package dataplex

import "errors"

var (

	// ErrInvalidPermissionConfig is the error value for invalid permission config
	ErrInvalidPermissionConfig = errors.New("invalid permission config type")
	// ErrUnableToEncryptNilCredentials is the error value if the to be encrypted credentials is nil
	ErrUnableToEncryptNilCredentials = errors.New("unable to encrypt nil credentials")
	// ErrUnableToDecryptNilCredentials is the error value if the to be decrypted credentials is nil
	ErrUnableToDecryptNilCredentials = errors.New("unable to decrypt nil credentials")
	ErrUnableToDecryptCredentials    = errors.New("unable to decrypt credentials")
	// ErrInvalidCredentialsType is the error value if the credentials value can't be casted into the bigquery.Credentials type
	ErrInvalidCredentialsType    = errors.New("invalid credentials type")
	ErrInvalidResourceFormatType = errors.New("invalid resource-name format, it should be projects/{project_id}/locations/{location}")
	ErrInvalidRole               = errors.New("invalid role")
	ErrInvalidResourceType       = errors.New("invalid resource type")

	ErrPermissionAlreadyExists = errors.New("permission already exists")
	ErrPermissionNotFound      = errors.New("permission not found")
	ErrNilProviderConfig       = errors.New("provider config can't be nil")

	ErrNilResource              = errors.New("designated resource can't be nil")
	ErrProviderTypeMismatch     = errors.New("provider type in the config and in the appeal don't match")
	ErrProviderURNMismatch      = errors.New("provider urn in the config and in the appeal don't match")
	ErrInvalidDatasetPermission = errors.New("provided permission is not supported for dataset resource")
)
