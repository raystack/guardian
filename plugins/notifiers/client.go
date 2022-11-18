package notifiers

import (
	"errors"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/plugins/notifiers/slack"
)

type Client interface {
	Notify([]domain.Notification) []error
	GetUserEmail(userId string) (string, error)
}

const (
	ProviderTypeSlack = "slack"
)

type Config struct {
	Provider string `mapstructure:"provider" validate:"omitempty,oneof=slack"`

	// slack
	AccessToken string `mapstructure:"access_token" validate:"required_if=Provider slack"`

	// custom messages
	Messages domain.NotificationMessages
}

func NewClient(config *Config) (Client, error) {
	if config.Provider == ProviderTypeSlack {
		return slack.New(&slack.Config{
			AccessToken: config.AccessToken,
			Messages:    config.Messages,
		}), nil
	}

	return nil, errors.New("invalid notifier provider type")
}
