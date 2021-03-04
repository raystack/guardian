package provider

import (
	"github.com/imdario/mergo"
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

// Find records
func (s *Service) Find() ([]*domain.Provider, error) {
	return s.providerRepository.Find()
}

// Update updates the non-zero value(s) only
func (s *Service) Update(p *domain.Provider) error {
	if p.ID == 0 {
		return ErrEmptyIDParam
	}

	currentProvider, err := s.providerRepository.GetOne(p.ID)
	if err != nil {
		return err
	}
	if currentProvider == nil {
		return ErrRecordNotFound
	}

	if err := mergo.Merge(p, currentProvider); err != nil {
		return err
	}

	return s.providerRepository.Update(p)
}
