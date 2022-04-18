package jobs

import (
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/plugins/notifiers"
	"github.com/odpf/salt/log"
)

type appealService interface {
	Find(*domain.ListAppealsFilter) ([]*domain.Appeal, error)
	Revoke(id, actor, reason string) (*domain.Appeal, error)
	Create(appeals []*domain.Appeal) error
}

type providerService interface {
	FetchResources() error
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
