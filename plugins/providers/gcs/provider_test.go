package gcs_test

import (
	"encoding/base64"
	"testing"

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
