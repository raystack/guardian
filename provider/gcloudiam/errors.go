package gcloudiam

import "errors"

var (
	ErrUnableToEncryptNilCredentials = errors.New("unable to encrypt nil credentials")
	ErrUnableToDecryptNilCredentials = errors.New("unable to decrypt nil credentials")
	ErrInvalidPermissionConfig       = errors.New("invalid permission config type")
	ErrInvalidCredentials            = errors.New("invalid credentials type")
	ErrPermissionAlreadyExists       = errors.New("permission already exists")
	ErrPermissionNotFound            = errors.New("permission not found")
	ErrInvalidResourceType           = errors.New("invalid resource type")
	ErrInvalidRole                   = errors.New("invalid role")
)
