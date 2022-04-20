package metabase

import "errors"

var (
	ErrInvalidCredentials            = errors.New("invalid credentials type")
	ErrInvalidPermissionConfig       = errors.New("invalid permission config type")
	ErrUnableToEncryptNilCredentials = errors.New("unable to encrypt nil credentials")
	ErrUnableToDecryptNilCredentials = errors.New("unable to decrypt nil credentials")
	ErrUserNotFound                  = errors.New("metabase user not found")
	ErrInvalidRole                   = errors.New("invalid role")
	ErrInvalidResourceType           = errors.New("invalid resource type")
	ErrPermissionNotFound            = errors.New("permission not found")
	ErrInvalidApiResponse            = errors.New("invalid api response")
	ErrInvalidDatabaseURN            = errors.New("database URN is invalid")
	ErrInvalidTableURN               = errors.New("table URN is invalid")
)
