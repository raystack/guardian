package metabase_test

import (
	"errors"
	"testing"

	"github.com/odpf/salt/log"

	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/plugins/providers/metabase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetType(t *testing.T) {
	t.Run("should return provider type name", func(t *testing.T) {
		expectedTypeName := domain.ProviderTypeMetabase
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		crypto := new(mocks.Crypto)
		p := metabase.NewProvider(expectedTypeName, crypto, logger)

		actualTypeName := p.GetType()

		assert.Equal(t, expectedTypeName, actualTypeName)
	})
}

func TestGetResources(t *testing.T) {
	t.Run("should return error if credentials is invalid", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		p := metabase.NewProvider("", crypto, logger)

		pc := &domain.ProviderConfig{
			Credentials: "invalid-creds",
		}

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.Error(t, actualError)
	})

	t.Run("should return error if there are any on client initialization", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		p := metabase.NewProvider("", crypto, logger)

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
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		p := metabase.NewProvider("", crypto, logger)
		p.Clients = map[string]metabase.MetabaseClient{
			providerURN: client,
		}

		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: map[string]interface{}{},
			Resources: []*domain.ResourceConfig{
				{
					Type: metabase.ResourceTypeDatabase,
				},
			},
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
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		p := metabase.NewProvider("", crypto, logger)
		p.Clients = map[string]metabase.MetabaseClient{
			providerURN: client,
		}

		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: map[string]interface{}{},
			Resources: []*domain.ResourceConfig{
				{
					Type: metabase.ResourceTypeCollection,
				},
			},
		}
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
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		p := metabase.NewProvider("", crypto, logger)
		p.Clients = map[string]metabase.MetabaseClient{
			providerURN: client,
		}

		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: map[string]interface{}{},
			Resources: []*domain.ResourceConfig{
				{
					Type: metabase.ResourceTypeDatabase,
				}, {
					Type: metabase.ResourceTypeTable,
				},
				{
					Type: metabase.ResourceTypeCollection,
				}, {
					Type: metabase.ResourceTypeGroup,
				},
			},
		}
		expectedDatabases := []*metabase.Database{
			{
				ID:     1,
				Name:   "db_1",
				Tables: []metabase.Table{{ID: 2, Name: "table_1", DbId: 1}},
			},
		}
		client.On("GetDatabases").Return(expectedDatabases, nil).Once()

		d := []*metabase.GroupResource{{Urn: "database:1", Permissions: []string{"read", "write"}}}
		c := []*metabase.GroupResource{{Urn: "collection:1", Permissions: []string{"read", "write"}}}
		group := metabase.Group{Name: "All Users", DatabaseResources: d, CollectionResources: c}

		client.On("GetGroups").Return([]*metabase.Group{&group},
			metabase.ResourceGroupDetails{"database:1": {{"urn": "group:1", "permissions": []string{"read", "write"}}}},
			metabase.ResourceGroupDetails{"collection:1": {{"urn": "group:1", "permissions": []string{"write"}}}}, nil).Once()

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
				URN:         "database:1",
				ProviderURN: providerURN,
				Name:        "db_1",
				Details: map[string]interface{}{
					"auto_run_queries":            false,
					"cache_field_values_schedule": "",
					"engine":                      "",
					"metadata_sync_schedule":      "",
					"native_permissions":          "",
					"timezone":                    "",
					"groups":                      []map[string]interface{}{{"urn": "group:1", "permissions": []string{"read", "write"}}},
				},
			}, {
				Type:        metabase.ResourceTypeTable,
				URN:         "table:1.2",
				ProviderURN: providerURN,
				Name:        "table_1",
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
				URN:         "collection:1",
				ProviderURN: providerURN,
				Name:        "col_1",
				Details: map[string]interface{}{
					"groups": []map[string]interface{}{{"urn": "group:1", "permissions": []string{"write"}}},
				},
			},
			{
				Type:        metabase.ResourceTypeGroup,
				URN:         "group:0",
				ProviderURN: providerURN,
				Name:        "All Users",
				Details: map[string]interface{}{
					"collection": []*metabase.GroupResource{{Name: "col_1", Type: "collection", Urn: "collection:1", Permissions: []string{"read", "write"}}},
					"database":   []*metabase.GroupResource{{Name: "db_1", Type: "database", Urn: "database:1", Permissions: []string{"read", "write"}}},
				},
			},
		}

		actualResources, actualError := p.GetResources(pc)

		assert.Equal(t, expectedResources, actualResources)
		assert.Nil(t, actualError)
	})
}

func TestGrantAccess(t *testing.T) {
	t.Run("should return an error if there is an error in getting permissions", func(t *testing.T) {
		var permission metabase.Permission
		invalidPermissionConfig := map[string]interface{}{}
		invalidPermissionConfigError := mapstructure.Decode(invalidPermissionConfig, &permission)

		testcases := []struct {
			resourceConfigs []*domain.ResourceConfig
			appeal          *domain.Appeal
			expectedError   error
		}{
			{
				appeal: &domain.Appeal{
					Resource: &domain.Resource{
						Type: "test-type",
					},
				},
				expectedError: metabase.ErrInvalidResourceType,
			},
			{
				resourceConfigs: []*domain.ResourceConfig{
					{
						Type: "test-type",
						Roles: []*domain.Role{
							{
								ID: "not-test-role",
							},
						},
					},
				},
				appeal: &domain.Appeal{
					Resource: &domain.Resource{
						Type: "test-type",
					},
					Role: "test-role",
				},
				expectedError: metabase.ErrInvalidRole,
			},
			{
				resourceConfigs: []*domain.ResourceConfig{
					{
						Type: "test-type",
						Roles: []*domain.Role{
							{
								ID: "test-role",
								Permissions: []interface{}{
									invalidPermissionConfig,
								},
							},
						},
					},
				},
				appeal: &domain.Appeal{
					Resource: &domain.Resource{
						Type: "test-type",
					},
					Role: "test-role",
				},
				expectedError: invalidPermissionConfigError,
			},
		}

		for _, tc := range testcases {
			crypto := new(mocks.Crypto)
			logger := log.NewLogrus(log.LogrusWithLevel("info"))
			p := metabase.NewProvider("", crypto, logger)

			providerConfig := &domain.ProviderConfig{
				Resources: tc.resourceConfigs,
			}

			actualError := p.GrantAccess(providerConfig, tc.appeal)
			assert.EqualError(t, actualError, tc.expectedError.Error())
		}
	})

	t.Run("should return error if credentials is invalid", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		p := metabase.NewProvider("", crypto, logger)

		pc := &domain.ProviderConfig{
			Credentials: "invalid-credentials",
			Resources: []*domain.ResourceConfig{
				{
					Type: "test-type",
					Roles: []*domain.Role{
						{
							ID:          "test-role",
							Permissions: []interface{}{"test-permission-config"},
						},
					},
				},
			},
		}
		a := &domain.Appeal{
			Resource: &domain.Resource{
				Type: "test-type",
			},
			Role: "test-role",
		}

		actualError := p.GrantAccess(pc, a)
		assert.Error(t, actualError)
	})

	t.Run("should return error if there are any on client initialization", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		p := metabase.NewProvider("", crypto, logger)
		expectedError := errors.New("decrypt error")
		crypto.On("Decrypt", "test-password").Return("", expectedError).Once()

		pc := &domain.ProviderConfig{
			Credentials: metabase.Credentials{
				Host:     "localhost",
				Username: "test-username",
				Password: "test-password",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: "test-type",
					Roles: []*domain.Role{
						{
							ID:          "test-role",
							Permissions: []interface{}{"test-permission-config"},
						},
					},
				},
			},
		}
		a := &domain.Appeal{
			Resource: &domain.Resource{
				Type: "test-type",
			},
			Role: "test-role",
		}

		actualError := p.GrantAccess(pc, a)

		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return error if resource type in unknown", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		p := metabase.NewProvider("", crypto, logger)
		expectedError := errors.New("invalid resource type")
		crypto.On("Decrypt", "test-password").Return("", expectedError).Once()

		pc := &domain.ProviderConfig{
			Credentials: metabase.Credentials{
				Host:     "localhost",
				Username: "test-username",
				Password: "test-password",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: "test-type",
					Roles: []*domain.Role{
						{
							ID:          "test-role",
							Permissions: []interface{}{"test-permission-config"},
						},
					},
				},
			},
			URN: "test-urn",
		}
		a := &domain.Appeal{
			Resource: &domain.Resource{
				Type: "test-type",
			},
			Role: "test-role",
		}

		actualError := p.GrantAccess(pc, a)

		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("given database resource", func(t *testing.T) {
		t.Run("should return error if there is an error in granting database access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			crypto := new(mocks.Crypto)
			client := new(mocks.MetabaseClient)
			logger := log.NewLogrus(log.LogrusWithLevel("info"))
			p := metabase.NewProvider("", crypto, logger)
			p.Clients = map[string]metabase.MetabaseClient{
				providerURN: client,
			}
			client.On("GrantDatabaseAccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()
			d := []*metabase.GroupResource{{Urn: "database:1", Permissions: []string{"read", "write"}}}
			c := []*metabase.GroupResource{{Urn: "collection:1", Permissions: []string{"read", "write"}}}
			group := metabase.Group{Name: "All Users", DatabaseResources: d, CollectionResources: c}
			client.On("GetGroups").Return([]*metabase.Group{&group},
				metabase.ResourceGroupDetails{"database:1": {{"urn": "group:1", "permissions": []string{"read", "write"}}}},
				metabase.ResourceGroupDetails{"collection:1": {{"urn": "group:1", "permissions": []string{"write"}}}}, nil).Once()

			pc := &domain.ProviderConfig{
				Credentials: metabase.Credentials{
					Host:     "localhost",
					Username: "test-username",
					Password: "test-password",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: metabase.ResourceTypeDatabase,
						Roles: []*domain.Role{
							{
								ID:          "test-role",
								Permissions: []interface{}{"test-permission-config"},
							},
						},
					},
				},
				URN: providerURN,
			}
			a := &domain.Appeal{
				Resource: &domain.Resource{
					Type: metabase.ResourceTypeDatabase,
					URN:  "database:999",
					Name: "test-database",
				},
				Role: "test-role",
			}

			actualError := p.GrantAccess(pc, a)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return nil error if granting access is successful", func(t *testing.T) {
			providerURN := "test-provider-urn"
			crypto := new(mocks.Crypto)
			logger := log.NewLogrus(log.LogrusWithLevel("info"))
			client := new(mocks.MetabaseClient)
			expectedDatabase := &metabase.Database{
				Name: "test-database",
				ID:   999,
			}
			expectedUser := "test@email.com"
			expectedRole := metabase.DatabaseRoleViewer
			p := metabase.NewProvider("", crypto, logger)
			p.Clients = map[string]metabase.MetabaseClient{
				providerURN: client,
			}
			client.On("GrantDatabaseAccess", expectedDatabase, expectedUser, expectedRole, mock.Anything).Return(nil).Once()

			d := []*metabase.GroupResource{{Urn: "database:1", Permissions: []string{"read", "write"}}}
			c := []*metabase.GroupResource{{Urn: "collection:1", Permissions: []string{"read", "write"}}}
			group := metabase.Group{Name: "All Users", DatabaseResources: d, CollectionResources: c}
			client.On("GetGroups").Return([]*metabase.Group{&group},
				metabase.ResourceGroupDetails{"database:1": {{"urn": "group:1", "permissions": []string{"read", "write"}}}},
				metabase.ResourceGroupDetails{"collection:1": {{"urn": "group:1", "permissions": []string{"write"}}}}, nil).Once()

			pc := &domain.ProviderConfig{
				Credentials: metabase.Credentials{
					Host:     "localhost",
					Username: "test-username",
					Password: "test-password",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: metabase.ResourceTypeDatabase,
						Roles: []*domain.Role{
							{
								ID:          "viewer",
								Permissions: []interface{}{expectedRole},
							},
						},
					},
				},
				URN: providerURN,
			}
			a := &domain.Appeal{
				Resource: &domain.Resource{
					Type: metabase.ResourceTypeDatabase,
					URN:  "database:999",
					Name: "test-database",
				},
				Role:       "viewer",
				AccountID:  expectedUser,
				ResourceID: "999",
				ID:         "999",
			}

			actualError := p.GrantAccess(pc, a)

			assert.Nil(t, actualError)
		})
	})

	t.Run("given collection resource", func(t *testing.T) {
		t.Run("should return error if there is an error in granting collection access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			crypto := new(mocks.Crypto)
			client := new(mocks.MetabaseClient)
			logger := log.NewLogrus(log.LogrusWithLevel("info"))
			p := metabase.NewProvider("", crypto, logger)
			p.Clients = map[string]metabase.MetabaseClient{
				providerURN: client,
			}
			client.On("GrantCollectionAccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

			d := []*metabase.GroupResource{{Urn: "database:1", Permissions: []string{"read", "write"}}}
			c := []*metabase.GroupResource{{Urn: "collection:1", Permissions: []string{"read", "write"}}}
			group := metabase.Group{Name: "All Users", DatabaseResources: d, CollectionResources: c}
			client.On("GetGroups").Return([]*metabase.Group{&group},
				metabase.ResourceGroupDetails{"database:1": {{"urn": "group:1", "permissions": []string{"read", "write"}}}},
				metabase.ResourceGroupDetails{"collection:1": {{"urn": "group:1", "permissions": []string{"write"}}}}, nil).Once()

			pc := &domain.ProviderConfig{
				Credentials: metabase.Credentials{
					Host:     "localhost",
					Username: "test-username",
					Password: "test-password",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: metabase.ResourceTypeCollection,
						Roles: []*domain.Role{
							{
								ID:          "test-role",
								Permissions: []interface{}{"test-permission-config"},
							},
						},
					},
				},
				URN: providerURN,
			}
			a := &domain.Appeal{
				Resource: &domain.Resource{
					Type: metabase.ResourceTypeCollection,
					URN:  "collection:999",
					Name: "test-collection",
				},
				Role: "test-role",
			}

			actualError := p.GrantAccess(pc, a)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return nil error if granting access is successful", func(t *testing.T) {
			providerURN := "test-provider-urn"
			crypto := new(mocks.Crypto)
			client := new(mocks.MetabaseClient)
			expectedCollection := &metabase.Collection{
				Name: "test-collection",
				ID:   999,
			}
			expectedUser := "test@email.com"
			expectedRole := "viewer"
			logger := log.NewLogrus(log.LogrusWithLevel("info"))
			p := metabase.NewProvider("", crypto, logger)

			p.Clients = map[string]metabase.MetabaseClient{
				providerURN: client,
			}
			client.On("GrantCollectionAccess", expectedCollection, expectedUser, expectedRole, mock.Anything).Return(nil).Once()

			d := []*metabase.GroupResource{{Urn: "database:1", Permissions: []string{"read", "write"}}}
			c := []*metabase.GroupResource{{Urn: "collection:1", Permissions: []string{"read", "write"}}}
			group := metabase.Group{Name: "All Users", DatabaseResources: d, CollectionResources: c}
			client.On("GetGroups").Return([]*metabase.Group{&group},
				metabase.ResourceGroupDetails{"database:1": {{"urn": "group:1", "permissions": []string{"read", "write"}}}},
				metabase.ResourceGroupDetails{"collection:1": {{"urn": "group:1", "permissions": []string{"write"}}}}, nil).Once()

			pc := &domain.ProviderConfig{
				Credentials: metabase.Credentials{
					Host:     "localhost",
					Username: "test-username",
					Password: "test-password",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: metabase.ResourceTypeCollection,
						Roles: []*domain.Role{
							{
								ID:          "viewer",
								Permissions: []interface{}{expectedRole},
							},
						},
					},
				},
				URN: providerURN,
			}
			a := &domain.Appeal{
				Resource: &domain.Resource{
					Type: metabase.ResourceTypeCollection,
					URN:  "collection:999",
					Name: "test-collection",
				},
				Role:       "viewer",
				AccountID:  expectedUser,
				ResourceID: "999",
				ID:         "999",
			}

			actualError := p.GrantAccess(pc, a)

			assert.Nil(t, actualError)
		})
	})
}
