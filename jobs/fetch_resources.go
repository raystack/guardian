package jobs

import (
	"context"

	"github.com/goto/guardian/domain"
	"github.com/goto/salt/audit"
)

func (h *handler) FetchResources(ctx context.Context, cfg Config) error {
	ctx = audit.WithActor(ctx, domain.SystemActorName)
	h.logger.Info(ctx, "running fetch resources job")
	return h.providerService.FetchResources(ctx)
}
