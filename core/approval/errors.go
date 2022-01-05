package approval

import "errors"

var (
	ErrPolicyNotFound                 = errors.New("policy not found")
	ErrDependencyApprovalStepNotFound = errors.New("unable to resolve approval step dependency")
	ErrApprovalStepConditionNotFound  = errors.New("unable to resolve designated condition")
	ErrNilResourceInAppeal            = errors.New("unable to resolve resource from the appeal")
	ErrInvalidConditionField          = errors.New("invalid condition field")
)
