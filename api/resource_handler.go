package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/resource"
	"github.com/odpf/guardian/utils"
)

type resourceUpdatePayload struct {
	Details map[string]interface{} `json:"details"`
	Labels  map[string]interface{} `json:"labels"`
}

func (p *resourceUpdatePayload) toDomain() *domain.Resource {
	return &domain.Resource{
		Details: p.Details,
		Labels:  p.Labels,
	}
}

// ResourceHandler for http service
type ResourceHandler struct {
	ResourceService domain.ResourceService
}

func NewResourceHandler(rs domain.ResourceService) *ResourceHandler {
	return &ResourceHandler{rs}
}

// Find handles http request for list of provider records
func (h *ResourceHandler) Find(w http.ResponseWriter, r *http.Request) {
	records, err := h.ResourceService.Find(map[string]interface{}{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.ReturnJSON(w, records)
	return
}

func (h *ResourceHandler) Update(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if id == 0 {
		http.Error(w, resource.ErrEmptyIDParam.Error(), http.StatusBadRequest)
		return
	}

	var payload resourceUpdatePayload
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
