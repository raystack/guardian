package iam

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

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
	GetManagersURL string `validate:"required,url" mapstructure:"get_managers_url"`
	HTTPClient     *http.Client

	Auth *HTTPAuthConfig `mapstructure:"auth" validate:"omitempty,dive"`
}

type managerEmailsResponse struct {
	Emails []string `json:"emails"`
}

// HTTPClient wraps the http client for external approver resolver service
type HTTPClient struct {
	getManagersURL string
	httpClient     *http.Client
	auth           *HTTPAuthConfig
}

// NewHTTPClient returns *iam.Client
func NewHTTPClient(config *HTTPClientConfig) (*HTTPClient, error) {
	if err := validator.New().Struct(config); err != nil {
		return nil, err
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &HTTPClient{
		getManagersURL: config.GetManagersURL,
		httpClient:     httpClient,
		auth:           config.Auth,
	}, nil
}

// GetManagerEmails fetches to external approver resolver service and returns approver emails
func (c *HTTPClient) GetManagerEmails(user string) ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, c.getManagersURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("user", user)
	req.URL.RawQuery = q.Encode()

	var approvers managerEmailsResponse
	if err := c.sendRequest(req, &approvers); err != nil {
		return nil, err
	}

	return approvers.Emails, nil
}

func (c *HTTPClient) sendRequest(req *http.Request, v interface{}) error {
	c.setAuth(req)
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if err := json.NewDecoder(res.Body).Decode(v); err != nil {
		return err
	}

	return nil
}

func (c *HTTPClient) setAuth(req *http.Request) {
	if c.auth != nil {
		switch c.auth.Type {
		case "basic":
			req.SetBasicAuth(c.auth.Username, c.auth.Password)
		case "api_key":
			switch c.auth.In {
			case "query":
				q := req.URL.Query()
				q.Add(c.auth.Key, c.auth.Value)
				req.URL.RawQuery = q.Encode()
			case "header":
				req.Header.Add(c.auth.Key, c.auth.Value)
			default:
			}
		case "bearer":
			req.Header.Add("Authorization", "Bearer "+c.auth.Token)
		default:
		}
	}
}
