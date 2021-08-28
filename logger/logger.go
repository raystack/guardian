package logger

import (
	"github.com/odpf/salt/log"
)

type Config struct {
	Level string `mapstructure:"level" default:"info"`
}

func New(config *Config) log.Logger {
	logger := log.NewLogrus(log.LogrusWithLevel(config.Level))
	return logger
}
