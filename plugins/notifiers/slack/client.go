package slack

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/goto/guardian/pkg/evaluator"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/goto/guardian/utils"

	"github.com/goto/guardian/domain"
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

type SlackWorkspace struct {
	WorkspaceName string `mapstructure:"workspace" validate:"required"`
	AccessToken   string `mapstructure:"access_token" validate:"required"`
	Criteria      string `mapstructure:"criteria" validate:"required"`
}

type Notifier struct {
	workspaces []SlackWorkspace

	slackIDCache        map[string]*slackIDCacheItem
	Messages            domain.NotificationMessages
	httpClient          utils.HTTPClient
	defaultMessageFiles embed.FS
}

type slackIDCacheItem struct {
	SlackID   string
	Workspace *SlackWorkspace
}

type Config struct {
	Workspaces []SlackWorkspace `mapstructure:"workspaces"`
	Messages   domain.NotificationMessages
}

//go:embed templates/*
var defaultTemplates embed.FS

func NewNotifier(config *Config, httpClient utils.HTTPClient) *Notifier {
	return &Notifier{
		workspaces:          config.Workspaces,
		slackIDCache:        map[string]*slackIDCacheItem{},
		Messages:            config.Messages,
		httpClient:          httpClient,
		defaultMessageFiles: defaultTemplates,
	}
}

func (n *Notifier) Notify(items []domain.Notification) []error {
	errs := make([]error, 0)
	for _, item := range items {
		var ws *SlackWorkspace
		var slackID string
		labelSlice := utils.MapToSlice(item.Labels)

		// check cache
		if n.slackIDCache[item.User] != nil {
			slackID = n.slackIDCache[item.User].SlackID
			ws = n.slackIDCache[item.User].Workspace
		} else {
			ws, err := n.GetSlackWorkspaceForUser(item.User)
			if err != nil {
				errs = append(errs, fmt.Errorf("%v | %w", labelSlice, err))
				continue
			}
			slackID, err = n.findSlackIDByEmail(item.User, *ws)
			if err != nil {
				errs = append(errs, fmt.Errorf("%v | %w", labelSlice, err))
				continue
			}

			// cache
			n.slackIDCache[item.User] = &slackIDCacheItem{
				SlackID:   slackID,
				Workspace: ws,
			}
		}

		msg, err := ParseMessage(item.Message, n.Messages, n.defaultMessageFiles)
		if err != nil {
			errs = append(errs, fmt.Errorf("%v | error parsing message : %w", labelSlice, err))
			continue
		}

		if err := n.sendMessage(*ws, slackID, msg); err != nil {
			errs = append(errs, fmt.Errorf("%v | error sending message to user:%s in workspace:%s | %w", labelSlice, item.User, ws.WorkspaceName, err))
			continue
		}
	}

	return errs
}

func (n *Notifier) sendMessage(workspace SlackWorkspace, channel, messageBlock string) error {
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
	req.Header.Add("Authorization", "Bearer "+workspace.AccessToken)
	req.Header.Add("Content-Type", "application/json")

	_, err = n.sendRequest(req)
	return err
}

func (n *Notifier) GetSlackWorkspaceForUser(email string) (*SlackWorkspace, error) {
	var ws *SlackWorkspace
	for _, workspace := range n.workspaces {
		v, err := evaluator.Expression(workspace.Criteria).EvaluateWithVars(map[string]interface{}{
			"email": email,
		})
		if err != nil {
			return ws, fmt.Errorf("error evaluating notifier expression: %w", err)
		}

		// if the expression evaluates to true, return the workspace
		if match, ok := v.(bool); !ok {
			return ws, errors.New("notifier expression did not evaluate to a boolean")
		} else if match {
			ws = &workspace
			break
		}
	}

	if ws == nil {
		return ws, errors.New(fmt.Sprintf("no slack workspace found for user: %s", email))
	}

	return ws, nil
}

func (n *Notifier) findSlackIDByEmail(email string, ws SlackWorkspace) (string, error) {
	slackURL := slackHost + "/api/users.lookupByEmail"
	form := url.Values{}
	form.Add("email", email)

	req, err := http.NewRequest(http.MethodPost, slackURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+ws.AccessToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	result, err := n.sendRequest(req)
	if err != nil {
		return "", fmt.Errorf("error finding slack id for email %s in workspace: %s - %s", email, ws.WorkspaceName, err)
	}
	if result.User == nil {
		return "", errors.New(fmt.Sprintf("user not found: %s in workspace: %s", email, ws.WorkspaceName))
	}

	return result.User.ID, nil
}

func (n *Notifier) sendRequest(req *http.Request) (*userResponse, error) {
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

func getDefaultTemplate(messageType string, defaultTemplateFiles embed.FS) (string, error) {
	content, err := defaultTemplateFiles.ReadFile(fmt.Sprintf("templates/%s.json", messageType))
	if err != nil {
		return "", fmt.Errorf("error finding default template for message type %s - %s", messageType, err)
	}
	return string(content), nil
}

func ParseMessage(message domain.NotificationMessage, templates domain.NotificationMessages, defaultTemplateFiles embed.FS) (string, error) {
	messageTypeTemplateMap := map[string]string{
		domain.NotificationTypeAccessRevoked:          templates.AccessRevoked,
		domain.NotificationTypeAppealApproved:         templates.AppealApproved,
		domain.NotificationTypeAppealRejected:         templates.AppealRejected,
		domain.NotificationTypeApproverNotification:   templates.ApproverNotification,
		domain.NotificationTypeExpirationReminder:     templates.ExpirationReminder,
		domain.NotificationTypeOnBehalfAppealApproved: templates.OthersAppealApproved,
		domain.NotificationTypeGrantOwnerChanged:      templates.GrantOwnerChanged,
	}

	messageBlock, ok := messageTypeTemplateMap[message.Type]
	if !ok {
		return "", fmt.Errorf("template not found for message type %s", message.Type)
	}

	if messageBlock == "" {
		defaultMsgBlock, err := getDefaultTemplate(message.Type, defaultTemplateFiles)
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
