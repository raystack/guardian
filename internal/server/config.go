package server

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/raystack/guardian/internal/store"
	"github.com/raystack/guardian/jobs"
	"github.com/raystack/guardian/pkg/auth"
	"github.com/raystack/guardian/pkg/tracing"
	"github.com/raystack/guardian/plugins/notifiers"
	"github.com/raystack/salt/config"
)

type DefaultAuth struct {
	HeaderKey string `mapstructure:"header_key" default:"X-Auth-Email"`
}

type Auth struct {
	Provider string              `mapstructure:"provider" default:"default"`
	Default  DefaultAuth         `mapstructure:"default"`
	OIDC     auth.OIDCAuth       `mapstructure:"oidc"`
	Frontier auth.FrontierConfig `mapstructure:"frontier"`
}

type Jobs struct {
	FetchResources             jobs.Job `mapstructure:"fetch_resources"`
	RevokeExpiredGrants        jobs.Job `mapstructure:"revoke_expired_grants"`
	ExpiringGrantNotification  jobs.Job `mapstructure:"expiring_grant_notification"`
	RevokeGrantsByUserCriteria jobs.Job `mapstructure:"revoke_grants_by_user_criteria"`
	GrantDormancyCheck         jobs.Job `mapstructure:"grant_dormancy_check"`

	// Deprecated: use ExpiringGrantNotification instead
	ExpiringAccessNotification jobs.Job `mapstructure:"expiring_access_notification"`
	// Deprecated: use RevokeExpiredGrants instead
	RevokeExpiredAccess jobs.Job `mapstructure:"revoke_expired_access"`
}

type Config struct {
	Port                   int              `mapstructure:"port" default:"8080"`
	GRPC                   GRPCConfig       `mapstructure:"grpc"`
	EncryptionSecretKeyKey string           `mapstructure:"encryption_secret_key"`
	Notifier               notifiers.Config `mapstructure:"notifier"`
	LogLevel               string           `mapstructure:"log_level" default:"info"`
	DB                     store.Config     `mapstructure:"db"`
	// Deprecated: use Auth.Default.HeaderKey instead note on the AuthenticatedUserHeaderKey
	AuthenticatedUserHeaderKey string         `mapstructure:"authenticated_user_header_key"`
	AuditLogTraceIDHeaderKey   string         `mapstructure:"audit_log_trace_id_header_key" default:"X-Trace-Id"`
	Jobs                       Jobs           `mapstructure:"jobs"`
	Telemetry                  tracing.Config `mapstructure:"telemetry"`
	Auth                       Auth           `mapstructure:"auth"`
}

type GRPCConfig struct {
	// TimeoutInSeconds is the maximum time in seconds a request can take before being cancelled. Default = 5
	TimeoutInSeconds int `mapstructure:"timeout_in_seconds" default:"5"`
	// MaxCallRecvMsgSize is the maximum message size the server can receive in bytes. Default = 1 << 25 (32MB)
	MaxCallRecvMsgSize int `mapstructure:"max_call_recv_msg_size" default:"33554432"`
	// MaxCallSendMsgSize is the maximum message size the server can send in bytes. Default = 1 << 25 (32MB)
	MaxCallSendMsgSize int `mapstructure:"max_call_send_msg_size" default:"33554432"`
}

func LoadConfig(serverConfigFileFromFlag string) (Config, error) {
	var cfg Config
	var options []config.LoaderOption
	options = append(options, config.WithName("config"))
	options = append(options, config.WithEnvKeyReplacer(".", "_"))
	options = append(options, config.WithEnvPrefix("GUARDIAN"))
	if p, err := os.Getwd(); err == nil {
		options = append(options, config.WithPath(p))
	}
	if execPath, err := os.Executable(); err == nil {
		options = append(options, config.WithPath(filepath.Dir(execPath)))
	}
	if currentHomeDir, err := os.UserHomeDir(); err == nil {
		options = append(options, config.WithPath(currentHomeDir))
		options = append(options, config.WithPath(filepath.Join(currentHomeDir, ".config")))
	}

	// override all config sources and prioritize one from file
	if serverConfigFileFromFlag != "" {
		options = append(options, config.WithFile(serverConfigFileFromFlag))
	}

	loader := config.NewLoader(options...)

	if err := loader.Load(&cfg); err != nil {
		if errors.As(err, &config.ConfigFileNotFoundError{}) {
			fmt.Println(err)
			return cfg, nil
		}
		return Config{}, err
	}

	// keep for backward-compatibility
	if cfg.AuthenticatedUserHeaderKey != "" {
		cfg.Auth.Default.HeaderKey = cfg.AuthenticatedUserHeaderKey
	}

	// fail if encryption secret key is not set
	if cfg.EncryptionSecretKeyKey == "" {
		fmt.Println("WARNING: encryption_secret_key is not set")
	}
	return cfg, nil
}
