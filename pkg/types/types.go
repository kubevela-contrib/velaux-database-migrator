package types

import "github.com/kubevela/velaux/pkg/server/infrastructure/datastore"

type MigratorConfig struct {
	Source      datastore.Config
	Target      datastore.Config
	Tables      []string
	ActionOnDup string
}

const (
	SkipOnDup   = "skip"
	UpdateOnDup = "update"
)
