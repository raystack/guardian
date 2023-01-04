package approval

import "errors"

var (
	ErrPolicyNotFound                 = errors.New("policy not found")
	ErrDependencyApprovalStepNotFound = errors.New("unable to resolve approval step dependency")
	ErrApprovalStepConditionNotFound  = errors.New("unable to resolve designated condition")
	ErrNilResourceInAppeal            = errors.New("unable to resolve resource from the appeal")
	ErrInvalidConditionField          = errors.New("invalid condition field")

	ErrAppealStatusCanceled           = errors.New("appeal already canceled")
	ErrAppealStatusApproved           = errors.New("appeal already approved")
	ErrAppealStatusRejected           = errors.New("appeal already rejected")
	ErrAppealStatusBlocked            = errors.New("approval is blocked")
	ErrAppealStatusUnrecognized       = errors.New("unrecognized appeal status")
	ErrAppealDuplicate                = errors.New("appeal with the same resource and role already exists")
	ErrAppealInvalidExtensionDuration = errors.New("invalid appeal extension duration")
	ErrAppealFoundActiveGrant         = errors.New("user still have an active grant")
	ErrGrantNotEligibleForExtension   = errors.New("existing grant is not eligible for extension")
	ErrCannotCreateAppealForOtherUser = errors.New("creating appeal for other individual user (account_type=\"user\") is not allowed")

	ErrApprovalDependencyIsBlocked = errors.New("found previous approval step that is still in blocked")
	ErrApprovalDependencyIsPending = errors.New("found previous approval step that is still in pending")
	ErrApprovalStatusApproved      = errors.New("approval already approved")
	ErrApprovalStatusRejected      = errors.New("approval already rejected")
	ErrApprovalStatusSkipped       = errors.New("approval already skipped")
	ErrApprovalStatusUnrecognized  = errors.New("unrecognized approval status")
	ErrApprovalNotFound            = errors.New("approval not found")
	ErrUnableToAddApprover         = errors.New("unable to add a new approver")
	ErrUnableToDeleteApprover      = errors.New("unable to remove approver")

	ErrActionForbidden    = errors.New("user is not allowed to make action on this approval step")
	ErrActionInvalidValue = errors.New("invalid action value")
)
