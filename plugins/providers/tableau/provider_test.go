package tableau_test

import (
	"errors"
	"testing"

	"github.com/mitchellh/mapstructure"
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

	t.Run("should return error if there are any on client initialization", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := tableau.NewProvider("", crypto)

		expectedError := errors.New("decrypt error")
		crypto.On("Decrypt", "test-password").Return("", expectedError).Once()
		pc := &domain.ProviderConfig{
			Credentials: map[string]interface{}{
				"password":    "test-password",
				"host":        "http://localhost",
				"content_url": "test-content-url",
				"username":    "username",
			},
		}

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.EqualError(t, actualError, expectedError.Error())
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
	})
}

func TestGrantAccess(t *testing.T) {
	t.Run("should return an error if there is an error in getting permissions", func(t *testing.T) {
		var permission tableau.Permission
		invalidPermissionConfig := "invalid-permisiion-config"
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
				expectedError: tableau.ErrInvalidResourceType,
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
				expectedError: tableau.ErrInvalidRole,
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
			p := tableau.NewProvider("", crypto)

			providerConfig := &domain.ProviderConfig{
				Resources: tc.resourceConfigs,
			}

			actualError := p.GrantAccess(providerConfig, tc.appeal)
			assert.EqualError(t, actualError, tc.expectedError.Error())
		}
	})

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

	t.Run("given workflow resource", func(t *testing.T) {
		t.Run("should return error if there is an error in granting workflow access", func(t *testing.T) {
			providerURN := "test-provider-urn"
			expectedError := errors.New("client error")
			crypto := new(mocks.Crypto)
			client := new(mocks.TableauClient)
			p := tableau.NewProvider("", crypto)
			p.Clients = map[string]tableau.TableauClient{
				providerURN: client,
			}
			client.On("GrantWorkbookAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

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
				Role: "test-role",
			}

			actualError := p.GrantAccess(pc, a)

			assert.EqualError(t, actualError, expectedError.Error())
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
				ResourceID: 999,
				ID:         999,
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
			client.On("GrantFlowAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

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
				Role: "test-role",
			}

			actualError := p.GrantAccess(pc, a)

			assert.EqualError(t, actualError, expectedError.Error())
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
				ResourceID: 999,
				ID:         999,
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
			client.On("GrantViewAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

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
				Role: "test-role",
			}

			actualError := p.GrantAccess(pc, a)

			assert.EqualError(t, actualError, expectedError.Error())
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
				ResourceID: 99,
				ID:         99,
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
			client.On("GrantMetricAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

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
				Role: "test-role",
			}

			actualError := p.GrantAccess(pc, a)

			assert.EqualError(t, actualError, expectedError.Error())
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
				ResourceID: 99,
				ID:         99,
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
			client.On("GrantDataSourceAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

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
				Role: "test-role",
			}

			actualError := p.GrantAccess(pc, a)

			assert.EqualError(t, actualError, expectedError.Error())
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
				ResourceID: 99,
				ID:         99,
			}

			actualError := p.GrantAccess(pc, a)

			assert.Nil(t, actualError)
		})
	})
}
