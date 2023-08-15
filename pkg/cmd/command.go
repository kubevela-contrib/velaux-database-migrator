package cmd

import (
	"github.com/spf13/cobra"
	"vela-migrator/pkg/utils"
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
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		config, err := utils.LoadConfig(args[0])
		if err != nil {
			return err
		}
		err = Migrate(cmd.Context(), config)
		if err != nil {
			return err
		}
		return nil
	}
	return cmd
}
