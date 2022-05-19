package metabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"io"
	"net/http"
	"net/url"
)

type ClientConfig struct {
	Host       string `validate:"required,url" mapstructure:"host"`
	Username   string `validate:"required" mapstructure:"username"`
	Password   string `validate:"required" mapstructure:"password"`
	HTTPClient HTTPClient
}

type user struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	MembershipID int    `json:"membership_id"`
}

type group struct {
	ID      int    `json:"id,omitempty"`
	Name    string `json:"name"`
	Members []user `json:"members"`
}

type member struct {
	MembershipId int `json:"membership_id,omitempty"`
	GroupId      int `json:"group_id"`
	UserId       int `json:"user_id"`
}

type SessionRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SessionResponse struct {
	ID string `json:"id"`
}

type databasePermission struct {
	Native  string `json:"native,omitempty" mapstructure:"native"`
	Schemas string `json:"schemas" mapstructure:"schemas"`
}

type databaseGraph struct {
	Revision int `json:"revision"`
	// Groups is a map[group_id]map[database_id]databasePermission
	Groups map[string]map[string]databasePermission `json:"groups"`
}

type collectionGraph struct {
	Revision int `json:"revision"`
	// Groups is a map[group_id]map[database_id]role string
	Groups map[string]map[string]string `json:"groups"`
}

type membershipRequest struct {
	GroupID int `json:"group_id"`
	UserID  int `json:"user_id"`
}

var (
	databaseViewerPermission = databasePermission{
		Schemas: "all",
	}
	databaseEditorPermission = databasePermission{
		Schemas: "all",
		Native:  "write",
	}
)

type client struct {
	baseURL *url.URL

	username     string
	password     string
	sessionToken string

	httpClient HTTPClient

	userIDs map[string]int
}

func NewClient(config *ClientConfig) (*client, error) {
	if err := validator.New().Struct(config); err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(config.Host)
	if err != nil {
		return nil, err
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	c := &client{
		baseURL:    baseURL,
		username:   config.Username,
		password:   config.Password,
		httpClient: httpClient,
		userIDs:    map[string]int{},
	}

	sessionToken, err := c.getSessionToken()
	if err != nil {
		return nil, err
	}
	c.sessionToken = sessionToken

	return c, nil
}

func (c *client) getUsers() ([]user, error) {
	req, err := c.newRequest(http.MethodGet, "/api/user", nil)
	if err != nil {
		return nil, err
	}

	var users []user
	if _, err := c.do(req, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (c *client) getGroups() ([]group, error) {
	req, err := c.newRequest(http.MethodGet, "/api/permissions/group", nil)
	if err != nil {
		return nil, err
	}

	var groups []group

	if _, err := c.do(req, &groups); err != nil {
		return nil, err
	}

	return groups, nil
}

func (c *client) getMembership() (map[string][]member, error) {
	req, err := c.newRequest(http.MethodGet, "/api/permissions/membership", nil)
	if err != nil {
		return nil, err
	}

	var members map[string][]member

	if _, err := c.do(req, &members); err != nil {
		return nil, err
	}

	return members, nil
}

func (c *client) getSessionToken() (string, error) {
	sessionRequest := &SessionRequest{
		Username: c.username,
		Password: c.password,
	}
	req, err := c.newRequest(http.MethodPost, "/api/session", sessionRequest)
	if err != nil {
		return "", err
	}

	var sessionResponse SessionResponse
	if _, err := c.do(req, &sessionResponse); err != nil {
		return "", err
	}

	return sessionResponse.ID, nil
}

func (c *client) getCollectionAccess() (*collectionGraph, error) {
	req, err := c.newRequest(http.MethodGet, "/api/collection/graph", nil)
	if err != nil {
		return nil, err
	}

	var graph collectionGraph
	if _, err := c.do(req, &graph); err != nil {
		return nil, err
	}

	return &graph, nil
}

func (c *client) getDatabaseAccess() (*databaseGraph, error) {
	req, err := c.newRequest(http.MethodGet, "/api/permissions/graph", nil)
	if err != nil {
		return nil, err
	}

	var dbGraph databaseGraph
	if _, err := c.do(req, &dbGraph); err != nil {
		return nil, err
	}

	return &dbGraph, nil
}

func (c *client) getGroup(id int) (*group, error) {
	url := fmt.Sprintf("/api/permissions/group/%d", id)
	req, err := c.newRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var group group

	if _, err := c.do(req, &group); err != nil {
		return nil, err
	}

	return &group, nil
}

func (c *client) newRequest(method, path string, body interface{}) (*http.Request, error) {
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
	req.Header.Set("X-Metabase-Session", c.sessionToken)
	return req, nil
}

func (c *client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		newSessionToken, err := c.getSessionToken()
		if err != nil {
			return nil, err
		}
		c.sessionToken = newSessionToken
		req.Header.Set("X-Metabase-Session", c.sessionToken)

		// re-do the request
		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
	}

	if v != nil {
		//all, _ := ioutil.ReadAll(resp.Body)
		//fmt.Println(string(all))
		err = json.NewDecoder(resp.Body).Decode(v)
	}
	return resp, err
}
