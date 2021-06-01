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
	Host string `mapstructure:"host"`

	// http config
	GetManagersURL string `mapstructure:"get_managers_url"`
}

func NewClient(config *ClientConfig) (domain.IAMClient, error) {
	if config.Provider == IAMProviderShield {
		return NewShieldClient(&ShieldClientConfig{
			Host: config.Host,
		})
	} else if config.Provider == IAMProviderHTTP {
		return NewHTTPClient(&HTTPClientConfig{
			GetManagersURL: config.GetManagersURL,
		})
	}

	return nil, errors.New("invalid iam provider type")
}
