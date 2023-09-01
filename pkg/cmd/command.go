package cmd

import (
	"vela-migrator/pkg/utils"

	"k8s.io/klog/v2"

	"github.com/spf13/cobra"
)

// NewMigrateCmd migrate database
func NewMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "migrate the database",
		Long:  "migrate the database",
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfgFile, err := cmd.Flags().GetString("config-file")
		if err != nil {
			return err
		}
		config, err := utils.LoadConfig(cfgFile)
		if err != nil {
			return err
		}
		return migrateCmd(cmd.Context(), config)
	}
	cmd.Flags().StringP("config-file", "c", "", "specify the path to the config file")
	err := cmd.MarkFlagRequired("config-file")
	if err != nil {
		klog.Info(err)
		return nil
	}
	return cmd
}

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
