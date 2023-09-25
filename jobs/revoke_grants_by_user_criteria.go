package jobs

import (
	"context"
	"fmt"

	"github.com/raystack/guardian/domain"
	"github.com/raystack/guardian/pkg/evaluator"
	"github.com/raystack/guardian/plugins/identities"
)

type RevokeGrantsByUserCriteriaConfig struct {
	IAM                 domain.IAMConfig     `mapstructure:"iam"`
	UserCriteria        evaluator.Expression `mapstructure:"user_criteria"`
	ReassignOwnershipTo evaluator.Expression `mapstructure:"reassign_ownership_to"`
	DryRun              bool                 `mapstructure:"dry_run"`
}

func (h *handler) RevokeGrantsByUserCriteria(ctx context.Context, c Config) error {
	h.logger.Info(fmt.Sprintf("starting %q job", TypeRevokeGrantsByUserCriteria))
	defer h.logger.Info(fmt.Sprintf("finished %q job", TypeRevokeGrantsByUserCriteria))

	var cfg RevokeGrantsByUserCriteriaConfig
	if err := c.Decode(&cfg); err != nil {
		return fmt.Errorf("invalid config for %s job: %w", TypeRevokeGrantsByUserCriteria, err)
	}

	iamManager := identities.NewManager(h.crypto, h.validator)
	iamConfig, err := iamManager.ParseConfig(&cfg.IAM)
	if err != nil {
		return fmt.Errorf("parsing IAM config: %w", err)
	}
	iamClient, err := iamManager.GetClient(iamConfig)
	if err != nil {
		return fmt.Errorf("initializing IAM client: %w", err)
	}

	h.logger.Info("getting active grants")
	activeGrants, err := h.grantService.List(ctx, domain.ListGrantsFilter{
		Statuses: []string{string(domain.GrantStatusActive)},
	})
	if err != nil {
		return fmt.Errorf("listing active grants: %w", err)
	}
	if len(activeGrants) == 0 {
		h.logger.Info("no active grants found")
		return nil
	}
	grantIDs := getGrantIDs(activeGrants)
	h.logger.Info(fmt.Sprintf("found %d active grants", len(activeGrants)), "grant_ids", grantIDs)

	grantsForUser := map[string][]*domain.Grant{}     // map[account_id][]grant
	grantsOwnedByUser := map[string][]*domain.Grant{} // map[owner][]grant
	uniqueUserEmails := map[string]bool{}             // map[account_id]bool
	for _, g := range activeGrants {
		if g.AccountType == domain.DefaultAppealAccountType {
			// collecting grants for individual users
			grantsForUser[g.AccountID] = append(grantsForUser[g.AccountID], &g)
			uniqueUserEmails[g.AccountID] = true
		} else if g.Owner != domain.SystemActorName {
			// collecting other grants owned by the user
			grantsOwnedByUser[g.Owner] = append(grantsOwnedByUser[g.AccountID], &g)
		}
	}
	h.logger.Info(fmt.Sprintf("found %d unique users", len(uniqueUserEmails)), "emails", uniqueUserEmails)

	counter := 0
	for email := range uniqueUserEmails {
		counter++
		fmt.Println("")
		h.logger.Info(fmt.Sprintf("processing user %d/%d", counter, len(uniqueUserEmails)), "email", email)

		h.logger.Info("fetching user details", "email", email)
		userDetails, err := fetchUserDetails(iamClient, email)
		if err != nil {
			h.logger.Error("failed to fetch user details", "email", email, "error", err)
			continue
		}

		h.logger.Info("checking criteria against user", "email", email, "criteria", cfg.UserCriteria.String())
		if criteriaSatisfied, err := evaluateCriteria(cfg.UserCriteria, userDetails); err != nil {
			h.logger.Error("failed to check criteria", "email", email, "error", err)
		} else if !criteriaSatisfied {
			h.logger.Info("criteria not satisfied", "email", email)
			continue
		}

		h.logger.Info("evaluating new owner", "email", email, "expression", cfg.ReassignOwnershipTo.String())
		newOwner, err := h.evaluateNewOwner(cfg.ReassignOwnershipTo, userDetails)
		if err != nil {
			h.logger.Error("evaluating new owner", "email", email, "error", err)
			continue
		}
		h.logger.Info(fmt.Sprintf("evaluated new owner: %q", newOwner), "email", email)

		if !cfg.DryRun {
			// revoking grants with account_id == email
			h.logger.Info("revoking user active grants", "email", email)
			if revokedGrants, err := h.revokeUserGrants(ctx, email); err != nil {
				h.logger.Error("failed to reovke grants", "email", email, "error", err)
			} else {
				revokedGrantIDs := []string{}
				for _, g := range revokedGrants {
					revokedGrantIDs = append(revokedGrantIDs, g.ID)
				}
				h.logger.Info("grant revocation successful", "count", len(revokedGrantIDs), "grant_ids", revokedGrantIDs)
			}

			// reassigning grants owned by the user to the new owner
			successfulGrants, failedGrants := h.reassignGrantsOwnership(ctx, grantsOwnedByUser[email], newOwner)
			if len(successfulGrants) > 0 {
				successfulGrantIDs := []string{}
				for _, g := range successfulGrants {
					successfulGrantIDs = append(successfulGrantIDs, g.ID)
				}
				h.logger.Info("grant ownership reassignment successful", "count", len(successfulGrantIDs), "grant_ids", successfulGrantIDs)
			}
			if len(failedGrants) > 0 {
				failedGrantIDs := []string{}
				for _, g := range failedGrants {
					failedGrantIDs = append(failedGrantIDs, g.ID)
				}
				h.logger.Error("grant ownership reassignment failed", "count", len(failedGrantIDs), "grant_ids", failedGrantIDs)
			}
		}
	}

	return nil
}

func fetchUserDetails(iamClient domain.IAMClient, email string) (map[string]interface{}, error) {
	user, err := iamClient.GetUser(email)
	if err != nil {
		return nil, err
	}
	userMap, ok := user.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("parsing user details: expected a map[string]interface{}, got %T instead; value is %q", user, user)
	}
	return userMap, nil
}

func evaluateCriteria(criteriaExpr evaluator.Expression, user map[string]interface{}) (bool, error) {
	criteriaEvaluation, err := criteriaExpr.EvaluateWithVars(map[string]interface{}{
		"user": user,
	})
	if err != nil {
		return false, fmt.Errorf("evaluating criteria: %w", err)
	}
	satisfied, ok := criteriaEvaluation.(bool)
	if !ok {
		return false, fmt.Errorf("invalid type for user_criteria evaluation result: expected boolean, got %T; value is %q", criteriaEvaluation, criteriaEvaluation)
	}

	return satisfied, nil
}

func (h *handler) revokeUserGrants(ctx context.Context, email string) ([]*domain.Grant, error) {
	revokeGrantsFilter := domain.RevokeGrantsFilter{
		AccountIDs: []string{email},
	}
	h.logger.Info("revoking grants", "account_id", email)
	revokedGrants, err := h.grantService.BulkRevoke(ctx, revokeGrantsFilter, domain.SystemActorName, "Revoked due to user deactivated")
	if err != nil {
		return nil, fmt.Errorf("revoking grants for %q: %w", email, err)
	}

	return revokedGrants, nil
}

func (h *handler) evaluateNewOwner(newOwnerExpr evaluator.Expression, user map[string]interface{}) (string, error) {
	newOwner, err := newOwnerExpr.EvaluateWithVars(map[string]interface{}{
		"user": user,
	})
	if err != nil {
		return "", fmt.Errorf("evaluating reassign_ownership_to: %w", err)
	}

	newOwnerStr, ok := newOwner.(string)
	// owner validation
	if !ok {
		return "", fmt.Errorf("invalid type for reassign_ownership_to evaluation result: expected string, got %T instead; value is %q", newOwner, newOwner)
	} else if newOwnerStr == "" {
		return "", fmt.Errorf("invalid value for reassign_ownership_to evaluation result: expected a non-empty string, got %q instead", newOwnerStr)
	} else if err := h.validator.Var(newOwnerStr, "email"); err != nil {
		return "", fmt.Errorf("invalid value for reassign_ownership_to evaluation result: expected a valid email address, got %q", newOwnerStr)
	}

	return newOwnerStr, nil
}

func (h *handler) reassignGrantsOwnership(ctx context.Context, ownedGrants []*domain.Grant, newOwner string) ([]*domain.Grant, []*domain.Grant) {
	var successfulGrants, failedGrants []*domain.Grant
	for _, g := range ownedGrants {
		g.Owner = newOwner
		if err := h.grantService.Update(ctx, g); err != nil {
			failedGrants = append(failedGrants, g)
			h.logger.Error("updating grant owner", "grant_id", g.ID, "existing_owner", g.Owner, "new_owner", newOwner, "error", err)
			continue
		}
		successfulGrants = append(successfulGrants, g)
	}

	return successfulGrants, failedGrants
}

func getGrantIDs(grants []domain.Grant) []string {
	var ids []string
	for _, g := range grants {
		ids = append(ids, g.ID)
	}
	return ids
}
