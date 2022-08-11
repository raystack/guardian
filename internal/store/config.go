package store

import "time"

// Config for database connection
type Config struct {
	Host            string        `mapstructure:"host" default:"localhost"`
	User            string        `mapstructure:"user" default:"postgres"`
	Password        string        `mapstructure:"password" default:""`
	Name            string        `mapstructure:"name" default:"postgres"`
	Port            string        `mapstructure:"port" default:"5432"`
	SslMode         string        `mapstructure:"sslmode" default:"disable"`
	MaxIdleConns    int           `yaml:"max_idle_conns" mapstructure:"max_idle_conns" default:"3"`
	MaxOpenConns    int           `yaml:"max_open_conns" mapstructure:"max_open_conns" default:"10"`
	ConnMaxLifeTime time.Duration `yaml:"conn_max_life_time" mapstructure:"conn_max_life_time" default:"10ms"`
	MaxIdleLifeTime time.Duration `yaml:"idle_max_life_time" mapstructure:"idle_max_life_time" default:"100ms"`
	LogLevel        string        `mapstructure:"log_level" default:"info"`
}
