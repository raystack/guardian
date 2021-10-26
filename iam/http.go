package iam

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mcuadros/go-defaults"
)

var ErrFailedRequest = errors.New("request failed")

const UserIDWildcard = "{user_id}"

type HTTPAuthConfig struct {
	Type string `mapstructure:"type" validate:"required,oneof=basic api_key bearer"`

	// basic auth
	Username string `mapstructure:"username" validate:"required_if=Type basic"`
	Password string `mapstructure:"password" validate:"required_if=Type basic"`

	// api key
	In    string `mapstructure:"in" validate:"required_if=Type api_key,omitempty,oneof=query header"`
	Key   string `mapstructure:"key" validate:"required_if=Type api_key"`
	Value string `mapstructure:"value" validate:"required_if=Type api_key"`

	// bearer
	Token string `mapstructure:"token" validate:"required_if=Type bearer"`
}

// HTTPClientConfig is the configuration required by iam.Client
type HTTPClientConfig struct {
	HTTPClient *http.Client

	URL     string            `mapstructure:"url" validate:"required,url"`
	Headers map[string]string `mapstructure:"headers"`
	Auth    *HTTPAuthConfig   `mapstructure:"auth" validate:"omitempty,dive"`
}

// HTTPClient wraps the http client for external approver resolver service
type HTTPClient struct {
	httpClient *http.Client
	config     *HTTPClientConfig

	url string
}

// NewHTTPClient returns *iam.Client
func NewHTTPClient(config *HTTPClientConfig) (*HTTPClient, error) {
	defaults.SetDefaults(config)
	if err := validator.New().Struct(config); err != nil {
		return nil, err
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &HTTPClient{
		httpClient: httpClient,
		config:     config,
		url:        config.URL,
	}, nil
}

// GetUser fetches user details to external
func (c *HTTPClient) GetUser(userID string) (interface{}, error) {
	req, err := c.createRequest(userID)
	if err != nil {
		return nil, err
	}

	var res interface{}
	if err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return res, nil
}

func (c *HTTPClient) createRequest(userID string) (*http.Request, error) {
	url := strings.Replace(c.config.URL, UserIDWildcard, userID, -1)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range c.config.Headers {
		if strings.Contains(v, UserIDWildcard) {
			req.Header.Set(k, strings.Replace(v, UserIDWildcard, userID, -1))
		} else {
			req.Header.Set(k, v)
		}
	}
	c.setAuth(req)

	req.Header.Set("Accept", "application/json")
	return req, nil
}

func (c *HTTPClient) sendRequest(req *http.Request, v interface{}) error {
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return json.NewDecoder(res.Body).Decode(v)
	}

	return fmt.Errorf("%w: %s", ErrFailedRequest, res.Status)
}

func (c *HTTPClient) setAuth(req *http.Request) {
	if c.config.Auth != nil {
		switch c.config.Auth.Type {
		case "basic":
			req.SetBasicAuth(c.config.Auth.Username, c.config.Auth.Password)
		case "api_key":
			switch c.config.Auth.In {
			case "query":
				q := req.URL.Query()
				q.Add(c.config.Auth.Key, c.config.Auth.Value)
				req.URL.RawQuery = q.Encode()
			case "header":
				req.Header.Add(c.config.Auth.Key, c.config.Auth.Value)
			default:
			}
		case "bearer":
			req.Header.Add("Authorization", "Bearer "+c.config.Auth.Token)
		default:
		}
	}
}
