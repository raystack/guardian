//go:generate mockery --name=appealService --exported
//go:generate mockery --name=providerService --exported

package jobs

import (
	"context"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/plugins/notifiers"
	"github.com/odpf/salt/log"
)

type appealService interface {
	Find(context.Context, *domain.ListAppealsFilter) ([]*domain.Appeal, error)
	Revoke(ctx context.Context, id, actor, reason string) (*domain.Appeal, error)
}

type providerService interface {
	FetchResources(context.Context) error
}

type handler struct {
	logger          log.Logger
	appealService   appealService
	providerService providerService
	notifier        notifiers.Client
}

func NewHandler(
	logger log.Logger,
	appealService appealService,
	providerService providerService,
	notifier notifiers.Client,
) *handler {
	return &handler{
		logger:          logger,
		appealService:   appealService,
		providerService: providerService,
		notifier:        notifier,
	}
}
