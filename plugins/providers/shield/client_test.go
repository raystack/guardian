package shield_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/goto/salt/log"

	"github.com/goto/guardian/mocks"
	"github.com/goto/guardian/plugins/providers/shield"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestNewClient(t *testing.T) {
	t.Run("should return error if config is invalid", func(t *testing.T) {
		invalidConfig := &shield.ClientConfig{}
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		actualClient, actualError := shield.NewClient(invalidConfig, logger)

		assert.Nil(t, actualClient)
		assert.Error(t, actualError)
	})

	t.Run("should return error if config.Host is not a valid url", func(t *testing.T) {
		invalidHostConfig := &shield.ClientConfig{
			AuthHeader: "X-Auth-Email",
			AuthEmail:  "test-email",
			Host:       "invalid-url",
		}
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		actualClient, actualError := shield.NewClient(invalidHostConfig, logger)

		assert.Nil(t, actualClient)
		assert.Error(t, actualError)
	})

	t.Run("should return client and nil error on success", func(t *testing.T) {
		// TODO: test http request execution
		mockHttpClient := new(mocks.HTTPClient)
		config := &shield.ClientConfig{
			AuthHeader: "X-Auth-Email",
			AuthEmail:  "test_email",
			Host:       "http://localhost",
			HTTPClient: mockHttpClient,
		}
		logger := log.NewLogrus(log.LogrusWithLevel("info"))

		_, actualError := shield.NewClient(config, logger)
		mockHttpClient.AssertExpectations(t)
		assert.Nil(t, actualError)
	})
}

type ClientTestSuite struct {
	suite.Suite

	mockHttpClient *mocks.HTTPClient
	client         shield.ShieldClient
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
	s.auth = "shield_admin"
	s.authHeader = "X-Auth-Email"
	client, err := shield.NewClient(&shield.ClientConfig{
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

func (s *ClientTestSuite) TestGetTeams() {
	s.Run("should get teams and nil error on success", func() {
		s.setup()

		testRequest, err := s.getTestRequest(http.MethodGet, "/admin/v1beta1/groups", nil, "")
		s.Require().NoError(err)

		teamResponseJSON := `{
		    "groups": [
		        {
		            "id": "team_id_1",
		            "name": "team_1",
		            "slug": "team_1",
		            "orgId": "org_id_1",
		            "metadata": {
		                "email": "team_1@email.com",
		                "privacy": "public",
		                "slack": "@team_1"
		            },
		            "createdAt": "2022-03-17T06:19:47.176089Z",
		            "updatedAt": "2022-03-17T06:19:47.176089Z"
		        },
		        {
		            "id": "team_id_2",
		            "name": "team_2",
		            "slug": "team_2",
		            "orgId": "org_id_1",
		            "metadata": {
		                "email": "team_2@email.com",
		                "privacy": "public",
		                "slack": "@team_2"
		            },
		            "createdAt": "2022-03-30T10:49:10.965916Z",
		            "updatedAt": "2022-03-30T10:49:10.965916Z"
		        }
		    ]
		}`
		teamResponse := http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(teamResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&teamResponse, nil).Once()

		expectedTeams := []shield.Team{
			{
				ID:    "team_id_1",
				Name:  "team_1",
				Slug:  "team_1",
				OrgId: "org_id_1",
				Metadata: shield.Metadata{
					Email:   "team_1@email.com",
					Privacy: "public",
					Slack:   "@team_1",
				},
				Admins: []string{"test.admin@email.com"},
			},
			{
				ID:    "team_id_2",
				Name:  "team_2",
				Slug:  "team_2",
				OrgId: "org_id_1",
				Metadata: shield.Metadata{
					Email:   "team_2@email.com",
					Privacy: "public",
					Slack:   "@team_2",
				},
				Admins: []string{"test.admin@email.com"},
			},
		}

		testAdminsRequest1, err := s.getTestRequest(http.MethodGet, "/admin/v1beta1/groups/team_id_1/admins", nil, "")
		s.Require().NoError(err)

		testAdminsRequest2, err := s.getTestRequest(http.MethodGet, "/admin/v1beta1/groups/team_id_2/admins", nil, "")
		s.Require().NoError(err)

		teamAdminResponse := `{
			"users": [
				{
					"id": "admin_id",
					"name": "Test_Admin",
					"slug": "Test_Admin",
					"email": "test.admin@email.com",
					"metadata": {
						"slack": "@Test_Admin"
					},
					"createdAt": "2022-03-17T09:43:12.391071Z",
					"updatedAt": "2022-03-17T09:43:12.391071Z"
				}]
		}`

		teamAdminResponse1 := http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(teamAdminResponse)))}
		s.mockHttpClient.On("Do", testAdminsRequest1).Return(&teamAdminResponse1, nil).Once()

		teamAdminResponse2 := http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(teamAdminResponse)))}
		s.mockHttpClient.On("Do", testAdminsRequest2).Return(&teamAdminResponse2, nil).Once()

		result, err1 := s.client.GetTeams()
		var teams []shield.Team
		for _, team := range result {
			teams = append(teams, *team)
		}
		s.Nil(err1)
		s.ElementsMatch(expectedTeams, teams)
	})
}

func (s *ClientTestSuite) TestGetProjects() {
	s.Run("should get projects and nil error on success", func() {
		s.setup()

		testRequest, err := s.getTestRequest(http.MethodGet, "/admin/v1beta1/projects", nil, "")
		s.Require().NoError(err)

		projectResponseJSON := `{
		    "projects": [
		        {
		            "id": "project_id_1",
		            "name": "project_1",
		            "slug": "project_1",
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
		projectResponse := http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(projectResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&projectResponse, nil).Once()

		expectedProjects := []shield.Project{
			{
				ID:     "project_id_1",
				Name:   "project_1",
				Slug:   "project_1",
				OrgId:  "org_id_1",
				Admins: []string{"test.admin@email.com"},
			},
		}

		testAdminsRequest, err := s.getTestRequest(http.MethodGet, "/admin/v1beta1/projects/project_id_1/admins", nil, "")
		s.Require().NoError(err)

		projectAdminResponse := `{
			"users": [
				{
					"id": "admin_id",
					"name": "Test_Admin",
					"slug": "Test_Admin",
					"email": "test.admin@email.com",
					"metadata": {
						"slack": "@Test_Admin"
					},
					"createdAt": "2022-03-17T09:43:12.391071Z",
					"updatedAt": "2022-03-17T09:43:12.391071Z"
				}
			]
		}`

		projectAdminResponse1 := http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(projectAdminResponse)))}
		s.mockHttpClient.On("Do", testAdminsRequest).Return(&projectAdminResponse1, nil).Once()

		result, err1 := s.client.GetProjects()
		var projects []shield.Project
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

		testRequest, err := s.getTestRequest(http.MethodGet, "/admin/v1beta1/organizations", nil, "")
		s.Require().NoError(err)

		organizationsResponseJSON := `{
		    "organizations": [
		        {
		            "id": "org_id_1",
		            "name": "org_1",
		            "slug": "org_1",
		            "metadata": {
		                "dwh_group_id": "dwh_group",
                		"dwh_team_name": "dwh_team"
		            },
		            "createdAt": "2022-03-17T06:19:47.176089Z",
		            "updatedAt": "2022-03-17T06:19:47.176089Z"
		        }
		    ]
		}`
		orgResponse := http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(organizationsResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&orgResponse, nil).Once()

		expectedOrganizations := []shield.Organization{
			{
				ID:     "org_id_1",
				Name:   "org_1",
				Slug:   "org_1",
				Admins: []string{"test.admin@email.com"},
			},
		}

		testAdminsRequest, err := s.getTestRequest(http.MethodGet, "/admin/v1beta1/organizations/org_id_1/admins", nil, "")
		s.Require().NoError(err)

		orgAdminResponse := `{
			"users": [
				{
					"id": "admin_id",
					"name": "Test_Admin",
					"slug": "Test_Admin",
					"email": "test.admin@email.com",
					"metadata": {
						"slack": "@Test_Admin"
					},
					"createdAt": "2022-03-17T09:43:12.391071Z",
					"updatedAt": "2022-03-17T09:43:12.391071Z"
				}
			]
		}`

		orgAdminResponse1 := http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(orgAdminResponse)))}
		s.mockHttpClient.On("Do", testAdminsRequest).Return(&orgAdminResponse1, nil).Once()

		result, err1 := s.client.GetOrganizations()
		var orgs []shield.Organization
		for _, org := range result {
			orgs = append(orgs, *org)
		}
		s.Nil(err1)
		s.ElementsMatch(expectedOrganizations, orgs)
	})
}

func (s *ClientTestSuite) TestGrantTeamAccess() {
	s.Run("should grant access to team and nil error on success", func() {
		s.setup()

		testUserId := "test_user_id"

		body := make(map[string][]string)
		body["userIds"] = append(body["userIds"], testUserId)

		var teamObj *shield.Team
		teamObj = new(shield.Team)
		teamObj.ID = "test_team_id"

		role := "users"

		responseJson := `{
			"users": [
				{
					"id": "test_user_id",
					"name": "Test_User",
					"slug": "Test_User",
					"email": "test.user@email.com",
					"metadata": {
						"slack": "@Test_Admin"
					},
					"createdAt": "2022-03-17T09:43:12.391071Z",
					"updatedAt": "2022-03-17T09:43:12.391071Z"
				}
			]
		}`

		responseUsers := http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(responseJson)))}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&responseUsers, nil).Once()

		actualError := s.client.GrantTeamAccess(teamObj, testUserId, role)
		s.Nil(actualError)
	})
}

func (s *ClientTestSuite) TestGrantProjectAccess() {
	s.Run("should grant access to project and nil error on success", func() {
		s.setup()

		testUserId := "test_user_id"

		body := make(map[string][]string)
		body["userIds"] = append(body["userIds"], testUserId)

		var projectObj *shield.Project
		projectObj = new(shield.Project)
		projectObj.ID = "test_project_id"

		role := "admins"

		responseJson := `{
			"users": [
				{
					"id": "test_user_id",
					"name": "Test_User",
					"slug": "Test_User",
					"email": "test.user@email.com",
					"metadata": {
						"slack": "@Test_Admin"
					},
					"createdAt": "2022-03-17T09:43:12.391071Z",
					"updatedAt": "2022-03-17T09:43:12.391071Z"
				}
			]
		}`

		responseUsers := http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(responseJson)))}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&responseUsers, nil).Once()

		actualError := s.client.GrantProjectAccess(projectObj, testUserId, role)
		s.Nil(actualError)
	})
}

func (s *ClientTestSuite) TestGrantOrganizationAccess() {
	s.Run("should grant access to organization and nil error on success", func() {
		s.setup()
		testUserId := "test_user_id"

		body := make(map[string][]string)
		body["userIds"] = append(body["userIds"], testUserId)

		var orgObj *shield.Organization
		orgObj = new(shield.Organization)
		orgObj.ID = "test_org_id"

		role := "admins"

		responseJson := `{
			"users": [
				{
					"id": "test_user_id",
					"name": "Test_User",
					"slug": "Test_User",
					"email": "test.user@email.com",
					"metadata": {
						"slack": "@Test_Admin"
					},
					"createdAt": "2022-03-17T09:43:12.391071Z",
					"updatedAt": "2022-03-17T09:43:12.391071Z"
				}
			]
		}`

		responseUsers := http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(responseJson)))}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&responseUsers, nil).Once()

		actualError := s.client.GrantOrganizationAccess(orgObj, testUserId, role)
		s.Nil(actualError)
	})
}

func (s *ClientTestSuite) TestRevokeTeamAccess() {
	s.Run("should revoke access to team and nil error on success", func() {
		s.setup()
		testUserId := "test_user_id"

		body := make(map[string][]string)
		body["userIds"] = append(body["userIds"], testUserId)

		var teamObj *shield.Team
		teamObj = new(shield.Team)
		teamObj.ID = "test_team_id"

		role := "users"

		responseJson := `{
			"message": "Removed User from group"
		}`

		responseUsers := http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(responseJson)))}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&responseUsers, nil).Once()

		actualError := s.client.RevokeTeamAccess(teamObj, testUserId, role)
		s.Nil(actualError)
	})
}

func (s *ClientTestSuite) TestRevokeProjectAccess() {
	s.Run("should revoke access to project and nil error on success", func() {
		s.setup()

		testUserId := "test_user_id"

		body := make(map[string][]string)
		body["userIds"] = append(body["userIds"], testUserId)

		var projectObj *shield.Project
		projectObj = new(shield.Project)
		projectObj.ID = "test_project_id"

		role := "admins"

		responseJson := `{
			"message": "Removed User from group"
		}`

		responseUsers := http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(responseJson)))}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&responseUsers, nil).Once()

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

		var orgObj *shield.Organization
		orgObj = new(shield.Organization)
		orgObj.ID = "test_org_id"

		role := "admins"

		responseJson := `{
			"message": "Removed User from group"
		}`

		responseUsers := http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(responseJson)))}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&responseUsers, nil).Once()

		actualError := s.client.RevokeOrganizationAccess(orgObj, testUserId, role)
		s.Nil(actualError)
	})
}

func (s *ClientTestSuite) TestGetSelfUser() {
	s.Run("Should return error user on empty email", func() {
		s.setup()
		testUserEmail := ""

		testGetSelfRequest, err := s.getTestRequest(http.MethodGet, "/admin/v1beta1/users/self", nil, testUserEmail)
		s.Require().NoError(err)

		responseJson := `{
			"code": 2,
			"message": "email id is empty",
			"details": []
		}`

		responseUser := http.Response{StatusCode: 500, Body: io.NopCloser(bytes.NewReader([]byte(responseJson)))}
		s.mockHttpClient.On("Do", testGetSelfRequest).Return(&responseUser, nil).Once()

		_, actualError := s.client.GetSelfUser(testUserEmail)
		s.NotNil(actualError)
	})
	s.Run("Should return shield user on success", func() {
		s.setup()
		testUserEmail := "test_user@email.com"

		testGetSelfRequest, err := s.getTestRequest(http.MethodGet, "/admin/v1beta1/users/self", nil, testUserEmail)
		s.Require().NoError(err)

		responseJson := `{
			 "user": {
				"id": "test-user-id",
				"name": "test-user",
				"slug": "",
				"email": "test_user@email.com",
				"metadata": {
					"slack": "test"
				},
				"createdAt": "2022-05-05T05:32:42.384021Z",
				"updatedAt": "2022-06-01T10:08:02.688055Z"
			}
		}`

		responseUser := http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(responseJson)))}
		s.mockHttpClient.On("Do", testGetSelfRequest).Return(&responseUser, nil).Once()

		expectedUser := &shield.User{
			ID:    "test-user-id",
			Name:  "test-user",
			Email: "test_user@email.com",
		}

		user, actualError := s.client.GetSelfUser(testUserEmail)
		s.EqualValues(expectedUser, user)
		s.Nil(actualError)
	})
}
