package iam

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-playground/validator/v10"
	"github.com/odpf/guardian/domain"
)

type ShieldClientConfig struct {
	Host string `mapstructure:"host" json:"host" yaml:"host" validate:"required,url"`

	validator *validator.Validate
	crypto    domain.Crypto
}

func (c *ShieldClientConfig) Validate() error {
	return c.validator.Struct(c)
}

func (c *ShieldClientConfig) Encrypt() error {
	return nil
}

func (c *ShieldClientConfig) Decrypt() error {
	return nil
}

type role struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayname"`
}

type policy struct {
	Subject  map[string]string `json:"subject"`
	Resource map[string]string `json:"resource"`
	Action   map[string]string `json:"action"`
}

type group struct {
	ID           string   `json:"id"`
	IsMember     bool     `json:"isMember"`
	UserPolicies []policy `json:"userPolicies"`
}

type user struct {
	ID       string            `json:"id"`
	Username string            `json:"username"`
	Metadata map[string]string `json:"metadata"`

	TeamLeads []string `json:"team_leads"`
}

type shieldClient struct {
	baseURL *url.URL

	teamAdminRoleID string
	userEmail       string
	users           map[string]user

	httpClient *http.Client
}

func NewShieldClient(config *ShieldClientConfig) (*shieldClient, error) {
	if err := validator.New().Struct(config); err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(config.Host)
	if err != nil {
		return nil, err
	}

	return &shieldClient{
		baseURL:    baseURL,
		httpClient: http.DefaultClient,
	}, nil
}

func (c *shieldClient) GetUser(userEmail string) (interface{}, error) {
	c.userEmail = userEmail

	if c.teamAdminRoleID == "" {
		roles, err := c.getRoles()
		if err != nil {
			return nil, err
		}

		for _, r := range roles {
			if r.DisplayName == "Team Admin" {
				c.teamAdminRoleID = r.ID
				break
			}
		}
		if c.teamAdminRoleID == "" {
			return nil, errors.New("team admin role id not found")
		}
	}

	groups, err := c.getGroups()
	if err != nil {
		return nil, err
	}

	users, err := c.getUsers()
	if err != nil {
		return nil, err
	}
	if c.users == nil {
		c.users = map[string]user{}
	}

	var userDetails user
	for _, u := range users {
		c.users[u.ID] = u
		if u.Metadata["email"] == c.userEmail {
			userDetails = u
		}
	}

	var teamLeadEmails []string
	for _, g := range groups {
		if g.IsMember {
			for _, p := range g.UserPolicies {
				teamLeadID := p.Subject["user"]
				teamLeadEmail := c.users[teamLeadID].Metadata["email"]
				teamLeadEmails = append(teamLeadEmails, teamLeadEmail)
			}
		}
	}

	userDetails.TeamLeads = teamLeadEmails

	jsonBytes, err := json.Marshal(userDetails)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *shieldClient) getRoles() ([]role, error) {
	req, err := c.newRequest(http.MethodGet, "/api/roles", nil)
	if err != nil {
		return nil, err
	}

	var roles []role
	_, err = c.do(req, &roles)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (c *shieldClient) getGroups() ([]group, error) {
	url := fmt.Sprintf("/api/groups?user_role=%s", c.teamAdminRoleID)
	req, err := c.newRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var groups []group
	_, err = c.do(req, &groups)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (c *shieldClient) getUsers() ([]user, error) {
	req, err := c.newRequest(http.MethodGet, "/api/users", nil)
	if err != nil {
		return nil, err
	}

	var users []user
	_, err = c.do(req, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (c *shieldClient) newRequest(method, path string, body interface{}) (*http.Request, error) {
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
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Goog-Authenticated-User-Email", c.userEmail)
	return req, nil
}

func (c *shieldClient) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(v)
	return resp, err
}
