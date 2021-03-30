package resource

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/utils"
)

// Handler for http service
type Handler struct {
	ResourceService domain.ResourceService
}

// SetupHandler registers api handlers to the endpoints
func SetupHandler(r *mux.Router, rs domain.ResourceService) {
	h := &Handler{rs}
	r.Methods(http.MethodGet).Path("/resources").HandlerFunc(h.Find)
	r.Methods(http.MethodPut).Path("/resources/{id}").HandlerFunc(h.Update)
}

// Find handles http request for list of provider records
func (h *Handler) Find(w http.ResponseWriter, r *http.Request) {
	records, err := h.ResourceService.Find(map[string]interface{}{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.ReturnJSON(w, records)
	return
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if id == 0 {
		http.Error(w, ErrEmptyIDParam.Error(), http.StatusBadRequest)
		return
	}

	var payload updatePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := utils.ValidateStruct(payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res := payload.toDomain()
	res.ID = uint(id)
	if err := h.ResourceService.Update(res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.ReturnJSON(w, res)
	return
}
