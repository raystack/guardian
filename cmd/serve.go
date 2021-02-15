package cmd

import (
	"github.com/odpf/guardian/app"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Run server",
		RunE:  serve,
	})
}

func serve(cmd *cobra.Command, args []string) error {
	c := app.LoadConfig()
	return app.RunServer(c)
}
