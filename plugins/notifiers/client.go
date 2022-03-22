package notifiers

import (
	"errors"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/plugins/notifiers/slack"
)

type Client interface {
	Notify([]domain.Notification) error
}

const (
	ProviderTypeSlack = "slack"
)

type Config struct {
	Provider string `mapstructure:"provider" validate:"omitempty,oneof=slack"`

	//console host url
	ConsoleUrl string `mapstructure:"console_url" validate:"required"`

	// slack
	AccessToken string `mapstructure:"access_token" validate:"required_if=Provider slack"`

	// custom messages
	Messages domain.NotificationMessages
}

func NewClient(config *Config) (Client, error) {
	if config.Provider == ProviderTypeSlack {
		variables := make(map[string]interface{}, 0)
		variables["console_url"] = config.ConsoleUrl
		return slack.New(&slack.Config{
			AccessToken: config.AccessToken,
			Variables:   variables,
			Messages:    config.Messages,
		}), nil
	}

	return nil, errors.New("invalid notifier provider type")
}
