package domain

type DatabaseConfig struct {
	Host     string `mapstructure:"host" default:"localhost"`
	User     string `mapstructure:"user" default:"postgres"`
	Password string `mapstructure:"password" default:""`
	Name     string `mapstructure:"name" default:"postgres"`
	Port     string `mapstructure:"port" default:"5432"`
	SslMode  string `mapstructure:"sslmode" default:"disable"`
	LogLevel string `mapstructure:"log_level" default:"info"`
}

type LogConfig struct {
	Level string `mapstructure:"level" default:"info"`
}

type Config struct {
	Port                   int                    `mapstructure:"port" default:"8080"`
	EncryptionSecretKeyKey string                 `mapstructure:"encryption_secret_key"`
	SlackAccessToken       string                 `mapstructure:"slack_access_token"`
	IAM                    map[string]interface{} `mapstructure:"iam"`
	Log                    LogConfig              `mapstructure:"log"`
	DB                     DatabaseConfig         `mapstructure:"db"`
}
