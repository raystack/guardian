package bigquery

import "errors"

var (
	// ErrInvalidCredentials is the error value for invalid credentials
	ErrInvalidCredentials = errors.New("invalid credentials type")
	// ErrInvalidPermissionConfig is the error value for invalid permission config
	ErrInvalidPermissionConfig = errors.New("invalid permission config type")
	// ErrUnableToEncryptNilCredentials is the error value if the to be encrypted credentials is nil
	ErrUnableToEncryptNilCredentials = errors.New("unable to encrypt nil credentials")
	// ErrUnableToDecryptNilCredentials is the error value if the to be decrypted credentials is nil
	ErrUnableToDecryptNilCredentials = errors.New("unable to decrypt nil credentials")
	// ErrInvalidCredentialsType is the error value if the credentials value can't be casted into the bigquery.Credentials type
	ErrInvalidCredentialsType = errors.New("invalid credentials type")
)
