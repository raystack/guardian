package migrations

const (
	Metabase        = "metabase"
	Username        = "username"
	Host            = "host"
	Password        = "password"
	Duration        = "duration"
	Group           = "group"
	Member          = "member"
	DefaultDuration = "720h"
)

type ResourceRequest struct {
	ID       string
	Name     string
	Role     string
	Duration string
}

type AppealRequest struct {
	AccountID string
	User      string
	Resource  ResourceRequest
}

type Client interface {
	GetType() string
	PopulateAccess() ([]AppealRequest, error)
}
