package api

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/provider"
	"github.com/odpf/guardian/utils"
	"gopkg.in/yaml.v3"
)

type providerUpdatePayload struct {
	Labels      map[string]string        `yaml:"labels"`
	Credentials interface{}              `yaml:"credentials"`
	Appeal      *domain.AppealConfig     `yaml:"appeal"`
	Resources   []*domain.ResourceConfig `yaml:"resources"`
}

func (p *providerUpdatePayload) toDomain() *domain.Provider {
	return &domain.Provider{
		Config: &domain.ProviderConfig{
			Labels:      p.Labels,
			Credentials: p.Credentials,
			Appeal:      p.Appeal,
			Resources:   p.Resources,
		},
	}
}

// ProviderHandler for http service
type ProviderHandler struct {
	ProviderService domain.ProviderService
}

func NewProviderHandler(ps domain.ProviderService) *ProviderHandler {
	return &ProviderHandler{ps}
}

// Create parses http request body to provider domain and passes it to the provider service
func (h *ProviderHandler) Create(w http.ResponseWriter, r *http.Request) {
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
}

// Find handles http request for list of provider records
func (h *ProviderHandler) Find(w http.ResponseWriter, r *http.Request) {
	records, err := h.ProviderService.Find()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, r := range records {
		r.Config.Credentials = nil
	}

	utils.ReturnJSON(w, records)
}

// Update handles http request for provider update
func (h *ProviderHandler) Update(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if id == 0 {
		http.Error(w, provider.ErrEmptyIDParam.Error(), http.StatusBadRequest)
		return
	}

	var payload providerUpdatePayload
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
		if err == provider.ErrRecordNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.ReturnJSON(w, p)
}
