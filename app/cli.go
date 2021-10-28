package app

import (
	"errors"
	"fmt"

	"github.com/odpf/salt/config"
)

var (
	CLIConfigFile = "./.guardian.yaml"
)

type CLIConfig struct {
	Host string `mapstructure:"host" default:"localhost"`
}

func LoadCLIConfig(configFile string) (*CLIConfig, error) {
	var cfg CLIConfig
	loader := config.NewLoader(config.WithFile(configFile))

	if err := loader.Load(&cfg); err != nil {
		if errors.As(err, &config.ConfigFileNotFoundError{}) {
			fmt.Println(err)
			return &cfg, nil
		}
		return &CLIConfig{}, err
	}
	return &cfg, nil
}
