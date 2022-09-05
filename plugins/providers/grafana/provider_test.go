package grafana_test

import (
	"errors"
	"testing"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/plugins/providers/grafana"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestCreateConfig(t *testing.T) {
	t.Run("should return error if there credentials are invalid", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.GrafanaClient)
		p := grafana.NewProvider("", crypto)
		p.Clients = map[string]grafana.GrafanaClient{
			providerURN: client,
		}

		testcases := []struct {
			name string
			pc   *domain.ProviderConfig
		}{
			{
				name: "invalid credentials struct",
				pc: &domain.ProviderConfig{
					Credentials: "invalid-credential-structure"},
			},
			{
				name: "empty mandatory credentials",
				pc: &domain.ProviderConfig{
					Credentials: grafana.Credentials{
						Host:     "",
						Username: "",
						Password: "",
					},
				},
			},
		}

		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				actualError := p.CreateConfig(tc.pc)
				assert.Error(t, actualError)
			})
		}
	})

	t.Run("should return error if there resource config is invalid", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.GrafanaClient)
		p := grafana.NewProvider("", crypto)
		p.Clients = map[string]grafana.GrafanaClient{
			providerURN: client,
		}

		testcases := []struct {
			pc *domain.ProviderConfig
		}{
			{
				pc: &domain.ProviderConfig{
					Credentials: grafana.Credentials{
						Host:     "localhost",
						Username: "test-username",
						Password: "test-password",
					},
					Resources: []*domain.ResourceConfig{ //resource type wrong requires one of "folder" or "dashboard"
						{
							Type: "invalid resource type",
						},
					},
				},
			},
			{
				pc: &domain.ProviderConfig{
					Credentials: grafana.Credentials{
						Host:     "localhost",
						Username: "test-username",
						Password: "test-password",
					},
					Resources: []*domain.ResourceConfig{
						{
							Type: grafana.ResourceTypeFolder, // Folder resource type
							Roles: []*domain.Role{
								{
									ID:          "viewer",
									Permissions: []interface{}{"wrong permissions"}, // requires "view" or "edit" or "admin" permissions
								},
							},
						},
					},
				},
			},
		}

		for _, tc := range testcases {
			actualError := p.CreateConfig(tc.pc)
			assert.Error(t, actualError)
		}
	})

	t.Run("should not return error if parse and valid of Credentials are correct", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.GrafanaClient)
		p := grafana.NewProvider("", crypto)
		p.Clients = map[string]grafana.GrafanaClient{
			providerURN: client,
		}
		crypto.On("Encrypt", "test-password").Return("encrypted-test-pasword", nil)

		testcases := []struct {
			pc            *domain.ProviderConfig
			expectedError error
		}{
			{
				pc: &domain.ProviderConfig{
					Credentials: grafana.Credentials{
						Host:     "http://localhost",
						Username: "test-username",
						Password: "test-password",
					},
					Resources: []*domain.ResourceConfig{
						{
							Type: grafana.ResourceTypeFolder,
							Roles: []*domain.Role{
								{
									ID:          "viewer",
									Permissions: []interface{}{"view"},
								},
							},
						},
					},
					URN: providerURN,
				},
				expectedError: nil,
			},
			{
				pc: &domain.ProviderConfig{
					Credentials: grafana.Credentials{
						Host:     "http://localhost",
						Username: "test-username",
						Password: "test-password",
					},
					Resources: []*domain.ResourceConfig{
						{
							Type: grafana.ResourceTypeDashboard,
							Roles: []*domain.Role{
								{
									ID:          "admin",
									Permissions: []interface{}{"admin"},
								},
							},
						},
					},
					URN: providerURN,
				},
				expectedError: nil,
			},
		}

		for _, tc := range testcases {
			actualError := p.CreateConfig(tc.pc)
			assert.Equal(t, tc.expectedError, actualError)
			crypto.AssertExpectations(t)
		}
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

func TestGrantAccess(t *testing.T) {
	t.Run("should return error if credentials is invalid", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := grafana.NewProvider("", crypto)

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
		a := domain.Grant{
			Resource: &domain.Resource{
				Type: "test-type",
			},
			Role: "test-role",
		}

		actualError := p.GrantAccess(pc, a)
		assert.Error(t, actualError)
	})

	t.Run("should return error if there are any on client initialization", func(t *testing.T) {
		password := "test-password"
		crypto := new(mocks.Crypto)
		p := grafana.NewProvider("", crypto)
		expectedError := errors.New("decrypt error")
		crypto.On("Decrypt", password).Return("", expectedError).Once()

		pc := &domain.ProviderConfig{
			Credentials: grafana.Credentials{
				Host:     "localhost",
				Username: "test-username",
				Password: password,
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
		a := domain.Grant{
			Resource: &domain.Resource{
				Type: "test-type",
			},
			Role: "test-role",
		}

		actualError := p.GrantAccess(pc, a)

		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return error if resource type in unknown", func(t *testing.T) {
		password := "test-password"
		crypto := new(mocks.Crypto)
		p := grafana.NewProvider("", crypto)
		expectedError := errors.New("invalid resource type")
		crypto.On("Decrypt", password).Return("", expectedError).Once()

		pc := &domain.ProviderConfig{
			Credentials: grafana.Credentials{
				Host:     "localhost",
				Username: "test-username",
				Password: password,
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
		a := domain.Grant{
			Resource: &domain.Resource{
				Type: "test-type",
			},
			Role: "test-role",
		}

		actualError := p.GrantAccess(pc, a)

		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("given dashboard resource", func(t *testing.T) {
		t.Run("should return error if there is an error in granting dashboard access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			crypto := new(mocks.Crypto)
			client := new(mocks.GrafanaClient)
			p := grafana.NewProvider("", crypto)
			p.Clients = map[string]grafana.GrafanaClient{
				providerURN: client,
			}
			client.On("GrantDashboardAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

			pc := &domain.ProviderConfig{
				Credentials: grafana.Credentials{
					Host:     "localhost",
					Username: "test-username",
					Password: "test-password",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: grafana.ResourceTypeDashboard,
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
			a := domain.Grant{
				Resource: &domain.Resource{
					Type: grafana.ResourceTypeDashboard,
					URN:  "999",
					Name: "test-dashboard",
				},
				Role:        "test-role",
				Permissions: []string{"test-permission-config"},
			}

			actualError := p.GrantAccess(pc, a)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return nil error if granting access is successful", func(t *testing.T) {
			providerURN := "test-provider-urn"
			crypto := new(mocks.Crypto)
			client := new(mocks.GrafanaClient)
			expectedDatabase := &grafana.Dashboard{
				Title: "test-dashboard",
				ID:    999,
			}
			expectedUser := "test@email.com"
			expectedRole := grafana.DashboardRoleViewer
			p := grafana.NewProvider("", crypto)
			p.Clients = map[string]grafana.GrafanaClient{
				providerURN: client,
			}
			client.On("GrantDashboardAccess", expectedDatabase, expectedUser, expectedRole).Return(nil).Once()

			pc := &domain.ProviderConfig{
				Credentials: grafana.Credentials{
					Host:     "localhost",
					Username: "test-username",
					Password: "test-password",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: grafana.ResourceTypeDashboard,
						Roles: []*domain.Role{
							{
								ID:          "viewer",
								Permissions: []interface{}{"view"},
							},
						},
					},
				},
				URN: providerURN,
			}
			a := domain.Grant{
				Resource: &domain.Resource{
					Type: grafana.ResourceTypeDashboard,
					URN:  "999",
					Name: "test-dashboard",
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

func TestRevokeAccess(t *testing.T) {
	t.Run("should return error if credentials is invalid", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := grafana.NewProvider("", crypto)

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
		a := domain.Grant{
			Resource: &domain.Resource{
				Type: "test-type",
			},
			Role: "test-role",
		}

		actualError := p.RevokeAccess(pc, a)
		assert.Error(t, actualError)
	})

	t.Run("should return error if there are any on client initialization", func(t *testing.T) {
		password := "test-password"
		crypto := new(mocks.Crypto)
		p := grafana.NewProvider("", crypto)
		expectedError := errors.New("decrypt error")
		crypto.On("Decrypt", password).Return("", expectedError).Once()

		pc := &domain.ProviderConfig{
			Credentials: grafana.Credentials{
				Host:     "localhost",
				Username: "test-username",
				Password: password,
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
		a := domain.Grant{
			Resource: &domain.Resource{
				Type: "test-type",
			},
			Role: "test-role",
		}

		actualError := p.RevokeAccess(pc, a)

		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return error if resource type in unknown", func(t *testing.T) {
		password := "test-password"
		crypto := new(mocks.Crypto)
		p := grafana.NewProvider("", crypto)
		expectedError := errors.New("invalid resource type")
		crypto.On("Decrypt", password).Return("", expectedError).Once()

		pc := &domain.ProviderConfig{
			Credentials: grafana.Credentials{
				Host:     "localhost",
				Username: "test-username",
				Password: password,
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
		a := domain.Grant{
			Resource: &domain.Resource{
				Type: "test-type",
			},
			Role: "test-role",
		}

		actualError := p.RevokeAccess(pc, a)

		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("given dashboard resource", func(t *testing.T) {
		t.Run("should return error if there is an error in revoking dashboard access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			crypto := new(mocks.Crypto)
			client := new(mocks.GrafanaClient)
			p := grafana.NewProvider("", crypto)
			p.Clients = map[string]grafana.GrafanaClient{
				providerURN: client,
			}
			client.On("RevokeDashboardAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

			pc := &domain.ProviderConfig{
				Credentials: grafana.Credentials{
					Host:     "localhost",
					Username: "test-username",
					Password: "test-password",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: grafana.ResourceTypeDashboard,
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
			a := domain.Grant{
				Resource: &domain.Resource{
					Type: grafana.ResourceTypeDashboard,
					URN:  "999",
					Name: "test-dashboard",
				},
				Role:        "test-role",
				Permissions: []string{"test-permission-config"},
			}

			actualError := p.RevokeAccess(pc, a)

			assert.EqualError(t, actualError, expectedError.Error())
		})
	})

	t.Run("should return nil error if revoking access is successful", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.GrafanaClient)
		expectedDatabase := &grafana.Dashboard{
			Title: "test-dashboard",
			ID:    999,
		}
		expectedUser := "test@email.com"
		expectedRole := grafana.DashboardRoleViewer
		p := grafana.NewProvider("", crypto)
		p.Clients = map[string]grafana.GrafanaClient{
			providerURN: client,
		}
		client.On("RevokeDashboardAccess", expectedDatabase, expectedUser, expectedRole).Return(nil).Once()

		pc := &domain.ProviderConfig{
			Credentials: grafana.Credentials{
				Host:     "localhost",
				Username: "test-username",
				Password: "test-password",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: grafana.ResourceTypeDashboard,
					Roles: []*domain.Role{
						{
							ID:          "viewer",
							Permissions: []interface{}{"view"},
						},
					},
				},
			},
			URN: providerURN,
		}
		a := domain.Grant{
			Resource: &domain.Resource{
				Type: grafana.ResourceTypeDashboard,
				URN:  "999",
				Name: "test-dashboard",
			},
			Role:        "viewer",
			AccountID:   expectedUser,
			ResourceID:  "999",
			ID:          "999",
			Permissions: []string{"view"},
		}

		actualError := p.RevokeAccess(pc, a)

		assert.Nil(t, actualError)
		client.AssertExpectations(t)
	})
}

func TestGetRoles(t *testing.T) {
	t.Run("should return an error if invalid resource type", func(t *testing.T) {
		expectedError := grafana.ErrInvalidResourceType
		crypto := new(mocks.Crypto)
		pv := grafana.NewProvider("grafana", crypto)
		validconfig := &domain.ProviderConfig{
			Type: "grafana",
			Credentials: grafana.Credentials{
				Host:     "localhost",
				Username: "test-username",
				Password: "test-password",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: grafana.ResourceTypeDashboard,
				},
			},
			URN: "test-URN",
		}

		actualRoles, actualError := pv.GetRoles(validconfig, "wrong-resource-type")

		assert.Nil(t, actualRoles)
		assert.Equal(t, expectedError, actualError)
	},
	)

	t.Run("should return no error if returned roles match the expected roles", func(t *testing.T) {
		expectedRoles := []*domain.Role{
			{
				ID:          "viewer",
				Permissions: []interface{}{"view"},
			},
		}
		crypto := new(mocks.Crypto)
		pv := grafana.NewProvider("grafana", crypto)
		validconfig := &domain.ProviderConfig{
			Type: "grafana",
			Credentials: grafana.Credentials{
				Host:     "localhost",
				Username: "test-username",
				Password: "test-password",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type:  grafana.ResourceTypeDashboard,
					Roles: expectedRoles,
				},
			},
			URN: "test-URN",
		}

		actualRoles, actualError := pv.GetRoles(validconfig, "dashboard")

		assert.Equal(t, expectedRoles, actualRoles)
		assert.NoError(t, actualError)
	},
	)
}

func TestGetAccountTypes(t *testing.T) {
	t.Run("should return the list of supported account types (user only)", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		pv := grafana.NewProvider("grafana", crypto)
		expectedAccountTypes := []string{"user"}

		actualAccountTypes := pv.GetAccountTypes()

		assert.Equal(t, expectedAccountTypes, actualAccountTypes)
	})
}
