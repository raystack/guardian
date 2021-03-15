package resource

import (
	"github.com/gorilla/mux"
	"github.com/odpf/guardian/domain"
)

// Handler for http service
type Handler struct {
	resourceService domain.ResourceService
}

// SetupHandler registers api handlers to the endpoints
func SetupHandler(r *mux.Router, rs domain.ResourceService) {}
