package notifier

import (
	"errors"

	"github.com/odpf/guardian/domain"
)

const (
	ProviderTypeSlack = "slack"
)

type ClientConfig struct {
	Provider string `mapstructure:"provider"`

	// slack
	AccessToken string `mapstructure:"access_token"`
}

func NewClient(config *ClientConfig) (domain.Notifier, error) {
	if config.Provider == ProviderTypeSlack {
		return NewSlackNotifier(&SlackConfig{
			AccessToken: config.AccessToken,
		}), nil
	}

	return nil, errors.New("invalid notifier provider type")
}
