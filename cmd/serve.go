package cmd

import (
	"github.com/odpf/guardian/app"
	"github.com/spf13/cobra"
)

// serveCommand start new guardian server
func serveCommand() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:     "serve",
		Aliases: []string{"s"},
		Short:   "Run guardian server",
		Annotations: map[string]string{
			"group:other": "dev",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.LoadConfig(configFile)
			if err != nil {
				return err
			}
			return app.RunServer(&cfg)
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "./config.yaml", "Config file path")
	return cmd
}
