package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"vela-migrator/pkg/types"
	"vela-migrator/pkg/utils"

	kubeCfg "github.com/kubevela/velaux/pkg/server/config"
	"github.com/kubevela/velaux/pkg/server/domain/model"
	"github.com/kubevela/velaux/pkg/server/infrastructure/clients"
	"github.com/kubevela/velaux/pkg/server/infrastructure/datastore"
	"github.com/kubevela/velaux/pkg/server/infrastructure/datastore/kubeapi"
	"github.com/kubevela/velaux/pkg/server/infrastructure/datastore/mongodb"
	"github.com/kubevela/velaux/pkg/server/infrastructure/datastore/mysql"
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
	case "kubeapi":
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

func FilterTables(models map[string]model.Interface, tableNames []string) (tables []datastore.Entity) {
	if len(tableNames) == 0 {
		for _, m := range models {
			tables = append(tables, m.(datastore.Entity))
		}
	} else {
		for _, tableName := range tableNames {
			m := models[tableName]
			if m != nil {
				tables = append(tables, m.(datastore.Entity))
			}
		}
	}
	return tables
}

func Rollback(ctx context.Context, target datastore.DataStore, tables []datastore.Entity, deleteTables map[string][]datastore.Entity, restoreTables map[string][]datastore.Entity) error {
	fmt.Println("Initiating rollback...")
	for _, table := range tables {
		var tableDeleted = false
		var tableRestored = false
		if deleteEntities, exists := deleteTables[table.TableName()]; exists {
			for _, deleteEntity := range deleteEntities {
				tableDeleted = true
				if err := target.Delete(ctx, deleteEntity); err != nil {
					fmt.Println("Error in rolling back")
					return err
				}
			}
		}
		if tableDeleted {
			fmt.Printf("Deleted the entries of %s table...\n", table.TableName())
		}
		if entities, exists := restoreTables[table.TableName()]; exists {
			for _, entity := range entities {
				tableRestored = true
				if err := target.Put(ctx, entity); err != nil {
					fmt.Println("Error in rolling back")
					return err
				}
			}
		}
		if tableRestored {
			fmt.Printf("Restored the entries of %s table...\n", table.TableName())
		}
	}
	fmt.Println("All changes are rolled back")
	return nil
}

func Migrate(ctx context.Context, source datastore.DataStore, target datastore.DataStore, actionOnDup string, tables []datastore.Entity) error {
	deleteTables := make(map[string][]datastore.Entity)
	restoreTables := make(map[string][]datastore.Entity)
	var err error
	rollback := func() error {
		errOnRollback := Rollback(ctx, target, tables, deleteTables, restoreTables)
		if errOnRollback != nil {
			return errOnRollback
		}
		return err
	}
	for _, table := range tables {
		skipOnDup := strings.ToLower(actionOnDup) == types.SkipOnDup
		updateOnDup := strings.ToLower(actionOnDup) == types.UpdateOnDup
		var tableMigrated = false
		rows, err := source.List(ctx, table, nil)
		if err != nil {
			fmt.Printf("Error migrating %s table\n", table.TableName())
			return rollback()
		}
		for _, saveEntity := range rows {
			tableMigrated = true
			initialEntity, err := utils.CloneEntity(saveEntity)
			if err != nil {
				return err
			}
			if err = target.Add(ctx, saveEntity); err != nil {
				if errors.Is(err, datastore.ErrRecordExist) {
					if skipOnDup {
						continue
					}
					if updateOnDup {
						if err = target.Put(ctx, initialEntity); err != nil {
							return rollback()
						}
						restoreTables[saveEntity.TableName()] = append(restoreTables[saveEntity.TableName()], initialEntity)
						continue
					}
					fmt.Printf("Error migrating %s table\n", table.TableName())
					return rollback()
				} else {
					fmt.Printf("Error migrating %s table\n", table.TableName())
					return rollback()
				}
			}
			deleteTables[saveEntity.TableName()] = append(deleteTables[saveEntity.TableName()], initialEntity)
		}
		if tableMigrated {
			fmt.Printf("Migrated table %s\n", table.TableName())
		}
	}
	fmt.Println("All tables are migrated")
	return nil
}
