package bigquery_test

import (
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/plugins/providers/bigquery"
	"github.com/stretchr/testify/assert"
)

func TestGetType(t *testing.T) {
	t.Run("should return provider type name", func(t *testing.T) {
		p := initProvider()
		expectedTypeName := domain.ProviderTypeBigQuery

		actualTypeName := p.GetType()

		assert.Equal(t, expectedTypeName, actualTypeName)
	})
}

func TestGrantAccess(t *testing.T) {
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
