package app

import (
	"fmt"

	"github.com/spf13/viper"
)

const (
	// PortKey is the key for port configuration
	PortKey = "PORT"
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
)

// Config contains the application configuration
type Config struct {
	Port int

	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string
	DBSslMode  string
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
		Port:       viper.GetInt(PortKey),
		DBHost:     viper.GetString(DBHostKey),
		DBUser:     viper.GetString(DBUserKey),
		DBPassword: viper.GetString(DBPasswordKey),
		DBName:     viper.GetString(DBNameKey),
		DBPort:     viper.GetString(DBPortKey),
		DBSslMode:  viper.GetString(DBSslModeKey),
	}
}
