package provider_test

import (
	"testing"

	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/domain"
	"github.com/stretchr/testify/assert"
)

func TestGetPermissions(t *testing.T) {
	pc := &domain.ProviderConfig{
		Resources: []*domain.ResourceConfig{
			{
				Type: "test-resource-type",
				Roles: []*domain.Role{
					{
						ID: "test-role",
						Permissions: []interface{}{
							"test-permission",
						},
					},
				},
			},
		},
	}

	t.Run("should return list of permission on success", func(t *testing.T) {
		expectedPermissions := pc.Resources[0].Roles[0].Permissions

		pm := provider.PermissionManager{}
		actualPermissions, actualError := pm.GetPermissions(pc, "test-resource-type", "test-role")

		assert.NoError(t, actualError)
		assert.Equal(t, expectedPermissions, actualPermissions)
	})

	t.Run("should return error if resource type not found", func(t *testing.T) {
		expectedError := provider.ErrInvalidResourceType

		pm := provider.PermissionManager{}
		actualPermissions, actualError := pm.GetPermissions(pc, "invalid-resource-type", "test-role")

		assert.ErrorIs(t, actualError, expectedError)
		assert.Nil(t, actualPermissions)
	})

	t.Run("should return error if role not found", func(t *testing.T) {
		expectedError := provider.ErrInvalidRole

		pm := provider.PermissionManager{}
		actualPermissions, actualError := pm.GetPermissions(pc, "test-resource-type", "invalid-role")

		assert.ErrorIs(t, actualError, expectedError)
		assert.Nil(t, actualPermissions)
	})
}
