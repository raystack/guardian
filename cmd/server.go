package cmd

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/guardian/app"
	"github.com/spf13/cobra"
)

func ServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server <command>",
		Aliases: []string{"s"},
		Short:   "Server management",
		Long:    "Server management commands.",
		Example: heredoc.Doc(`
			$ guardian server start
			$ guardian server start -c ./config.yaml
			$ guardian server migrate
			$ guardian server migrate -c ./config.yaml
		`),
	}

	cmd.AddCommand(startCommand())
	cmd.AddCommand(migrateCommand())

	return cmd
}

func startCommand() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:     "start",
		Aliases: []string{"s"},
		Short:   "Start the server",
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

func migrateCommand() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
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
