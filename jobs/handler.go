package jobs

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/raystack/guardian/core/grant"
	"github.com/raystack/guardian/domain"
	"github.com/raystack/guardian/plugins/notifiers"
	"github.com/raystack/salt/log"
)

//go:generate mockery --name=grantService --exported --with-expecter
type grantService interface {
	List(context.Context, domain.ListGrantsFilter) ([]domain.Grant, error)
	Revoke(ctx context.Context, id, actor, reason string, opts ...grant.Option) (*domain.Grant, error)
	BulkRevoke(ctx context.Context, filter domain.RevokeGrantsFilter, actor, reason string) ([]*domain.Grant, error)
	Update(context.Context, *domain.Grant) error
	DormancyCheck(context.Context, domain.DormancyCheckCriteria) error
}

//go:generate mockery --name=providerService --exported
type providerService interface {
	FetchResources(context.Context) error
	Find(context.Context, domain.ProviderFilter) ([]*domain.Provider, error)
}

type crypto interface {
	domain.Crypto
}

type handler struct {
	logger          log.Logger
	grantService    grantService
	providerService providerService
	notifier        notifiers.Client
	crypto          crypto
	validator       *validator.Validate
}

func NewHandler(
	logger log.Logger,
	grantService grantService,
	providerService providerService,
	notifier notifiers.Client,
	crypto crypto,
	validator *validator.Validate,
) *handler {
	return &handler{
		logger:          logger,
		grantService:    grantService,
		providerService: providerService,
		notifier:        notifier,
		crypto:          crypto,
		validator:       validator,
	}
}
