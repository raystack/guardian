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

	"github.com/odpf/salt/log"

	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/plugins/providers/metabase"
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
	s.Run("should get databases and nil error on success", func() {
		s.setup()

		testRequest, err := s.getTestRequest(http.MethodGet, "/api/database?include=tables", nil)
		s.Require().NoError(err)

		databasesResponseJSON := `[{"id":1,"name":"test-Name","cache_field_values_schedule":"testCache","timezone":"test-time","auto_run_queries":true,"metadata_sync_schedule":"test-sync","engine":"test-engine","native_permissions":"per" }]`
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

		d := []*metabase.GroupResource{{Urn: "database:1", Permissions: []string{"read", "write"}}}   //expected Database Permsissions
		c := []*metabase.GroupResource{{Urn: "collection:1", Permissions: []string{"read", "write"}}} //expected Collection Permissions
		group := metabase.Group{Name: "All Users", DatabaseResources: d, CollectionResources: c}

		expectedgroupResponse := []*metabase.Group{&group}
		//expectedDatabasePermission := metabase.ResourceGroupDetails{"database:1": {{"urn": "group:1", "permissions": []string{"read", "write"}}}}
		//	expectedCollectionPermission:= metabase.ResourceGroupDetails{"collection:1": {{"urn": "group:1", "permissions": []string{"write"}}}}

		groupResponseJSON := `[{"name":"All Users","database":[{"name":"", "permission":["read","write"],"urn":"database:1","type":""}], "collection":[{"name":"","permission":["read","write"], "urn":"collection:1","type":""}] }]`
		//{&metabase.Group{ID: 999,Name: "Test-Group-Name",DatabaseResources: []*metabase.GroupResource{{Name: "Database-Name",Permissions: []string{"Viewer-Permissions"},Urn: "test_URN",Type: ""   }}}}
		groupResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(groupResponseJSON)))}
		s.mockHttpClient.On("Do", fetchGroupstestRequest).Return(&groupResponse, nil).Once()

		fetchDatabasePermissionstestRequest, err := s.getTestRequest(http.MethodGet, "/api/permissions/graph", nil)
		s.Require().NoError(err)

		expectedDatabaseGroupResponse := metabase.ResourceGroupDetails{}
		//TODO			expectedDatabasePermission
		//metabase.ResourceGroupDetails{"database:1": {{"urn": "group:1", "permission": []string{"read", "write"}}}}
		databaseResourceGroupsResponseJSON := `{[ "database:1": [{"urn":"group:1", "permissions":["read","write"]}] ]}`

		//Working with minor fix required-  `{ "database:1": [{"urn":"group:1", "permissions":["read","write"]}] }`

		//`{"database:1": { "permissions":["read","write"] } }`
		//`"database:1": {{ "permissions":["read","write"], "urn":"group:1" }} `

		//[[{"name":"", "permission":["read","write"],"urn":"database:1","type":""}]] `

		//`{ "database:1": [ { "permission":["read","write"],"urn":"group:1" } ] }`

		//	[{"urn":"group:1","permission":["read","write"] }]`
		databaseResourceGroupsResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(databaseResourceGroupsResponseJSON)))}
		s.mockHttpClient.On("Do", fetchDatabasePermissionstestRequest).Return(&databaseResourceGroupsResponse, nil).Once()

		fetchCollectionPermissionstestRequest, err := s.getTestRequest(http.MethodGet, "/api/collection/graph", nil)
		s.Require().NoError(err)

		fetchCollectionPermissionsResponseJSON := ``
		fetchCollectionPermissionsResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(fetchCollectionPermissionsResponseJSON)))}
		s.mockHttpClient.On("Do", fetchCollectionPermissionstestRequest).Return(&fetchCollectionPermissionsResponse, nil).Once()

		actualGroupResponse, actualDatabaseGroupResponse, _, err := s.client.GetGroups()

		s.Nil(err)
		s.Equal(expectedgroupResponse, actualGroupResponse)
		s.Equal(expectedDatabaseGroupResponse, actualDatabaseGroupResponse)
	})

}

func (s *ClientTestSuite) TestGrantDatabaseAccess() {
	s.Run("should grant access to database and nil error on success", func() {
		s.setup()

		testRequest, err := s.getTestRequest(http.MethodGet, "/api/permissions/graph", nil)
		databasesResponseJSON := `{"revision":1,"groups":[{"gid_1":[{"db_1":{"schema":"all"}},{"db_2":{"native":"write"}}]}  ] }`
		databaseResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(databasesResponseJSON)))}
		s.mockHttpClient.On("Do", testRequest).Return(&databaseResponse, nil).Once()

		s.Require().NoError(err)

		expectedDatabase := &metabase.Database{
			Name: "test-database",
			ID:   999,
		}
		//[{"name":"All Users","database":[{"name":"", "permission":["read","write"],"urn":"database:1","type":""}], "collection":[{"name":"","permission":["read","write"], "urn":"collection:1","type":""}] }]
		groups := map[string]*metabase.Group{
			"gid_1": {ID: 1, Name: "db_1"},
			"gid_2": {ID: 2, Name: "db_2"},
		}
		role := "schemas:all"
		resource := expectedDatabase
		email := "test-email@gojek.com" //test for getuser()
		getUserUrl := fmt.Sprintf("/api/user?query=%s", email)
		testUserRequest, err := s.getTestRequest(http.MethodGet, getUserUrl, nil)

		userResponseJSON := `{"id":1,"email":"test-email@gojek.com"}`
		userResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(userResponseJSON)))}
		s.mockHttpClient.On("Do", testUserRequest).Return(&userResponse, nil).Once()

		//test for addGroupMember()

		membershipURL := "/api/permissions/membership"
		membershipRequest := `{"group_id":1,"user_id":5}`
		testGroupMemberRequest, err := s.getTestRequest(http.MethodPost, membershipURL, membershipRequest)

		response := http.Response{StatusCode: 200}
		s.mockHttpClient.On("Do", testGroupMemberRequest).Return(&response, nil).Once()

		actualError := s.client.GrantDatabaseAccess(resource, email, role, groups)
		s.Nil(actualError)
	})
}
