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
	GetManagersURL string          `mapstructure:"get_managers_url" validate:"required_if=Provider http"`
	Auth           *HTTPAuthConfig `mapstructure:"auth" valdiate:"omitempty,dive"`
}

func NewClient(config *ClientConfig) (domain.IAMClient, error) {
	if config.Provider == IAMProviderShield {
		return NewShieldClient(&ShieldClientConfig{
			Host: config.Host,
		})
	} else if config.Provider == IAMProviderHTTP {
		return NewHTTPClient(&HTTPClientConfig{
			GetManagersURL: config.GetManagersURL,
			Auth:           config.Auth,
		})
	}

	return nil, errors.New("invalid iam provider type")
}
