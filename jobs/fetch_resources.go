package jobs

import "context"

func (h *handler) FetchResources(ctx context.Context) error {
	return h.providerService.FetchResources(ctx)
}
