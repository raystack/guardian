package policy

import "errors"

var (
	// ErrEmptyIDParam is the error value if the policy id is empty
	ErrEmptyIDParam = errors.New("id can't be empty")
	// ErrPolicyNotFound is the error value if the designated policy not found
	ErrPolicyNotFound              = errors.New("policy not found")
	ErrIDContainsWhitespaces       = errors.New("id should not contain whitespaces")
	ErrStepNameContainsWhitespaces = errors.New("step name should not contain whitespaces")
	ErrInvalidApprovers            = errors.New("invalid approvers")
	ErrStepDependencyDoesNotExists = errors.New("step dependency does not exists")
)
