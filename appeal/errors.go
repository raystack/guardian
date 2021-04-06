package appeal

import "errors"

var (
	ErrAppealIDEmptyParam = errors.New("appeal id is required")

	ErrAppealStatusApproved     = errors.New("appeal already approved")
	ErrAppealStatusRejected     = errors.New("appeal already rejected")
	ErrAppealStatusTerminated   = errors.New("appeal already terminated")
	ErrAppealStatusUnrecognized = errors.New("unrecognized appeal status")

	ErrApprovalDependencyIsPending = errors.New("found previous approval step that is still in pending")
	ErrApprovalStatusApproved      = errors.New("approval already approved")
	ErrApprovalStatusRejected      = errors.New("approval already rejected")
	ErrApprovalStatusSkipped       = errors.New("approval already skipped")
	ErrApprovalStatusUnrecognized  = errors.New("unrecognized approval status")
	ErrApprovalNameNotFound        = errors.New("approval step name not found")

	ErrActionForbidden    = errors.New("user is not allowed to make action on this approval step")
	ErrActionInvalidValue = errors.New("invalid action value")

	ErrProviderTypeNotFound  = errors.New("provider is not registered")
	ErrProviderURNNotFound   = errors.New("provider with specified urn is not registered")
	ErrPolicyConfigNotFound  = errors.New("unable to find matching approval policy config for specified resource")
	ErrPolicyIDNotFound      = errors.New("unable to find approval policy for specified id")
	ErrPolicyVersionNotFound = errors.New("unable to find approval policy for specified version")

	ErrApproverKeyNotRecognized = errors.New("unrecognized approvers key")
	ErrApproverInvalidType      = errors.New("invalid approver type, expected an email or array of email")
)
