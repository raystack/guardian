package iam

import (
	"errors"

	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/domain"
)

const (
	IAMProviderShield = "shield"
	IAMProviderHTTP   = "http"
)

func NewClient(config map[string]interface{}) (domain.IAMClient, error) {
	if config["provider"] == IAMProviderShield {
		var shieldConfig ShieldClientOptions
		if err := mapstructure.Decode(config, &shieldConfig); err != nil {
			return nil, err
		}

		return NewShieldClient(shieldConfig)
	}

	return nil, errors.New("invalid iam provider type")
}
