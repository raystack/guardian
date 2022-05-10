package jobs

import (
	"context"

	"github.com/odpf/guardian/pkg/audit"
)

func (h *handler) FetchResources(ctx context.Context) error {
	ctx = audit.WithActor(ctx, "system")
	return h.providerService.FetchResources(ctx)
}
