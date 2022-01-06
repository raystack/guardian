package resource

import "errors"

var (
	// ErrEmptyIDParam is the error value if the resource id is empty
	ErrEmptyIDParam = errors.New("id can't be empty")
	// ErrRecordNotFound is the error value if the designated record id is not exists
	ErrRecordNotFound = errors.New("record not found")
)
