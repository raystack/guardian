package jobs

import (
	"context"

	"github.com/goto/guardian/core/grant"
	"github.com/goto/guardian/domain"
	"github.com/goto/guardian/plugins/notifiers"
	"github.com/goto/salt/log"
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
