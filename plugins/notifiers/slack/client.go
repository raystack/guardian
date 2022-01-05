package slack

import (
	"bytes"
	"encoding/json"
	"errors"
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

func (n *notifier) Notify(items []domain.Notification) error {
	for _, item := range items {
		slackID, err := n.findSlackIDByEmail(item.User)
		if err != nil {
			return err
		}

		msg, err := parseMessage(item.Message, n.Messages)
		if err != nil {
			return err
		}

		if err := n.sendMessage(slackID, msg); err != nil {
			return err
		}
	}

	return nil
}

func (n *notifier) sendMessage(channel, text string) error {
	url := slackHost + "/api/chat.postMessage"
	data, err := json.Marshal(map[string]string{
		"channel": channel,
		"text":    text,
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
