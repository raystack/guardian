package frontier_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/raystack/salt/log"

	"github.com/raystack/guardian/mocks"
	"github.com/raystack/guardian/plugins/providers/frontier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestNewClient(t *testing.T) {
	t.Run("should return error if config is invalid", func(t *testing.T) {
		invalidConfig := &frontier.ClientConfig{}
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		actualClient, actualError := frontier.NewClient(invalidConfig, logger)

		assert.Nil(t, actualClient)
		assert.Error(t, actualError)
	})

	t.Run("should return error if config.Host is not a valid url", func(t *testing.T) {
		invalidHostConfig := &frontier.ClientConfig{
			AuthHeader: "X-Frontier-Email",
			AuthEmail:  "test-email",
			Host:       "invalid-url",
			HTTPClient: new(mocks.HTTPClient),
		}
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		actualClient, actualError := frontier.NewClient(invalidHostConfig, logger)

		assert.Nil(t, actualClient)
		assert.Error(t, actualError)
	})

	t.Run("should return client and nil error on success", func(t *testing.T) {
		// TODO: test http request execution
		mockHttpClient := new(mocks.HTTPClient)
		config := &frontier.ClientConfig{
			AuthHeader: "X-Frontier-Email",
			AuthEmail:  "test_email",
			Host:       "http://localhost",
			HTTPClient: mockHttpClient,
		}
		logger := log.NewLogrus(log.LogrusWithLevel("info"))

		_, actualError := frontier.NewClient(config, logger)
		mockHttpClient.AssertExpectations(t)
		assert.Nil(t, actualError)
	})
}

type ClientTestSuite struct {
	suite.Suite

	mockHttpClient *mocks.HTTPClient
	client         frontier.Client
	authHeader     string
	auth           string
	host           string
}

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (s *ClientTestSuite) setup() {
	logger := log.NewNoop()
	s.mockHttpClient = new(mocks.HTTPClient)

	s.host = "http://localhost"
	s.auth = "frontier_admin"
	s.authHeader = "X-Frontier-Email"
	client, err := frontier.NewClient(&frontier.ClientConfig{
		AuthHeader: s.authHeader,
		AuthEmail:  s.auth,
		Host:       s.host,
		HTTPClient: s.mockHttpClient,
	}, logger)
	s.Require().NoError(err)
	s.client = client
}

func (s *ClientTestSuite) getTestRequest(method, path string, body interface{}, authEmail string) (*http.Request, error) {
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
	if authEmail == "" {
		req.Header.Set(s.authHeader, s.auth)
	} else {
		req.Header.Set(s.authHeader, authEmail)
	}
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func (s *ClientTestSuite) TestGetGroups() {
	s.Run("should get groups and nil error on success", func() {
		s.setup()

		testRequest, err := s.getTestRequest(http.MethodGet, "/v1beta1/organizations/org_id_1/groups", nil, "")
		s.Require().NoError(err)

		groupResponseJSON := `{
		    "groups": [
		        {
		            "id": "group_id_1",
		            "name": "group_1",
		            "title": "group_1",
		            "orgId": "org_id_1",
		            "metadata": {
		                "email": "group_1@email.com",
		                "privacy": "public",
		                "slack": "@group_1"
		            },
		            "createdAt": "2022-03-17T06:19:47.176089Z",
		            "updatedAt": "2022-03-17T06:19:47.176089Z"
		        },
		        {
		            "id": "group_id_2",
		            "name": "group_2",
		            "title": "group_2",
		            "orgId": "org_id_1",
		            "metadata": {
		                "email": "group_2@email.com",
		                "privacy": "public",
		                "slack": "@group_2"
		            },
		            "createdAt": "2022-03-30T10:49:10.965916Z",
		            "updatedAt": "2022-03-30T10:49:10.965916Z"
		        }
		    ]
		}`
		groupResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(groupResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&groupResponse, nil).Once()

		expectedTeams := []frontier.Group{
			{
				ID:    "group_id_1",
				Name:  "group_1",
				Title: "group_1",
				OrgId: "org_id_1",
				Metadata: frontier.Metadata{
					Email:   "group_1@email.com",
					Privacy: "public",
					Slack:   "@group_1",
				},
			},
			{
				ID:    "group_id_2",
				Name:  "group_2",
				Title: "group_2",
				OrgId: "org_id_1",
				Metadata: frontier.Metadata{
					Email:   "group_2@email.com",
					Privacy: "public",
					Slack:   "@group_2",
				},
			},
		}

		result, err1 := s.client.GetGroups("org_id_1")
		var groups []frontier.Group
		for _, group := range result {
			groups = append(groups, *group)
		}
		s.Nil(err1)
		s.ElementsMatch(expectedTeams, groups)
	})
}

func (s *ClientTestSuite) TestGetProjects() {
	s.Run("should get projects and nil error on success", func() {
		s.setup()

		testRequest, err := s.getTestRequest(http.MethodGet, "/v1beta1/organizations/org_id_1/projects", nil, "")
		s.Require().NoError(err)

		projectResponseJSON := `{
		    "projects": [
		        {
		            "id": "project_id_1",
		            "name": "project_1",
		            "title": "project_1",
		            "orgId": "org_id_1",
		            "metadata": {
		                "environment": "integration",
                		"landscape": "core",
                		"organization": "gtfn"
		            },
		            "createdAt": "2022-03-17T06:19:47.176089Z",
		            "updatedAt": "2022-03-17T06:19:47.176089Z"
		        }
		    ]
		}`
		projectResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(projectResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&projectResponse, nil).Once()

		expectedProjects := []frontier.Project{
			{
				ID:     "project_id_1",
				Name:   "project_1",
				Title:  "project_1",
				OrgId:  "org_id_1",
				Admins: []string{"test.admin@email.com"},
			},
		}

		testAdminsRequest, err := s.getTestRequest(http.MethodGet, "/v1beta1/projects/project_id_1/admins", nil, "")
		s.Require().NoError(err)

		projectAdminResponse := `{
			"users": [
				{
					"id": "admin_id",
					"name": "Test_Admin",
					"title": "Test_Admin",
					"email": "test.admin@email.com",
					"metadata": {
						"slack": "@Test_Admin"
					},
					"createdAt": "2022-03-17T09:43:12.391071Z",
					"updatedAt": "2022-03-17T09:43:12.391071Z"
				}
			]
		}`

		projectAdminResponse1 := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(projectAdminResponse)))}
		s.mockHttpClient.On("Do", testAdminsRequest).Return(&projectAdminResponse1, nil).Once()

		result, err1 := s.client.GetProjects("org_id_1")
		var projects []frontier.Project
		for _, project := range result {
			projects = append(projects, *project)
		}
		s.Nil(err1)
		s.ElementsMatch(expectedProjects, projects)
	})
}

func (s *ClientTestSuite) TestGetOrganizations() {
	s.Run("should get organizations and nil error on success", func() {
		s.setup()

		testRequest, err := s.getTestRequest(http.MethodGet, "/v1beta1/organizations", nil, "")
		s.Require().NoError(err)

		organizationsResponseJSON := `{
		    "organizations": [
		        {
		            "id": "org_id_1",
		            "name": "org_1",
		            "title": "org_1",
		            "metadata": {
		                "dwh_group_id": "dwh_group",
                		"dwh_group_name": "dwh_group"
		            },
		            "createdAt": "2022-03-17T06:19:47.176089Z",
		            "updatedAt": "2022-03-17T06:19:47.176089Z"
		        }
		    ]
		}`
		orgResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(organizationsResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&orgResponse, nil).Once()

		expectedOrganizations := []frontier.Organization{
			{
				ID:     "org_id_1",
				Name:   "org_1",
				Title:  "org_1",
				Admins: []string{"test.admin@email.com"},
			},
		}

		testAdminsRequest, err := s.getTestRequest(http.MethodGet, "/v1beta1/organizations/org_id_1/admins", nil, "")
		s.Require().NoError(err)

		orgAdminResponse := `{
			"users": [
				{
					"id": "admin_id",
					"name": "Test_Admin",
					"title": "Test_Admin",
					"email": "test.admin@email.com",
					"metadata": {
						"slack": "@Test_Admin"
					},
					"createdAt": "2022-03-17T09:43:12.391071Z",
					"updatedAt": "2022-03-17T09:43:12.391071Z"
				}
			]
		}`

		orgAdminResponse1 := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(orgAdminResponse)))}
		s.mockHttpClient.On("Do", testAdminsRequest).Return(&orgAdminResponse1, nil).Once()

		result, err1 := s.client.GetOrganizations()
		var orgs []frontier.Organization
		for _, org := range result {
			orgs = append(orgs, *org)
		}
		s.Nil(err1)
		s.ElementsMatch(expectedOrganizations, orgs)
	})
}

func (s *ClientTestSuite) TestGrantTeamAccess() {
	s.Run("should grant access to group and nil error on success", func() {
		s.setup()

		testUserId := "test_user_id"
		var groupObj *frontier.Group
		groupObj = new(frontier.Group)
		groupObj.ID = "test_group_id"
		role := "users"

		body := make(map[string]string)
		body["principal"] = fmt.Sprintf("%s:%s", "app/user", testUserId)
		body["resource"] = fmt.Sprintf("%s:%s", "app/group", groupObj.ID)
		body["roleId"] = role

		responseJson := `{}`

		responseUsers := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(responseJson)))}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&responseUsers, nil).Once()

		actualError := s.client.GrantGroupAccess(groupObj, testUserId, role)
		s.Nil(actualError)
	})
}

func (s *ClientTestSuite) TestGrantProjectAccess() {
	s.Run("should grant access to project and nil error on success", func() {
		s.setup()

		testUserId := "test_user_id"
		body := make(map[string]string)
		var projectObj *frontier.Project
		projectObj = new(frontier.Project)
		projectObj.ID = "test_project_id"
		role := "admins"

		body["principal"] = fmt.Sprintf("%s:%s", "app/user", testUserId)
		body["resource"] = fmt.Sprintf("%s:%s", "app/project", projectObj.ID)
		body["roleId"] = role

		responseJson := `{}`

		responseUsers := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(responseJson)))}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&responseUsers, nil).Once()

		actualError := s.client.GrantProjectAccess(projectObj, testUserId, role)
		s.Nil(actualError)
	})
}

func (s *ClientTestSuite) TestGrantOrganizationAccess() {
	s.Run("should grant access to organization and nil error on success", func() {
		s.setup()
		testUserId := "test_user_id"
		body := make(map[string]string)
		var orgObj *frontier.Organization
		orgObj = new(frontier.Organization)
		orgObj.ID = "test_org_id"
		role := "admins"

		body["roleId"] = role
		body["resource"] = fmt.Sprintf("%s:%s", "app/organization", orgObj.ID)
		body["principal"] = fmt.Sprintf("%s:%s", "app/user", testUserId)

		responseJson := `{}`

		responseUsers := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(responseJson)))}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&responseUsers, nil).Once()

		actualError := s.client.GrantOrganizationAccess(orgObj, testUserId, role)
		s.Nil(actualError)
	})
}

func (s *ClientTestSuite) TestRevokeTeamAccess() {
	s.Run("should revoke access to group and nil error on success", func() {
		s.setup()
		testUserId := "test_user_id"
		role := frontier.RoleGroupMember
		var groupObj *frontier.Group
		groupObj = new(frontier.Group)
		groupObj.ID = "test_group_id"

		responseJson := `{
			"policies": [
				{
					"id": "1"
				}
			]
		}`
		responseUsers := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(responseJson)))}
		s.mockHttpClient.On("Do", mock.Anything).Return(&responseUsers, nil).Once()

		responseJson = `{}`
		deletePolicyResp := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(responseJson)))}
		s.mockHttpClient.On("Do", mock.Anything).Return(&deletePolicyResp, nil).Once()

		actualError := s.client.RevokeGroupAccess(groupObj, testUserId, role)
		s.Nil(actualError)
	})
}

func (s *ClientTestSuite) TestRevokeProjectAccess() {
	s.Run("should revoke access to project and nil error on success", func() {
		s.setup()

		testUserId := "test_user_id"

		body := make(map[string][]string)
		body["userIds"] = append(body["userIds"], testUserId)

		var projectObj *frontier.Project
		projectObj = new(frontier.Project)
		projectObj.ID = "test_project_id"

		role := "admins"

		responseJson := `{
			"policies": [
				{
					"id": "1"
				}
			]
		}`
		responseUsers := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(responseJson)))}
		s.mockHttpClient.On("Do", mock.Anything).Return(&responseUsers, nil).Once()

		responseJson = `{}`
		deletePolicyResp := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(responseJson)))}
		s.mockHttpClient.On("Do", mock.Anything).Return(&deletePolicyResp, nil).Once()

		actualError := s.client.RevokeProjectAccess(projectObj, testUserId, role)
		s.Nil(actualError)
	})
}

func (s *ClientTestSuite) TestRevokeOrganizationAccess() {
	s.Run("should revoke access to organization and nil error on success", func() {
		s.setup()
		testUserId := "test_user_id"

		body := make(map[string][]string)
		body["userIds"] = append(body["userIds"], testUserId)

		var orgObj *frontier.Organization
		orgObj = new(frontier.Organization)
		orgObj.ID = "test_org_id"

		role := "admins"

		responseJson := `{
			"policies": [
				{
					"id": "1"
				}
			]
		}`
		responseUsers := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(responseJson)))}
		s.mockHttpClient.On("Do", mock.Anything).Return(&responseUsers, nil).Once()

		responseJson = `{}`
		deletePolicyResp := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(responseJson)))}
		s.mockHttpClient.On("Do", mock.Anything).Return(&deletePolicyResp, nil).Once()

		actualError := s.client.RevokeOrganizationAccess(orgObj, testUserId, role)
		s.Nil(actualError)
	})
}

func (s *ClientTestSuite) TestGetSelfUser() {
	s.Run("Should return error user on empty email", func() {
		s.setup()
		testUserEmail := ""

		testGetSelfRequest, err := s.getTestRequest(http.MethodGet, "/v1beta1/users/self", nil, testUserEmail)
		s.Require().NoError(err)

		responseJson := `{
			"code": 2,
			"message": "email id is empty",
			"details": []
		}`

		responseUser := http.Response{StatusCode: 500, Body: ioutil.NopCloser(bytes.NewReader([]byte(responseJson)))}
		s.mockHttpClient.On("Do", testGetSelfRequest).Return(&responseUser, nil).Once()

		_, actualError := s.client.GetSelfUser(testUserEmail)
		s.NotNil(actualError)
	})
	s.Run("Should return frontier user on success", func() {
		s.setup()
		testUserEmail := "test_user@email.com"

		testGetSelfRequest, err := s.getTestRequest(http.MethodGet, "/v1beta1/users/self", nil, testUserEmail)
		s.Require().NoError(err)

		responseJson := `{
			 "user": {
				"id": "test-user-id",
				"name": "test-user",
				"title": "",
				"email": "test_user@email.com",
				"metadata": {
					"slack": "test"
				},
				"createdAt": "2022-05-05T05:32:42.384021Z",
				"updatedAt": "2022-06-01T10:08:02.688055Z"
			}
		}`

		responseUser := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(responseJson)))}
		s.mockHttpClient.On("Do", testGetSelfRequest).Return(&responseUser, nil).Once()

		expectedUser := &frontier.User{
			ID:    "test-user-id",
			Name:  "test-user",
			Email: "test_user@email.com",
		}

		user, actualError := s.client.GetSelfUser(testUserEmail)
		s.EqualValues(expectedUser, user)
		s.Nil(actualError)
	})
}
