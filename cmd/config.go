package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/mcuadros/go-defaults"
	"github.com/odpf/guardian/app"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func configCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "manage guardian CLI configuration",
	}
	cmd.AddCommand(configInitCommand())
	return cmd
}

func configInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "initialize CLI configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			var config app.CLIConfig
			defaults.SetDefaults(&config)

			b, err := yaml.Marshal(&config)
			if err != nil {
				return err
			}

			filepath := fmt.Sprintf("%v", app.CLIConfigFile)
			if err := ioutil.WriteFile(filepath, b, 0655); err != nil {
				return err
			}
			fmt.Printf("config created: %v\n", filepath)

			return nil
		},
	}
}
