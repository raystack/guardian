//go:generate mockery --name=repository --exported
//go:generate mockery --name=Client --exported
//go:generate mockery --name=resourceService --exported
//go:generate mockery --name=auditLogger --exported

package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/imdario/mergo"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/plugins/providers"
	"github.com/odpf/guardian/store"
	"github.com/odpf/guardian/utils"
	"github.com/odpf/salt/audit"
	"github.com/odpf/salt/log"
)

const (
	AuditKeyCreate = "provider.create"
	AuditKeyUpdate = "provider.update"
)

type repository interface {
	store.ProviderRepository
}

type Client interface {
	providers.Client
}

type resourceService interface {
	Find(context.Context, map[string]interface{}) ([]*domain.Resource, error)
	BulkUpsert(context.Context, []*domain.Resource) error
}

type auditLogger interface {
	Log(ctx context.Context, action string, data interface{}) error
}

// Service handling the business logics
type Service struct {
	repository      repository
	resourceService resourceService
	clients         map[string]Client

	validator   *validator.Validate
	logger      log.Logger
	auditLogger auditLogger
}

type ServiceDeps struct {
	Repository      repository
	ResourceService resourceService
	Clients         []Client

	Validator   *validator.Validate
	Logger      log.Logger
	AuditLogger auditLogger
}

// NewService returns service struct
func NewService(deps ServiceDeps) *Service {
	mapProviderClients := make(map[string]Client)
	for _, c := range deps.Clients {
		mapProviderClients[c.GetType()] = c
	}

	return &Service{
		deps.Repository,
		deps.ResourceService,
		mapProviderClients,

		deps.Validator,
		deps.Logger,
		deps.AuditLogger,
	}
}

// Create record
func (s *Service) Create(ctx context.Context, p *domain.Provider) error {
	c := s.getClient(p.Type)
	if c == nil {
		return ErrInvalidProviderType
	}

	accountTypes := c.GetAccountTypes()
	if err := s.validateAccountTypes(p.Config, accountTypes); err != nil {
		return err
	}

	if p.Config.Appeal != nil {
		if err := s.validateAppealConfig(p.Config.Appeal); err != nil {
			return err
		}
	}

	if err := c.CreateConfig(p.Config); err != nil {
		return err
	}

	if err := s.repository.Create(p); err != nil {
		return err
	}

	if err := s.auditLogger.Log(ctx, AuditKeyCreate, p); err != nil {
		s.logger.Error("failed to record audit log", "error", err)
	}

	go func() {
		s.logger.Info("fetching resources", "provider_urn", p.URN)
		ctx := audit.WithActor(context.Background(), domain.SystemActorName)
		resources, err := s.getResources(ctx, p)
		if err != nil {
			s.logger.Error("failed to fetch resources", "error", err)
		}
		if err := s.resourceService.BulkUpsert(ctx, resources); err != nil {
			s.logger.Error("failed to insert resources to db", "error", err)
		} else {
			s.logger.Info("resources added",
				"provider_urn", p.URN,
				"count", len(resources),
			)
		}
	}()

	return nil
}

// Find records
func (s *Service) Find(ctx context.Context) ([]*domain.Provider, error) {
	providers, err := s.repository.Find()
	if err != nil {
		return nil, err
	}

	return providers, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*domain.Provider, error) {
	return s.repository.GetByID(id)
}

func (s *Service) GetTypes(ctx context.Context) ([]domain.ProviderType, error) {
	return s.repository.GetTypes()
}

func (s *Service) GetOne(ctx context.Context, pType, urn string) (*domain.Provider, error) {
	return s.repository.GetOne(pType, urn)
}

// Update updates the non-zero value(s) only
func (s *Service) Update(ctx context.Context, p *domain.Provider) error {
	var currentProvider *domain.Provider
	var err error

	if len(p.ID) > 0 {
		currentProvider, err = s.GetByID(ctx, p.ID)
	} else {
		currentProvider, err = s.GetOne(ctx, p.Type, p.URN)
	}
	if err != nil {
		return err
	}

	if err := mergo.Merge(p, currentProvider); err != nil {
		return err
	}

	c := s.getClient(p.Type)
	if c == nil {
		return ErrInvalidProviderType
	}

	accountTypes := c.GetAccountTypes()
	if err := s.validateAccountTypes(p.Config, accountTypes); err != nil {
		return err
	}

	if p.Config.Appeal != nil {
		if err := s.validateAppealConfig(p.Config.Appeal); err != nil {
			return err
		}
	}

	if err := c.CreateConfig(p.Config); err != nil {
		return err
	}

	if err := s.repository.Update(p); err != nil {
		return err
	}

	if err := s.auditLogger.Log(ctx, AuditKeyUpdate, p); err != nil {
		s.logger.Error("failed to record audit log", "error", err)
	}

	return nil
}

// FetchResources fetches all resources for all registered providers
func (s *Service) FetchResources(ctx context.Context) error {
	providers, err := s.repository.Find()
	if err != nil {
		return err
	}

	resources := []*domain.Resource{}
	for _, p := range providers {
		s.logger.Info("fetching resources", "provider_urn", p.URN)
		res, err := s.getResources(ctx, p)
		if err != nil {
			s.logger.Error("failed to send notifications", "error", err)
			continue
		}
		s.logger.Info("resources added",
			"provider_urn", p.URN,
			"count", len(resources),
		)
		resources = append(resources, res...)
	}

	return s.resourceService.BulkUpsert(ctx, resources)
}

func (s *Service) GetRoles(ctx context.Context, id string, resourceType string) ([]*domain.Role, error) {
	p, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	c := s.getClient(p.Type)
	return c.GetRoles(p.Config, resourceType)
}

func (s *Service) ValidateAppeal(ctx context.Context, a *domain.Appeal, p *domain.Provider) error {
	if err := s.validateAppealParam(a); err != nil {
		return err
	}

	resourceType := a.Resource.Type
	c := s.getClient(p.Type)
	if c == nil {
		return ErrInvalidProviderType
	}

	if !utils.ContainsString(p.Config.AllowedAccountTypes, a.AccountType) {
		allowedAccountTypesStr := strings.Join(p.Config.AllowedAccountTypes, ", ")
		return fmt.Errorf("invalid account type: %v. allowed account types for %v: %v", a.AccountType, p.Type, allowedAccountTypesStr)
	}

	roles, err := c.GetRoles(p.Config, resourceType)
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

func (s *Service) GrantAccess(ctx context.Context, a *domain.Appeal) error {
	if err := s.validateAppealParam(a); err != nil {
		return err
	}

	c := s.getClient(a.Resource.ProviderType)
	if c == nil {
		return ErrInvalidProviderType
	}

	p, err := s.getProviderConfig(ctx, a.Resource.ProviderType, a.Resource.ProviderURN)
	if err != nil {
		return err
	}

	return c.GrantAccess(p.Config, a)
}

func (s *Service) RevokeAccess(ctx context.Context, a *domain.Appeal) error {
	if err := s.validateAppealParam(a); err != nil {
		return err
	}

	c := s.getClient(a.Resource.ProviderType)
	if c == nil {
		return ErrInvalidProviderType
	}

	p, err := s.getProviderConfig(ctx, a.Resource.ProviderType, a.Resource.ProviderURN)
	if err != nil {
		return err
	}

	return c.RevokeAccess(p.Config, a)
	// TODO: handle if permission for the given user with the given role is not found
	// handle the resolution for the appeal status
}

func (s *Service) getResources(ctx context.Context, p *domain.Provider) ([]*domain.Resource, error) {
	resources := []*domain.Resource{}
	c := s.getClient(p.Type)
	if c == nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidProviderType, p.Type)
	}

	existingResources, err := s.resourceService.Find(ctx, map[string]interface{}{
		"provider_type": p.Type,
		"provider_urn":  p.URN,
	})
	if err != nil {
		return nil, err
	}

	res, err := c.GetResources(p.Config)
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

func (s *Service) getClient(pType string) providers.Client {
	return s.clients[pType]
}

func (s *Service) getProviderConfig(ctx context.Context, pType, urn string) (*domain.Provider, error) {
	p, err := s.GetOne(ctx, pType, urn)
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
