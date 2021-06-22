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
	Host     string `validate:"required,url" mapstructure:"host"`
	Username string `validate:"required" mapstructure:"username"`
	Password string `validate:"required" mapstructure:"password"`
}

type sessionRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type sessionResponse struct {
	ID string `json:"id"`
}

type client struct {
	baseURL *url.URL

	username     string
	password     string
	sessionToken string

	httpClient *http.Client
}

func newClient(config *ClientConfig) (*client, error) {
	if err := validator.New().Struct(config); err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(config.Host)
	if err != nil {
		return nil, err
	}

	c := &client{
		baseURL:    baseURL,
		username:   config.Username,
		password:   config.Password,
		httpClient: &http.Client{},
	}

	sessionToken, err := c.getSessionToken()
	if err != nil {
		return nil, err
	}
	c.sessionToken = sessionToken

	return c, nil
}

func (c *client) GetDatabases() ([]*Database, error) {
	req, err := c.newRequest(http.MethodGet, "/api/database", nil)
	if err != nil {
		return nil, err
	}

	var databases []*Database
	if _, err := c.do(req, &databases); err != nil {
		return nil, err
	}

	return databases, nil
}

func (c *client) GetCollections() ([]*Collection, error) {
	req, err := c.newRequest(http.MethodGet, "/api/collection", nil)
	if err != nil {
		return nil, err
	}

	var collection []*Collection
	if _, err := c.do(req, &collection); err != nil {
		return nil, err
	}

	return collection, nil
}

func (c *client) getSessionToken() (string, error) {
	sessionRequest := &sessionRequest{
		Username: c.username,
		Password: c.password,
	}
	req, err := c.newRequest(http.MethodPost, "/api/session", sessionRequest)
	if err != nil {
		return "", nil
	}

	var sessionResponse sessionResponse
	if _, err := c.do(req, &sessionResponse); err != nil {
		return "", nil
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

	err = json.NewDecoder(resp.Body).Decode(v)
	return resp, err
}
