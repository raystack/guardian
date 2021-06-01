package appeal

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
	AppealService domain.AppealService
}

func NewHTTPHandler(as domain.AppealService) *Handler {
	return &Handler{as}
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) Find(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) GetPendingApprovals(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")

	approvals, err := h.AppealService.GetPendingApprovals(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.ReturnJSON(w, approvals)
	return
}

func (h *Handler) MakeAction(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	appealID, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	approvalName := params["name"]

	var payload actionPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	actor := getActor(r)

	appeal, err := h.AppealService.MakeAction(domain.ApprovalAction{
		AppealID:     uint(appealID),
		ApprovalName: approvalName,
		Actor:        actor,
		Action:       payload.Action,
	})
	if err != nil {
		var statusCode int
		switch err {
		case ErrAppealStatusCanceled,
			ErrAppealStatusApproved,
			ErrAppealStatusRejected,
			ErrAppealStatusTerminated,
			ErrAppealStatusUnrecognized,
			ErrApprovalDependencyIsPending,
			ErrAppealStatusRejected,
			ErrApprovalStatusUnrecognized,
			ErrApprovalStatusApproved,
			ErrApprovalStatusRejected,
			ErrApprovalStatusSkipped,
			ErrActionInvalidValue:
			statusCode = http.StatusBadRequest
		case ErrActionForbidden:
			statusCode = http.StatusForbidden
		case ErrApprovalNameNotFound:
			statusCode = http.StatusNotFound
		default:
			statusCode = http.StatusInternalServerError
		}
		http.Error(w, err.Error(), statusCode)
		return
	}

	utils.ReturnJSON(w, appeal)
	return
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	appealID, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	appeal, err := h.AppealService.Cancel(uint(appealID))
	if err != nil {
		var statusCode int
		switch err {
		case ErrAppealStatusCanceled,
			ErrAppealStatusApproved,
			ErrAppealStatusRejected,
			ErrAppealStatusTerminated,
			ErrAppealStatusUnrecognized:
			statusCode = http.StatusBadRequest
		default:
			statusCode = http.StatusInternalServerError
		}

		http.Error(w, err.Error(), statusCode)
		return
	}

	utils.ReturnJSON(w, appeal)
	return
}

func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	appealID, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	actor := getActor(r)

	appeal, err := h.AppealService.Revoke(uint(appealID), actor)
	if err != nil {
		var statusCode int
		switch err {
		case ErrRevokeAppealForbidden:
			statusCode = http.StatusForbidden
		case ErrAppealNotFound:
			statusCode = http.StatusNotFound
		default:
			statusCode = http.StatusInternalServerError
		}

		http.Error(w, err.Error(), statusCode)
		return
	}

	utils.ReturnJSON(w, appeal)
	return
}

func getActor(r *http.Request) string {
	return r.Header.Get(domain.AuthenticatedEmailHeaderKey)
}
