package metabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"github.com/go-playground/validator/v10"
)

type ClientConfig struct {
	Host     string `validate:"required,url" mapstructure:"host"`
	Username string `validate:"required" mapstructure:"username"`
	Password string `validate:"required" mapstructure:"password"`
}

type user struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	MembershipID int    `json:"membership_id"`
}

type group struct {
	ID      int    `json:"id,omitempty"`
	Name    string `json:"name"`
	Members []user `json:"members"`
}

type sessionRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type sessionResponse struct {
	ID string `json:"id"`
}

type databasePermission struct {
	Native  string `json:"native" mapstructure:"native"`
	Schemas string `json:"schemas" mapstructure:"schemas"`
}

type databaseGraph struct {
	Revision int `json:"revision"`
	// Groups is a map[group_id]map[database_id]databasePermission
	Groups map[string]map[string]databasePermission `json:"groups"`
}

type collectionGraph struct {
	Revision int `json:"revision"`
	// Groups is a map[group_id]map[database_id]role string
	Groups map[string]map[string]string `json:"groups"`
}

type membershipRequest struct {
	GroupID int `json:"group_id"`
	UserID  int `json:"user_id"`
}

type client struct {
	baseURL *url.URL

	username     string
	password     string
	sessionToken string

	httpClient *http.Client

	userIDs map[string]int
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
		httpClient: &http.Client{},
		userIDs:    map[string]int{},
	}

	sessionToken, err := c.getSessionToken()
	if err != nil {
		return nil, err
	}
	c.sessionToken = sessionToken

	return c, nil
}

func (c *client) GetDatabases() ([]*Database, error) {
	req, err := c.newRequest(http.MethodGet, "/api/database", nil)
	if err != nil {
		return nil, err
	}

	var databases []*Database
	if _, err := c.do(req, &databases); err != nil {
		return nil, err
	}

	return databases, nil
}

func (c *client) GetCollections() ([]*Collection, error) {
	req, err := c.newRequest(http.MethodGet, "/api/collection", nil)
	if err != nil {
		return nil, err
	}

	var collection []*Collection
	if _, err := c.do(req, &collection); err != nil {
		return nil, err
	}

	return collection, nil
}

func (c *client) GrantDatabaseAccess(resource *Database, user, role string) error {
	access, err := c.getDatabaseAccess()
	if err != nil {
		return err
	}

	var groupNameSuffix string
	var designatedRole databasePermission
	if role == DatabaseRoleViewer {
		designatedRole = databasePermission{
			Schemas: "all",
		}
		groupNameSuffix = "Viewer"
	} else if role == DatabaseRoleEditor {
		designatedRole = databasePermission{
			Schemas: "all",
			Native:  "write",
		}
		groupNameSuffix = "Editor"
	}

	var designatedGroupID int
	resourceID := fmt.Sprintf("%v", resource.ID)
	for groupID, permissions := range access.Groups {
		for databaseID, permission := range permissions {
			if databaseID == resourceID && reflect.DeepEqual(designatedRole, permission) {
				groupIDint, err := strconv.Atoi(groupID)
				if err != nil {
					return err
				}

				designatedGroupID = groupIDint
				break
			}
		}
	}

	if designatedGroupID == 0 {
		group := &group{
			Name: fmt.Sprintf("%s - %s", resource.Name, groupNameSuffix),
		}
		if err := c.createGroup(group); err != nil {
			return err
		}
		designatedGroupID = group.ID

		groupIDstr := strconv.Itoa(group.ID)
		collectionID := fmt.Sprintf("%v", resource.ID)
		access.Groups[groupIDstr] = map[string]databasePermission{}
		access.Groups[groupIDstr][collectionID] = designatedRole
		if err := c.updateDatabaseAccess(access); err != nil {
			return err
		}
	}

	userID, err := c.getUserID(user)
	if err != nil {
		return err
	}
	return c.addGroupMember(designatedGroupID, userID)
}

func (c *client) RevokeDatabaseAccess(resource *Database, user, role string) error {
	access, err := c.getDatabaseAccess()
	if err != nil {
		return err
	}

	var designatedRole databasePermission
	if role == DatabaseRoleViewer {
		designatedRole = databasePermission{
			Schemas: "all",
		}
	} else if role == DatabaseRoleEditor {
		designatedRole = databasePermission{
			Schemas: "all",
			Native:  "write",
		}
	}

	resourceID := fmt.Sprintf("%v", resource.ID)
	var designatedGroupID int
	for groupID, permissions := range access.Groups {

		for databaseID, permission := range permissions {
			if databaseID == resourceID && reflect.DeepEqual(designatedRole, permission) {
				groupIDInt, err := strconv.Atoi(groupID)
				if err != nil {
					return err
				}
				designatedGroupID = groupIDInt
				break
			}
		}
		if designatedGroupID != 0 {
			break
		}
	}

	if designatedGroupID != 0 {
		return ErrPermissionNotFound
	}

	group, err := c.getGroup(designatedGroupID)
	if err != nil {
		return err
	}

	var membershipID int
	for _, member := range group.Members {
		if member.Email == user {
			membershipID = member.MembershipID
			break
		}
	}
	if membershipID == 0 {
		return ErrPermissionNotFound
	}

	return c.removeGroupMember(membershipID)
}

func (c *client) GrantCollectionAccess(resource *Collection, user, role string) error {
	access, err := c.getCollectionAccess()
	if err != nil {
		return err
	}

	var designatedGroupID int
	for groupID, collections := range access.Groups {
		for collectionID, currentRole := range collections {
			if collectionID == resource.ID && currentRole == role {
				groupIDint, err := strconv.Atoi(groupID)
				if err != nil {
					return err
				}

				designatedGroupID = groupIDint
				break
			}
		}
	}

	if designatedGroupID == 0 {
		group := &group{
			Name: fmt.Sprintf("%s - %s", resource.Name, role),
		}
		if err := c.createGroup(group); err != nil {
			return err
		}
		designatedGroupID = group.ID

		groupIDstr := strconv.Itoa(group.ID)
		collectionID := fmt.Sprintf("%v", resource.ID)
		access.Groups[groupIDstr] = map[string]string{}
		access.Groups[groupIDstr][collectionID] = role
		if err := c.updateCollectionAccess(access); err != nil {
			return err
		}
	}

	userID, err := c.getUserID(user)
	if err != nil {
		return err
	}

	return c.addGroupMember(designatedGroupID, userID)
}

func (c *client) getUserID(email string) (int, error) {
	if c.userIDs[email] != 0 {
		return c.userIDs[email], nil
	}

	users, err := c.getUsers()
	if err != nil {
		return 0, err
	}

	userIDs := map[string]int{}
	for _, u := range users {
		userIDs[u.Email] = u.ID
	}
	c.userIDs = userIDs

	if c.userIDs[email] == 0 {
		return 0, ErrUserNotFound
	}
	return c.userIDs[email], nil
}

func (c *client) getUsers() ([]user, error) {
	req, err := c.newRequest(http.MethodGet, "/api/user", nil)
	if err != nil {
		return nil, err
	}

	var users []user
	if _, err := c.do(req, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (c *client) getSessionToken() (string, error) {
	sessionRequest := &sessionRequest{
		Username: c.username,
		Password: c.password,
	}
	req, err := c.newRequest(http.MethodPost, "/api/session", sessionRequest)
	if err != nil {
		return "", err
	}

	var sessionResponse sessionResponse
	if _, err := c.do(req, &sessionResponse); err != nil {
		return "", err
	}

	return sessionResponse.ID, nil
}

func (c *client) getCollectionAccess() (*collectionGraph, error) {
	req, err := c.newRequest(http.MethodGet, "/api/collection/graph", nil)
	if err != nil {
		return nil, err
	}

	var graph collectionGraph
	if _, err := c.do(req, &graph); err != nil {
		return nil, err
	}

	return &graph, nil
}

func (c *client) updateCollectionAccess(access *collectionGraph) error {
	req, err := c.newRequest(http.MethodPut, "/api/collection/graph", access)
	if err != nil {
		return err
	}

	if _, err := c.do(req, &access); err != nil {
		return err
	}

	return nil
}

func (c *client) getDatabaseAccess() (*databaseGraph, error) {
	req, err := c.newRequest(http.MethodGet, "/api/permissions/graph", nil)
	if err != nil {
		return nil, err
	}

	var dbGraph databaseGraph
	if _, err := c.do(req, &dbGraph); err != nil {
		return nil, err
	}

	return &dbGraph, nil
}

func (c *client) updateDatabaseAccess(dbGraph *databaseGraph) error {
	req, err := c.newRequest(http.MethodPut, "/api/permissions/graph", dbGraph)
	if err != nil {
		return err
	}

	if _, err := c.do(req, &dbGraph); err != nil {
		return err
	}

	return nil
}

func (c *client) createGroup(group *group) error {
	req, err := c.newRequest(http.MethodPost, "/api/permissions/group", group)
	if err != nil {
		return err
	}

	if _, err := c.do(req, group); err != nil {
		return err
	}

	return nil
}

func (c *client) getGroup(id int) (*group, error) {
	url := fmt.Sprintf("/api/permissions/group/%d", id)
	req, err := c.newRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var group group

	if _, err := c.do(req, &group); err != nil {
		return nil, err
	}

	return &group, nil
}

func (c *client) addGroupMember(groupID, userID int) error {
	req, err := c.newRequest(http.MethodPost, "/api/permissions/membership", membershipRequest{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		return err
	}

	if _, err := c.do(req, nil); err != nil {
		return err
	}

	return nil
}

func (c *client) removeGroupMember(membershipID int) error {
	url := fmt.Sprintf("/api/permissions/membership/%d", membershipID)
	req, err := c.newRequest(http.MethodDelete, url, nil)

	if err != nil {
		return err
	}

	if _, err := c.do(req, nil); err != nil {
		return err
	}

	return nil
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
	req.Header.Set("X-Metabase-Session", c.sessionToken)
	return req, nil
}

func (c *client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		newSessionToken, err := c.getSessionToken()
		if err != nil {
			return nil, err
		}
		c.sessionToken = newSessionToken
		req.Header.Set("X-Metabase-Session", c.sessionToken)

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
