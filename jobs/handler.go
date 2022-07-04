//go:generate mockery --name=appealService --exported
//go:generate mockery --name=providerService --exported

package jobs

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/odpf/guardian/plugins/identities"

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

type handler struct {
	logger          log.Logger
	appealService   appealService
	providerService providerService
	policyService   policyService
	notifier        notifiers.Client
	iam             *identities.Manager
}

func NewHandler(
	logger log.Logger,
	appealService appealService,
	providerService providerService,
	policyService policyService,
	notifier notifiers.Client,
	validator *validator.Validate,
	crypto domain.Crypto,
) *handler {
	return &handler{
		logger:          logger,
		appealService:   appealService,
		providerService: providerService,
		policyService:   policyService,
		notifier:        notifier,
		iam:             identities.NewManager(crypto, validator),
	}
}
