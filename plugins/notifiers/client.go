package notifiers

import (
	"errors"
	"net/http"
	"time"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/plugins/notifiers/slack"
)

type Client interface {
	Notify([]domain.Notification) []error
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
		slackConfig := &slack.Config{
			AccessToken: config.AccessToken,
			Messages:    config.Messages,
		}
		httpClient := &http.Client{Timeout: 10 * time.Second}
		slackIDCache := map[string]string{}
		return slack.NewNotifier(slackConfig, slackIDCache, httpClient), nil
	}

	return nil, errors.New("invalid notifier provider type")
}
