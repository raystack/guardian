package gcs

import "errors"

var (
	ErrInvalidPermissionConfig       = errors.New("invalid permission config type")
	ErrUnableToDecryptNilCredentials = errors.New("unable to decrypt nil credentials")

	ErrInvalidResourceType           = errors.New("invalid resource type")
	ErrUnableToEncryptNilCredentials = errors.New("unable to encrypt nil credentials")
	ErrInvalidCredentialsType        = errors.New("invalid credentials type")

	ErrNilProviderConfig    = errors.New("provider config can't be nil")
	ErrNilAppeal            = errors.New("appeal can't be nil")
	ErrNilResource          = errors.New("designated resource can't be nil")
	ErrProviderTypeMismatch = errors.New("provider type in the config and in the appeal don't match")
	ErrProviderURNMismatch  = errors.New("provider urn in the config and in the appeal don't match")

	ErrInvalidRole             = errors.New("invalid role")
	ErrPermissionAlreadyExists = errors.New("permission already exists")
)
