package provider

import (
	"log"

	"github.com/odpf/guardian/domain"
)

// JobHandler for cronjob
type JobHandler struct {
	providerService domain.ProviderService
}

// NewJobHandler returns *JobHandler
func NewJobHandler(ps domain.ProviderService) *JobHandler {
	return &JobHandler{ps}
}

// GetResources fetches all resources for all registered providers
func (h *JobHandler) GetResources() error {
	log.Print("GetResources")
	return h.providerService.FetchResources()
}
