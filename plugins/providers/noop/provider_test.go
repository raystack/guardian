package noop_test

import (
	"testing"

	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/plugins/providers/noop"
	"github.com/odpf/salt/log"
	"github.com/stretchr/testify/assert"
)

func TestGetType(t *testing.T) {
	t.Run("should return the proper type name based on the initialization", func(t *testing.T) {
		logger := log.NewLogrus()

		expectedTypeName := "test-type-name"
		p := noop.NewProvider(expectedTypeName, logger)

		actualTypeName := p.GetType()

		assert.Equal(t, expectedTypeName, actualTypeName)
	})
}

func TestCreateConfig(t *testing.T) {
	t.Run("should have config type equal to noop", func(t *testing.T) {
		p := initProvider()

		config := &domain.ProviderConfig{
			Type: "invalid-type-test",
		}

		actualError := p.CreateConfig(config)

		assert.NotNil(t, actualError)
		assert.Error(t, actualError)
		assert.ErrorIs(t, actualError, noop.ErrInvalidProviderType)
	})

	t.Run("should have only 'user' as the allowed account types", func(t *testing.T) {
		p := initProvider()

		config := &domain.ProviderConfig{
			Type:                domain.ProviderTypeNoOp,
			AllowedAccountTypes: []string{"invalid-account-type"},
		}

		actualError := p.CreateConfig(config)

		assert.NotNil(t, actualError)
		assert.Error(t, actualError)
		assert.ErrorIs(t, actualError, noop.ErrInvalidAllowedAccountTypes)
	})

	t.Run("credentials should be nil", func(t *testing.T) {
		p := initProvider()

		config := &domain.ProviderConfig{
			Type:                domain.ProviderTypeNoOp,
			AllowedAccountTypes: []string{"user"},
			Credentials:         "test-creds",
		}

		actualError := p.CreateConfig(config)

		assert.NotNil(t, actualError)
		assert.Error(t, actualError)
		assert.ErrorIs(t, actualError, noop.ErrInvalidCredentials)
	})

	t.Run("resources", func(t *testing.T) {
		t.Run("should only have one resource config", func(t *testing.T) {
			p := initProvider()

			config := &domain.ProviderConfig{
				Type:                domain.ProviderTypeNoOp,
				AllowedAccountTypes: []string{"user"},
				Resources: []*domain.ResourceConfig{
					{Type: "test-type"},
					{Type: "test-type-2"},
				},
			}

			actualError := p.CreateConfig(config)

			assert.NotNil(t, actualError)
			assert.Error(t, actualError)
			assert.ErrorIs(t, actualError, noop.ErrInvalidResourceConfigLength)
		})

		t.Run("should have 'noop' resource config type", func(t *testing.T) {
			p := initProvider()

			config := &domain.ProviderConfig{
				Type:                domain.ProviderTypeNoOp,
				AllowedAccountTypes: []string{"user"},
				Resources: []*domain.ResourceConfig{
					{Type: "test-type"},
				},
			}

			actualError := p.CreateConfig(config)

			assert.NotNil(t, actualError)
			assert.Error(t, actualError)
			assert.ErrorIs(t, actualError, noop.ErrInvalidResourceConfigType)
		})

		t.Run("roles", func(t *testing.T) {
			t.Run("permissions should be empty", func(t *testing.T) {
				p := initProvider()

				config := &domain.ProviderConfig{
					Type:                domain.ProviderTypeNoOp,
					AllowedAccountTypes: []string{"user"},
					Resources: []*domain.ResourceConfig{
						{
							Type: noop.ResourceTypeNoOp,
							Roles: []*domain.Role{
								{
									ID:          "test-role",
									Name:        "Test Role",
									Permissions: []interface{}{"test-permission"},
								},
							},
						},
					},
				}

				actualError := p.CreateConfig(config)

				assert.NotNil(t, actualError)
				assert.Error(t, actualError)
				assert.ErrorIs(t, actualError, noop.ErrInvalidRolePermissions)
			})
		})

		t.Run("should return nil error if the provider config is all valid", func(t *testing.T) {
			p := initProvider()

			expectedResourceType := noop.ResourceTypeNoOp
			validConfig := &domain.ProviderConfig{
				Type:                domain.ProviderTypeNoOp,
				URN:                 "test-noop",
				AllowedAccountTypes: []string{"user"},
				Credentials:         nil,
				Appeal: &domain.AppealConfig{
					AllowPermanentAccess:         true,
					AllowActiveAccessExtensionIn: "1h",
				},
				Resources: []*domain.ResourceConfig{
					{
						Type: expectedResourceType,
						Policy: &domain.PolicyConfig{
							ID:      "test-policy",
							Version: 1,
						},
						Roles: []*domain.Role{
							{
								ID:          "test-role",
								Name:        "Test Role",
								Permissions: []interface{}{},
							},
						},
					},
				},
			}

			actualError := p.CreateConfig(validConfig)

			assert.NoError(t, actualError)
			assert.Equal(t, expectedResourceType, validConfig.Resources[0].Type)
		})
	})
}

func TestGetResources(t *testing.T) {
	t.Run("should return one no-op resource", func(t *testing.T) {
		p := initProvider()
		validConfig := &domain.ProviderConfig{
			Type:                domain.ProviderTypeNoOp,
			URN:                 "test-noop",
			AllowedAccountTypes: []string{"user"},
			Credentials:         nil,
			Appeal: &domain.AppealConfig{
				AllowPermanentAccess:         true,
				AllowActiveAccessExtensionIn: "1h",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: noop.ResourceTypeNoOp,
					Policy: &domain.PolicyConfig{
						ID:      "test-policy",
						Version: 1,
					},
					Roles: []*domain.Role{
						{
							ID:   "test-role",
							Name: "Test Role",
						},
					},
				},
			},
		}

		expectedResource := &domain.Resource{
			ProviderType: domain.ProviderTypeNoOp,
			ProviderURN:  validConfig.URN,
			Type:         noop.ResourceTypeNoOp,
			URN:          validConfig.URN,
			Name:         validConfig.URN,
		}

		actualResources, actualError := p.GetResources(validConfig)

		assert.NoError(t, actualError)
		assert.Equal(t, []*domain.Resource{expectedResource}, actualResources)
	})
}

func TestGrantAccess(t *testing.T) {
	t.Run("should return nil", func(t *testing.T) {
		p := initProvider()

		actualError := p.GrantAccess(nil, domain.Grant{})

		assert.NoError(t, actualError)
	})
}

func TestRevokeAccess(t *testing.T) {
	t.Run("should return nil", func(t *testing.T) {
		p := initProvider()

		actualError := p.RevokeAccess(nil, domain.Grant{})

		assert.NoError(t, actualError)
	})
}

func TestGetRoles(t *testing.T) {
	t.Run("should return error if resource type is invalid", func(t *testing.T) {
		p := initProvider()

		validConfig := &domain.ProviderConfig{
			Type:                domain.ProviderTypeNoOp,
			URN:                 "test-URN",
			AllowedAccountTypes: []string{"user"},
			Resources: []*domain.ResourceConfig{
				{
					Type: noop.ResourceTypeNoOp,
					Policy: &domain.PolicyConfig{
						ID:      "test-policy",
						Version: 1,
					},
				},
			},
		}

		actualRoles, actualError := p.GetRoles(validConfig, "wrong_resource_type")

		assert.Nil(t, actualRoles)
		assert.ErrorIs(t, actualError, provider.ErrInvalidResourceType)
	})

	t.Run("should return roles specified in the provider config", func(t *testing.T) {
		p := initProvider()
		expectedRoles := []*domain.Role{
			{
				ID:   "test-role",
				Name: "Test Role",
			},
		}

		validConfig := &domain.ProviderConfig{
			Type:                domain.ProviderTypeNoOp,
			URN:                 "test-URN",
			AllowedAccountTypes: []string{"user"},
			Resources: []*domain.ResourceConfig{
				{
					Type: noop.ResourceTypeNoOp,
					Policy: &domain.PolicyConfig{
						ID:      "test-policy",
						Version: 1,
					},
					Roles: expectedRoles,
				},
			},
		}

		actualRoles, actualError := p.GetRoles(validConfig, noop.ResourceTypeNoOp)

		assert.Equal(t, expectedRoles, actualRoles)
		assert.NoError(t, actualError)
	})
}

func TestGetAccountTypes(t *testing.T) {
	t.Run("should return one role \"user\"", func(t *testing.T) {
		p := initProvider()
		expectedAccountTypes := []string{"user"}

		actualAccountTypes := p.GetAccountTypes()

		assert.Equal(t, expectedAccountTypes, actualAccountTypes)
	})
}

func initProvider() *noop.Provider {
	logger := log.NewLogrus()
	return noop.NewProvider("noop", logger)
}
