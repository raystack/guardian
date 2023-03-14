package bigquery_test

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/logging"
	"github.com/google/go-cmp/cmp"
	"github.com/goto/guardian/core/provider"
	"github.com/goto/guardian/domain"
	"github.com/goto/guardian/plugins/providers/bigquery"
	"github.com/goto/guardian/plugins/providers/bigquery/mocks"
	"github.com/goto/salt/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/genproto/googleapis/api/monitoredres"
	"google.golang.org/genproto/googleapis/cloud/audit"
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

type BigQueryProviderTestSuite struct {
	suite.Suite

	mockBigQueryClient     *mocks.BigQueryClient
	mockCloudLoggingClient *mocks.CloudLoggingClientI
	mockEncryptor          *mocks.Encryptor
	dummyProjectID         string
	provider               *bigquery.Provider

	validProvider *domain.Provider
}

func TestBigQueryProvider(t *testing.T) {
	suite.Run(t, new(BigQueryProviderTestSuite))
}

func (s *BigQueryProviderTestSuite) SetupTest() {
	s.mockBigQueryClient = new(mocks.BigQueryClient)
	s.mockCloudLoggingClient = new(mocks.CloudLoggingClientI)
	s.mockEncryptor = new(mocks.Encryptor)
	s.provider = bigquery.NewProvider("bigquery", s.mockEncryptor, log.NewNoop())
	s.dummyProjectID = "test-project-id"
	s.provider.Clients[s.dummyProjectID] = s.mockBigQueryClient
	s.provider.LogClients[s.dummyProjectID] = s.mockCloudLoggingClient

	s.validProvider = &domain.Provider{
		Type: "bigquery",
		URN:  s.dummyProjectID,
		Config: &domain.ProviderConfig{
			Type:                "bigquery",
			URN:                 s.dummyProjectID,
			AllowedAccountTypes: []string{"user", "serviceAccount"},
			Credentials: map[string]interface{}{
				"resource_name":       fmt.Sprintf("projects/%s", s.dummyProjectID),
				"service_account_key": "dummy-credentials",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: "dataset",
					Policy: &domain.PolicyConfig{
						ID:      "test-policy-1",
						Version: 1,
					},
					Roles: []*domain.Role{
						{
							ID:          "test-role",
							Name:        "test_role_name",
							Permissions: []interface{}{bigquery.DatasetRoleReader},
						},
						{
							ID:          "test-role-2",
							Name:        "test_role_name_2",
							Permissions: []interface{}{bigquery.DatasetRoleWriter},
						},
					},
				},
			},
		},
	}

	s.mockEncryptor.EXPECT().Decrypt("dummy-credentials").Return("dummy-credentials", nil)
}

func (s *BigQueryProviderTestSuite) TestListAccess() {
	s.Run("return error if initializing client fails", func() {
		s.mockEncryptor.EXPECT().Decrypt("invalid-key").Return("", errors.New("invalid-key"))

		ctx := context.Background()
		_, err := s.provider.ListAccess(ctx, domain.ProviderConfig{
			Type: "bigquery",
			URN:  "new-urn",
			Credentials: map[string]interface{}{
				"service_account_key": "invalid-key",
			},
		}, []*domain.Resource{})

		s.EqualError(err, "initializing bigquery client: bigquery: constructing client: invalid character 'i' looking for beginning of value")
	})

	s.Run("return nil error on success", func() {
		expectedResourcesAccess := domain.MapResourceAccess{}
		expectedResources := []*domain.Resource{}
		s.mockBigQueryClient.EXPECT().
			ListAccess(mock.AnythingOfType("*context.emptyCtx"), expectedResources).
			Return(expectedResourcesAccess, nil).Once()

		ctx := context.Background()
		actualResourcesAccess, actualError := s.provider.ListAccess(ctx, *s.validProvider.Config, expectedResources)

		s.mockBigQueryClient.AssertExpectations(s.T())
		s.NoError(actualError)
		s.Equal(expectedResourcesAccess, actualResourcesAccess)
	})
}

func (s *BigQueryProviderTestSuite) TestGetActivities_Success() {
	s.Run("should map bigquery logging entries to guardian activities", func() {
		now := time.Now()

		expectedBigQueryActivities := []*bigquery.Activity{
			{
				&logging.Entry{
					Timestamp: now,
					InsertID:  "test-activity-id",
					Payload: &audit.AuditLog{
						ResourceName: "projects/test-project-id/datasets/test-dataset-id",
						ServiceName:  "bigquery.googleapis.com",
						AuthenticationInfo: &audit.AuthenticationInfo{
							PrincipalEmail: "user@example.com",
						},
						AuthorizationInfo: []*audit.AuthorizationInfo{
							{
								Permission: "bigquery.datasets.get",
							},
						},
					},
					Resource: &monitoredres.MonitoredResource{
						Type: "bigquery_dataset",
						Labels: map[string]string{
							"dataset_id": "test-dataset-id",
							"project_id": "test-project-id",
						},
					},
				},
			},
		}
		s.mockCloudLoggingClient.EXPECT().
			ListLogEntries(mock.AnythingOfType("*context.emptyCtx"), bigquery.ImportActivitiesFilter{
				Types: bigquery.BigQueryAuditMetadataMethods,
			}).Return(expectedBigQueryActivities, nil).Once()
		s.mockBigQueryClient.EXPECT().
			GetRolePermissions(mock.AnythingOfType("*context.emptyCtx"), "roles/bigquery.dataViewer").Return([]string{"bigquery.datasets.get"}, nil).Once()
		s.mockBigQueryClient.EXPECT().
			GetRolePermissions(mock.AnythingOfType("*context.emptyCtx"), "roles/bigquery.dataEditor").Return([]string{"bigquery.datasets.get"}, nil).Once()

		expectedActivities := []*domain.Activity{
			{
				ProviderID:         s.validProvider.ID,
				ProviderActivityID: "test-activity-id",
				AccountType:        "user",
				AccountID:          "user@example.com",
				Timestamp:          now,
				Authorizations:     []string{"bigquery.datasets.get"},
				Resource: &domain.Resource{
					ProviderType: s.validProvider.Type,
					ProviderURN:  s.validProvider.URN,
					Type:         "dataset",
					URN:          "test-project-id:test-dataset-id",
					Name:         "test-dataset-id",
				},
				RelatedPermissions: []string{bigquery.DatasetRoleReader, bigquery.DatasetRoleWriter},
				Metadata: map[string]interface{}{
					"logging_entry": map[string]interface{}{
						"insert_id": "test-activity-id",
						"labels":    nil,
						"operation": nil,
						"payload": map[string]interface{}{
							"authentication_info": map[string]interface{}{
								"principal_email": "user@example.com",
							},
							"authorization_info": []interface{}{
								map[string]interface{}{
									"permission": "bigquery.datasets.get",
								},
							},
							"resource_name": "projects/test-project-id/datasets/test-dataset-id",
							"service_name":  "bigquery.googleapis.com",
						},
						"resource": map[string]interface{}{
							"labels": map[string]interface{}{
								"dataset_id": "test-dataset-id",
								"project_id": "test-project-id",
							},
							"type": "bigquery_dataset",
						},
						"severity":        float64(0),
						"source_location": nil,
						"span_id":         "",
						"timestamp":       now.Format(time.RFC3339Nano),
						"trace":           "",
						"trace_sampled":   false,
					},
				},
			},
		}

		actualActivities, err := s.provider.GetActivities(context.Background(), *s.validProvider, domain.ImportActivitiesFilter{})

		s.mockCloudLoggingClient.AssertExpectations(s.T())
		s.mockBigQueryClient.AssertExpectations(s.T())
		s.NoError(err)
		s.Empty(cmp.Diff(expectedActivities, actualActivities))
	})

	s.Run("should return error if there is an error on initializing logging client", func() {
		expectedError := errors.New("error")

		s.mockEncryptor.EXPECT().
			Decrypt("invalid-key").Return("", expectedError).Once()

		invalidProvider := &domain.Provider{
			Config: &domain.ProviderConfig{
				Type: "bigquery",
				URN:  "new-urn",
				Credentials: map[string]interface{}{
					"service_account_key": "invalid-key",
				},
			},
		}
		_, err := s.provider.GetActivities(context.Background(), *invalidProvider, domain.ImportActivitiesFilter{})

		s.mockEncryptor.AssertExpectations(s.T())
		s.ErrorIs(err, expectedError)
	})

	s.Run("should return error if there is an error on listing log entries", func() {
		expectedError := errors.New("error")
		s.mockCloudLoggingClient.EXPECT().
			ListLogEntries(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("bigquery.ImportActivitiesFilter")).Return(nil, expectedError).Once()

		_, err := s.provider.GetActivities(context.Background(), *s.validProvider, domain.ImportActivitiesFilter{})

		s.mockCloudLoggingClient.AssertExpectations(s.T())
		s.ErrorIs(err, expectedError)
	})
}

func initProvider() *bigquery.Provider {
	crypto := new(mocks.Encryptor)
	l := log.NewNoop()
	return bigquery.NewProvider("bigquery", crypto, l)
}
