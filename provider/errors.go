package provider

import "errors"

var (
	// ErrInvalidProviderType is the error value if provider is unable to find the matching provider type
	ErrInvalidProviderType = errors.New("unable to find provider based on provider type")
)
