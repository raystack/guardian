package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/raystack/guardian/domain"
)

func (h *handler) GrantExpirationReminder(ctx context.Context, cfg Config) error {
	h.logger.Info("running grant expiration reminder job")

	daysBeforeExpired := []int{7, 3, 1}
	for _, d := range daysBeforeExpired {
		h.logger.Info("retrieving active grants", "expiration_window_in_days", d)

		now := time.Now().AddDate(0, 0, d)
		year, month, day := now.Date()
		from := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
		to := time.Date(year, month, day, 23, 59, 59, 999999999, now.Location())
		filters := domain.ListGrantsFilter{
			Statuses:                  []string{string(domain.GrantStatusActive)},
			ExpirationDateGreaterThan: from,
			ExpirationDateLessThan:    to,
		}
		grants, err := h.grantService.List(ctx, filters)
		if err != nil {
			h.logger.Error("failed to retrieve active grants",
				"expiration_window_in_days", d,
				"error", err,
			)
			continue
		}

		// TODO: group notifications by username
		var notifications []domain.Notification
		for _, g := range grants {
			notifications = append(notifications, domain.Notification{
				User: g.CreatedBy,
				Labels: map[string]string{
					"appeal_id": g.AppealID,
					"grant_id":  g.ID,
				},
				Message: domain.NotificationMessage{
					Type: domain.NotificationTypeExpirationReminder,
					Variables: map[string]interface{}{
						"resource_name":             fmt.Sprintf("%s (%s: %s)", g.Resource.Name, g.Resource.ProviderType, g.Resource.URN),
						"role":                      g.Role,
						"expiration_date":           g.ExpirationDate,
						"expiration_date_formatted": g.ExpirationDate.Format("Jan 02, 2006 15:04:05 UTC"),
						"account_id":                g.AccountID,
						"requestor":                 g.Owner,
					},
				},
			})
		}

		if errs := h.notifier.Notify(notifications); errs != nil {
			for _, err1 := range errs {
				h.logger.Error("failed to send notifications", "error", err1)
			}
		}
	}

	return nil
}
