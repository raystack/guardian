package frontier_test

import (
	"testing"

	"github.com/raystack/guardian/plugins/providers/frontier"

	"github.com/raystack/guardian/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	t.Run("should return shield config containing the same provider config", func(t *testing.T) {
		pc := &domain.ProviderConfig{}
		expectedProviderConfig := pc

		c := frontier.NewConfig(pc)
		actualProviderConfig := c.ProviderConfig

		assert.NotNil(t, c)
		assert.Equal(t, expectedProviderConfig, actualProviderConfig)
	})
}

func TestValidate(t *testing.T) {
	validCredentials := frontier.Credentials{
		Host:      "http://localhost:1234",
		AuthEmail: "guardian_test@test.com",
	}
	invalidCredentials := "invalid-credentials"

	validResourceConfig := []*domain.ResourceConfig{
		{
			Type: frontier.ResourceTypeGroup,
			Roles: []*domain.Role{
				{
					Permissions: []interface{}{frontier.RoleGroupMember, frontier.RoleGroupOwner},
				},
			},
		},
		{
			Type: frontier.ResourceTypeProject,
			Roles: []*domain.Role{
				{
					Permissions: []interface{}{frontier.RoleProjectOwner, frontier.RoleProjectManager, frontier.RoleProjectViewer},
				},
			},
		},
		{
			Type: frontier.ResourceTypeOrganization,
			Roles: []*domain.Role{
				{
					Permissions: []interface{}{frontier.RoleOrgOwner, frontier.RoleOrgManager, frontier.RoleOrgAccessManager, frontier.RoleOrgViewer},
				},
			},
		},
	}

	inValidResourcePermissionConfig := []*domain.ResourceConfig{
		{
			Type: frontier.ResourceTypeGroup,
			Roles: []*domain.Role{
				{
					Permissions: []interface{}{"member"},
				},
			},
		},
		{
			Type: frontier.ResourceTypeProject,
			Roles: []*domain.Role{
				{
					Permissions: []interface{}{"admin"},
				},
			},
		},
	}

	inValidResourceTypeConfig := []*domain.ResourceConfig{
		{
			Type: "resource-type",
			Roles: []*domain.Role{
				{
					Permissions: []interface{}{"users"},
				},
			},
		},
	}

	t.Run("error validations", func(t *testing.T) {
		testCases := []struct {
			name           string
			credentials    interface{}
			resourceConfig []*domain.ResourceConfig
		}{
			{
				name:           "should return error when invalid credentials",
				credentials:    invalidCredentials,
				resourceConfig: validResourceConfig,
			},
			{
				name:           "should return error when invalid permission config",
				credentials:    validCredentials,
				resourceConfig: inValidResourcePermissionConfig,
			},
			{
				name:           "should return error if invalid resource type config",
				credentials:    validCredentials,
				resourceConfig: inValidResourceTypeConfig,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				pc := &domain.ProviderConfig{
					Credentials: tc.credentials,
					Resources:   tc.resourceConfig,
				}
				err := frontier.NewConfig(pc).ParseAndValidate()
				assert.Error(t, err)
			})
		}
	})
}
