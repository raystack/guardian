package cmd

import (
	"github.com/odpf/guardian/app"
	"github.com/spf13/cobra"
)

func migrateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Migrate database schema",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := app.LoadConfig()
			return app.Migrate(c)
		},
	}
}
