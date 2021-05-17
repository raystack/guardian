package app

import (
	"fmt"

	"github.com/spf13/viper"
)

const (
	// PortKey is the key for port configuration
	PortKey = "PORT"
	// EncryptionSecretKeyKey is the key for encryption secret key
	EncryptionSecretKeyKey = "ENCRYPTION_SECRET_KEY"
	// DBHostKey is the key for database host configuration
	DBHostKey = "DB_HOST"
	// DBUserKey is the key for database user configuration
	DBUserKey = "DB_USER"
	// DBPasswordKey is the key for database password configuration
	DBPasswordKey = "DB_PASSWORD"
	// DBNameKey is the key for database name configuration
	DBNameKey = "DB_NAME"
	// DBPortKey is the key for database port configuration
	DBPortKey = "DB_PORT"
	// DBSslModeKey is the key for database ssl mode configuration
	DBSslModeKey = "DB_SSLMODE"
	// IdentityManagerURL is the key for external identity manager url
	IdentityManagerURL = "IDENTITY_MANAGER_URL"
	// SlackAccessToken is the key for slack access token
	SlackAccessTokenKey = "SLACK_ACCESS_TOKEN"
)

// Config contains the application configuration
type Config struct {
	Port                   int
	EncryptionSecretKeyKey string

	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string
	DBSslMode  string

	SlackAccessToken string

	IdentityManagerURL string
}

// LoadConfig returns application configuration
func LoadConfig() *Config {
	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	viper.AddConfigPath("../")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error reading config: %s", err))
	}

	viper.SetDefault(PortKey, 3000)
	viper.SetDefault(DBSslModeKey, "disable")

	return &Config{
		Port:                   viper.GetInt(PortKey),
		EncryptionSecretKeyKey: viper.GetString(EncryptionSecretKeyKey),

		DBHost:     viper.GetString(DBHostKey),
		DBUser:     viper.GetString(DBUserKey),
		DBPassword: viper.GetString(DBPasswordKey),
		DBName:     viper.GetString(DBNameKey),
		DBPort:     viper.GetString(DBPortKey),
		DBSslMode:  viper.GetString(DBSslModeKey),

		SlackAccessToken: viper.GetString(SlackAccessTokenKey),

		IdentityManagerURL: viper.GetString(IdentityManagerURL),
	}
}
