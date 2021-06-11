package provider

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/utils"
	"gopkg.in/yaml.v3"
)

// Handler for http service
type Handler struct {
	ProviderService domain.ProviderService
}

func NewHTTPHandler(ps domain.ProviderService) *Handler {
	return &Handler{ps}
}

// Create parses http request body to provider domain and passes it to the provider service
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var providerConfig domain.ProviderConfig

	if err := yaml.NewDecoder(r.Body).Decode(&providerConfig); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := utils.ValidateStruct(providerConfig); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p := &domain.Provider{
		Type:   providerConfig.Type,
		URN:    providerConfig.URN,
		Config: &providerConfig,
	}
	if err := h.ProviderService.Create(p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.ReturnJSON(w, p)
	return
}

// Find handles http request for list of provider records
func (h *Handler) Find(w http.ResponseWriter, r *http.Request) {
	records, err := h.ProviderService.Find()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, r := range records {
		r.Config.Credentials = nil
	}

	utils.ReturnJSON(w, records)
	return
}

// Update handles http request for provider update
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
	if err := yaml.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := utils.ValidateStruct(payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p := payload.toDomain()
	p.ID = uint(id)
	if err := h.ProviderService.Update(p); err != nil {
		if err == ErrRecordNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.ReturnJSON(w, p)
	return
}
