package appeal

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/plugins/notifiers"
	"github.com/odpf/salt/log"
)

type AppealService interface {
	Find(*domain.ListAppealsFilter) ([]*domain.Appeal, error)
	Revoke(id, actor, reason string) (*domain.Appeal, error)
}

type JobHandler struct {
	logger        log.Logger
	appealService AppealService
	notifier      notifiers.Client
}

func NewJobHandler(logger log.Logger, as AppealService, notifier notifiers.Client) *JobHandler {
	return &JobHandler{
		logger,
		as,
		notifier,
	}
}

func (h *JobHandler) RevokeExpiredAccess() error {
	filters := &domain.ListAppealsFilter{
		Statuses:               []string{domain.AppealStatusActive},
		ExpirationDateLessThan: time.Now(),
	}

	h.logger.Info("retrieving access...")
	appeals, err := h.appealService.Find(filters)
	if err != nil {
		return err
	}
	h.logger.Info(fmt.Sprintf("found %d access that should be expired\n", len(appeals)))

	successRevoke := []string{}
	failedRevoke := []map[string]interface{}{}
	for _, a := range appeals {
		h.logger.Info(fmt.Sprintf("revoking access with appeal id: %s\n", a.ID))
		if _, err := h.appealService.Revoke(a.ID, domain.SystemActorName, "Automatically revoked"); err != nil {
			h.logger.Info(fmt.Sprintf("failed to revoke access %s, error: %s\n", a.ID, err.Error()))
			failedRevoke = append(failedRevoke, map[string]interface{}{
				"id":    a.ID,
				"error": err.Error(),
			})
		} else {
			h.logger.Info(fmt.Sprintf("access %s revoked successfully\n", a.ID))
			successRevoke = append(successRevoke, a.ID)
		}
	}

	result, err := json.Marshal(map[string]interface{}{
		"success": successRevoke,
		"failed":  failedRevoke,
	})
	if err != nil {
		return err
	}

	h.logger.Info("done!")
	h.logger.Info(string(result))
	return nil
}

func (h *JobHandler) NotifyAboutToExpireAccess() error {
	daysBeforeExpired := []int{7, 3, 1}
	for _, d := range daysBeforeExpired {
		h.logger.Info(fmt.Sprintf("collecting access that will expire in %v day(s)", d))

		now := time.Now().AddDate(0, 0, d)
		year, month, day := now.Date()
		from := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
		to := time.Date(year, month, day, 23, 59, 59, 999999999, now.Location())

		filters := &domain.ListAppealsFilter{
			Statuses:                  []string{domain.AppealStatusActive},
			ExpirationDateGreaterThan: from,
			ExpirationDateLessThan:    to,
		}

		appeals, err := h.appealService.Find(filters)
		if err != nil {
			h.logger.Error(fmt.Sprintf("unable to list appeals: %v", err))
			continue
		}

		// TODO: group notifications by username

		var notifications []domain.Notification
		for _, a := range appeals {
			notifications = append(notifications, domain.Notification{
				User: a.AccountID,
				Message: domain.NotificationMessage{
					Type: domain.NotificationTypeExpirationReminder,
					Variables: map[string]interface{}{
						"resource_name":   fmt.Sprintf("%s (%s: %s)", a.Resource.Name, a.Resource.ProviderType, a.Resource.URN),
						"role":            a.Role,
						"expiration_date": *a.Options.ExpirationDate,
					},
				},
			})
		}

		if err := h.notifier.Notify(notifications); err != nil {
			h.logger.Error(fmt.Sprintf("unable to send notifications: %v", err))
		}
	}

	return nil
}
