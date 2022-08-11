package access

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/plugins/notifiers"
	"github.com/odpf/salt/log"
)

const (
	AuditKeyRevoke = "access.revoke"
)

//go:generate mockery --name=repository --exported --with-expecter
type repository interface {
	List(context.Context, domain.ListAccessesFilter) ([]domain.Access, error)
	GetByID(context.Context, string) (*domain.Access, error)
	Update(context.Context, *domain.Access) error
}

//go:generate mockery --name=providerService --exported --with-expecter
type providerService interface {
	RevokeAccess(context.Context, *domain.Appeal) error
}

//go:generate mockery --name=auditLogger --exported --with-expecter
type auditLogger interface {
	Log(ctx context.Context, action string, data interface{}) error
}

//go:generate mockery --name=notifier --exported --with-expecter
type notifier interface {
	notifiers.Client
}

type accessCreation struct {
	AppealStatus string `validate:"required,eq=active"`
	AccountID    string `validate:"required"`
	AccountType  string `validate:"required"`
	ResourceID   string `validate:"required"`
}

type Service struct {
	repo            repository
	providerService providerService

	notifier    notifier
	validator   *validator.Validate
	logger      log.Logger
	auditLogger auditLogger
}

type ServiceDeps struct {
	Repository      repository
	ProviderService providerService

	Notifier    notifier
	Validator   *validator.Validate
	Logger      log.Logger
	AuditLogger auditLogger
}

func NewService(deps ServiceDeps) *Service {
	return &Service{
		repo:            deps.Repository,
		providerService: deps.ProviderService,

		notifier:    deps.Notifier,
		validator:   deps.Validator,
		logger:      deps.Logger,
		auditLogger: deps.AuditLogger,
	}
}

func (s *Service) List(ctx context.Context, filter domain.ListAccessesFilter) ([]domain.Access, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) GetByID(ctx context.Context, id string) (*domain.Access, error) {
	if id == "" {
		return nil, ErrEmptyIDParam
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Prepare(ctx context.Context, appeal domain.Appeal) (*domain.Access, error) {
	if err := s.validator.Struct(accessCreation{
		AppealStatus: appeal.Status,
		AccountID:    appeal.AccountID,
		AccountType:  appeal.AccountType,
		ResourceID:   appeal.ResourceID,
	}); err != nil {
		return nil, fmt.Errorf("validating appeal: %w", err)
	}

	return appeal.ToAccess()
}

func (s *Service) Revoke(ctx context.Context, id, actor, reason string, opts ...Option) (*domain.Access, error) {
	access, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting access details: %w", err)
	}

	if err := access.Revoke(actor, reason); err != nil {
		return nil, err
	}

	options := s.getOptions(opts...)

	if !options.skipRevokeInProvider {
		if err := s.providerService.RevokeAccess(ctx, access.Appeal); err != nil {
			return nil, fmt.Errorf("removing access in provider: %w", err)
		}
	}
	if err := s.repo.Update(ctx, access); err != nil {
		return nil, fmt.Errorf("updating access record in db: %w", err)
	}

	if !options.skipNotification {
		if errs := s.notifier.Notify([]domain.Notification{{
			User: access.Appeal.CreatedBy,
			Message: domain.NotificationMessage{
				Type: domain.NotificationTypeAccessRevoked,
				Variables: map[string]interface{}{
					"resource_name": fmt.Sprintf("%s (%s: %s)", access.Resource.Name, access.Resource.ProviderType, access.Resource.URN),
					"role":          access.Appeal.Role,
					"account_type":  access.AccountType,
					"account_id":    access.AccountID,
				},
			},
		}}); errs != nil {
			for _, err1 := range errs {
				s.logger.Error("failed to send notifications", "error", err1.Error())
			}
		}
	}

	if err := s.auditLogger.Log(ctx, AuditKeyRevoke, map[string]interface{}{
		"access_id": id,
		"reason":    reason,
	}); err != nil {
		s.logger.Error("failed to record audit log", "error", err)
	}

	return access, nil
}
