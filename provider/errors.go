package provider

import "errors"

var (
	// ErrInvalidProviderType is the error value if provider is unable to find the matching provider type
	ErrInvalidProviderType = errors.New("unable to find provider based on provider type")
	// ErrEmptyIDParam is the error value if the policy id is empty
	ErrEmptyIDParam = errors.New("id can't be empty")
	// ErrRecordNotFound is the error value if the designated record id is not exists
	ErrRecordNotFound = errors.New("record not found")
)
