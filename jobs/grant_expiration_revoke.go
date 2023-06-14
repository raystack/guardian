package jobs

import (
	"context"
	"time"

	"github.com/raystack/guardian/domain"
	"github.com/raystack/salt/audit"
)

func (h *handler) RevokeExpiredGrants(ctx context.Context) error {
	h.logger.Info("running revoke expired grants job")

	falseBool := false
	filters := domain.ListGrantsFilter{
		Statuses:               []string{string(domain.GrantStatusActive)},
		ExpirationDateLessThan: time.Now(),
		IsPermanent:            &falseBool,
	}

	h.logger.Info("retrieving active grant...")
	grants, err := h.grantService.List(ctx, filters)
	if err != nil {
		return err
	}

	successRevoke := []string{}
	failedRevoke := []map[string]interface{}{}
	for _, g := range grants {
		h.logger.Info("revoking grant", "id", g.ID)

		ctx = audit.WithActor(ctx, domain.SystemActorName)
		if _, err := h.grantService.Revoke(ctx, g.ID, domain.SystemActorName, "Automatically revoked"); err != nil {
			h.logger.Error("failed to revoke grant",
				"id", g.ID,
				"error", err,
			)

			failedRevoke = append(failedRevoke, map[string]interface{}{
				"id":    g.ID,
				"error": err.Error(),
			})
		} else {
			h.logger.Info("grant revoked", "id", g.ID)
			successRevoke = append(successRevoke, g.ID)
		}
	}

	if err != nil {
		return err
	}

	h.logger.Info("successful grant revocation", "count", len(successRevoke), "ids", successRevoke)
	if len(failedRevoke) > 0 {
		h.logger.Info("failed grant revocation", "count", len(failedRevoke), "ids", failedRevoke)
	}

	return nil
}
