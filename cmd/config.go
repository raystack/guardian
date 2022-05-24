package cmd

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/salt/cmdx"
	"github.com/spf13/cobra"
)

var cliConfig *Config

type Config struct {
	Host                string `mapstructure:"host"`
	EncryptionSecretKey string `mapstructure:"encryption_secret_key"`
}

func LoadConfig() (*Config, error) {
	var config Config

	cfg := cmdx.SetConfig("guardian")
	err := cfg.Load(&config)

	return &config, err
}

func bindFlagsFromConfig(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("host", "H", "", "Guardian service to connect to")

	if cliConfig != nil {
		if cliConfig.Host != "" {
			cmd.PersistentFlags().Set("host", cliConfig.Host)
		}
	}
}

func configCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <command>",
		Short: "Manage client configurations",
		Example: heredoc.Doc(`
			$ guardian config init
			$ guardian config list`),
	}

	cmd.AddCommand(configInitCommand())
	cmd.AddCommand(configListCommand())

	return cmd
}

func configInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new client configuration",
		Example: heredoc.Doc(`
			$ guardian config init
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := cmdx.SetConfig("guardian")

			if err := cfg.Init(&Config{}); err != nil {
				return err
			}

			fmt.Printf("config created: %v\n", cfg.File())
			return nil
		},
	}
}

func configListCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List client configuration settings",
		Example: heredoc.Doc(`
			$ guardian config list
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
