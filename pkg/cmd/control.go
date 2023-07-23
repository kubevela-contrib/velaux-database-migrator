package cmd

import (
	"context"
	"fmt"
	"github.com/kubevela/velaux/pkg/server/domain/model"
	"github.com/kubevela/velaux/pkg/server/infrastructure/datastore"
	"github.com/kubevela/velaux/pkg/server/infrastructure/datastore/mongodb"
	"github.com/kubevela/velaux/pkg/server/infrastructure/datastore/mysql"
)

// initDataStore init datastore with given config
func initDataStore(ctx context.Context, config datastore.Config) (datastore.DataStore, error) {
	var dbType = config.Type
	switch dbType {
	case "mongodb":
		return mongodb.New(ctx, config)
	case "mysql":
		return mysql.New(ctx, config)
	default:
		return nil, fmt.Errorf("database type is invalid")
	}
}

func migrate(ctx context.Context, source datastore.DataStore, target datastore.DataStore) error {
	application := model.Application{}
	list, err := source.List(ctx, &application, nil)
	if err != nil {
		return err
	}
	err = target.BatchAdd(ctx, list)
	if err != nil {
		return err
	}
	return nil
}

func Migrate(ctx context.Context) datastore.DataStore {
	source, err := initDataStore(ctx, datastore.Config{
		Type:     "",
		URL:      "",
		Database: "",
	})
	if err != nil {
		return nil
	}
	target, err := initDataStore(ctx, datastore.Config{
		Type:     "",
		URL:      "",
		Database: "",
	})
	if err != nil {
		return nil
	}
	if err = migrate(ctx, source, target); err != nil {
		return nil
	}
	return source
}
