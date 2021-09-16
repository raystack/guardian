package policy

import "errors"

var (
	// ErrEmptyIDParam is the error value if the policy id is empty
	ErrEmptyIDParam = errors.New("id can't be empty")
	// ErrPolicyDoesNotExists is the error value if the designated policy is not exists
	ErrPolicyDoesNotExists         = errors.New("policy does not exists")
	ErrIDContainsWhitespaces       = errors.New("id should not contain whitespaces")
	ErrStepNameContainsWhitespaces = errors.New("step name should not contain whitespaces")
	ErrInvalidApprovers            = errors.New("invalid approvers")
	ErrStepDependencyDoesNotExists = errors.New("step dependency does not exists")
)
