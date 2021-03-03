package policy

import "errors"

var (
	// ErrEmptyIDParam is the error value if the policy id is empty
	ErrEmptyIDParam = errors.New("id can't be empty")
	// ErrPolicyDoesNotExists is the error value if the designated policy is not exists
	ErrPolicyDoesNotExists = errors.New("policy does not exists")
)
