package resource

import (
	"encoding/json"
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
	r.Methods(http.MethodPut).Path("/resources").HandlerFunc(h.BulkUpsert)
}

func (h *Handler) BulkUpsert(w http.ResponseWriter, r *http.Request) {
	var payload []*domain.Resource
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.resourceService.BulkUpsert(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.ReturnJSON(w, payload)
	return
}
