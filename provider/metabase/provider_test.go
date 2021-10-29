package metabase_test

import (
	"errors"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/provider/metabase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
			p := metabase.NewProvider("", crypto)

			providerConfig := &domain.ProviderConfig{
				Resources: tc.resourceConfigs,
			}

			actualError := p.GrantAccess(providerConfig, tc.appeal)
			assert.EqualError(t, actualError, tc.expectedError.Error())
		}
	})

	t.Run("should return error if credentials is invalid", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := metabase.NewProvider("", crypto)

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
		p := metabase.NewProvider("", crypto)
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
		p := metabase.NewProvider("", crypto)
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
			p := metabase.NewProvider("", crypto)
			p.Clients = map[string]metabase.MetabaseClient{
				providerURN: client,
			}
			client.On("GrantDatabaseAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

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
					URN:  "999",
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
			client := new(mocks.MetabaseClient)
			expectedDatabase := &metabase.Database{
				Name: "test-database",
				ID:   999,
			}
			expectedUser := "test@email.com"
			expectedRole := metabase.DatabaseRoleViewer
			p := metabase.NewProvider("", crypto)
			p.Clients = map[string]metabase.MetabaseClient{
				providerURN: client,
			}
			client.On("GrantDatabaseAccess", expectedDatabase, expectedUser, expectedRole).Return(nil).Once()

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
					URN:  "999",
					Name: "test-database",
				},
				Role:       "viewer",
				AccountID:  expectedUser,
				ResourceID: 999,
				ID:         999,
			}

			actualError := p.GrantAccess(pc, a)

			assert.Nil(t, actualError)
		})
	})

	t.Run("given collection resource", func(t *testing.T) {
		t.Run("should return error if there is an error in grandting collection access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			crypto := new(mocks.Crypto)
			client := new(mocks.MetabaseClient)
			p := metabase.NewProvider("", crypto)
			p.Clients = map[string]metabase.MetabaseClient{
				providerURN: client,
			}
			client.On("GrantCollectionAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

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
					URN:  "999",
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
			p := metabase.NewProvider("", crypto)

			p.Clients = map[string]metabase.MetabaseClient{
				providerURN: client,
			}
			client.On("GrantCollectionAccess", expectedCollection, expectedUser, expectedRole).Return(nil).Once()

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
					URN:  "999",
					Name: "test-collection",
				},
				Role:       "viewer",
				AccountID:  expectedUser,
				ResourceID: 999,
				ID:         999,
			}

			actualError := p.GrantAccess(pc, a)

			assert.Nil(t, actualError)
		})
	})
}
