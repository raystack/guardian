package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/policy"
	"github.com/odpf/guardian/utils"
	"gopkg.in/yaml.v3"
)

type policyCreatePayload struct {
	ID          string                 `yaml:"id" validate:"required"`
	Description string                 `yaml:"description"`
	Steps       []*domain.Step         `yaml:"steps" validate:"required"`
	Labels      map[string]interface{} `yaml:"labels"`
}

func (p *policyCreatePayload) toDomain() *domain.Policy {
	return &domain.Policy{
		ID:          p.ID,
		Description: p.Description,
		Steps:       p.Steps,
		Labels:      p.Labels,
	}
}

type policyUpdatePayload struct {
	Description string                 `yaml:"description"`
	Steps       []*domain.Step         `yaml:"steps" validate:"required"`
	Labels      map[string]interface{} `yaml:"labels"`
}

func (p *policyUpdatePayload) toDomain() *domain.Policy {
	return &domain.Policy{
		Description: p.Description,
		Steps:       p.Steps,
		Labels:      p.Labels,
	}
}

// PolicyHandler for http service
type PolicyHandler struct {
	PolicyService domain.PolicyService
}

func NewPolicyHandler(ps domain.PolicyService) *PolicyHandler {
	return &PolicyHandler{ps}
}

// Create parses http request body to policy domain and passes it to the policy service
func (h *PolicyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var payload policyCreatePayload

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
}

// Find handles http request for list of policy records
func (h *PolicyHandler) Find(w http.ResponseWriter, r *http.Request) {
	policies, err := h.PolicyService.Find()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.ReturnJSON(w, policies)
}

// Update is the http handler for policy update
func (h *PolicyHandler) Update(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	policyID := params["id"]
	if policyID == "" {
		http.Error(w, policy.ErrEmptyIDParam.Error(), http.StatusBadRequest)
		return
	}

	var payload policyUpdatePayload
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
		if err == policy.ErrPolicyDoesNotExists || err == policy.ErrEmptyIDParam {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.ReturnJSON(w, p)
}
