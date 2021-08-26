package gcloudiam_test

import (
	"errors"
	"testing"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/provider/gcloudiam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetType(t *testing.T) {
	t.Run("should return provider type name", func(t *testing.T) {
		expectedTypeName := domain.ProviderTypeMetabase
		crypto := new(mocks.Crypto)
		p := gcloudiam.NewProvider(expectedTypeName, crypto)

		actualTypeName := p.GetType()

		assert.Equal(t, expectedTypeName, actualTypeName)
	})
}

func TestGetResources(t *testing.T) {
	t.Run("should return error if credentials is invalid", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := gcloudiam.NewProvider("", crypto)

		pc := &domain.ProviderConfig{
			Credentials: "invalid-creds",
		}

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.Error(t, actualError)
	})

	t.Run("should return error if got any on getting role resources", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.GcloudIamClient)
		p := gcloudiam.NewProvider("", crypto)
		p.Clients = map[string]gcloudiam.GcloudIamClient{
			providerURN: client,
		}

		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: map[string]interface{}{},
		}
		expectedError := errors.New("client error")
		client.On("GetRoles", mock.Anything).Return(nil, expectedError).Once()

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return list of resources and nil error on success", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.GcloudIamClient)
		p := gcloudiam.NewProvider("", crypto)
		p.Clients = map[string]gcloudiam.GcloudIamClient{
			providerURN: client,
		}

		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: map[string]interface{}{},
		}
		expectedDatabases := []*gcloudiam.Role{
			{
				Name:        "role",
				Title:       "Role Title",
				Description: "test description",
			},
		}
		client.On("GetRoles", mock.Anything).Return(expectedDatabases, nil).Once()
		expectedResources := []*domain.Resource{
			{
				Type:        gcloudiam.ResourceTypeRole,
				URN:         "role",
				ProviderURN: providerURN,
				Name:        "Role Title",
				Details: map[string]interface{}{
					"description": "test description",
				},
			},
		}

		actualResources, actualError := p.GetResources(pc)

		assert.Equal(t, expectedResources, actualResources)
		assert.Nil(t, actualError)
	})
}

func TestGrantAccess(t *testing.T) {
	t.Run("should return error if credentials is invalid", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := gcloudiam.NewProvider("", crypto)

		pc := &domain.ProviderConfig{
			Credentials: "invalid-credentials",
			Resources: []*domain.ResourceConfig{
				{
					Type: "test-type",
					Roles: []*domain.RoleConfig{
						{
							ID: "test-role",
							Permissions: []interface{}{
								gcloudiam.PermissionConfig{
									Name: "test-permission-config",
								},
							},
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

	t.Run("should return error if resource type is unknown", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.GcloudIamClient)
		p := gcloudiam.NewProvider("", crypto)
		p.Clients = map[string]gcloudiam.GcloudIamClient{
			providerURN: client,
		}
		expectedError := errors.New("invalid resource type")

		pc := &domain.ProviderConfig{
			URN: providerURN,
			Resources: []*domain.ResourceConfig{
				{
					Type: "test-type",
					Roles: []*domain.RoleConfig{
						{
							ID: "test-role",
							Permissions: []interface{}{
								gcloudiam.PermissionConfig{
									Name: "test-permission-config",
								},
							},
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

	t.Run("given database resource", func(t *testing.T) {
		t.Run("should return error if there is an error in granting the access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			crypto := new(mocks.Crypto)
			client := new(mocks.GcloudIamClient)
			p := gcloudiam.NewProvider("", crypto)
			p.Clients = map[string]gcloudiam.GcloudIamClient{
				providerURN: client,
			}
			client.On("GrantAccess", mock.Anything, mock.Anything).Return(expectedError).Once()

			pc := &domain.ProviderConfig{
				Resources: []*domain.ResourceConfig{
					{
						Type: gcloudiam.ResourceTypeRole,
						Roles: []*domain.RoleConfig{
							{
								ID: "test-role",
								Permissions: []interface{}{
									gcloudiam.PermissionConfig{
										Name: "test-permission-config",
									},
								},
							},
						},
					},
				},
				URN: providerURN,
			}
			a := &domain.Appeal{
				Resource: &domain.Resource{
					Type: gcloudiam.ResourceTypeRole,
					URN:  "999",
					Name: "test-role",
				},
				Role: "test-role",
			}

			actualError := p.GrantAccess(pc, a)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return nil error if granting access is successful", func(t *testing.T) {
			providerURN := "test-provider-urn"
			crypto := new(mocks.Crypto)
			client := new(mocks.GcloudIamClient)
			expectedResource := &gcloudiam.Role{
				Name: "test-role",
			}
			expectedUser := "test@email.com"
			p := gcloudiam.NewProvider("", crypto)
			p.Clients = map[string]gcloudiam.GcloudIamClient{
				providerURN: client,
			}
			client.On("GrantAccess", expectedResource, expectedUser).Return(nil).Once()

			pc := &domain.ProviderConfig{
				Resources: []*domain.ResourceConfig{
					{
						Type: gcloudiam.ResourceTypeRole,
						Roles: []*domain.RoleConfig{
							{
								ID: "allow",
								Permissions: []interface{}{
									gcloudiam.PermissionConfig{
										Name: "allow",
									},
								},
							},
						},
					},
				},
				URN: providerURN,
			}
			a := &domain.Appeal{
				Resource: &domain.Resource{
					Type: gcloudiam.ResourceTypeRole,
					URN:  "test-role",
				},
				Role:       "viewer",
				User:       expectedUser,
				ResourceID: 999,
				ID:         999,
			}

			actualError := p.GrantAccess(pc, a)

			assert.Nil(t, actualError)
		})
	})
}

func TestRevokeAccess(t *testing.T) {
	t.Run("should return error if resource type is unknown", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.GcloudIamClient)
		p := gcloudiam.NewProvider("", crypto)
		p.Clients = map[string]gcloudiam.GcloudIamClient{
			providerURN: client,
		}
		expectedError := errors.New("invalid resource type")

		pc := &domain.ProviderConfig{
			URN: providerURN,
			Resources: []*domain.ResourceConfig{
				{
					Type: "test-type",
					Roles: []*domain.RoleConfig{
						{
							ID: "test-role",
							Permissions: []interface{}{
								gcloudiam.PermissionConfig{
									Name: "test-permission-config",
								},
							},
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

		actualError := p.RevokeAccess(pc, a)

		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("given database resource", func(t *testing.T) {
		t.Run("should return error if there is an error in granting the access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			crypto := new(mocks.Crypto)
			client := new(mocks.GcloudIamClient)
			p := gcloudiam.NewProvider("", crypto)
			p.Clients = map[string]gcloudiam.GcloudIamClient{
				providerURN: client,
			}
			client.On("RevokeAccess", mock.Anything, mock.Anything).Return(expectedError).Once()

			pc := &domain.ProviderConfig{
				Resources: []*domain.ResourceConfig{
					{
						Type: gcloudiam.ResourceTypeRole,
						Roles: []*domain.RoleConfig{
							{
								ID: "test-role",
								Permissions: []interface{}{
									gcloudiam.PermissionConfig{
										Name: "test-permission-config",
									},
								},
							},
						},
					},
				},
				URN: providerURN,
			}
			a := &domain.Appeal{
				Resource: &domain.Resource{
					Type: gcloudiam.ResourceTypeRole,
					URN:  "999",
					Name: "test-role",
				},
				Role: "test-role",
			}

			actualError := p.RevokeAccess(pc, a)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return nil error if revoking access is successful", func(t *testing.T) {
			providerURN := "test-provider-urn"
			crypto := new(mocks.Crypto)
			client := new(mocks.GcloudIamClient)
			expectedResource := &gcloudiam.Role{
				Name: "test-role",
			}
			expectedUser := "test@email.com"
			p := gcloudiam.NewProvider("", crypto)
			p.Clients = map[string]gcloudiam.GcloudIamClient{
				providerURN: client,
			}
			client.On("RevokeAccess", expectedResource, expectedUser).Return(nil).Once()

			pc := &domain.ProviderConfig{
				Resources: []*domain.ResourceConfig{
					{
						Type: gcloudiam.ResourceTypeRole,
						Roles: []*domain.RoleConfig{
							{
								ID: "allow",
								Permissions: []interface{}{
									gcloudiam.PermissionConfig{
										Name: "allow",
									},
								},
							},
						},
					},
				},
				URN: providerURN,
			}
			a := &domain.Appeal{
				Resource: &domain.Resource{
					Type: gcloudiam.ResourceTypeRole,
					URN:  "test-role",
				},
				Role:       "viewer",
				User:       expectedUser,
				ResourceID: 999,
				ID:         999,
			}

			actualError := p.RevokeAccess(pc, a)

			assert.Nil(t, actualError)
		})
	})
}
