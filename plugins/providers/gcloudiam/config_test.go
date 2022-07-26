package gcloudiam_test

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/plugins/providers/gcloudiam"
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

