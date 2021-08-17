package tableau

import (
	"bytes"
	"encoding/json"
	"errors"
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
	GetFlows() ([]*Flow, error)
	GetDataSources() ([]*DataSource, error)
	GetViews() ([]*View, error)
	GetMetrics() ([]*Metric, error)
	UpdateSiteRole(user, role string) error
	GrantWorkbookAccess(resource *Workbook, user, role string) error
	RevokeWorkbookAccess(resource *Workbook, user, role string) error
	GrantFlowAccess(resource *Flow, user, role string) error
	RevokeFlowAccess(resource *Flow, user, role string) error
	GrantDataSourceAccess(resource *DataSource, user, role string) error
	RevokeDataSourceAccess(resource *DataSource, user, role string) error
	GrantViewAccess(resource *View, user, role string) error
	RevokeViewAccess(resource *View, user, role string) error
	GrantMetricAccess(resource *Metric, user, role string) error
	RevokeMetricAccess(resource *Metric, user, role string) error
}

type ClientConfig struct {
	Host       string `validate:"required,url" mapstructure:"host"`
	Username   string `validate:"required" mapstructure:"username"`
	Password   string `validate:"required" mapstructure:"password"`
	ContentURL string `validate:"required" mapstructure:"content_url"`
	APIVersion string `mapstructure:"apiVersion" default:"3.12"`
	HTTPClient HTTPClient
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

type SessionResponse struct {
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

	httpClient HTTPClient
}

type workbookPermissions struct {
	Permissions workbookPermission `json:"permissions"`
}

type flowPermissions struct {
	Permissions flowPermission `json:"permissions"`
}

type datasourcePermissions struct {
	Permissions datasourcePermission `json:"permissions"`
}

type viewPermissions struct {
	Permissions viewPermission `json:"permissions"`
}

type metricPermissions struct {
	Permissions metricPermission `json:"permissions"`
}

type resourceDetails struct {
	ID string `json:"id"`
}

type userDetails struct {
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
	User         userDetails  `json:"user"`
	Capabilities capabilities `json:"capabilities"`
}
type workbookPermission struct {
	Workbook            resourceDetails       `json:"workbook"`
	GranteeCapabilities []granteeCapabilities `json:"granteeCapabilities"`
}

type flowPermission struct {
	Flow                resourceDetails       `json:"flow"`
	GranteeCapabilities []granteeCapabilities `json:"granteeCapabilities"`
}

type datasourcePermission struct {
	DataSource          resourceDetails       `json:"datasource"`
	GranteeCapabilities []granteeCapabilities `json:"granteeCapabilities"`
}

type viewPermission struct {
	View                resourceDetails       `json:"view"`
	GranteeCapabilities []granteeCapabilities `json:"granteeCapabilities"`
}

type metricPermission struct {
	Metric              resourceDetails       `json:"metric"`
	GranteeCapabilities []granteeCapabilities `json:"granteeCapabilities"`
}

type responseWorkbooks struct {
	Pagination pagination `json:"pagination"`
	Workbooks  workbooks  `json:"workbooks"`
}

type responseFlows struct {
	Pagination pagination `json:"pagination"`
	Flows      flows      `json:"flows"`
}

type responseDataSources struct {
	Pagination  pagination  `json:"pagination"`
	DataSources datasources `json:"datasources"`
}

type responseViews struct {
	Pagination pagination `json:"pagination"`
	Views      views      `json:"views"`
}

type responseMetrics struct {
	Pagination pagination `json:"pagination"`
	Metrics    metrics    `json:"metrics"`
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

type flows struct {
	Flow []*Flow `json:"flow"`
}

type datasources struct {
	DataSource []*DataSource `json:"datasource"`
}

type views struct {
	View []*View `json:"view"`
}

type metrics struct {
	Metric []*Metric `json:"metric"`
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

type errorTemplate struct {
	Error struct {
		Summary string `json:"summary"`
		Detail  string `json:"detail"`
		Code    string `json:"code"`
	} `json:"error"`
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

func (c *client) GetFlows() ([]*Flow, error) {
	url := fmt.Sprintf("/api/%v/sites/%v/flows", c.apiVersion, c.siteID)
	req, err := c.newRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var flows responseFlows
	if _, err := c.do(req, &flows); err != nil {
		return nil, err
	}
	return flows.Flows.Flow, nil
}

func (c *client) GetDataSources() ([]*DataSource, error) {
	url := fmt.Sprintf("/api/%v/sites/%v/datasources", c.apiVersion, c.siteID)
	req, err := c.newRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var datasources responseDataSources
	if _, err := c.do(req, &datasources); err != nil {
		return nil, err
	}
	return datasources.DataSources.DataSource, nil
}

func (c *client) GetViews() ([]*View, error) {
	url := fmt.Sprintf("/api/%v/sites/%v/views", c.apiVersion, c.siteID)
	req, err := c.newRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var views responseViews
	if _, err := c.do(req, &views); err != nil {
		return nil, err
	}
	return views.Views.View, nil
}

func (c *client) GetMetrics() ([]*Metric, error) {
	url := fmt.Sprintf("/api/%v/sites/%v/metrics", c.apiVersion, c.siteID)
	req, err := c.newRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var metrics responseMetrics
	if _, err := c.do(req, &metrics); err != nil {
		return nil, err
	}
	return metrics.Metrics.Metric, nil
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
	requestWorkbook := resourceDetails{
		ID: resource.ID,
	}
	foundUser, err := c.getUser(user)
	if err != nil {
		return err
	}
	userId := foundUser.Users.User[0].ID
	requestUser := userDetails{
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

	permission := workbookPermission{
		Workbook:            requestWorkbook,
		GranteeCapabilities: granteeCapabilities,
	}
	err = c.addWorkbookPermissions(resource.ID, permission)
	return err
}

func (c *client) RevokeWorkbookAccess(resource *Workbook, user, role string) error {
	foundUser, err := c.getUser(user)
	if err != nil {
		return err
	}
	userId := foundUser.Users.User[0].ID
	err = c.deleteWorkbookPermissions(resource.ID, userId, role)
	return err
}

func (c *client) GrantFlowAccess(resource *Flow, user, role string) error {
	requestFlow := resourceDetails{
		ID: resource.ID,
	}
	foundUser, err := c.getUser(user)
	if err != nil {
		return err
	}
	userId := foundUser.Users.User[0].ID
	requestUser := userDetails{
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

	permission := flowPermission{
		Flow:                requestFlow,
		GranteeCapabilities: granteeCapabilities,
	}
	err = c.addFlowPermissions(resource.ID, permission)
	return err
}

func (c *client) RevokeFlowAccess(resource *Flow, user, role string) error {
	foundUser, err := c.getUser(user)
	if err != nil {
		return err
	}
	userId := foundUser.Users.User[0].ID
	err = c.deleteFlowPermissions(resource.ID, userId, role)
	return err
}

func (c *client) GrantMetricAccess(resource *Metric, user, role string) error {
	requestMetric := resourceDetails{
		ID: resource.ID,
	}
	foundUser, err := c.getUser(user)
	if err != nil {
		return err
	}
	userId := foundUser.Users.User[0].ID
	requestUser := userDetails{
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

	permission := metricPermission{
		Metric:              requestMetric,
		GranteeCapabilities: granteeCapabilities,
	}
	err = c.addMetricPermissions(resource.ID, permission)
	return err
}

func (c *client) RevokeMetricAccess(resource *Metric, user, role string) error {
	foundUser, err := c.getUser(user)
	if err != nil {
		return err
	}
	userId := foundUser.Users.User[0].ID
	err = c.deleteMetricPermissions(resource.ID, userId, role)
	return err
}

func (c *client) GrantDataSourceAccess(resource *DataSource, user, role string) error {
	requestDataSource := resourceDetails{
		ID: resource.ID,
	}
	foundUser, err := c.getUser(user)
	if err != nil {
		return err
	}
	userId := foundUser.Users.User[0].ID
	requestUser := userDetails{
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

	permission := datasourcePermission{
		DataSource:          requestDataSource,
		GranteeCapabilities: granteeCapabilities,
	}
	err = c.addDataSourcePermissions(resource.ID, permission)
	return err
}

func (c *client) RevokeDataSourceAccess(resource *DataSource, user, role string) error {
	foundUser, err := c.getUser(user)
	if err != nil {
		return err
	}
	userId := foundUser.Users.User[0].ID
	err = c.deleteDataSourcePermissions(resource.ID, userId, role)
	return err
}

func (c *client) GrantViewAccess(resource *View, user, role string) error {
	requestView := resourceDetails{
		ID: resource.ID,
	}
	foundUser, err := c.getUser(user)
	if err != nil {
		return err
	}
	userId := foundUser.Users.User[0].ID
	requestUser := userDetails{
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

	permission := viewPermission{
		View:                requestView,
		GranteeCapabilities: granteeCapabilities,
	}
	err = c.addViewPermissions(resource.ID, permission)
	return err
}

func (c *client) RevokeViewAccess(resource *View, user, role string) error {
	foundUser, err := c.getUser(user)
	if err != nil {
		return err
	}
	userId := foundUser.Users.User[0].ID
	err = c.deleteViewPermissions(resource.ID, userId, role)
	return err
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

func NewClient(config *ClientConfig) (*client, error) {
	defaults.SetDefaults(config)
	if err := validator.New().Struct(config); err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(config.Host)
	if err != nil {
		return nil, err
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	c := &client{
		baseURL:    baseURL,
		username:   config.Username,
		password:   config.Password,
		contentUrl: config.ContentURL,
		apiVersion: config.APIVersion,
		httpClient: httpClient,
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

func (c *client) addWorkbookPermissions(id string, permissions workbookPermission) error {
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

func (c *client) addFlowPermissions(id string, permissions flowPermission) error {
	body := flowPermissions{
		Permissions: permissions,
	}
	url := fmt.Sprintf("/api/%v/sites/%v/flows/%v/permissions", c.apiVersion, c.siteID, id)
	req, err := c.newRequest(http.MethodPut, url, body)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err

}

func (c *client) addMetricPermissions(id string, permissions metricPermission) error {
	body := metricPermissions{
		Permissions: permissions,
	}
	url := fmt.Sprintf("/api/%v/sites/%v/metrics/%v/permissions", c.apiVersion, c.siteID, id)
	req, err := c.newRequest(http.MethodPut, url, body)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err

}

func (c *client) addDataSourcePermissions(id string, permissions datasourcePermission) error {
	body := datasourcePermissions{
		Permissions: permissions,
	}
	url := fmt.Sprintf("/api/%v/sites/%v/datasources/%v/permissions", c.apiVersion, c.siteID, id)
	req, err := c.newRequest(http.MethodPut, url, body)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err

}

func (c *client) addViewPermissions(id string, permissions viewPermission) error {
	body := viewPermissions{
		Permissions: permissions,
	}
	url := fmt.Sprintf("/api/%v/sites/%v/views/%v/permissions", c.apiVersion, c.siteID, id)
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

func (c *client) deleteFlowPermissions(id, user, role string) error {
	split := strings.Split(role, ":")
	capabilityName := split[0]
	capabilityMode := split[1]
	url := fmt.Sprintf("/api/%v/sites/%v/flows/%v/permissions/users/%v/%v/%v", c.apiVersion, c.siteID, id, user, capabilityName, capabilityMode)
	req, err := c.newRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err

}

func (c *client) deleteMetricPermissions(id, user, role string) error {
	split := strings.Split(role, ":")
	capabilityName := split[0]
	capabilityMode := split[1]
	url := fmt.Sprintf("/api/%v/sites/%v/metrics/%v/permissions/users/%v/%v/%v", c.apiVersion, c.siteID, id, user, capabilityName, capabilityMode)
	req, err := c.newRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err

}

func (c *client) deleteDataSourcePermissions(id, user, role string) error {
	split := strings.Split(role, ":")
	capabilityName := split[0]
	capabilityMode := split[1]
	url := fmt.Sprintf("/api/%v/sites/%v/datasources/%v/permissions/users/%v/%v/%v", c.apiVersion, c.siteID, id, user, capabilityName, capabilityMode)
	req, err := c.newRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err

}

func (c *client) deleteViewPermissions(id, user, role string) error {
	split := strings.Split(role, ":")
	capabilityName := split[0]
	capabilityMode := split[1]
	url := fmt.Sprintf("/api/%v/sites/%v/views/%v/permissions/users/%v/%v/%v", c.apiVersion, c.siteID, id, user, capabilityName, capabilityMode)
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
		return "", "", "", err
	}

	var sessionResponse SessionResponse
	if _, err := c.do(req, &sessionResponse); err != nil {
		return "", "", "", err
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
	} else if resp.StatusCode == http.StatusForbidden {
		responseError := &errorTemplate{}
		json.NewDecoder(resp.Body).Decode(responseError)
		return nil, errors.New(responseError.Error.Detail)
	}

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}
