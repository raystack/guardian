package iam

import (
	"errors"

	"github.com/odpf/guardian/domain"
)

const (
	IAMProviderShield = "shield"
	IAMProviderHTTP   = "http"
)

type ClientConfig struct {
	Provider string `mapstructure:"provider"`

	// shield config
	Host string `mapstructure:"host" validate:"required_if=Provider shield"`

	// http config
	URL     string            `mapstructure:"url" validate:"required_if=Provider http"`
	Headers map[string]string `mapstructure:"headers"`
	Auth    *HTTPAuthConfig   `mapstructure:"auth" validate:"omitempty,dive"`
}

func NewClient(config *ClientConfig) (domain.IAMClient, error) {
	if config.Provider == IAMProviderShield {
		return NewShieldClient(&ShieldClientConfig{
			Host: config.Host,
		})
	} else if config.Provider == IAMProviderHTTP {
		return NewHTTPClient(&HTTPClientConfig{
			URL:     config.URL,
			Auth:    config.Auth,
			Headers: config.Headers,
		})
	}

	return nil, errors.New("invalid iam provider type")
}
