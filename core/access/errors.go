package access

import "errors"

var (
	ErrEmptyIDParam   = errors.New("access id can't be empty")
	ErrAccessNotFound = errors.New("access not found")
)
