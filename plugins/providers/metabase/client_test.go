package metabase_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/goto/salt/log"

	"github.com/goto/guardian/mocks"
	"github.com/goto/guardian/plugins/providers/metabase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestNewClient(t *testing.T) {
	t.Run("should return error if config is invalid", func(t *testing.T) {
		invalidConfig := &metabase.ClientConfig{}
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		actualClient, actualError := metabase.NewClient(invalidConfig, logger)

		assert.Nil(t, actualClient)
		assert.Error(t, actualError)
	})

	t.Run("should return error if config.Host is not a valid url", func(t *testing.T) {
		invalidHostConfig := &metabase.ClientConfig{
			Username: "test-username",
			Password: "test-password",
			Host:     "invalid-url",
		}
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		actualClient, actualError := metabase.NewClient(invalidHostConfig, logger)

		assert.Nil(t, actualClient)
		assert.Error(t, actualError)
	})

	t.Run("should return error if got error retrieving the session token", func(t *testing.T) {
		mockHttpClient := new(mocks.HTTPClient)
		config := &metabase.ClientConfig{
			Username:   "test-username",
			Password:   "test-password",
			Host:       "http://localhost",
			HTTPClient: mockHttpClient,
		}
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		expectedError := errors.New("request error")
		mockHttpClient.On("Do", mock.Anything).Return(nil, expectedError).Once()

		actualClient, actualError := metabase.NewClient(config, logger)

		mockHttpClient.AssertExpectations(t)
		assert.Nil(t, actualClient)
		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return client and nil error on success", func(t *testing.T) {
		// TODO: test http request execution
		mockHttpClient := new(mocks.HTTPClient)
		config := &metabase.ClientConfig{
			Username:   "test-username",
			Password:   "test-password",
			Host:       "http://localhost",
			HTTPClient: mockHttpClient,
		}
		logger := log.NewLogrus(log.LogrusWithLevel("info"))

		sessionToken := "93df71b4-6887-46bd-b4bf-7ad3b94bd6fe"
		responseJSON := `{"id":"` + sessionToken + `"}`
		response := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(responseJSON)))}
		mockHttpClient.On("Do", mock.Anything).Return(&response, nil).Once()

		_, actualError := metabase.NewClient(config, logger)
		mockHttpClient.AssertExpectations(t)
		assert.Nil(t, actualError)
	})
}

type ClientTestSuite struct {
	suite.Suite

	mockHttpClient *mocks.HTTPClient
	client         metabase.MetabaseClient
	sessionToken   string
	host           string
}

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (s *ClientTestSuite) setup() {
	logger := log.NewNoop()
	s.mockHttpClient = new(mocks.HTTPClient)

	s.sessionToken = "93df71b4-6887-46bd-b4bf-7ad3b94bd6fe"
	sessionResponse := http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"` + s.sessionToken + `"}`))),
	}
	s.mockHttpClient.On("Do", mock.Anything).Return(&sessionResponse, nil).Once()

	s.host = "http://localhost"
	client, err := metabase.NewClient(&metabase.ClientConfig{
		Username:   "test-username",
		Password:   "test-password",
		Host:       s.host,
		HTTPClient: s.mockHttpClient,
	}, logger)
	s.Require().NoError(err)
	s.client = client
}

func (s *ClientTestSuite) TestGetCollections() {
	s.Run("should get collections and nil error on success", func() {
		s.setup()

		testRequest, err := s.getTestRequest(http.MethodGet, "/api/collection", nil)
		s.Require().NoError(err)

		collectionResponseJSON := `[{"authority_level":null,"name":"Our analytics","id":"root","parent_id":null,"effective_location":null,"effective_ancestors":[],"can_write":true},{"authority_level":null,"description":null,"archived":false,"slug":"cabfares","color":"#509EE3","can_write":true,"name":"CabFares","personal_owner_id":null,"id":2,"location":"/","namespace":null},{"authority_level":null,"description":null,"archived":false,"slug":"countries","color":"#509EE3","can_write":true,"name":"Countries","personal_owner_id":null,"id":5,"location":"/4/","namespace":null},{"authority_level":null,"description":null,"archived":false,"slug":"ds_analysis","color":"#509EE3","can_write":true,"name":"DS Analysis","personal_owner_id":null,"id":3,"location":"/2/","namespace":null},{"authority_level":null,"description":null,"archived":false,"slug":"ds_analysis","color":"#509EE3","can_write":true,"name":"DS Analysis","personal_owner_id":null,"id":6,"location":"/4/5/","namespace":null},{"authority_level":null,"description":null,"archived":false,"slug":"spending","color":"#509EE3","can_write":true,"name":"Spending","personal_owner_id":null,"id":4,"location":"/","namespace":null},{"authority_level":null,"description":null,"archived":false,"slug":"summary","color":"#509EE3","can_write":true,"name":"Summary","personal_owner_id":null,"id":7,"location":"/2/3/","namespace":null},{"authority_level":null,"description":null,"archived":false,"slug":"alex_s_personal_collection","color":"#31698A","can_write":true,"name":"Alex's Personal Collection","personal_owner_id":1,"id":1,"location":"/","namespace":null}]`
		collectionResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(collectionResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&collectionResponse, nil).Once()

		expectedCollections := []metabase.Collection{
			{ID: float64(2), Name: "CabFares", Slug: "cabfares", Location: "/"},
			{ID: float64(5), Name: "Spending/Countries", Slug: "countries", Location: "/4/"},
			{ID: float64(3), Name: "CabFares/DS Analysis", Slug: "ds_analysis", Location: "/2/"},
			{ID: float64(6), Name: "Spending/Countries/DS Analysis", Slug: "ds_analysis", Location: "/4/5/"},
			{ID: float64(4), Name: "Spending", Slug: "spending", Location: "/", Namespace: ""},
			{ID: float64(7), Name: "CabFares/DS Analysis/Summary", Slug: "summary", Location: "/2/3/"},
		}

		result, err1 := s.client.GetCollections()
		var collections []metabase.Collection
		for _, coll := range result {
			collections = append(collections, *coll)
		}
		s.Nil(err1)
		s.ElementsMatch(expectedCollections, collections)
	})
}

func (s *ClientTestSuite) TestGetDatabases() {
	s.Run("should return error bad request, status code 400", func() {
		s.setup()

		testRequest, err := s.getTestRequest(http.MethodGet, "/api/database?include=tables", nil)
		s.Require().NoError(err)

		databaseResponse := http.Response{StatusCode: 400, Body: ioutil.NopCloser(bytes.NewReader([]byte(nil)))}
		s.mockHttpClient.On("Do", testRequest).Return(&databaseResponse, nil).Once()

		result, err1 := s.client.GetDatabases()
		s.Nil(result)
		s.Error(err1)
	})

	s.Run("should return error internal server error, status code 500", func() {
		s.setup()

		testRequest, err := s.getTestRequest(http.MethodGet, "/api/database?include=tables", nil)
		s.Require().NoError(err)

		databaseResponse := http.Response{StatusCode: 500, Body: ioutil.NopCloser(bytes.NewReader([]byte(nil)))}
		s.mockHttpClient.On("Do", testRequest).Return(&databaseResponse, nil).Once()

		result, err1 := s.client.GetDatabases()
		s.Nil(result)
		s.Error(err1)
	})

	s.Run("if user unauthorised, get the session again and retry. Should return databases on success", func() {
		s.setup()

		testRequest, err := s.getTestRequest(http.MethodGet, "/api/database?include=tables", nil)
		s.Require().NoError(err)

		databaseResponse := http.Response{StatusCode: 401, Body: ioutil.NopCloser(bytes.NewReader([]byte(nil)))}
		s.mockHttpClient.On("Do", testRequest).Return(&databaseResponse, nil).Once()

		s.sessionToken = "93df71b4-6887-46bd-b4bf-7ad3b94bd6fe"
		sessionResponse := http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"` + s.sessionToken + `"}`))),
		}
		s.mockHttpClient.On("Do", mock.Anything).Return(&sessionResponse, nil).Once()

		databasesResponseJSON := `[{"id":1,"name":"test-Name","cache_field_values_schedule":"testCache","timezone":"test-time","auto_run_queries":true,"metadata_sync_schedule":"test-sync","engine":"test-engine","native_permissions":"per" }]`
		correctdatabaseResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(databasesResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&correctdatabaseResponse, nil).Once()

		expectedDatabases := []metabase.Database{
			{ID: 1, Name: "test-Name", CacheFieldValuesSchedule: "testCache", Timezone: "test-time", AutoRunQueries: true, MetadataSyncSchedule: "test-sync", Engine: "test-engine", NativePermissions: "per"},
		}

		result, err1 := s.client.GetDatabases()
		var databases []metabase.Database
		for _, db := range result {
			databases = append(databases, *db)
		}
		s.Nil(err1)
		s.Equal(expectedDatabases, databases)
		s.mockHttpClient.AssertExpectations(s.T())
	})

	s.Run("should get databases and nil error on success", func() {
		s.setup()

		testRequest, err := s.getTestRequest(http.MethodGet, "/api/database?include=tables", nil)
		s.Require().NoError(err)

		databasesResponseJSON := `{"data":[{"id":1,"name":"test-Name","cache_field_values_schedule":"testCache","timezone":"test-time","auto_run_queries":true,"metadata_sync_schedule":"test-sync","engine":"test-engine","native_permissions":"per" }]}`
		databaseResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(databasesResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&databaseResponse, nil).Once()

		expectedDatabases := []metabase.Database{
			{ID: 1, Name: "test-Name", CacheFieldValuesSchedule: "testCache", Timezone: "test-time", AutoRunQueries: true, MetadataSyncSchedule: "test-sync", Engine: "test-engine", NativePermissions: "per"},
			//Tables: []metabase.Table{{ID: 2, Name: "tab1", DbId: 1, Database: &domain.Resource{ID: "5", ProviderType: "metabase", ProviderURN: "test-URN", Type: "database"} }} },
		}

		result, err1 := s.client.GetDatabases()
		var databases []metabase.Database
		for _, db := range result {
			databases = append(databases, *db)
		}
		s.Nil(err1)
		s.Equal(expectedDatabases, databases)
		s.mockHttpClient.AssertExpectations(s.T())
	})
}

func (s *ClientTestSuite) getTestRequest(method, path string, body interface{}) (*http.Request, error) {
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}
	url := fmt.Sprintf("%s%s", s.host, path)
	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Metabase-Session", s.sessionToken)
	return req, nil
}

func (s *ClientTestSuite) TestGetGroups() {
	s.Run("should fetch groups and nil error on success", func() {
		s.setup()

		fetchGroupstestRequest, err := s.getTestRequest(http.MethodGet, "/api/permissions/group", nil)
		s.Require().NoError(err)

		//test fetech group permissions
		groups := []*metabase.Group{}
		d := []*metabase.GroupResource{{Urn: "database:1", Permissions: []string{"schema:all"}}, {Urn: "database:2", Permissions: []string{"native:write"}}, {Urn: "database:3", Permissions: []string{"schema:all"}}}                          //expected Database Permsissions
		c := []*metabase.GroupResource{{Name: "All Users", Urn: "collection:1", Permissions: []string{"read"}}, {Name: "All Users", Urn: "collection:2", Permissions: []string{"write"}}, {Urn: "collection:3", Permissions: []string{"read"}}} //expected Collection Permissions
		group := metabase.Group{ID: 1, Name: "All Users", DatabaseResources: d, CollectionResources: c}

		groups = append(groups, &group)
		expectedgroupResponse := groups

		groupResponseJSON := `[{"id":1,"name":"All Users","database":[{"permission":["schema:all"],"urn":"database:1"},{"permission":["native:write"],"urn":"database:2"}], "collection":[{"name":"All Users","permission":["read"], "urn":"collection:1"},{"name":"All Users","permission":["write"], "urn":"collection:2"}] }]`
		groupResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(groupResponseJSON)))}
		s.mockHttpClient.On("Do", fetchGroupstestRequest).Return(&groupResponse, nil).Once()

		//test fetch database permissions
		fetchDatabasePermissionstestRequest, err := s.getTestRequest(http.MethodGet, "/api/permissions/graph", nil)
		s.Require().NoError(err)
		expectedDatabaseGroupResponse := metabase.ResourceGroupDetails{"database:3": []map[string]interface{}{{"name": "All Users", "permissions": []string{"schema:all"}, "urn": "group:1"}}}
		databaseResourceGroupsResponseJSON := `{"groups":{"1":{ "3":{"schema":"all"}} } }`

		databaseResourceGroupsResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(databaseResourceGroupsResponseJSON)))}
		s.mockHttpClient.On("Do", fetchDatabasePermissionstestRequest).Return(&databaseResourceGroupsResponse, nil).Once()

		fetchCollectionPermissionstestRequest, err := s.getTestRequest(http.MethodGet, "/api/collection/graph", nil)
		s.Require().NoError(err)

		fetchCollectionPermissionsResponseJSON := `{"groups":{"1":{ "3":"read"} } }`
		fetchCollectionPermissionsResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(fetchCollectionPermissionsResponseJSON)))}
		s.mockHttpClient.On("Do", fetchCollectionPermissionstestRequest).Return(&fetchCollectionPermissionsResponse, nil).Once()

		actualGroupResponse, actualDatabaseGroupResponse, _, err := s.client.GetGroups()

		s.Nil(err)
		s.Equal(expectedgroupResponse, actualGroupResponse)
		s.Equal(expectedDatabaseGroupResponse, actualDatabaseGroupResponse)
		s.mockHttpClient.AssertExpectations(s.T())
	})
}

func (s *ClientTestSuite) TestGrantDatabaseAccess() {
	s.Run("should grant access to database and nil error on success", func() {
		s.setup()

		testRequest, err := s.getTestRequest(http.MethodGet, "/api/permissions/graph", nil)
		s.Require().NoError(err)
		databasesResponseJSON := `{"revision":1,"groups":{"gid_1":{"db_1":{"schema":"all"},"db_2":{"native":"write"}}}}`
		databaseResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(databasesResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&databaseResponse, nil).Once()

		// createGroup mock
		createGroupResponseJSON := `{"id":1,"name":"","members":[{"id":1,"email":"","membership_id":1,"group_ids":[1,2]}]}`
		createGroupResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(createGroupResponseJSON)))}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).
			Return(&createGroupResponse, nil).Once()

		// updateDatabaseAccess mock
		updateDatabaseAccessResponseJSON := `{"revision":1,"groups":{"gid_1":{"db_1":{"schema":"all"},"db_2":{"native":"write"}}}}`
		updateDatabaseAccessResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(updateDatabaseAccessResponseJSON)))}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).
			Return(&updateDatabaseAccessResponse, nil).Once()

		email := "test-email@gojek.com" //test for getuser()

		getUserUrl := fmt.Sprintf("/api/user?query=%s", email)
		testUserRequest, err := s.getTestRequest(http.MethodGet, getUserUrl, nil)
		s.Require().NoError(err)
		userResponseJSON := `{"data":[{"id":1,"email":"test-email@gojek.com"}]}`
		userResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(userResponseJSON)))}
		s.mockHttpClient.On("Do", testUserRequest).Return(&userResponse, nil).Once()

		//test for addGroupMember()

		response := http.Response{StatusCode: 200, Body: ioutil.NopCloser(nil)}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&response, nil).Once()

		role := "schemas:all"
		expectedDatabase := &metabase.Database{
			Name: "test-database",
			ID:   999,
		}
		resource := expectedDatabase
		groups := map[string]*metabase.Group{
			"gid_1": {ID: 1, Name: "db_1"},
			"gid_2": {ID: 2, Name: "db_2"},
		}
		actualError := s.client.GrantDatabaseAccess(resource, email, role, groups)
		s.Nil(actualError)
		s.mockHttpClient.AssertExpectations(s.T())
	})
}

func (s *ClientTestSuite) TestGrantCollectionAccess() {
	s.Run("should grant access to collection and nil error on success", func() {
		s.setup()

		testRequest, err := s.getTestRequest(http.MethodGet, "/api/collection/graph", nil)
		s.Require().NoError(err)
		collectionResponseJSON := `{"revision":1,"groups":{"51": {"999":"write"},"52":{"1000":"read"} } }`
		collectionResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(collectionResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&collectionResponse, nil).Once()

		groupResJSON := `{"id":53,"name":"_guardian_collection_999_read"}`
		res := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(groupResJSON)))}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&res, nil).Once()

		updatedcollectionResponseJSON := `{"revision":1,"groups":{"51": {"999":"write"},"52":{"1000":"read"},"53":{"999":"read"} } }`
		updatedcollectionResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(updatedcollectionResponseJSON)))}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&updatedcollectionResponse, nil).Once()

		email := "test-email@gojek.com"

		getUserUrl := fmt.Sprintf("/api/user?query=%s", email)
		testUserRequest, err := s.getTestRequest(http.MethodGet, getUserUrl, nil)
		s.Require().NoError(err)
		userResponseJSON := `{"data":[{"id":1,"email":"test-email@gojek.com"}]}`
		userResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(userResponseJSON)))}
		s.mockHttpClient.On("Do", testUserRequest).Return(&userResponse, nil).Once()

		res2 := http.Response{StatusCode: 200, Body: ioutil.NopCloser(nil)}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&res2, nil).Once()

		role := "read" //valid collection roles are "read" and "write"
		expectedCollection := &metabase.Collection{
			ID:   999,
			Name: "test-collection",
		}
		resource := expectedCollection
		actualError := s.client.GrantCollectionAccess(resource, email, role)
		s.Nil(actualError)
		s.mockHttpClient.AssertExpectations(s.T())
	})

	s.Run("should grant access to collection and nil error on success", func() {
		s.setup()

		testRequest, err := s.getTestRequest(http.MethodGet, "/api/collection/graph", nil)
		s.Require().NoError(err)
		collectionResponseJSON := `{"revision":1,"groups":{"51": {"999":"write"},"52":{"1000":"read"} } }`
		collectionResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(collectionResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&collectionResponse, nil).Once()

		email := "test-email@gojek.com"

		getUserUrl := fmt.Sprintf("/api/user?query=%s", email)
		testUserRequest, err := s.getTestRequest(http.MethodGet, getUserUrl, nil)
		s.Require().NoError(err)
		userResponseJSON := `{"data":[{"id":1,"email":"test-email@gojek.com"}]}`
		userResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(userResponseJSON)))}
		s.mockHttpClient.On("Do", testUserRequest).Return(&userResponse, nil).Once()

		res := http.Response{StatusCode: 200, Body: ioutil.NopCloser(nil)}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&res, nil).Once()

		role := "write" //valid collection roles are "read" and "write"
		expectedCollection := &metabase.Collection{
			ID:   999,
			Name: "test-collection",
		}
		resource := expectedCollection
		actualError := s.client.GrantCollectionAccess(resource, email, role)
		s.Nil(actualError)
		s.mockHttpClient.AssertExpectations(s.T())
	})
}

func (s *ClientTestSuite) TestRevokeCollectionAccess() {
	s.Run("should grant access to collection and nil error on success", func() {
		s.setup()

		testRequest, err := s.getTestRequest(http.MethodGet, "/api/collection/graph", nil) //test getcollection access
		s.Require().NoError(err)
		collectionResponseJSON := `{"revision":1,"groups":{"51": {"999":"write"},"52":{"1000":"read"} } }`
		collectionResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(collectionResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&collectionResponse, nil).Once()

		email := "test-email@gojek.com"

		groupIDInt := 51 //test GetGroups
		url := fmt.Sprintf("/api/permissions/group/%d", groupIDInt)
		req, err2 := s.getTestRequest(http.MethodGet, url, nil)
		s.Require().NoError(err2)
		groupResponseJSON := `{"id":51 ,"name":"999", "members":[{"id":1,"email":"test-email@gojek.com","membership_id":500,"group_ids":[51,52,53]}] }`
		groupResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(groupResponseJSON)))}
		s.mockHttpClient.On("Do", req).Return(&groupResponse, nil).Once()

		membershipID := 500 //test removeGroupMember
		revokeGroupMemeberURL := fmt.Sprintf("/api/permissions/membership/%d", membershipID)
		revokeGroupMemeberRequest, err3 := s.getTestRequest(http.MethodDelete, revokeGroupMemeberURL, nil)
		s.Require().NoError(err3)
		revokeGroupMemeberResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(nil)}
		s.mockHttpClient.On("Do", revokeGroupMemeberRequest).Return(&revokeGroupMemeberResponse, nil).Once()

		role := "write" //valid collection roles are "read" and "write"
		expectedCollection := &metabase.Collection{
			ID:   999,
			Name: "test-collection",
		}
		resource := expectedCollection
		actualError := s.client.RevokeCollectionAccess(resource, email, role)
		s.Nil(actualError)
		s.mockHttpClient.AssertExpectations(s.T())
	})
}

func (s *ClientTestSuite) TestGrantGroupAccesss() {
	s.Run("should return nil if user is already part of the group", func() {
		s.setup()

		email := "test-email@gojek.com"
		getUserUrl := fmt.Sprintf("/api/user?query=%s", email)
		testUserRequest, err := s.getTestRequest(http.MethodGet, getUserUrl, nil)
		s.Require().NoError(err)
		userResponseJSON := `{"data":[{"id":1,"email":"test-email@gojek.com","membership_id":500,"group_ids":[51,52,53]}]}`
		userResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(userResponseJSON)))}
		s.mockHttpClient.On("Do", testUserRequest).Return(&userResponse, nil).Once()

		groupID := 53

		actualError := s.client.GrantGroupAccess(groupID, email)

		s.Nil(actualError)
		s.mockHttpClient.AssertExpectations(s.T())
	})

	s.Run("should add member to group and nil error on success", func() {
		s.setup()

		email := "test-email@gojek.com"
		getUserUrl := fmt.Sprintf("/api/user?query=%s", email)
		testUserRequest, err := s.getTestRequest(http.MethodGet, getUserUrl, nil)
		s.Require().NoError(err)
		userResponseJSON := `{"data":[{"id":1,"email":"test-email@gojek.com","membership_id":500,"group_ids":[51,52,53]}]}`
		userResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(userResponseJSON)))}
		s.mockHttpClient.On("Do", testUserRequest).Return(&userResponse, nil).Once()

		groupID := 54

		res := http.Response{StatusCode: 200, Body: ioutil.NopCloser(nil)}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&res, nil).Once()

		actualError := s.client.GrantGroupAccess(groupID, email)

		s.Nil(actualError)
		s.mockHttpClient.AssertExpectations(s.T())
	})
}

func (s *ClientTestSuite) TestGrantTableAccess() {
	s.Run("should create the group, if not already there, grant access to database and nil error on success", func() {
		s.setup()

		//test get database access
		testRequest, err := s.getTestRequest(http.MethodGet, "/api/permissions/graph", nil)
		s.Require().NoError(err)
		databasesResponseJSON := `{"revision":1,"groups":{"999":{"1":{"schema":"all"},"2":{"native":"write"}}}}`
		databaseResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(databasesResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&databaseResponse, nil).Once()

		expectedTable := &metabase.Table{
			Name: "test-table",
			ID:   999,
			DbId: 1,
		}
		resource := expectedTable
		role := "all"

		groups := map[string]*metabase.Group{
			"table_1_999_all1": {ID: 1, Name: "1", DatabaseResources: []*metabase.GroupResource{{Permissions: []string{"schema:all"}, Urn: "database:1"}}}}

		groupResponseJSON := `{"id":0,"name":"_guardian_table_1_999_all"}`
		response2 := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(groupResponseJSON)))} // test for createGroup
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&response2, nil).Once()

		updateDatabaseAccessResponseJSON := `{"revision":1,"groups":{"999":{"1":{"schema":"all"},"2":{"native":"write"}},"1":{"1":{"schemas":{"public":{"999":"all"} } } }  }}`
		updateDatabaseAccessResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(updateDatabaseAccessResponseJSON)))}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).
			Return(&updateDatabaseAccessResponse, nil).Once()

		email := "test-email@gojek.com" //test for getuser()

		getUserUrl := fmt.Sprintf("/api/user?query=%s", email)
		testUserRequest, err := s.getTestRequest(http.MethodGet, getUserUrl, nil)
		s.Require().NoError(err)
		userResponseJSON := `{"data":[{"id":1,"email":"test-email@gojek.com"}]}`
		userResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(userResponseJSON)))}
		s.mockHttpClient.On("Do", testUserRequest).Return(&userResponse, nil).Once()

		response := http.Response{StatusCode: 200, Body: ioutil.NopCloser(nil)} // test for addGroupMember
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&response, nil).Once()

		actualError := s.client.GrantTableAccess(resource, email, role, groups)

		s.Nil(actualError)
	})

	s.Run("if group already there should grant access to database and nil error on success", func() {
		s.setup()

		//test get database access
		testRequest, err := s.getTestRequest(http.MethodGet, "/api/permissions/graph", nil)
		s.Require().NoError(err)
		databasesResponseJSON := `{"revision":1,"groups":{"999":{"1":{"schema":"all"},"2":{"native":"write"}}}}`
		databaseResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(databasesResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&databaseResponse, nil).Once()

		expectedTable := &metabase.Table{
			Name: "test-table",
			ID:   999,
			DbId: 1,
		}
		resource := expectedTable
		role := "all"

		groups := map[string]*metabase.Group{
			"table_1_999_all": {ID: 1, Name: "1", DatabaseResources: []*metabase.GroupResource{{Permissions: []string{"schema:all"}, Urn: "database:1"}}}}

		updateDatabaseAccessResponseJSON := `{"revision":1,"groups":{"999":{"1":{"schema":"all"},"2":{"native":"write"}},"1":{"1":{"schemas":{"public":{"999":"all"} } } }  }}`
		updateDatabaseAccessResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(updateDatabaseAccessResponseJSON)))}
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).
			Return(&updateDatabaseAccessResponse, nil).Once()

		email := "test-email@gojek.com" //test for getuser()

		getUserUrl := fmt.Sprintf("/api/user?query=%s", email)
		testUserRequest, err := s.getTestRequest(http.MethodGet, getUserUrl, nil)
		s.Require().NoError(err)
		userResponseJSON := `{"data":[{"id":1,"email":"test-email@gojek.com"}]}`
		userResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(userResponseJSON)))}
		s.mockHttpClient.On("Do", testUserRequest).Return(&userResponse, nil).Once()

		response := http.Response{StatusCode: 200, Body: ioutil.NopCloser(nil)} // test for addGroupMember
		s.mockHttpClient.On("Do", mock.AnythingOfType("*http.Request")).Return(&response, nil).Once()

		actualError := s.client.GrantTableAccess(resource, email, role, groups)

		s.Nil(actualError)
	})
}

func (s *ClientTestSuite) TestRevokeDatabaseAccess() {
	s.Run("should grant access to database and nil error on success", func() {
		s.setup()

		testRequest, err := s.getTestRequest(http.MethodGet, "/api/permissions/graph", nil)
		s.Require().NoError(err)
		databasesResponseJSON := `{"revision":1,"groups":{"51":{"999":{"schemas":"all"}}}}`
		databaseResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(databasesResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&databaseResponse, nil).Once()

		expectedDatabase := &metabase.Database{
			Name: "test-database",
			ID:   999,
		}
		resource := expectedDatabase
		email := "test-email@gojek.com"
		role := "schemas:all"

		groupIDInt := 51 //test GetGroup(groupID)
		url := fmt.Sprintf("/api/permissions/group/%d", groupIDInt)
		req, err2 := s.getTestRequest(http.MethodGet, url, nil)
		s.Require().NoError(err2)
		groupResponseJSON := `{"id":51 ,"name":"999", "members":[{"id":1,"email":"test-email@gojek.com","membership_id":500,"group_ids":[51]}] }`
		groupResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(groupResponseJSON)))}
		s.mockHttpClient.On("Do", req).Return(&groupResponse, nil).Once()

		membershipID := 500 //test removeGroupMember
		revokeGroupMemeberURL := fmt.Sprintf("/api/permissions/membership/%d", membershipID)
		revokeGroupMemeberRequest, err3 := s.getTestRequest(http.MethodDelete, revokeGroupMemeberURL, nil)
		s.Require().NoError(err3)
		revokeGroupMemeberResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(nil)}
		s.mockHttpClient.On("Do", revokeGroupMemeberRequest).Return(&revokeGroupMemeberResponse, nil).Once()

		actualError := s.client.RevokeDatabaseAccess(resource, email, role)

		s.Nil(actualError)
	})
}
