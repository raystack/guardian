package gcs_test

import (
	"encoding/base64"
	"testing"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/plugins/providers/gcs"
	"github.com/stretchr/testify/assert"
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
							Permissions: []interface{}{"WRITER"},
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

// func TestGetResources(t *testing.T) {
// 	t.Run("should get the bucket resources defined in the provider config", func(t *testing.T) {
// 		crypto := new(mocks.Crypto)
// 		p := gcs.NewProvider("gcs", crypto)
// 		providerURN := "test-resource-name"
// 		pc := &domain.ProviderConfig{
// 			Type: domain.ProviderTypeGCS,
// 			URN:  providerURN,
// 			Credentials: gcs.Credentials{
// 				ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte(`{"type":"service_account"}`)),
// 				ResourceName:      "projects/test-resource-name",
// 			},
// 			Resources: []*domain.ResourceConfig{
// 				{
// 					Type: gcs.ResourceTypeBucket,
// 					Roles: []*domain.Role{
// 						{
// 							ID:          "Storage Legacy Bucket Writer",
// 							Name:        "Storage Legacy Bucket Writer",
// 							Description: "Read access to buckets with object listing/creation/deletion",
// 							Permissions: []interface{}{"roles/storage.legacyBucketWriter"},
// 						},
// 					},
// 				},
// 			},
// 		}
// 		expectedResources := []*domain.Resource{
// 			{
// 				ProviderType: pc.Type,
// 				ProviderURN:  pc.URN,
// 				Type:         gcs.ResourceTypeBucket,
// 				URN:          "projects/test-resource-name",
// 			},
// 		}

// 		crypto.On("Decrypt", "eyJ0eXBlIjoic2VydmljZV9hY2NvdW50In0=").Return(`{"type":"service_account"}`, nil)

// 		actualResources, actualError := p.GetResources(pc)
// 		assert.NoError(t, actualError)
// 		assert.Equal(t, expectedResources, actualResources)
// 	})
// }

// func TestGrantAccess(t *testing.T) {
// 	t.Run("test", func(t *testing.T) {
// 		p := initProvider()
// 		providerURN := "test-URN"
// 		pc := &domain.ProviderConfig{
// 			URN:         providerURN,
// 			Credentials: "valid-Credentials",
// 			Resources:   []*domain.ResourceConfig{{}},
// 		}
// 		a := &domain.Appeal{}

// 		actualError := p.GrantAccess(pc, a)
// 		assert.Nil(t, actualError)
// 	})
// }

func TestRevokeAccess(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		p := initProvider()
		providerURN := "test-URN"
		pc := &domain.ProviderConfig{
			URN:         providerURN,
			Credentials: "valid-Credentials",
			Resources:   []*domain.ResourceConfig{{}},
		}
		a := &domain.Appeal{}

		actualError := p.RevokeAccess(pc, a)
		assert.Nil(t, actualError)
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

		//	assert.Nil(t, actualError)
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
