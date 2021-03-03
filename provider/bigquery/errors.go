package bigquery

import "errors"

var (
	// ErrInvalidCredentials is the error value for invalid credentials
	ErrInvalidCredentials = errors.New("invalid credentials type")
	// ErrInvalidPermissionConfig is the error value for invalid permission config
	ErrInvalidPermissionConfig = errors.New("invalid permission config type")
)
