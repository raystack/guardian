package appeal

import "errors"

var (
	ErrProviderTypeNotFound = errors.New("provider is not registered")
	ErrProviderURNNotFound  = errors.New("provider with specified urn is not registered")
	ErrPolicyConfigNotFound = errors.New("unable to find matching approval policy config for specified resource")
)
