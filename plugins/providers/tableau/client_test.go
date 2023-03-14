package tableau_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/goto/guardian/mocks"
	"github.com/goto/guardian/plugins/providers/tableau"
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
	userID         string
}

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (s *ClientTestSuite) setup() {
	s.mockHttpClient = new(mocks.HTTPClient)

	s.sessionToken = "93df71b4-6887-46bd-b4bf-7ad3b94bd6fe"
	s.siteID = "1.0"
	s.userID = "test-user-id"
	sessionResponseJSON := fmt.Sprintf(`{"credentials":{"token":"%s","site":{"id":"%s"},"user":{"id":"%s"}}}`, s.sessionToken, s.siteID, s.userID)
	sessionResponse := http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(sessionResponseJSON))),
	}
	s.mockHttpClient.On("Do", mock.Anything).Return(&sessionResponse, nil).Once()
	s.apiVersion = "3.12"
	s.host = "http://localhost"
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

func (s *ClientTestSuite) TestGetWorkbooks() {
	s.Run("should return error if status code is 403, user forbidden to get workbooks", func() {
		s.setup()

		url := fmt.Sprintf("/api/%v/sites/%v/workbooks", s.apiVersion, s.siteID)
		testRequest, err := s.getTestRequest(http.MethodGet, url, nil)
		s.Require().NoError(err)

		workbookResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"workbooks":{"workbook":[{"id":"test-workbook"}]}}`
		workbookResponse := http.Response{StatusCode: 403, Body: ioutil.NopCloser(bytes.NewReader([]byte(workbookResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&workbookResponse, nil).Once()

		actualWorkbooks, err1 := s.client.GetWorkbooks()

		s.Nil(actualWorkbooks)
		s.Error(err1)
	})

	s.Run("should get workbooks and nil error on success", func() {
		s.setup()

		url := fmt.Sprintf("/api/%v/sites/%v/workbooks", s.apiVersion, s.siteID)
		testRequest, err := s.getTestRequest(http.MethodGet, url, nil)
		s.Require().NoError(err)

		workbookResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"workbooks":{"workbook":[{"id":"test-workbook"}]}}`
		workbookResponse := http.Response{StatusCode: 400, Body: ioutil.NopCloser(bytes.NewReader([]byte(workbookResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&workbookResponse, nil).Once()

		expectedWorkbooks := []*tableau.Workbook{
			{
				ID: "test-workbook",
			},
		}

		actualWorkbooks, err1 := s.client.GetWorkbooks()

		s.Nil(err1)
		s.Equal(expectedWorkbooks, actualWorkbooks)
	})
}

func (s *ClientTestSuite) getTestRequest(method, path string, body interface{}) (*http.Request, error) {
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}
	url := fmt.Sprintf("%s%s", s.host, path)
	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Tableau-Auth", s.sessionToken)
	return req, nil
}

func (s *ClientTestSuite) TestGetFlows() {
	s.Run("should get error if user is forbidden, Status Code 403", func() {
		s.setup()

		path := fmt.Sprintf("/api/%v/sites/%v/flows", s.apiVersion, "1.0")

		testRequest, err := s.getTestRequest(http.MethodGet, path, nil)
		s.Require().NoError(err)

		flowResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"flows": {"flow":[{"id":"flow-1","name":"fl_1"}]}} `

		folderResponse := http.Response{StatusCode: 403, Body: ioutil.NopCloser(bytes.NewReader([]byte(flowResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&folderResponse, nil).Once()

		actualFlows, err1 := s.client.GetFlows()

		s.Nil(actualFlows)
		s.Error(err1)
	})

	s.Run("should get flows and nil error on success, after status code 401, get the session and retry the http request", func() {
		s.setup()

		path := fmt.Sprintf("/api/%v/sites/%v/flows", s.apiVersion, "1.0")

		testRequest, err := s.getTestRequest(http.MethodGet, path, nil)
		s.Require().NoError(err)

		expectedFlows := []*tableau.Flow{
			{
				ID:   "flow-1",
				Name: "fl_1",
			},
		}

		flowResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"flows": {"flow":[{"id":"flow-1","name":"fl_1"}]}} `

		folderResponse := http.Response{StatusCode: 401, Body: ioutil.NopCloser(bytes.NewReader([]byte(flowResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&folderResponse, nil).Once()

		sessionResponseJSON := fmt.Sprintf(`{"credentials":{"token":"%s","site":{"id":"%s"},"user":{"id":"%s"}}}`, s.sessionToken, s.siteID, s.userID)
		sessionResponse := http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(sessionResponseJSON))),
		}

		s.mockHttpClient.On("Do", mock.Anything).Return(&sessionResponse, nil).Once()

		folderResponse2 := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(flowResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&folderResponse2, nil).Once()

		actualFlows, err1 := s.client.GetFlows()

		s.Nil(err1)
		s.Equal(expectedFlows, actualFlows)
	})

	s.Run("should get flows and nil error on success", func() {
		s.setup()

		path := fmt.Sprintf("/api/%v/sites/%v/flows", s.apiVersion, "1.0")

		testRequest, err := s.getTestRequest(http.MethodGet, path, nil)
		s.Require().NoError(err)

		expectedFlows := []*tableau.Flow{
			{
				ID:   "flow-1",
				Name: "fl_1",
			},
		}

		flowResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"flows": {"flow":[{"id":"flow-1","name":"fl_1"}]}} `

		folderResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(flowResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&folderResponse, nil).Once()

		actualFlows, err1 := s.client.GetFlows()

		s.Nil(err1)
		s.Equal(expectedFlows, actualFlows)
	})
}

func (s *ClientTestSuite) TestGetDataSources() {
	s.Run("should return error if status code is 403, user forbidden to get workbooks", func() {
		s.setup()

		path := fmt.Sprintf("/api/%v/sites/%v/datasources", s.apiVersion, "1.0")

		testRequest, err := s.getTestRequest(http.MethodGet, path, nil)
		s.Require().NoError(err)

		DataSourcesJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"datasources": {"datasource":[{"id":"datasource-1","name":"ds_1"}]}} `
		DataSourcesResponse := http.Response{StatusCode: 403, Body: ioutil.NopCloser(bytes.NewReader([]byte(DataSourcesJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&DataSourcesResponse, nil).Once()

		actualDataSources, err1 := s.client.GetDataSources()

		s.Nil(actualDataSources)
		s.Error(err1)
	})

	s.Run("should get datasources and nil error on success", func() {
		s.setup()

		path := fmt.Sprintf("/api/%v/sites/%v/datasources", s.apiVersion, "1.0")

		testRequest, err := s.getTestRequest(http.MethodGet, path, nil)
		s.Require().NoError(err)

		expectedDataSources := []*tableau.DataSource{
			{
				ID:   "datasource-1",
				Name: "ds_1",
			},
		}

		DataSourcesJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"datasources": {"datasource":[{"id":"datasource-1","name":"ds_1"}]}} `
		DataSourcesResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(DataSourcesJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&DataSourcesResponse, nil).Once()

		actualDataSources, err1 := s.client.GetDataSources()

		s.Nil(err1)
		s.Equal(expectedDataSources, actualDataSources)
	})
}

func (s *ClientTestSuite) TestGetViews() {
	s.Run("should return error if status code is 403, user forbidden to get resource", func() {
		s.setup()

		path := fmt.Sprintf("/api/%v/sites/%v/views", s.apiVersion, "1.0")

		testRequest, err := s.getTestRequest(http.MethodGet, path, nil)
		s.Require().NoError(err)

		ViewsResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"views": {"view":[{"id":"view-1","name":"vw_1"}]}} `

		ViewsResponse := http.Response{StatusCode: 403, Body: ioutil.NopCloser(bytes.NewReader([]byte(ViewsResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&ViewsResponse, nil).Once()

		actualViews, err1 := s.client.GetViews()

		s.Nil(actualViews)
		s.Error(err1)
	})

	s.Run("should get views and nil error on success", func() {
		s.setup()

		path := fmt.Sprintf("/api/%v/sites/%v/views", s.apiVersion, "1.0")

		testRequest, err := s.getTestRequest(http.MethodGet, path, nil)
		s.Require().NoError(err)

		expectedViews := []*tableau.View{
			{
				ID:   "view-1",
				Name: "vw_1",
			},
		}

		ViewsResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"views": {"view":[{"id":"view-1","name":"vw_1"}]}} `

		ViewsResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(ViewsResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&ViewsResponse, nil).Once()

		actualViews, err1 := s.client.GetViews()

		s.Nil(err1)
		s.Equal(expectedViews, actualViews)
	})
}

func (s *ClientTestSuite) TestGetMetrics() {
	s.Run("should return error if status code is 403, user forbidden to get resource", func() {
		s.setup()

		path := fmt.Sprintf("/api/%v/sites/%v/metrics", s.apiVersion, "1.0")

		testRequest, err := s.getTestRequest(http.MethodGet, path, nil)
		s.Require().NoError(err)
		MetricsResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"metrics": {"metric":[{"id": "metric-1","name":"mt_1"}]}} `

		MetricsResponse := http.Response{StatusCode: 403, Body: ioutil.NopCloser(bytes.NewReader([]byte(MetricsResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&MetricsResponse, nil).Once()

		actualMetrics, err1 := s.client.GetMetrics()

		s.Nil(actualMetrics)
		s.Error(err1)
	})

	s.Run("should get metrics and nil error on success", func() {
		s.setup()

		path := fmt.Sprintf("/api/%v/sites/%v/metrics", s.apiVersion, "1.0")

		testRequest, err := s.getTestRequest(http.MethodGet, path, nil)
		s.Require().NoError(err)

		expectedMetrics := []*tableau.Metric{
			{
				ID:   "metric-1",
				Name: "mt_1",
			},
		}

		MetricsResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"metrics": {"metric":[{"id": "metric-1","name":"mt_1"}]}} `

		MetricsResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(MetricsResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&MetricsResponse, nil).Once()

		actualMetrics, err1 := s.client.GetMetrics()

		s.Nil(err1)
		s.Equal(expectedMetrics, actualMetrics)
		s.mockHttpClient.AssertExpectations(s.T())
	})
}

func (s *ClientTestSuite) TestUpdateSiteRole() {
	s.Run("should return error if error in getting the user", func() {
		s.setup()

		expectedError := tableau.ErrUserNotFound
		userEmail := "test-email@gojek.com"
		filter := fmt.Sprintf("name:eq:%v", userEmail)
		path := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", s.apiVersion, s.siteID, filter)

		GetUserRequest, err := s.getTestRequest(http.MethodGet, path, nil) //testing getUser()
		s.Require().NoError(err)

		GetUserResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"users": {"user":[]}} `

		GetUserResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(GetUserResponseJSON)))}
		s.mockHttpClient.On("Do", GetUserRequest).Return(&GetUserResponse, nil).Once()

		role := "Viewer"

		actualError := s.client.UpdateSiteRole(userEmail, role)

		s.Equal(expectedError, actualError)
	})

	s.Run("should Update the site role on success and return no error", func() {
		s.setup()

		userEmail := "test-email@gojek.com"
		filter := fmt.Sprintf("name:eq:%v", userEmail)
		path := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", s.apiVersion, s.siteID, filter)

		GetUserRequest, err := s.getTestRequest(http.MethodGet, path, nil) //testing getUser()
		s.Require().NoError(err)

		GetUserResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"users": {"user":[{"id": "userID-1"}]}} `

		GetUserResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(GetUserResponseJSON)))}
		s.mockHttpClient.On("Do", GetUserRequest).Return(&GetUserResponse, nil).Once()

		response := http.Response{StatusCode: 200, Body: ioutil.NopCloser(nil)}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&response, nil).Once()
		role := "Viewer"

		actualError := s.client.UpdateSiteRole(userEmail, role)

		s.Nil(actualError)
		s.mockHttpClient.AssertExpectations(s.T())
	})
}

func (s *ClientTestSuite) TestGrantWorkbookAccess() { //the body have to be updated later after fix getTestRequest
	s.Run("should return error if error in getting the user", func() {
		s.setup()

		expectedError := tableau.ErrUserNotFound
		userEmail := "test-email@gojek.com"
		filter := fmt.Sprintf("name:eq:%v", userEmail)
		path := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", s.apiVersion, s.siteID, filter)

		GetUserRequest, err := s.getTestRequest(http.MethodGet, path, nil) //testing getUser()
		s.Require().NoError(err)

		GetUserResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"users": {"user":[]}} `

		GetUserResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(GetUserResponseJSON)))}
		s.mockHttpClient.On("Do", GetUserRequest).Return(&GetUserResponse, nil).Once()

		role := "write:allow"
		resource := &tableau.Workbook{
			ID: "wb_1",
		}
		actualError := s.client.GrantWorkbookAccess(resource, userEmail, role)

		s.Equal(expectedError, actualError)
	})

	s.Run("should grant access to the workbook and return no error on success", func() {
		s.setup()

		userEmail := "test-email@gojek.com"
		filter := fmt.Sprintf("name:eq:%v", userEmail) //test getUser
		path := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", s.apiVersion, s.siteID, filter)

		GetUserRequest, err := s.getTestRequest(http.MethodGet, path, nil)
		s.Require().NoError(err)

		GetUserResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"users": {"user":[{"id": "userID-1"}]}} `

		GetUserResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(GetUserResponseJSON)))}
		s.mockHttpClient.On("Do", GetUserRequest).Return(&GetUserResponse, nil).Once()

		response := http.Response{StatusCode: 200, Body: ioutil.NopCloser(nil)}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&response, nil).Once()

		role := "write:allow"
		resource := &tableau.Workbook{
			ID: "wb_1",
		}
		actualError := s.client.GrantWorkbookAccess(resource, userEmail, role)

		s.Nil(actualError)
		s.mockHttpClient.AssertExpectations(s.T())
	})
}

func (s *ClientTestSuite) TestGrantFlowAccess() { //the body have to be updated later after fix getTestRequest
	s.Run("should return error if error in getting the user", func() {
		s.setup()

		expectedError := tableau.ErrUserNotFound
		userEmail := "test-email@gojek.com"
		filter := fmt.Sprintf("name:eq:%v", userEmail)
		path := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", s.apiVersion, s.siteID, filter)

		GetUserRequest, err := s.getTestRequest(http.MethodGet, path, nil) //testing getUser()
		s.Require().NoError(err)

		GetUserResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"users": {"user":[]}} `

		GetUserResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(GetUserResponseJSON)))}
		s.mockHttpClient.On("Do", GetUserRequest).Return(&GetUserResponse, nil).Once()

		role := "write:allow"
		resource := &tableau.Flow{
			ID: "fl_1",
		}
		actualError := s.client.GrantFlowAccess(resource, userEmail, role)

		s.Equal(expectedError, actualError)
	})

	s.Run("should grant access to flow and return no error on success", func() {
		s.setup()

		userEmail := "test-email@gojek.com"
		filter := fmt.Sprintf("name:eq:%v", userEmail) //test getUser
		path := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", s.apiVersion, s.siteID, filter)

		GetUserRequest, err := s.getTestRequest(http.MethodGet, path, nil)
		s.Require().NoError(err)

		GetUserResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"users": {"user":[{"id": "userID-1"}]}} `

		GetUserResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(GetUserResponseJSON)))}
		s.mockHttpClient.On("Do", GetUserRequest).Return(&GetUserResponse, nil).Once()

		response := http.Response{StatusCode: 200, Body: ioutil.NopCloser(nil)}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&response, nil).Once()

		role := "write:allow"
		resource := &tableau.Flow{
			ID: "fl_1",
		}
		actualError := s.client.GrantFlowAccess(resource, userEmail, role)

		s.Nil(actualError)
		s.mockHttpClient.AssertExpectations(s.T())
	})
}

func (s *ClientTestSuite) TestGrantMetricAccess() { //the body have to be updated later after fix getTestRequest
	s.Run("should return error if error in getting the user", func() {
		s.setup()

		expectedError := tableau.ErrUserNotFound
		userEmail := "test-email@gojek.com"
		filter := fmt.Sprintf("name:eq:%v", userEmail)
		path := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", s.apiVersion, s.siteID, filter)

		GetUserRequest, err := s.getTestRequest(http.MethodGet, path, nil) //testing getUser()
		s.Require().NoError(err)

		GetUserResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"users": {"user":[]}} `

		GetUserResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(GetUserResponseJSON)))}
		s.mockHttpClient.On("Do", GetUserRequest).Return(&GetUserResponse, nil).Once()

		role := "write:allow"
		resource := &tableau.Metric{
			ID: "mt_1",
		}
		actualError := s.client.GrantMetricAccess(resource, userEmail, role)

		s.Equal(expectedError, actualError)
	})
	s.Run("should grant access to metric and return no error on success", func() {
		s.setup()

		userEmail := "test-email@gojek.com"
		filter := fmt.Sprintf("name:eq:%v", userEmail) //test getUser
		path := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", s.apiVersion, s.siteID, filter)

		GetUserRequest, err := s.getTestRequest(http.MethodGet, path, nil)
		s.Require().NoError(err)

		GetUserResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"users": {"user":[{"id": "userID-1"}]}} `

		GetUserResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(GetUserResponseJSON)))}
		s.mockHttpClient.On("Do", GetUserRequest).Return(&GetUserResponse, nil).Once()

		response := http.Response{StatusCode: 200, Body: ioutil.NopCloser(nil)}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&response, nil).Once()

		role := "write:allow"
		resource := &tableau.Metric{
			ID: "mt_1",
		}
		actualError := s.client.GrantMetricAccess(resource, userEmail, role)

		s.Nil(actualError)
		s.mockHttpClient.AssertExpectations(s.T())
	})
}

func (s *ClientTestSuite) TestGrantDataSourceAccess() { //the body have to be updated later after fix getTestRequest
	s.Run("should return error if error in getting the user", func() {
		s.setup()

		expectedError := tableau.ErrUserNotFound
		userEmail := "test-email@gojek.com"
		filter := fmt.Sprintf("name:eq:%v", userEmail)
		path := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", s.apiVersion, s.siteID, filter)

		GetUserRequest, err := s.getTestRequest(http.MethodGet, path, nil) //testing getUser()
		s.Require().NoError(err)

		GetUserResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"users": {"user":[]}} `

		GetUserResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(GetUserResponseJSON)))}
		s.mockHttpClient.On("Do", GetUserRequest).Return(&GetUserResponse, nil).Once()

		role := "write:allow"
		resource := &tableau.DataSource{
			ID: "ds_1",
		}
		actualError := s.client.GrantDataSourceAccess(resource, userEmail, role)

		s.Equal(expectedError, actualError)
	})
	s.Run("should grant access to datasource and return nil error on success", func() {
		s.setup()

		userEmail := "test-email@gojek.com"
		filter := fmt.Sprintf("name:eq:%v", userEmail) //test getUser
		path := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", s.apiVersion, s.siteID, filter)

		GetUserRequest, err := s.getTestRequest(http.MethodGet, path, nil)
		s.Require().NoError(err)

		GetUserResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"users": {"user":[{"id": "userID-1"}]}} `

		GetUserResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(GetUserResponseJSON)))}
		s.mockHttpClient.On("Do", GetUserRequest).Return(&GetUserResponse, nil).Once()

		//body:=
		//request:=
		response := http.Response{StatusCode: 200, Body: ioutil.NopCloser(nil)}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&response, nil).Once()

		role := "write:allow"
		resource := &tableau.DataSource{
			ID: "ds_1",
		}
		actualError := s.client.GrantDataSourceAccess(resource, userEmail, role)

		s.Nil(actualError)
	})
}

func (s *ClientTestSuite) TestGrantViewAccess() { //the body have to be updated later after fix getTestRequest
	s.Run("should return error if error in getting the user", func() {
		s.setup()

		expectedError := tableau.ErrUserNotFound
		userEmail := "test-email@gojek.com"
		filter := fmt.Sprintf("name:eq:%v", userEmail)
		path := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", s.apiVersion, s.siteID, filter)

		GetUserRequest, err := s.getTestRequest(http.MethodGet, path, nil) //testing getUser()
		s.Require().NoError(err)

		GetUserResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"users": {"user":[]}} `

		GetUserResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(GetUserResponseJSON)))}
		s.mockHttpClient.On("Do", GetUserRequest).Return(&GetUserResponse, nil).Once()

		role := "write:allow"
		resource := &tableau.View{
			ID: "vw_1",
		}
		actualError := s.client.GrantViewAccess(resource, userEmail, role)

		s.Equal(expectedError, actualError)
	})

	s.Run("should grant access to view and return nil error on success", func() {
		s.setup()

		userEmail := "test-email@gojek.com"
		filter := fmt.Sprintf("name:eq:%v", userEmail) //test getUser
		path := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", s.apiVersion, s.siteID, filter)

		GetUserRequest, err := s.getTestRequest(http.MethodGet, path, nil)
		s.Require().NoError(err)

		GetUserResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"users": {"user":[{"id": "userID-1"}]}} `

		GetUserResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(GetUserResponseJSON)))}
		s.mockHttpClient.On("Do", GetUserRequest).Return(&GetUserResponse, nil).Once()

		//body:=
		//request:=
		response := http.Response{StatusCode: 200, Body: ioutil.NopCloser(nil)}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&response, nil).Once()

		role := "write:allow"
		resource := &tableau.View{
			ID: "vw_1",
		}
		actualError := s.client.GrantViewAccess(resource, userEmail, role)

		s.Nil(actualError)
	})
}

func (s *ClientTestSuite) TestRevokeWorkbookAccess() {
	s.Run("should return error if error in getting the user", func() {
		s.setup()

		expectedError := tableau.ErrUserNotFound
		userEmail := "test-email@gojek.com"
		filter := fmt.Sprintf("name:eq:%v", userEmail)
		path := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", s.apiVersion, s.siteID, filter)

		GetUserRequest, err := s.getTestRequest(http.MethodGet, path, nil) //testing getUser()
		s.Require().NoError(err)

		GetUserResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"users": {"user":[]}} `

		GetUserResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(GetUserResponseJSON)))}
		s.mockHttpClient.On("Do", GetUserRequest).Return(&GetUserResponse, nil).Once()

		role := "write:allow"
		resource := &tableau.Workbook{
			ID: "wb_1",
		}
		actualError := s.client.RevokeWorkbookAccess(resource, userEmail, role)

		s.Equal(expectedError, actualError)
	})
	s.Run("should revoke access to workbook and return no error on success", func() {
		s.setup()

		userEmail := "test-email@gojek.com"
		filter := fmt.Sprintf("name:eq:%v", userEmail) //test getUser
		path := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", s.apiVersion, s.siteID, filter)

		GetUserRequest, err := s.getTestRequest(http.MethodGet, path, nil)
		s.Require().NoError(err)

		GetUserResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"users": {"user":[{"id": "userID-1"}]}} `

		GetUserResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(GetUserResponseJSON)))}
		s.mockHttpClient.On("Do", GetUserRequest).Return(&GetUserResponse, nil).Once()

		userID := "userID-1"
		role := "write:allow"
		split := strings.Split(role, ":")
		capabilityName := split[0]
		capabilityMode := split[1]
		resource := &tableau.Workbook{
			ID: "wb_1",
		}
		deleteWbPath := fmt.Sprintf("/api/%v/sites/%v/workbooks/%v/permissions/users/%v/%v/%v", s.apiVersion, s.siteID, resource.ID, userID, capabilityName, capabilityMode)

		deleteWbPermissionRequest, err := s.getTestRequest(http.MethodDelete, deleteWbPath, nil)
		s.Require().NoError(err)

		deleteWbPermissionResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(nil)}
		s.mockHttpClient.On("Do", deleteWbPermissionRequest).Return(&deleteWbPermissionResponse, nil).Once()

		actualError := s.client.RevokeWorkbookAccess(resource, userEmail, role)

		s.Nil(actualError)
	})
}

func (s *ClientTestSuite) TestRevokeFlowAccess() {
	s.Run("should return error if error in getting the user", func() {
		s.setup()

		expectedError := tableau.ErrUserNotFound
		userEmail := "test-email@gojek.com"
		filter := fmt.Sprintf("name:eq:%v", userEmail)
		path := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", s.apiVersion, s.siteID, filter)

		GetUserRequest, err := s.getTestRequest(http.MethodGet, path, nil) //testing getUser()
		s.Require().NoError(err)

		GetUserResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"users": {"user":[]}} `

		GetUserResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(GetUserResponseJSON)))}
		s.mockHttpClient.On("Do", GetUserRequest).Return(&GetUserResponse, nil).Once()

		role := "write:allow"
		resource := &tableau.Flow{
			ID: "fl_1",
		}
		actualError := s.client.RevokeFlowAccess(resource, userEmail, role)

		s.Equal(expectedError, actualError)
	})

	s.Run("should revoke access to flow and return nil error on success", func() {
		s.setup()

		userEmail := "test-email@gojek.com"
		filter := fmt.Sprintf("name:eq:%v", userEmail) //test getUser
		path := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", s.apiVersion, s.siteID, filter)

		GetUserRequest, err := s.getTestRequest(http.MethodGet, path, nil)
		s.Require().NoError(err)

		GetUserResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"users": {"user":[{"id": "userID-1"}]}} `

		GetUserResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(GetUserResponseJSON)))}
		s.mockHttpClient.On("Do", GetUserRequest).Return(&GetUserResponse, nil).Once()

		userID := "userID-1"
		role := "write:allow"
		split := strings.Split(role, ":")
		capabilityName := split[0]
		capabilityMode := split[1]
		resource := &tableau.Flow{
			ID: "fl_1",
		}
		deleteFlowPath := fmt.Sprintf("/api/%v/sites/%v/flows/%v/permissions/users/%v/%v/%v", s.apiVersion, s.siteID, resource.ID, userID, capabilityName, capabilityMode)

		deleteFlowPermissionRequest, err := s.getTestRequest(http.MethodDelete, deleteFlowPath, nil)
		s.Require().NoError(err)

		deleteFlowPermissionResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(nil)}
		s.mockHttpClient.On("Do", deleteFlowPermissionRequest).Return(&deleteFlowPermissionResponse, nil).Once()

		actualError := s.client.RevokeFlowAccess(resource, userEmail, role)

		s.Nil(actualError)
	})
}

func (s *ClientTestSuite) TestRevokeMetricAccess() {
	s.Run("should return error if error in getting the user", func() {
		s.setup()

		expectedError := tableau.ErrUserNotFound
		userEmail := "test-email@gojek.com"
		filter := fmt.Sprintf("name:eq:%v", userEmail)
		path := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", s.apiVersion, s.siteID, filter)

		GetUserRequest, err := s.getTestRequest(http.MethodGet, path, nil) //testing getUser()
		s.Require().NoError(err)

		GetUserResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"users": {"user":[]}} `

		GetUserResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(GetUserResponseJSON)))}
		s.mockHttpClient.On("Do", GetUserRequest).Return(&GetUserResponse, nil).Once()

		role := "write:allow"
		resource := &tableau.Metric{
			ID: "mt_1",
		}
		actualError := s.client.RevokeMetricAccess(resource, userEmail, role)

		s.Equal(expectedError, actualError)
	})
	s.Run("should revoke access to metric and return nil error on success", func() {
		s.setup()

		userEmail := "test-email@gojek.com"
		filter := fmt.Sprintf("name:eq:%v", userEmail) //test getUser
		path := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", s.apiVersion, s.siteID, filter)

		GetUserRequest, err := s.getTestRequest(http.MethodGet, path, nil)
		s.Require().NoError(err)

		GetUserResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"users": {"user":[{"id": "userID-1"}]}} `

		GetUserResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(GetUserResponseJSON)))}
		s.mockHttpClient.On("Do", GetUserRequest).Return(&GetUserResponse, nil).Once()

		userID := "userID-1"
		role := "write:allow"
		split := strings.Split(role, ":")
		capabilityName := split[0]
		capabilityMode := split[1]
		resource := &tableau.Metric{
			ID: "mt_1",
		}
		deleteMetricPath := fmt.Sprintf("/api/%v/sites/%v/metrics/%v/permissions/users/%v/%v/%v", s.apiVersion, s.siteID, resource.ID, userID, capabilityName, capabilityMode)

		deleteMetricPermissionRequest, err := s.getTestRequest(http.MethodDelete, deleteMetricPath, nil)
		s.Require().NoError(err)

		deleteMetricPermissionResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(nil)}
		s.mockHttpClient.On("Do", deleteMetricPermissionRequest).Return(&deleteMetricPermissionResponse, nil).Once()

		actualError := s.client.RevokeMetricAccess(resource, userEmail, role)

		s.Nil(actualError)
	})
}

func (s *ClientTestSuite) TestRevokeDataSourceAccess() {
	s.Run("should return error if error in getting the user", func() {
		s.setup()

		expectedError := tableau.ErrUserNotFound
		userEmail := "test-email@gojek.com"
		filter := fmt.Sprintf("name:eq:%v", userEmail)
		path := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", s.apiVersion, s.siteID, filter)

		GetUserRequest, err := s.getTestRequest(http.MethodGet, path, nil) //testing getUser()
		s.Require().NoError(err)

		GetUserResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"users": {"user":[]}} `

		GetUserResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(GetUserResponseJSON)))}
		s.mockHttpClient.On("Do", GetUserRequest).Return(&GetUserResponse, nil).Once()

		role := "write:allow"
		resource := &tableau.DataSource{
			ID: "ds_1",
		}
		actualError := s.client.RevokeDataSourceAccess(resource, userEmail, role)

		s.Equal(expectedError, actualError)
	})
	s.Run("should revoke access to datasource and return nil error on success", func() {
		s.setup()

		userEmail := "test-email@gojek.com"
		filter := fmt.Sprintf("name:eq:%v", userEmail) //test getUser
		path := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", s.apiVersion, s.siteID, filter)

		GetUserRequest, err := s.getTestRequest(http.MethodGet, path, nil)
		s.Require().NoError(err)

		GetUserResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"users": {"user":[{"id": "userID-1"}]}} `

		GetUserResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(GetUserResponseJSON)))}
		s.mockHttpClient.On("Do", GetUserRequest).Return(&GetUserResponse, nil).Once()

		userID := "userID-1"
		role := "write:allow"
		split := strings.Split(role, ":")
		capabilityName := split[0]
		capabilityMode := split[1]
		resource := &tableau.DataSource{
			ID: "ds_1",
		}
		deleteDsPath := fmt.Sprintf("/api/%v/sites/%v/datasources/%v/permissions/users/%v/%v/%v", s.apiVersion, s.siteID, resource.ID, userID, capabilityName, capabilityMode)

		deleteDsPermissionRequest, err := s.getTestRequest(http.MethodDelete, deleteDsPath, nil)
		s.Require().NoError(err)

		deleteDsPermissionResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(nil)}
		s.mockHttpClient.On("Do", deleteDsPermissionRequest).Return(&deleteDsPermissionResponse, nil).Once()

		actualError := s.client.RevokeDataSourceAccess(resource, userEmail, role)

		s.Nil(actualError)
	})
}

func (s *ClientTestSuite) TestRevokeViewAccess() {
	s.Run("should return error if error in getting the user", func() {
		s.setup()

		expectedError := tableau.ErrUserNotFound
		userEmail := "test-email@gojek.com"
		filter := fmt.Sprintf("name:eq:%v", userEmail)
		path := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", s.apiVersion, s.siteID, filter)

		GetUserRequest, err := s.getTestRequest(http.MethodGet, path, nil) //testing getUser()
		s.Require().NoError(err)

		GetUserResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"users": {"user":[]}} `

		GetUserResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(GetUserResponseJSON)))}
		s.mockHttpClient.On("Do", GetUserRequest).Return(&GetUserResponse, nil).Once()

		role := "write:allow"
		resource := &tableau.View{
			ID: "vw_1",
		}
		actualError := s.client.RevokeViewAccess(resource, userEmail, role)

		s.Equal(expectedError, actualError)
	})
	s.Run("should revoke access to view and return nil error on success", func() {
		s.setup()

		userEmail := "test-email@gojek.com"
		filter := fmt.Sprintf("name:eq:%v", userEmail) //test getUser
		path := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", s.apiVersion, s.siteID, filter)

		GetUserRequest, err := s.getTestRequest(http.MethodGet, path, nil)
		s.Require().NoError(err)

		GetUserResponseJSON := `{"pagination":{"pageNumber":"1","pageSize":"1","totalAvailable":"1"},"users": {"user":[{"id": "userID-1"}]}} `

		GetUserResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(GetUserResponseJSON)))}
		s.mockHttpClient.On("Do", GetUserRequest).Return(&GetUserResponse, nil).Once()

		userID := "userID-1"
		role := "write:allow"
		split := strings.Split(role, ":")
		capabilityName := split[0]
		capabilityMode := split[1]
		resource := &tableau.View{
			ID: "vw_1",
		}
		deleteViewPath := fmt.Sprintf("/api/%v/sites/%v/views/%v/permissions/users/%v/%v/%v", s.apiVersion, s.siteID, resource.ID, userID, capabilityName, capabilityMode)

		deleteViewPermissionRequest, err := s.getTestRequest(http.MethodDelete, deleteViewPath, nil)
		s.Require().NoError(err)

		deleteViewPermissionResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(nil)}
		s.mockHttpClient.On("Do", deleteViewPermissionRequest).Return(&deleteViewPermissionResponse, nil).Once()

		actualError := s.client.RevokeViewAccess(resource, userEmail, role)

		s.Nil(actualError)
	})
}

// t.Run("should return error and if session response is not success", func(t *testing.T) {
// 	// TODO: test http request execution
// 	mockHttpClient := new(mocks.HTTPClient)
// 	config := &tableau.ClientConfig{
// 		Username:   "test-username",
// 		Password:   "test-password",
// 		Host:       "http://localhost",
// 		ContentURL: "test-content-url",
// 		HTTPClient: mockHttpClient,
// 		APIVersion: "3.2",
// 	}

// 	sessionToken := "93df71b4-6887-46bd-b4bf-7ad3b94bd6fe"
// 	siteID := "1.0"
// 	userID := "test-user-id"
// 	sessionResponseJSON := fmt.Sprintf(`{"credentials":{"token":"%s","site":{"id":"%s"},"user":{"id":"%s"}}}`, sessionToken, siteID, userID)
// 	sessionResponse := http.Response{
// 		StatusCode: 401,
// 		Body:       ioutil.NopCloser(bytes.NewReader([]byte(sessionResponseJSON))),
// 	}
// 	mockHttpClient.On("Do", mock.Anything).Return(&sessionResponse, nil).Once()

// 	// sessionToken := "93df71b4-6887-46bd-b4bf-7ad3b94bd6fe"
// 	// responseJSON := `{"id":"` + sessionToken + `"}`
// 	// response := http.Response{StatusCode: 401, Body: ioutil.NopCloser(bytes.NewReader([]byte(responseJSON)))}
// 	// mockHttpClient.On("Do", mock.Anything).Return(&response, nil).Once()

// 	//sessionResponseJSON := fmt.Sprintf(`{"credentials":{"token":"%s","site":{"id":"%s"},"user":{"id":"%s"}}}`, sessionToken, siteID, userID)

// 	// apiVersion := "3.2"
// 	// url:= fmt.Sprintf("/api/%v/auth/signin", apiVersion)
// 	// testRequest, err := s.getTestRequest(http.MethodGet, url, nil)

// 	sessionResponse2 := http.Response{
// 		StatusCode: 200,
// 		Body:       ioutil.NopCloser(bytes.NewReader([]byte(sessionResponseJSON))),
// 	}
// 	mockHttpClient.On("Do", mock.Anything).Return(&sessionResponse2, nil).Once()

// 	mockHttpClient.On("Do", mock.Anything).Return(&sessionResponse2, nil).Once()

// 	_, actualError := tableau.NewClient(config)
// 	//mockHttpClient.AssertExpectations(t)
// 	assert.Nil(t, actualError)
// })
