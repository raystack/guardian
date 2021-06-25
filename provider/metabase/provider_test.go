package metabase_test

import (
	"errors"
	"testing"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/provider/metabase"
	"github.com/stretchr/testify/assert"
)

func TestGetType(t *testing.T) {
	t.Run("should return provider type name", func(t *testing.T) {
		expectedTypeName := domain.ProviderTypeMetabase
		crypto := new(mocks.Crypto)
		p := metabase.NewProvider(expectedTypeName, crypto)

		actualTypeName := p.GetType()

		assert.Equal(t, expectedTypeName, actualTypeName)
	})
}

func TestGetResources(t *testing.T) {
	t.Run("should return error if credentials is invalid", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := metabase.NewProvider("", crypto)

		pc := &domain.ProviderConfig{
			Credentials: "invalid-creds",
		}

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.Error(t, actualError)
	})

	t.Run("should return error if there are any on client initialization", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := metabase.NewProvider("", crypto)

		expectedError := errors.New("decrypt error")
		crypto.On("Decrypt", "test-password").Return("", expectedError).Once()
		pc := &domain.ProviderConfig{
			Credentials: map[string]interface{}{
				"password": "test-password",
			},
		}

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return error if got any on getting database resources", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.MetabaseClient)
		p := metabase.NewProvider("", crypto)
		p.Clients = map[string]metabase.MetabaseClient{
			providerURN: client,
		}

		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: map[string]interface{}{},
		}
		expectedError := errors.New("client error")
		client.On("GetDatabases").Return(nil, expectedError).Once()

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return error if got any on getting collection resources", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.MetabaseClient)
		p := metabase.NewProvider("", crypto)
		p.Clients = map[string]metabase.MetabaseClient{
			providerURN: client,
		}

		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: map[string]interface{}{},
		}
		client.On("GetDatabases").Return([]*metabase.Database{}, nil).Once()
		expectedError := errors.New("client error")
		client.On("GetCollections").Return(nil, expectedError).Once()

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return list of resources and nil error on success", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.MetabaseClient)
		p := metabase.NewProvider("", crypto)
		p.Clients = map[string]metabase.MetabaseClient{
			providerURN: client,
		}

		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: map[string]interface{}{},
		}
		expectedDatabases := []*metabase.Database{
			{
				ID:   1,
				Name: "db_1",
			},
		}
		client.On("GetDatabases").Return(expectedDatabases, nil).Once()
		expectedCollections := []*metabase.Collection{
			{
				ID:   1,
				Name: "col_1",
			},
		}
		client.On("GetCollections").Return(expectedCollections, nil).Once()
		expectedResources := []*domain.Resource{
			{
				Type:        metabase.ResourceTypeDatabase,
				URN:         "1",
				ProviderURN: providerURN,
				Name:        "db_1",
				Details: map[string]interface{}{
					"auto_run_queries":            false,
					"cache_field_values_schedule": "",
					"engine":                      "",
					"metadata_sync_schedule":      "",
					"native_permissions":          "",
					"timezone":                    "",
				},
			},
			{
				Type:        metabase.ResourceTypeCollection,
				URN:         "1",
				ProviderURN: providerURN,
				Name:        "col_1",
				Details:     map[string]interface{}{},
			},
		}

		actualResources, actualError := p.GetResources(pc)

		assert.Equal(t, expectedResources, actualResources)
		assert.Nil(t, actualError)
	})
}
