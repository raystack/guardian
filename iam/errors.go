package iam

import "errors"

var (
	// ErrEmptyUserEmailParam is the error value when the passed user email is empty
	ErrEmptyUserEmailParam = errors.New("user email param is required")
	// ErrEmptyApprovers is the error value when the returned approver emails are zero/empty
	ErrEmptyApprovers = errors.New("got zero approver")
)
