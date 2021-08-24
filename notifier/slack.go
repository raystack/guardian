package notifier

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/odpf/guardian/domain"
)

const (
	slackHost = "https://slack.com"
)

type slackUser struct {
	ID       string `json:"id"`
	TeamID   string `json:"team_id"`
	Name     string `json:"name"`
	RealName string `json:"real_name"`
}

type slackUserResponse struct {
	OK    bool       `json:"ok"`
	User  *slackUser `json:"user"`
	Error string     `json:"error"`
}

type slackNotifier struct {
	accessToken string

	slackIDCache map[string]string
}

type SlackConfig struct {
	AccessToken string `mapstructure:"access_token"`
}

func NewSlackNotifier(config *SlackConfig) *slackNotifier {
	return &slackNotifier{config.AccessToken, map[string]string{}}
}

func (n *slackNotifier) Notify(items []domain.Notification) error {
	for _, item := range items {
		slackID, err := n.findSlackIDByEmail(item.User)
		if err != nil {
			return err
		}

		if err := n.sendMessage(slackID, item.Message); err != nil {
			return err
		}
	}

	return nil
}

func (n *slackNotifier) sendMessage(channel, text string) error {
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

func (n *slackNotifier) findSlackIDByEmail(email string) (string, error) {
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

func (n *slackNotifier) sendRequest(req *http.Request) (*slackUserResponse, error) {
	Client := &http.Client{Timeout: 10 * time.Second}
	resp, err := Client.Do(req)
	if err != nil {
		return nil, err
	}

	var result slackUserResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	if !result.OK {
		return &result, errors.New(result.Error)
	}

	return &result, nil
}
