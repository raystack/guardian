package grafana

import "errors"

var (
	ErrInvalidRole                   = errors.New("invalid role")
	ErrInvalidPermissionType         = errors.New("invalid permission type")
	ErrUserNotFound                  = errors.New("cannot find user with the given email")
	ErrInvalidResourceType           = errors.New("invalid resource type")
	ErrInvalidCredentials            = errors.New("invalid credentials type")
	ErrInvalidPermissionConfig       = errors.New("invalid permission config type")
	ErrUnableToEncryptNilCredentials = errors.New("unable to encrypt nil credentials")
	ErrUnableToDecryptNilCredentials = errors.New("unable to decrypt nil credentials")
)
