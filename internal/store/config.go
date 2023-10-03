package store

// Config for database connection
type Config struct {
	Host                string `mapstructure:"host" default:"localhost"`
	User                string `mapstructure:"user" default:"postgres"`
	Password            string `mapstructure:"password" default:""`
	Name                string `mapstructure:"name" default:"postgres"`
	Port                string `mapstructure:"port" default:"5432"`
	SslMode             string `mapstructure:"sslmode" default:"disable"`
	LogLevel            string `mapstructure:"log_level" default:"info"`
	ConnMaxIdleTimeInMs int    `mapstructure:"conn_max_idle_time_in_ms"`
	ConnMaxLifetimeInMs int    `mapstructure:"conn_max_lifetime_in_ms"`
	MaxIdleConns        int    `mapstructure:"max_idle_conns"`
	MaxOpenConns        int    `mapstructure:"max_open_conns"`
}
