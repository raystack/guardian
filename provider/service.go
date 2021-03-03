package provider

import (
	"github.com/odpf/guardian/domain"
)

// Service handling the business logics
type Service struct {
	providerRepository domain.ProviderRepository

	providers map[string]domain.ProviderInterface
}

// NewService returns service struct
func NewService(pr domain.ProviderRepository, providers []domain.ProviderInterface) *Service {
	mapProviders := make(map[string]domain.ProviderInterface)
	for _, p := range providers {
		mapProviders[p.GetType()] = p
	}

	return &Service{
		providerRepository: pr,
		providers:          mapProviders,
	}
}

// Create record
func (s *Service) Create(p *domain.Provider) error {
	provider := s.getProvider(p.Type)
	if provider == nil {
		return ErrInvalidProviderType
	}

	if err := provider.ValidateConfig(p.Config); err != nil {
		return err
	}

	return s.providerRepository.Create(p)
}

func (s *Service) getProvider(pType string) domain.ProviderInterface {
	return s.providers[pType]
}
