package metabase

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/goto/guardian/pkg/tracing"
	"github.com/goto/salt/log"

	"github.com/mitchellh/mapstructure"

	"github.com/go-playground/validator/v10"
)

const (
	databaseEndpoint             = "/api/database?include=tables"
	collectionEndpoint           = "/api/collection"
	groupEndpoint                = "/api/permissions/group"
	databasePermissionEndpoint   = "/api/permissions/graph"
	collectionPermissionEndpoint = "/api/collection/graph"

	data             = "data"
	database         = "database"
	collection       = "collection"
	groups           = "groups"
	table            = "table"
	none             = "none"
	urn              = "urn"
	name             = "name"
	permissionsConst = "permissions"
	groupConst       = "group"

	pathSeparator = "/"
)

type ResourceGroupDetails map[string][]map[string]interface{}

type MetabaseClient interface {
	GetDatabases() ([]*Database, error)
	GetCollections() ([]*Collection, error)
	GetGroups() ([]*Group, ResourceGroupDetails, ResourceGroupDetails, error)
	GrantDatabaseAccess(resource *Database, user, role string, groups map[string]*Group) error
	RevokeDatabaseAccess(resource *Database, user, role string) error
	GrantCollectionAccess(resource *Collection, user, role string) error
	RevokeCollectionAccess(resource *Collection, user, role string) error
	GrantTableAccess(resource *Table, user, role string, groups map[string]*Group) error
	RevokeTableAccess(resource *Table, user, role string) error
	GrantGroupAccess(groupID int, email string) error
	RevokeGroupAccess(groupID int, email string) error
}

type ClientConfig struct {
	Host       string `validate:"required,url" mapstructure:"host"`
	Username   string `validate:"required" mapstructure:"username"`
	Password   string `validate:"required" mapstructure:"password"`
	HTTPClient HTTPClient
}

type user struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	MembershipID int    `json:"membership_id" mapstructure:"membership_id"`
	GroupIds     []int  `json:"group_ids" mapstructure:"group_ids"`
}

type group struct {
	ID      int    `json:"id,omitempty"`
	Name    string `json:"name"`
	Members []user `json:"members"`
}

type SessionRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SessionResponse struct {
	ID string `json:"id"`
}

type databasePermission map[string]interface{}

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

var (
	databaseViewerPermission = databasePermission{
		"schemas": "all",
	}
	databaseEditorPermission = databasePermission{
		"native":  "write",
		"schemas": "all",
	}
)

type client struct {
	baseURL *url.URL

	username     string
	password     string
	sessionToken string

	httpClient HTTPClient

	userIDs map[string]int

	logger log.Logger
}

func NewClient(config *ClientConfig, logger log.Logger) (*client, error) {
	if err := validator.New().Struct(config); err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(config.Host)
	if err != nil {
		return nil, err
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = tracing.NewHttpClient("MetabaseHttpClient")
	}

	c := &client{
		baseURL:    baseURL,
		username:   config.Username,
		password:   config.Password,
		httpClient: httpClient,
		userIDs:    map[string]int{},
		logger:     logger,
	}

	sessionToken, err := c.getSessionToken()
	if err != nil {
		return nil, err
	}
	c.sessionToken = sessionToken

	return c, nil
}

func (c *client) GetDatabases() ([]*Database, error) {
	req, err := c.newRequest(http.MethodGet, databaseEndpoint, nil)
	if err != nil {
		return nil, err
	}

	var databases []*Database
	var response interface{}

	_, err = c.do(req, &response)
	if err != nil {
		return databases, err
	}

	if v, ok := response.([]interface{}); ok {
		err = mapstructure.Decode(v, &databases) // this is for metabase v0.37
	} else if v, ok := response.(map[string]interface{}); ok && v[data] != nil {
		err = mapstructure.Decode(v[data], &databases) // this is for metabase v0.42
	} else {
		return databases, ErrInvalidApiResponse
	}

	if err != nil {
		return databases, err
	}
	c.logger.Info("Fetch database from request", "total", len(databases), req.URL)
	return databases, err
}

func (c *client) GetCollections() ([]*Collection, error) {
	req, err := c.newRequest(http.MethodGet, collectionEndpoint, nil)
	if err != nil {
		return nil, err
	}

	var collections []*Collection
	result := make([]*Collection, 0)
	if _, err := c.do(req, &collections); err != nil {
		return nil, err
	}
	c.logger.Info("Fetch collections from request", "total", len(collections), req.URL)

	collectionIdNameMap := make(map[string]string, 0)
	for _, collection := range collections {
		collectionIdNameMap[fmt.Sprintf("%v", collection.ID)] = collection.Name
	}

	for _, collection := range collections {
		// don't add personal collection
		if collection.PersonalOwnerId == nil {
			locationPath := ""
			locations := strings.Split(collection.Location, pathSeparator)
			if len(locations) > 1 {
				for _, id := range locations {
					if name, ok := collectionIdNameMap[id]; ok && len(id) > 0 {
						locationPath = locationPath + name + pathSeparator
					}
				}
				//populate resource name as hierarchy of its parent name
				collection.Name = locationPath + collection.Name
				result = append(result, collection)
			}
		}
	}

	return result, nil
}

func (c *client) GetGroups() ([]*Group, ResourceGroupDetails, ResourceGroupDetails, error) {
	wg := sync.WaitGroup{}
	wg.Add(3)

	var groups []*Group
	var err error
	go c.fetchGroups(&wg, &groups, err)

	databaseResourceGroups := make(ResourceGroupDetails, 0)
	go c.fetchDatabasePermissions(&wg, databaseResourceGroups, err)

	collectionResourceGroups := make(ResourceGroupDetails, 0)
	go c.fetchCollectionPermissions(&wg, collectionResourceGroups, err)

	wg.Wait()

	groupMap := make(map[string]*Group, 0)
	for _, group := range groups {
		groupMap[fmt.Sprintf("group:%d", group.ID)] = group
	}

	addResourceToGroup(databaseResourceGroups, groupMap, database)
	addResourceToGroup(collectionResourceGroups, groupMap, collection)

	return groups, databaseResourceGroups, collectionResourceGroups, err
}

func (c *client) fetchGroups(wg *sync.WaitGroup, groups *[]*Group, err error) {
	defer wg.Done()
	req, err := c.newRequest(http.MethodGet, groupEndpoint, nil)
	if err != nil {
		return
	}

	_, err = c.do(req, &groups)
	if err != nil {
		return
	}
	c.logger.Info("Fetch groups from request", "total", len(*groups), req.URL)
}

func (c *client) fetchDatabasePermissions(wg *sync.WaitGroup, resourceGroups ResourceGroupDetails, err error) {
	defer wg.Done()

	req, err := c.newRequest(http.MethodGet, databasePermissionEndpoint, nil)
	if err != nil {
		return
	}

	graphs := make(map[string]interface{}, 0)
	_, err = c.do(req, &graphs)
	if err != nil {
		return
	}

	for groupId, r := range graphs[groups].(map[string]interface{}) {
		for dbId, role := range r.(map[string]interface{}) {
			if roles, ok := role.(map[string]interface{}); ok {
				permissions := make([]string, 0)
				for key, value := range roles {
					if tables, ok := value.(map[string]interface{}); ok {
						for _, tables := range tables {
							if tables, ok := tables.(map[string]interface{}); ok {
								for tableId, tablePermission := range tables {
									perm, ok := tablePermission.(string)
									if !ok {
										c.logger.Warn("Invalid permission type for metabase group", "dbId", dbId, "tableId", tableId, "groupId", groupId, "permission", tablePermission, "type", reflect.TypeOf(tablePermission))
										continue
									}
									addGroupToResource(resourceGroups, fmt.Sprintf("%s:%s.%s", table, dbId, tableId), groupId, []string{perm}, err)
								}
							}
						}
					} else {
						permissions = append(permissions, fmt.Sprintf("%s:%s", key, value))
					}
				}
				addGroupToResource(resourceGroups, fmt.Sprintf("%s:%s", database, dbId), groupId, permissions, err)
			}
		}
	}
}

func (c *client) fetchCollectionPermissions(wg *sync.WaitGroup, resourceGroups ResourceGroupDetails, err error) {
	defer wg.Done()

	req, err := c.newRequest(http.MethodGet, collectionPermissionEndpoint, nil)
	if err != nil {
		return
	}

	graphs := make(map[string]interface{}, 0)
	_, err = c.do(req, &graphs)
	if err != nil {
		return
	}
	c.logger.Info(fmt.Sprintf("Fetch permissions for collections from request: %v", req.URL))
	for groupId, r := range graphs[groups].(map[string]interface{}) {
		for collectionId, permission := range r.(map[string]interface{}) {
			if permission != none {
				p, ok := permission.(string)
				if !ok {
					c.logger.Warn("Invalid permission type for metabase collection", "collectionId", collectionId, "groupId", groupId, "permission", permission, "type", reflect.TypeOf(permission))
					continue
				}
				addGroupToResource(resourceGroups, fmt.Sprintf("%s:%s", collection, collectionId), groupId, []string{p}, err)
			}
		}
	}
}

func addResourceToGroup(resourceGroups ResourceGroupDetails, groupMap map[string]*Group, resourceType string) {
	for resourceId, groups := range resourceGroups {
		for _, groupDetails := range groups {
			groupID := groupDetails[urn].(string)
			if group, ok := groupMap[groupID]; ok {
				if strings.HasPrefix(group.Name, GuardianGroupPrefix) {
					continue
				}
				groupDetails[name] = group.Name
				if resourceType == database {
					group.DatabaseResources = append(group.DatabaseResources, &GroupResource{Urn: resourceId, Permissions: groupDetails[permissionsConst].([]string)})
				}
				if resourceType == collection {
					group.CollectionResources = append(group.CollectionResources, &GroupResource{Urn: resourceId, Permissions: groupDetails[permissionsConst].([]string)})
				}
			}
		}
	}
}

func addGroupToResource(resourceGroups ResourceGroupDetails, resourceId string, groupId string, permissions []string, err error) {
	id, err := strconv.Atoi(groupId)
	if err != nil {
		return
	}
	if groups, ok := resourceGroups[resourceId]; ok {
		groups = append(groups, map[string]interface{}{urn: fmt.Sprintf("%s:%d", groupConst, id), permissionsConst: permissions})
		resourceGroups[resourceId] = groups
	} else {
		resourceGroups[resourceId] = []map[string]interface{}{{urn: fmt.Sprintf("%s:%d", groupConst, id), permissionsConst: permissions}}
	}
}

func (c *client) GrantDatabaseAccess(resource *Database, email, role string, groups map[string]*Group) error {
	access, err := c.getDatabaseAccess()
	if err != nil {
		return err
	}

	var dbPermission databasePermission
	if role == DatabaseRoleViewer {
		dbPermission = databaseViewerPermission
	} else if role == DatabaseRoleEditor {
		dbPermission = databaseEditorPermission
	} else {
		return ErrInvalidRole
	}

	resourceIDStr := strconv.Itoa(resource.ID)
	toBrGroupName := fmt.Sprintf("%s_%v_%s", ResourceTypeDatabase, resource.ID, role)
	groupID := c.findDatabaseAccessGroup(access, resourceIDStr, dbPermission)

	if groupID == "" {
		if g, ok := groups[toBrGroupName]; ok {
			groupID = strconv.Itoa(g.ID)
		} else {
			g := &group{
				Name: toBrGroupName,
			}
			if err := c.createGroup(g); err != nil {
				return err
			}

			groupID = strconv.Itoa(g.ID)
		}

		databaseID := fmt.Sprintf("%v", resource.ID)

		access.Groups[groupID] = map[string]databasePermission{}
		access.Groups[groupID][databaseID] = dbPermission
		if err := c.updateDatabaseAccess(access); err != nil {
			return err
		}
	}

	groupIDInt, err := strconv.Atoi(groupID)
	if err != nil {
		return err
	}

	user, err := c.getUser(email)
	if err != nil {
		return err
	}

	for _, groupId := range user.GroupIds {
		if groupId == groupIDInt {
			return nil
		}
	}

	return c.addGroupMember(groupIDInt, user.ID)
}

func (c *client) RevokeDatabaseAccess(resource *Database, user, role string) error {
	access, err := c.getDatabaseAccess()
	if err != nil {
		return err
	}

	var dbPermission databasePermission
	if role == DatabaseRoleViewer {
		dbPermission = databaseViewerPermission
	} else if role == DatabaseRoleEditor {
		dbPermission = databaseEditorPermission
	}

	resourceIDStr := strconv.Itoa(resource.ID)
	groupID := c.findDatabaseAccessGroup(access, resourceIDStr, dbPermission)

	if groupID == "" {
		return ErrPermissionNotFound
	}

	groupIDInt, err := strconv.Atoi(groupID)
	if err != nil {
		return err
	}
	return c.removeMembership(groupIDInt, user)
}

func (c *client) GrantCollectionAccess(resource *Collection, email, role string) error {
	access, err := c.getCollectionAccess()
	if err != nil {
		return err
	}

	resourceIDStr := fmt.Sprintf("%v", resource.ID)
	if role != CollectionRoleViewer && role != CollectionRoleCurate {
		return ErrInvalidRole
	}

	groupID := c.findCollectionAccessGroup(access, resourceIDStr, role)

	if groupID == "" {
		g := &group{
			Name: fmt.Sprintf("%s_%s_%s", ResourceTypeCollection, resourceIDStr, role),
		}
		if err := c.createGroup(g); err != nil {
			return err
		}

		groupID = strconv.Itoa(g.ID)
		collectionID := fmt.Sprintf("%v", resource.ID)

		access.Groups[groupID] = map[string]string{}
		access.Groups[groupID][collectionID] = role
		if err := c.updateCollectionAccess(access); err != nil {
			return err
		}
	}

	groupIDInt, err := strconv.Atoi(groupID)
	if err != nil {
		return err
	}
	user, err := c.getUser(email)
	if err != nil {
		return err
	}

	for _, groupId := range user.GroupIds {
		if groupId == groupIDInt {
			return nil
		}
	}

	return c.addGroupMember(groupIDInt, user.ID)
}

func (c *client) RevokeCollectionAccess(resource *Collection, user, role string) error {
	access, err := c.getCollectionAccess()
	if err != nil {
		return err
	}

	resourceIDStr := fmt.Sprintf("%v", resource.ID)
	groupID := c.findCollectionAccessGroup(access, resourceIDStr, role)

	if groupID == "" {
		return ErrPermissionNotFound
	}

	groupIDInt, err := strconv.Atoi(groupID)
	if err != nil {
		return err
	}
	return c.removeMembership(groupIDInt, user)
}

func (c *client) GrantTableAccess(resource *Table, email, role string, groups map[string]*Group) error {
	access, err := c.getDatabaseAccess()
	if err != nil {
		return err
	}

	var dbPermission databasePermission
	resourceIDStr := strconv.Itoa(resource.ID)
	databaseId := resource.DbId
	databaseIdStr := strconv.Itoa(databaseId)
	if role == TableRoleViewer {
		dbPermission = map[string]interface{}{
			"schemas": map[string]interface{}{
				"public": map[string]interface{}{
					resourceIDStr: "all",
				},
			},
		}
	} else {
		return ErrInvalidRole
	}

	toBrGroupName := fmt.Sprintf("%s_%v_%v_%s", ResourceTypeTable, databaseId, resource.ID, role)
	groupID := c.findTableAccessGroup(access, databaseIdStr, dbPermission)

	if groupID == "" {
		if g, ok := groups[toBrGroupName]; ok {
			groupID = strconv.Itoa(g.ID)
		} else {
			g := &group{
				Name: toBrGroupName,
			}
			if err := c.createGroup(g); err != nil {
				return err
			}

			groupID = strconv.Itoa(g.ID)
		}

		access.Groups[groupID] = map[string]databasePermission{}
		access.Groups[groupID][databaseIdStr] = dbPermission
		if err := c.updateDatabaseAccess(access); err != nil {
			return err
		}
	}

	groupIDInt, err := strconv.Atoi(groupID)
	if err != nil {
		return err
	}

	user, err := c.getUser(email)
	if err != nil {
		return err
	}

	for _, groupId := range user.GroupIds {
		if groupId == groupIDInt {
			return nil
		}
	}

	return c.addGroupMember(groupIDInt, user.ID)
}

func (c *client) RevokeTableAccess(resource *Table, user, role string) error {
	access, err := c.getDatabaseAccess()
	if err != nil {
		return err
	}

	var dbPermission databasePermission
	resourceIDStr := strconv.Itoa(resource.ID)
	databaseId := resource.DbId
	databaseIdStr := strconv.Itoa(databaseId)
	if role == TableRoleViewer {
		dbPermission = map[string]interface{}{
			"schemas": map[string]interface{}{
				"public": map[string]interface{}{
					resourceIDStr: "all",
				},
			},
		}
	} else {
		return ErrInvalidRole
	}

	groupID := c.findTableAccessGroup(access, databaseIdStr, dbPermission)

	if groupID == "" {
		return ErrPermissionNotFound
	}

	groupIDInt, err := strconv.Atoi(groupID)
	if err != nil {
		return err
	}
	return c.removeMembership(groupIDInt, user)
}

func (c *client) GrantGroupAccess(groupID int, email string) error {
	user, err := c.getUser(email)
	if err != nil {
		return err
	}

	for _, userGroupId := range user.GroupIds {
		if userGroupId == groupID {
			c.logger.Warn(fmt.Sprintf("User %s is already member of group %d", email, groupID))
			return nil
		}
	}

	return c.addGroupMember(groupID, user.ID)
}

func (c *client) RevokeGroupAccess(groupID int, email string) error {
	return c.removeMembership(groupID, email)
}

func (c *client) removeMembership(groupID int, user string) error {
	group, err := c.getGroup(groupID)
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

func (c *client) getUser(email string) (user, error) {
	req, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/user?query=%s", email), nil)
	if err != nil {
		return user{}, err
	}

	var users []user
	var response interface{}
	if _, err := c.do(req, &response); err != nil {
		return user{}, err
	}

	if v, ok := response.([]interface{}); ok {
		err = mapstructure.Decode(v, &users) // this is for metabase v0.37
	} else if v, ok := response.(map[string]interface{}); ok && v[data] != nil {
		err = mapstructure.Decode(v[data], &users) // this is for metabase v0.42
	} else {
		return user{}, ErrInvalidApiResponse
	}

	if err != nil {
		return user{}, ErrUserNotFound
	}

	for _, u := range users {
		if u.Email == email {
			return u, nil
		}
	}

	return user{}, ErrUserNotFound
}

func (c *client) getSessionToken() (string, error) {
	sessionRequest := &SessionRequest{
		Username: c.username,
		Password: c.password,
	}
	req, err := c.newRequest(http.MethodPost, "/api/session", sessionRequest)
	if err != nil {
		return "", err
	}

	var sessionResponse SessionResponse
	if _, err := c.do(req, &sessionResponse); err != nil {
		return "", err
	}

	return sessionResponse.ID, nil
}

func (c *client) getCollectionAccess() (*collectionGraph, error) {
	req, err := c.newRequest(http.MethodGet, collectionPermissionEndpoint, nil)
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
	req, err := c.newRequest(http.MethodPut, collectionPermissionEndpoint, access)
	if err != nil {
		return err
	}

	if _, err := c.do(req, &access); err != nil {
		return err
	}

	return nil
}

func (c *client) getDatabaseAccess() (*databaseGraph, error) {
	req, err := c.newRequest(http.MethodGet, databasePermissionEndpoint, nil)
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
	req, err := c.newRequest(http.MethodPut, databasePermissionEndpoint, dbGraph)
	if err != nil {
		return err
	}

	if _, err := c.do(req, &dbGraph); err != nil {
		return err
	}

	return nil
}

func (c *client) createGroup(group *group) error {
	group.Name = GuardianGroupPrefix + group.Name
	req, err := c.newRequest(http.MethodPost, groupEndpoint, group)
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

func (c *client) findCollectionAccessGroup(access *collectionGraph, resourceID, role string) string {
group:
	for groupID, collections := range access.Groups {
		if collections[resourceID] == role {
			for collectionID, restRole := range collections {
				if collectionID == resourceID {
					continue
				}

				if restRole != "none" {
					continue group
				}
			}

			return groupID
		}
	}

	return ""
}

func (c *client) findDatabaseAccessGroup(access *databaseGraph, resourceID string, role databasePermission) string {
	expectedDatabasePermission := map[string]databasePermission{
		resourceID: role,
	}

	for groupID, databasePermissions := range access.Groups {
		if reflect.DeepEqual(databasePermissions, expectedDatabasePermission) {
			return groupID
		}
	}

	return ""
}

func (c *client) findTableAccessGroup(access *databaseGraph, databaseID string, role databasePermission) string {
	expectedDatabasePermission := map[string]databasePermission{
		databaseID: role,
	}

	for groupID, databasePermissions := range access.Groups {
		if reflect.DeepEqual(databasePermissions, expectedDatabasePermission) {
			return groupID
		}
	}

	return ""
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
		c.logger.Error(fmt.Sprintf("Failed to execute request %v with error %v", req.URL, err))
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

	if resp.StatusCode == http.StatusBadRequest {
		byteData, _ := io.ReadAll(resp.Body)
		return nil, errors.New(string(byteData))
	}

	if resp.StatusCode == http.StatusInternalServerError {
		byteData, _ := io.ReadAll(resp.Body)
		return nil, errors.New(string(byteData))
	}

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
	}
	return resp, err
}
