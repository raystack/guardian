package migrations

type ResourceRequest struct {
	ID       string
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
