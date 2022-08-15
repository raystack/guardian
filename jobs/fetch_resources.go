package jobs

import (
	"context"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/salt/audit"
)

func (h *handler) FetchResources(ctx context.Context) error {
	ctx = audit.WithActor(ctx, domain.SystemActorName)
	h.logger.Info("running fetch resources job")
	return h.providerService.FetchResources(ctx)
}
