package bigquery_test

import (
	"encoding/base64"
	"testing"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/provider/bigquery"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	t.Run("should return bigquery config containing the same provider config", func(t *testing.T) {
		pc := &domain.ProviderConfig{}
		expectedProviderConfig := pc

		c := bigquery.NewConfig(pc)
		actualProviderConfig := c.ProviderConfig

		assert.NotNil(t, c)
		assert.Equal(t, expectedProviderConfig, actualProviderConfig)
	})
}

func TestValidate(t *testing.T) {
	validCredentials := base64.StdEncoding.EncodeToString([]byte("service-account-key-json"))
	validPermissionConfig := map[string]interface{}{
		"name": "roleName",
	}

	t.Run("error validations", func(t *testing.T) {
		testCases := []struct {
			name             string
			credentials      interface{}
			permissionConfig interface{}
		}{
			{
				name:             "should return error if credentials is not a base64 string",
				credentials:      "non-base64-value",
				permissionConfig: validPermissionConfig,
			},
			{
				name:             "should return error if permission type is invalid",
				credentials:      validCredentials,
				permissionConfig: 0,
			},
			{
				name:        "should return error if permission config does not contain name field",
				credentials: validCredentials,
				permissionConfig: map[string]interface{}{
					"target": "target_value",
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				pc := &domain.ProviderConfig{
					Credentials: tc.credentials,
					Resources: []*domain.ResourceConfig{
						{
							Roles: []*domain.RoleConfig{
								{
									Permissions: []interface{}{tc.permissionConfig},
								},
							},
						},
					},
				}

				err := bigquery.NewConfig(pc).Validate()
				assert.Error(t, err)
			})
		}
	})

	t.Run("should update credentials and permission config values into castable bigquery config", func(t *testing.T) {
		pc := &domain.ProviderConfig{
			Credentials: validCredentials,
			Resources: []*domain.ResourceConfig{
				{
					Roles: []*domain.RoleConfig{
						{
							Permissions: []interface{}{validPermissionConfig},
						},
					},
				},
			},
		}

		err := bigquery.NewConfig(pc).Validate()
		_, credentialsOk := pc.Credentials.(*bigquery.Credentials)
		_, permissionConfigOk := pc.Resources[0].Roles[0].Permissions[0].(*bigquery.PermissionConfig)

		assert.Nil(t, err)
		assert.True(t, credentialsOk)
		assert.True(t, permissionConfigOk)
	})
}
