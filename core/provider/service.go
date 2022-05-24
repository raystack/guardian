package provider

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/imdario/mergo"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/plugins/providers"
	"github.com/odpf/guardian/store"
	"github.com/odpf/guardian/utils"
	"github.com/odpf/salt/log"
)

type resourceService interface {
	Find(map[string]interface{}) ([]*domain.Resource, error)
	BulkUpsert([]*domain.Resource) error
	BatchDelete([]string) error
}

// Service handling the business logics
type Service struct {
	logger             log.Logger
	validator          *validator.Validate
	providerRepository store.ProviderRepository
	resourceService    resourceService

	providerClients map[string]providers.Client
}

// NewService returns service struct
func NewService(
	logger log.Logger,
	validator *validator.Validate,
	pr store.ProviderRepository,
	rs resourceService,
	providerClients []providers.Client,
) *Service {
	mapProviderClients := make(map[string]providers.Client)
	for _, c := range providerClients {
		mapProviderClients[c.GetType()] = c
	}

	return &Service{
		logger:             logger,
		validator:          validator,
		providerRepository: pr,
		resourceService:    rs,
		providerClients:    mapProviderClients,
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

	if err := s.providerRepository.Create(p); err != nil {
		return err
	}

	go func() {
		s.logger.Info(fmt.Sprintf("fetching resources for %s", p.URN))
		resources, err := s.getResources(p)
		if err != nil {
			s.logger.Error(fmt.Sprintf("fetching resources: %s", err))
		}
		if err := s.resourceService.BulkUpsert(resources); err != nil {
			s.logger.Error(fmt.Sprintf("inserting resources to db: %s", err))
		} else {
			s.logger.Info(fmt.Sprintf("added %v resources for %s", len(resources), p.URN))
		}
	}()

	return nil
}

// Find records
func (s *Service) Find() ([]*domain.Provider, error) {
	providers, err := s.providerRepository.Find()
	if err != nil {
		return nil, err
	}

	return providers, nil
}

func (s *Service) GetByID(id string) (*domain.Provider, error) {
	return s.providerRepository.GetByID(id)
}

func (s *Service) GetTypes() ([]domain.ProviderType, error) {
	return s.providerRepository.GetTypes()
}

func (s *Service) GetOne(pType, urn string) (*domain.Provider, error) {
	return s.providerRepository.GetOne(pType, urn)
}

// Update updates the non-zero value(s) only
func (s *Service) Update(p *domain.Provider) error {
	var currentProvider *domain.Provider
	var err error

	if len(p.ID) > 0 {
		currentProvider, err = s.GetByID(p.ID)
	} else {
		currentProvider, err = s.GetOne(p.Type, p.URN)
	}
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
		s.logger.Info(fmt.Sprintf("fetching resources for %s", p.URN))
		res, err := s.getResources(p)
		if err != nil {
			s.logger.Error(err.Error())
			continue
		}
		s.logger.Info(fmt.Sprintf("got %v resources for %s", len(res), p.URN))
		resources = append(resources, res...)
	}

	return s.resourceService.BulkUpsert(resources)
}

func (s *Service) GetRoles(id string, resourceType string) ([]*domain.Role, error) {
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

	isRoleExists := len(roles) == 0
	for _, role := range roles {
		if a.Role == role.ID {
			isRoleExists = true
			break
		}
	}

	if !isRoleExists {
		return ErrInvalidRole
	}

	if p.Config.Appeal != nil && !p.Config.Appeal.AllowPermanentAccess {
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

func (s *Service) Delete(id string) error {
	p, err := s.providerRepository.GetByID(id)
	if err != nil {
		return fmt.Errorf("getting provider details: %w", err)
	}

	resources, err := s.resourceService.Find(map[string]interface{}{
		"provider_type": p.Type,
		"provider_urn":  p.URN,
	})
	if err != nil {
		return fmt.Errorf("retrieving related resources: %w", err)
	}
	var resourceIds []string
	for _, r := range resources {
		resourceIds = append(resourceIds, r.ID)
	}
	// TODO: execute in transaction
	if err := s.resourceService.BatchDelete(resourceIds); err != nil {
		return fmt.Errorf("batch deleting resources: %w", err)
	}

	return s.providerRepository.Delete(id)
}

func (s *Service) getResources(p *domain.Provider) ([]*domain.Resource, error) {
	resources := []*domain.Resource{}
	provider := s.getProvider(p.Type)
	if provider == nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidProviderType, p.Type)
	}

	existingResources, err := s.resourceService.Find(map[string]interface{}{
		"provider_type": p.Type,
		"provider_urn":  p.URN,
	})
	if err != nil {
		return nil, err
	}

	res, err := provider.GetResources(p.Config)
	if err != nil {
		return nil, fmt.Errorf("error fetching resources for %v: %w", p.ID, err)
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

	return resources, nil
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

func (s *Service) getProvider(pType string) providers.Client {
	return s.providerClients[pType]
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
