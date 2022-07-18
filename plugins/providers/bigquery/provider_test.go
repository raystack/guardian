package bigquery_test

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/plugins/providers/bigquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetType(t *testing.T) {
	t.Run("should return provider type name", func(t *testing.T) {
		p := initProvider()
		expectedTypeName := domain.ProviderTypeBigQuery

		actualTypeName := p.GetType()

		assert.Equal(t, expectedTypeName, actualTypeName)
	})
}

func TestCreateConfig(t *testing.T) {

	t.Run("should return error if error in credentials are invalid/mandatory fields are missing", func(t *testing.T) {
		//providerURN := "test-URN"
		crypto := new(mocks.Crypto)
		client := new(mocks.BigQueryClient)
		p := bigquery.NewProvider("", crypto)
		p.Clients = map[string]bigquery.BigQueryClient{
			"resource-name": client,
		}

		testcases := []struct {
			pc   *domain.ProviderConfig
			name string
		}{
			{
				name: "invalid credentials struct",
				pc: &domain.ProviderConfig{
					Credentials: "invalid-credential-structure",
					Resources: []*domain.ResourceConfig{
						{
							Type: bigquery.ResourceTypeDataset,
						},
					},
				},
			},
			{
				name: "empty mandatory credentials",
				pc: &domain.ProviderConfig{
					Credentials: bigquery.Credentials{
						ServiceAccountKey: "",
						ResourceName:      "",
					},
					Resources: []*domain.ResourceConfig{
						{
							Type: bigquery.ResourceTypeDataset,
						},
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
		providerURN := "test-URN"
		crypto := new(mocks.Crypto)
		client := new(mocks.BigQueryClient)
		p := bigquery.NewProvider("", crypto)
		p.Clients = map[string]bigquery.BigQueryClient{
			providerURN: client,
		}

		testcases := []struct {
			pc *domain.ProviderConfig
		}{
			{
				pc: &domain.ProviderConfig{
					Credentials: bigquery.Credentials{
						ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service-account-key-json")),
						ResourceName:      "projects/test-resource-name",
					},
				},
			},
			{
				pc: &domain.ProviderConfig{
					Credentials: bigquery.Credentials{
						ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service-account-key-json")),
						ResourceName:      "projects/test-resource-name",
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
					Credentials: bigquery.Credentials{
						ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service-account-key-json")),
						ResourceName:      "projects/test-resource-name",
					},
					Resources: []*domain.ResourceConfig{
						{
							Type: bigquery.ResourceTypeDataset,
							Roles: []*domain.Role{
								{
									ID:          "viewer",
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

	// t.Run("should return error if error in encrypting the credentials", func(t *testing.T) {
	// 	providerURN := "test-URN"
	// 	crypto := new(mocks.Crypto)
	// 	client := new(mocks.BigQueryClient)
	// 	p := bigquery.NewProvider("", crypto)
	// 	p.Clients = map[string]bigquery.BigQueryClient{
	// 		"test-resource-name": client,
	// 	}
	// 	expectedError := errors.New("error in encrypting SAK")
	// 	crypto.On("Encrypt", "service-account-key-json").Return("", expectedError)
	// 	pc := &domain.ProviderConfig{
	// 		Resources: []*domain.ResourceConfig{
	// 			{
	// 				Type:  bigquery.ResourceTypeDataset,
	// 				Roles: []*domain.Role{},
	// 			},
	// 		},
	// 		//	base64.StdEncoding.EncodeToString([]byte("service-account-key-json"))
	// 		Credentials: bigquery.Credentials{
	// 			ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service-account-key-json")),
	// 			ResourceName:      "projects/test-resource-name",
	// 		},
	// 		URN: providerURN,
	// 	}

	// 	actualError := p.CreateConfig(pc)

	// 	assert.Equal(t, expectedError, actualError)
	// })

	// t.Run("should return nil error and create the config on success", func(t *testing.T) {
	// 	providerURN := "test-URN"
	// 	crypto := new(mocks.Crypto)
	// 	client := new(mocks.BigQueryClient)
	// 	p := bigquery.NewProvider("", crypto)
	// 	p.Clients = map[string]bigquery.BigQueryClient{
	// 		"test-resource-name": client,
	// 	}
	// 	crypto.On("Encrypt", "service-account-key-json").Return("service-account-key-json", nil)
	// 	pc := &domain.ProviderConfig{
	// 		Resources: []*domain.ResourceConfig{
	// 			{
	// 				Type:  bigquery.ResourceTypeDataset,
	// 				Roles: []*domain.Role{},
	// 			},
	// 		},
	// 		//	base64.StdEncoding.EncodeToString([]byte("service-account-key-json"))
	// 		Credentials: bigquery.Credentials{
	// 			ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service-account-key-json")),
	// 			ResourceName:      "projects/test-resource-name",
	// 		},
	// 		URN: providerURN,
	// 	}

	// 	actualError := p.CreateConfig(pc)

	// 	assert.NoError(t, actualError)
	// })
}

func TestGetResources(t *testing.T) {
	t.Run("should error when credentials are invalid", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := bigquery.NewProvider("", crypto)
		pc := &domain.ProviderConfig{
			Type:        domain.ProviderTypeBigQuery,
			URN:         "test-project-id",
			Credentials: "invalid-creds",
		}

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.Error(t, actualError)
	})

	t.Run("should return dataset resource object", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		client := new(mocks.BigQueryClient)
		p := bigquery.NewProvider("", crypto)
		p.Clients = map[string]bigquery.BigQueryClient{
			"resource-name": client,
		}
		validCredentials := bigquery.Credentials{
			ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service-account-key-json")),
			ResourceName:      "projects/resource-name",
		}
		pc := &domain.ProviderConfig{
			Type:        domain.ProviderTypeBigQuery,
			URN:         "test-project-id",
			Credentials: validCredentials,
			Resources: []*domain.ResourceConfig{
				{
					Type: "dataset",
					Roles: []*domain.Role{
						{
							ID:          "VIEWER",
							Name:        "VIEWER",
							Permissions: []interface{}{"VIEWER"},
						},
					},
				},
				{
					Type: "table",
					Roles: []*domain.Role{
						{
							ID:          "VIEWER",
							Name:        "VIEWER",
							Permissions: []interface{}{"VIEWER"},
						},
					},
				},
			},
		}
		expectedDatasets := []*bigquery.Dataset{
			{
				ProjectID: "p_id",
				DatasetID: "d_id",
			},
		}
		expectedTables := []*bigquery.Table{
			{
				ProjectID: "p_id",
				DatasetID: "d_id",
				TableID:   "t_id",
			},
		}

		client.On("GetDatasets", mock.Anything).Return(expectedDatasets, nil).Once()
		client.On("GetTables", mock.Anything, mock.Anything).Return(expectedTables, nil).Once()
		expectedResources := []*domain.Resource{
			{
				ProviderType: domain.ProviderTypeBigQuery,
				ProviderURN:  "test-project-id",
				Type:         "dataset",
				Name:         "d_id",
				URN:          "p_id:d_id",
			},
			{
				ProviderType: domain.ProviderTypeBigQuery,
				ProviderURN:  "test-project-id",
				Name:         "t_id",
				URN:          "p_id:d_id.t_id",
				Type:         "table",
			},
		}
		actualResources, actualError := p.GetResources(pc)

		assert.Equal(t, expectedResources, actualResources)
		assert.Nil(t, actualError)
	})
}

func TestGrantAccess(t *testing.T) {

	t.Run("should return error if Provider Config or Appeal doesn't have required paramters", func(t *testing.T) {
		testCases := []struct {
			providerConfig *domain.ProviderConfig
			appeal         *domain.Appeal
			expectedError  error
		}{
			{
				providerConfig: nil,
				expectedError:  bigquery.ErrNilProviderConfig,
			},
			{
				providerConfig: &domain.ProviderConfig{
					Type:                "bigquery",
					URN:                 "test-URN",
					AllowedAccountTypes: []string{"user", "serviceAccount"},
				},
				appeal:        nil,
				expectedError: bigquery.ErrNilAppeal,
			},
			{
				providerConfig: &domain.ProviderConfig{
					Type:                "bigquery",
					URN:                 "test-URN",
					AllowedAccountTypes: []string{"user", "serviceAccount"},
				},
				appeal: &domain.Appeal{
					ID:          "test-appeal-id",
					AccountType: "user",
				},
				expectedError: bigquery.ErrNilResource,
			},
			{
				providerConfig: &domain.ProviderConfig{
					Type:                "bigquery",
					URN:                 "test-URN-1",
					AllowedAccountTypes: []string{"user", "serviceAccount"},
				},
				appeal: &domain.Appeal{
					ID:          "test-appeal-id",
					AccountType: "user",
					Resource: &domain.Resource{
						ID:           "test-resource-id",
						ProviderType: "metabase",
					},
				},
				expectedError: bigquery.ErrProviderTypeMismatch,
			},
			{
				providerConfig: &domain.ProviderConfig{
					Type:                "bigquery",
					URN:                 "test-URN-1",
					AllowedAccountTypes: []string{"user", "serviceAccount"},
				},
				appeal: &domain.Appeal{
					ID:          "test-appeal-id",
					AccountType: "user",
					Resource: &domain.Resource{
						ID:           "test-resource-id",
						ProviderType: "bigquery",
						ProviderURN:  "not-test-URN-1",
					},
				},
				expectedError: bigquery.ErrProviderURNMismatch,
			},
		}

		for _, tc := range testCases {

			p := initProvider()
			pc := tc.providerConfig
			a := tc.appeal

			actualError := p.GrantAccess(pc, a)
			assert.EqualError(t, actualError, tc.expectedError.Error())
		}
	})

	t.Run("should return an error if there is an error in getting permissions", func(t *testing.T) {
		var permission bigquery.Permission
		invalidPermissionConfig := map[string]interface{}{}
		invalidPermissionConfigError := mapstructure.Decode(invalidPermissionConfig, &permission)

		testCases := []struct {
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
				expectedError: bigquery.ErrInvalidResourceType,
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
				expectedError: bigquery.ErrInvalidRole,
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

		for _, tc := range testCases {
			crypto := new(mocks.Crypto)
			p := bigquery.NewProvider("", crypto)

			providerConfig := &domain.ProviderConfig{
				Resources: tc.resourceConfigs,
			}

			actualError := p.GrantAccess(providerConfig, tc.appeal)
			assert.EqualError(t, actualError, tc.expectedError.Error())
		}
	},
	)

	t.Run("should return error if credentials is invalid", func(t *testing.T) {
		p := initProvider()

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

	t.Run("should return error if GrantDataset Access returns an error", func(t *testing.T) {
		//	providerURN := "test-URN"
		expectedError := errors.New("Test-Error")
		expectedAccountType := "user"
		expectedAccountID := "test@email.com"
		crypto := new(mocks.Crypto)
		client := new(mocks.BigQueryClient)
		p := bigquery.NewProvider("bigquery", crypto)
		p.Clients = map[string]bigquery.BigQueryClient{
			"resource-name": client,
		}
		validCredentials := bigquery.Credentials{
			ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service-account-key-json")),
			ResourceName:      "projects/resource-name",
		}
		client.On("GrantDatasetAccess", mock.Anything, mock.Anything, expectedAccountID, mock.Anything).Return(expectedError).Once()

		pc := &domain.ProviderConfig{
			Type:        "bigquery",
			URN:         "p_id:d_id",
			Credentials: validCredentials,
			Resources: []*domain.ResourceConfig{
				{
					Type: "dataset",
					Roles: []*domain.Role{
						{
							ID:          "VIEWER",
							Name:        "VIEWER",
							Permissions: []interface{}{"VIEWER"},
						},
					},
				},
			},
		}
		a := &domain.Appeal{
			Role: "VIEWER",
			Resource: &domain.Resource{
				ProviderType: "bigquery",
				ProviderURN:  "p_id:d_id",
				Type:         "dataset",
			},
			ID:          "999",
			ResourceID:  "999",
			AccountType: expectedAccountType,
			AccountID:   expectedAccountID,
		}

		actualError := p.GrantAccess(pc, a)

		assert.Equal(t, expectedError, actualError)
	})

	t.Run("should grant access to dataset resource and return no error on success", func(t *testing.T) {
		//	providerURN := "test-URN"
		expectedAccountType := "user"
		expectedAccountID := "test@email.com"
		crypto := new(mocks.Crypto)
		client := new(mocks.BigQueryClient)
		p := bigquery.NewProvider("bigquery", crypto)
		p.Clients = map[string]bigquery.BigQueryClient{
			"resource-name": client,
		}
		validCredentials := bigquery.Credentials{
			ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service-account-key-json")),
			ResourceName:      "projects/resource-name",
		}
		client.On("GrantDatasetAccess", mock.Anything, mock.Anything, expectedAccountID, mock.Anything).Return(bigquery.ErrPermissionAlreadyExists).Once()

		pc := &domain.ProviderConfig{
			Type:        "bigquery",
			URN:         "p_id:d_id",
			Credentials: validCredentials,
			Resources: []*domain.ResourceConfig{
				{
					Type: "dataset",
					Roles: []*domain.Role{
						{
							ID:          "VIEWER",
							Name:        "VIEWER",
							Permissions: []interface{}{"VIEWER"},
						},
					},
				},
			},
		}
		a := &domain.Appeal{
			Role: "VIEWER",
			Resource: &domain.Resource{
				ProviderType: "bigquery",
				ProviderURN:  "p_id:d_id",
				Type:         "dataset",
			},
			ID:          "999",
			ResourceID:  "999",
			AccountType: expectedAccountType,
			AccountID:   expectedAccountID,
		}

		actualError := p.GrantAccess(pc, a)

		assert.Nil(t, actualError)
	})

	t.Run("should grant access to table resource and return no error on success", func(t *testing.T) {
		providerURN := "test-URN"
		expectedAccountType := "user"
		expectedAccountID := "test@email.com"
		crypto := new(mocks.Crypto)
		client := new(mocks.BigQueryClient)
		p := bigquery.NewProvider("bigquery", crypto)
		p.Clients = map[string]bigquery.BigQueryClient{
			"resource-name": client,
		}
		validCredentials := bigquery.Credentials{
			ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service-account-key-json")),
			ResourceName:      "projects/resource-name",
		}
		client.On("GrantTableAccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(bigquery.ErrPermissionAlreadyExists).Once()

		pc := &domain.ProviderConfig{
			Type:        "bigquery",
			URN:         providerURN,
			Credentials: validCredentials,
			Resources: []*domain.ResourceConfig{
				{
					Type: "table",
					Roles: []*domain.Role{
						{
							ID:          "VIEWER",
							Name:        "VIEWER",
							Permissions: []interface{}{"VIEWER"},
						},
					},
				},
			},
		}
		a := &domain.Appeal{
			Role: "VIEWER",
			Resource: &domain.Resource{
				URN:          "p_id:d_id.t_id",
				Name:         "t_id",
				ProviderType: "bigquery",
				ProviderURN:  "test-URN",
				Type:         "table",
			},
			ID:          "999",
			ResourceID:  "999",
			AccountType: expectedAccountType,
			AccountID:   expectedAccountID,
		}

		actualError := p.GrantAccess(pc, a)

		assert.Nil(t, actualError)
	})
}

func TestRevokeAccess(t *testing.T) {

	t.Run("should return error if Provider Config or Appeal doesn't have required paramters", func(t *testing.T) {
		testCases := []struct {
			providerConfig *domain.ProviderConfig
			appeal         *domain.Appeal
			expectedError  error
		}{
			{
				providerConfig: nil,
				expectedError:  bigquery.ErrNilProviderConfig,
			},
			{
				providerConfig: &domain.ProviderConfig{
					Type:                "bigquery",
					URN:                 "test-URN",
					AllowedAccountTypes: []string{"user", "serviceAccount"},
				},
				appeal:        nil,
				expectedError: bigquery.ErrNilAppeal,
			},
			{
				providerConfig: &domain.ProviderConfig{
					Type:                "bigquery",
					URN:                 "test-URN",
					AllowedAccountTypes: []string{"user", "serviceAccount"},
				},
				appeal: &domain.Appeal{
					ID:          "test-appeal-id",
					AccountType: "user",
				},
				expectedError: bigquery.ErrNilResource,
			},
			{
				providerConfig: &domain.ProviderConfig{
					Type:                "bigquery",
					URN:                 "test-URN-1",
					AllowedAccountTypes: []string{"user", "serviceAccount"},
				},
				appeal: &domain.Appeal{
					ID:          "test-appeal-id",
					AccountType: "user",
					Resource: &domain.Resource{
						ID:           "test-resource-id",
						ProviderType: "metabase",
					},
				},
				expectedError: bigquery.ErrProviderTypeMismatch,
			},
			{
				providerConfig: &domain.ProviderConfig{
					Type:                "bigquery",
					URN:                 "test-URN-1",
					AllowedAccountTypes: []string{"user", "serviceAccount"},
				},
				appeal: &domain.Appeal{
					ID:          "test-appeal-id",
					AccountType: "user",
					Resource: &domain.Resource{
						ID:           "test-resource-id",
						ProviderType: "bigquery",
						ProviderURN:  "not-test-URN-1",
					},
				},
				expectedError: bigquery.ErrProviderURNMismatch,
			},
		}

		for _, tc := range testCases {

			p := initProvider()
			pc := tc.providerConfig
			a := tc.appeal

			actualError := p.RevokeAccess(pc, a)
			assert.EqualError(t, actualError, tc.expectedError.Error())
		}
	})

	t.Run("should return an error if there is an error in getting permissions", func(t *testing.T) {
		var permission bigquery.Permission
		invalidPermissionConfig := map[string]interface{}{}
		invalidPermissionConfigError := mapstructure.Decode(invalidPermissionConfig, &permission)

		testCases := []struct {
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
				expectedError: bigquery.ErrInvalidResourceType,
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
				expectedError: bigquery.ErrInvalidRole,
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

		for _, tc := range testCases {
			crypto := new(mocks.Crypto)
			p := bigquery.NewProvider("", crypto)

			providerConfig := &domain.ProviderConfig{
				Resources: tc.resourceConfigs,
			}

			actualError := p.RevokeAccess(providerConfig, tc.appeal)
			assert.EqualError(t, actualError, tc.expectedError.Error())
		}
	},
	)

	t.Run("should return error if credentials is invalid", func(t *testing.T) {
		p := initProvider()

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

		actualError := p.RevokeAccess(pc, a)
		assert.Error(t, actualError)
	})

	t.Run("should return error if Revoke Dataset Access returns an error", func(t *testing.T) {
		//	providerURN := "test-URN"
		expectedError := errors.New("Test-Error")
		expectedAccountType := "user"
		expectedAccountID := "test@email.com"
		crypto := new(mocks.Crypto)
		client := new(mocks.BigQueryClient)
		p := bigquery.NewProvider("bigquery", crypto)
		p.Clients = map[string]bigquery.BigQueryClient{
			"resource-name": client,
		}
		validCredentials := bigquery.Credentials{
			ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service-account-key-json")),
			ResourceName:      "projects/resource-name",
		}
		client.On("RevokeDatasetAccess", mock.Anything, mock.Anything, expectedAccountID, mock.Anything).Return(expectedError).Once()

		pc := &domain.ProviderConfig{
			Type:        "bigquery",
			URN:         "p_id:d_id",
			Credentials: validCredentials,
			Resources: []*domain.ResourceConfig{
				{
					Type: "dataset",
					Roles: []*domain.Role{
						{
							ID:          "VIEWER",
							Name:        "VIEWER",
							Permissions: []interface{}{"VIEWER"},
						},
					},
				},
			},
		}
		a := &domain.Appeal{
			Role: "VIEWER",
			Resource: &domain.Resource{
				ProviderType: "bigquery",
				ProviderURN:  "p_id:d_id",
				Type:         "dataset",
			},
			ID:          "999",
			ResourceID:  "999",
			AccountType: expectedAccountType,
			AccountID:   expectedAccountID,
		}

		actualError := p.RevokeAccess(pc, a)

		assert.Equal(t, expectedError, actualError)
	})

	t.Run("should Revoke access to dataset resource and return no error on success", func(t *testing.T) {
		//	providerURN := "test-URN"
		expectedAccountType := "user"
		expectedAccountID := "test@email.com"
		crypto := new(mocks.Crypto)
		client := new(mocks.BigQueryClient)
		p := bigquery.NewProvider("bigquery", crypto)
		p.Clients = map[string]bigquery.BigQueryClient{
			"resource-name": client,
		}
		validCredentials := bigquery.Credentials{
			ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service-account-key-json")),
			ResourceName:      "projects/resource-name",
		}
		client.On("RevokeDatasetAccess", mock.Anything, mock.Anything, expectedAccountID, mock.Anything).Return(bigquery.ErrPermissionNotFound).Once()

		pc := &domain.ProviderConfig{
			Type:        "bigquery",
			URN:         "p_id:d_id",
			Credentials: validCredentials,
			Resources: []*domain.ResourceConfig{
				{
					Type: "dataset",
					Roles: []*domain.Role{
						{
							ID:          "VIEWER",
							Name:        "VIEWER",
							Permissions: []interface{}{"VIEWER"},
						},
					},
				},
			},
		}
		a := &domain.Appeal{
			Role: "VIEWER",
			Resource: &domain.Resource{
				ProviderType: "bigquery",
				ProviderURN:  "p_id:d_id",
				Type:         "dataset",
			},
			ID:          "999",
			ResourceID:  "999",
			AccountType: expectedAccountType,
			AccountID:   expectedAccountID,
		}

		actualError := p.RevokeAccess(pc, a)

		assert.Nil(t, actualError)
	})

	t.Run("should Revoke access to table resource and return no error on success", func(t *testing.T) {
		providerURN := "test-URN"
		expectedAccountType := "user"
		expectedAccountID := "test@email.com"
		crypto := new(mocks.Crypto)
		client := new(mocks.BigQueryClient)
		p := bigquery.NewProvider("bigquery", crypto)
		p.Clients = map[string]bigquery.BigQueryClient{
			"resource-name": client,
		}
		validCredentials := bigquery.Credentials{
			ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service-account-key-json")),
			ResourceName:      "projects/resource-name",
		}
		client.On("RevokeTableAccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(bigquery.ErrPermissionNotFound).Once()

		pc := &domain.ProviderConfig{
			Type:        "bigquery",
			URN:         providerURN,
			Credentials: validCredentials,
			Resources: []*domain.ResourceConfig{
				{
					Type: "table",
					Roles: []*domain.Role{
						{
							ID:          "VIEWER",
							Name:        "VIEWER",
							Permissions: []interface{}{"VIEWER"},
						},
					},
				},
			},
		}
		a := &domain.Appeal{
			Role: "VIEWER",
			Resource: &domain.Resource{
				URN:          "p_id:d_id.t_id",
				Name:         "t_id",
				ProviderType: "bigquery",
				ProviderURN:  "test-URN",
				Type:         "table",
			},
			ID:          "999",
			ResourceID:  "999",
			AccountType: expectedAccountType,
			AccountID:   expectedAccountID,
		}

		actualError := p.RevokeAccess(pc, a)

		assert.Nil(t, actualError)
	})
}

func TestGetAccountTypes(t *testing.T) {
	t.Run("should return the supported account types \"user\" and \"serviceAccount\"", func(t *testing.T) {
		p := initProvider()
		expectedAccountTypes := []string{"user", "serviceAccount"}

		actualAccountTypes := p.GetAccountTypes()

		assert.Equal(t, expectedAccountTypes, actualAccountTypes)
	})
}

func TestGetRoles(t *testing.T) {
	t.Run("should return error if resource type is invalid", func(t *testing.T) {
		p := initProvider()

		validConfig := &domain.ProviderConfig{
			Type:                "bigquery",
			URN:                 "test-URN",
			AllowedAccountTypes: []string{"user", "serviceAccount"},
			Credentials:         map[string]interface{}{},
			Resources: []*domain.ResourceConfig{
				{
					Type: "dataset",
					Policy: &domain.PolicyConfig{
						ID:      "test-policy-1",
						Version: 1,
					},
				},
				{
					Type: "table",
					Policy: &domain.PolicyConfig{
						ID:      "test-policy-2",
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
		p := initProvider()
		expectedRoles := []*domain.Role{
			{
				ID:   "test-role",
				Name: "test_role_name",
			},
		}

		validConfig := &domain.ProviderConfig{
			Type:                "bigquery",
			URN:                 "test-URN",
			AllowedAccountTypes: []string{"user", "serviceAccount"},
			Credentials:         map[string]interface{}{},
			Resources: []*domain.ResourceConfig{
				{
					Type: "dataset",
					Policy: &domain.PolicyConfig{
						ID:      "test-policy-1",
						Version: 1,
					},
					Roles: expectedRoles,
				},
			},
		}

		actualRoles, actualError := p.GetRoles(validConfig, "dataset")

		assert.NoError(t, actualError)
		assert.Equal(t, expectedRoles, actualRoles)
	})
}

func initProvider() *bigquery.Provider {
	crypto := new(mocks.Crypto)
	return bigquery.NewProvider("bigquery", crypto)
}
