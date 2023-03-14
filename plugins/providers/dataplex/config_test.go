package dataplex_test

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/goto/guardian/plugins/providers/bigquery"

	"github.com/goto/guardian/domain"
	"github.com/goto/guardian/mocks"
	"github.com/goto/guardian/plugins/providers/dataplex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCredentials(t *testing.T) {
	encryptor := new(mocks.Encryptor)
	decryptor := new(mocks.Decryptor)

	t.Run("encrypt", func(t *testing.T) {
		t.Run("should return error if creds is nil", func(t *testing.T) {
			var creds *dataplex.Credentials
			expectedError := dataplex.ErrUnableToEncryptNilCredentials

			actualError := creds.Encrypt(encryptor)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return error if encryptor failed to encrypt the creds", func(t *testing.T) {
			creds := dataplex.Credentials{}
			expectedError := errors.New("encryptor error")
			encryptor.On("Encrypt", mock.Anything).Return("", expectedError).Once()

			actualError := creds.Encrypt(encryptor)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return encrypted password inside Credentials on success", func(t *testing.T) {
			service_account_key := base64.StdEncoding.EncodeToString([]byte("service-account-key-json"))
			resourceName := "resource_name"
			creds := bigquery.Credentials{
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
			var creds *dataplex.Credentials

			expectedError := dataplex.ErrUnableToDecryptNilCredentials

			actualError := creds.Decrypt(decryptor)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return error if decryptor failed to decrypt the creds", func(t *testing.T) {
			creds := dataplex.Credentials{}
			expectedError := errors.New("decryptor error")
			decryptor.On("Decrypt", mock.Anything).Return("", expectedError).Once()

			actualError := creds.Decrypt(decryptor)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return decrypted service account key on success", func(t *testing.T) {
			service_account_key := base64.StdEncoding.EncodeToString([]byte("service-account-key-json"))
			resourceName := "resource_name"
			creds := bigquery.Credentials{
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

func TestNewConfig(t *testing.T) {
	t.Run("should return dataplex config containing the same provider config", func(t *testing.T) {
		mockCrypto := new(mocks.Crypto)
		pc := &domain.ProviderConfig{}
		expectedProviderConfig := pc

		c := dataplex.NewConfig(pc, mockCrypto)
		actualProviderConfig := c.ProviderConfig

		assert.NotNil(t, c)
		assert.Equal(t, expectedProviderConfig, actualProviderConfig)
	})
}

func TestValidate(t *testing.T) {
	mockCrypto := new(mocks.Crypto)
	validCredentials := dataplex.Credentials{
		ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service-account-key-json")),
		ResourceName:      "projects/project-name/location/us",
	}
	credentialsWithoutBaseEncodedSAKey := dataplex.Credentials{
		ServiceAccountKey: "non-base64-value",
		ResourceName:      "projects/project-name/location/us",
	}
	credentialsWithoutResourceName := dataplex.Credentials{
		ServiceAccountKey: base64.StdEncoding.EncodeToString([]byte("service-account-key-json")),
	}
	validPermissionConfig := "permission-name"

	t.Run("error validations", func(t *testing.T) {
		testCases := []struct {
			name             string
			credentials      interface{}
			permissionConfig interface{}
		}{
			{
				name:             "should return error if service account key of credentials is not a base64 string",
				credentials:      credentialsWithoutBaseEncodedSAKey,
				permissionConfig: validPermissionConfig,
			},
			{
				name:             "should return error if permission type is invalid",
				credentials:      validCredentials,
				permissionConfig: 0,
			},
			{
				name:             "should return error if credentials config does not contain resource name field",
				credentials:      credentialsWithoutResourceName,
				permissionConfig: validPermissionConfig,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				pc := &domain.ProviderConfig{
					Credentials: tc.credentials,
					Resources: []*domain.ResourceConfig{
						{
							Roles: []*domain.Role{
								{
									Permissions: []interface{}{tc.permissionConfig},
								},
							},
						},
					},
				}
				mockCrypto.On("Encrypt", mock.Anything).Return("", nil).Once()

				err := bigquery.NewConfig(pc, mockCrypto).ParseAndValidate()
				assert.Error(t, err)
			})
		}
	})

	/*t.Run("should update credentials and permission config values into castable bigquery config", func(t *testing.T) {
		pc := &domain.ProviderConfig{
			Credentials: validCredentials,
			Resources: []*domain.ResourceConfig{
				{
					Roles: []*domain.Role{
						{
							Permissions: []interface{}{validPermissionConfig},
						},
					},
				},
			},
		}
		mockCrypto.On("Encrypt", mock.Anything).Return("", nil).Once()

		err := bigquery.NewConfig(pc, mockCrypto).ParseAndValidate()
		_, credentialsOk := pc.Credentials.(*bigquery.Credentials)
		_, permissionConfigOk := pc.Resources[0].Roles[0].Permissions[0].(*bigquery.PermissionConfig)

		assert.Nil(t, err)
		assert.True(t, credentialsOk)
		assert.True(t, permissionConfigOk)
	})*/
}
