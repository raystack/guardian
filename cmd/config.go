package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	DefaultHost   = "localhost"
	FileName      = ".guardian"
	FileExtension = "yaml"
)

type config struct {
	Host string `yaml:"host"`
}

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
			config := config{
				Host: DefaultHost,
			}

			b, err := yaml.Marshal(config)
			if err != nil {
				return err
			}

			filepath := fmt.Sprintf("%v.%v", FileName, FileExtension)
			if err := ioutil.WriteFile(filepath, b, 0655); err != nil {
				return err
			}
			fmt.Printf("config created: %v", filepath)

			return nil
		},
	}
}

func readConfig() (*config, error) {
	var c config
	filepath := fmt.Sprintf("%v.%v", FileName, FileExtension)
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, err
	}

	return &c, nil
}
