package cmd

import (
	"github.com/odpf/guardian/app"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "migrate",
		Short: "Migrate database schema",
		RunE:  migrate,
	})
}

func migrate(cmd *cobra.Command, args []string) error {
	c := app.LoadConfig()
	return app.Migrate(c)
}
