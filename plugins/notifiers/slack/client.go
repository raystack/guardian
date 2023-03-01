package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/odpf/guardian/utils"
	"html/template"
	"net/http"
	"net/url"
	"os"
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
	httpClient   utils.HTTPClient
}

type Config struct {
	AccessToken string `mapstructure:"access_token"`
	Messages    domain.NotificationMessages
}

func New(config *Config) *notifier {
	return &notifier{
		accessToken:  config.AccessToken,
		slackIDCache: map[string]string{},
		Messages:     config.Messages,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
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
		return fmt.Errorf("error in parsing message block %s", err)
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
	resp, err := n.httpClient.Do(req)
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

func getDefaultTemplate(messageType string) (string, error) {
	content, err := os.ReadFile(fmt.Sprintf("plugins/notifiers/slack/templates/%s.json", messageType))
	if err != nil {
		return "", fmt.Errorf("error finding default template for message type %s - %s", messageType, err)
	}
	return string(content), nil
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

	messageBlock, ok := messageTypeTemplateMap[message.Type]
	if !ok {
		return "", fmt.Errorf("template not found for message type %s", message.Type)
	}

	if messageBlock == "" {
		defaultMsgBlock, err := getDefaultTemplate(message.Type)
		if err != nil {
			return "", err
		}
		messageBlock = defaultMsgBlock
	}

	t, err := template.New("notification_messages").Parse(messageBlock)
	if err != nil {
		return "", err
	}

	var buff bytes.Buffer
	if err := t.Execute(&buff, message.Variables); err != nil {
		return "", err
	}

	return buff.String(), nil
}
