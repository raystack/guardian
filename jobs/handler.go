package jobs

import (
	"context"

	"github.com/odpf/guardian/core/grant"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/plugins/notifiers"
	"github.com/odpf/salt/log"
)

//go:generate mockery --name=grantService --exported
type grantService interface {
	List(context.Context, domain.ListGrantsFilter) ([]domain.Grant, error)
	Revoke(ctx context.Context, id, actor, reason string, opts ...grant.Option) (*domain.Grant, error)
}

//go:generate mockery --name=providerService --exported
type providerService interface {
	FetchResources(context.Context) error
}

type handler struct {
	logger          log.Logger
	grantService    grantService
	providerService providerService
	notifier        notifiers.Client
}

func NewHandler(
	logger log.Logger,
	grantService grantService,
	providerService providerService,
	notifier notifiers.Client,
) *handler {
	return &handler{
		logger:          logger,
		grantService:    grantService,
		providerService: providerService,
		notifier:        notifier,
	}
}
