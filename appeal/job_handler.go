package appeal

import (
	"encoding/json"
	"log"
	"time"

	"github.com/odpf/guardian/domain"
)

type JobHandler struct {
	appealService domain.AppealService
}

func NewJobHandler(as domain.AppealService) *JobHandler {
	return &JobHandler{as}
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
		if _, err := h.appealService.Revoke(a.ID, domain.SystemActorName); err != nil {
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
