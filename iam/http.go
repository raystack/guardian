package iam

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/odpf/guardian/domain"
)

// HTTPClientConfig is the configuration required by iam.Client
type HTTPClientConfig struct {
	GetManagersURL string `validate:"required,url" mapstructure:"get_managers_url"`
	HTTPClient     *http.Client
}

// HTTPClient wraps the http client for external approver resolver service
type HTTPClient struct {
	getManagersURL string
	httpClient     *http.Client
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

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	var approvers domain.ApproversResponse
	if err := json.NewDecoder(res.Body).Decode(&approvers); err != nil {
		return nil, err
	}

	return approvers.Emails, nil
}
