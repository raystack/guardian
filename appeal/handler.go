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

// SetupHandler registers api handlers to the endpoints
func SetupHandler(r *mux.Router, as domain.AppealService) {
	h := &Handler{as}
	r.Methods(http.MethodPost).Path("/appeals").HandlerFunc(h.Create)
	r.Methods(http.MethodGet).Path("/appeals").HandlerFunc(h.Find)
	r.Methods(http.MethodGet).Path("/appeals/approvals").HandlerFunc(h.GetPendingApprovals)
	r.Methods(http.MethodPost).Path("/appeals/{id}/approvals/{name}").HandlerFunc(h.MakeAction)
	r.Methods(http.MethodPut).Path("/appeals/{id}/cancel").HandlerFunc(h.Cancel)
	r.Methods(http.MethodGet).Path("/appeals/{id}").HandlerFunc(h.GetByID)
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

	appeal, err := h.AppealService.MakeAction(domain.ApprovalAction{
		AppealID:     uint(appealID),
		ApprovalName: approvalName,
		Actor:        payload.Actor,
		Action:       payload.Action,
	})
	if err != nil {
		var statusCode int
		switch err {
		case ErrAppealStatusCancelled,
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
		case ErrAppealStatusCancelled,
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
