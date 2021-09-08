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
	ErrInvalidCredentialsType      = errors.New("invalid credentials type")
	ErrInvalidRole                 = errors.New("invalid role")
	ErrInvalidResourceType         = errors.New("invalid resource type")
	ErrInvalidTableURN             = errors.New("table URN is invalid")
	ErrPermissionAlreadyExists     = errors.New("permission already exists")
	ErrPermissionNotFound          = errors.New("permission not found")
	ErrNilProviderConfig           = errors.New("provider config can't be nil")
	ErrNilAppeal                   = errors.New("appeal can't be nil")
	ErrNilResource                 = errors.New("designated resource can't be nil")
	ErrProviderTypeMismatch        = errors.New("provider type in the config and in the appeal don't match")
	ErrProviderURNMismatch         = errors.New("provider urn in the config and in the appeal don't match")
	ErrInvalidDatasetPermission    = errors.New("provided permission is not supported for dataset resource")
	ErrInvalidTablePermission      = errors.New("provided permission is not supported for table resource")
	ErrEmptyResource               = errors.New("this bigquery project has no resources")
	ErrCannotVerifyTablePermission = errors.New("cannot verify the table permissions since this bigquery project does not have any tables")
)
