package provider

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/utils"
	"gopkg.in/yaml.v3"
)

// Handler for http service
type Handler struct {
	ProviderService domain.ProviderService
}

// SetupHandler registers api handlers to the endpoints
func SetupHandler(r *mux.Router, ps domain.ProviderService) {
	h := &Handler{ps}
	r.Methods(http.MethodPost).Path("/provider").HandlerFunc(h.Create)
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
