package policy

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/utils"
	"gopkg.in/yaml.v3"
)

// Handler for http service
type Handler struct {
	PolicyService domain.PolicyService
}

// SetupHandler registers api handlers to the endpoints
func SetupHandler(r *mux.Router, ps domain.PolicyService) {
	h := &Handler{ps}
	r.Methods(http.MethodGet).Path("/policies").HandlerFunc(h.Find)
	r.Methods(http.MethodPost).Path("/policies").HandlerFunc(h.Create)
	r.Methods(http.MethodPut).Path("/policies/{id}").HandlerFunc(h.Update)
}

// Create parses http request body to policy domain and passes it to the policy service
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var payload createPayload

	if err := yaml.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := utils.ValidateStruct(payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p := payload.toDomain()
	if err := h.PolicyService.Create(p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.ReturnJSON(w, p)
	return
}

// Find handles http request for list of policy records
func (h *Handler) Find(w http.ResponseWriter, r *http.Request) {
	policies, err := h.PolicyService.Find()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.ReturnJSON(w, policies)
	return
}

// Update is the http handler for policy update
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	policyID := params["id"]
	if policyID == "" {
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
	p.ID = policyID
	if err := h.PolicyService.Update(p); err != nil {
		if err == ErrPolicyDoesNotExists || err == ErrEmptyIDParam {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.ReturnJSON(w, p)
	return
}
