package dataplex_test

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"

	"github.com/raystack/guardian/core/provider"
	"github.com/raystack/guardian/domain"
	"github.com/raystack/guardian/plugins/providers/dataplex"
	"github.com/raystack/guardian/plugins/providers/dataplex/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestGetType(t *testing.T) {
	t.Run("should return provider type name", func(t *testing.T) {
		p := initProvider()
		expectedTypeName := domain.ProviderTypePolicyTag

		actualTypeName := p.GetType()

		assert.Equal(t, expectedTypeName, actualTypeName)
	})
}

func TestCreateConfig(t *testing.T) {
	t.Run("should return error if error in credentials are invalid/mandatory fields are missing", func(t *testing.T) {
		encryptor := new(mocks.Encryptor)
		client := new(mocks.DataplexClient)
		p := dataplex.NewProvider("", encryptor)
		p.Clients = map[string]dataplex.PolicyTagClient{
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
							Type: dataplex.ResourceTypeTag,
						},
					},
				},
			},
			{
				name: "empty mandatory credentials",
				pc: &domain.ProviderConfig{
					Credentials: dataplex.Credentials{
						ServiceAccountKey: "",
						ResourceName:      "",
					},
					Resources: []*domain.ResourceConfig{
						{
							Type: dataplex.ResourceTypeTag,
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
		client := new(mocks.DataplexClient)
		p := dataplex.NewProvider("", encryptor)
		p.Clients = map[string]dataplex.PolicyTagClient{
			"test-resource-name": client,
		}

		testcases := []struct {
			name string
			pc   *domain.ProviderConfig
		}{
			{
				name: "resource type invalid",
				pc: &domain.ProviderConfig{
					Credentials: dataplex.Credentials{
						ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte(`{"type":"service_account"}`)),
						ResourceName:      "projects/project-name/location/us",
					},
					Resources: []*domain.ResourceConfig{
						{
							Type: "not tag type",
							Roles: []*domain.Role{
								{
									ID:          "fineGrainReader",
									Permissions: []interface{}{"wrong permissions"},
								},
							},
						},
					},
				},
			},
			{
				name: "wrong permissions for tag type",
				pc: &domain.ProviderConfig{
					Credentials: dataplex.Credentials{
						ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte(`{"type":"service_account"}`)),
						ResourceName:      "projects/project-name/location/us",
					},
					Resources: []*domain.ResourceConfig{
						{
							Type: dataplex.ResourceTypeTag,
							Roles: []*domain.Role{
								{
									ID:          "fineGrainReader",
									Permissions: []interface{}{"wrong permissions"},
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
		client := new(mocks.DataplexClient)
		p := dataplex.NewProvider("", encryptor)
		p.Clients = map[string]dataplex.PolicyTagClient{
			"test-resource-name": client,
		}
		pc := &domain.ProviderConfig{
			Resources: []*domain.ResourceConfig{
				{
					Type: dataplex.ResourceTypeTag,
					Roles: []*domain.Role{
						{
							ID:          "fineGrainReader",
							Name:        "Fine Grain Reader",
							Permissions: []interface{}{"invalid permissions for resource type"},
						},
					},
				},
			},
			Credentials: dataplex.Credentials{
				ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte(`{"type":"service_account"}`)), // private_key
				ResourceName:      "projects/project-name/location/us",
			},
			URN: providerURN,
		}

		actualError := p.CreateConfig(pc)

		assert.Error(t, actualError)
	})

	t.Run("should return error if error in encrypting the credentials", func(t *testing.T) {
		providerURN := "test-URN"
		encryptor := new(mocks.Encryptor)
		client := new(mocks.DataplexClient)
		p := dataplex.NewProvider("", encryptor)
		p.Clients = map[string]dataplex.PolicyTagClient{
			"test-resource-name": client,
		}
		pc := &domain.ProviderConfig{
			Resources: []*domain.ResourceConfig{
				{
					Type:  dataplex.ResourceTypeTag,
					Roles: []*domain.Role{},
				},
			},
			Credentials: dataplex.Credentials{
				ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte(`{"type":"service_account"}`)),
				ResourceName:      "projects/project-name/location/us",
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
		client := new(mocks.DataplexClient)
		p := dataplex.NewProvider("", encryptor)
		p.Clients = map[string]dataplex.PolicyTagClient{
			"test-resource-name": client,
		}
		pc := &domain.ProviderConfig{
			Resources: []*domain.ResourceConfig{
				{
					Type: dataplex.ResourceTypeTag,
					Roles: []*domain.Role{
						{
							ID:          "fineGrainReader",
							Name:        "Fine Grain Reader",
							Permissions: []interface{}{"roles/datacatalog.categoryFineGrainedReader"},
						},
					},
				},
			},
			Credentials: dataplex.Credentials{
				ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte(`{"type":"service_account"}`)), // private_key
				ResourceName:      "projects/project-name/location/us",
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
		p := dataplex.NewProvider("", encryptor)
		pc := &domain.ProviderConfig{
			Type:        domain.ProviderTypePolicyTag,
			URN:         "test-project-id",
			Credentials: "invalid-creds",
		}

		actualResources, actualError := p.GetResources(pc)

		assert.Nil(t, actualResources)
		assert.Error(t, actualError)
	})

	t.Run("should return policy resource object", func(t *testing.T) {
		providerUrn := "policy-tag-urn"
		encryptor := new(mocks.Encryptor)
		client := new(mocks.DataplexClient)
		p := dataplex.NewProvider("", encryptor)
		p.Clients = map[string]dataplex.PolicyTagClient{
			"project-name": client,
		}
		validCredentials := dataplex.Credentials{
			ServiceAccountKey: "12345",
			ResourceName:      "projects/project-name/locations/us",
		}
		pc := &domain.ProviderConfig{
			Type:        domain.ProviderTypePolicyTag,
			URN:         providerUrn,
			Credentials: validCredentials,
			Resources: []*domain.ResourceConfig{
				{
					Type: "tag",
					Roles: []*domain.Role{
						{
							ID:          "fineGrainReader",
							Name:        "Fine Grain Reader",
							Permissions: []interface{}{"roles/datacatalog.categoryFineGrainedReader"},
						},
					},
				},
			},
		}
		expectedPolicies := []*dataplex.Policy{
			{
				Name:        "p_name",
				DisplayName: "p_displayname",
				Description: "p_description",
			},
		}
		client.On("GetPolicies", mock.Anything).Return(expectedPolicies, nil).Once()

		expectedResources := []*domain.Resource{
			{
				ProviderType: domain.ProviderTypePolicyTag,
				ProviderURN:  providerUrn,
				Type:         "tag",
				Name:         "p_displayname",
				URN:          "p_name",
				Details: map[string]interface{}{
					"description": "p_description",
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
				expectedError:  dataplex.ErrNilProviderConfig,
			},
			{
				providerConfig: &domain.ProviderConfig{
					Type:                "dataplex",
					URN:                 "test-URN",
					AllowedAccountTypes: []string{"user", "serviceAccount"},
				},
				grant: domain.Grant{
					ID:          "test-appeal-id",
					AccountType: "user",
				},
				expectedError: dataplex.ErrNilResource,
			},
			{
				providerConfig: &domain.ProviderConfig{
					Type:                "dataplex",
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
				expectedError: dataplex.ErrProviderTypeMismatch,
			},
			{
				providerConfig: &domain.ProviderConfig{
					Type:                "dataplex",
					URN:                 "test-URN-1",
					AllowedAccountTypes: []string{"user", "serviceAccount"},
				},
				grant: domain.Grant{
					ID:          "test-appeal-id",
					AccountType: "user",
					Resource: &domain.Resource{
						ID:           "test-resource-id",
						ProviderType: "dataplex",
						ProviderURN:  "not-test-URN-1",
					},
				},
				expectedError: dataplex.ErrProviderURNMismatch,
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

	t.Run("should return error if Grant Policy Access returns an error", func(t *testing.T) {
		expectedError := errors.New("Test-Error")
		expectedAccountType := "user"
		expectedAccountID := "test@email.com"
		encryptor := new(mocks.Encryptor)
		client := new(mocks.DataplexClient)
		p := dataplex.NewProvider("dataplex", encryptor)
		p.Clients = map[string]dataplex.PolicyTagClient{
			"resource-name": client,
		}
		validCredentials := dataplex.Credentials{
			ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service-account-key-json")),
			ResourceName:      "projects/resource-name/locations/us",
		}
		policy := &dataplex.Policy{}
		client.On("GrantPolicyAccess", mock.AnythingOfType("*context.emptyCtx"), policy, "user:test@email.com", "roles/datacatalog.categoryFineGrainedReader").Return(expectedError).Once()

		pc := &domain.ProviderConfig{
			Type:        "dataplex",
			URN:         "p_id:d_id",
			Credentials: validCredentials,
			Resources: []*domain.ResourceConfig{
				{
					Type: "tag",
					Roles: []*domain.Role{
						{
							ID:          "fineGrainReader",
							Name:        "Fine Grain Reader",
							Permissions: []interface{}{"roles/datacatalog.categoryFineGrainedReader"},
						},
					},
				},
			},
		}
		g := domain.Grant{
			Role: "fineGrainReader",
			Resource: &domain.Resource{
				ProviderType: "dataplex",
				ProviderURN:  "p_id:d_id",
				Type:         "tag",
			},
			ID:          "999",
			ResourceID:  "999",
			AccountType: expectedAccountType,
			AccountID:   expectedAccountID,
			Permissions: []string{"roles/datacatalog.categoryFineGrainedReader"},
		}

		actualError := p.GrantAccess(pc, g)

		assert.Equal(t, expectedError, actualError)
	})

	t.Run("should grant access to policy resource and return no error on success", func(t *testing.T) {
		providerURN := "test-URN"
		expectedAccountType := "user"
		expectedAccountID := "test@email.com"
		encryptor := new(mocks.Encryptor)
		client := new(mocks.DataplexClient)
		p := dataplex.NewProvider("dataplex", encryptor)
		p.Clients = map[string]dataplex.PolicyTagClient{
			"resource-name": client,
		}
		validCredentials := dataplex.Credentials{
			ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service-account-key-json")),
			ResourceName:      "projects/resource-name/locations/us",
		}
		policy := &dataplex.Policy{
			Name:        "p_id:d_id.t_id",
			DisplayName: "",
			Description: "",
		}
		client.On("GrantPolicyAccess", mock.AnythingOfType("*context.emptyCtx"), policy, "user:test@email.com", "roles/datacatalog.categoryFineGrainedReader").Return(dataplex.ErrPermissionAlreadyExists).Once()

		pc := &domain.ProviderConfig{
			Type:        "dataplex",
			URN:         providerURN,
			Credentials: validCredentials,
			Resources: []*domain.ResourceConfig{
				{
					Type: "tag",
					Roles: []*domain.Role{
						{
							ID:          "fineGrainReader",
							Name:        "Fine Grain Reader",
							Permissions: []interface{}{"roles/datacatalog.categoryFineGrainedReader"},
						},
					},
				},
			},
		}
		g := domain.Grant{
			Role: "fineGrainReader",
			Resource: &domain.Resource{
				URN:          "p_id:d_id.t_id",
				Name:         "t_id",
				ProviderType: "dataplex",
				ProviderURN:  "test-URN",
				Type:         "tag",
			},
			ID:          "999",
			ResourceID:  "999",
			AccountType: expectedAccountType,
			AccountID:   expectedAccountID,
			Permissions: []string{"roles/datacatalog.categoryFineGrainedReader"},
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
				expectedError:  dataplex.ErrNilProviderConfig,
			},
			{
				providerConfig: &domain.ProviderConfig{
					Type:                "dataplex",
					URN:                 "test-URN",
					AllowedAccountTypes: []string{"user", "serviceAccount"},
				},
				grant: domain.Grant{
					ID:          "test-appeal-id",
					AccountType: "user",
				},
				expectedError: dataplex.ErrNilResource,
			},
			{
				providerConfig: &domain.ProviderConfig{
					Type:                "dataplex",
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
				expectedError: dataplex.ErrProviderTypeMismatch,
			},
			{
				providerConfig: &domain.ProviderConfig{
					Type:                "dataplex",
					URN:                 "test-URN-1",
					AllowedAccountTypes: []string{"user", "serviceAccount"},
				},
				grant: domain.Grant{
					ID:          "test-appeal-id",
					AccountType: "user",
					Resource: &domain.Resource{
						ID:           "test-resource-id",
						ProviderType: "dataplex",
						ProviderURN:  "not-test-URN-1",
					},
				},
				expectedError: dataplex.ErrProviderURNMismatch,
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

	t.Run("should return error if Revoke policy Access returns an error", func(t *testing.T) {
		expectedError := errors.New("Test-Error")
		expectedAccountType := "user"
		expectedAccountID := "test@email.com"
		encryptor := new(mocks.Encryptor)
		client := new(mocks.DataplexClient)
		p := dataplex.NewProvider("dataplex", encryptor)
		p.Clients = map[string]dataplex.PolicyTagClient{
			"resource-name": client,
		}

		validCredentials := dataplex.Credentials{
			ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service-account-key-json")),
			ResourceName:      "projects/resource-name/locations/us",
		}
		policy := &dataplex.Policy{}

		client.On("RevokePolicyAccess", mock.AnythingOfType("*context.emptyCtx"), policy, "user:test@email.com", "roles/datacatalog.categoryFineGrainedReader").Return(expectedError).Once()

		pc := &domain.ProviderConfig{
			Type:        "dataplex",
			URN:         "p_id:d_id",
			Credentials: validCredentials,
			Resources: []*domain.ResourceConfig{
				{
					Type: "tag",
					Roles: []*domain.Role{
						{
							ID:          "fineGrainReader",
							Name:        "Fine Grain Reader",
							Permissions: []interface{}{"roles/datacatalog.categoryFineGrainedReader"},
						},
					},
				},
			},
		}
		g := domain.Grant{
			Role: "fineGrainReader",
			Resource: &domain.Resource{
				ProviderType: "dataplex",
				ProviderURN:  "p_id:d_id",
				Type:         "tag",
			},
			ID:          "999",
			ResourceID:  "999",
			AccountType: expectedAccountType,
			AccountID:   expectedAccountID,
			Permissions: []string{"roles/datacatalog.categoryFineGrainedReader"},
		}

		actualError := p.RevokeAccess(pc, g)

		assert.Equal(t, expectedError, actualError)
	})

	t.Run("should Revoke access to policy resource and return no error on success", func(t *testing.T) {
		expectedAccountType := "user"
		expectedAccountID := "test@email.com"
		encryptor := new(mocks.Encryptor)
		client := new(mocks.DataplexClient)
		p := dataplex.NewProvider("dataplex", encryptor)
		p.Clients = map[string]dataplex.PolicyTagClient{
			"resource-name": client,
		}
		validCredentials := dataplex.Credentials{
			ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service-account-key-json")),
			ResourceName:      "projects/resource-name/locations/us",
		}
		policy := &dataplex.Policy{}

		client.On("RevokePolicyAccess", mock.AnythingOfType("*context.emptyCtx"), policy, "user:test@email.com", "roles/datacatalog.categoryFineGrainedReader").Return(dataplex.ErrPermissionNotFound).Once()

		pc := &domain.ProviderConfig{
			Type:        "dataplex",
			URN:         "p_id:d_id",
			Credentials: validCredentials,
			Resources: []*domain.ResourceConfig{
				{
					Type: "tag",
					Roles: []*domain.Role{
						{
							ID:          "fineGrainReader",
							Name:        "Fine Grain Reader",
							Permissions: []interface{}{"roles/datacatalog.categoryFineGrainedReader"},
						},
					},
				},
			},
		}
		g := domain.Grant{
			Role: "fineGrainReader",
			Resource: &domain.Resource{
				ProviderType: "dataplex",
				ProviderURN:  "p_id:d_id",
				Type:         "tag",
			},
			ID:          "999",
			ResourceID:  "999",
			AccountType: expectedAccountType,
			AccountID:   expectedAccountID,
			Permissions: []string{"roles/datacatalog.categoryFineGrainedReader"},
		}

		actualError := p.RevokeAccess(pc, g)

		assert.Nil(t, actualError)
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
			Type:                "dataplex",
			URN:                 "test-URN",
			AllowedAccountTypes: []string{"user", "serviceAccount"},
			Credentials:         map[string]interface{}{},
			Resources: []*domain.ResourceConfig{
				{
					Type: "tag",
					Policy: &domain.PolicyConfig{
						ID:      "test-policy-1",
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
			Type:                "dataplex",
			URN:                 "test-URN",
			AllowedAccountTypes: []string{"user", "serviceAccount"},
			Credentials:         map[string]interface{}{},
			Resources: []*domain.ResourceConfig{
				{
					Type: "tag",
					Policy: &domain.PolicyConfig{
						ID:      "test-policy-1",
						Version: 1,
					},
					Roles: expectedRoles,
				},
			},
		}

		actualRoles, actualError := p.GetRoles(validConfig, "tag")

		assert.NoError(t, actualError)
		assert.Equal(t, expectedRoles, actualRoles)
	})
}

type DataplexProviderTestSuite struct {
	suite.Suite

	mockDataplexClient *mocks.DataplexClient
	//mockCloudLoggingClient *mocks.CloudLoggingClientI
	mockEncryptor  *mocks.Encryptor
	dummyProjectID string
	provider       *dataplex.Provider

	validProvider *domain.Provider
}

func TestDataplexProvider(t *testing.T) {
	suite.Run(t, new(DataplexProviderTestSuite))
}

func (s *DataplexProviderTestSuite) SetupTest() {
	s.mockDataplexClient = new(mocks.DataplexClient)
	s.mockEncryptor = new(mocks.Encryptor)
	s.provider = dataplex.NewProvider(domain.ProviderTypePolicyTag, s.mockEncryptor)
	s.dummyProjectID = "test-project-id"
	s.provider.Clients[s.dummyProjectID] = s.mockDataplexClient

	s.validProvider = &domain.Provider{
		Type: "dataplex",
		URN:  s.dummyProjectID,
		Config: &domain.ProviderConfig{
			Type:                "dataplex",
			URN:                 s.dummyProjectID,
			AllowedAccountTypes: []string{"user", "serviceAccount"},
			Credentials: map[string]interface{}{
				"resource_name":       fmt.Sprintf("projects/%s/location/us", s.dummyProjectID),
				"service_account_key": "dummy-credentials",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: "tag",
					Policy: &domain.PolicyConfig{
						ID:      "test-policy-1",
						Version: 1,
					},
					Roles: []*domain.Role{
						{
							ID:          "test-role",
							Name:        "test_role_name",
							Permissions: []interface{}{"roles/datacatalog.categoryFineGrainedReader"},
						},
					},
				},
			},
		},
	}

	s.mockEncryptor.On("Decrypt", "12345").Return(`{"type":"service_account"}`, nil) // tests the newIamClient when p.Clients is not initialised in the provider config
}

func (s *DataplexProviderTestSuite) TestListAccess() {
	s.Run("return error if initializing client fails", func() {
		s.mockEncryptor.EXPECT().Decrypt("invalid-key").Return("", errors.New("invalid-key")).Once()

		ctx := context.Background()
		_, err := s.provider.ListAccess(ctx, domain.ProviderConfig{
			Type: "dataplex",
			URN:  "new-urn",
			Credentials: map[string]interface{}{
				"service_account_key": "invalid-key",
				"resource_name":       "projects/project-name/locations/u",
			},
		}, []*domain.Resource{})

		s.EqualError(err, "initializing dataplex client: unable to decrypt credentials")
	})
}

func initProvider() *dataplex.Provider {
	crypto := new(mocks.Encryptor)
	return dataplex.NewProvider(domain.ProviderTypePolicyTag, crypto)
}
