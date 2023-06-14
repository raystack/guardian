package jobs

import (
	"context"

	"github.com/raystack/guardian/domain"
	"github.com/raystack/salt/audit"
)

func (h *handler) FetchResources(ctx context.Context, cfg Config) error {
	ctx = audit.WithActor(ctx, domain.SystemActorName)
	h.logger.Info("running fetch resources job")
	return h.providerService.FetchResources(ctx)
}
