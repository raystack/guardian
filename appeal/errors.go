package appeal

import "errors"

var (
	ErrProviderTypeNotFound  = errors.New("provider is not registered")
	ErrProviderURNNotFound   = errors.New("provider with specified urn is not registered")
	ErrPolicyConfigNotFound  = errors.New("unable to find matching approval policy config for specified resource")
	ErrPolicyIDNotFound      = errors.New("unable to find approval policy for specified id")
	ErrPolicyVersionNotFound = errors.New("unable to find approval policy for specified version")

	ErrApproverKeyNotRecognized = errors.New("unrecognized approvers key")
	ErrApproverInvalidType      = errors.New("invalid approver type, expected an email or array of email")
)
