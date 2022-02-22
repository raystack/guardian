package jobs

func (h *handler) FetchResources() error {
	return h.providerService.FetchResources()
}
