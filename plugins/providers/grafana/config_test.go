package grafana_test

import (
	"errors"
	"testing"

	"github.com/goto/guardian/mocks"
	"github.com/goto/guardian/plugins/providers/grafana"
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
			username := "username"
			password := "password"
			creds := grafana.Credentials{
				Host:     "http://localhost:4000",
				Username: username,
				Password: password,
			}

			expectedEncryptedApiKey := "encrypted_api_key"
			encryptor.On("Encrypt", password).Return(expectedEncryptedApiKey, nil).Once()

			actualError := creds.Encrypt(encryptor)

			assert.Nil(t, actualError)
			assert.Equal(t, expectedEncryptedApiKey, creds.Password)
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
			username := "username"
			password := "encrypted_password"
			creds := grafana.Credentials{
				Host:     "http://localhost:3000",
				Username: username,
				Password: password,
			}

			expectedDecryptedApiKey := "decrypted_api_key"
			decryptor.On("Decrypt", password).Return(expectedDecryptedApiKey, nil).Once()

			actualError := creds.Decrypt(decryptor)

			assert.Nil(t, actualError)
			assert.Equal(t, expectedDecryptedApiKey, creds.Password)
		})
	})
}
