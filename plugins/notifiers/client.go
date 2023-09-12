package notifiers

import (
	"errors"
	"net/http"
	"time"

	"github.com/goto/guardian/domain"
	"github.com/goto/guardian/plugins/notifiers/slack"
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
	AccessToken string                 `mapstructure:"access_token" validate:"required_without=Workspaces"`
	Workspaces  []slack.SlackWorkspace `mapstructure:"workspaces" validate:"required_without=AccessToken,dive"`

	// custom messages
	Messages domain.NotificationMessages
}

func NewClient(config *Config) (Client, error) {
	if config.Provider == ProviderTypeSlack {

		slackConfig, err := NewSlackConfig(config)
		if err != nil {
			return nil, err
		}

		httpClient := &http.Client{Timeout: 10 * time.Second}
		return slack.NewNotifier(slackConfig, httpClient), nil
	}

	return nil, errors.New("invalid notifier provider type")
}

func NewSlackConfig(config *Config) (*slack.Config, error) {

	// validation
	if config.AccessToken == "" && len(config.Workspaces) == 0 {
		return nil, errors.New("slack access token or workspaces must be provided")
	}
	if config.AccessToken != "" && len(config.Workspaces) != 0 {
		return nil, errors.New("slack access token and workspaces cannot be provided at the same time")
	}

	var slackConfig *slack.Config
	if config.AccessToken != "" {
		workspaces := []slack.SlackWorkspace{
			{
				WorkspaceName: "default",
				AccessToken:   config.AccessToken,
				Criteria:      "1==1",
			},
		}
		slackConfig = &slack.Config{
			Workspaces: workspaces,
			Messages:   config.Messages,
		}
		return slackConfig, nil
	}

	slackConfig = &slack.Config{
		Workspaces: config.Workspaces,
		Messages:   config.Messages,
	}

	return slackConfig, nil
}
