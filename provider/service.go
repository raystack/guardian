package provider

import (
	"fmt"
	"time"

	"github.com/imdario/mergo"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/salt/log"
)

// Service handling the business logics
type Service struct {
	logger             *log.Logrus
	providerRepository domain.ProviderRepository
	resourceService    domain.ResourceService

	providers map[string]domain.ProviderInterface
}

// NewService returns service struct
func NewService(logger *log.Logrus, pr domain.ProviderRepository, rs domain.ResourceService, providers []domain.ProviderInterface) *Service {
	mapProviders := make(map[string]domain.ProviderInterface)
	for _, p := range providers {
		mapProviders[p.GetType()] = p
	}

	return &Service{
		logger:             logger,
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

	if p.Config.Appeal != nil {
		if err := s.validateAppealConfig(p.Config.Appeal); err != nil {
			return err
		}
	}

	if err := provider.CreateConfig(p.Config); err != nil {
		return err
	}

	return s.providerRepository.Create(p)
}

// Find records
func (s *Service) Find() ([]*domain.Provider, error) {
	providers, err := s.providerRepository.Find()
	if err != nil {
		return nil, err
	}

	for _, p := range providers {
		p.Config.Credentials = nil
	}

	return providers, nil
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

	if p.Config.Appeal != nil {
		if err := s.validateAppealConfig(p.Config.Appeal); err != nil {
			return err
		}
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
			s.logger.Error(fmt.Sprintf("%v: %v", ErrInvalidProviderType, p.Type))
			continue
		}

		res, err := provider.GetResources(p.Config)
		if err != nil {
			s.logger.Error(fmt.Sprintf("error fetching resources for %v: %v", p.ID, err))
			continue
		}

		resources = append(resources, res...)
	}

	return s.resourceService.BulkUpsert(resources)
}

func (s *Service) GetRoles(id uint, resourceType string) ([]*domain.Role, error) {
	p, err := s.providerRepository.GetByID(id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrRecordNotFound
	}

	provider := s.getProvider(p.Type)
	return provider.GetRoles(p.Config, resourceType)
}

func (s *Service) ValidateAppeal(a *domain.Appeal, p *domain.Provider) error {
	if err := s.validateAppealParam(a); err != nil {
		return err
	}

	resourceType := a.Resource.Type
	provider := s.getProvider(p.Type)
	roles, err := provider.GetRoles(p.Config, resourceType)
	if err != nil {
		return err
	}

	isRoleExists := false
	for _, role := range roles {
		if a.Role == role.Name {
			isRoleExists = true
			break
		}
	}

	if !isRoleExists {
		return ErrInvalidRole
	}

	if !p.Config.Appeal.AllowPermanentAccess {
		if a.Options == nil || a.Options.ExpirationDate == nil {
			return ErrOptionsExpirationDateOptionNotFound
		} else if a.Options.ExpirationDate.IsZero() {
			return ErrExpirationDateIsRequired
		}
	}

	return nil
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
	// TODO: handle if permission for the given user with the given role is not found
	// handle the resolution for the appeal status
}

func (s *Service) validateAppealParam(a *domain.Appeal) error {
	if a == nil {
		return ErrNilAppeal
	}
	if a.Resource == nil {
		return ErrNilResource
	}
	//TO-DO
	//Make sure the user and role is required
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

func (s *Service) validateAppealConfig(a *domain.AppealConfig) error {
	if a.AllowActiveAccessExtensionIn != "" {
		if _, err := time.ParseDuration(a.AllowActiveAccessExtensionIn); err != nil {
			return fmt.Errorf("parsing appeal extension duration: %v", err)
		}
	}

	return nil
}
