package cmd

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/guardian/app"
	"github.com/odpf/salt/cmdx"
	"github.com/spf13/cobra"
)

func configCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <command>",
		Short: "Manage client configuration settings",
		Example: heredoc.Doc(`
			$ stencil config init
			$ stencil config get
			$ stencil config list`),
	}

	cmd.AddCommand(configInitCommand())
	cmd.AddCommand(configListCommand())
	cmd.AddCommand(configGetCommand())

	return cmd
}

func configInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new client configuration",
		Example: heredoc.Doc(`
			$ stencil init --path=.stencil.yml
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := cmdx.SetConfig("guardian")

			if err := cfg.Init(&app.CLIConfig{}); err != nil {
				return err
			}

			fmt.Printf("config created: %v\n", cfg.File())
			return nil
		},
	}
}

func configGetCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "get <key>",
		Short: "Update configuration with a value for the given key",
		Example: heredoc.Doc(`
			$ stencil config get --path=.stencil.yml
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := app.LoadCLIConfig()
			if err != nil {
				return err
			}
			fmt.Println(config)
			return nil
		},
	}
	return cmd
}

func configListCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "Update configuration with a value for the given key",
		Example: heredoc.Doc(`
			$ stencil config list --path=.stencil.yml
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := cmdx.SetConfig("guardian")

			data, err := cfg.Read()
			if err != nil {
				return err
			}

			fmt.Println(data)
			return nil
		},
	}
	return cmd
}
