package gcs_test

import (
	"encoding/base64"
	"errors"
	"fmt"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/plugins/providers/gcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetType(t *testing.T) {
	t.Run("should return the typeName of the provider", func(t *testing.T) {
		expectedTypeName := "test-typeName"
		crypto := new(mocks.Crypto)
		p := gcs.NewProvider(expectedTypeName, crypto)

		actualTypeName := p.GetType()

		assert.Equal(t, expectedTypeName, actualTypeName)
	})
}

func TestCreateConfig(t *testing.T) {
	t.Run("should return error if error in parse and validate configurations", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		client := new(mocks.GcsClient)
		p := gcs.NewProvider("", crypto)
		p.Clients = map[string]gcs.GcsClient{
			"test-resource-name": client,
		}

		testcases := []struct {
			name string
			pc   *domain.ProviderConfig
		}{
			{
				name: "invalid resource type",
				pc: &domain.ProviderConfig{
					Credentials: gcs.Credentials{
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
				name: "invalid permissions for bucket resource type",
				pc: &domain.ProviderConfig{
					Credentials: gcs.Credentials{
						ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte(`{"type":"service_account"}`)),
						ResourceName:      "projects/test-resource-name",
					},
					Resources: []*domain.ResourceConfig{
						{
							Type: gcs.ResourceTypeBucket,
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
		crypto.On("Encrypt", `{"type":"service_account"}`).Return(`{"type":"service_account"}`, nil)

		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				actualError := p.CreateConfig(tc.pc)
				assert.Error(t, actualError)
			})
		}
	})

	t.Run("should return error if error in encrypting the credentials", func(t *testing.T) {
		providerURN := "test-URN"
		crypto := new(mocks.Crypto)
		client := new(mocks.GcsClient)
		p := gcs.NewProvider("", crypto)
		p.Clients = map[string]gcs.GcsClient{
			"test-resource-name": client,
		}
		pc := &domain.ProviderConfig{
			Resources: []*domain.ResourceConfig{
				{
					Type:  gcs.ResourceTypeBucket,
					Roles: []*domain.Role{},
				},
			},
			Credentials: gcs.Credentials{
				ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte(`{"type":"service_account"}`)),
				ResourceName:      "projects/test-resource-name",
			},
			URN: providerURN,
		}
		expectedError := errors.New("error in encrypting SAK")
		crypto.On("Encrypt", `{"type":"service_account"}`).Return("", expectedError)
		actualError := p.CreateConfig(pc)

		assert.Equal(t, expectedError, actualError)
	})

	t.Run("should make the provider config, parse and validate the credentials and permissions and return nil error on success", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := gcs.NewProvider("gcs", crypto)
		providerURN := "test-resource-name"
		pc := &domain.ProviderConfig{
			Type: domain.ProviderTypeGCS,
			URN:  providerURN,
			Credentials: gcs.Credentials{
				ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte(`{"type":"service_account"}`)),
				ResourceName:      "projects/test-resource-name",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: gcs.ResourceTypeBucket,
					Roles: []*domain.Role{
						{
							ID:          "Storage Legacy Bucket Writer",
							Name:        "Storage Legacy Bucket Writer",
							Description: "Read access to buckets with object listing/creation/deletion",
							Permissions: []interface{}{"roles/storage.legacyBucketWriter"},
						},
					},
				},
			},
		}
		crypto.On("Encrypt", `{"type":"service_account"}`).Return("encrypted Service Account Key", nil)

		actualError := p.CreateConfig(pc)
		assert.NoError(t, actualError)
		crypto.AssertExpectations(t)
	})
}

func TestGetResources(t *testing.T) {
	t.Run("should get the bucket resources defined in the provider config", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		client := new(mocks.GcsClient)
		p := gcs.NewProvider("gcs", crypto)
		p.Clients = map[string]gcs.GcsClient{
			"test-resource-name": client,
		}
		providerURN := "test-resource-name"
		crypto.On("Decrypt", "c2VydmljZV9hY2NvdW50LWtleS1qc29u").Return(`{"type":"service_account"}`, nil)

		pc := &domain.ProviderConfig{
			Type: domain.ProviderTypeGCS,
			URN:  providerURN,
			Credentials: gcs.Credentials{
				ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service_account-key-json")),
				ResourceName:      "projects/test-resource-name",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: gcs.ResourceTypeBucket,
					Roles: []*domain.Role{
						{
							ID:          "Storage Legacy Bucket Writer",
							Name:        "Storage Legacy Bucket Writer",
							Description: "Read access to buckets with object listing/creation/deletion",
							Permissions: []interface{}{"roles/storage.legacyBucketWriter"},
						},
					},
				},
			},
		}
		expectedBuckets := []*gcs.Bucket{
			{
				Name: "test-bucket-name",
			},
		}
		client.On("GetBuckets", mock.Anything, mock.Anything).Return(expectedBuckets, nil).Once()
		expectedResources := []*domain.Resource{
			{
				ProviderType: pc.Type,
				ProviderURN:  pc.URN,
				Type:         gcs.ResourceTypeBucket,
				URN:          "test-bucket-name",
				Name:         "test-bucket-name",
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
			appeal         *domain.Appeal
			expectedError  error
		}{
			{
				name:           "nil provider config",
				providerConfig: nil,
				expectedError:  fmt.Errorf("invalid provider/appeal config: %w", gcs.ErrNilProviderConfig),
			},
			{
				name: "nil appeal config",
				providerConfig: &domain.ProviderConfig{
					Type:                domain.ProviderTypeGCS,
					URN:                 "test-URN",
					AllowedAccountTypes: []string{"user", "serviceAccount"},
				},
				appeal:        nil,
				expectedError: fmt.Errorf("invalid provider/appeal config: %w", gcs.ErrNilAppeal),
			},
			{
				name: "nil resource config",
				providerConfig: &domain.ProviderConfig{
					Type:                domain.ProviderTypeGCS,
					URN:                 "test-URN",
					AllowedAccountTypes: []string{"user", "serviceAccount"},
				},
				appeal: &domain.Appeal{
					ID:          "test-appeal-id",
					AccountType: "user",
				},
				expectedError: fmt.Errorf("invalid provider/appeal config: %w", gcs.ErrNilResource),
			},
			{
				name: "provider type doesnt match",
				providerConfig: &domain.ProviderConfig{
					Type:                domain.ProviderTypeGCS,
					URN:                 "test-URN-1",
					AllowedAccountTypes: []string{"user", "serviceAccount"},
				},
				appeal: &domain.Appeal{
					ID:          "test-appeal-id",
					AccountType: "user",
					Resource: &domain.Resource{
						ID:           "test-resource-id",
						ProviderType: "not-gcs",
					},
				},
				expectedError: fmt.Errorf("invalid provider/appeal config: %w", gcs.ErrProviderTypeMismatch),
			},
			{
				name: "provider urn doesnt match",
				providerConfig: &domain.ProviderConfig{
					Type:                domain.ProviderTypeGCS,
					URN:                 "test-URN-1",
					AllowedAccountTypes: []string{"user", "serviceAccount"},
				},
				appeal: &domain.Appeal{
					ID:          "test-appeal-id",
					AccountType: "user",
					Resource: &domain.Resource{
						ID:           "test-resource-id",
						ProviderType: domain.ProviderTypeGCS,
						ProviderURN:  "not-test-URN-1",
					},
				},
				expectedError: fmt.Errorf("invalid provider/appeal config: %w", gcs.ErrProviderURNMismatch),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				p := initProvider()
				pc := tc.providerConfig
				a := tc.appeal

				actualError := p.GrantAccess(pc, a)
				assert.EqualError(t, actualError, tc.expectedError.Error())
			})
		}
	})

	t.Run("should return an error if there is an error in getting permissions", func(t *testing.T) {
		var permission gcs.Permission
		invalidPermissionConfig := map[string]interface{}{}
		invalidPermissionConfigError := mapstructure.Decode(invalidPermissionConfig, &permission)

		testCases := []struct {
			name            string
			resourceConfigs []*domain.ResourceConfig
			appeal          *domain.Appeal
			expectedError   error
		}{
			{
				name: "invalid resource type",
				appeal: &domain.Appeal{
					Resource: &domain.Resource{
						Type: "test-type",
					},
				},
				expectedError: fmt.Errorf("error in getting permissions: %w", gcs.ErrInvalidResourceType),
			},
			{
				name: "invalid role",
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
				expectedError: fmt.Errorf("error in getting permissions: %w", gcs.ErrInvalidRole),
			},
			{
				name: "invalid permissions config",
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
				expectedError: fmt.Errorf("error in getting permissions: %w", invalidPermissionConfigError),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				crypto := new(mocks.Crypto)
				p := gcs.NewProvider("", crypto)

				providerConfig := &domain.ProviderConfig{
					Resources: tc.resourceConfigs,
				}

				actualError := p.GrantAccess(providerConfig, tc.appeal)
				assert.EqualError(t, actualError, tc.expectedError.Error())
			})
		}
	},
	)

	t.Run("should return error if error in decoding credentials", func(t *testing.T) {
		p := initProvider()

		pc := &domain.ProviderConfig{
			Credentials: "invalid-credentials-struct",
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

	t.Run("should return error if error in decrypting the service account key", func(t *testing.T) {
		expectedAccountType := "user"
		expectedAccountID := "test@email.com"
		crypto := new(mocks.Crypto)
		client := new(mocks.GcsClient)
		p := gcs.NewProvider("gcs", crypto)
		p.Clients = map[string]gcs.GcsClient{
			"test-resource-name": client,
		}
		providerURN := "test-resource-name"
		expectedError := errors.New("Error in decrypting service account key")
		crypto.On("Decrypt", "c2VydmljZV9hY2NvdW50LWtleS1qc29u").Return(`{"type":"service_account"}`, expectedError)

		pc := &domain.ProviderConfig{
			Type: domain.ProviderTypeGCS,
			URN:  providerURN,
			Credentials: gcs.Credentials{
				ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service_account-key-json")),
				ResourceName:      "projects/test-resource-name",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: gcs.ResourceTypeBucket,
					Roles: []*domain.Role{
						{
							ID:          "Storage Legacy Bucket Writer",
							Name:        "Storage Legacy Bucket Writer",
							Description: "Read access to buckets with object listing/creation/deletion",
							Permissions: []interface{}{"roles/storage.legacyBucketWriter"},
						},
					},
				},
			},
		}
		a := &domain.Appeal{
			Role: "Storage Legacy Bucket Writer",
			Resource: &domain.Resource{
				URN:          "test-bucket-name",
				Name:         "test-bucket-name",
				ProviderType: "gcs",
				ProviderURN:  "test-resource-name",
				Type:         "bucket",
			},
			ID:          "999",
			ResourceID:  "999",
			AccountType: expectedAccountType,
			AccountID:   expectedAccountID,
		}

		actualError := p.GrantAccess(pc, a)

		assert.Error(t, actualError)
	})

	t.Run("should return error if error in getting the gcs client", func(t *testing.T) {
		expectedAccountType := "user"
		expectedAccountID := "test@email.com"
		crypto := new(mocks.Crypto)
		p := gcs.NewProvider("gcs", crypto)
		providerURN := "test-resource-name"

		crypto.On("Decrypt", "c2VydmljZV9hY2NvdW50LWtleS1qc29u").Return(`{"type":"service_account"}`, nil)

		pc := &domain.ProviderConfig{
			Type: domain.ProviderTypeGCS,
			URN:  providerURN,
			Credentials: gcs.Credentials{
				ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service_account-key-json")),
				ResourceName:      "projects/test-resource-name",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: gcs.ResourceTypeBucket,
					Roles: []*domain.Role{
						{
							ID:          "Storage Legacy Bucket Writer",
							Name:        "Storage Legacy Bucket Writer",
							Description: "Read access to buckets with object listing/creation/deletion",
							Permissions: []interface{}{"roles/storage.legacyBucketWriter"},
						},
					},
				},
			},
		}
		a := &domain.Appeal{
			Role: "Storage Legacy Bucket Writer",
			Resource: &domain.Resource{
				URN:          "test-bucket-name",
				Name:         "test-bucket-name",
				ProviderType: "gcs",
				ProviderURN:  "test-resource-name",
				Type:         "bucket",
			},
			ID:          "999",
			ResourceID:  "999",
			AccountType: expectedAccountType,
			AccountID:   expectedAccountID,
		}

		actualError := p.GrantAccess(pc, a)

		assert.Error(t, actualError)
	})

	t.Run("should grant the access to bucket resource and return nil error", func(t *testing.T) {

		expectedAccountType := "user"
		expectedAccountID := "test@email.com"

		crypto := new(mocks.Crypto)
		client := new(mocks.GcsClient)
		p := gcs.NewProvider("gcs", crypto)
		p.Clients = map[string]gcs.GcsClient{
			"test-resource-name": client,
		}
		providerURN := "test-resource-name"

		crypto.On("Decrypt", "c2VydmljZV9hY2NvdW50LWtleS1qc29u").Return(`{"type":"service_account"}`, nil)
		client.On("GrantBucketAccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		pc := &domain.ProviderConfig{
			Type: domain.ProviderTypeGCS,
			URN:  providerURN,
			Credentials: gcs.Credentials{
				ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service_account-key-json")),
				ResourceName:      "projects/test-resource-name",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: gcs.ResourceTypeBucket,
					Roles: []*domain.Role{
						{
							ID:          "Storage Legacy Bucket Writer",
							Name:        "Storage Legacy Bucket Writer",
							Description: "Read access to buckets with object listing/creation/deletion",
							Permissions: []interface{}{"roles/storage.legacyBucketWriter"},
						},
					},
				},
			},
		}

		a := &domain.Appeal{
			Role: "Storage Legacy Bucket Writer",
			Resource: &domain.Resource{
				URN:          "test-bucket-name",
				Name:         "test-bucket-name",
				ProviderType: "gcs",
				ProviderURN:  "test-resource-name",
				Type:         "bucket",
			},
			ID:          "999",
			ResourceID:  "999",
			AccountType: expectedAccountType,
			AccountID:   expectedAccountID,
		}

		actualError := p.GrantAccess(pc, a)
		assert.Nil(t, actualError)
		client.AssertExpectations(t)
	})
}

func TestRevokeAccess(t *testing.T) {

	t.Run("should return error if Provider Config or Appeal doesn't have required parameters", func(t *testing.T) {
		testCases := []struct {
			name           string
			providerConfig *domain.ProviderConfig
			appeal         *domain.Appeal
			expectedError  error
		}{
			{
				name:           "nil provider config",
				providerConfig: nil,
				expectedError:  fmt.Errorf("invalid provider/appeal config: %w", gcs.ErrNilProviderConfig),
			},
			{
				name: "nil appeal config",
				providerConfig: &domain.ProviderConfig{
					Type:                domain.ProviderTypeGCS,
					URN:                 "test-URN",
					AllowedAccountTypes: []string{"user", "serviceAccount"},
				},
				appeal:        nil,
				expectedError: fmt.Errorf("invalid provider/appeal config: %w", gcs.ErrNilAppeal),
			},
			{
				name: "nil resource config",
				providerConfig: &domain.ProviderConfig{
					Type:                domain.ProviderTypeGCS,
					URN:                 "test-URN",
					AllowedAccountTypes: []string{"user", "serviceAccount"},
				},
				appeal: &domain.Appeal{
					ID:          "test-appeal-id",
					AccountType: "user",
				},
				expectedError: fmt.Errorf("invalid provider/appeal config: %w", gcs.ErrNilResource),
			},
			{
				name: "provider type doesnt match",
				providerConfig: &domain.ProviderConfig{
					Type:                domain.ProviderTypeGCS,
					URN:                 "test-URN-1",
					AllowedAccountTypes: []string{"user", "serviceAccount"},
				},
				appeal: &domain.Appeal{
					ID:          "test-appeal-id",
					AccountType: "user",
					Resource: &domain.Resource{
						ID:           "test-resource-id",
						ProviderType: "not-gcs",
					},
				},
				expectedError: fmt.Errorf("invalid provider/appeal config: %w", gcs.ErrProviderTypeMismatch),
			},
			{
				name: "provider urn doesnt match",
				providerConfig: &domain.ProviderConfig{
					Type:                domain.ProviderTypeGCS,
					URN:                 "test-URN-1",
					AllowedAccountTypes: []string{"user", "serviceAccount"},
				},
				appeal: &domain.Appeal{
					ID:          "test-appeal-id",
					AccountType: "user",
					Resource: &domain.Resource{
						ID:           "test-resource-id",
						ProviderType: domain.ProviderTypeGCS,
						ProviderURN:  "not-test-URN-1",
					},
				},
				expectedError: fmt.Errorf("invalid provider/appeal config: %w", gcs.ErrProviderURNMismatch),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				p := initProvider()
				pc := tc.providerConfig
				a := tc.appeal

				actualError := p.RevokeAccess(pc, a)
				assert.EqualError(t, actualError, tc.expectedError.Error())
			})
		}
	})

	t.Run("should return an error if there is an error in getting permissions", func(t *testing.T) {
		var permission gcs.Permission
		invalidPermissionConfig := map[string]interface{}{}
		invalidPermissionConfigError := mapstructure.Decode(invalidPermissionConfig, &permission)

		testCases := []struct {
			name            string
			resourceConfigs []*domain.ResourceConfig
			appeal          *domain.Appeal
			expectedError   error
		}{
			{
				name: "invalid resource type",
				appeal: &domain.Appeal{
					Resource: &domain.Resource{
						Type: "test-type",
					},
				},
				expectedError: fmt.Errorf("error in getting permissions: %w", gcs.ErrInvalidResourceType),
			},
			{
				name: "invalid role",
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
				expectedError: fmt.Errorf("error in getting permissions: %w", gcs.ErrInvalidRole),
			},
			{
				name: "invalid permissions config",
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
				expectedError: fmt.Errorf("error in getting permissions: %w", invalidPermissionConfigError),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				crypto := new(mocks.Crypto)
				p := gcs.NewProvider("", crypto)

				providerConfig := &domain.ProviderConfig{
					Resources: tc.resourceConfigs,
				}

				actualError := p.RevokeAccess(providerConfig, tc.appeal)
				assert.EqualError(t, actualError, tc.expectedError.Error())
			})
		}
	},
	)

	t.Run("should return error if error in decoding credentials", func(t *testing.T) {
		p := initProvider()

		pc := &domain.ProviderConfig{
			Credentials: "invalid-credentials-struct",
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

	t.Run("should return error if error in decrypting the service account key", func(t *testing.T) {
		expectedAccountType := "user"
		expectedAccountID := "test@email.com"
		crypto := new(mocks.Crypto)
		client := new(mocks.GcsClient)
		p := gcs.NewProvider("gcs", crypto)
		p.Clients = map[string]gcs.GcsClient{
			"test-resource-name": client,
		}
		providerURN := "test-resource-name"
		expectedError := errors.New("Error in decrypting service account key")
		crypto.On("Decrypt", "c2VydmljZV9hY2NvdW50LWtleS1qc29u").Return(`{"type":"service_account"}`, expectedError)

		pc := &domain.ProviderConfig{
			Type: domain.ProviderTypeGCS,
			URN:  providerURN,
			Credentials: gcs.Credentials{
				ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service_account-key-json")),
				ResourceName:      "projects/test-resource-name",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: gcs.ResourceTypeBucket,
					Roles: []*domain.Role{
						{
							ID:          "Storage Legacy Bucket Writer",
							Name:        "Storage Legacy Bucket Writer",
							Description: "Read access to buckets with object listing/creation/deletion",
							Permissions: []interface{}{"roles/storage.legacyBucketWriter"},
						},
					},
				},
			},
		}
		a := &domain.Appeal{
			Role: "Storage Legacy Bucket Writer",
			Resource: &domain.Resource{
				URN:          "test-bucket-name",
				Name:         "test-bucket-name",
				ProviderType: "gcs",
				ProviderURN:  "test-resource-name",
				Type:         "bucket",
			},
			ID:          "999",
			ResourceID:  "999",
			AccountType: expectedAccountType,
			AccountID:   expectedAccountID,
		}

		actualError := p.RevokeAccess(pc, a)

		assert.Error(t, actualError)
	})

	t.Run("should revoke the access to bucket resource and return nil error", func(t *testing.T) {

		expectedAccountType := "user"
		expectedAccountID := "test@email.com"

		crypto := new(mocks.Crypto)
		client := new(mocks.GcsClient)
		p := gcs.NewProvider("gcs", crypto)
		p.Clients = map[string]gcs.GcsClient{
			"test-resource-name": client,
		}
		providerURN := "test-resource-name"

		crypto.On("Decrypt", "c2VydmljZV9hY2NvdW50LWtleS1qc29u").Return(`{"type":"service_account"}`, nil)
		client.On("RevokeBucketAccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		pc := &domain.ProviderConfig{
			Type: domain.ProviderTypeGCS,
			URN:  providerURN,
			Credentials: gcs.Credentials{
				ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service_account-key-json")),
				ResourceName:      "projects/test-resource-name",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: gcs.ResourceTypeBucket,
					Roles: []*domain.Role{
						{
							ID:          "Storage Legacy Bucket Writer",
							Name:        "Storage Legacy Bucket Writer",
							Description: "Read access to buckets with object listing/creation/deletion",
							Permissions: []interface{}{"roles/storage.legacyBucketWriter"},
						},
					},
				},
			},
		}

		a := &domain.Appeal{
			Role: "Storage Legacy Bucket Writer",
			Resource: &domain.Resource{
				URN:          "test-bucket-name",
				Name:         "test-bucket-name",
				ProviderType: "gcs",
				ProviderURN:  "test-resource-name",
				Type:         "bucket",
			},
			ID:          "999",
			ResourceID:  "999",
			AccountType: expectedAccountType,
			AccountID:   expectedAccountID,
		}

		actualError := p.RevokeAccess(pc, a)
		assert.Nil(t, actualError)
		client.AssertExpectations(t)
	})
}

func TestGetRoles(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		p := initProvider()
		providerURN := "test-URN"
		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: "valid-Credentials",
			Resources:   []*domain.ResourceConfig{{}},
		}
		expectedRoles := []*domain.Role(nil)

		actualRoles, _ := p.GetRoles(pc, gcs.ResourceTypeBucket)

		assert.Equal(t, expectedRoles, actualRoles)
	})
}

func TestGetAccountType(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		p := initProvider()
		expectedAccountTypes := []string{"user", "serviceAccount", "group", "domain"}

		actualAccountypes := p.GetAccountTypes()

		assert.Equal(t, expectedAccountTypes, actualAccountypes)
	})
}

func initProvider() *gcs.Provider {
	crypto := new(mocks.Crypto)
	return gcs.NewProvider("gcs", crypto)
}
