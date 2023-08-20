package utils

import (
	"reflect"
	"vela-migrator/pkg/types"

	"github.com/jinzhu/copier"
	"github.com/spf13/viper"

	"github.com/kubevela/velaux/pkg/server/infrastructure/datastore"
)

func LoadConfig(path string) (config types.MigratorConfig, err error) {
	viper.SetConfigFile(path)
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}

func CloneEntity(input datastore.Entity) (datastore.Entity, error) {
	temp := reflect.New(reflect.ValueOf(input).Elem().Type()).Interface().(datastore.Entity)
	err := copier.CopyWithOption(temp, input, copier.Option{})
	if err != nil {
		return nil, err
	}
	return temp, nil
}
