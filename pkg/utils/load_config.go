package utils

import (
	"github.com/kubevela/velaux/pkg/server/infrastructure/datastore"
	"github.com/spf13/viper"
)

type MigratorConfig struct {
	Source     datastore.Config
	Target     datastore.Config
	Tables     []string
	ErrorOnDup string
}

func LoadConfig(path string) (config MigratorConfig, err error) {
	viper.SetConfigFile(path)
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}
