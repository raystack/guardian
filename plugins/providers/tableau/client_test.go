package tableau_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/plugins/providers/tableau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestNewClient(t *testing.T) {
	t.Run("should return error if config is invalid", func(t *testing.T) {
		invalidConfig := &tableau.ClientConfig{}

		actualClient, actualError := tableau.NewClient(invalidConfig)

		assert.Nil(t, actualClient)
		assert.Error(t, actualError)
	})

	t.Run("should return error if config.Host is not a valid url", func(t *testing.T) {
		invalidHostConfig := &tableau.ClientConfig{
			Username:   "test-username",
			Password:   "test-password",
			ContentURL: "test-content-url",
			Host:       "invalid-url",
		}

		actualClient, actualError := tableau.NewClient(invalidHostConfig)

		assert.Nil(t, actualClient)
		assert.Error(t, actualError)
	})

	t.Run("should return error if got error retrieving the session token", func(t *testing.T) {
		mockHttpClient := new(mocks.HTTPClient)
		config := &tableau.ClientConfig{
			Username:   "test-username",
			Password:   "test-password",
			Host:       "http://localhost",
			ContentURL: "test-content-url",
			HTTPClient: mockHttpClient,
		}

		expectedError := errors.New("request error")
		mockHttpClient.On("Do", mock.Anything).Return(nil, expectedError).Once()
		actualClient, actualError := tableau.NewClient(config)

		mockHttpClient.AssertExpectations(t)
		assert.Nil(t, actualClient)
		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return client and nil error on success", func(t *testing.T) {
		// TODO: test http request execution
		mockHttpClient := new(mocks.HTTPClient)
		config := &tableau.ClientConfig{
			Username:   "test-username",
			Password:   "test-password",
			Host:       "http://localhost",
			ContentURL: "test-content-url",
			HTTPClient: mockHttpClient,
			APIVersion: "3.2",
		}

		sessionToken := "93df71b4-6887-46bd-b4bf-7ad3b94bd6fe"
		responseJSON := `{"id":"` + sessionToken + `"}`
		response := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(responseJSON)))}
		mockHttpClient.On("Do", mock.Anything).Return(&response, nil).Once()

		_, actualError := tableau.NewClient(config)
		mockHttpClient.AssertExpectations(t)
		assert.Nil(t, actualError)
	})
}

type ClientTestSuite struct {
	suite.Suite

	mockHttpClient *mocks.HTTPClient
	client         tableau.TableauClient
	sessionToken   string
	host           string
	apiVersion     string
	siteID         string
}

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (s *ClientTestSuite) setup() {
	s.mockHttpClient = new(mocks.HTTPClient)

	s.sessionToken = "93df71b4-6887-46bd-b4bf-7ad3b94bd6fe"
	sessionResponse := http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"` + s.sessionToken + `"}`))),
	}
	s.mockHttpClient.On("Do", mock.Anything).Return(&sessionResponse, nil).Once()
	s.apiVersion = "3.12"
	s.siteID = "1.0"
	s.host = "http://localhost"
	//siteID := "test-site"
	client, err := tableau.NewClient(&tableau.ClientConfig{
		Username:   "test-username",
		Password:   "test-password",
		Host:       s.host,
		HTTPClient: s.mockHttpClient,
		ContentURL: "test-content-url",
		APIVersion: "3.12",
	})

	s.Require().NoError(err)
	s.client = client
}

// func (s *ClientTestSuite) TestGetWorkbooks() {
// 	s.Run("should get workbooks and nil error on success", func() {
// 		s.setup()

// 		url := fmt.Sprintf("/api/%v/sites/%v/workbooks", s.apiVersion, s.siteID)

// 		testRequest, err := s.getTestRequest(http.MethodGet, url, nil)
// 		s.Require().NoError(err)

// 		workbookResponseJSON := `[{}]`
// 		workbookResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(workbookResponseJSON)))}
// 		s.mockHttpClient.On("Do", testRequest).Return(&workbookResponse, nil).Once()

// 		expectedWorkbooks := []tableau.Workbook{
// 			{},
// 		}

// 		result, err1 := s.client.GetWorkbooks()
// 		var workbooks []tableau.Workbook
// 		for _, wb := range result {
// 			workbooks = append(workbooks, *wb)
// 		}
// 		s.Nil(err1)
// 		s.ElementsMatch(expectedWorkbooks, workbooks)

// 	})
// }

// func (s *ClientTestSuite) getTestRequest(method, path string, body interface{}) (*http.Request, error) {
// 	var buf io.ReadWriter
// 	if body != nil {
// 		buf = new(bytes.Buffer)
// 		err := json.NewEncoder(buf).Encode(body)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	url := fmt.Sprintf("%s%s", s.host, path)
// 	req, err := http.NewRequest(method, url, buf)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if body != nil {
// 		req.Header.Set("Content-Type", "application/json")
// 	}
// 	req.Header.Set("Accept", "application/json")
// 	req.Header.Set("X-Tableau-Auth", s.sessionToken)
// 	return req, nil
// }

// func (s *ClientTestSuite) TestGetFlows() {
// 	s.Run("should get folders and nil error on success", func() {
// 		s.setup()

// 		path := fmt.Sprintf("/api/%v/sites/%v/flows", s.apiVersion, "1.0")

// 		testRequest, err := s.getTestRequest(http.MethodGet, path, nil)
// 		s.Require().NoError(err)

// 		expectedFlows := []tableau.Flow{
// 			{
// 				ID:   "flow-1",
// 				Name: "fl_1",
// 			},
// 			{
// 				ID:   "flow-2",
// 				Name: "fl_2",
// 			},
// 		}

// 		flowResponseJSON := `[{"id":"flow-1","name":"fl_1"},{"id":"flow-2","name":"fl_2"}]`
// 		folderResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(flowResponseJSON)))}
// 		s.mockHttpClient.On("Do", testRequest).Return(&folderResponse, nil).Once()

// 		result, err1 := s.client.GetFlows()
// 		var flows []tableau.Flow
// 		for _, fl := range result {
// 			flows = append(flows, *fl)
// 		}

// 		s.Nil(err1)
// 		s.Equal(expectedFlows, flows)
// 	})
// }
