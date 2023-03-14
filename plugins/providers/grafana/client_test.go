package grafana_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/goto/guardian/mocks"
	"github.com/goto/guardian/plugins/providers/grafana"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestNewClient(t *testing.T) {
	t.Run("should return error if config is invalid", func(t *testing.T) {
		invalidConfig := &grafana.ClientConfig{}

		actualClient, actualError := grafana.NewClient(invalidConfig)

		assert.Nil(t, actualClient)
		assert.Error(t, actualError)
	})

	t.Run("should return error if config.Host is not a valid url", func(t *testing.T) {
		invalidHostConfig := &grafana.ClientConfig{
			Host:     "invalid-url",
			Username: "test-username",
			Password: "test-password",
		}

		actualClient, actualError := grafana.NewClient(invalidHostConfig)

		assert.Nil(t, actualClient)
		assert.Error(t, actualError)
	})

	t.Run("should return client and nil error on success", func(t *testing.T) {
		mockHttpClient := new(mocks.HTTPClient)
		testCases := []struct {
			name   string
			config grafana.ClientConfig
		}{
			{
				name: "config without httpClient",
				config: grafana.ClientConfig{
					Username: "test-username",
					Password: "test-password",
					Host:     "http://localhost",
					Org:      "test-Org",
				},
			},
			{
				name: "config with mockHttpClient",
				config: grafana.ClientConfig{
					Username:   "test-username",
					Password:   "test-password",
					Host:       "http://localhost",
					Org:        "test-Org",
					HTTPClient: mockHttpClient,
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				config := tc.config
				_, actualError := grafana.NewClient(&config)
				mockHttpClient.AssertExpectations(t)
				assert.Nil(t, actualError)
			})
		}
	})
}

type ClientTestSuite struct {
	suite.Suite

	mockHttpClient *mocks.HTTPClient
	client         grafana.GrafanaClient
	host           string
	org            string
}

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
func (s *ClientTestSuite) setup() {
	s.mockHttpClient = new(mocks.HTTPClient)

	s.host = "http://localhost"
	s.org = "test-Org"
	client, err := grafana.NewClient(&grafana.ClientConfig{
		Username:   "test-username",
		Password:   "test-password",
		Host:       s.host,
		Org:        s.org,
		HTTPClient: s.mockHttpClient,
	})
	s.Require().NoError(err)
	s.client = client
}

func (s *ClientTestSuite) TestGetFolders() {
	s.Run("should get folders and nil error on success", func() {
		s.setup()

		testRequest, err := s.getTestRequest(http.MethodGet, "/api/folders", nil)
		s.Require().NoError(err)

		folderResponseJSON := `[]`
		folderResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(folderResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&folderResponse, errors.New("Http Client Error")).Once()

		actualFolders, err1 := s.client.GetFolders()

		s.Nil(actualFolders)
		s.Error(err1)
	})

	s.Run("should get folders and nil error on success", func() {
		s.setup()

		testRequest, err := s.getTestRequest(http.MethodGet, "/api/folders", nil)
		s.Require().NoError(err)

		expectedFolders := []grafana.Folder{
			{
				ID:    1,
				Title: "fd_1",
			},
			{
				ID:    2,
				Title: "fd_2",
			},
		}
		folderResponseJSON := `[{"id":1,"title":"fd_1"},{"id":2,"title":"fd_2"}]`
		folderResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(folderResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&folderResponse, nil).Once()

		result, err1 := s.client.GetFolders()
		var folders []grafana.Folder
		for _, fd := range result {
			folders = append(folders, *fd)
		}

		s.Nil(err1)
		s.Equal(expectedFolders, folders)
	})
}

func (s *ClientTestSuite) TestGetDashboards() {
	s.Run("should get folders and nil error on success", func() {
		s.setup()

		folderID := 50
		url := fmt.Sprintf("/api/search?folderIds=%d&type=dash-db", folderID)
		testRequest, err := s.getTestRequest(http.MethodGet, url, nil)
		s.Require().NoError(err)

		dashboardResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte("")))}
		s.mockHttpClient.On("Do", testRequest).Return(&dashboardResponse, errors.New("http client error")).Once()

		result, err1 := s.client.GetDashboards(folderID)

		s.Nil(result)
		s.Error(err1)
	})

	s.Run("should get folders and nil error on success", func() {
		s.setup()

		folderID := 50
		url := fmt.Sprintf("/api/search?folderIds=%d&type=dash-db", folderID)
		testRequest, err := s.getTestRequest(http.MethodGet, url, nil)
		s.Require().NoError(err)

		expectedDashboards := []grafana.Dashboard{
			{
				ID:    1,
				Title: "db_1",
			},
			{
				ID:    2,
				Title: "db_2",
			},
		}
		dashboardResponseJSON := `[{"id":1,"title":"db_1"},{"id":2,"title":"db_2"}]`
		dashboardResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(dashboardResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&dashboardResponse, nil).Once()

		result, err1 := s.client.GetDashboards(folderID)
		var dashboards []grafana.Dashboard
		for _, db := range result {
			dashboards = append(dashboards, *db)
		}

		s.Nil(err1)
		s.Equal(expectedDashboards, dashboards)
	})
}

func (s *ClientTestSuite) getTestRequest(method, path string, body interface{}) (*http.Request, error) {
	basicKey := "dGVzdC11c2VybmFtZTp0ZXN0LXBhc3N3b3Jk"
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
	req.Header.Set("Authorization", "Basic "+basicKey)
	req.Header.Set("X-Grafana-Org-Id", s.org)

	return req, nil
}

func (s *ClientTestSuite) TestGrantDashboardAccess() {
	s.Run("should return error if user not found", func() {
		s.setup()

		user := "test-email@gojek.com"
		role := "view"
		expectedError := errors.New("Error in getting user")
		url := fmt.Sprintf("/api/users/lookup?loginOrEmail=%s", user) //testing the getUser(user) Response
		testRequest, err := s.getTestRequest(http.MethodGet, url, nil)
		s.Require().NoError(err)

		userResponse := http.Response{StatusCode: 404, Body: ioutil.NopCloser(bytes.NewReader([]byte(nil)))}
		s.mockHttpClient.On("Do", testRequest).Return(&userResponse, expectedError).Once()

		resource := grafana.Dashboard{
			ID:          1,
			Title:       "db_1",
			FolderID:    50,
			FolderTitle: "fd_1",
		}

		actualError := s.client.GrantDashboardAccess(&resource, user, role)
		s.Equal(expectedError, actualError)
	})

	s.Run("should return error if user not found", func() {
		s.setup()

		user := "test-email@gojek.com"
		role := "view"

		url := fmt.Sprintf("/api/users/lookup?loginOrEmail=%s", user) //testing the getUser(user) Response
		testRequest, err := s.getTestRequest(http.MethodGet, url, nil)
		s.Require().NoError(err)

		userResponseJSON := `{ "id":55,"email":"test-email@gojek.com" }`
		userResponse := http.Response{StatusCode: 404, Body: ioutil.NopCloser(bytes.NewReader([]byte(userResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&userResponse, nil).Once()

		resource := grafana.Dashboard{
			ID:          1,
			Title:       "db_1",
			FolderID:    50,
			FolderTitle: "fd_1",
		}

		actualError := s.client.GrantDashboardAccess(&resource, user, role)
		e := grafana.ErrUserNotFound
		s.Equal(e, actualError)
	})

	s.Run("should return an error if permissions are invalid for dashboard access", func() {
		s.setup()

		user := "test-email@gojek.com"
		role := "invalid role"
		expectedError := grafana.ErrInvalidPermissionType

		url := fmt.Sprintf("/api/users/lookup?loginOrEmail=%s", user) //testing the getUser(user) Response
		testRequest, err := s.getTestRequest(http.MethodGet, url, nil)
		s.Require().NoError(err)

		userResponseJSON := `{ "id":55,"email":"test-email@gojek.com" }`
		userResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(userResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&userResponse, nil).Once()

		resource := grafana.Dashboard{
			ID:          1,
			Title:       "db_1",
			FolderID:    50,
			FolderTitle: "fd_1",
		}

		actualError := s.client.GrantDashboardAccess(&resource, user, role)
		s.Equal(expectedError, actualError)
	})

	s.Run("should grant dashboard access for inherited permissions", func() {
		s.setup()

		user := "test-email@gojek.com"
		role := "view" //valid roles are "view", "edit", "admin"

		url := fmt.Sprintf("/api/users/lookup?loginOrEmail=%s", user) //testing the getUser(user) Response
		testRequest, err := s.getTestRequest(http.MethodGet, url, nil)
		s.Require().NoError(err)

		userResponseJSON := `{ "id":55,"email":"test-email@gojek.com" }`
		userResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(userResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&userResponse, nil).Once()

		resource := grafana.Dashboard{
			ID:          1,
			Title:       "db_1",
			FolderID:    50,
			FolderTitle: "fd_1",
		}

		id := resource.ID //test for getDashboardPermissions(resource.ID)
		permissionsUrl := fmt.Sprintf("/api/dashboards/id/%d/permissions", id)
		permissionsRequest, err2 := s.getTestRequest(http.MethodGet, permissionsUrl, nil)
		s.Require().NoError(err2)

		permissionsResponseJSON := `[{"userID":55, "permission":1,"inherited":false}]` //permission codes are: "view": 1, "edit": 2, "admin": 4
		permissionsResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(permissionsResponseJSON)))}
		s.mockHttpClient.On("Do", permissionsRequest).Return(&permissionsResponse, nil).Once()

		updatePermissionsResponseJSON := `[{"permission":1,"inherited":true}]`
		updatePermissionsResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(updatePermissionsResponseJSON)))}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&updatePermissionsResponse, nil).Once()

		actualError := s.client.GrantDashboardAccess(&resource, user, role)
		s.Nil(actualError)
		s.mockHttpClient.AssertExpectations(s.T())
	})
	s.Run("should grant dashboard access for non inherited permissions", func() {
		s.setup()

		user := "test-email@gojek.com"
		role := "view" //valid roles are "view", "edit", "admin"

		url := fmt.Sprintf("/api/users/lookup?loginOrEmail=%s", user) //testing the getUser(user) Response
		testRequest, err := s.getTestRequest(http.MethodGet, url, nil)
		s.Require().NoError(err)

		userResponseJSON := `{ "id":1,"email":"test-email@gojek.com" }`
		userResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(userResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&userResponse, nil).Once()

		resource := grafana.Dashboard{
			ID:          1,
			Title:       "db_1",
			FolderID:    50,
			FolderTitle: "fd_1",
		}

		id := resource.ID //test for getDashboardPermissions(resource.ID)
		permissionsUrl := fmt.Sprintf("/api/dashboards/id/%d/permissions", id)
		permissionsRequest, err2 := s.getTestRequest(http.MethodGet, permissionsUrl, nil)
		s.Require().NoError(err2)

		permissionsResponseJSON := `[{"permission":1,"inherited":true}]` //permission codes are: "view": 1, "edit": 2, "admin": 4
		permissionsResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(permissionsResponseJSON)))}
		s.mockHttpClient.On("Do", permissionsRequest).Return(&permissionsResponse, nil).Once()

		updatePermissionsResponseJSON := `[{"permission":1,"inherited":true}]`
		updatePermissionsResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(updatePermissionsResponseJSON)))}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&updatePermissionsResponse, nil).Once()

		actualError := s.client.GrantDashboardAccess(&resource, user, role)
		s.Nil(actualError)
		s.mockHttpClient.AssertExpectations(s.T())
	})
}

func (s *ClientTestSuite) TestRevokeDashboardAccess() {
	s.Run("should return error if user not found", func() {
		s.setup()

		user := "test-email@gojek.com"
		role := "view"
		expectedError := errors.New("Error in getting user")
		url := fmt.Sprintf("/api/users/lookup?loginOrEmail=%s", user) //testing the getUser(user) Response
		testRequest, err := s.getTestRequest(http.MethodGet, url, nil)
		s.Require().NoError(err)

		userResponse := http.Response{StatusCode: 404, Body: ioutil.NopCloser(bytes.NewReader([]byte(nil)))}
		s.mockHttpClient.On("Do", testRequest).Return(&userResponse, expectedError).Once()

		resource := grafana.Dashboard{
			ID:          1,
			Title:       "db_1",
			FolderID:    50,
			FolderTitle: "fd_1",
		}

		actualError := s.client.RevokeDashboardAccess(&resource, user, role)
		s.Equal(expectedError, actualError)
	})

	s.Run("should return an error if permissions are invalid for dashboard access", func() {
		s.setup()

		user := "test-email@gojek.com"
		role := "invalid role" //valid roles are "view", "edit", "admin"
		expectedError := grafana.ErrInvalidPermissionType

		url := fmt.Sprintf("/api/users/lookup?loginOrEmail=%s", user) //testing the getUser(user) Response
		testRequest, err := s.getTestRequest(http.MethodGet, url, nil)
		s.Require().NoError(err)

		userResponseJSON := `{ "id":55,"email":"test-email@gojek.com" }`
		userResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(userResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&userResponse, nil).Once()

		resource := grafana.Dashboard{
			ID:          1,
			Title:       "db_1",
			FolderID:    50,
			FolderTitle: "fd_1",
		}

		actualError := s.client.RevokeDashboardAccess(&resource, user, role)
		s.Equal(expectedError, actualError)
	})

	s.Run("should return error if requested permission to be revoked isn't already there in the existing user permissions", func() {
		s.setup()

		user := "test-email@gojek.com"
		role := "view" //valid roles are "view", "edit", "admin"
		expectedError := grafana.ErrPermissionNotFound
		url := fmt.Sprintf("/api/users/lookup?loginOrEmail=%s", user) //testing the getUser(user) Response
		testRequest, err := s.getTestRequest(http.MethodGet, url, nil)
		s.Require().NoError(err)
		userResponseJSON := `{ "id":1,"email":"test-email@gojek.com" }`
		userResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(userResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&userResponse, nil).Once()

		resource := grafana.Dashboard{
			ID:          1,
			Title:       "db_1",
			FolderID:    50,
			FolderTitle: "fd_1",
		}
		id := resource.ID
		permissionsUrl := fmt.Sprintf("/api/dashboards/id/%d/permissions", id)
		permissionsRequest, err2 := s.getTestRequest(http.MethodGet, permissionsUrl, nil)
		s.Require().NoError(err2)
		permissionsResponseJSON := `[{"permission":1,"inherited":false,"userID":0}]` //permission codes are: "view": 1, "edit": 2, "admin": 4
		permissionsResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(permissionsResponseJSON)))}
		s.mockHttpClient.On("Do", permissionsRequest).Return(&permissionsResponse, nil).Once()

		updatePermissionsResponseJSON := `[{"permission":1,"inherited":false}]`
		updatePermissionsResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(updatePermissionsResponseJSON)))}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&updatePermissionsResponse, nil).Once()

		actualError := s.client.RevokeDashboardAccess(&resource, user, role)
		s.Equal(expectedError, actualError)
	})

	s.Run("should revoke access to dashboard and return nil error", func() {
		s.setup()

		user := "test-email@gojek.com"
		role := "view" //valid roles are "view", "edit", "admin"

		url := fmt.Sprintf("/api/users/lookup?loginOrEmail=%s", user) //testing the getUser(user) Response
		testRequest, err := s.getTestRequest(http.MethodGet, url, nil)
		s.Require().NoError(err)
		userResponseJSON := `{ "id":1,"email":"test-email@gojek.com" }`
		userResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(userResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&userResponse, nil).Once()

		resource := grafana.Dashboard{
			ID:          1,
			Title:       "db_1",
			FolderID:    50,
			FolderTitle: "fd_1",
		}
		id := resource.ID
		permissionsUrl := fmt.Sprintf("/api/dashboards/id/%d/permissions", id)
		permissionsRequest, err2 := s.getTestRequest(http.MethodGet, permissionsUrl, nil)
		s.Require().NoError(err2)
		permissionsResponseJSON := `[{"permission":1,"inherited":false,"userID":1}]` //permission codes are: "view": 1, "edit": 2, "admin": 4
		permissionsResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(permissionsResponseJSON)))}
		s.mockHttpClient.On("Do", permissionsRequest).Return(&permissionsResponse, nil).Once()

		updatePermissionsResponseJSON := `[{"permission":1,"inherited":false}]`
		updatePermissionsResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(updatePermissionsResponseJSON)))}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&updatePermissionsResponse, nil).Once()

		actualError := s.client.RevokeDashboardAccess(&resource, user, role)
		s.Nil(actualError)
	})
}
