package jobs

import (
	"context"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/pkg/audit"
)

func (h *handler) FetchResources(ctx context.Context) error {
	ctx = audit.WithActor(ctx, domain.SystemActorName)
	return h.providerService.FetchResources(ctx)
}
