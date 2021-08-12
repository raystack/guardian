package cmd

import (
	"github.com/odpf/guardian/app"
	"github.com/spf13/cobra"
)

func serveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Run server",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := app.LoadConfig()
			return app.RunServer(c)
		},
	}
}
