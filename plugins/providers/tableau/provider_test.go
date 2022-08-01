package tableau_test

import (
	"errors"
	"testing"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/plugins/providers/tableau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetType(t *testing.T) {
	t.Run("should return provider type name", func(t *testing.T) {
		expectedTypeName := domain.ProviderTypeTableau
		crypto := new(mocks.Crypto)
		p := tableau.NewProvider(expectedTypeName, crypto)

		actualTypeName := p.GetType()

		assert.Equal(t, expectedTypeName, actualTypeName)
	})
}

func TestCreateConfig(t *testing.T) {
	t.Run("should return error if there credentials are invalid", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.TableauClient)
		p := tableau.NewProvider("", crypto)
		p.Clients = map[string]tableau.TableauClient{
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
					Credentials: tableau.Credentials{
						Host:       "",
						Username:   "",
						Password:   "",
						ContentURL: "",
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
		client := new(mocks.TableauClient)
		p := tableau.NewProvider("", crypto)
		p.Clients = map[string]tableau.TableauClient{
			providerURN: client,
		}

		testcases := []struct {
			pc *domain.ProviderConfig
		}{
			{
				pc: &domain.ProviderConfig{
					Credentials: tableau.Credentials{
						Host:       "localhost",
						Username:   "test-username",
						Password:   "test-password",
						ContentURL: "test-content-url",
					},
					Resources: []*domain.ResourceConfig{ //resource type wrong requires one of "workbook" or "flow" or "datasource" or "view" or "metric"
						{
							Type: "invalid resource type",
						},
					},
				},
			},
			{
				pc: &domain.ProviderConfig{
					Credentials: tableau.Credentials{
						Host:       "localhost",
						Username:   "test-username",
						Password:   "test-password",
						ContentURL: "test-content-url",
					},
					Resources: []*domain.ResourceConfig{
						{
							Type: tableau.ResourceTypeWorkbook, // Workbook resource type
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
		client := new(mocks.TableauClient)
		validCredentials := tableau.Credentials{
			Host:       "http://localhost",
			Username:   "test-username",
			Password:   "test-password",
			ContentURL: "test-content-url",
		}
		p := tableau.NewProvider("", crypto)
		p.Clients = map[string]tableau.TableauClient{
			providerURN: client,
		}
		crypto.On("Encrypt", "test-password").Return("encrypted-test-pasword", nil)

		testcases := []struct {
			pc            *domain.ProviderConfig
			expectedError error
		}{
			{
				pc: &domain.ProviderConfig{
					Type:        "tableau",
					Credentials: validCredentials,
					Resources: []*domain.ResourceConfig{
						{
							Type: tableau.ResourceTypeWorkbook,
							Roles: []*domain.Role{
								{
									ID:   "read",
									Name: "Read",
									Permissions: []interface{}{
										map[string]interface{}{
											"name": "Read:Allow", //Valid permissions:  "AddComment", "ChangeHierarchy", "ChangePermissions", "Delete", "ExportData", "ExportImage", "ExportXml", "Filter", "Read", "ShareView", "ViewComments", "ViewUnderlyingData", "WebAuthoring", "Write"
										},
									},
								},
								{
									ID:   "viewer",
									Name: "Viewer",
									Permissions: []interface{}{
										map[string]interface{}{
											"name": "Viewer", // Permissions for site-role : "Creator", "Explorer", "ExplorerCanPublish", "SiteAdministratorExplorer", "SiteAdministratorCreator", "Unlicensed", "Viewer",
											"type": "site_role",
										},
									},
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
					Type:        "tableau",
					Credentials: validCredentials,
					Resources: []*domain.ResourceConfig{
						{
							Type: tableau.ResourceTypeFlow,
							Roles: []*domain.Role{
								{
									ID:   "change-hierarchy",
									Name: "test-name",
									Permissions: []interface{}{
										map[string]interface{}{
											"name": "ChangeHierarchy:Allow", // Valid permissions : "ChangeHierarchy", "ChangePermissions", "Delete", "Execute", "ExportXml", "Read", "Write"
										},
									},
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
					Type:        "tableau",
					Credentials: validCredentials,
					Resources: []*domain.ResourceConfig{
						{
							Type: tableau.ResourceTypeDataSource,
							Roles: []*domain.Role{
								{
									ID:   "data-source",
									Name: "test-name",
									Permissions: []interface{}{
										map[string]interface{}{
											"name": "ChangePermissions:Allow", // valid permissions - "ChangePermissions", "Connect", "Delete", "ExportXml", "Read", "Write"
										},
									},
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
					Type:        "tableau",
					Credentials: validCredentials,
					Resources: []*domain.ResourceConfig{
						{
							Type: tableau.ResourceTypeView,
							Roles: []*domain.Role{
								{
									ID:   "view",
									Name: "test-name",
									Permissions: []interface{}{
										map[string]interface{}{
											"name": "AddComment:Allow", // valid permissions : "ChangePermissions", "Connect", "Delete", "ExportXml", "Read", "Write"
										},
									},
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
					Type:        "tableau",
					Credentials: validCredentials,
					Resources: []*domain.ResourceConfig{
						{
							Type: tableau.ResourceTypeMetric,
							Roles: []*domain.Role{
								{
									ID:   "metric",
									Name: "test-name",
									Permissions: []interface{}{
										map[string]interface{}{
											"name": "Delete:Allow", // valid permissions:  "Delete", "Read", "Write" , modes: "Allow", "Deny"
										},
									},
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
		crypto := new(mocks.Crypto)
		p := tableau.NewProvider("", crypto)

		pc := &domain.ProviderConfig{
			Credentials: "invalid-creds",
		}

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.Error(t, actualError)
	})

	t.Run("should return error if there credentials couldnt be decrypted", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := tableau.NewProvider("", crypto)

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

	t.Run("should return error if there HTTP client isn't valid", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := tableau.NewProvider("", crypto)

		crypto.On("Decrypt", "test-password").Return("correct-password", nil).Once()
		pc := &domain.ProviderConfig{
			Type: "test-URN",
			Credentials: map[string]interface{}{
				"password":    "test-password",
				"host":        "http://localhost",
				"content_url": "test-content-url",
				"username":    "username",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: "test-type",
					Roles: []*domain.Role{
						{
							ID: "test-role",
							Permissions: []interface{}{
								tableau.Permission{
									Name: "test-permission-config",
								},
							},
						},
					},
				},
			},
		}

		actualResources, actualError := p.GetResources(pc)

		assert.Error(t, actualError)
		assert.Nil(t, actualResources)
	})

	t.Run("should return error if got any on getting workbook resources", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.TableauClient)
		p := tableau.NewProvider("", crypto)
		p.Clients = map[string]tableau.TableauClient{
			providerURN: client,
		}

		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: map[string]interface{}{},
			Resources: []*domain.ResourceConfig{
				{
					Type: tableau.ResourceTypeWorkbook,
				},
			},
		}
		expectedError := errors.New("client error")
		client.On("GetWorkbooks").Return(nil, expectedError).Once()

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return error if got any on getting flow resources", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.TableauClient)
		p := tableau.NewProvider("", crypto)
		p.Clients = map[string]tableau.TableauClient{
			providerURN: client,
		}

		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: map[string]interface{}{},
			Resources: []*domain.ResourceConfig{
				{
					Type: tableau.ResourceTypeFlow,
				},
			},
		}
		expectedError := errors.New("client error")
		client.On("GetFlows").Return(nil, expectedError).Once()

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return error if got any on getting datasource resources", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.TableauClient)
		p := tableau.NewProvider("", crypto)
		p.Clients = map[string]tableau.TableauClient{
			providerURN: client,
		}

		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: map[string]interface{}{},
			Resources: []*domain.ResourceConfig{
				{
					Type: tableau.ResourceTypeDataSource,
				},
			},
		}
		expectedError := errors.New("client error")
		client.On("GetDataSources").Return(nil, expectedError).Once()

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return error if got any on getting view resources", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.TableauClient)
		p := tableau.NewProvider("", crypto)
		p.Clients = map[string]tableau.TableauClient{
			providerURN: client,
		}

		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: map[string]interface{}{},
			Resources: []*domain.ResourceConfig{
				{
					Type: tableau.ResourceTypeView,
				},
			},
		}
		expectedError := errors.New("client error")
		client.On("GetViews").Return(nil, expectedError).Once()

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return error if got any on getting metric resources", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.TableauClient)
		p := tableau.NewProvider("", crypto)
		p.Clients = map[string]tableau.TableauClient{
			providerURN: client,
		}

		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: map[string]interface{}{},
			Resources: []*domain.ResourceConfig{
				{
					Type: tableau.ResourceTypeMetric,
				},
			},
		}
		expectedError := errors.New("client error")
		client.On("GetMetrics").Return(nil, expectedError).Once()

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return list of resources and nil error on success", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.TableauClient)
		p := tableau.NewProvider("", crypto)
		p.Clients = map[string]tableau.TableauClient{
			providerURN: client,
		}

		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: map[string]interface{}{},
			Resources: []*domain.ResourceConfig{
				{
					Type: tableau.ResourceTypeWorkbook,
				},
				{
					Type: tableau.ResourceTypeFlow,
				},
				{
					Type: tableau.ResourceTypeDataSource,
				},
				{
					Type: tableau.ResourceTypeView,
				},
				{
					Type: tableau.ResourceTypeMetric,
				},
			},
		}
		expectedWorkbooks := []*tableau.Workbook{
			{
				ID:   "workbook-1",
				Name: "wb_1",
			},
		}
		client.On("GetWorkbooks").Return(expectedWorkbooks, nil).Once()
		expectedFlows := []*tableau.Flow{
			{
				ID:   "flow-1",
				Name: "fl_1",
			},
		}
		client.On("GetFlows").Return(expectedFlows, nil).Once()
		expectedDataSources := []*tableau.DataSource{
			{
				ID:   "datasource-1",
				Name: "ds_1",
			},
		}
		client.On("GetDataSources").Return(expectedDataSources, nil).Once()
		expectedViews := []*tableau.View{
			{
				ID:   "view-1",
				Name: "vw_1",
			},
		}
		client.On("GetViews").Return(expectedViews, nil).Once()
		expectedMetrics := []*tableau.Metric{
			{
				ID:   "metric-1",
				Name: "mt_1",
			},
		}
		client.On("GetMetrics").Return(expectedMetrics, nil).Once()

		expectedResources := []*domain.Resource{
			{
				Type:        tableau.ResourceTypeWorkbook,
				URN:         "workbook-1",
				ProviderURN: providerURN,
				Name:        "wb_1",
				Details: map[string]interface{}{
					"project_name":    "",
					"project_id":      "",
					"owner_name":      "",
					"owner_id":        "",
					"content_url":     "",
					"webpage_url":     "",
					"size":            "",
					"default_view_id": "",
					"tags":            nil,
					"show_tabs":       "",
				},
			},
			{
				Type:        tableau.ResourceTypeFlow,
				URN:         "flow-1",
				ProviderURN: providerURN,
				Name:        "fl_1",
				Details: map[string]interface{}{
					"project_name": "",
					"project_id":   "",
					"owner_id":     "",
					"webpage_url":  "",
					"tags":         nil,
					"fileType":     "",
				},
			},
			{
				Type:        tableau.ResourceTypeDataSource,
				URN:         "datasource-1",
				ProviderURN: providerURN,
				Name:        "ds_1",
				Details: map[string]interface{}{
					"project_name":        "",
					"project_id":          "",
					"owner_id":            "",
					"content_url":         "",
					"webpage_url":         "",
					"tags":                nil,
					"encryptExtracts":     "",
					"hasExtracts":         false,
					"isCertified":         false,
					"type":                "",
					"useRemoteQueryAgent": false,
				},
			}, {
				Type:        tableau.ResourceTypeView,
				URN:         "view-1",
				ProviderURN: providerURN,
				Name:        "vw_1",
				Details: map[string]interface{}{
					"project_name": "",
					"project_id":   "",
					"owner_name":   "",
					"workbook_id":  "",
					"owner_id":     "",
					"content_url":  "",
					"tags":         nil,
					"viewUrlName":  "",
				},
			}, {
				Type:        tableau.ResourceTypeMetric,
				URN:         "metric-1",
				ProviderURN: providerURN,
				Name:        "mt_1",
				Details: map[string]interface{}{
					"project_name":   "",
					"project_id":     "",
					"owner_id":       "",
					"webpage_url":    "",
					"tags":           nil,
					"description":    "",
					"underlyingView": tableau.UnderlyingView{ID: ""},
					"suspended":      false,
				},
			},
		}

		actualResources, actualError := p.GetResources(pc)

		assert.Equal(t, expectedResources, actualResources)
		assert.Nil(t, actualError)
		client.AssertExpectations(t)
	})
}

func TestGrantAccess(t *testing.T) {
	t.Run("should return error if credentials is invalid", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := tableau.NewProvider("", crypto)

		pc := &domain.ProviderConfig{
			Credentials: "invalid-credentials",
			Resources: []*domain.ResourceConfig{
				{
					Type: "test-type",
					Roles: []*domain.Role{
						{
							ID: "test-role",
							Permissions: []interface{}{
								tableau.Permission{
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

	t.Run("should return error if there are any on client initialization", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := tableau.NewProvider("", crypto)
		expectedError := errors.New("decrypt error")
		crypto.On("Decrypt", "test-password").Return("", expectedError).Once()

		pc := &domain.ProviderConfig{
			Credentials: tableau.Credentials{
				Host:       "localhost",
				Username:   "test-username",
				Password:   "test-password",
				ContentURL: "test-content-url",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: "test-type",
					Roles: []*domain.Role{
						{
							ID: "test-role",
							Permissions: []interface{}{
								tableau.Permission{
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

	t.Run("should return error if resource type in unknown", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := tableau.NewProvider("", crypto)
		expectedError := errors.New("invalid resource type")
		crypto.On("Decrypt", "test-password").Return("", expectedError).Once()

		pc := &domain.ProviderConfig{
			Credentials: tableau.Credentials{
				Host:       "localhost",
				Username:   "test-username",
				Password:   "test-password",
				ContentURL: "test-content-url",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: "test-type",
					Roles: []*domain.Role{
						{
							ID: "test-role",
							Permissions: []interface{}{
								tableau.Permission{
									Name: "test-permission-config",
								},
							},
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

	t.Run("given workbook resource", func(t *testing.T) {
		t.Run("should return error if there is an error in granting workbook access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")

			crypto := new(mocks.Crypto)
			client := new(mocks.TableauClient)
			p := tableau.NewProvider("", crypto)
			p.Clients = map[string]tableau.TableauClient{
				providerURN: client,
			}
			validSiteRole := []*domain.Role{
				{
					ID: "test-role",
					Permissions: []interface{}{
						tableau.Permission{
							Name: "Creator",
							Type: "site_role",
						},
					},
				},
			}
			validRole := []*domain.Role{
				{
					ID: "test-role",
					Permissions: []interface{}{
						tableau.Permission{
							Name: "Read:Allow",
						},
					},
				},
			}
			validCredentials := tableau.Credentials{
				Host:       "localhost",
				Username:   "test-username",
				Password:   "test-password",
				ContentURL: "test-content-url",
			}

			client.On("GrantWorkbookAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()
			client.On("UpdateSiteRole", mock.Anything, mock.Anything).Return(expectedError).Once()

			testcases := []struct {
				pc   *domain.ProviderConfig
				name string
				a    *domain.Appeal
			}{
				{
					pc: &domain.ProviderConfig{
						Credentials: validCredentials,
						Resources: []*domain.ResourceConfig{
							{
								Type:  tableau.ResourceTypeWorkbook,
								Roles: validSiteRole,
							},
						},
						URN: providerURN,
					},
					name: "Provider Config with Site Role Permissions",
					a: &domain.Appeal{
						Resource: &domain.Resource{
							Type: tableau.ResourceTypeWorkbook,
							URN:  "999",
							Name: "test-workbook",
						},
						Role:        "test-role",
						Permissions: []string{"Creator@site_role"},
					},
				},
				{
					pc: &domain.ProviderConfig{
						Credentials: validCredentials,
						Resources: []*domain.ResourceConfig{
							{
								Type:  tableau.ResourceTypeWorkbook,
								Roles: validRole,
							},
						},
						URN: providerURN,
					},
					name: "Provider Config with Workook Permissions without site role permission",
					a: &domain.Appeal{
						Resource: &domain.Resource{
							Type: tableau.ResourceTypeWorkbook,
							URN:  "999",
							Name: "test-workbook",
						},
						Role:        "test-role",
						Permissions: []string{"Read:Allow"},
					},
				},
			}

			for _, tc := range testcases {
				t.Run(tc.name, func(t *testing.T) {
					actualError := p.GrantAccess(tc.pc, tc.a)

					assert.EqualError(t, actualError, expectedError.Error())
				})
			}
		})

		t.Run("should return nil error if granting access is successful", func(t *testing.T) {
			providerURN := "test-provider-urn"
			crypto := new(mocks.Crypto)
			client := new(mocks.TableauClient)
			expectedWorkbook := &tableau.Workbook{
				Name: "test-workbook",
				ID:   "workbook-id",
			}
			expectedUser := "test@email.com"
			expectedRole := tableau.PermissionNames[tableau.ResourceTypeWorkbook][0] + ":" + tableau.PermissionModes[0]
			p := tableau.NewProvider("", crypto)
			p.Clients = map[string]tableau.TableauClient{
				providerURN: client,
			}
			client.On("GrantWorkbookAccess", expectedWorkbook, expectedUser, expectedRole).Return(nil).Once()

			pc := &domain.ProviderConfig{
				Credentials: tableau.Credentials{
					Host:       "localhost",
					Username:   "test-username",
					Password:   "test-password",
					ContentURL: "test-content-url",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: tableau.ResourceTypeWorkbook,
						Roles: []*domain.Role{
							{
								ID: "viewer",
								Permissions: []interface{}{
									tableau.Permission{
										Name: expectedRole,
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
					Type: tableau.ResourceTypeWorkbook,
					URN:  "workbook-id",
					Name: "test-workbook",
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

	t.Run("given flow resource", func(t *testing.T) {
		t.Run("should return error if there is an error in granting flow access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			crypto := new(mocks.Crypto)
			client := new(mocks.TableauClient)
			p := tableau.NewProvider("", crypto)
			p.Clients = map[string]tableau.TableauClient{
				providerURN: client,
			}

			validSiteRole := []*domain.Role{
				{
					ID: "test-role",
					Permissions: []interface{}{
						tableau.Permission{
							Name: "Creator",
							Type: "site_role",
						},
					},
				},
			}
			validRole := []*domain.Role{
				{
					ID: "test-role",
					Permissions: []interface{}{
						tableau.Permission{
							Name: "ChangeHierarchy:Allow",
						},
					},
				},
			}
			validCredentials := tableau.Credentials{
				Host:       "localhost",
				Username:   "test-username",
				Password:   "test-password",
				ContentURL: "test-content-url",
			}

			client.On("GrantFlowAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()
			client.On("UpdateSiteRole", mock.Anything, mock.Anything).Return(expectedError).Once()

			testcases := []struct {
				pc   *domain.ProviderConfig
				name string
				a    *domain.Appeal
			}{
				{
					pc: &domain.ProviderConfig{
						Credentials: validCredentials,
						Resources: []*domain.ResourceConfig{
							{
								Type:  tableau.ResourceTypeFlow,
								Roles: validSiteRole,
							},
						},
						URN: providerURN,
					},
					name: "Provider Config to Update Site Role Permissions",
					a: &domain.Appeal{
						Resource: &domain.Resource{
							Type: tableau.ResourceTypeFlow,
							URN:  "999",
							Name: "test-flow",
						},
						Role:        "test-role",
						Permissions: []string{"Creator@site_role"},
					},
				},
				{
					pc: &domain.ProviderConfig{
						Credentials: validCredentials,
						Resources: []*domain.ResourceConfig{
							{
								Type:  tableau.ResourceTypeFlow,
								Roles: validRole,
							},
						},
						URN: providerURN,
					},
					name: "Provider Config with to Grant Flow Permission",
					a: &domain.Appeal{
						Resource: &domain.Resource{
							Type: tableau.ResourceTypeFlow,
							URN:  "999",
							Name: "test-flow",
						},
						Role:        "test-role",
						Permissions: []string{"ChangeHierarchy:Allow"},
					},
				},
			}

			for _, tc := range testcases {
				t.Run(tc.name, func(t *testing.T) {
					actualError := p.GrantAccess(tc.pc, tc.a)

					assert.EqualError(t, actualError, expectedError.Error())
				})
			}
		})

		t.Run("should return nil error if granting access is successful", func(t *testing.T) {
			providerURN := "test-provider-urn"
			crypto := new(mocks.Crypto)
			client := new(mocks.TableauClient)
			expectedFlow := &tableau.Flow{
				Name: "test-flow",
				ID:   "flow-id",
			}
			expectedUser := "test@email.com"
			expectedRole := tableau.PermissionNames[tableau.ResourceTypeFlow][0] + ":" + tableau.PermissionModes[0]
			p := tableau.NewProvider("", crypto)
			p.Clients = map[string]tableau.TableauClient{
				providerURN: client,
			}
			client.On("GrantFlowAccess", expectedFlow, expectedUser, expectedRole).Return(nil).Once()

			pc := &domain.ProviderConfig{
				Credentials: tableau.Credentials{
					Host:       "localhost",
					Username:   "test-username",
					Password:   "test-password",
					ContentURL: "test-content-url",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: tableau.ResourceTypeFlow,
						Roles: []*domain.Role{
							{
								ID: "viewer",
								Permissions: []interface{}{
									tableau.Permission{
										Name: expectedRole,
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
					Type: tableau.ResourceTypeFlow,
					URN:  "flow-id",
					Name: "test-flow",
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

	t.Run("given view resource", func(t *testing.T) {
		t.Run("should return error if there is an error in granting view access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			crypto := new(mocks.Crypto)
			client := new(mocks.TableauClient)
			p := tableau.NewProvider("", crypto)
			p.Clients = map[string]tableau.TableauClient{
				providerURN: client,
			}

			validSiteRole := []*domain.Role{
				{
					ID: "test-role",
					Permissions: []interface{}{
						tableau.Permission{
							Name: "Creator",
							Type: "site_role",
						},
					},
				},
			}
			validRole := []*domain.Role{
				{
					ID: "test-role",
					Permissions: []interface{}{
						tableau.Permission{
							Name: "Connect:Allow",
						},
					},
				},
			}
			validCredentials := tableau.Credentials{
				Host:       "localhost",
				Username:   "test-username",
				Password:   "test-password",
				ContentURL: "test-content-url",
			}

			client.On("GrantViewAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()
			client.On("UpdateSiteRole", mock.Anything, mock.Anything).Return(expectedError).Once()

			testcases := []struct {
				pc   *domain.ProviderConfig
				name string
				a    *domain.Appeal
			}{
				{
					pc: &domain.ProviderConfig{
						Credentials: validCredentials,
						Resources: []*domain.ResourceConfig{
							{
								Type:  tableau.ResourceTypeView,
								Roles: validSiteRole,
							},
						},
						URN: providerURN,
					},
					name: "Provider Config to Update Site Role Permissions",
					a: &domain.Appeal{
						Resource: &domain.Resource{
							Type: tableau.ResourceTypeView,
							URN:  "999",
							Name: "test-view",
						},
						Role:        "test-role",
						Permissions: []string{"Creator@site_role"},
					},
				},
				{
					pc: &domain.ProviderConfig{
						Credentials: validCredentials,
						Resources: []*domain.ResourceConfig{
							{
								Type:  tableau.ResourceTypeView,
								Roles: validRole,
							},
						},
						URN: providerURN,
					},
					name: "Provider Config with to Grant View Permission",
					a: &domain.Appeal{
						Resource: &domain.Resource{
							Type: tableau.ResourceTypeView,
							URN:  "999",
							Name: "test-view",
						},
						Role:        "test-role",
						Permissions: []string{"Connect:Allow"},
					},
				},
			}

			for _, tc := range testcases {
				t.Run(tc.name, func(t *testing.T) {
					actualError := p.GrantAccess(tc.pc, tc.a)

					assert.EqualError(t, actualError, expectedError.Error())
				})
			}
		})

		t.Run("should return nil error if granting access is successful", func(t *testing.T) {
			providerURN := "test-provider-urn"
			crypto := new(mocks.Crypto)
			client := new(mocks.TableauClient)
			expectedView := &tableau.View{
				Name: "test-view",
				ID:   "view-id",
			}
			expectedUser := "test@email.com"
			expectedRole := tableau.PermissionNames[tableau.ResourceTypeView][0] + ":" + tableau.PermissionModes[0]
			p := tableau.NewProvider("", crypto)
			p.Clients = map[string]tableau.TableauClient{
				providerURN: client,
			}
			client.On("GrantViewAccess", expectedView, expectedUser, expectedRole).Return(nil).Once()

			pc := &domain.ProviderConfig{
				Credentials: tableau.Credentials{
					Host:       "localhost",
					Username:   "test-username",
					Password:   "test-password",
					ContentURL: "test-content-url",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: tableau.ResourceTypeView,
						Roles: []*domain.Role{
							{
								ID: "viewer",
								Permissions: []interface{}{
									tableau.Permission{
										Name: expectedRole,
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
					Type: tableau.ResourceTypeView,
					URN:  "view-id",
					Name: "test-view",
				},
				Role:       "viewer",
				AccountID:  expectedUser,
				ResourceID: "99",
				ID:         "99",
			}

			actualError := p.GrantAccess(pc, a)

			assert.Nil(t, actualError)
		})
	})

	t.Run("given metric resource", func(t *testing.T) {
		t.Run("should return error if there is an error in granting metric access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			crypto := new(mocks.Crypto)
			client := new(mocks.TableauClient)
			p := tableau.NewProvider("", crypto)
			p.Clients = map[string]tableau.TableauClient{
				providerURN: client,
			}
			validSiteRole := []*domain.Role{
				{
					ID: "test-role",
					Permissions: []interface{}{
						tableau.Permission{
							Name: "Creator",
							Type: "site_role",
						},
					},
				},
			}
			validRole := []*domain.Role{
				{
					ID: "test-role",
					Permissions: []interface{}{
						tableau.Permission{
							Name: "Delete:Allow",
						},
					},
				},
			}
			validCredentials := tableau.Credentials{
				Host:       "localhost",
				Username:   "test-username",
				Password:   "test-password",
				ContentURL: "test-content-url",
			}

			client.On("GrantMetricAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()
			client.On("UpdateSiteRole", mock.Anything, mock.Anything).Return(expectedError).Once()

			testcases := []struct {
				pc   *domain.ProviderConfig
				name string
				a    *domain.Appeal
			}{
				{
					pc: &domain.ProviderConfig{
						Credentials: validCredentials,
						Resources: []*domain.ResourceConfig{
							{
								Type:  tableau.ResourceTypeMetric,
								Roles: validSiteRole,
							},
						},
						URN: providerURN,
					},
					name: "Provider Config to Update Site Role Permissions",
					a: &domain.Appeal{
						Resource: &domain.Resource{
							Type: tableau.ResourceTypeMetric,
							URN:  "999",
							Name: "test-metric",
						},
						Role:        "test-role",
						Permissions: []string{"Creator@site_role"},
					},
				},
				{
					pc: &domain.ProviderConfig{
						Credentials: validCredentials,
						Resources: []*domain.ResourceConfig{
							{
								Type:  tableau.ResourceTypeMetric,
								Roles: validRole,
							},
						},
						URN: providerURN,
					},
					name: "Provider Config with to Grant Metric Permission",
					a: &domain.Appeal{
						Resource: &domain.Resource{
							Type: tableau.ResourceTypeMetric,
							URN:  "999",
							Name: "test-metric",
						},
						Role:        "test-role",
						Permissions: []string{"Delete:Allow"},
					},
				},
			}

			for _, tc := range testcases {
				t.Run(tc.name, func(t *testing.T) {
					actualError := p.GrantAccess(tc.pc, tc.a)

					assert.EqualError(t, actualError, expectedError.Error())
				})
			}
		})

		t.Run("should return nil error if granting access is successful", func(t *testing.T) {
			providerURN := "test-provider-urn"
			crypto := new(mocks.Crypto)
			client := new(mocks.TableauClient)
			expectedMetric := &tableau.Metric{
				Name: "test-metric",
				ID:   "metric-id",
			}
			expectedUser := "test@email.com"
			expectedRole := tableau.PermissionNames[tableau.ResourceTypeMetric][0] + ":" + tableau.PermissionModes[0]
			p := tableau.NewProvider("", crypto)
			p.Clients = map[string]tableau.TableauClient{
				providerURN: client,
			}
			client.On("GrantMetricAccess", expectedMetric, expectedUser, expectedRole).Return(nil).Once()

			pc := &domain.ProviderConfig{
				Credentials: tableau.Credentials{
					Host:       "localhost",
					Username:   "test-username",
					Password:   "test-password",
					ContentURL: "test-content-url",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: tableau.ResourceTypeMetric,
						Roles: []*domain.Role{
							{
								ID: "viewer",
								Permissions: []interface{}{
									tableau.Permission{
										Name: expectedRole,
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
					Type: tableau.ResourceTypeMetric,
					URN:  "metric-id",
					Name: "test-metric",
				},
				Role:       "viewer",
				AccountID:  expectedUser,
				ResourceID: "99",
				ID:         "99",
			}

			actualError := p.GrantAccess(pc, a)

			assert.Nil(t, actualError)
		})
	})

	t.Run("given datasource resource", func(t *testing.T) {
		t.Run("should return error if there is an error in granting datasource access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			crypto := new(mocks.Crypto)
			client := new(mocks.TableauClient)
			p := tableau.NewProvider("", crypto)
			p.Clients = map[string]tableau.TableauClient{
				providerURN: client,
			}

			validSiteRole := []*domain.Role{
				{
					ID: "test-role",
					Permissions: []interface{}{
						tableau.Permission{
							Name: "Creator",
							Type: "site_role",
						},
					},
				},
			}
			validRole := []*domain.Role{
				{
					ID: "test-role",
					Permissions: []interface{}{
						tableau.Permission{
							Name: "ChangePermissions:Allow",
						},
					},
				},
			}
			validCredentials := tableau.Credentials{
				Host:       "localhost",
				Username:   "test-username",
				Password:   "test-password",
				ContentURL: "test-content-url",
			}

			client.On("GrantDataSourceAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()
			client.On("UpdateSiteRole", mock.Anything, mock.Anything).Return(expectedError).Once()

			testcases := []struct {
				pc   *domain.ProviderConfig
				name string
				a    *domain.Appeal
			}{
				{
					pc: &domain.ProviderConfig{
						Credentials: validCredentials,
						Resources: []*domain.ResourceConfig{
							{
								Type:  tableau.ResourceTypeDataSource,
								Roles: validSiteRole,
							},
						},
						URN: providerURN,
					},
					name: "Provider Config to Update Site Role Permissions",
					a: &domain.Appeal{
						Resource: &domain.Resource{
							Type: tableau.ResourceTypeDataSource,
							URN:  "999",
							Name: "test-DataSource",
						},
						Role:        "test-role",
						Permissions: []string{"Creator@site_role"},
					},
				},
				{
					pc: &domain.ProviderConfig{
						Credentials: validCredentials,
						Resources: []*domain.ResourceConfig{
							{
								Type:  tableau.ResourceTypeDataSource,
								Roles: validRole,
							},
						},
						URN: providerURN,
					},
					name: "Provider Config with to Grant DataSource Permission",
					a: &domain.Appeal{
						Resource: &domain.Resource{
							Type: tableau.ResourceTypeDataSource,
							URN:  "999",
							Name: "test-DataSource",
						},
						Role:        "test-role",
						Permissions: []string{"ChangePermissions:Allow"},
					},
				},
			}

			for _, tc := range testcases {
				t.Run(tc.name, func(t *testing.T) {
					actualError := p.GrantAccess(tc.pc, tc.a)

					assert.EqualError(t, actualError, expectedError.Error())
				})
			}
		})

		t.Run("should return nil error if granting access is successful", func(t *testing.T) {
			providerURN := "test-provider-urn"
			crypto := new(mocks.Crypto)
			client := new(mocks.TableauClient)
			expectedDatasource := &tableau.DataSource{
				Name: "test-datasource",
				ID:   "datasource-id",
			}
			expectedUser := "test@email.com"
			expectedRole := tableau.PermissionNames[tableau.ResourceTypeDataSource][0] + ":" + tableau.PermissionModes[0]
			p := tableau.NewProvider("", crypto)
			p.Clients = map[string]tableau.TableauClient{
				providerURN: client,
			}
			client.On("GrantDataSourceAccess", expectedDatasource, expectedUser, expectedRole).Return(nil).Once()

			pc := &domain.ProviderConfig{
				Credentials: tableau.Credentials{
					Host:       "localhost",
					Username:   "test-username",
					Password:   "test-password",
					ContentURL: "test-content-url",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: tableau.ResourceTypeDataSource,
						Roles: []*domain.Role{
							{
								ID: "viewer",
								Permissions: []interface{}{
									tableau.Permission{
										Name: expectedRole,
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
					Type: tableau.ResourceTypeDataSource,
					URN:  "datasource-id",
					Name: "test-datasource",
				},
				Role:       "viewer",
				AccountID:  expectedUser,
				ResourceID: "99",
				ID:         "99",
			}

			actualError := p.GrantAccess(pc, a)

			assert.Nil(t, actualError)
		})
	})
}

func TestRevokeAccess(t *testing.T) {
	t.Run("should return error if credentials is invalid", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := tableau.NewProvider("", crypto)

		pc := &domain.ProviderConfig{
			Credentials: "invalid-credentials",
			Resources: []*domain.ResourceConfig{
				{
					Type: "test-type",
					Roles: []*domain.Role{
						{
							ID: "test-role",
							Permissions: []interface{}{
								tableau.Permission{
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
		assert.Error(t, actualError)
	})

	t.Run("should return error if there are any on client initialization", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := tableau.NewProvider("", crypto)
		expectedError := errors.New("decrypt error")
		crypto.On("Decrypt", "test-password").Return("", expectedError).Once()

		pc := &domain.ProviderConfig{
			Credentials: tableau.Credentials{
				Host:       "localhost",
				Username:   "test-username",
				Password:   "test-password",
				ContentURL: "test-content-url",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: "test-type",
					Roles: []*domain.Role{
						{
							ID: "test-role",
							Permissions: []interface{}{
								tableau.Permission{
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

	t.Run("should return error if resource type in unknown", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := tableau.NewProvider("", crypto)
		expectedError := errors.New("invalid resource type")
		crypto.On("Decrypt", "test-password").Return("", expectedError).Once()

		pc := &domain.ProviderConfig{
			Credentials: tableau.Credentials{
				Host:       "localhost",
				Username:   "test-username",
				Password:   "test-password",
				ContentURL: "test-content-url",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: "test-type",
					Roles: []*domain.Role{
						{
							ID: "test-role",
							Permissions: []interface{}{
								tableau.Permission{
									Name: "test-permission-config",
								},
							},
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

		actualError := p.RevokeAccess(pc, a)

		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("given workflow resource", func(t *testing.T) {
		t.Run("should return error if there is an error in revoking workflow access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			crypto := new(mocks.Crypto)
			client := new(mocks.TableauClient)
			p := tableau.NewProvider("", crypto)
			p.Clients = map[string]tableau.TableauClient{
				providerURN: client,
			}
			client.On("RevokeWorkbookAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

			pc := &domain.ProviderConfig{
				Credentials: tableau.Credentials{
					Host:       "localhost",
					Username:   "test-username",
					Password:   "test-password",
					ContentURL: "test-content-url",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: tableau.ResourceTypeWorkbook,
						Roles: []*domain.Role{
							{
								ID: "test-role",
								Permissions: []interface{}{
									tableau.Permission{
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
					Type: tableau.ResourceTypeWorkbook,
					URN:  "999",
					Name: "test-workbook",
				},
				Role:        "test-role",
				Permissions: []string{"test-permission-config"},
			}

			actualError := p.RevokeAccess(pc, a)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return nil error if revoking access is successful", func(t *testing.T) {
			providerURN := "test-provider-urn"
			crypto := new(mocks.Crypto)
			client := new(mocks.TableauClient)
			expectedWorkbook := &tableau.Workbook{
				Name: "test-workbook",
				ID:   "workbook-id",
			}
			expectedUser := "test@email.com"
			expectedRole := tableau.PermissionNames[tableau.ResourceTypeWorkbook][0] + ":" + tableau.PermissionModes[0]
			p := tableau.NewProvider("", crypto)
			p.Clients = map[string]tableau.TableauClient{
				providerURN: client,
			}
			client.On("RevokeWorkbookAccess", expectedWorkbook, expectedUser, expectedRole).Return(nil).Once()
			client.On("UpdateSiteRole", expectedUser, "Unlicensed").Return(nil).Once()
			pc := &domain.ProviderConfig{
				Credentials: tableau.Credentials{
					Host:       "localhost",
					Username:   "test-username",
					Password:   "test-password",
					ContentURL: "test-content-url",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: tableau.ResourceTypeWorkbook,
						Roles: []*domain.Role{
							{
								ID: "viewer",
								Permissions: []interface{}{
									tableau.Permission{
										Name: expectedRole,
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
					Type: tableau.ResourceTypeWorkbook,
					URN:  "workbook-id",
					Name: "test-workbook",
				},
				Role:       "viewer",
				AccountID:  expectedUser,
				ResourceID: "999",
				ID:         "999",
			}

			actualError := p.RevokeAccess(pc, a)

			assert.Nil(t, actualError)
		})
	})

	t.Run("given flow resource", func(t *testing.T) {
		t.Run("should return error if there is an error in revoking flow access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			crypto := new(mocks.Crypto)
			client := new(mocks.TableauClient)
			p := tableau.NewProvider("", crypto)
			p.Clients = map[string]tableau.TableauClient{
				providerURN: client,
			}
			client.On("RevokeFlowAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

			pc := &domain.ProviderConfig{
				Credentials: tableau.Credentials{
					Host:       "localhost",
					Username:   "test-username",
					Password:   "test-password",
					ContentURL: "test-content-url",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: tableau.ResourceTypeFlow,
						Roles: []*domain.Role{
							{
								ID: "test-role",
								Permissions: []interface{}{
									tableau.Permission{
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
					Type: tableau.ResourceTypeFlow,
					URN:  "999",
					Name: "test-flow",
				},
				Role:        "test-role",
				Permissions: []string{"test-permission-config"},
			}

			actualError := p.RevokeAccess(pc, a)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return nil error if revoking access is successful", func(t *testing.T) {
			providerURN := "test-provider-urn"
			crypto := new(mocks.Crypto)
			client := new(mocks.TableauClient)
			expectedFlow := &tableau.Flow{
				Name: "test-flow",
				ID:   "flow-id",
			}
			expectedUser := "test@email.com"
			expectedRole := tableau.PermissionNames[tableau.ResourceTypeFlow][0] + ":" + tableau.PermissionModes[0]
			p := tableau.NewProvider("", crypto)
			p.Clients = map[string]tableau.TableauClient{
				providerURN: client,
			}
			client.On("RevokeFlowAccess", expectedFlow, expectedUser, expectedRole).Return(nil).Once()
			client.On("UpdateSiteRole", expectedUser, "Unlicensed").Return(nil).Once()

			pc := &domain.ProviderConfig{
				Credentials: tableau.Credentials{
					Host:       "localhost",
					Username:   "test-username",
					Password:   "test-password",
					ContentURL: "test-content-url",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: tableau.ResourceTypeFlow,
						Roles: []*domain.Role{
							{
								ID: "viewer",
								Permissions: []interface{}{
									tableau.Permission{
										Name: expectedRole,
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
					Type: tableau.ResourceTypeFlow,
					URN:  "flow-id",
					Name: "test-flow",
				},
				Role:       "viewer",
				AccountID:  expectedUser,
				ResourceID: "999",
				ID:         "999",
			}

			actualError := p.RevokeAccess(pc, a)

			assert.Nil(t, actualError)
		})
	})

	t.Run("given view resource", func(t *testing.T) {
		t.Run("should return error if there is an error in revoking view access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			crypto := new(mocks.Crypto)
			client := new(mocks.TableauClient)
			p := tableau.NewProvider("", crypto)
			p.Clients = map[string]tableau.TableauClient{
				providerURN: client,
			}
			client.On("RevokeViewAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

			pc := &domain.ProviderConfig{
				Credentials: tableau.Credentials{
					Host:       "localhost",
					Username:   "test-username",
					Password:   "test-password",
					ContentURL: "test-content-url",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: tableau.ResourceTypeView,
						Roles: []*domain.Role{
							{
								ID: "test-role",
								Permissions: []interface{}{
									tableau.Permission{
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
					Type: tableau.ResourceTypeView,
					URN:  "99",
					Name: "test-view",
				},
				Role:        "test-role",
				Permissions: []string{"test-permission-config"},
			}

			actualError := p.RevokeAccess(pc, a)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return nil error if revoking access is successful", func(t *testing.T) {
			providerURN := "test-provider-urn"
			crypto := new(mocks.Crypto)
			client := new(mocks.TableauClient)
			expectedView := &tableau.View{
				Name: "test-view",
				ID:   "view-id",
			}
			expectedUser := "test@email.com"
			expectedRole := tableau.PermissionNames[tableau.ResourceTypeView][0] + ":" + tableau.PermissionModes[0]
			p := tableau.NewProvider("", crypto)
			p.Clients = map[string]tableau.TableauClient{
				providerURN: client,
			}
			client.On("RevokeViewAccess", expectedView, expectedUser, expectedRole).Return(nil).Once()
			client.On("UpdateSiteRole", expectedUser, "Unlicensed").Return(nil).Once()

			pc := &domain.ProviderConfig{
				Credentials: tableau.Credentials{
					Host:       "localhost",
					Username:   "test-username",
					Password:   "test-password",
					ContentURL: "test-content-url",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: tableau.ResourceTypeView,
						Roles: []*domain.Role{
							{
								ID: "viewer",
								Permissions: []interface{}{
									tableau.Permission{
										Name: expectedRole,
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
					Type: tableau.ResourceTypeView,
					URN:  "view-id",
					Name: "test-view",
				},
				Role:       "viewer",
				AccountID:  expectedUser,
				ResourceID: "99",
				ID:         "99",
			}

			actualError := p.RevokeAccess(pc, a)

			assert.Nil(t, actualError)
		})
	})

	t.Run("given metric resource", func(t *testing.T) {
		t.Run("should return error if there is an error in revoking metric access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			crypto := new(mocks.Crypto)
			client := new(mocks.TableauClient)
			p := tableau.NewProvider("", crypto)
			p.Clients = map[string]tableau.TableauClient{
				providerURN: client,
			}
			client.On("RevokeMetricAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

			pc := &domain.ProviderConfig{
				Credentials: tableau.Credentials{
					Host:       "localhost",
					Username:   "test-username",
					Password:   "test-password",
					ContentURL: "test-content-url",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: tableau.ResourceTypeMetric,
						Roles: []*domain.Role{
							{
								ID: "test-role",
								Permissions: []interface{}{
									tableau.Permission{
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
					Type: tableau.ResourceTypeMetric,
					URN:  "99",
					Name: "test-metric",
				},
				Role:        "test-role",
				Permissions: []string{"test-permission-config"},
			}

			actualError := p.RevokeAccess(pc, a)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return nil error if revoking access is successful", func(t *testing.T) {
			providerURN := "test-provider-urn"
			crypto := new(mocks.Crypto)
			client := new(mocks.TableauClient)
			expectedMetric := &tableau.Metric{
				Name: "test-metric",
				ID:   "metric-id",
			}
			expectedUser := "test@email.com"
			expectedRole := tableau.PermissionNames[tableau.ResourceTypeMetric][0] + ":" + tableau.PermissionModes[0]
			p := tableau.NewProvider("", crypto)
			p.Clients = map[string]tableau.TableauClient{
				providerURN: client,
			}
			client.On("RevokeMetricAccess", expectedMetric, expectedUser, expectedRole).Return(nil).Once()
			client.On("UpdateSiteRole", expectedUser, "Unlicensed").Return(nil).Once()

			pc := &domain.ProviderConfig{
				Credentials: tableau.Credentials{
					Host:       "localhost",
					Username:   "test-username",
					Password:   "test-password",
					ContentURL: "test-content-url",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: tableau.ResourceTypeMetric,
						Roles: []*domain.Role{
							{
								ID: "viewer",
								Permissions: []interface{}{
									tableau.Permission{
										Name: expectedRole,
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
					Type: tableau.ResourceTypeMetric,
					URN:  "metric-id",
					Name: "test-metric",
				},
				Role:       "viewer",
				AccountID:  expectedUser,
				ResourceID: "99",
				ID:         "99",
			}

			actualError := p.RevokeAccess(pc, a)

			assert.Nil(t, actualError)
		})
	})

	t.Run("given datasource resource", func(t *testing.T) {
		t.Run("should return error if there is an error in revoking datasource access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			crypto := new(mocks.Crypto)
			client := new(mocks.TableauClient)
			p := tableau.NewProvider("", crypto)
			p.Clients = map[string]tableau.TableauClient{
				providerURN: client,
			}
			client.On("RevokeDataSourceAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

			pc := &domain.ProviderConfig{
				Credentials: tableau.Credentials{
					Host:       "localhost",
					Username:   "test-username",
					Password:   "test-password",
					ContentURL: "test-content-url",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: tableau.ResourceTypeDataSource,
						Roles: []*domain.Role{
							{
								ID: "test-role",
								Permissions: []interface{}{
									tableau.Permission{
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
					Type: tableau.ResourceTypeDataSource,
					URN:  "99",
					Name: "test-datasource",
				},
				Role:        "test-role",
				Permissions: []string{"test-permission-config"},
			}

			actualError := p.RevokeAccess(pc, a)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return nil error if revoking access is successful", func(t *testing.T) {
			providerURN := "test-provider-urn"
			crypto := new(mocks.Crypto)
			client := new(mocks.TableauClient)
			expectedDatasource := &tableau.DataSource{
				Name: "test-datasource",
				ID:   "datasource-id",
			}
			expectedUser := "test@email.com"
			expectedRole := tableau.PermissionNames[tableau.ResourceTypeDataSource][0] + ":" + tableau.PermissionModes[0]
			p := tableau.NewProvider("", crypto)
			p.Clients = map[string]tableau.TableauClient{
				providerURN: client,
			}
			client.On("RevokeDataSourceAccess", expectedDatasource, expectedUser, expectedRole).Return(nil).Once()
			client.On("UpdateSiteRole", expectedUser, "Unlicensed").Return(nil).Once()

			pc := &domain.ProviderConfig{
				Credentials: tableau.Credentials{
					Host:       "localhost",
					Username:   "test-username",
					Password:   "test-password",
					ContentURL: "test-content-url",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: tableau.ResourceTypeDataSource,
						Roles: []*domain.Role{
							{
								ID: "viewer",
								Permissions: []interface{}{
									tableau.Permission{
										Name: expectedRole,
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
					Type: tableau.ResourceTypeDataSource,
					URN:  "datasource-id",
					Name: "test-datasource",
				},
				Role:       "viewer",
				AccountID:  expectedUser,
				ResourceID: "99",
				ID:         "99",
			}

			actualError := p.RevokeAccess(pc, a)

			assert.Nil(t, actualError)
		})
	})
}

func TestGetAccountTypes(t *testing.T) {
	t.Run("should return the valid Account Types \"user\"", func(t *testing.T) {
		expectedAccountTypes := []string{"user"}
		crypto := new(mocks.Crypto)
		p := tableau.NewProvider("", crypto)

		actualAccountTypes := p.GetAccountTypes()

		assert.Equal(t, expectedAccountTypes, actualAccountTypes)
	})
}

func TestGetRoles(t *testing.T) {
	t.Run("should return error if resource type is invalid", func(t *testing.T) {

	})

	t.Run("should return roles specified in the provider config", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := tableau.NewProvider("", crypto)
		expectedRoles := []*domain.Role{
			{
				ID:   "read",
				Name: "Read",
				Permissions: []interface{}{
					map[string]interface{}{
						"name": "Read:Allow",
					},
				},
			},
		}

		validConfig := &domain.ProviderConfig{
			Type: "tableau",
			Credentials: tableau.Credentials{
				Host:       "http://localhost",
				Username:   "test-username",
				Password:   "test-password",
				ContentURL: "test-content-url",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type:  tableau.ResourceTypeWorkbook,
					Roles: expectedRoles,
				},
			},
			URN: "test-URN",
		}

		actualRoles, actualError := p.GetRoles(validConfig, "workbook")

		assert.NoError(t, actualError)
		assert.Equal(t, expectedRoles, actualRoles)
	})
}
