package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/odpf/guardian/domain"
)

const (
	slackHost = "https://slack.com"
)

type user struct {
	ID       string `json:"id"`
	TeamID   string `json:"team_id"`
	Name     string `json:"name"`
	RealName string `json:"real_name"`
}

type userResponse struct {
	OK    bool   `json:"ok"`
	User  *user  `json:"user"`
	Error string `json:"error"`
}

type notifier struct {
	accessToken string

	slackIDCache map[string]string
	Messages     domain.NotificationMessages
}

type Config struct {
	AccessToken string `mapstructure:"access_token"`
	Messages    domain.NotificationMessages
}

func New(config *Config) *notifier {
	return &notifier{config.AccessToken, map[string]string{}, config.Messages}
}

func (n *notifier) Notify(items []domain.Notification) []error {
	errs := make([]error, 0)
	for _, item := range items {
		slackID, err := n.findSlackIDByEmail(item.User)
		if err != nil {
			errs = append(errs, err)
		}

		msg, err := parseMessage(item.Message, n.Messages)
		if err != nil {
			errs = append(errs, err)
		}

		if err := n.sendMessage(slackID, msg); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (n *notifier) sendMessage(channel, messageBlock string) error {
	url := slackHost + "/api/chat.postMessage"
	var messageblockList []interface{}

	if err := json.Unmarshal([]byte(messageBlock), &messageblockList); err != nil {
		return err
	}
	data, err := json.Marshal(map[string]interface{}{
		"channel": channel,
		"blocks":  messageblockList,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+n.accessToken)
	req.Header.Add("Content-Type", "application/json")

	_, err = n.sendRequest(req)
	return err
}

func (n *notifier) findSlackIDByEmail(email string) (string, error) {
	if n.slackIDCache[email] != "" {
		return n.slackIDCache[email], nil
	}

	slackURL := slackHost + "/api/users.lookupByEmail"
	form := url.Values{}
	form.Add("email", email)

	req, err := http.NewRequest(http.MethodPost, slackURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+n.accessToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	result, err := n.sendRequest(req)
	if err != nil {
		return "", err
	}
	if result.User == nil {
		return "", errors.New("user not found")
	}

	n.slackIDCache[email] = result.User.ID
	return result.User.ID, nil
}

func (n *notifier) sendRequest(req *http.Request) (*userResponse, error) {
	Client := &http.Client{Timeout: 10 * time.Second}
	resp, err := Client.Do(req)
	if err != nil {
		return nil, err
	}

	var result userResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	if !result.OK {
		return &result, errors.New(result.Error)
	}

	return &result, nil
}

func parseMessage(message domain.NotificationMessage, templates domain.NotificationMessages) (string, error) {
	messageTypeTemplateMap := map[string]string{
		domain.NotificationTypeAccessRevoked:          templates.AccessRevoked,
		domain.NotificationTypeAppealApproved:         templates.AppealApproved,
		domain.NotificationTypeAppealRejected:         templates.AppealRejected,
		domain.NotificationTypeApproverNotification:   templates.ApproverNotification,
		domain.NotificationTypeExpirationReminder:     templates.ExpirationReminder,
		domain.NotificationTypeOnBehalfAppealApproved: templates.OthersAppealApproved,
	}

	block, ok := messageTypeTemplateMap[message.Type]
	if !ok {
		return "", fmt.Errorf("template not found for message type %s", message.Type)
	}

	t, err := template.New("notification_messages").Parse(block)
	if err != nil {
		return "", err
	}

	var buff bytes.Buffer
	if err := t.Execute(&buff, message.Variables); err != nil {
		return "", err
	}

	return buff.String(), nil
}
