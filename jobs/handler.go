//go:generate mockery --name=appealService --exported
//go:generate mockery --name=providerService --exported
//go:generate mockery --name=policyService --exported

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

type policyService interface {
	GetOne(ctx context.Context, id string, version uint) (*domain.Policy, error)
}

type iamManager interface {
	domain.IAMManager
}

type handler struct {
	logger          log.Logger
	appealService   appealService
	providerService providerService
	policyService   policyService
	notifier        notifiers.Client
	iam             domain.IAMManager
}

func NewHandler(
	logger log.Logger,
	appealService appealService,
	providerService providerService,
	policyService policyService,
	notifier notifiers.Client,
	manager iamManager,
) *handler {
	return &handler{
		logger:          logger,
		appealService:   appealService,
		providerService: providerService,
		policyService:   policyService,
		notifier:        notifier,
		iam:             manager,
	}
}
