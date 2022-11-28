package shield_test

import (
	"errors"
	"testing"

	"github.com/odpf/salt/log"

	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/plugins/providers/shield"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetType(t *testing.T) {
	t.Run("should return provider type name", func(t *testing.T) {
		expectedTypeName := domain.ProviderTypeShield
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		p := shield.NewProvider(expectedTypeName, logger)

		actualTypeName := p.GetType()

		assert.Equal(t, expectedTypeName, actualTypeName)
	})
}

func TestCreateConfig(t *testing.T) {
	t.Run("should return error if there resource config is invalid", func(t *testing.T) {
		providerURN := "test-provider-urn"
		client := new(mocks.ShieldClient)
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		p := shield.NewProvider("", logger)
		p.Clients = map[string]shield.ShieldClient{
			providerURN: client,
		}

		testcases := []struct {
			pc *domain.ProviderConfig
		}{
			{
				pc: &domain.ProviderConfig{
					Credentials: shield.Credentials{
						Host:      "localhost",
						AuthEmail: "test-email",
					},
					Resources: []*domain.ResourceConfig{
						{
							Type: "invalid resource type",
						},
					},
				},
			},
			{
				pc: &domain.ProviderConfig{
					Credentials: shield.Credentials{
						Host:      "localhost",
						AuthEmail: "test-email",
					},
					Resources: []*domain.ResourceConfig{
						{
							Type: shield.ResourceTypeTeam,
							Roles: []*domain.Role{
								{
									ID:          "member",
									Permissions: []interface{}{"wrong permissions"},
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
		client := new(mocks.ShieldClient)
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		p := shield.NewProvider("", logger)
		p.Clients = map[string]shield.ShieldClient{
			providerURN: client,
		}

		testcases := []struct {
			pc            *domain.ProviderConfig
			expectedError error
		}{
			{
				pc: &domain.ProviderConfig{
					Credentials: shield.Credentials{
						Host:       "localhost",
						AuthEmail:  "test-email",
						AuthHeader: "X-Auth-Email",
					},
					Resources: []*domain.ResourceConfig{
						{
							Type: shield.ResourceTypeTeam,
							Roles: []*domain.Role{
								{
									ID:          "member",
									Permissions: []interface{}{"users"},
								},
								{
									ID:          "admin",
									Permissions: []interface{}{"admins"},
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
					Credentials: shield.Credentials{
						Host:       "localhost",
						AuthEmail:  "test-email",
						AuthHeader: "X-Auth-Email",
					},
					Resources: []*domain.ResourceConfig{
						{
							Type: shield.ResourceTypeProject,
							Roles: []*domain.Role{
								{
									ID:          "admin",
									Permissions: []interface{}{"admins"},
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
					Credentials: shield.Credentials{
						Host:       "localhost",
						AuthEmail:  "test-email",
						AuthHeader: "X-Auth-Email",
					},
					Resources: []*domain.ResourceConfig{
						{
							Type: shield.ResourceTypeOrganization,
							Roles: []*domain.Role{
								{
									ID:          "admin",
									Permissions: []interface{}{"admins"},
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
		}
	})
}

func TestGetResources(t *testing.T) {
	t.Run("should return error if credentials is invalid", func(t *testing.T) {
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		p := shield.NewProvider("", logger)

		pc := &domain.ProviderConfig{
			Credentials: "invalid-creds",
		}

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.Error(t, actualError)
	})

	t.Run("should return error if got any on getting team resources", func(t *testing.T) {
		providerURN := "test-provider-urn"
		client := new(mocks.ShieldClient)
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		p := shield.NewProvider("", logger)
		p.Clients = map[string]shield.ShieldClient{
			providerURN: client,
		}

		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: map[string]interface{}{},
			Resources: []*domain.ResourceConfig{
				{
					Type: shield.ResourceTypeTeam,
				},
			},
		}
		expectedError := errors.New("client error")
		client.On("GetTeams").Return(nil, expectedError).Once()

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return error if got any on getting project resources", func(t *testing.T) {
		providerURN := "test-provider-urn"
		client := new(mocks.ShieldClient)
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		p := shield.NewProvider("", logger)
		p.Clients = map[string]shield.ShieldClient{
			providerURN: client,
		}

		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: map[string]interface{}{},
			Resources: []*domain.ResourceConfig{
				{
					Type: shield.ResourceTypeProject,
				},
			},
		}
		expectedError := errors.New("client error")
		client.On("GetProjects").Return(nil, expectedError).Once()

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return error if got any on getting organization resources", func(t *testing.T) {
		providerURN := "test-provider-urn"
		client := new(mocks.ShieldClient)
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		p := shield.NewProvider("", logger)
		p.Clients = map[string]shield.ShieldClient{
			providerURN: client,
		}

		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: map[string]interface{}{},
			Resources: []*domain.ResourceConfig{
				{
					Type: shield.ResourceTypeOrganization,
				},
			},
		}
		expectedError := errors.New("client error")
		client.On("GetOrganizations").Return(nil, expectedError).Once()

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return list of resources and nil error on success", func(t *testing.T) {
		providerURN := "test-provider-urn"
		client := new(mocks.ShieldClient)
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		p := shield.NewProvider("", logger)
		p.Clients = map[string]shield.ShieldClient{
			providerURN: client,
		}

		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: map[string]interface{}{},
			Resources: []*domain.ResourceConfig{
				{
					Type: shield.ResourceTypeTeam,
				},
				{
					Type: shield.ResourceTypeProject,
				},
				{
					Type: shield.ResourceTypeOrganization,
				},
			},
		}
		expectedTeams := []*shield.Team{
			{
				ID:    "team_id",
				Name:  "team_1",
				OrgId: "org_id",
				Metadata: shield.Metadata{
					Email:   "team_1@gojek.com",
					Privacy: "public",
					Slack:   "team_1_slack",
				},
				Admins: []string{"testTeamAdmin@gmail.com"},
			},
		}
		client.On("GetTeams").Return(expectedTeams, nil).Once()

		expectedProjects := []*shield.Project{
			{
				ID:     "project_id",
				Name:   "project_1",
				OrgId:  "org_id",
				Admins: []string{"testProjectAdmin@gmail.com"},
			},
		}
		client.On("GetProjects").Return(expectedProjects, nil).Once()

		expectedOrganizations := []*shield.Organization{
			{
				ID:     "org_id",
				Name:   "org_1",
				Admins: []string{"testOrganizationAdmin@gmail.com"},
			},
		}

		client.On("GetOrganizations").Return(expectedOrganizations, nil).Once()

		expectedResources := []*domain.Resource{
			{
				Type:        shield.ResourceTypeTeam,
				URN:         "team:team_id",
				ProviderURN: providerURN,
				Name:        "team_1",
				Details: map[string]interface{}{
					"id":     "team_id",
					"orgId":  "org_id",
					"admins": []string{"testTeamAdmin@gmail.com"},
					"metadata": shield.Metadata{
						Email:   "team_1@gojek.com",
						Privacy: "public",
						Slack:   "team_1_slack",
					},
				},
			}, {
				Type:        shield.ResourceTypeProject,
				URN:         "project:project_id",
				ProviderURN: providerURN,
				Name:        "project_1",
				Details: map[string]interface{}{
					"id":     "project_id",
					"orgId":  "org_id",
					"admins": []string{"testProjectAdmin@gmail.com"},
				},
			},
			{
				Type:        shield.ResourceTypeOrganization,
				URN:         "organization:org_id",
				ProviderURN: providerURN,
				Name:        "org_1",
				Details: map[string]interface{}{
					"id":     "org_id",
					"admins": []string{"testOrganizationAdmin@gmail.com"},
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
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		p := shield.NewProvider("", logger)

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

	t.Run("should return error if resource type in unknown", func(t *testing.T) {
		providerURN := "test-provider-urn"
		expectedError := errors.New("invalid resource type")
		client := new(mocks.ShieldClient)
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		p := shield.NewProvider("", logger)

		p.Clients = map[string]shield.ShieldClient{
			providerURN: client,
		}

		expectedUserEmail := "test@email.com"
		expectedUser := &shield.User{
			ID:    "test_user_id",
			Name:  "test_user",
			Email: expectedUserEmail,
		}

		client.On("GetSelfUser", expectedUserEmail).Return(expectedUser, nil).Once()

		pc := &domain.ProviderConfig{
			Credentials: shield.Credentials{
				Host:      "http://localhost/",
				AuthEmail: "test-email",
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
			URN: providerURN,
		}

		a := domain.Grant{
			Resource: &domain.Resource{
				Type: "test-type",
			},
			Role:      "test-role",
			AccountID: expectedUserEmail,
		}

		actualError := p.GrantAccess(pc, a)

		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("given team resource", func(t *testing.T) {
		t.Run("should return error if there is an error in granting team access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			client := new(mocks.ShieldClient)
			logger := log.NewLogrus(log.LogrusWithLevel("info"))
			p := shield.NewProvider("", logger)
			p.Clients = map[string]shield.ShieldClient{
				providerURN: client,
			}

			expectedUserEmail := "test@email.com"
			expectedUser := &shield.User{
				ID:    "test_user_id",
				Name:  "test_user",
				Email: expectedUserEmail,
			}

			client.On("GetSelfUser", expectedUserEmail).Return(expectedUser, nil).Once()
			client.On("GrantTeamAccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

			pc := &domain.ProviderConfig{
				Credentials: shield.Credentials{
					Host:      "localhost",
					AuthEmail: "test_email",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: shield.ResourceTypeTeam,
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
					Type: shield.ResourceTypeTeam,
					URN:  "team:team_id",
					Name: "team_1",
					Details: map[string]interface{}{
						"id":     "team_id",
						"orgId":  "456",
						"admins": []interface{}{"testAdmin@email.com"},
						"metadata": shield.Metadata{
							Email:   "team_1@gojek.com",
							Privacy: "public",
							Slack:   "team_1_slack",
						},
					},
				},
				Role:        "test-role",
				AccountID:   expectedUserEmail,
				Permissions: []string{"test-permission-config"},
			}

			actualError := p.GrantAccess(pc, a)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return nil error if granting access is successful", func(t *testing.T) {
			providerURN := "test-provider-urn"
			logger := log.NewLogrus(log.LogrusWithLevel("info"))
			client := new(mocks.ShieldClient)
			expectedTeam := &shield.Team{
				Name: "team_1",
				ID:   "team_id",
			}

			expectedUserEmail := "test@email.com"
			expectedUser := &shield.User{
				ID:    "test_user_id",
				Name:  "test_user",
				Email: expectedUserEmail,
			}

			expectedRole := "users"
			p := shield.NewProvider("", logger)
			p.Clients = map[string]shield.ShieldClient{
				providerURN: client,
			}
			client.On("GetSelfUser", expectedUserEmail).Return(expectedUser, nil).Once()
			client.On("GrantTeamAccess", expectedTeam, expectedUser.ID, expectedRole).Return(nil).Once()

			pc := &domain.ProviderConfig{
				Credentials: shield.Credentials{
					Host:      "localhost",
					AuthEmail: "test_email",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: shield.ResourceTypeTeam,
						Roles: []*domain.Role{
							{
								ID:          "member",
								Permissions: []interface{}{expectedRole},
							},
						},
					},
				},
				URN: providerURN,
			}
			a := domain.Grant{
				Resource: &domain.Resource{
					Type: shield.ResourceTypeTeam,
					URN:  "team:team_id",
					Name: "team_1",
					Details: map[string]interface{}{
						"id":     "team_id",
						"orgId":  "456",
						"admins": []interface{}{"testAdmin@email.com"},
						"metadata": shield.Metadata{
							Email:   "team_1@gojek.com",
							Privacy: "public",
							Slack:   "team_1_slack",
						},
					},
				},
				Role:       "member",
				AccountID:  expectedUserEmail,
				ResourceID: "999",
				ID:         "999",
			}

			actualError := p.GrantAccess(pc, a)

			assert.Nil(t, actualError)
		})
	})

	t.Run("given project resource", func(t *testing.T) {
		t.Run("should return error if there is an error in granting project access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			client := new(mocks.ShieldClient)
			logger := log.NewLogrus(log.LogrusWithLevel("info"))
			p := shield.NewProvider("", logger)
			p.Clients = map[string]shield.ShieldClient{
				providerURN: client,
			}

			expectedUserEmail := "test@email.com"
			expectedUser := &shield.User{
				ID:    "test_user_id",
				Name:  "test_user",
				Email: expectedUserEmail,
			}

			client.On("GetSelfUser", expectedUserEmail).Return(expectedUser, nil).Once()
			client.On("GrantProjectAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

			pc := &domain.ProviderConfig{
				Credentials: shield.Credentials{
					Host:      "localhost",
					AuthEmail: "test_email",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: shield.ResourceTypeProject,
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
					Type: shield.ResourceTypeProject,
					URN:  "project:project_id",
					Name: "project_1",
					Details: map[string]interface{}{
						"id":     "project_id",
						"orgId":  "456",
						"admins": []interface{}{"testAdmin@email.com"},
					},
				},
				Role:        "test-role",
				AccountID:   expectedUserEmail,
				Permissions: []string{"test-permission-config"},
			}

			actualError := p.GrantAccess(pc, a)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return nil error if granting access is successful", func(t *testing.T) {
			providerURN := "test-provider-urn"
			logger := log.NewLogrus(log.LogrusWithLevel("info"))
			client := new(mocks.ShieldClient)
			expectedProject := &shield.Project{
				Name: "project_1",
				ID:   "project_id",
			}
			expectedUserEmail := "test@email.com"
			expectedUser := &shield.User{
				ID:    "test_user_id",
				Name:  "test_user",
				Email: expectedUserEmail,
			}

			expectedRole := "admins"
			p := shield.NewProvider("", logger)
			p.Clients = map[string]shield.ShieldClient{
				providerURN: client,
			}

			client.On("GetSelfUser", expectedUserEmail).Return(expectedUser, nil).Once()
			client.On("GrantProjectAccess", expectedProject, expectedUser.ID, expectedRole).Return(nil).Once()

			pc := &domain.ProviderConfig{
				Credentials: shield.Credentials{
					Host:      "localhost",
					AuthEmail: "test_email",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: shield.ResourceTypeProject,
						Roles: []*domain.Role{
							{
								ID:          "admin",
								Permissions: []interface{}{expectedRole},
							},
						},
					},
				},
				URN: providerURN,
			}
			a := domain.Grant{
				Resource: &domain.Resource{
					Type: shield.ResourceTypeProject,
					URN:  "project:project_id",
					Name: "project_1",
					Details: map[string]interface{}{
						"id":     "project_id",
						"orgId":  "456",
						"admins": []interface{}{"testAdmin@email.com"},
					},
				},
				Role:       "admin",
				AccountID:  expectedUserEmail,
				ResourceID: "999",
				ID:         "999",
			}

			actualError := p.GrantAccess(pc, a)

			assert.Nil(t, actualError)
		})
	})

	t.Run("given organization resource", func(t *testing.T) {
		t.Run("should return error if there is an error in granting organization access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			client := new(mocks.ShieldClient)
			logger := log.NewLogrus(log.LogrusWithLevel("info"))
			p := shield.NewProvider("", logger)
			p.Clients = map[string]shield.ShieldClient{
				providerURN: client,
			}

			expectedUserEmail := "test@email.com"
			expectedUser := &shield.User{
				ID:    "test_user_id",
				Name:  "test_user",
				Email: expectedUserEmail,
			}

			client.On("GetSelfUser", expectedUserEmail).Return(expectedUser, nil).Once()
			client.On("GrantOrganizationAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

			pc := &domain.ProviderConfig{
				Credentials: shield.Credentials{
					Host:      "localhost",
					AuthEmail: "test_email",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: shield.ResourceTypeOrganization,
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
					Type: shield.ResourceTypeOrganization,
					URN:  "organization:org_id",
					Name: "org_1",
					Details: map[string]interface{}{
						"id":     "org_id",
						"admins": []interface{}{"testAdmin@email.com"},
					},
				},
				Role:        "test-role",
				AccountID:   expectedUserEmail,
				Permissions: []string{"test-permission-config"},
			}

			actualError := p.GrantAccess(pc, a)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return nil error if granting access is successful", func(t *testing.T) {
			providerURN := "test-provider-urn"
			logger := log.NewLogrus(log.LogrusWithLevel("info"))
			client := new(mocks.ShieldClient)
			expectedOrganization := &shield.Organization{
				Name: "org_1",
				ID:   "org_id",
			}
			expectedUserEmail := "test@email.com"
			expectedUser := &shield.User{
				ID:    "test_user_id",
				Name:  "test_user",
				Email: expectedUserEmail,
			}

			expectedRole := "admins"
			p := shield.NewProvider("", logger)
			p.Clients = map[string]shield.ShieldClient{
				providerURN: client,
			}

			client.On("GetSelfUser", expectedUserEmail).Return(expectedUser, nil).Once()
			client.On("GrantOrganizationAccess", expectedOrganization, expectedUser.ID, expectedRole).Return(nil).Once()

			pc := &domain.ProviderConfig{
				Credentials: shield.Credentials{
					Host:      "localhost",
					AuthEmail: "test_email",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: shield.ResourceTypeOrganization,
						Roles: []*domain.Role{
							{
								ID:          "admin",
								Permissions: []interface{}{expectedRole},
							},
						},
					},
				},
				URN: providerURN,
			}
			a := domain.Grant{
				Resource: &domain.Resource{
					Type: shield.ResourceTypeOrganization,
					URN:  "organization:org_id",
					Name: "org_1",
					Details: map[string]interface{}{
						"id":     "org_id",
						"admins": []interface{}{"testAdmin@email.com"},
					},
				},
				Role:       "admin",
				AccountID:  expectedUserEmail,
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
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		p := shield.NewProvider("", logger)

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

	t.Run("should return error if resource type in unknown", func(t *testing.T) {
		providerURN := "test-provider-urn"
		client := new(mocks.ShieldClient)
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		p := shield.NewProvider("", logger)
		p.Clients = map[string]shield.ShieldClient{
			providerURN: client,
		}

		expectedError := errors.New("invalid resource type")
		expectedUserEmail := "test@email.com"
		expectedUser := &shield.User{
			ID:    "test_user_id",
			Name:  "test_user",
			Email: expectedUserEmail,
		}

		client.On("GetSelfUser", expectedUserEmail).Return(expectedUser, nil).Once()

		pc := &domain.ProviderConfig{
			Credentials: shield.Credentials{
				Host:      "http://localhost/",
				AuthEmail: "test_email",
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
			URN: providerURN,
		}
		a := domain.Grant{
			Resource: &domain.Resource{
				Type: "test-type",
			},
			Role:      "test-role",
			AccountID: expectedUserEmail,
		}

		actualError := p.RevokeAccess(pc, a)
		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("given team resource", func(t *testing.T) {
		t.Run("should return error if there is an error in revoking team access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			client := new(mocks.ShieldClient)
			logger := log.NewLogrus(log.LogrusWithLevel("info"))
			p := shield.NewProvider("", logger)
			p.Clients = map[string]shield.ShieldClient{
				providerURN: client,
			}

			expectedUserEmail := "test@email.com"
			expectedUser := &shield.User{
				ID:    "test_user_id",
				Name:  "test_user",
				Email: expectedUserEmail,
			}

			client.On("GetSelfUser", expectedUserEmail).Return(expectedUser, nil).Once()
			client.On("RevokeTeamAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

			pc := &domain.ProviderConfig{
				Credentials: shield.Credentials{
					Host:      "localhost",
					AuthEmail: "test_email",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: shield.ResourceTypeTeam,
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
					Type: shield.ResourceTypeTeam,
					URN:  "team:team_id",
					Name: "team_1",
					Details: map[string]interface{}{
						"id":     "team_id",
						"orgId":  "456",
						"admins": []interface{}{"testAdmin@email.com"},
						"metadata": shield.Metadata{
							Email:   "team_1@gojek.com",
							Privacy: "public",
							Slack:   "team_1_slack",
						},
					},
				},
				Role:        "test-role",
				AccountID:   expectedUserEmail,
				Permissions: []string{"test-permission-config"},
			}

			actualError := p.RevokeAccess(pc, a)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return nil error if revoking team access is successful", func(t *testing.T) {
			providerURN := "test-provider-urn"
			logger := log.NewLogrus(log.LogrusWithLevel("info"))
			client := new(mocks.ShieldClient)
			expectedTeam := &shield.Team{
				Name:  "team_1",
				ID:    "team_id",
				OrgId: "456",
				Metadata: shield.Metadata{
					Email:   "team_1@gojek.com",
					Privacy: "public",
					Slack:   "team_1_slack",
				},
				Admins: []string{"testAdmin@email.com"},
			}

			expectedRole := "admins"
			p := shield.NewProvider("", logger)
			p.Clients = map[string]shield.ShieldClient{
				providerURN: client,
			}

			expectedUserEmail := "test@email.com"
			expectedUser := &shield.User{
				ID:    "test_user_id",
				Name:  "test_user",
				Email: expectedUserEmail,
			}

			client.On("GetSelfUser", expectedUserEmail).Return(expectedUser, nil).Once()
			client.On("RevokeTeamAccess", expectedTeam, expectedUser.ID, expectedRole).Return(nil).Once()

			pc := &domain.ProviderConfig{
				Credentials: shield.Credentials{
					Host:      "localhost",
					AuthEmail: "test_email",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: shield.ResourceTypeTeam,
						Roles: []*domain.Role{
							{
								ID:          "admin",
								Permissions: []interface{}{expectedRole},
							},
						},
					},
				},
				URN: providerURN,
			}
			a := domain.Grant{
				Resource: &domain.Resource{
					Type: shield.ResourceTypeTeam,
					URN:  "team:team_id",
					Name: "team_1",
					Details: map[string]interface{}{
						"id":     "team_id",
						"orgId":  "456",
						"admins": []interface{}{"testAdmin@email.com"},
						"metadata": shield.Metadata{
							Email:   "team_1@gojek.com",
							Privacy: "public",
							Slack:   "team_1_slack",
						},
					},
				},
				Role:        "admin",
				Permissions: []string{expectedRole},
				AccountID:   expectedUserEmail,
				ResourceID:  "999",
				ID:          "999",
			}

			actualError := p.RevokeAccess(pc, a)

			assert.Nil(t, actualError)
			client.AssertExpectations(t)
		})
	})

	t.Run("given project resource", func(t *testing.T) {
		t.Run("should return error if there is an error in revoking project access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			client := new(mocks.ShieldClient)
			logger := log.NewLogrus(log.LogrusWithLevel("info"))
			p := shield.NewProvider("", logger)
			p.Clients = map[string]shield.ShieldClient{
				providerURN: client,
			}

			expectedUserEmail := "test@email.com"
			expectedUser := &shield.User{
				ID:    "test_user_id",
				Name:  "test_user",
				Email: expectedUserEmail,
			}

			client.On("GetSelfUser", expectedUserEmail).Return(expectedUser, nil).Once()

			client.On("RevokeProjectAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

			pc := &domain.ProviderConfig{
				Credentials: shield.Credentials{
					Host:      "localhost",
					AuthEmail: "test_email",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: shield.ResourceTypeProject,
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
					Type: shield.ResourceTypeProject,
					URN:  "project:project_id",
					Name: "project_1",
					Details: map[string]interface{}{
						"id":     "project_id",
						"orgId":  "456",
						"admins": []interface{}{"testAdmin@email.com"},
					},
				},
				Role:        "test-role",
				AccountID:   expectedUserEmail,
				Permissions: []string{"test-permission-config"},
			}

			actualError := p.RevokeAccess(pc, a)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return nil error if revoking access is successful", func(t *testing.T) {
			providerURN := "test-provider-urn"
			client := new(mocks.ShieldClient)
			expectedProject := &shield.Project{
				Name:   "project_1",
				ID:     "project_id",
				OrgId:  "456",
				Admins: []string{"testAdmin@email.com"},
			}
			expectedRole := "admins"
			logger := log.NewLogrus(log.LogrusWithLevel("info"))
			p := shield.NewProvider("", logger)

			p.Clients = map[string]shield.ShieldClient{
				providerURN: client,
			}

			expectedUserEmail := "test@email.com"
			expectedUser := &shield.User{
				ID:    "test_user_id",
				Name:  "test_user",
				Email: expectedUserEmail,
			}

			client.On("GetSelfUser", expectedUserEmail).Return(expectedUser, nil).Once()
			client.On("RevokeProjectAccess", expectedProject, expectedUser.ID, expectedRole).Return(nil).Once()

			pc := &domain.ProviderConfig{
				Credentials: shield.Credentials{
					Host:      "localhost",
					AuthEmail: "test_email",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: shield.ResourceTypeProject,
						Roles: []*domain.Role{
							{
								ID:          "admin",
								Permissions: []interface{}{expectedRole},
							},
						},
					},
				},
				URN: providerURN,
			}
			a := domain.Grant{
				Resource: &domain.Resource{
					Type: shield.ResourceTypeProject,
					URN:  "project:project_id",
					Name: "project_1",
					Details: map[string]interface{}{
						"id":     "project_id",
						"orgId":  "456",
						"admins": []interface{}{"testAdmin@email.com"},
					},
				},
				Role:        "admin",
				Permissions: []string{expectedRole},
				AccountID:   expectedUserEmail,
				ResourceID:  "999",
				ID:          "999",
			}

			actualError := p.RevokeAccess(pc, a)

			assert.Nil(t, actualError)
			client.AssertExpectations(t)
		})
	})

	t.Run("given Organization resource", func(t *testing.T) {
		t.Run("should return error if there is an error in revoking organization access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			client := new(mocks.ShieldClient)
			logger := log.NewLogrus(log.LogrusWithLevel("info"))
			p := shield.NewProvider("", logger)
			p.Clients = map[string]shield.ShieldClient{
				providerURN: client,
			}

			expectedUserEmail := "test@email.com"
			expectedUser := &shield.User{
				ID:    "test_user_id",
				Name:  "test_user",
				Email: expectedUserEmail,
			}

			client.On("GetSelfUser", expectedUserEmail).Return(expectedUser, nil).Once()
			client.On("RevokeOrganizationAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

			pc := &domain.ProviderConfig{
				Credentials: shield.Credentials{
					Host:      "localhost",
					AuthEmail: "test_email",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: shield.ResourceTypeOrganization,
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
					Type: shield.ResourceTypeOrganization,
					URN:  "organization:org_id",
					Name: "org_1",
					Details: map[string]interface{}{
						"id":     "org_id",
						"admins": []interface{}{"testAdmin@email.com"},
					},
				},
				Role:        "test-role",
				AccountID:   expectedUserEmail,
				Permissions: []string{"test-permission-config"},
			}

			actualError := p.RevokeAccess(pc, a)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return nil error if revoking access is successful", func(t *testing.T) {
			providerURN := "test-provider-urn"
			client := new(mocks.ShieldClient)
			expectedOrganization := &shield.Organization{
				Name:   "org_1",
				ID:     "org_id",
				Admins: []string{"testAdmin@email.com"},
			}
			expectedRole := "admins"
			logger := log.NewLogrus(log.LogrusWithLevel("info"))
			p := shield.NewProvider("", logger)

			p.Clients = map[string]shield.ShieldClient{
				providerURN: client,
			}
			expectedUserEmail := "test@email.com"
			expectedUser := &shield.User{
				ID:    "test_user_id",
				Name:  "test_user",
				Email: expectedUserEmail,
			}

			client.On("GetSelfUser", expectedUserEmail).Return(expectedUser, nil).Once()
			client.On("RevokeOrganizationAccess", expectedOrganization, expectedUser.ID, expectedRole).Return(nil).Once()

			pc := &domain.ProviderConfig{
				Credentials: shield.Credentials{
					Host:      "localhost",
					AuthEmail: "test_email",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: shield.ResourceTypeOrganization,
						Roles: []*domain.Role{
							{
								ID:          "admin",
								Permissions: []interface{}{expectedRole},
							},
						},
					},
				},
				URN: providerURN,
			}
			a := domain.Grant{
				Resource: &domain.Resource{
					Type: shield.ResourceTypeOrganization,
					URN:  "organization:org_id",
					Name: "org_1",
					Details: map[string]interface{}{
						"id":     "org_id",
						"admins": []interface{}{"testAdmin@email.com"},
					},
				},
				Role:        "admin",
				Permissions: []string{expectedRole},
				AccountID:   expectedUserEmail,
				ResourceID:  "999",
				ID:          "999",
			}

			actualError := p.RevokeAccess(pc, a)

			assert.Nil(t, actualError)
			client.AssertExpectations(t)
		})
	})
}

func TestGetAccountTypes(t *testing.T) {
	expectedAccountType := []string{"user"}
	logger := log.NewLogrus(log.LogrusWithLevel("info"))
	p := shield.NewProvider("", logger)

	actualAccountType := p.GetAccountTypes()

	assert.Equal(t, expectedAccountType, actualAccountType)
}

func TestGetRoles(t *testing.T) {
	t.Run("should return error if resource type is invalid", func(t *testing.T) {
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		p := shield.NewProvider("shield", logger)
		validConfig := &domain.ProviderConfig{
			Type:                "shield",
			URN:                 "test-URN",
			AllowedAccountTypes: []string{"user"},
			Credentials:         map[string]interface{}{},
			Resources: []*domain.ResourceConfig{
				{
					Type: "team",
					Policy: &domain.PolicyConfig{
						ID:      "test-policy-1",
						Version: 1,
					},
				},
				{
					Type: "project",
					Policy: &domain.PolicyConfig{
						ID:      "test-policy-2",
						Version: 1,
					},
				},
				{
					Type: "organization",
					Policy: &domain.PolicyConfig{
						ID:      "test-policy-3",
						Version: 1,
					},
				},
			},
		}

		actualRoles, actualError := p.GetRoles(validConfig, "invalid_resource_type")

		assert.Nil(t, actualRoles)
		assert.ErrorIs(t, actualError, provider.ErrInvalidResourceType)
	})

	t.Run("should return roles specified in the provider config", func(t *testing.T) {
		logger := log.NewLogrus(log.LogrusWithLevel("info"))
		p := shield.NewProvider("shield", logger)

		expectedRoles := []*domain.Role{
			{
				ID:   "test-role",
				Name: "test_role_name",
			},
		}

		validConfig := &domain.ProviderConfig{
			Type:                "shield",
			URN:                 "test-URN",
			AllowedAccountTypes: []string{"user"},
			Credentials:         map[string]interface{}{},
			Resources: []*domain.ResourceConfig{
				{
					Type: "team",
					Policy: &domain.PolicyConfig{
						ID:      "test-policy-1",
						Version: 1,
					},
					Roles: expectedRoles,
				},
			},
		}

		actualRoles, actualError := p.GetRoles(validConfig, "team")

		assert.NoError(t, actualError)
		assert.Equal(t, expectedRoles, actualRoles)
	})
}
