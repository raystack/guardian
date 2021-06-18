package metabase_test

import (
	"errors"
	"testing"

	"github.com/odpf/guardian/mocks"
	"github.com/odpf/guardian/provider/metabase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCredentials(t *testing.T) {
	encryptor := new(mocks.Encryptor)
	decryptor := new(mocks.Decryptor)

	t.Run("encrypt", func(t *testing.T) {
		t.Run("should return error if creds is nil", func(t *testing.T) {
			var creds *metabase.Credentials
			expectedError := metabase.ErrUnableToEncryptNilCredentials

			actualError := creds.Encrypt(encryptor)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return error if encryptor failed to encrypt the creds", func(t *testing.T) {
			creds := metabase.Credentials{}
			expectedError := errors.New("encryptor error")
			encryptor.On("Encrypt", mock.Anything).Return("", expectedError).Once()

			actualError := creds.Encrypt(encryptor)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return encrypt the password inside Credentials on success", func(t *testing.T) {
			password := "password"
			creds := metabase.Credentials{
				Host:     "http://localhost:3000",
				Username: "test@email.com",
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
			var creds *metabase.Credentials

			expectedError := metabase.ErrUnableToDecryptNilCredentials

			actualError := creds.Decrypt(decryptor)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return error if decryptor failed to decrypt the creds", func(t *testing.T) {
			creds := metabase.Credentials{}
			expectedError := errors.New("decryptor error")
			decryptor.On("Decrypt", mock.Anything).Return("", expectedError).Once()

			actualError := creds.Decrypt(decryptor)

			assert.EqualError(t, actualError, expectedError.Error())
		})

		t.Run("should return encrypt the password inside Credentials on success", func(t *testing.T) {
			password := "encrypted_password"
			creds := metabase.Credentials{
				Host:     "http://localhost:3000",
				Username: "test@email.com",
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
