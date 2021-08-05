package tableau

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
	ContentURL string `validate:"required" mapstructure:"content_url"`
}

type sessionRequest struct {
	Credentials requestCredentials `json:"credentials"`
}

type requestCredentials struct {
	Site     requestSite `json:"site"`
	Name     string      `json:"name"`
	Password string      `json:"password"`
}

type requestSite struct {
	ContentURL string `json:"contentUrl"`
}

type sessionResponse struct {
	Credentials responseCredentials `json:"credentials"`
}

type responseCredentials struct {
	Site  responseSite `json:"site"`
	User  responseUser `json:"user"`
	Token string       `json:"token"`
}

type responseUser struct {
	ID string `json:"id"`
}

type responseSite struct {
	ID         string `json:"id"`
	ContentURL string `json:"contentUrl"`
}

type client struct {
	baseURL *url.URL

	username     string
	password     string
	contentUrl   string
	sessionToken string
	siteID       string
	userID       string

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
		contentUrl: config.ContentURL,
		httpClient: &http.Client{},
	}

	sessionToken, siteID, userID, err := c.getSession()
	if err != nil {
		return nil, err
	}
	c.sessionToken = sessionToken
	c.siteID = siteID
	c.userID = userID
	return c, nil
}

func (c *client) getSession() (string, string, string, error) {
	sessionRequest := &sessionRequest{
		Credentials: requestCredentials{
			Name:     c.username,
			Password: c.password,
			Site: requestSite{
				ContentURL: c.contentUrl,
			},
		},
	}

	req, err := c.newRequest(http.MethodPost, "/api/3.12/auth/signin", sessionRequest)
	if err != nil {
		return "", "", "", nil
	}

	var sessionResponse sessionResponse
	if _, err := c.do(req, &sessionResponse); err != nil {
		return "", "", "", nil
	}

	return sessionResponse.Credentials.Token, sessionResponse.Credentials.Site.ID, sessionResponse.Credentials.User.ID, nil
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
	req.Header.Set("X-Tableau-Auth", c.sessionToken)
	return req, nil
}

func (c *client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		newSessionToken, _, _, err := c.getSession()
		if err != nil {
			return nil, err
		}
		c.sessionToken = newSessionToken
		req.Header.Set("X-Tableau-Auth", c.sessionToken)

		// re-do the request
		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
	}

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
	}
	return resp, err
}
