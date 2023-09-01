package types

import "github.com/kubevela/velaux/pkg/server/infrastructure/datastore"

// MigratorConfig defines the structure of the config file needed for the migration
type MigratorConfig struct {
	Source      datastore.Config
	Target      datastore.Config
	Tables      []string
	ActionOnDup string
}

const (
	// SkipOnDup the entry will be skipped if any error occurs during migration
	SkipOnDup = "skip"
	// UpdateOnDup the entry will be skipped if any error occurs during migration
	UpdateOnDup = "update"
)
