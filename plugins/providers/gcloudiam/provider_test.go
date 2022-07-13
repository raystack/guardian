package gcloudiam_test

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/plugins/providers/gcloudiam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateConfig(t *testing.T) {
	providerURN := "test-URN"
	crypto := new(mocks.Crypto)
	client := new(mocks.GcloudIamClient)
	p := gcloudiam.NewProvider("", crypto)
	p.Clients = map[string]gcloudiam.GcloudIamClient{
		providerURN: client,
	}
	crypto.On("Encrypt", "service-account-key-json").Return("encrypted-SAK", nil)
	pc := &domain.ProviderConfig{
		Resources: []*domain.ResourceConfig{
			{
				Type:  gcloudiam.ResourceTypeProject,
				Roles: []*domain.Role{},
			},
		},
		//	base64.StdEncoding.EncodeToString([]byte("service-account-key-json"))

		// Credentials: map[string]interface{}{
		// 	"resource_name": "projects/test-resource-name",
		// },
		Credentials: gcloudiam.Credentials{
			ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service-account-key-json")),
			ResourceName:      "projects/test-resource-name",
		},
		URN: providerURN,
	}

	actualError := p.CreateConfig(pc)

	assert.NoError(t, actualError)
}

func TestGetType(t *testing.T) {
	t.Run("should return provider type name", func(t *testing.T) {
		expectedTypeName := domain.ProviderTypeMetabase
		crypto := new(mocks.Crypto)
		p := gcloudiam.NewProvider(expectedTypeName, crypto)

		actualTypeName := p.GetType()

		assert.Equal(t, expectedTypeName, actualTypeName)
	})
}

func TestGetResources(t *testing.T) {

	t.Run("should error when credentials are invalid", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := gcloudiam.NewProvider("", crypto)
		pc := &domain.ProviderConfig{
			Type:        domain.ProviderTypeGCloudIAM,
			URN:         "test-project-id",
			Credentials: "invalid-creds",
		}

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.Error(t, actualError)
	})

	t.Run("should return project resource object", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := gcloudiam.NewProvider("", crypto)
		pc := &domain.ProviderConfig{
			Type: domain.ProviderTypeGCloudIAM,
			URN:  "test-project-id",
			Credentials: map[string]interface{}{
				"resource_name": "project/test-resource-name",
			},
		}

		expectedResources := []*domain.Resource{
			{
				ProviderType: pc.Type,
				ProviderURN:  pc.URN,
				Type:         gcloudiam.ResourceTypeProject,
				URN:          "project/test-resource-name",
				Name:         "project/test-resource-name - GCP IAM",
			},
		}

		actualResources, actualError := p.GetResources(pc)

		assert.Equal(t, expectedResources, actualResources)
		assert.Nil(t, actualError)
	})

	t.Run("should return organization resource object", func(t *testing.T) {
		crypto := new(mocks.Crypto)
		p := gcloudiam.NewProvider("", crypto)
		pc := &domain.ProviderConfig{
			Type: domain.ProviderTypeGCloudIAM,
			URN:  "test-project-id",
			Credentials: map[string]interface{}{
				"resource_name": "organization/test-resource-name",
			},
		}

		expectedResources := []*domain.Resource{
			{
				ProviderType: pc.Type,
				ProviderURN:  pc.URN,
				Type:         gcloudiam.ResourceTypeOrganization,
				URN:          "organization/test-resource-name",
				Name:         "organization/test-resource-name - GCP IAM",
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
		p := gcloudiam.NewProvider("", crypto)

		pc := &domain.ProviderConfig{
			Credentials: "invalid-credentials",
			Resources: []*domain.ResourceConfig{
				{
					Type: "test-type",
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

	t.Run("should return error if resource type is unknown", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.GcloudIamClient)
		p := gcloudiam.NewProvider("", crypto)
		p.Clients = map[string]gcloudiam.GcloudIamClient{
			providerURN: client,
		}
		expectedError := errors.New("invalid resource type")

		pc := &domain.ProviderConfig{
			URN: providerURN,
			Resources: []*domain.ResourceConfig{
				{
					Type: "test-type",
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

	t.Run("should return error if there is an error in granting the access", func(t *testing.T) {
		providerURN := "test-provider-urn"
		expectedError := errors.New("client error")
		crypto := new(mocks.Crypto)
		client := new(mocks.GcloudIamClient)
		p := gcloudiam.NewProvider("", crypto)
		p.Clients = map[string]gcloudiam.GcloudIamClient{
			providerURN: client,
		}
		client.On("GrantAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

		pc := &domain.ProviderConfig{
			Resources: []*domain.ResourceConfig{
				{
					Type: gcloudiam.ResourceTypeProject,
				},
			},
			URN: providerURN,
		}
		a := &domain.Appeal{
			Resource: &domain.Resource{
				Type: gcloudiam.ResourceTypeProject,
				URN:  "999",
				Name: "test-role",
			},
			Role: "test-role",
		}

		actualError := p.GrantAccess(pc, a)

		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return nil error if granting access is successful", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.GcloudIamClient)
		expectedRole := "test-role"
		expectedAccountType := "user"
		expectedAccountID := "test@email.com"
		p := gcloudiam.NewProvider("", crypto)
		p.Clients = map[string]gcloudiam.GcloudIamClient{
			providerURN: client,
		}
		client.On("GrantAccess", expectedAccountType, expectedAccountID, expectedRole).Return(nil).Once()

		pc := &domain.ProviderConfig{
			Resources: []*domain.ResourceConfig{
				{
					Type: gcloudiam.ResourceTypeProject,
				},
			},
			URN: providerURN,
		}
		a := &domain.Appeal{
			Resource: &domain.Resource{
				Type: gcloudiam.ResourceTypeProject,
				URN:  "test-role",
			},
			Role:        expectedRole,
			AccountType: expectedAccountType,
			AccountID:   expectedAccountID,
			ResourceID:  "999",
			ID:          "999",
		}

		actualError := p.GrantAccess(pc, a)

		assert.Nil(t, actualError)
	})
}

func TestRevokeAccess(t *testing.T) {
	t.Run("should return error if resource type is unknown", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.GcloudIamClient)
		p := gcloudiam.NewProvider("", crypto)
		p.Clients = map[string]gcloudiam.GcloudIamClient{
			providerURN: client,
		}
		expectedError := errors.New("invalid resource type")

		pc := &domain.ProviderConfig{
			URN: providerURN,
			Resources: []*domain.ResourceConfig{
				{
					Type: "test-type",
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

	t.Run("should return error if there is an error in granting the access", func(t *testing.T) {
		providerURN := "test-provider-urn"
		expectedError := errors.New("client error")
		crypto := new(mocks.Crypto)
		client := new(mocks.GcloudIamClient)
		p := gcloudiam.NewProvider("", crypto)
		p.Clients = map[string]gcloudiam.GcloudIamClient{
			providerURN: client,
		}
		client.On("RevokeAccess", mock.Anything, mock.Anything, mock.Anything).Return(expectedError).Once()

		pc := &domain.ProviderConfig{
			Resources: []*domain.ResourceConfig{
				{
					Type: gcloudiam.ResourceTypeProject,
				},
			},
			URN: providerURN,
		}
		a := &domain.Appeal{
			Resource: &domain.Resource{
				Type: gcloudiam.ResourceTypeProject,
				URN:  "999",
				Name: "test-role",
			},
			Role: "test-role",
		}

		actualError := p.RevokeAccess(pc, a)

		assert.EqualError(t, actualError, expectedError.Error())
	})

	t.Run("should return nil error if revoking access is successful", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.GcloudIamClient)
		expectedRole := "test-role"
		expectedAccountType := "user"
		expectedAccountID := "test@email.com"
		p := gcloudiam.NewProvider("", crypto)
		p.Clients = map[string]gcloudiam.GcloudIamClient{
			providerURN: client,
		}
		client.On("RevokeAccess", expectedAccountType, expectedAccountID, expectedRole).Return(nil).Once()

		pc := &domain.ProviderConfig{
			Resources: []*domain.ResourceConfig{
				{
					Type: gcloudiam.ResourceTypeProject,
				},
			},
			URN: providerURN,
		}
		a := &domain.Appeal{
			Resource: &domain.Resource{
				Type: gcloudiam.ResourceTypeProject,
				URN:  "test-role",
			},
			Role:        expectedRole,
			AccountType: expectedAccountType,
			AccountID:   expectedAccountID,
			ResourceID:  "999",
			ID:          "999",
		}

		actualError := p.RevokeAccess(pc, a)

		assert.Nil(t, actualError)
	})
}

func TestGetRoles(t *testing.T) {

	t.Run("should return error if resource type is not project or organisation", func(t *testing.T) {
		expectedError := gcloudiam.ErrInvalidResourceType
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.GcloudIamClient)
		p := gcloudiam.NewProvider("", crypto)
		p.Clients = map[string]gcloudiam.GcloudIamClient{
			providerURN: client,
		}

		pc := &domain.ProviderConfig{
			Resources: []*domain.ResourceConfig{
				{
					Type: gcloudiam.ResourceTypeProject,
				},
			},
			URN: providerURN,
		}

		_, actualError := p.GetRoles(pc, "not_proj_not_org")

		assert.Equal(t, expectedError, actualError)
	})

	t.Run("should get the expected roles and no error", func(t *testing.T) {
		providerURN := "test-provider-urn"
		crypto := new(mocks.Crypto)
		client := new(mocks.GcloudIamClient)
		expectedRoles := []*domain.Role{
			{
				ID:          "test-role",
				Name:        "title",
				Description: "About the description",
			},
		}
		returnedRoles := []*gcloudiam.Role{
			{
				Name:        "test-role",
				Title:       "title",
				Description: "About the description",
			},
		}
		p := gcloudiam.NewProvider("", crypto)
		p.Clients = map[string]gcloudiam.GcloudIamClient{
			providerURN: client,
		}
		client.On("GetRoles").Return(returnedRoles, nil).Once()

		pc := &domain.ProviderConfig{
			Resources: []*domain.ResourceConfig{
				{
					Type: gcloudiam.ResourceTypeProject,
					Roles: []*domain.Role{
						{
							Name: "test-role",
						},
					},
				},
			},
			URN: providerURN,
		}

		actualRoles, actualError := p.GetRoles(pc, "project")

		assert.Equal(t, expectedRoles, actualRoles)
		assert.Nil(t, actualError)
	})

}

func TestGetAccountTypes(t *testing.T) {
	t.Run("should return the supported account types: user, serviceAccount", func(t *testing.T) {
		expectedAccountTypes := []string{"user", "serviceAccount"}
		crypto := new(mocks.Crypto)
		p := gcloudiam.NewProvider("", crypto)

		actualAccountTypes := p.GetAccountTypes()

		assert.Equal(t, expectedAccountTypes, actualAccountTypes)
	})
}
