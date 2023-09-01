package cmd

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"time"
	"vela-migrator/pkg/types"
	"vela-migrator/pkg/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	mysqlgorm "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/kubevela/velaux/pkg/server/domain/model"
	"github.com/kubevela/velaux/pkg/server/infrastructure/clients"
	"github.com/kubevela/velaux/pkg/server/infrastructure/datastore"
	"github.com/kubevela/velaux/pkg/server/infrastructure/datastore/kubeapi"
	"github.com/kubevela/velaux/pkg/server/infrastructure/datastore/mongodb"
	"github.com/kubevela/velaux/pkg/server/infrastructure/datastore/mysql"
)

type migrationTestData struct {
	SourceData   []datastore.Entity
	TargetData   []datastore.Entity
	ExpectedData []datastore.Entity
	ActionOnDup  string
}

func insertIntoDatabase(db datastore.DataStore, entities []datastore.Entity) error {
	for _, entity := range entities {
		err := db.Add(context.TODO(), entity)
		if err != nil {
			return err
		}
	}
	return nil
}

func initMysqlTestDs() (datastore.DataStore, error) {
	db, err := gorm.Open(mysqlgorm.Open("root:kubevelaSQL123@tcp(127.0.0.1:3306)/kubevela?parseTime=True"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	for _, v := range model.GetRegisterModels() {
		err := db.Migrator().DropTable(v)
		if err != nil {
			return nil, err
		}
	}
	mysqlDriver, err := mysql.New(context.TODO(), datastore.Config{
		URL:      "root:kubevelaSQL123@tcp(127.0.0.1:3306)/kubevela?parseTime=True",
		Database: "kubevela",
	})
	if err != nil {
		return nil, err
	}
	return mysqlDriver, nil
}

func initKubeapiTestDs() (datastore.DataStore, error) {
	var testScheme = runtime.NewScheme()
	testEnv := &envtest.Environment{
		ControlPlaneStartTimeout: time.Minute * 3,
		ControlPlaneStopTimeout:  time.Minute,
		UseExistingCluster:       pointer.Bool(false),
	}
	cfg, err := testEnv.Start()
	if err != nil {
		return nil, err
	}
	err = scheme.AddToScheme(testScheme)
	if err != nil {
		return nil, err
	}
	cfg.Timeout = time.Minute * 2
	k8sClient, err := client.New(cfg, client.Options{Scheme: testScheme})
	if err != nil {
		return nil, err
	}
	clients.SetKubeClient(k8sClient)
	kubeStore, err := kubeapi.New(context.TODO(), datastore.Config{Database: "test"}, k8sClient)
	if err != nil {
		return nil, err
	}
	return kubeStore, nil
}

func initMongodbTestDs() (datastore.DataStore, error) {
	clientOpts := options.Client().ApplyURI("mongodb+srv://ngdev21:21102002n@clusterdev.eodjv7o.mongodb.net/?retryWrites=true&w=majority")
	mongoClient, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		return nil, err
	}
	err = mongoClient.Database("kubevela").Drop(context.TODO())
	if err != nil {
		return nil, err
	}
	mongodbDriver, err := mongodb.New(context.TODO(), datastore.Config{
		URL:      "mongodb+srv://ngdev21:21102002n@clusterdev.eodjv7o.mongodb.net/?retryWrites=true&w=majority",
		Database: "kubevela",
	})
	if err != nil {
		return nil, err
	}
	return mongodbDriver, nil
}

func initTestDb(dbType string) (datastore.DataStore, error) {
	switch dbType {
	case "mysql":
		return initMysqlTestDs()
	case "mongodb":
		return initMongodbTestDs()
	case "kubeapi":
		return initKubeapiTestDs()
	default:
		return nil, fmt.Errorf("invalid database type")
	}
}

func checkEqualDataSlice(output []datastore.Entity, expected []datastore.Entity) bool {
	if len(output) != len(expected) {
		return false
	}
	sort.Slice(output, func(i, j int) bool {
		return output[i].PrimaryKey() > output[j].PrimaryKey()
	})
	sort.Slice(expected, func(i, j int) bool {
		return expected[i].PrimaryKey() > expected[j].PrimaryKey()
	})
	for i := range output {
		cloneOutput, err := utils.CloneEntity(output[i])
		if err != nil {
			return false
		}
		cloneExpected, err := utils.CloneEntity(expected[i])
		if err != nil {
			return false
		}
		cloneOutput.SetUpdateTime(time.Time{})
		cloneOutput.SetCreateTime(time.Time{})
		cloneExpected.SetUpdateTime(time.Time{})
		cloneExpected.SetCreateTime(time.Time{})
		if !reflect.DeepEqual(cloneExpected, cloneOutput) {
			return false
		}
	}
	return true
}

var _ = Describe("Test helper functions", func() {
	It("Test FilterTable functions", func() {
		models := model.GetRegisterModels()
		tableNames := []string{"vela_application", "vela_user"}
		outTables := []datastore.Entity{models["vela_application"].(datastore.Entity), models["vela_user"].(datastore.Entity)}
		tables := FilterTables(models, tableNames)
		Expect(len(tables)).Should(Equal(len(outTables)))
		tables = FilterTables(models, []string{})
		Expect(len(tables)).Should(Equal(len(models)))
		tables = FilterTables(models, []string{"dummy"})
		Expect(len(tables)).Should(Equal(0))
	})
	It("Test migrate Function", func() {
		dbTypes := []string{"kubeapi", "mongodb"}

		updTestData := migrationTestData{
			SourceData: []datastore.Entity{
				&model.Application{
					Name:        "app1",
					Description: "newApp",
				},
				&model.User{
					Name:  "user1",
					Alias: "newUser",
				},
			},
			TargetData: []datastore.Entity{
				&model.Application{
					Name:        "app1",
					Description: "oldApp",
				},
				&model.User{
					Name:  "user2",
					Alias: "oldUser",
				},
			},
			ExpectedData: []datastore.Entity{
				&model.Application{
					Name:        "app1",
					Description: "newApp",
				},
				&model.User{
					Name:  "user1",
					Alias: "newUser",
				},
				&model.User{
					Name:  "user2",
					Alias: "oldUser",
				},
			},
			ActionOnDup: "update",
		}

		skipTestData := migrationTestData{
			SourceData: []datastore.Entity{
				&model.Application{
					Name:        "app1",
					Description: "newApp",
				},
				&model.User{
					Name:  "user1",
					Alias: "newUser",
				},
			},
			TargetData: []datastore.Entity{
				&model.Application{
					Name:        "app1",
					Description: "oldApp",
				},
				&model.User{
					Name:  "user2",
					Alias: "oldUser",
				},
			},
			ExpectedData: []datastore.Entity{
				&model.Application{
					Name:        "app1",
					Description: "oldApp",
				},
				&model.User{
					Name:  "user1",
					Alias: "newUser",
				},
				&model.User{
					Name:  "user2",
					Alias: "oldUser",
				},
			},
			ActionOnDup: "skip",
		}

		errTestData := migrationTestData{
			SourceData: []datastore.Entity{
				&model.User{
					Name:  "user1",
					Alias: "newUser",
				},
				&model.Application{
					Name:        "app1",
					Description: "newApp",
				},
			},
			TargetData: []datastore.Entity{
				&model.User{
					Name:  "user2",
					Alias: "oldUser",
				},
				&model.Application{
					Name:        "app1",
					Description: "oldApp",
				},
			},
			ExpectedData: []datastore.Entity{
				&model.User{
					Name:  "user2",
					Alias: "oldUser",
				},
				&model.Application{
					Name:        "app1",
					Description: "oldApp",
				},
			},
		}

		for i := range dbTypes {
			for j := range dbTypes {
				if i == j {
					continue
				}
				By("Test migrator function with update on duplicate")
				testMigrate(dbTypes[i], dbTypes[j], updTestData.ActionOnDup, []string{"vela_user", "vela_application"}, updTestData.SourceData, updTestData.TargetData, updTestData.ExpectedData, false)
				By("Test migrator function with skip on duplicate")
				testMigrate(dbTypes[i], dbTypes[j], skipTestData.ActionOnDup, []string{"vela_user", "vela_application"}, skipTestData.SourceData, skipTestData.TargetData, skipTestData.ExpectedData, false)
				By("Test migrator function with err on duplicate")
				testMigrate(dbTypes[i], dbTypes[j], errTestData.ActionOnDup, []string{"vela_user", "vela_application"}, errTestData.SourceData, errTestData.TargetData, errTestData.ExpectedData, true)
			}
		}
	})
})

func testMigrate(sourceDbType string, targetDbType string, actionOnDup string, tables []string, sourceDbData []datastore.Entity, targetDbData []datastore.Entity, expectedData []datastore.Entity, isError bool) {
	source, err := initTestDb(sourceDbType)
	Expect(err).Should(BeNil())
	target, err := initTestDb(targetDbType)
	Expect(err).Should(BeNil())
	err = insertIntoDatabase(source, sourceDbData)
	Expect(err).Should(BeNil())
	err = insertIntoDatabase(target, targetDbData)
	Expect(err).Should(BeNil())
	tablesToMigrate := FilterTables(model.GetRegisterModels(), tables)
	err = Migrate(context.TODO(), source, target, actionOnDup, tablesToMigrate)
	switch actionOnDup {
	case types.SkipOnDup:
		{
			Expect(err).Should(BeNil())
		}
	case types.UpdateOnDup:
		{
			Expect(err).Should(BeNil())
		}
	default:
		{
			if isError {
				Expect(err).ShouldNot(BeNil())
			} else {
				Expect(err).Should(BeNil())
			}
		}
	}
	var outputData []datastore.Entity
	for _, tables := range tablesToMigrate {
		list, err := target.List(context.TODO(), tables, nil)
		Expect(err).ShouldNot(HaveOccurred())
		outputData = append(outputData, list...)
	}
	Expect(checkEqualDataSlice(outputData, expectedData)).Should(Equal(true))
}
