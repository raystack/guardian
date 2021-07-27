package grafana

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-playground/validator/v10"
)

type GrafanaClient interface {
	GetDashboards(folderId int) ([]*Dashboard, error)
	GetFolders() ([]*Folder, error)
	GrantDashboardAccess(resource *Dashboard, user, role string) error
	RevokeDashboardAccess(resource *Dashboard, user, role string) error
}

type ClientConfig struct {
	Host       string `validate:"required,url" mapstructure:"host"`
	Username   string `validate:"required" mapstructure:"username"`
	Password   string `validate:"required" mapstructure:"password"`
	Org        string `validate:"required" mapstructure:"org"`
	HTTPClient HTTPClient
}

type permission struct {
	UserID     int    `json:"userId,omitempty"`
	TeamID     int    `json:"teamId,omitempty"`
	Permission int    `json:"permission"`
	Role       string `json:"role,omitempty"`
	Inherited  bool   `json:"inherited,omitempty"`
}

type user struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

type updatePermissionRequest struct {
	Items []*permission `json:"items"`
}

type client struct {
	baseURL *url.URL

	username string
	password string
	org      string

	httpClient HTTPClient
}

func NewClient(config *ClientConfig) (*client, error) {
	if err := validator.New().Struct(config); err != nil {
		return nil, err
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	baseURL, err := url.Parse(config.Host)
	if err != nil {
		return nil, err
	}

	username := config.Username
	password := config.Password
	org := config.Org

	c := &client{
		baseURL:    baseURL,
		username:   username,
		password:   password,
		org:        org,
		httpClient: httpClient,
	}

	return c, nil
}

func (c *client) GrantDashboardAccess(resource *Dashboard, user, role string) error {
	userDetails, err := c.getUser(user)
	if err != nil {
		return err
	}

	permissionCode := permissionCodes[role]
	if permissionCode == 0 {
		return ErrInvalidPermissionType
	}

	permissions, err := c.getDashboardPermissions(resource.ID)
	if err != nil {
		return err
	}

	nonInheritedPermissions := []*permission{}
	isPermissionUpdated := false
	for _, permission := range permissions {
		if !permission.Inherited {
			p := permission
			if permission.UserID == userDetails.ID {
				p.Permission = permissionCode
				isPermissionUpdated = true
			}

			nonInheritedPermissions = append(nonInheritedPermissions, p)
		}
	}

	if !isPermissionUpdated {
		nonInheritedPermissions = append(nonInheritedPermissions, &permission{
			UserID:     userDetails.ID,
			Permission: permissionCode,
		})
	}
	return c.updateDashboardPermissions(resource.ID, nonInheritedPermissions)
}

func (c *client) RevokeDashboardAccess(resource *Dashboard, user, role string) error {
	userDetails, err := c.getUser(user)
	if err != nil {
		return err
	}
	permissionCode := permissionCodes[role]
	if permissionCode == 0 {
		return ErrInvalidPermissionType
	}

	permissions, err := c.getDashboardPermissions(resource.ID)
	if err != nil {
		return err
	}

	nonInheritedPermissions := []*permission{}
	isPermissionFound := false
	for _, permission := range permissions {
		if !permission.Inherited {
			p := permission
			if permission.UserID == userDetails.ID && permission.Permission == permissionCode {
				isPermissionFound = true
			} else {
				nonInheritedPermissions = append(nonInheritedPermissions, p)
			}
		}
	}

	if !isPermissionFound {
		return ErrPermissionNotFound
	}

	return c.updateDashboardPermissions(resource.ID, nonInheritedPermissions)
}

func (c *client) base64Encode() string {
	data := c.username + ":" + c.password
	basicKeyEncoded := b64.StdEncoding.EncodeToString([]byte(data))

	return basicKeyEncoded
}

func (c *client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	basicKey := c.base64Encode()
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
	req.Header.Set("Authorization", "Basic "+basicKey)
	req.Header.Set("X-Grafana-Org-Id", c.org)

	return req, nil
}

func (c *client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
	}
	return resp, err
}

func (c *client) GetFolders() ([]*Folder, error) {
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

func (c *client) GetDashboards(folderId int) ([]*Dashboard, error) {
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

func (c *client) getDashboardPermissions(id int) ([]*permission, error) {
	url := fmt.Sprintf("/api/dashboards/id/%d/permissions", id)
	req, err := c.newRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var permissions []*permission
	if _, err := c.do(req, &permissions); err != nil {
		return nil, err
	}

	return permissions, nil
}

func (c *client) updateDashboardPermissions(id int, permissions []*permission) error {
	body := updatePermissionRequest{
		Items: permissions,
	}
	url := fmt.Sprintf("/api/dashboards/id/%d/permissions", id)
	req, err := c.newRequest(http.MethodPost, url, body)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err

}

func (c *client) getUser(email string) (*user, error) {
	url := fmt.Sprintf("/api/users/lookup?loginOrEmail=%s", email)
	req, err := c.newRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var user *user
	res, err := c.do(req, &user)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, ErrUserNotFound
	}

	return user, nil
}
