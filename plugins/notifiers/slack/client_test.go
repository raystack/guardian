package slack_test

import (
	"bytes"
	"embed"
	"errors"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/plugins/notifiers/slack"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"testing"
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
	s.accessToken = "XXXXX-TOKEN-XXXXX"
	s.messages = domain.NotificationMessages{
		AppealRejected: "[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"Your appeal to {{.resource_name}} with role {{.role}} has been rejected\"}}]",
	}

	conf := &slack.Config{
		AccessToken: s.accessToken,
		Messages:    s.messages,
	}

	s.notifier = slack.NewNotifier(conf, s.mockHttpClient)
}

func (s *ClientTestSuite) TestNotify() {
	s.Run("should return error if slack id not found", func() {
		s.setup()

		slackAPIResponse := `{"ok":false,"error":"users_not_found"}`
		resp := &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(slackAPIResponse)))}
		s.mockHttpClient.On("Do", mock.Anything).Return(resp, nil)
		expectedErrs := []error{errors.New("users_not_found"), errors.New("EOF")}

		notifications := []domain.Notification{
			{
				User: "test-user@abc.com",
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
	s.Run("should be able to parse message", func() {
		notificationMsg := domain.NotificationMessage{
			Type: domain.NotificationTypeAppealRejected,
			Variables: map[string]interface{}{
				"resource_name": "test-resource",
				"role":          "test-role",
			},
		}
		expectedMsg := `[{"type":"section","text":{"type":"mrkdwn","text":"Your appeal to test-resource with role test-role has been rejected"}}]`
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

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}