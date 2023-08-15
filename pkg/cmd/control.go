package cmd

import (
	"context"
	"errors"
	"fmt"
	kubeCfg "github.com/kubevela/velaux/pkg/server/config"
	"github.com/kubevela/velaux/pkg/server/domain/model"
	"github.com/kubevela/velaux/pkg/server/infrastructure/clients"
	"github.com/kubevela/velaux/pkg/server/infrastructure/datastore"
	"github.com/kubevela/velaux/pkg/server/infrastructure/datastore/kubeapi"
	"github.com/kubevela/velaux/pkg/server/infrastructure/datastore/mongodb"
	"github.com/kubevela/velaux/pkg/server/infrastructure/datastore/mysql"
	"strings"
	"vela-migrator/pkg/utils"
)

// InitDataStore initDataStore init datastore with given config
func InitDataStore(ctx context.Context, config datastore.Config) (datastore.DataStore, error) {
	fmt.Printf("Initialising the %s database\n", config.Type)
	var dbType = config.Type
	switch dbType {
	case "mongodb":
		return mongodb.New(ctx, config)
	case "mysql":
		return mysql.New(ctx, config)
	case "kube-api":
		err := clients.SetKubeConfig(*kubeCfg.NewConfig())
		if err != nil {
			return nil, err
		}
		kubeClient, err := clients.GetKubeClient()
		if err != nil {
			return nil, err
		}
		return kubeapi.New(context.Background(), config, kubeClient)
	default:
		return nil, fmt.Errorf("database type is invalid")
	}
}

func filterTables(models map[string]model.Interface, tableNames []string) (tables []datastore.Entity) {
	tableMap := make(map[string]bool)
	for _, t := range tableNames {
		tableMap[t] = true
	}
	if len(tableNames) == 0 {
		for _, m := range models {
			tableMap[m.TableName()] = true
		}
	}
	for _, m := range models {
		datastoreModel := m.(datastore.Entity)
		if exists, _ := tableMap[datastoreModel.TableName()]; exists {
			tables = append(tables, datastoreModel)
		}
	}
	return tables
}

func rollback(ctx context.Context, target datastore.DataStore, tables []datastore.Entity, deleteTables map[string][]datastore.Entity, restoreTables map[string][]datastore.Entity) error {
	fmt.Println("Initiating rollback...")
	for _, table := range tables {
		fmt.Printf("Deleting the entries of %s table...\n", table.TableName())
		if deleteEntities, exists := deleteTables[table.TableName()]; exists {
			for _, deleteEntity := range deleteEntities {
				if err := target.Delete(ctx, deleteEntity); err != nil {
					fmt.Println("Error in rolling back")
					return err
				}
			}
		}
	}
	for _, table := range tables {
		fmt.Printf("Restoring the entries of %s table...\n", table.TableName())
		if entities, exists := restoreTables[table.TableName()]; exists {
			for _, entity := range entities {
				if err := target.Put(ctx, entity); err != nil {
					fmt.Println("Error in rolling back")
					return err
				}
			}
		}
	}
	fmt.Println("All changes are rolled back")
	return nil
}

func Migrate(ctx context.Context, config utils.MigratorConfig) error {
	source, err := InitDataStore(ctx, config.Source)
	if err != nil {
		fmt.Println("Error initialising source database")
		return err
	}
	target, err := InitDataStore(ctx, config.Target)
	if err != nil {
		fmt.Println("Error initialising source database")
		return err
	}
	fmt.Println("Initiating migration...")
	models := model.GetRegisterModels()
	tables := filterTables(models, config.Tables)
	deleteTables := make(map[string][]datastore.Entity)
	restoreTables := make(map[string][]datastore.Entity)
	for _, table := range tables {
		fmt.Printf("Migrating %s table...\n", table.TableName())
		skipOnDup := strings.ToLower(config.ErrorOnDup) == "skip"
		updateOnDup := strings.ToLower(config.ErrorOnDup) == "update"
		rows, err := source.List(ctx, table, nil)
		if err != nil {
			fmt.Printf("Error migrating %s table\n", table.TableName())
			err := rollback(ctx, target, tables, deleteTables, restoreTables)
			if err != nil {
				return err
			}
			return err
		}
		for _, saveEntity := range rows {
			var initialEntity = saveEntity
			if err = target.Add(ctx, saveEntity); err != nil {
				if errors.Is(err, datastore.ErrRecordExist) {
					if skipOnDup {
						continue
					}
					if updateOnDup {
						if err = target.Put(ctx, initialEntity); err != nil {
							err := rollback(ctx, target, tables, deleteTables, restoreTables)
							if err != nil {
								return err
							}
							return err
						}
						restoreTables[saveEntity.TableName()] = append(restoreTables[saveEntity.TableName()], saveEntity)
						continue
					}
					fmt.Printf("Error migrating %s table\n", table.TableName())
					err := rollback(ctx, target, tables, deleteTables, restoreTables)
					if err != nil {
						return err
					}
					return err
				} else {
					fmt.Printf("Error migrating %s table\n", table.TableName())
					err := rollback(ctx, target, tables, deleteTables, restoreTables)
					if err != nil {
						return err
					}
					return err
				}
			}
			deleteTables[saveEntity.TableName()] = append(deleteTables[saveEntity.TableName()], initialEntity)
		}
	}
	fmt.Println("All tables are migrated")
	return nil
}
