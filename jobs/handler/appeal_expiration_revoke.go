package handler

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/odpf/guardian/domain"
)

func (h *handler) RevokeExpiredAppeals() error {
	h.logger.Info("Revoke Expired Appeals")

	filters := &domain.ListAppealsFilter{
		Statuses:               []string{domain.AppealStatusActive},
		ExpirationDateLessThan: time.Now(),
	}

	h.logger.Info("Retrieving active appeals...")

	appeals, err := h.appealService.Find(filters)
	if err != nil {
		return err
	}

	successRevoke := []string{}
	failedRevoke := []map[string]interface{}{}
	for _, a := range appeals {
		h.logger.Info(fmt.Sprintf("Revoking appeal ID: %s", a.ID))

		if _, err := h.appealService.Revoke(a.ID, domain.SystemActorName, "Automatically revoked"); err != nil {
			h.logger.Info(fmt.Sprintf("Failed to revoke appeal ID: %s, error: %s", a.ID, err.Error()))

			failedRevoke = append(failedRevoke, map[string]interface{}{
				"id":    a.ID,
				"error": err.Error(),
			})
		} else {
			h.logger.Info(fmt.Sprintf("Appeal ID %s has been revoked successfully", a.ID))
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

	if len(successRevoke) > 0 || len(failedRevoke) > 0 {
		h.logger.Info(fmt.Sprintf("Done! %v appeals revoked", len(successRevoke)))
		if len(failedRevoke) > 0 {
			h.logger.Info(fmt.Sprintf("But unable to revoke %v appeals", len(failedRevoke)))
		}
		h.logger.Info(string(result))
	} else {
		h.logger.Info("Done! No active appeals revoked")
	}

	return nil
}
