package usecases

import "github.com/odpf/guardian/repositories"

// UseCases contains all usecases
type UseCases struct {
	Provider Provider
}

// New returns usecases
func New(r *repositories.Repositories) *UseCases {
	return &UseCases{
		Provider: NewProvider(r.Provider),
	}
}
