package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/odpf/guardian/domain"
)

func (h *handler) AppealExpirationReminder(ctx context.Context) error {
	h.logger.Info("running appeal expiration reminder job")

	daysBeforeExpired := []int{7, 3, 1}
	for _, d := range daysBeforeExpired {
		h.logger.Info("retrieving active appeals", "expiration_window_in_days", d)

		now := time.Now().AddDate(0, 0, d)
		year, month, day := now.Date()
		from := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
		to := time.Date(year, month, day, 23, 59, 59, 999999999, now.Location())
		filters := &domain.ListAppealsFilter{
			Statuses:                  []string{domain.AppealStatusActive},
			ExpirationDateGreaterThan: from,
			ExpirationDateLessThan:    to,
		}
		appeals, err := h.appealService.Find(ctx, filters)
		if err != nil {
			h.logger.Error("failed to retrieve active appeals",
				"expiration_window_in_days", d,
				"error", err,
			)
			continue
		}

		// TODO: group notifications by username
		var notifications []domain.Notification
		for _, a := range appeals {
			notifications = append(notifications, domain.Notification{
				User: a.CreatedBy,
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
			h.logger.Error("failed to send notifications", "error", err)
			return err
		}
	}

	return nil
}
