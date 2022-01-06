package identities

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/domain"
)

var (
	ErrInvalidConfig       = errors.New("invalid client config")
	ErrUnknownProviderType = errors.New("unknown provider type")
)

type manager struct {
	crypto    domain.Crypto
	validator *validator.Validate
}

func NewManager(crypto domain.Crypto, validator *validator.Validate) *manager {
	return &manager{crypto, validator}
}

func (m *manager) ParseConfig(iamConfig *domain.IAMConfig) (domain.SensitiveConfig, error) {
	switch iamConfig.Provider {
	case domain.IAMProviderTypeHTTP:
		var clientConfig HTTPClientConfig
		if err := mapstructure.Decode(iamConfig.Config, &clientConfig); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidConfig, err)
		}
		clientConfig.crypto = m.crypto
		clientConfig.validator = m.validator
		return &clientConfig, nil
	case domain.IAMProviderTypeShield:
		var clientConfig ShieldClientConfig
		if err := mapstructure.Decode(iamConfig.Config, &clientConfig); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidConfig, err)
		}
		clientConfig.crypto = m.crypto
		clientConfig.validator = m.validator
		return &clientConfig, nil
	}
	return nil, ErrUnknownProviderType
}

func (m *manager) GetClient(config domain.SensitiveConfig) (domain.IAMClient, error) {
	if clientConfig, ok := config.(*HTTPClientConfig); ok {
		return NewHTTPClient(clientConfig)
	}
	if clientConfig, ok := config.(*ShieldClientConfig); ok {
		return NewShieldClient(clientConfig)
	}

	return nil, ErrInvalidConfig
}
