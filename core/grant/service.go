package grant

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
	AuditKeyRevoke = "grant.revoke"
)

//go:generate mockery --name=repository --exported --with-expecter
type repository interface {
	List(context.Context, domain.ListGrantsFilter) ([]domain.Grant, error)
	GetByID(context.Context, string) (*domain.Grant, error)
	Update(context.Context, *domain.Grant) error
}

//go:generate mockery --name=providerService --exported --with-expecter
type providerService interface {
	RevokeAccess(context.Context, domain.Grant) error
}

//go:generate mockery --name=auditLogger --exported --with-expecter
type auditLogger interface {
	Log(ctx context.Context, action string, data interface{}) error
}

//go:generate mockery --name=notifier --exported --with-expecter
type notifier interface {
	notifiers.Client
}

type grantCreation struct {
	AppealStatus string `validate:"required,eq=approved"`
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

func (s *Service) List(ctx context.Context, filter domain.ListGrantsFilter) ([]domain.Grant, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) GetByID(ctx context.Context, id string) (*domain.Grant, error) {
	if id == "" {
		return nil, ErrEmptyIDParam
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Prepare(ctx context.Context, appeal domain.Appeal) (*domain.Grant, error) {
	if err := s.validator.Struct(grantCreation{
		AppealStatus: appeal.Status,
		AccountID:    appeal.AccountID,
		AccountType:  appeal.AccountType,
		ResourceID:   appeal.ResourceID,
	}); err != nil {
		return nil, fmt.Errorf("validating appeal: %w", err)
	}

	return appeal.ToGrant()
}

func (s *Service) Revoke(ctx context.Context, id, actor, reason string, opts ...Option) (*domain.Grant, error) {
	grant, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting grant details: %w", err)
	}

	revokedGrant := &domain.Grant{}
	*revokedGrant = *grant
	if err := grant.Revoke(actor, reason); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, grant); err != nil {
		return nil, fmt.Errorf("updating grant record in db: %w", err)
	}

	options := s.getOptions(opts...)

	if !options.skipRevokeInProvider {
		if err := s.providerService.RevokeAccess(ctx, *grant); err != nil {
			if err := s.repo.Update(ctx, grant); err != nil {
				return nil, fmt.Errorf("failed to rollback grant status: %w", err)
			}
			return nil, fmt.Errorf("removing grant in provider: %w", err)
		}
	}

	if !options.skipNotification {
		if errs := s.notifier.Notify([]domain.Notification{{
			User: grant.CreatedBy,
			Message: domain.NotificationMessage{
				Type: domain.NotificationTypeAccessRevoked,
				Variables: map[string]interface{}{
					"resource_name": fmt.Sprintf("%s (%s: %s)", grant.Resource.Name, grant.Resource.ProviderType, grant.Resource.URN),
					"role":          grant.Role,
					"account_type":  grant.AccountType,
					"account_id":    grant.AccountID,
				},
			},
		}}); errs != nil {
			for _, err1 := range errs {
				s.logger.Error("failed to send notifications", "error", err1.Error())
			}
		}
	}

	if err := s.auditLogger.Log(ctx, AuditKeyRevoke, map[string]interface{}{
		"grant_id": id,
		"reason":   reason,
	}); err != nil {
		s.logger.Error("failed to record audit log", "error", err)
	}

	return grant, nil
}

func (s *Service) BulkRevoke(ctx context.Context, filter domain.RevokeGrantsFilter, actor, reason string) ([]*domain.Grant, error) {
	if filter.AccountIDs == nil || len(filter.AccountIDs) == 0 {
		return nil, fmt.Errorf("account_ids is required")
	}

	grants, err := s.List(ctx, domain.ListGrantsFilter{
		Statuses:      []string{string(domain.GrantStatusActive)},
		AccountIDs:    filter.AccountIDs,
		ProviderTypes: filter.ProviderTypes,
		ProviderURNs:  filter.ProviderURNs,
		ResourceTypes: filter.ResourceTypes,
		ResourceURNs:  filter.ResourceURNs,
	})
	if err != nil {
		return nil, fmt.Errorf("listing active grants: %w", err)
	}
	if len(grants) == 0 {
		return nil, nil
	}

	result := make([]*domain.Grant, 0)
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

	totalRequests := len(grants)
	done := make(chan *domain.Grant, totalRequests)
	resourceGrantMap := make(map[string][]*domain.Grant, 0)

	for i, grant := range grants {
		var resourceGrants []*domain.Grant
		var ok bool
		if resourceGrants, ok = resourceGrantMap[grant.ResourceID]; ok {
			resourceGrants = append(resourceGrants, &grants[i])
		} else {
			resourceGrants = []*domain.Grant{&grants[i]}
		}
		resourceGrantMap[grant.ResourceID] = resourceGrants
	}

	for _, resourceGrants := range resourceGrantMap {
		go s.expiredInActiveUserAccess(ctx, timeLimiter, done, actor, reason, resourceGrants)
	}

	var successRevoke []string
	var failedRevoke []string
	for {
		select {
		case grant := <-done:
			if grant.Status == domain.GrantStatusInactive {
				successRevoke = append(successRevoke, grant.ID)
			} else {
				failedRevoke = append(failedRevoke, grant.ID)
			}
			result = append(result, grant)
			if len(result) == totalRequests {
				s.logger.Info("successful grant revocation", "count", len(successRevoke), "ids", successRevoke)
				s.logger.Info("failed grant revocation", "count", len(failedRevoke), "ids", failedRevoke)
				return result, nil
			}
		}
	}
}

func (s *Service) expiredInActiveUserAccess(ctx context.Context, timeLimiter chan int, done chan *domain.Grant, actor string, reason string, grants []*domain.Grant) {
	for _, grant := range grants {
		<-timeLimiter

		revokedGrant := &domain.Grant{}
		*revokedGrant = *grant
		if err := revokedGrant.Revoke(actor, reason); err != nil {
			s.logger.Error("failed to revoke grant", "id", grant.ID, "error", err)
			return
		}
		if err := s.providerService.RevokeAccess(ctx, *grant); err != nil {
			done <- grant
			s.logger.Error("failed to revoke grant in provider", "id", grant.ID, "error", err)
			return
		}

		revokedGrant.Status = domain.GrantStatusInactive
		if err := s.repo.Update(ctx, revokedGrant); err != nil {
			done <- grant
			s.logger.Error("failed to update access-revoke status", "id", grant.ID, "error", err)
			return
		} else {
			done <- revokedGrant
			s.logger.Info("grant revoked", "id", grant.ID)
		}
	}
}
