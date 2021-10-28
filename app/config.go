package app

import (
	"github.com/odpf/guardian/iam"
	"github.com/odpf/guardian/notifier"
	"github.com/odpf/guardian/store"
	"github.com/odpf/salt/config"
)

type Config struct {
	Port                       int                   `mapstructure:"port" default:"8080"`
	EncryptionSecretKeyKey     string                `mapstructure:"encryption_secret_key"`
	Notifier                   notifier.ClientConfig `mapstructure:"notifier"`
	IAM                        iam.ClientConfig      `mapstructure:"iam"`
	LogLevel                   string                `mapstructure:"log_level" default:"info"`
	DB                         store.Config          `mapstructure:"db"`
	AuthenticatedUserHeaderKey string                `mapstructure:"authenticated_user_header_key"`
	Jobs                       Jobs                  `mapstructure:"jobs"`
}

//  LoadConfig the configuration from the given path.
func LoadConfig(configFile string) (Config, error) {
	var cfg Config
	loader := config.NewLoader(config.WithFile(configFile))

	if err := loader.Load(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
