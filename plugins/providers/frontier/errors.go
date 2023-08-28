package frontier

import "errors"

var (
	ErrInvalidPermissionConfig = errors.New("invalid permission config type")
	ErrInvalidResourceType     = errors.New("invalid resource type")
)
