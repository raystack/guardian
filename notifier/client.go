package notifier

import (
	"errors"
	"fmt"
	"strings"

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

func parseMessage(message domain.NotificationMessage, templates domain.NotificationMessages) string {
	var result string
	switch message.Type {
	case domain.NotificationTypeAccessRevoked:
		result = templates.AccessRevoked
	case domain.NotificationTypeAppealApproved:
		result = templates.AppealApproved
	case domain.NotificationTypeAppealRejected:
		result = templates.AppealRejected
	case domain.NotificationTypeApproverNotification:
		result = templates.ApproverNotification
	case domain.NotificationTypeExpirationReminder:
		result = templates.ExpirationReminder
	}

	result = strings.Replace(result, "{{resource_name}}", message.Variables.ResourceName, -1)
	result = strings.Replace(result, "{{role}}", message.Variables.Role, -1)
	result = strings.Replace(result, "{{expiration_date}}", message.Variables.ExpirationDate.String(), -1)
	result = strings.Replace(result, "{{requestor}}", message.Variables.Requestor, -1)
	result = strings.Replace(result, "{{appeal_id}}", fmt.Sprintf("%v", message.Variables.AppealID), -1)

	return result
}
