package cmd

import (
	"fmt"
	"os"

	v1 "github.com/odpf/guardian/api/handler/v1"
	"github.com/spf13/cobra"
)

// Execute runs the command line interface
func Execute() {
	var rootCmd = &cobra.Command{
		Use: "guardian",
	}

	protoAdapter := v1.NewAdapter()

	cliConfig, err := readConfig()
	if err != nil {
		fmt.Println(err)
	}

	rootCmd.AddCommand(serveCommand())
	rootCmd.AddCommand(migrateCommand())
	rootCmd.AddCommand(configCommand())
	rootCmd.AddCommand(resourcesCommand(cliConfig))
	rootCmd.AddCommand(providersCommand(cliConfig, protoAdapter))
	rootCmd.AddCommand(policiesCommand(cliConfig, protoAdapter))
	rootCmd.AddCommand(appealsCommand(cliConfig))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
