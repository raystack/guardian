package slack_test

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/raystack/guardian/domain"
	"github.com/raystack/guardian/mocks"
	"github.com/raystack/guardian/plugins/notifiers/slack"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ClientTestSuite struct {
	suite.Suite
	mockHttpClient *mocks.HTTPClient
	accessToken    string
	messages       domain.NotificationMessages
	notifier       *slack.Notifier
}

func (s *ClientTestSuite) setup() {
	s.mockHttpClient = new(mocks.HTTPClient)
	workspaces := []slack.SlackWorkspace{
		{
			WorkspaceName: "ws-1",
			AccessToken:   "XXXXX-TOKEN-1-XXXXX",
			Criteria:      "$email contains '@abc'",
		},
		{
			WorkspaceName: "ws-2",
			AccessToken:   "XXXXX-TOKEN-2-XXXXX",
			Criteria:      "$email contains '@xyz'",
		},
	}
	s.accessToken = "XXXXX-TOKEN-XXXXX"
	s.messages = domain.NotificationMessages{
		AppealRejected: "[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"Your appeal to {{.resource_name}} with role {{.role}} has been rejected\"}}]",
	}

	conf := &slack.Config{
		Workspaces: workspaces,
		Messages:   s.messages,
	}

	s.notifier = slack.NewNotifier(conf, s.mockHttpClient)
}

func (s *ClientTestSuite) TestNotify() {
	s.Run("should return error if slack id not found", func() {
		s.setup()

		slackAPIResponse := `{"ok":false,"error":"users_not_found"}`
		resp := &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(slackAPIResponse)))}
		s.mockHttpClient.On("Do", mock.Anything).Return(resp, nil)
		expectedErrs := []error{
			fmt.Errorf("[appeal_id=test-appeal-id] | %w", errors.New("error finding slack id for email test-user@abc.com in workspace: ws-1 - users_not_found")),
		}

		notifications := []domain.Notification{
			{
				User: "test-user@abc.com",
				Labels: map[string]string{
					"appeal_id": "test-appeal-id",
				},
				Message: domain.NotificationMessage{
					Type: domain.NotificationTypeAppealRejected,
					Variables: map[string]interface{}{
						"resource_name": "test-resource",
						"role":          "test-role",
					},
				},
			},
		}
		actualErrs := s.notifier.Notify(notifications)

		s.Equal(expectedErrs, actualErrs)
	})
}

func (s *ClientTestSuite) TestParseMessage() {
	s.setup()
	s.messages = domain.NotificationMessages{
		ApproverNotification: "[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"You have an appeal created by *{{.requestor}}* requesting access to *{{.resource_name}}* with role *{{.role}}*. User's manager: {{.creator.manager_email}} and belongs to {{.creator.org_name}}.\\n Appeal ID: *{{.appeal_id}}*\"}}]",
	}
	s.Run("should be able to parse message", func() {
		creator := map[string]interface{}{
			"manager_email": "user-manager@example.com",
			"org_name":      "test-org",
		}

		notificationMsg := domain.NotificationMessage{
			Type: domain.NotificationTypeApproverNotification,
			Variables: map[string]interface{}{
				"resource_name": "test-resource",
				"role":          "test-role",
				"creator":       creator,
				"appeal_id":     "test-appeal-id",
				"requestor":     "test-user",
			},
		}
		expectedMsg := `[{"type":"section","text":{"type":"mrkdwn","text":"You have an appeal created by *test-user* requesting access to *test-resource* with role *test-role*. User's manager: user-manager@example.com and belongs to test-org.\n Appeal ID: *test-appeal-id*"}}]`
		message, err := slack.ParseMessage(notificationMsg, s.messages, embed.FS{})

		s.Nil(err)
		s.Equal(expectedMsg, message)
	})

	s.Run("should return error if message template not found", func() {
		notificationMsg := domain.NotificationMessage{
			Type: "AppealSuspended", // not found in messages
			Variables: map[string]interface{}{
				"resource_name": "test-resource",
				"role":          "test-role",
			},
		}
		_, err := slack.ParseMessage(notificationMsg, s.messages, embed.FS{})
		s.Errorf(err, "template not found for message type AppealSuspended")
	})
}

func (s *ClientTestSuite) TestGetSlackWorkspaceForUser() {
	s.setup()
	s.Run("should return slack workspace 1 for user", func() {
		userEmail := "example-user@abc.com"
		expectedWs := &slack.SlackWorkspace{
			WorkspaceName: "ws-1",
			AccessToken:   "XXXXX-TOKEN-1-XXXXX",
			Criteria:      "$email contains '@abc'",
		}
		actualWs, err := s.notifier.GetSlackWorkspaceForUser(userEmail)
		s.Nil(err)
		s.Equal(expectedWs, actualWs)
	})

	s.Run("should return slack workspace 2 for user", func() {
		userEmail := "example-user@xyz.com"
		expectedWs := &slack.SlackWorkspace{
			WorkspaceName: "ws-2",
			AccessToken:   "XXXXX-TOKEN-2-XXXXX",
			Criteria:      "$email contains '@xyz'",
		}
		actualWs, err := s.notifier.GetSlackWorkspaceForUser(userEmail)
		s.Nil(err)
		s.Equal(expectedWs, actualWs)
	})

	s.Run("should return error if slack workspace not found for user", func() {
		userEmail := "example-user@def.com"
		expectedErr := fmt.Errorf("no slack workspace found for user: %s", userEmail)
		_, actualErr := s.notifier.GetSlackWorkspaceForUser(userEmail)
		s.Equal(expectedErr, actualErr)
	})

}

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
