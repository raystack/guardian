package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/odpf/guardian/domain"
	"github.com/odpf/salt/audit"
)

type Key struct {
	AccountId     string
	PolicyId      string
	PolicyVersion uint
}

func (h *handler) DormantAccountAppealRevoke(ctx context.Context) error {
	h.logger.Info("running dormant account appeal revoke job")
	ctx = audit.WithActor(ctx, domain.SystemActorName)

	filters := &domain.ListAppealsFilter{
		AccountType: domain.DefaultAppealAccountType,
		Statuses:    []string{domain.AppealStatusActive},
	}
	appeals, err := h.appealService.Find(ctx, filters)
	if err != nil {
		return err
	}

	if len(appeals) == 0 {
		return nil
	}

	batchSize := 10
	timeLimiter := make(chan int, batchSize)

	for i := 1; i <= batchSize; i++ {
		timeLimiter <- i
	}

	go func() {
		for range time.Tick(1 * time.Second) {
			for i := 1; i <= batchSize; i++ {
				timeLimiter <- i
			}
		}
	}()

	accountAppealsMap := make(map[Key][]*domain.Appeal, 0)
	for _, appeal := range appeals {
		groupKey := Key{
			AccountId:     appeal.AccountID,
			PolicyId:      appeal.PolicyID,
			PolicyVersion: appeal.PolicyVersion,
		}

		var appealList = accountAppealsMap[groupKey]
		if appealList == nil {
			appealList = make([]*domain.Appeal, 0)
		}
		appealList = append(appealList, appeal)
		accountAppealsMap[groupKey] = appealList
	}

	totalRequests := len(accountAppealsMap)
	done := make(chan string, totalRequests)
	for key, appealList := range accountAppealsMap {
		if len(appealList) > 0 {
			policy, err := h.policyService.GetOne(ctx, key.PolicyId, key.PolicyVersion)
			if err != nil {
				h.logger.Error("failed to get policy", "policy_id", key.PolicyId, "policy_version", key.PolicyVersion, "error", err)
			} else {
				iamConfig, err := h.iam.ParseConfig(policy.IAM)
				if err != nil {
					return fmt.Errorf("parsing iam config: %w", err)
				}
				iamClient, err := h.iam.GetClient(iamConfig)
				if err != nil {
					return fmt.Errorf("getting iam client: %w", err)
				}
				go h.expiredDormantUserAppeal(ctx, iamClient, timeLimiter, done, key.AccountId, appealList)
			}
		}
	}

	for {
		if len(done) == totalRequests {
			break
		}
	}
	return nil
}

func (h *handler) expiredDormantUserAppeal(ctx context.Context, iamClient domain.IAMClient, timeLimiter chan int, done chan string, accountId string, appeals []*domain.Appeal) {
	<-timeLimiter
	var successRevoke []string
	var failedRevoke []map[string]interface{}
	isActive, err := iamClient.IsActiveUser(accountId)
	if err != nil {
		h.logger.Error("failed to revoke appeal", "user", accountId, "error", err)
	}

	if !isActive {
		for _, appeal := range appeals {
			if _, err := h.appealService.Revoke(ctx, appeal.ID, domain.SystemActorName, "Automatically revoked since account is dormant"); err != nil {
				h.logger.Error("failed to revoke appeal", "id", appeal.ID, "error", err)
				failedRevoke = append(failedRevoke, map[string]interface{}{"id": appeal.ID, "error": err.Error()})
			} else {
				h.logger.Info("appeal revoked", "id", appeal.ID)
				successRevoke = append(successRevoke, appeal.ID)
			}
		}
	}
	h.logger.Info("successful appeal revocation", "user", accountId, "count", len(successRevoke), "ids", successRevoke)
	h.logger.Info("failed appeal revocation", "user", accountId, "count", len(failedRevoke), "ids", failedRevoke)
	done <- accountId
}
