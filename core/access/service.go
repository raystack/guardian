package access

import (
	"context"
	"fmt"
	"time"

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
	RevokeAccess(context.Context, domain.Access) error
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

	revokedAccess := &domain.Access{}
	*revokedAccess = *access
	if err := access.Revoke(actor, reason); err != nil {
		return nil, err
	}
	// TODO: remove below logic in future release when appeal no longer for managing access
	if err := access.Appeal.Revoke(actor, reason); err != nil {
		return nil, fmt.Errorf("updating appeal status: %s", err)
	}
	if err := s.repo.Update(ctx, access); err != nil {
		return nil, fmt.Errorf("updating access record in db: %w", err)
	}

	options := s.getOptions(opts...)

	if !options.skipRevokeInProvider {
		if err := s.providerService.RevokeAccess(ctx, *access); err != nil {
			if err := s.repo.Update(ctx, access); err != nil {
				return nil, fmt.Errorf("failed to rollback access status: %w", err)
			}
			return nil, fmt.Errorf("removing access in provider: %w", err)
		}
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

func (s *Service) BulkRevoke(ctx context.Context, filter domain.RevokeAccessesFilter, actor, reason string) ([]*domain.Access, error) {
	if filter.AccountIDs == nil || len(filter.AccountIDs) == 0 {
		return nil, fmt.Errorf("account_ids is required")
	}

	accesses, err := s.List(ctx, domain.ListAccessesFilter{
		Statuses:      []string{domain.AppealStatusActive},
		AccountIDs:    filter.AccountIDs,
		ProviderTypes: filter.ProviderTypes,
		ProviderURNs:  filter.ProviderURNs,
		ResourceTypes: filter.ResourceTypes,
		ResourceURNs:  filter.ResourceURNs,
	})
	if err != nil {
		return nil, fmt.Errorf("listing active accesses: %w", err)
	}
	if len(accesses) == 0 {
		return nil, nil
	}

	result := make([]*domain.Access, 0)
	batchSize := 10
	timeLimiter := make(chan int, batchSize)

	for i := 1; i <= batchSize; i++ {
		timeLimiter <- i
	}

	go func() {
		for range time.Tick(1 * time.Second) {
			for i := 1; i <= batchSize; i++ {
				timeLimiter <- i
			}
		}
	}()

	totalRequests := len(accesses)
	done := make(chan *domain.Access, totalRequests)
	resourceAccessMap := make(map[string][]*domain.Access, 0)

	for i, access := range accesses {
		var resourceAccesses []*domain.Access
		var ok bool
		if resourceAccesses, ok = resourceAccessMap[access.ResourceID]; ok {
			resourceAccesses = append(resourceAccesses, &accesses[i])
		} else {
			resourceAccesses = []*domain.Access{&accesses[i]}
		}
		resourceAccessMap[access.ResourceID] = resourceAccesses
	}

	for _, resourceAccesses := range resourceAccessMap {
		go s.expiredInActiveUserAccess(ctx, timeLimiter, done, actor, reason, resourceAccesses)
	}

	var successRevoke []string
	var failedRevoke []string
	for {
		select {
		case access := <-done:
			if access.Status == domain.AccessStatusInactive {
				successRevoke = append(successRevoke, access.ID)
			} else {
				failedRevoke = append(failedRevoke, access.ID)
			}
			result = append(result, access)
			if len(result) == totalRequests {
				s.logger.Info("successful access revocation", "count", len(successRevoke), "ids", successRevoke)
				s.logger.Info("failed access revocation", "count", len(failedRevoke), "ids", failedRevoke)
				return result, nil
			}
		}
	}
}

func (s *Service) expiredInActiveUserAccess(ctx context.Context, timeLimiter chan int, done chan *domain.Access, actor string, reason string, accesses []*domain.Access) {
	for _, access := range accesses {
		<-timeLimiter

		revokedAccess := &domain.Access{}
		*revokedAccess = *access
		if err := revokedAccess.Revoke(actor, reason); err != nil {
			s.logger.Error("failed to revoke access", "id", access.ID, "error", err)
			return
		}
		// TODO: remove below logic in future release when appeal no longer for managing access
		if err := access.Appeal.Revoke(actor, reason); err != nil {
			s.logger.Error("updating appeal status", "id", access.Appeal.ID, "error", err)
			return
		}

		if err := s.providerService.RevokeAccess(ctx, *access); err != nil {
			done <- access
			s.logger.Error("failed to revoke access in provider", "id", access.ID, "error", err)
			return
		}

		revokedAccess.Status = domain.AccessStatusInactive
		if err := s.repo.Update(ctx, revokedAccess); err != nil {
			done <- access
			s.logger.Error("failed to update access-revoke status", "id", access.ID, "error", err)
			return
		} else {
			done <- revokedAccess
			s.logger.Info("access revoked", "id", access.ID)
		}
	}
}
