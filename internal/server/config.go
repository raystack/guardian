package server

import (
	"errors"
	"fmt"

	"github.com/odpf/guardian/internal/store"
	"github.com/odpf/guardian/plugins/notifiers"
	"github.com/odpf/salt/config"
)

type JobType string

const (
	FetchResources             JobType = "fetch_resources"
	ExpiringAccessNotification JobType = "appeal_expiration_reminder"
	RevokeExpiredAccess        JobType = "appeal_expiration_revocation"
)

type JobConfig struct {
	JobType  JobType
	Enabled  bool   `mapstructure:"enabled" default:"true"`
	Interval string `mapstructure:"interval"`
}

type Jobs struct {
	FetchResources             JobConfig `mapstructure:"fetch_resources"`
	RevokeExpiredAccess        JobConfig `mapstructure:"revoke_expired_access"`
	ExpiringAccessNotification JobConfig `mapstructure:"expiring_access_notification"`
}

type Config struct {
	Port                       int              `mapstructure:"port" default:"8080"`
	EncryptionSecretKeyKey     string           `mapstructure:"encryption_secret_key"`
	Notifier                   notifiers.Config `mapstructure:"notifier"`
	LogLevel                   string           `mapstructure:"log_level" default:"info"`
	DB                         store.Config     `mapstructure:"db"`
	AuthenticatedUserHeaderKey string           `mapstructure:"authenticated_user_header_key"`
	AuditLogTraceIDHeaderKey   string           `mapstructure:"audit_log_trace_id_header_key" default:"X-Trace-Id"`
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
