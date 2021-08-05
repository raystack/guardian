package tableau

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mcuadros/go-defaults"
)

type TableauClient interface {
	GetWorkbooks() ([]*Workbook, error)
	UpdateSiteRole(user, role string) error
	GrantWorkbookAccess(resource *Workbook, user, role string) error
	RevokeWorkbookAccess(resource *Workbook, user, role string) error
}

type ClientConfig struct {
	Host       string `validate:"required,url" mapstructure:"host"`
	Username   string `validate:"required" mapstructure:"username"`
	Password   string `validate:"required" mapstructure:"password"`
	ContentURL string `validate:"required" mapstructure:"content_url"`
	APIVersion string `mapstructure:"apiVersion" default:"3.12"`
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
	apiVersion   string
	sessionToken string
	siteID       string
	userID       string

	httpClient *http.Client
}

type workbookPermissions struct {
	Permissions permissions `json:"permissions"`
}
type workbook struct {
	ID string `json:"id"`
}
type workbookUser struct {
	ID string `json:"id"`
}
type capability struct {
	Name string `json:"name"`
	Mode string `json:"mode"`
}
type capabilities struct {
	Capability []capability `json:"capability"`
}
type granteeCapabilities struct {
	User         workbookUser `json:"user"`
	Capabilities capabilities `json:"capabilities"`
}
type permissions struct {
	Workbook            workbook              `json:"workbook"`
	GranteeCapabilities []granteeCapabilities `json:"granteeCapabilities"`
}

type responseWorkbooks struct {
	Pagination pagination `json:"pagination"`
	Workbooks  workbooks  `json:"workbooks"`
}

type siteUsers struct {
	Pagination pagination    `json:"pagination"`
	Users      responseUsers `json:"users"`
}

type responseUsers struct {
	User []responseUser `json:"user"`
}

type workbooks struct {
	Workbook []*Workbook `json:"workbook"`
}
type pagination struct {
	PageNumber     string `json:"pageNumber"`
	PageSize       string `json:"pageSize"`
	TotalAvailable string `json:"totalAvailable"`
}

type userSiteRoleData struct {
	User userSiteRole `json:"user"`
}

type userSiteRole struct {
	SiteRole string `json:"siteRole"`
}

func (c *client) GetWorkbooks() ([]*Workbook, error) {
	url := fmt.Sprintf("/api/%v/sites/%v/workbooks", c.apiVersion, c.siteID)
	req, err := c.newRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var workbooks responseWorkbooks
	if _, err := c.do(req, &workbooks); err != nil {
		return nil, err
	}
	return workbooks.Workbooks.Workbook, nil
}

func (c *client) UpdateSiteRole(user, role string) error {
	foundUser, err := c.getUser(user)
	if err != nil {
		return err
	}
	userId := foundUser.Users.User[0].ID

	body := userSiteRoleData{
		User: userSiteRole{
			SiteRole: role,
		},
	}
	url := fmt.Sprintf("/api/%v/sites/%v/users/%v", c.apiVersion, c.siteID, userId)
	req, err := c.newRequest(http.MethodPut, url, body)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *client) GrantWorkbookAccess(resource *Workbook, user, role string) error {
	requestWorkbook := workbook{
		ID: resource.ID,
	}
	foundUser, err := c.getUser(user)
	if err != nil {
		return err
	}
	userId := foundUser.Users.User[0].ID
	requestUser := workbookUser{
		ID: userId,
	}

	split := strings.Split(role, ":")
	requestCapability := capability{
		Name: split[0],
		Mode: split[1],
	}

	capabilityArr := []capability{requestCapability}
	requestCapabilities := capabilities{
		Capability: capabilityArr,
	}

	requestGranteeCapabilities := granteeCapabilities{
		User:         requestUser,
		Capabilities: requestCapabilities,
	}

	granteeCapabilities := []granteeCapabilities{requestGranteeCapabilities}

	permission := permissions{
		Workbook:            requestWorkbook,
		GranteeCapabilities: granteeCapabilities,
	}
	c.addWorkbookPermissions(resource.ID, permission)

	return nil
}

func (c *client) RevokeWorkbookAccess(resource *Workbook, user, role string) error {
	foundUser, err := c.getUser(user)
	if err != nil {
		return err
	}
	userId := foundUser.Users.User[0].ID
	c.deleteWorkbookPermissions(resource.ID, userId, role)
	return nil
}

func (c *client) getUser(email string) (*siteUsers, error) {
	filter := fmt.Sprintf("name:eq:%v", email)
	url := fmt.Sprintf("/api/%v/sites/%v/users?filter=%v", c.apiVersion, c.siteID, filter)
	req, err := c.newRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var user *siteUsers
	_, err = c.do(req, &user)
	if err != nil {
		return nil, err
	}

	if len(user.Users.User) == 0 {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func newClient(config *ClientConfig) (*client, error) {
	defaults.SetDefaults(config)
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
		apiVersion: config.APIVersion,
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

func (c *client) addWorkbookPermissions(id string, permissions permissions) error {
	body := workbookPermissions{
		Permissions: permissions,
	}
	url := fmt.Sprintf("/api/%v/sites/%v/workbooks/%v/permissions", c.apiVersion, c.siteID, id)
	req, err := c.newRequest(http.MethodPut, url, body)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err

}

func (c *client) deleteWorkbookPermissions(id, user, role string) error {
	split := strings.Split(role, ":")
	capabilityName := split[0]
	capabilityMode := split[1]
	url := fmt.Sprintf("/api/%v/sites/%v/workbooks/%v/permissions/users/%v/%v/%v", c.apiVersion, c.siteID, id, user, capabilityName, capabilityMode)
	req, err := c.newRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err

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

	url := fmt.Sprintf("/api/%v/auth/signin", c.apiVersion)
	req, err := c.newRequest(http.MethodPost, url, sessionRequest)
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
