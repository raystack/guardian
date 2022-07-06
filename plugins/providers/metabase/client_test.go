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

	t.Run("should get collections and nil error on success", func(t *testing.T) {
		mockHttpClient := new(mocks.HTTPClient)
		config := getTestClientConfig()
		config.HTTPClient = mockHttpClient
		logger := log.NewLogrus(log.LogrusWithLevel("info"))

		sessionToken := "93df71b4-6887-46bd-b4bf-7ad3b94bd6fe"
		responseJSON := `{"id":"` + sessionToken + `"}`
		response := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(responseJSON)))}

		mockHttpClient.On("Do", mock.Anything).Return(&response, nil).Once()

		client, err := metabase.NewClient(config, logger)
		assert.Nil(t, err)
		assert.NotNil(t, client)

		testRequest, err := getTestRequest(sessionToken, http.MethodGet, fmt.Sprintf("%s%s", config.Host, "/api/collection"), nil)
		assert.Nil(t, err)

		collectionResponseJSON := `[{"authority_level":null,"name":"Our analytics","id":"root","parent_id":null,"effective_location":null,"effective_ancestors":[],"can_write":true},{"authority_level":null,"description":null,"archived":false,"slug":"cabfares","color":"#509EE3","can_write":true,"name":"CabFares","personal_owner_id":null,"id":2,"location":"/","namespace":null},{"authority_level":null,"description":null,"archived":false,"slug":"countries","color":"#509EE3","can_write":true,"name":"Countries","personal_owner_id":null,"id":5,"location":"/4/","namespace":null},{"authority_level":null,"description":null,"archived":false,"slug":"ds_analysis","color":"#509EE3","can_write":true,"name":"DS Analysis","personal_owner_id":null,"id":3,"location":"/2/","namespace":null},{"authority_level":null,"description":null,"archived":false,"slug":"ds_analysis","color":"#509EE3","can_write":true,"name":"DS Analysis","personal_owner_id":null,"id":6,"location":"/4/5/","namespace":null},{"authority_level":null,"description":null,"archived":false,"slug":"spending","color":"#509EE3","can_write":true,"name":"Spending","personal_owner_id":null,"id":4,"location":"/","namespace":null},{"authority_level":null,"description":null,"archived":false,"slug":"summary","color":"#509EE3","can_write":true,"name":"Summary","personal_owner_id":null,"id":7,"location":"/2/3/","namespace":null},{"authority_level":null,"description":null,"archived":false,"slug":"alex_s_personal_collection","color":"#31698A","can_write":true,"name":"Alex's Personal Collection","personal_owner_id":1,"id":1,"location":"/","namespace":null}]`
		collectionResponse := http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte(collectionResponseJSON)))}
		mockHttpClient.On("Do", testRequest).Return(&collectionResponse, nil).Once()

		expectedCollections := []metabase.Collection{
			{ID: float64(2), Name: "CabFares", Slug: "cabfares", Location: "/"},
			{ID: float64(5), Name: "Spending/Countries", Slug: "countries", Location: "/4/"},
			{ID: float64(3), Name: "CabFares/DS Analysis", Slug: "ds_analysis", Location: "/2/"},
			{ID: float64(6), Name: "Spending/Countries/DS Analysis", Slug: "ds_analysis", Location: "/4/5/"},
			{ID: float64(4), Name: "Spending", Slug: "spending", Location: "/", Namespace: ""},
			{ID: float64(7), Name: "CabFares/DS Analysis/Summary", Slug: "summary", Location: "/2/3/"},
		}

		result, err1 := client.GetCollections()
		var collections []metabase.Collection
		for _, coll := range result {
			collections = append(collections, *coll)
		}
		assert.Nil(t, err1)
		assert.ElementsMatch(t, expectedCollections, collections)
	})
}

func getTestClientConfig() *metabase.ClientConfig {
	return &metabase.ClientConfig{
		Username: "test-username",
		Password: "test-password",
		Host:     "http://localhost",
	}
}

func getTestRequest(sessionToken, method, url string, body interface{}) (*http.Request, error) {
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Metabase-Session", sessionToken)
	return req, nil
}
