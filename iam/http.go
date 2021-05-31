package iam

import (
	"encoding/json"
	"net/http"

	"github.com/odpf/guardian/domain"
)

// HTTPClient abstracts the http client
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// ClientConfig is the configuration required by iam.Client
type ClientConfig struct {
	URL        string
	HttpClient HTTPClient
}

// Client wraps the http client for external approver resolver service
type Client struct {
	url        string
	httpClient HTTPClient
}

// NewHTTPClient returns *iam.Client
func NewHTTPClient(config *ClientConfig) *Client {
	return &Client{
		url:        config.URL,
		httpClient: config.HttpClient,
	}
}

// GetUserApproverEmails fetches to external approver resolver service and returns approver emails
func (c *Client) GetUserApproverEmails(query map[string]string) ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, c.url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, v)
	}
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
