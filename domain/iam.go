package domain

// IAMClient interface
type IAMClient interface {
	GetUser(id string) (interface{}, error)
}
