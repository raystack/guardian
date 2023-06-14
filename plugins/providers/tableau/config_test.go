package tableau_test

import (
	"errors"
	"testing"

	"github.com/raystack/guardian/mocks"
	"github.com/raystack/guardian/plugins/providers/tableau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCredentials(t *testing.T) {
	encryptor := new(mocks.Encryptor)
	decryptor := new(mocks.Decryptor)

	t.Run("encrypt", func(t *testing.T) {
		t.Run("should return error if creds is nil", func(t *testing.T) {
			var creds *tableau.Credentials
			expectedError := tableau.ErrUnableToEncryptNilCredentials

			actualError := creds.Encrypt(encryptor)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return error if encryptor failed to encrypt the creds", func(t *testing.T) {
			creds := tableau.Credentials{}
			expectedError := errors.New("encryptor error")
			encryptor.On("Encrypt", mock.Anything).Return("", expectedError).Once()

			actualError := creds.Encrypt(encryptor)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return encrypted password inside Credentials on success", func(t *testing.T) {
			username := "test_user_name"
			password := "test_password"
			creds := tableau.Credentials{
				Host:     "http://localhost:4000",
				Username: username,
				Password: password,
			}

			expectedEncryptedPassword := "encrypted_password"
			encryptor.On("Encrypt", password).Return(expectedEncryptedPassword, nil).Once()

			actualError := creds.Encrypt(encryptor)

			assert.Nil(t, actualError)
			assert.Equal(t, expectedEncryptedPassword, creds.Password)
		})
	})

	t.Run("decrypt", func(t *testing.T) {
		t.Run("should return error if creds is nil", func(t *testing.T) {
			var creds *tableau.Credentials

			expectedError := tableau.ErrUnableToDecryptNilCredentials

			actualError := creds.Decrypt(decryptor)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return error if decryptor failed to decrypt the creds", func(t *testing.T) {
			creds := tableau.Credentials{}
			expectedError := errors.New("decryptor error")
			decryptor.On("Decrypt", mock.Anything).Return("", expectedError).Once()

			actualError := creds.Decrypt(decryptor)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return decrypted password inside Credentials on success", func(t *testing.T) {
			username := "test_user_name"
			password := "test_password"
			creds := tableau.Credentials{
				Host:     "http://localhost:4000",
				Username: username,
				Password: password,
			}

			expectedDecryptedPassword := "decrypted_password"
			decryptor.On("Decrypt", password).Return(expectedDecryptedPassword, nil).Once()

			actualError := creds.Decrypt(decryptor)

			assert.Nil(t, actualError)
			assert.Equal(t, expectedDecryptedPassword, creds.Password)
		})
	})
}
