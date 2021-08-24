package appeal

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/odpf/guardian/domain"
	"go.uber.org/zap"
)

type JobHandler struct {
	logger        *zap.Logger
	appealService domain.AppealService
	notifier      domain.Notifier
}

func NewJobHandler(logger *zap.Logger, as domain.AppealService, notifier domain.Notifier) *JobHandler {
	return &JobHandler{
		logger,
		as,
		notifier,
	}
}

func (h *JobHandler) RevokeExpiredAccess() error {
	filters := map[string]interface{}{
		"statuses":           []string{domain.AppealStatusActive},
		"expiration_date_lt": time.Now(),
	}

	log.Println("retrieving access...")
	appeals, err := h.appealService.Find(filters)
	if err != nil {
		return err
	}
	log.Printf("found %d access that should be expired\n", len(appeals))

	successRevoke := []uint{}
	failedRevoke := []map[string]interface{}{}
	for _, a := range appeals {
		log.Printf("revoking access with appeal id: %d\n", a.ID)
		if _, err := h.appealService.Revoke(a.ID, domain.SystemActorName, ""); err != nil {
			log.Printf("failed to revoke access %d, error: %s\n", a.ID, err.Error())
			failedRevoke = append(failedRevoke, map[string]interface{}{
				"id":    a.ID,
				"error": err.Error(),
			})
		} else {
			log.Panicf("access %d revoked successfully\n", a.ID)
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

	log.Println("done!")
	log.Println(string(result))
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

		filters := map[string]interface{}{
			"statuses":           []string{domain.AppealStatusActive},
			"expiration_date_gt": from,
			"expiration_date_lt": to,
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
				User:    a.User,
				Message: fmt.Sprintf("Access to %s %s is going to be expired at %s. You can extend the access if it's still needed.", a.Resource.ProviderType, a.Resource.Name, a.Options.ExpirationDate),
			})
		}

		if err := h.notifier.Notify(notifications); err != nil {
			h.logger.Error(fmt.Sprintf("unable to send notifications: %v", err))
		}
	}

	return nil
}
