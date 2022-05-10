package jobs

import (
	"context"
	"time"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/pkg/audit"
)

func (h *handler) RevokeExpiredAppeals(ctx context.Context) error {
	h.logger.Info("running revoke expired appeals job")

	filters := &domain.ListAppealsFilter{
		Statuses:               []string{domain.AppealStatusActive},
		ExpirationDateLessThan: time.Now(),
	}

	h.logger.Info("retrieving active appeals...")
	appeals, err := h.appealService.Find(ctx, filters)
	if err != nil {
		return err
	}

	successRevoke := []string{}
	failedRevoke := []map[string]interface{}{}
	for _, a := range appeals {
		h.logger.Info("revoking appeal", "id", a.ID)

		ctx = audit.WithActor(ctx, "system")
		if _, err := h.appealService.Revoke(ctx, a.ID, domain.SystemActorName, "Automatically revoked"); err != nil {
			h.logger.Error("failed to revoke appeal",
				"id", a.ID,
				"error", err,
			)

			failedRevoke = append(failedRevoke, map[string]interface{}{
				"id":    a.ID,
				"error": err.Error(),
			})
		} else {
			h.logger.Info("appeal revoked", "id", a.ID)
			successRevoke = append(successRevoke, a.ID)
		}
	}

	if err != nil {
		return err
	}

	h.logger.Info("successful appeal revocation",
		"count", len(successRevoke),
		"ids", successRevoke,
	)
	h.logger.Info("failed appeal revocation",
		"count", len(failedRevoke),
		"ids", failedRevoke,
	)

	return nil
}
