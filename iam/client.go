package iam

import (
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/domain"
)

var (
	ErrInvalidConfig       = errors.New("invalid client config")
	ErrUnknownProviderType = errors.New("unknown provider type")
)

func NewClient(config *domain.IAMConfig) (domain.IAMClient, error) {
	if config.Provider == domain.IAMProviderTypeShield {
		var clientConfig ShieldClientConfig
		if err := mapstructure.Decode(config.Config, &clientConfig); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidConfig, err)
		}
		return NewShieldClient(&clientConfig)
	} else if config.Provider == domain.IAMProviderTypeHTTP {
		var clientConfig HTTPClientConfig
		if err := mapstructure.Decode(config.Config, &clientConfig); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidConfig, err)
		}
		return NewHTTPClient(&clientConfig)
	}

	return nil, ErrUnknownProviderType
}
