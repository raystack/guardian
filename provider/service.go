package provider

import (
	"github.com/imdario/mergo"
	"github.com/odpf/guardian/domain"
)

// Service handling the business logics
type Service struct {
	providerRepository domain.ProviderRepository
	resourceService    domain.ResourceService

	providers map[string]domain.ProviderInterface
}

// NewService returns service struct
func NewService(pr domain.ProviderRepository, rs domain.ResourceService, providers []domain.ProviderInterface) *Service {
	mapProviders := make(map[string]domain.ProviderInterface)
	for _, p := range providers {
		mapProviders[p.GetType()] = p
	}

	return &Service{
		providerRepository: pr,
		resourceService:    rs,
		providers:          mapProviders,
	}
}

// Create record
func (s *Service) Create(p *domain.Provider) error {
	provider := s.getProvider(p.Type)
	if provider == nil {
		return ErrInvalidProviderType
	}

	if err := provider.CreateConfig(p.Config); err != nil {
		return err
	}

	return s.providerRepository.Create(p)
}

// Find records
func (s *Service) Find() ([]*domain.Provider, error) {
	return s.providerRepository.Find()
}

// Update updates the non-zero value(s) only
func (s *Service) Update(p *domain.Provider) error {
	currentProvider, err := s.providerRepository.GetByID(p.ID)
	if err != nil {
		return err
	}
	if currentProvider == nil {
		return ErrRecordNotFound
	}

	if err := mergo.Merge(p, currentProvider); err != nil {
		return err
	}

	provider := s.getProvider(p.Type)
	if provider == nil {
		return ErrInvalidProviderType
	}
	if err := provider.CreateConfig(p.Config); err != nil {
		return err
	}

	return s.providerRepository.Update(p)
}

// FetchResources fetches all resources for all registered providers
func (s *Service) FetchResources() error {
	providers, err := s.providerRepository.Find()
	if err != nil {
		return err
	}

	resources := []*domain.Resource{}
	for _, p := range providers {
		provider := s.getProvider(p.Type)
		if provider == nil {
			return ErrInvalidProviderType
		}

		res, err := provider.GetResources(p.Config)
		if err != nil {
			return err
		}

		resources = append(resources, res...)
	}

	return s.resourceService.BulkUpsert(resources)
}

func (s *Service) GrantAccess(a *domain.Appeal) error {
	if err := s.validateAppealParam(a); err != nil {
		return err
	}

	provider := s.getProvider(a.Resource.ProviderType)
	if provider == nil {
		return ErrInvalidProviderType
	}

	p, err := s.getProviderConfig(a.Resource.ProviderType, a.Resource.ProviderURN)
	if err != nil {
		return err
	}

	return provider.GrantAccess(p.Config, a)
}

func (s *Service) RevokeAccess(a *domain.Appeal) error {
	if err := s.validateAppealParam(a); err != nil {
		return err
	}

	provider := s.getProvider(a.Resource.ProviderType)
	if provider == nil {
		return ErrInvalidProviderType
	}

	p, err := s.getProviderConfig(a.Resource.ProviderType, a.Resource.ProviderURN)
	if err != nil {
		return err
	}

	return provider.RevokeAccess(p.Config, a)
}

func (s *Service) validateAppealParam(a *domain.Appeal) error {
	if a == nil {
		return ErrNilAppeal
	}
	if a.Resource == nil {
		return ErrNilResource
	}
	return nil
}

func (s *Service) getProvider(pType string) domain.ProviderInterface {
	return s.providers[pType]
}

func (s *Service) getProviderConfig(pType, urn string) (*domain.Provider, error) {
	p, err := s.providerRepository.GetOne(pType, urn)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrProviderNotFound
	}
	return p, nil
}
