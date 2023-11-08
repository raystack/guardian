package jobs

import (
	"context"
	"time"

	"github.com/goto/guardian/domain"
	"github.com/goto/salt/audit"
)

func (h *handler) RevokeExpiredGrants(ctx context.Context, cfg Config) error {
	h.logger.Info(ctx, "running revoke expired grants job")

	falseBool := false
	filters := domain.ListGrantsFilter{
		Statuses:               []string{string(domain.GrantStatusActive)},
		ExpirationDateLessThan: time.Now(),
		IsPermanent:            &falseBool,
	}

	h.logger.Info(ctx, "retrieving active grants...")
	grants, err := h.grantService.List(ctx, filters)
	if err != nil {
		return err
	}
	h.logger.Info(ctx, "retrieved active grants", "count", len(grants))

	successRevoke := []string{}
	failedRevoke := []map[string]interface{}{}
	for _, g := range grants {
		h.logger.Info(ctx, "revoking grant", "id", g.ID)

		ctx = audit.WithActor(ctx, domain.SystemActorName)
		if _, err := h.grantService.Revoke(ctx, g.ID, domain.SystemActorName, "Automatically revoked"); err != nil {
			h.logger.Error(ctx, "failed to revoke grant",
				"id", g.ID,
				"error", err,
			)

			failedRevoke = append(failedRevoke, map[string]interface{}{
				"id":    g.ID,
				"error": err.Error(),
			})
		} else {
			h.logger.Info(ctx, "grant revoked", "id", g.ID)
			successRevoke = append(successRevoke, g.ID)
		}
	}

	if err != nil {
		return err
	}

	h.logger.Info(ctx, "successful grant revocation", "count", len(successRevoke), "ids", successRevoke)
	if len(failedRevoke) > 0 {
		h.logger.Info(ctx, "failed grant revocation", "count", len(failedRevoke), "ids", failedRevoke)
	}

	return nil
}
