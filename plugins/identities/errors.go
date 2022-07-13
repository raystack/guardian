package identities

import "errors"

var (
	ErrUserActiveEmptyMetadata          = errors.New("user active metadata is required")
	ErrUserAccountStatusKeyShouldBeBool = errors.New("user account status key should be boolean")
)
