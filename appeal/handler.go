package appeal

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/utils"
)

// Handler for http service
type Handler struct {
	AppealService domain.AppealService
}

// SetupHandler registers api handlers to the endpoints
func SetupHandler(r *mux.Router, as domain.AppealService) {
	h := &Handler{as}
	r.Methods(http.MethodPost).Path("/appeals").HandlerFunc(h.Create)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var payload createPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := utils.ValidateStruct(payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	appeals, err := payload.toDomain()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.AppealService.Create(appeals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.ReturnJSON(w, appeals)
	return
}
