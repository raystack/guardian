package notifier

import (
	"bytes"
	"errors"
	"text/template"

	"github.com/odpf/guardian/domain"
)

const (
	ProviderTypeSlack = "slack"
)

type ClientConfig struct {
	Provider string `mapstructure:"provider"`

	// slack
	AccessToken string `mapstructure:"access_token"`

	// custom messages
	Messages domain.NotificationMessages
}

func NewClient(config *ClientConfig) (domain.Notifier, error) {
	if config.Provider == ProviderTypeSlack {
		return NewSlackNotifier(&SlackConfig{
			AccessToken: config.AccessToken,
			Messages:    config.Messages,
		}), nil
	}

	return nil, errors.New("invalid notifier provider type")
}

func parseMessage(message domain.NotificationMessage, templates domain.NotificationMessages) (string, error) {
	var text string
	switch message.Type {
	case domain.NotificationTypeAccessRevoked:
		text = templates.AccessRevoked
	case domain.NotificationTypeAppealApproved:
		text = templates.AppealApproved
	case domain.NotificationTypeAppealRejected:
		text = templates.AppealRejected
	case domain.NotificationTypeApproverNotification:
		text = templates.ApproverNotification
	case domain.NotificationTypeExpirationReminder:
		text = templates.ExpirationReminder
	}

	t, err := template.New("notification_messages").Parse(text)
	if err != nil {
		return "", err
	}

	var buff bytes.Buffer
	if err := t.Execute(&buff, message.Variables); err != nil {
		return "", err
	}

	return buff.String(), nil
}
