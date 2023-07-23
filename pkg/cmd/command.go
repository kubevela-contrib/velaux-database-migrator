package cmd

import (
	"github.com/spf13/cobra"
)

// NewMigratorCommand create migrator command
func NewMigratorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrator",
		Short: "VelaUX migrator",
		Long:  "VelaUX migrator",
	}
	cmd.AddCommand(NewMigrateCmd())
	return cmd
}

// NewMigrateCmd migrate database
func NewMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "velamg",
		Short: "migrate the database",
		Long:  "migrate the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			Migrate(cmd.Context())
			return nil
		},
	}
	return cmd
}
