package gcloudiam_test

import (
	"encoding/base64"
	"errors"
	"github.com/goto/guardian/domain"
	"github.com/goto/guardian/pkg/crypto"
	"testing"

	"github.com/goto/guardian/mocks"
	"github.com/goto/guardian/plugins/providers/gcloudiam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCredentials(t *testing.T) {
	encryptor := new(mocks.Encryptor)
	decryptor := new(mocks.Decryptor)

	t.Run("encrypt", func(t *testing.T) {
		t.Run("should return error if creds is nil", func(t *testing.T) {
			var creds *gcloudiam.Credentials
			expectedError := gcloudiam.ErrUnableToEncryptNilCredentials

			actualError := creds.Encrypt(encryptor)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return error if encryptor failed to encrypt the creds", func(t *testing.T) {
			creds := gcloudiam.Credentials{}
			expectedError := errors.New("encryptor error")
			encryptor.On("Encrypt", mock.Anything).Return("", expectedError).Once()

			actualError := creds.Encrypt(encryptor)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return encrypted password inside Credentials on success", func(t *testing.T) {
			service_account_key := base64.StdEncoding.EncodeToString([]byte("service-account-key-json"))
			resourceName := "resource_name"
			creds := gcloudiam.Credentials{
				ServiceAccountKey: service_account_key,
				ResourceName:      resourceName,
			}

			expectedEncryptedServiceAccountKey := "encrypted_service_account_key"
			encryptor.On("Encrypt", service_account_key).Return(expectedEncryptedServiceAccountKey, nil).Once()

			actualError := creds.Encrypt(encryptor)

			assert.Nil(t, actualError)
			assert.Equal(t, expectedEncryptedServiceAccountKey, creds.ServiceAccountKey)
			encryptor.AssertExpectations(t)
		})
	})

	t.Run("decrypt", func(t *testing.T) {
		t.Run("should return error if creds is nil", func(t *testing.T) {
			var creds *gcloudiam.Credentials

			expectedError := gcloudiam.ErrUnableToDecryptNilCredentials

			actualError := creds.Decrypt(decryptor)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return error if decryptor failed to decrypt the creds", func(t *testing.T) {
			creds := gcloudiam.Credentials{}
			expectedError := errors.New("decryptor error")
			decryptor.On("Decrypt", mock.Anything).Return("", expectedError).Once()

			actualError := creds.Decrypt(decryptor)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return decrypted service account key on success", func(t *testing.T) {
			service_account_key := base64.StdEncoding.EncodeToString([]byte("service-account-key-json"))
			resourceName := "resource_name"
			creds := gcloudiam.Credentials{
				ServiceAccountKey: service_account_key,
				ResourceName:      resourceName,
			}
			expectedDecryptedServiceAccountKey := "decrypted_service_account_key"
			decryptor.On("Decrypt", service_account_key).Return(expectedDecryptedServiceAccountKey, nil).Once()

			actualError := creds.Decrypt(decryptor)

			assert.Nil(t, actualError)
			assert.Equal(t, expectedDecryptedServiceAccountKey, creds.ServiceAccountKey)
			decryptor.AssertExpectations(t)
		})
	})
}

func TestConfig_ParseAndValidate(t *testing.T) {

	t.Run("should return error if resource config is nil", func(t *testing.T) {
		crypo := crypto.NewAES("encryption_key")
		providerConfig := domain.ProviderConfig{
			Type: "gcloudiam",
			URN:  "test-urn",
			Credentials: gcloudiam.Credentials{
				ServiceAccountKey: getBase64EncodedString(),
				ResourceName:      "projects/test-project",
			},
			Resources: nil,
		}
		config := gcloudiam.NewConfig(&providerConfig, crypo)

		actualErr := config.ParseAndValidate()
		assert.EqualError(t, actualErr, "empty resource config")
	})

	t.Run("should return error if service account key is not base64", func(t *testing.T) {
		crypo := crypto.NewAES("encryption_key")
		providerConfig := domain.ProviderConfig{
			Type: "gcloudiam",
			URN:  "test-urn",
			Credentials: gcloudiam.Credentials{
				ServiceAccountKey: "test-service-account-key",
				ResourceName:      "projects/test-project",
			},
			Resources: nil,
		}
		config := gcloudiam.NewConfig(&providerConfig, crypo)

		actualErr := config.ParseAndValidate()
		assert.NotNil(t, actualErr)
	})
	t.Run("should return error if duplicate resource type is present", func(t *testing.T) {
		crypo := crypto.NewAES("encryption_key")
		providerConfig := domain.ProviderConfig{
			Type: "gcloudiam",
			URN:  "test-urn",
			Credentials: gcloudiam.Credentials{
				ServiceAccountKey: getBase64EncodedString(),
				ResourceName:      "projects/test-project",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: "project",
					Roles: []*domain.Role{
						{ID: "test-roleA", Name: "test role A", Permissions: []interface{}{"test-permission-a"}},
					},
				},
				{
					Type: "project",
					Roles: []*domain.Role{
						{ID: "test-roleB", Name: "test role B", Permissions: []interface{}{"test-permission-b"}},
					}},
			},
		}
		expectedErr := "duplicate resource type: \"project\""
		config := gcloudiam.NewConfig(&providerConfig, crypo)

		actualErr := config.ParseAndValidate()
		assert.Equal(t, expectedErr, actualErr.Error())
	})

	t.Run("should return error for invalid resource type", func(t *testing.T) {
		crypo := crypto.NewAES("encryption_key")
		providerConfig := domain.ProviderConfig{
			Type: "gcloudiam",
			URN:  "test-urn",
			Credentials: gcloudiam.Credentials{
				ServiceAccountKey: getBase64EncodedString(),
				ResourceName:      "projects/test-project",
			},
			Resources: []*domain.ResourceConfig{
				{Type: "invalid-resource-type"},
			},
		}
		expectedErr := "invalid resource type: \"invalid-resource-type\"\ngcloud_iam provider should not have empty roles"
		config := gcloudiam.NewConfig(&providerConfig, crypo)

		actualErr := config.ParseAndValidate()
		assert.Equal(t, expectedErr, actualErr.Error())
	})

	t.Run("should return nil if config is valid", func(t *testing.T) {
		crypo := crypto.NewAES("encryption_key")
		providerConfig := domain.ProviderConfig{
			Type: "gcloudiam",
			URN:  "test-urn",
			Credentials: gcloudiam.Credentials{
				ServiceAccountKey: getBase64EncodedString(),
				ResourceName:      "projects/test-project",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: "project",
					Roles: []*domain.Role{
						{ID: "test-role", Name: "test role", Permissions: []interface{}{"test-permission"}},
					},
				},
			},
		}
		config := gcloudiam.NewConfig(&providerConfig, crypo)

		actualErr := config.ParseAndValidate()
		assert.Nil(t, actualErr)
	})

	t.Run("should return error if duplicate role is configured", func(t *testing.T) {
		crypo := crypto.NewAES("encryption_key")
		providerConfig := domain.ProviderConfig{
			Type: "gcloudiam",
			URN:  "test-urn",
			Credentials: gcloudiam.Credentials{
				ServiceAccountKey: getBase64EncodedString(),
				ResourceName:      "projects/test-project",
			},
			Resources: []*domain.ResourceConfig{
				{
					Type: "project",
					Roles: []*domain.Role{
						{ID: "test-role1", Name: "test role 1", Permissions: []interface{}{"test-permission-1"}},
						{ID: "test-role2", Name: "test role 2", Permissions: []interface{}{"test-permission-2"}},
						{ID: "test-role1", Name: "test role 1", Permissions: []interface{}{"test-permission-11"}},
						{ID: "test-role3", Name: "test role 3", Permissions: []interface{}{"test-permission-3"}},
					},
				},
			},
		}

		expectedErr := "duplicate role: \"test-role1\""

		config := gcloudiam.NewConfig(&providerConfig, crypo)

		actualErr := config.ParseAndValidate()
		assert.EqualError(t, actualErr, expectedErr)
	})

}

func getBase64EncodedString() string {
	return base64.StdEncoding.EncodeToString([]byte("test-service-account-key"))
}
