package appeal

import "errors"

var (
	ErrAppealIDEmptyParam = errors.New("appeal id is required")

	ErrAppealStatusCanceled           = errors.New("appeal already canceled")
	ErrAppealStatusApproved           = errors.New("appeal already approved")
	ErrAppealStatusRejected           = errors.New("appeal already rejected")
	ErrAppealStatusTerminated         = errors.New("appeal already terminated")
	ErrAppealStatusBlocked            = errors.New("approval is blocked")
	ErrAppealStatusUnrecognized       = errors.New("unrecognized appeal status")
	ErrAppealDuplicate                = errors.New("appeal with the same resource and role already exists")
	ErrAppealInvalidExtensionDuration = errors.New("invalid appeal extension duration")
	ErrAppealFoundActiveAccess        = errors.New("user still have an active access")
	ErrAppealNotEligibleForExtension  = errors.New("appeal is not eligible for extension")

	ErrApprovalDependencyIsBlocked = errors.New("found previous approval step that is still in blocked")
	ErrApprovalDependencyIsPending = errors.New("found previous approval step that is still in pending")
	ErrApprovalStatusApproved      = errors.New("approval already approved")
	ErrApprovalStatusRejected      = errors.New("approval already rejected")
	ErrApprovalStatusSkipped       = errors.New("approval already skipped")
	ErrApprovalStatusUnrecognized  = errors.New("unrecognized approval status")
	ErrApprovalNameNotFound        = errors.New("approval step name not found")

	ErrActionForbidden    = errors.New("user is not allowed to make action on this approval step")
	ErrActionInvalidValue = errors.New("invalid action value")

	ErrProviderTypeNotFound    = errors.New("provider is not registered")
	ErrProviderURNNotFound     = errors.New("provider with specified urn is not registered")
	ErrResourceTypeNotFound    = errors.New("unable to find matching resource config for specified resource type")
	ErrOptionsDurationNotFound = errors.New("duration option not found")
	ErrInvalidRole             = errors.New("invalid role")
	ErrDurationIsRequired      = errors.New("having permanent access to this resource is not allowed, access duration is required")
	ErrPolicyIDNotFound        = errors.New("unable to find approval policy for specified id")
	ErrPolicyVersionNotFound   = errors.New("unable to find approval policy for specified version")
	ErrResourceNotFound        = errors.New("resource not found")
	ErrAppealNotFound          = errors.New("appeal not found")

	ErrApproverKeyNotRecognized = errors.New("unrecognized approvers key")
	ErrApproverInvalidType      = errors.New("invalid approver type, expected an email or array of email")
)
