package provider

import "errors"

var (
	// ErrInvalidProviderType is the error value if provider is unable to find the matching provider type
	ErrInvalidProviderType = errors.New("unable to find provider based on provider type")
	// ErrEmptyIDParam is the error value if the policy id is empty
	ErrEmptyIDParam = errors.New("id can't be empty")
	// ErrRecordNotFound is the error value if the designated record id is not exists
	ErrRecordNotFound                      = errors.New("record not found")
	ErrEmptyProviderType                   = errors.New("provider type can't be nil")
	ErrEmptyProviderURN                    = errors.New("provider urn can't be nil")
	ErrNilAppeal                           = errors.New("appeal can't be nil")
	ErrNilResource                         = errors.New("resource can't be nil")
	ErrProviderNotFound                    = errors.New("provider config not found")
	ErrInvalidResourceType                 = errors.New("invalid resource type")
	ErrInvalidRole                         = errors.New("invalid role")
	ErrExpirationDateIsRequired            = errors.New("having permanent access to this resource is not allowed, access duration is required")
	ErrOptionsExpirationDateOptionNotFound = errors.New("expiration date is required, unable to find expiration date option")
)
