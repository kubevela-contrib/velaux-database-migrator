package cmd

import (
	"context"
	"fmt"
	"vela-migrator/pkg/types"

	"github.com/kubevela/velaux/pkg/server/domain/model"
)

func migrateCmd(ctx context.Context, config types.MigratorConfig) error {
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
	tables := FilterTables(models, config.Tables)
	return Migrate(ctx, source, target, config.ActionOnDup, tables)
}
