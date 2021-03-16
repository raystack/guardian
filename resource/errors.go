package resource

import "errors"

var (
	// ErrEmptyIDParam is the error value if the resource id is empty
	ErrEmptyIDParam = errors.New("id can't be empty")
)
