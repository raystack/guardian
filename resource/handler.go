package resource

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/utils"
)

// Handler for http service
type Handler struct {
	resourceService domain.ResourceService
}

// SetupHandler registers api handlers to the endpoints
func SetupHandler(r *mux.Router, rs domain.ResourceService) {
	h := &Handler{rs}
	r.Methods(http.MethodGet).Path("/resources").HandlerFunc(h.Find)
}

// Find handles http request for list of provider records
func (h *Handler) Find(w http.ResponseWriter, r *http.Request) {
	records, err := h.resourceService.Find()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.ReturnJSON(w, records)
	return
}
