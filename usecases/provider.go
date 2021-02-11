package usecases

import (
	"github.com/odpf/guardian/models"
	"github.com/odpf/guardian/repositories"
)

// Provider use case interface
type Provider interface {
	Create(*models.Provider) error
}

// ProviderUseCase usecase
type ProviderUseCase struct {
	providerRepository repositories.Provider
}

// NewProvider returns provider usecase struct
func NewProvider(pr repositories.Provider) *ProviderUseCase {
	return &ProviderUseCase{pr}
}

// Create provider record
func (u *ProviderUseCase) Create(m *models.Provider) error {
	return u.providerRepository.Create(m)
}
