package grafana_test

import (
	"errors"
	"testing"

	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/provider/grafana"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCredentials(t *testing.T) {
	encryptor := new(mocks.Encryptor)
	decryptor := new(mocks.Decryptor)

	t.Run("encrypt", func(t *testing.T) {
		t.Run("should return error if creds is nil", func(t *testing.T) {
			var creds *grafana.Credentials
			expectedError := grafana.ErrUnableToEncryptNilCredentials

			actualError := creds.Encrypt(encryptor)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return error if encryptor failed to encrypt the creds", func(t *testing.T) {
			creds := grafana.Credentials{}
			expectedError := errors.New("encryptor error")
			encryptor.On("Encrypt", mock.Anything).Return("", expectedError).Once()

			actualError := creds.Encrypt(encryptor)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return encrypted password inside Credentials on success", func(t *testing.T) {
			api_key := "test_api_key"
			creds := grafana.Credentials{
				Host:   "http://localhost:4000",
				ApiKey: api_key,
			}

			expectedEncryptedApiKey := "encrypted_api_key"
			encryptor.On("Encrypt", api_key).Return(expectedEncryptedApiKey, nil).Once()

			actualError := creds.Encrypt(encryptor)

			assert.Nil(t, actualError)
			assert.Equal(t, expectedEncryptedApiKey, creds.ApiKey)
		})
	})

	t.Run("decrypt", func(t *testing.T) {
		t.Run("should return error if creds is nil", func(t *testing.T) {
			var creds *grafana.Credentials

			expectedError := grafana.ErrUnableToDecryptNilCredentials

			actualError := creds.Decrypt(decryptor)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return error if decryptor failed to decrypt the creds", func(t *testing.T) {
			creds := grafana.Credentials{}
			expectedError := errors.New("decryptor error")
			decryptor.On("Decrypt", mock.Anything).Return("", expectedError).Once()

			actualError := creds.Decrypt(decryptor)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return decrypted password inside Credentials on success", func(t *testing.T) {
			api_key := "encrypted_api_key"
			creds := grafana.Credentials{
				Host:   "http://localhost:3000",
				ApiKey: api_key,
			}

			expectedDecryptedApiKey := "decrypted_api_key"
			decryptor.On("Decrypt", api_key).Return(expectedDecryptedApiKey, nil).Once()

			actualError := creds.Decrypt(decryptor)

			assert.Nil(t, actualError)
			assert.Equal(t, expectedDecryptedApiKey, creds.ApiKey)
		})
	})
}
