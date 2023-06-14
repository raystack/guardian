package identities

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mcuadros/go-defaults"
	"github.com/raystack/guardian/domain"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

var ErrFailedRequest = errors.New("request failed")

const UserIDWildcard = "{user_id}"

type HTTPAuthConfig struct {
	Type string `mapstructure:"type" json:"type" yaml:"type" validate:"required,oneof=basic api_key bearer google_idtoken"`

	// basic auth
	Username string `mapstructure:"username,omitempty" json:"username,omitempty" yaml:"username,omitempty" validate:"required_if=Type basic"`
	Password string `mapstructure:"password,omitempty" json:"password,omitempty" yaml:"password,omitempty" validate:"required_if=Type basic"`

	// api key
	In    string `mapstructure:"in,omitempty" json:"in,omitempty" yaml:"in,omitempty" validate:"required_if=Type api_key,omitempty,oneof=query header"`
	Key   string `mapstructure:"key,omitempty" json:"key,omitempty" yaml:"key,omitempty" validate:"required_if=Type api_key"`
	Value string `mapstructure:"value,omitempty" json:"value,omitempty" yaml:"value,omitempty" validate:"required_if=Type api_key"`

	// bearer
	Token string `mapstructure:"token,omitempty" json:"token,omitempty" yaml:"token,omitempty" validate:"required_if=Type bearer"`

	// google_idtoken
	Audience string `mapstructure:"audience,omitempty" json:"audience,omitempty" yaml:"audience,omitempty" validate:"required_if=Type google_idtoken"`
	// TODO: allow base64 encoded credentials
	CredentialsJSON string `mapstructure:"credentials_json,omitempty" json:"credentials_json,omitempty" yaml:"credentials_json,omitempty" validate:"required_if=Type google_idtoken"`
}

// HTTPClientConfig is the configuration required by iam.Client
type HTTPClientConfig struct {
	URL     string            `mapstructure:"url" json:"url" yaml:"url" validate:"required,url"`
	Headers map[string]string `mapstructure:"headers,omitempty" json:"headers,omitempty" yaml:"headers,omitempty"`
	Auth    *HTTPAuthConfig   `mapstructure:"auth,omitempty" json:"auth,omitempty" yaml:"auth,omitempty" validate:"omitempty,dive"`

	HTTPClient *http.Client `mapstructure:"-" json:"-" yaml:"-"`
	validator  *validator.Validate
	crypto     domain.Crypto
}

func (c *HTTPClientConfig) Validate() error {
	return c.validator.Struct(c)
}

func (c *HTTPClientConfig) Encrypt() error {
	if c.Auth != nil {
		if c.Auth.Password != "" {
			encryptedValue, err := c.crypto.Encrypt(c.Auth.Password)
			if err != nil {
				return err
			}
			c.Auth.Password = encryptedValue
		}

		if c.Auth.Value != "" {
			encryptedValue, err := c.crypto.Encrypt(c.Auth.Value)
			if err != nil {
				return err
			}
			c.Auth.Value = encryptedValue
		}

		if c.Auth.Token != "" {
			encryptedValue, err := c.crypto.Encrypt(c.Auth.Token)
			if err != nil {
				return err
			}
			c.Auth.Token = encryptedValue
		}

		if c.Auth.CredentialsJSON != "" {
			encryptedValue, err := c.crypto.Encrypt(c.Auth.CredentialsJSON)
			if err != nil {
				return err
			}
			c.Auth.CredentialsJSON = encryptedValue
		}
	}

	return nil
}

func (c *HTTPClientConfig) Decrypt() error {
	if c.Auth != nil {
		if c.Auth.Password != "" {
			decryptedValue, err := c.crypto.Decrypt(c.Auth.Password)
			if err != nil {
				return err
			}
			c.Auth.Password = decryptedValue
		}

		if c.Auth.Value != "" {
			decryptedValue, err := c.crypto.Decrypt(c.Auth.Value)
			if err != nil {
				return err
			}
			c.Auth.Value = decryptedValue
		}

		if c.Auth.Token != "" {
			decryptedValue, err := c.crypto.Decrypt(c.Auth.Token)
			if err != nil {
				return err
			}
			c.Auth.Token = decryptedValue
		}

		if c.Auth.CredentialsJSON != "" {
			decryptedValue, err := c.crypto.Decrypt(c.Auth.CredentialsJSON)
			if err != nil {
				return err
			}
			c.Auth.CredentialsJSON = decryptedValue
		}
	}

	return nil
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

	if config.Auth.Type == "google_idtoken" {
		ctx := context.Background()
		ts, err := idtoken.NewTokenSource(ctx, config.Auth.Audience, idtoken.WithCredentialsJSON([]byte(config.Auth.CredentialsJSON)))
		if err != nil {
			return nil, err
		}
		httpClient = oauth2.NewClient(ctx, ts)
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
