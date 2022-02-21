package provider

type ProviderService interface {
	FetchResources() error
}

// JobHandler for cronjob
type JobHandler struct {
	providerService ProviderService
}

// NewJobHandler returns *JobHandler
func NewJobHandler(ps ProviderService) *JobHandler {
	return &JobHandler{ps}
}

// GetResources fetches all resources for all registered providers
func (h *JobHandler) GetResources() error {
	return h.providerService.FetchResources()
}
