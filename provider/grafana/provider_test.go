package grafana_test

import (
	"errors"
	"testing"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/provider/grafana"
	"github.com/stretchr/testify/assert"
)

func TestGetType(t *testing.T) {
	t.Run("should return provider type name", func(t *testing.T) {
		expectedTypeName := domain.ProviderTypeGrafana
		crypto := new(mocks.Crypto)
		p := grafana.NewProvider(expectedTypeName, crypto)

		actualTypeName := p.GetType()

		assert.Equal(t, expectedTypeName, actualTypeName)
	})
}

func TestGetResources(t *testing.T) {
	t.Run("should return error if credentials is invalid", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := grafana.NewProvider("", crypto)

		pc := &domain.ProviderConfig{
			Credentials: "invalid-creds",
		}

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.Error(t, actualError)
	})

	t.Run("should return error if there are any on client initialization", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := grafana.NewProvider("", crypto)

		expectedError := errors.New("decrypt error")
		crypto.On("Decrypt", "test-api-key").Return("", expectedError).Once()
		pc := &domain.ProviderConfig{
			Credentials: map[string]interface{}{
				"api_key": "test-api-key",
			},
		}

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return error if got any on getting folder resources", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.GrafanaClient)
		p := grafana.NewProvider("", crypto)
		p.Clients = map[string]grafana.GrafanaClient{
			providerURN: client,
		}

		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: map[string]interface{}{},
		}
		expectedError := errors.New("client error")
		client.On("GetFolders").Return(nil, expectedError).Once()

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return error if got any on getting dashboard resources", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.GrafanaClient)
		p := grafana.NewProvider("", crypto)
		p.Clients = map[string]grafana.GrafanaClient{
			providerURN: client,
		}

		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: map[string]interface{}{},
		}
		expectedError := errors.New("client error")
		expectedFolders := []*grafana.Folder{
			{
				ID:    1,
				Title: "fd_1",
			},
		}
		client.On("GetFolders").Return(expectedFolders, nil).Once()
		client.On("GetDashboards", 1).Return(nil, expectedError).Times(2)

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return list of resources and nil error on success", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.GrafanaClient)
		p := grafana.NewProvider("", crypto)
		p.Clients = map[string]grafana.GrafanaClient{
			providerURN: client,
		}

		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: map[string]interface{}{},
		}
		expectedFolders := []*grafana.Folder{
			{
				ID:    1,
				Title: "fd_1",
			},
		}
		client.On("GetFolders").Return(expectedFolders, nil).Once()
		expectedDashboards := []*grafana.Dashboard{
			{
				ID:    1,
				Title: "db_1",
			},
		}
		client.On("GetDashboards", 1).Return(expectedDashboards, nil).Once()
		expectedResources := []*domain.Resource{
			{
				Type:        grafana.ResourceTypeDashboard,
				URN:         "1",
				ProviderURN: providerURN,
				Name:        "db_1",
				Details:     map[string]interface{}{},
			},
		}

		actualResources, actualError := p.GetResources(pc)

		assert.Equal(t, expectedResources, actualResources)
		assert.Nil(t, actualError)
	})
}
