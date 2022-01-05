package app

import (
	"errors"
	"fmt"

	"github.com/odpf/guardian/plugins/notifiers"
	"github.com/odpf/guardian/store"
	"github.com/odpf/salt/cmdx"
	"github.com/odpf/salt/config"
)

type Config struct {
	Port                       int              `mapstructure:"port" default:"8080"`
	EncryptionSecretKeyKey     string           `mapstructure:"encryption_secret_key"`
	Notifier                   notifiers.Config `mapstructure:"notifier"`
	LogLevel                   string           `mapstructure:"log_level" default:"info"`
	DB                         store.Config     `mapstructure:"db"`
	AuthenticatedUserHeaderKey string           `mapstructure:"authenticated_user_header_key"`
	Jobs                       Jobs             `mapstructure:"jobs"`
}

type CLIConfig struct {
	Host string `mapstructure:"host" default:"localhost"`
}

func LoadConfig(configFile string) (Config, error) {
	var cfg Config
	loader := config.NewLoader(config.WithFile(configFile))

	if err := loader.Load(&cfg); err != nil {
		if errors.As(err, &config.ConfigFileNotFoundError{}) {
			fmt.Println(err)
			return cfg, nil
		}
		return Config{}, err
	}
	return cfg, nil
}

func LoadCLIConfig() (*CLIConfig, error) {
	var config CLIConfig

	cfg := cmdx.SetConfig("guardian")
	err := cfg.Load(&config)

	return &config, err
}
