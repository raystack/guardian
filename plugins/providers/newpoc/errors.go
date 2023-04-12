package newpoc

import "errors"

var (
	ErrUnableToEncryptNilCredentials = errors.New("unable to encrypt nil credentials")
	ErrUnableToDecryptNilCredentials = errors.New("unable to decrypt nil credentials")
	ErrInvalidPermissionConfig       = errors.New("invalid permission config type")
	ErrPermissionAlreadyExists       = errors.New("permission already exists")
	ErrPermissionNotFound            = errors.New("permission not found")
	ErrInvalidResourceType           = errors.New("invalid resource type")
	ErrInvalidRole                   = errors.New("invalid role")
	ErrInvalidResourceName           = errors.New("invalid resource name: resource name should be projects/{{project-id}} or organizations/{{org-id}}")
	ErrInvalidProjectRole            = errors.New("provided role is not supported for project in gcloud")
	ErrInvalidProviderType           = errors.New("invalid provider type")
)
