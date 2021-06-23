package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/appeal"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/utils"
)

var TimeNow = time.Now

type resourceOptions struct {
	Duration string `json:"duration"`
}

type appealResource struct {
	ID      uint                   `json:"id" validate:"required"`
	Role    string                 `json:"role" validate:"required"`
	Options map[string]interface{} `json:"options"`
}

type appealCreatePayload struct {
	User      string           `json:"user" validate:"required"`
	Resources []appealResource `json:"resources" validate:"required,min=1"`
}

func (p *appealCreatePayload) toDomain() ([]*domain.Appeal, error) {
	appeals := []*domain.Appeal{}
	for _, r := range p.Resources {
		var options *domain.AppealOptions

		var resOptions *resourceOptions
		if err := mapstructure.Decode(r.Options, &resOptions); err != nil {
			return nil, err
		}
		if resOptions != nil {
			if err := utils.ValidateStruct(resOptions); err != nil {
				return nil, err
			}
			var expirationDate time.Time
			if resOptions.Duration != "" {
				duration, err := time.ParseDuration(resOptions.Duration)
				if err != nil {
					return nil, err
				}
				expirationDate = TimeNow().Add(duration)
			}

			options = &domain.AppealOptions{
				ExpirationDate: &expirationDate,
			}
		}

		appeals = append(appeals, &domain.Appeal{
			User:       p.User,
			ResourceID: r.ID,
			Role:       r.Role,
			Options:    options,
		})
	}

	return appeals, nil
}

type appealActionPayload struct {
	Action string `json:"action"`
}

const (
	AuthenticatedEmailHeaderKey = "X-Goog-Authenticated-User-Email"
)

// AppealHandler for http service
type AppealHandler struct {
	AppealService domain.AppealService
}

func NewAppealHandler(as domain.AppealService) *AppealHandler {
	return &AppealHandler{as}
}

func (h *AppealHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	a, err := h.AppealService.GetByID(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if a == nil {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	utils.ReturnJSON(w, a)
	return
}

func (h *AppealHandler) Find(w http.ResponseWriter, r *http.Request) {
	filters := map[string]interface{}{
		"user": r.URL.Query().Get("user"),
	}

	records, err := h.AppealService.Find(filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.ReturnJSON(w, records)
	return
}

func (h *AppealHandler) Create(w http.ResponseWriter, r *http.Request) {
	var payload appealCreatePayload
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
		status := http.StatusInternalServerError
		if errors.Is(err, appeal.ErrAppealDuplicate) {
			status = http.StatusConflict
		}

		http.Error(w, err.Error(), status)
		return
	}

	utils.ReturnJSON(w, appeals)
	return
}

func (h *AppealHandler) GetPendingApprovals(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")

	approvals, err := h.AppealService.GetPendingApprovals(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.ReturnJSON(w, approvals)
	return
}

func (h *AppealHandler) MakeAction(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	appealID, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	approvalName := params["name"]

	var payload appealActionPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	actor := getActor(r)

	a, err := h.AppealService.MakeAction(domain.ApprovalAction{
		AppealID:     uint(appealID),
		ApprovalName: approvalName,
		Actor:        actor,
		Action:       payload.Action,
	})
	if err != nil {
		var statusCode int
		switch err {
		case appeal.ErrAppealStatusCanceled,
			appeal.ErrAppealStatusApproved,
			appeal.ErrAppealStatusRejected,
			appeal.ErrAppealStatusTerminated,
			appeal.ErrAppealStatusUnrecognized,
			appeal.ErrApprovalDependencyIsPending,
			appeal.ErrAppealStatusRejected,
			appeal.ErrApprovalStatusUnrecognized,
			appeal.ErrApprovalStatusApproved,
			appeal.ErrApprovalStatusRejected,
			appeal.ErrApprovalStatusSkipped,
			appeal.ErrActionInvalidValue:
			statusCode = http.StatusBadRequest
		case appeal.ErrActionForbidden:
			statusCode = http.StatusForbidden
		case appeal.ErrApprovalNameNotFound:
			statusCode = http.StatusNotFound
		default:
			statusCode = http.StatusInternalServerError
		}
		http.Error(w, err.Error(), statusCode)
		return
	}

	utils.ReturnJSON(w, a)
	return
}

func (h *AppealHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	appealID, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	a, err := h.AppealService.Cancel(uint(appealID))
	if err != nil {
		var statusCode int
		switch err {
		case appeal.ErrAppealStatusCanceled,
			appeal.ErrAppealStatusApproved,
			appeal.ErrAppealStatusRejected,
			appeal.ErrAppealStatusTerminated,
			appeal.ErrAppealStatusUnrecognized:
			statusCode = http.StatusBadRequest
		default:
			statusCode = http.StatusInternalServerError
		}

		http.Error(w, err.Error(), statusCode)
		return
	}

	utils.ReturnJSON(w, a)
	return
}

func (h *AppealHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	appealID, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	actor := getActor(r)

	a, err := h.AppealService.Revoke(uint(appealID), actor)
	if err != nil {
		var statusCode int
		switch err {
		case appeal.ErrRevokeAppealForbidden:
			statusCode = http.StatusForbidden
		case appeal.ErrAppealNotFound:
			statusCode = http.StatusNotFound
		default:
			statusCode = http.StatusInternalServerError
		}

		http.Error(w, err.Error(), statusCode)
		return
	}

	utils.ReturnJSON(w, a)
	return
}

func getActor(r *http.Request) string {
	return r.Header.Get(AuthenticatedEmailHeaderKey)
}
