package slack

import (
	"bytes"
	"errors"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
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
	slackIDCache   map[string]string
	notifier       notifier
}

func (s *ClientTestSuite) setup() {
	s.mockHttpClient = new(mocks.HTTPClient)
	s.accessToken = "XXXXX-TOKEN-XXXXX"
	s.messages = domain.NotificationMessages{
		AppealRejected: "[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"Your appeal to {{.resource_name}} with role {{.role}} has been rejected\"}}]",
	}
	s.slackIDCache = map[string]string{}
	s.notifier = notifier{
		accessToken:  s.accessToken,
		slackIDCache: s.slackIDCache,
		Messages:     s.messages,
		httpClient:   s.mockHttpClient,
	}
}

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (s *ClientTestSuite) TestNotify() {

	s.Run("should return error if slack id not found", func() {
		s.setup()

		slackAPIResponse := `{"ok":false,"error":"users_not_found"}`
		resp := &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(slackAPIResponse)))}
		s.mockHttpClient.On("Do", mock.Anything).Return(resp, nil)
		expectedErrs := []error{errors.New("users_not_found"), errors.New("EOF")}

		actualErrs := s.notifier.Notify([]domain.Notification{
			{
				User: "test-user@abc.com",
				Message: domain.NotificationMessage{
					Type: domain.NotificationTypeAppealRejected,
					Variables: map[string]interface{}{
						"ResourceName": "test-resource",
					},
				},
			},
		})

		s.Equal(expectedErrs, actualErrs)
	})
}
