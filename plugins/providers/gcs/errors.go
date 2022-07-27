package gcs

import "errors"

var (
	ErrInvalidPermissionConfig       = errors.New("invalid permission config type")
	ErrUnableToDecryptNilCredentials = errors.New("unable to decrypt nil credentials")

	ErrInvalidResourceType = errors.New("invalid resource type")
)
