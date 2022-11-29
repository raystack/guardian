package bigquery_test

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/plugins/providers/bigquery"
	"github.com/odpf/guardian/plugins/providers/bigquery/mocks"
	"github.com/odpf/salt/log"
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
		encryptor := new(mocks.Encryptor)
		client := new(mocks.BigQueryClient)
		l := log.NewNoop()
		p := bigquery.NewProvider("", encryptor, l)
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

	t.Run("should return error if error in parse and validate configurations", func(t *testing.T) {
		encryptor := new(mocks.Encryptor)
		client := new(mocks.BigQueryClient)
		l := log.NewNoop()
		p := bigquery.NewProvider("", encryptor, l)
		p.Clients = map[string]bigquery.BigQueryClient{
			"test-resource-name": client,
		}

		testcases := []struct {
			name string
			pc   *domain.ProviderConfig
		}{
			{
				name: "resource type invalid",
				pc: &domain.ProviderConfig{
					Credentials: bigquery.Credentials{
						ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte(`{"type":"service_account"}`)),
						ResourceName:      "projects/test-resource-name",
					},
					Resources: []*domain.ResourceConfig{
						{
							Type: "not dataset or table resource type",
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
			{
				name: "wrong permissions for dataset type",
				pc: &domain.ProviderConfig{
					Credentials: bigquery.Credentials{
						ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte(`{"type":"service_account"}`)),
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
			{
				name: "wrong permissions for table resource type",
				pc: &domain.ProviderConfig{
					Credentials: bigquery.Credentials{
						ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte(`{"type":"service_account"}`)),
						ResourceName:      "projects/test-resource-name",
					},
					Resources: []*domain.ResourceConfig{
						{
							Type: bigquery.ResourceTypeTable,
							Roles: []*domain.Role{
								{
									ID:          "viewer",
									Permissions: []interface{}{""},
								},
							},
						},
					},
				},
			},
		}
		encryptor.On("Encrypt", `{"type":"service_account"}`).Return(`{"type":"service_account"}`, nil)

		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				actualError := p.CreateConfig(tc.pc)
				assert.Error(t, actualError)
			})
		}
	})

	t.Run("should return error if error in parsing or validaing permissions", func(t *testing.T) {
		providerURN := "test-URN"
		encryptor := new(mocks.Encryptor)
		client := new(mocks.BigQueryClient)
		l := log.NewNoop()
		p := bigquery.NewProvider("", encryptor, l)
		p.Clients = map[string]bigquery.BigQueryClient{
			"test-resource-name": client,
		}
		pc := &domain.ProviderConfig{
			Resources: []*domain.ResourceConfig{
				{
					Type: bigquery.ResourceTypeDataset,
					Roles: []*domain.Role{
						{
							ID:          "VIEWER",
							Name:        "VIEWER",
							Permissions: []interface{}{"invalid permissions for resource type"},
						},
					},
				},
			},
			Credentials: bigquery.Credentials{
				ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte(`{"type":"service_account"}`)), // private_key
				ResourceName:      "projects/test-resource-name",
			},
			URN: providerURN,
		}

		actualError := p.CreateConfig(pc)

		assert.Error(t, actualError)
	})

	t.Run("should return error if error in encrypting the credentials", func(t *testing.T) {
		providerURN := "test-URN"
		encryptor := new(mocks.Encryptor)
		client := new(mocks.BigQueryClient)
		l := log.NewNoop()
		p := bigquery.NewProvider("", encryptor, l)
		p.Clients = map[string]bigquery.BigQueryClient{
			"test-resource-name": client,
		}
		pc := &domain.ProviderConfig{
			Resources: []*domain.ResourceConfig{
				{
					Type:  bigquery.ResourceTypeDataset,
					Roles: []*domain.Role{},
				},
			},
			Credentials: bigquery.Credentials{
				ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte(`{"type":"service_account"}`)),
				ResourceName:      "projects/test-resource-name",
			},
			URN: providerURN,
		}
		expectedError := errors.New("error in encrypting SAK")
		encryptor.On("Encrypt", `{"type":"service_account"}`).Return("", expectedError)
		actualError := p.CreateConfig(pc)

		assert.Equal(t, expectedError, actualError)
	})

	t.Run("should return nil error and create the config on success", func(t *testing.T) {
		providerURN := "test-URN"
		encryptor := new(mocks.Encryptor)
		client := new(mocks.BigQueryClient)
		l := log.NewNoop()
		p := bigquery.NewProvider("", encryptor, l)
		p.Clients = map[string]bigquery.BigQueryClient{
			"test-resource-name": client,
		}
		pc := &domain.ProviderConfig{
			Resources: []*domain.ResourceConfig{
				{
					Type: bigquery.ResourceTypeDataset,
					Roles: []*domain.Role{
						{
							ID:          "VIEWER",
							Name:        "VIEWER",
							Permissions: []interface{}{"READER"},
						},
					},
				},
			},
			Credentials: bigquery.Credentials{
				ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte(`{"type":"service_account"}`)), // private_key
				ResourceName:      "projects/test-resource-name",
			},
			URN: providerURN,
		}
		encryptor.On("Encrypt", `{"type":"service_account"}`).Return(`{"type":"service_account"}`, nil)

		actualError := p.CreateConfig(pc)

		assert.NoError(t, actualError)
		encryptor.AssertExpectations(t)
	})
}

func TestGetResources(t *testing.T) {
	t.Run("should error when credentials are invalid", func(t *testing.T) {
		encryptor := new(mocks.Encryptor)
		l := log.NewNoop()
		p := bigquery.NewProvider("", encryptor, l)
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
		encryptor := new(mocks.Encryptor)
		client := new(mocks.BigQueryClient)
		l := log.NewNoop()
		p := bigquery.NewProvider("", encryptor, l)
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
				Children: []*domain.Resource{
					{
						ProviderType: domain.ProviderTypeBigQuery,
						ProviderURN:  "test-project-id",
						Name:         "t_id",
						URN:          "p_id:d_id.t_id",
						Type:         "table",
					},
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
	t.Run("should return error if Provider Config or Appeal doesn't have required parameters", func(t *testing.T) {
		testCases := []struct {
			name           string
			providerConfig *domain.ProviderConfig
			grant          domain.Grant
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
				grant: domain.Grant{
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
				grant: domain.Grant{
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
				grant: domain.Grant{
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
			a := tc.grant

			actualError := p.GrantAccess(pc, a)
			assert.EqualError(t, actualError, tc.expectedError.Error())
		}
	})

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
		g := domain.Grant{
			Resource: &domain.Resource{
				Type: "test-type",
			},
			Role: "test-role",
		}

		actualError := p.GrantAccess(pc, g)
		assert.Error(t, actualError)
	})

	t.Run("should return error if GrantDataset Access returns an error", func(t *testing.T) {
		expectedError := errors.New("Test-Error")
		expectedAccountType := "user"
		expectedAccountID := "test@email.com"
		encryptor := new(mocks.Encryptor)
		client := new(mocks.BigQueryClient)
		l := log.NewNoop()
		p := bigquery.NewProvider("bigquery", encryptor, l)
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
		g := domain.Grant{
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
			Permissions: []string{"VIEWER"},
		}

		actualError := p.GrantAccess(pc, g)

		assert.Equal(t, expectedError, actualError)
	})

	t.Run("should grant access to dataset resource and return no error on success", func(t *testing.T) {
		expectedAccountType := "user"
		expectedAccountID := "test@email.com"
		encryptor := new(mocks.Encryptor)
		client := new(mocks.BigQueryClient)
		l := log.NewNoop()
		p := bigquery.NewProvider("bigquery", encryptor, l)
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
		g := domain.Grant{
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
			Permissions: []string{"VIEWER"},
		}

		actualError := p.GrantAccess(pc, g)

		assert.Nil(t, actualError)
		client.AssertExpectations(t)
	})

	t.Run("should grant access to table resource and return no error on success", func(t *testing.T) {
		providerURN := "test-URN"
		expectedAccountType := "user"
		expectedAccountID := "test@email.com"
		encryptor := new(mocks.Encryptor)
		client := new(mocks.BigQueryClient)
		l := log.NewNoop()
		p := bigquery.NewProvider("bigquery", encryptor, l)
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
		g := domain.Grant{
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
			Permissions: []string{"VIEWER"},
		}

		actualError := p.GrantAccess(pc, g)

		assert.Nil(t, actualError)
		client.AssertExpectations(t)
	})
}

func TestRevokeAccess(t *testing.T) {
	t.Run("should return error if Provider Config or Appeal doesn't have required parameters", func(t *testing.T) {
		testCases := []struct {
			providerConfig *domain.ProviderConfig
			grant          domain.Grant
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
				grant: domain.Grant{
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
				grant: domain.Grant{
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
				grant: domain.Grant{
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
			a := tc.grant

			actualError := p.RevokeAccess(pc, a)
			assert.EqualError(t, actualError, tc.expectedError.Error())
		}
	})

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
		g := domain.Grant{
			Resource: &domain.Resource{
				Type: "test-type",
			},
			Role: "test-role",
		}

		actualError := p.RevokeAccess(pc, g)
		assert.Error(t, actualError)
	})

	t.Run("should return error if Revoke Dataset Access returns an error", func(t *testing.T) {
		expectedError := errors.New("Test-Error")
		expectedAccountType := "user"
		expectedAccountID := "test@email.com"
		encryptor := new(mocks.Encryptor)
		client := new(mocks.BigQueryClient)
		l := log.NewNoop()
		p := bigquery.NewProvider("bigquery", encryptor, l)
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
		g := domain.Grant{
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
			Permissions: []string{"VIEWER"},
		}

		actualError := p.RevokeAccess(pc, g)

		assert.Equal(t, expectedError, actualError)
	})

	t.Run("should Revoke access to dataset resource and return no error on success", func(t *testing.T) {
		expectedAccountType := "user"
		expectedAccountID := "test@email.com"
		encryptor := new(mocks.Encryptor)
		client := new(mocks.BigQueryClient)
		l := log.NewNoop()
		p := bigquery.NewProvider("bigquery", encryptor, l)
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
		g := domain.Grant{
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
			Permissions: []string{"VIEWER"},
		}

		actualError := p.RevokeAccess(pc, g)

		assert.Nil(t, actualError)
	})

	t.Run("should Revoke access to table resource and return no error on success", func(t *testing.T) {
		providerURN := "test-URN"
		expectedAccountType := "user"
		expectedAccountID := "test@email.com"
		encryptor := new(mocks.Encryptor)
		client := new(mocks.BigQueryClient)
		l := log.NewNoop()
		p := bigquery.NewProvider("bigquery", encryptor, l)
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
		g := domain.Grant{
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
			Permissions: []string{"VIEWER"},
		}

		actualError := p.RevokeAccess(pc, g)

		assert.Nil(t, actualError)
		client.AssertExpectations(t)
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
	crypto := new(mocks.Encryptor)
	l := log.NewNoop()
	return bigquery.NewProvider("bigquery", crypto, l)
}
