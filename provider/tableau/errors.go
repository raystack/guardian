package tableau

import "errors"

var (
	ErrInvalidCredentials            = errors.New("invalid credentials type")
	ErrInvalidPermissionConfig       = errors.New("invalid permission config type")
	ErrUnableToEncryptNilCredentials = errors.New("unable to encrypt nil credentials")
	ErrUnableToDecryptNilCredentials = errors.New("unable to decrypt nil credentials")
)
