package domain

// IAMClient interface
type IAMClient interface {
	GetUser(id string) (interface{}, error)
}

// IAMService interface
type IAMService interface {
	GetUser(id string) (interface{}, error)
}
