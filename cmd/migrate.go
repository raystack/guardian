package cmd

import (
	"github.com/odpf/guardian/app"
	"github.com/spf13/cobra"
)

func migrateCommand() *cobra.Command {

	var configFile string

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Annotations: map[string]string{
			"group:other": "dev",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.LoadConfig(configFile)
			if err != nil {
				return err
			}
			return app.Migrate(&cfg)
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "./config.yaml", "Config file path")
	return cmd
}
