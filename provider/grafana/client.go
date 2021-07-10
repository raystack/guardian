package grafana

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-playground/validator/v10"
)

type ClientConfig struct {
	Host   string `validate:"required,url" mapstructure:"host"`
	ApiKey string `validate:"required" mapstructure:"api_key"`
}

type client struct {
	baseURL *url.URL

	apiKey string

	httpClient *http.Client
}

func NewClient(config *ClientConfig) (*client, error) {
	if err := validator.New().Struct(config); err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(config.Host)
	if err != nil {
		return nil, err
	}

	apiKey := config.ApiKey

	c := &client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}

	return c, nil
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
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	return req, nil
}

func (c *client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(v)
	return resp, err
}

func (c *client) getFolders() ([]*Folder, error) {
	req, err := c.newRequest(http.MethodGet, "/api/folders", nil)
	if err != nil {
		return nil, err
	}

	var folders []*Folder
	if _, err := c.do(req, &folders); err != nil {
		return nil, err
	}
	return folders, nil
}

func (c *client) getDashboards(folderId uint) ([]*Dashboard, error) {
	url := fmt.Sprintf("/api/search?folderIds=%d&type=dash-db", folderId)
	req, err := c.newRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var dashboard []*Dashboard
	if _, err := c.do(req, &dashboard); err != nil {
		return nil, err
	}

	return dashboard, nil
}
