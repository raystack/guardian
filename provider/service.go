package provider

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/imdario/mergo"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/utils"
	"github.com/odpf/salt/log"
)

// Service handling the business logics
type Service struct {
	logger             *log.Logrus
	validator          *validator.Validate
	providerRepository domain.ProviderRepository
	resourceService    domain.ResourceService

	providers map[string]domain.ProviderInterface
}

// NewService returns service struct
func NewService(logger *log.Logrus, validator *validator.Validate, pr domain.ProviderRepository, rs domain.ResourceService, providers []domain.ProviderInterface) *Service {
	mapProviders := make(map[string]domain.ProviderInterface)
	for _, p := range providers {
		mapProviders[p.GetType()] = p
	}

	return &Service{
		logger:             logger,
		validator:          validator,
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

	accountTypes := provider.GetAccountTypes()
	if err := s.validateAccountTypes(p.Config, accountTypes); err != nil {
		return err
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

	return providers, nil
}

func (s *Service) GetByID(id uint) (*domain.Provider, error) {
	return s.providerRepository.GetByID(id)
}

func (s *Service) GetOne(pType, urn string) (*domain.Provider, error) {
	return s.providerRepository.GetOne(pType, urn)
}

// Update updates the non-zero value(s) only
func (s *Service) Update(p *domain.Provider) error {
	currentProvider, err := s.GetByID(p.ID)
	if err != nil {
		return err
	}

	if err := mergo.Merge(p, currentProvider); err != nil {
		return err
	}

	provider := s.getProvider(p.Type)
	if provider == nil {
		return ErrInvalidProviderType
	}

	accountTypes := provider.GetAccountTypes()
	if err := s.validateAccountTypes(p.Config, accountTypes); err != nil {
		return err
	}

	if p.Config.Appeal != nil {
		if err := s.validateAppealConfig(p.Config.Appeal); err != nil {
			return err
		}
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

		existingResources, err := s.resourceService.Find(map[string]interface{}{
			"provider_type": p.Type,
			"provider_urn":  p.URN,
		})
		if err != nil {
			return err
		}

		res, err := provider.GetResources(p.Config)
		if err != nil {
			s.logger.Error(fmt.Sprintf("error fetching resources for %v: %v", p.ID, err))
			continue
		}

		for _, er := range existingResources {
			isFound := false
			for _, r := range res {
				if er.URN == r.URN {
					resources = append(resources, r)
					isFound = true
					break
				}
			}
			if !isFound {
				er.IsDeleted = true
				resources = append(resources, er)
			}
		}
		for _, r := range res {
			isAdded := false
			for _, rr := range resources {
				if r.URN == rr.URN {
					isAdded = true
					break
				}
			}
			if !isAdded {
				resources = append(resources, r)
			}
		}
	}
	return s.resourceService.BulkUpsert(resources)
}

func (s *Service) GetRoles(id uint, resourceType string) ([]*domain.Role, error) {
	p, err := s.GetByID(id)
	if err != nil {
		return nil, err
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
	if provider == nil {
		return ErrInvalidProviderType
	}

	if !utils.ContainsString(p.Config.AllowedAccountTypes, a.AccountType) {
		allowedAccountTypesStr := strings.Join(p.Config.AllowedAccountTypes, ", ")
		return fmt.Errorf("invalid account type: %v. allowed account types for %v: %v", a.AccountType, p.Type, allowedAccountTypesStr)
	}

	roles, err := provider.GetRoles(p.Config, resourceType)
	if err != nil {
		return err
	}

	isRoleExists := false
	for _, role := range roles {
		if a.Role == role.ID {
			isRoleExists = true
			break
		}
	}

	if !isRoleExists {
		return ErrInvalidRole
	}

	if !p.Config.Appeal.AllowPermanentAccess {
		if a.Options == nil {
			return ErrOptionsDurationNotFound
		}

		if a.Options.Duration == "" {
			return ErrDurationIsRequired
		}

		if err := validateDuration(a.Options.Duration); err != nil {
			return fmt.Errorf("invalid duration: %v", err)
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
	p, err := s.GetOne(pType, urn)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) validateAccountTypes(pc *domain.ProviderConfig, accountTypes []string) error {
	if pc.AllowedAccountTypes == nil {
		pc.AllowedAccountTypes = accountTypes
	} else {
		if err := s.validator.Var(pc.AllowedAccountTypes, "min=1,unique"); err != nil {
			return err
		}

		for _, at := range pc.AllowedAccountTypes {
			accountTypesStr := strings.Join(accountTypes, " ")
			if err := s.validator.Var(at, fmt.Sprintf("oneof=%v", accountTypesStr)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Service) validateAppealConfig(a *domain.AppealConfig) error {
	if a.AllowActiveAccessExtensionIn != "" {
		if err := validateDuration(a.AllowActiveAccessExtensionIn); err != nil {
			return fmt.Errorf("invalid appeal extension policy: %v", err)
		}
	}

	return nil
}

func validateDuration(d string) error {
	if _, err := time.ParseDuration(d); err != nil {
		return fmt.Errorf("parsing duration: %v", err)
	}
	return nil
}
