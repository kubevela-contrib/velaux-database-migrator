package cmd

import (
	"fmt"
	"vela-migrator/pkg/utils"

	"github.com/spf13/cobra"
)

// NewMigratorCommand create migrator command
func NewMigratorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "velamg",
		Short: "VelaUX migrator",
		Long:  "VelaUX migrator",
	}
	cmd.AddCommand(NewMigrateCmd())
	return cmd
}

// NewMigrateCmd migrate database
func NewMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "migrate the database",
		Long:  "migrate the database",
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("too many arguments")
		}
		if len(args) == 0 {
			return fmt.Errorf("provide the path of the config file")
		}
		config, err := utils.LoadConfig(args[0])
		if err != nil {
			return err
		}
		err = migrateCmd(cmd.Context(), config)
		if err != nil {
			return err
		}
		return nil
	}
	return cmd
}
