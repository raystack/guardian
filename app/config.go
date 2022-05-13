package app

import (
	"errors"
	"fmt"

	"github.com/odpf/guardian/plugins/notifiers"
	"github.com/odpf/guardian/store"
	"github.com/odpf/salt/config"
)

type Jobs struct {
	FetchResourcesInterval             string `mapstructure:"fetch_resources_interval" default:"0 */2 * * *"`
	RevokeExpiredAccessInterval        string `mapstructure:"revoke_expired_access_interval" default:"*/20 * * * *"`
	ExpiringAccessNotificationInterval string `mapstructure:"expiring_access_notification_interval" default:"0 9 * * *"`
}

type Config struct {
	Port                       int              `mapstructure:"port" default:"8080"`
	EncryptionSecretKeyKey     string           `mapstructure:"encryption_secret_key"`
	Notifier                   notifiers.Config `mapstructure:"notifier"`
	LogLevel                   string           `mapstructure:"log_level" default:"info"`
	DB                         store.Config     `mapstructure:"db"`
	AuthenticatedUserHeaderKey string           `mapstructure:"authenticated_user_header_key"`
	Jobs                       Jobs             `mapstructure:"jobs"`
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
