package provider

import "github.com/odpf/guardian/domain"

// Usecase handling the business logics
type Usecase struct {
	providerRepository domain.ProviderRepository
}

// NewUsecase returns usecase struct
func NewUsecase(pr domain.ProviderRepository) *Usecase {
	return &Usecase{pr}
}

// Create record
func (u *Usecase) Create(p *domain.Provider) error {
	return u.providerRepository.Create(p)
}
