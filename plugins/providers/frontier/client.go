package frontier

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/raystack/guardian/pkg/tracing"
	"github.com/raystack/salt/log"
)

const (
	groupsEndpoint       = "/v1beta1/groups"
	projectsEndpoint     = "/v1beta1/projects"
	organizationEndpoint = "/v1beta1/organizations"
	selfUserEndpoint     = "/v1beta1/users/self"

	groupsConst        = "groups"
	projectsConst      = "projects"
	organizationsConst = "organizations"
	usersConst         = "users"
	userConst          = "user"
)

type successAccess interface{}

type Client interface {
	GetTeams() ([]*Team, error)
	GetProjects() ([]*Project, error)
	GetOrganizations() ([]*Organization, error)
	GrantTeamAccess(team *Team, userId string, role string) error
	RevokeTeamAccess(team *Team, userId string, role string) error
	GrantProjectAccess(project *Project, userId string, role string) error
	RevokeProjectAccess(project *Project, userId string, role string) error
	GrantOrganizationAccess(organization *Organization, userId string, role string) error
	RevokeOrganizationAccess(organization *Organization, userId string, role string) error
	GetSelfUser(email string) (*User, error)
}

type client struct {
	baseURL *url.URL

	authHeader string
	authEmail  string

	httpClient HTTPClient
	logger     log.Logger
}

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type ClientConfig struct {
	Host       string `validate:"required,url" mapstructure:"host"`
	AuthHeader string `validate:"required" mapstructure:"auth_header"`
	AuthEmail  string `validate:"required" mapstructure:"auth_email"`
	HTTPClient HTTPClient
}

func NewClient(config *ClientConfig, logger log.Logger) (*client, error) {
	if err := validator.New().Struct(config); err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(config.Host)
	if err != nil {
		return nil, err
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = tracing.NewHttpClient("ShieldHttpClient")
	}

	c := &client{
		baseURL:    baseURL,
		authHeader: config.AuthHeader,
		authEmail:  config.AuthEmail,
		httpClient: httpClient,
		logger:     logger,
	}

	return c, nil
}

func (c *client) newRequest(method, path string, body interface{}, authEmail string) (*http.Request, error) {
	u, err := c.baseURL.Parse(path)
	if err != nil {
		return nil, err
	}
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if authEmail == "" {
		req.Header.Set(c.authHeader, c.authEmail)
	} else {
		req.Header.Set(c.authHeader, authEmail)
	}
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func (c *client) GetAdminsOfGivenResourceType(id string, resourceTypeEndPoint string) ([]string, error) {
	endPoint := path.Join(resourceTypeEndPoint, "/", id, "/admins")
	req, err := c.newRequest(http.MethodGet, endPoint, nil, "")
	if err != nil {
		return nil, err
	}

	var users []*User
	var response interface{}
	if _, err := c.do(req, &response); err != nil {
		return nil, err
	}
	if v, ok := response.(map[string]interface{}); ok && v[usersConst] != nil {
		err = mapstructure.Decode(v[usersConst], &users)
	}

	var userEmails []string
	for _, user := range users {
		userEmails = append(userEmails, user.Email)
	}

	return userEmails, err
}

func (c *client) GetTeams() ([]*Team, error) {
	req, err := c.newRequest(http.MethodGet, groupsEndpoint, nil, "")
	if err != nil {
		return nil, err
	}

	var teams []*Team
	var response interface{}
	if _, err := c.do(req, &response); err != nil {
		return nil, err
	}

	if v, ok := response.(map[string]interface{}); ok && v[groupsConst] != nil {
		err = mapstructure.Decode(v[groupsConst], &teams)
	}

	for _, team := range teams {
		admins, err := c.GetAdminsOfGivenResourceType(team.ID, groupsEndpoint)
		if err != nil {
			return nil, err
		}
		team.Admins = admins
	}

	c.logger.Info("Fetch teams from request", "total", len(teams), req.URL)

	return teams, err
}

func (c *client) GetProjects() ([]*Project, error) {
	req, err := c.newRequest(http.MethodGet, projectsEndpoint, nil, "")
	if err != nil {
		return nil, err
	}

	var projects []*Project
	var response interface{}

	if _, err := c.do(req, &response); err != nil {
		return nil, err
	}

	if v, ok := response.(map[string]interface{}); ok && v[projectsConst] != nil {
		err = mapstructure.Decode(v[projectsConst], &projects)
	}

	for _, project := range projects {
		admins, err := c.GetAdminsOfGivenResourceType(project.ID, projectsEndpoint)
		if err != nil {
			return nil, err
		}
		project.Admins = admins
	}

	c.logger.Info("Fetch projects from request", "total", len(projects), req.URL)

	return projects, err
}

func (c *client) GetOrganizations() ([]*Organization, error) {
	req, err := c.newRequest(http.MethodGet, organizationEndpoint, nil, "")
	if err != nil {
		return nil, err
	}

	var organizations []*Organization
	var response interface{}
	if _, err := c.do(req, &response); err != nil {
		return nil, err
	}

	if v, ok := response.(map[string]interface{}); ok && v[organizationsConst] != nil {
		err = mapstructure.Decode(v[organizationsConst], &organizations)
	}

	for _, org := range organizations {
		admins, err := c.GetAdminsOfGivenResourceType(org.ID, organizationEndpoint)
		if err != nil {
			return nil, err
		}
		org.Admins = admins
	}

	c.logger.Info("Fetch organizations from request", "total", len(organizations), req.URL)

	return organizations, err
}

func (c *client) GrantTeamAccess(resource *Team, userId string, role string) error {
	body := make(map[string][]string)
	body["userIds"] = append(body["userIds"], userId)

	endPoint := path.Join(groupsEndpoint, "/", resource.ID, "/", role)
	req, err := c.newRequest(http.MethodPost, endPoint, body, "")
	if err != nil {
		return err
	}

	var users []*User
	var response interface{}
	if _, err := c.do(req, &response); err != nil {
		return err
	}

	if v, ok := response.(map[string]interface{}); ok && v[usersConst] != nil {
		err = mapstructure.Decode(v[usersConst], &users)
		if err != nil {
			return err
		}
	}

	c.logger.Info("Team access to the user,", "total users", len(users), req.URL)

	return nil
}

func (c *client) GrantProjectAccess(resource *Project, userId string, role string) error {
	body := make(map[string][]string)
	body["userIds"] = append(body["userIds"], userId)

	endPoint := path.Join(projectsEndpoint, "/", resource.ID, "/", role)
	req, err := c.newRequest(http.MethodPost, endPoint, body, "")
	if err != nil {
		return err
	}

	var users []*User
	var response interface{}
	if _, err := c.do(req, &response); err != nil {
		return err
	}

	if v, ok := response.(map[string]interface{}); ok && v[usersConst] != nil {
		err = mapstructure.Decode(v[usersConst], &users)
		if err != nil {
			return err
		}
	}

	c.logger.Info("Project access to the user,", "total users", len(users), req.URL)
	return nil
}

func (c *client) GrantOrganizationAccess(resource *Organization, userId string, role string) error {
	body := make(map[string][]string)
	body["userIds"] = append(body["userIds"], userId)

	endPoint := path.Join(organizationEndpoint, "/", resource.ID, "/", role)
	req, err := c.newRequest(http.MethodPost, endPoint, body, "")

	if err != nil {
		return err
	}

	var users []*User
	var response interface{}
	if _, err := c.do(req, &response); err != nil {
		return err
	}

	if v, ok := response.(map[string]interface{}); ok && v[usersConst] != nil {
		err = mapstructure.Decode(v[usersConst], &users)
		if err != nil {
			return err
		}
	}

	c.logger.Info("Organization access to the user,", "total users", len(users), req.URL)
	return nil
}

func (c *client) RevokeTeamAccess(resource *Team, userId string, role string) error {
	endPoint := path.Join(groupsEndpoint, "/", resource.ID, "/", role, "/", userId)
	req, err := c.newRequest(http.MethodDelete, endPoint, "", "")
	if err != nil {
		return err
	}

	var success successAccess
	var response interface{}
	if _, err := c.do(req, &response); err != nil {
		return err
	}

	if v, ok := response.(map[string]interface{}); ok && v != nil {
		err = mapstructure.Decode(v, &success)
		if err != nil {
			return err
		}
	}

	c.logger.Info("Remove access of the user from team,", "Users", userId, req.URL)
	return nil
}

func (c *client) RevokeProjectAccess(resource *Project, userId string, role string) error {
	endPoint := path.Join(projectsEndpoint, "/", resource.ID, "/", role, "/", userId)
	req, err := c.newRequest(http.MethodDelete, endPoint, "", "")
	if err != nil {
		return err
	}

	var success successAccess
	var response interface{}
	if _, err := c.do(req, &response); err != nil {
		return err
	}

	if v, ok := response.(map[string]interface{}); ok && v != nil {
		err = mapstructure.Decode(v, &success)
		if err != nil {
			return err
		}
	}

	c.logger.Info("Remove access of the user from project", "Users", userId, req.URL)
	return nil
}

func (c *client) RevokeOrganizationAccess(resource *Organization, userId string, role string) error {
	endPoint := path.Join(organizationEndpoint, "/", resource.ID, "/", role, "/", userId)
	req, err := c.newRequest(http.MethodDelete, endPoint, "", "")
	if err != nil {
		return err
	}

	var success successAccess
	var response interface{}
	if _, err := c.do(req, &response); err != nil {
		return err
	}

	if v, ok := response.(map[string]interface{}); ok && v != nil {
		err = mapstructure.Decode(v, &success)
		if err != nil {
			return err
		}
	}

	c.logger.Info("Remove access of the user from organization", "Users", userId, req.URL)
	return nil
}

func (c *client) GetSelfUser(email string) (*User, error) {
	req, err := c.newRequest(http.MethodGet, selfUserEndpoint, nil, email)
	if err != nil {
		return nil, err
	}

	var user *User
	var response interface{}
	if _, err := c.do(req, &response); err != nil {
		return nil, err
	}

	if v, ok := response.(map[string]interface{}); ok && v[userConst] != nil {
		err = mapstructure.Decode(v[userConst], &user)
	}

	c.logger.Info("Fetch user from request", "Id", user.ID, req.URL)

	return user, err
}

func (c *client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error(fmt.Sprintf("Failed to execute request %v with error %v", req.URL, err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusInternalServerError {
		byteData, _ := io.ReadAll(resp.Body)
		return nil, errors.New(string(byteData))
	}

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
	}

	return resp, err
}
