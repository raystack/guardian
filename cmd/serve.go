package cmd

import (
	"github.com/odpf/guardian/app"
	"github.com/odpf/guardian/config"
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
	c := config.Load()
	return app.RunServer(c)
}
