package metabase

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/go-playground/validator/v10"
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
