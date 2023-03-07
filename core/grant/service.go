package grant

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/plugins/notifiers"
	"github.com/odpf/guardian/utils"
	"github.com/odpf/salt/log"
)

const (
	AuditKeyRevoke = "grant.revoke"
	AuditKeyUpdate = "grant.update"
)

//go:generate mockery --name=repository --exported --with-expecter
type repository interface {
	List(context.Context, domain.ListGrantsFilter) ([]domain.Grant, error)
	GetByID(context.Context, string) (*domain.Grant, error)
	Update(context.Context, *domain.Grant) error
	BulkUpsert(context.Context, []*domain.Grant) error
}

//go:generate mockery --name=providerService --exported --with-expecter
type providerService interface {
	GetByID(context.Context, string) (*domain.Provider, error)
	RevokeAccess(context.Context, domain.Grant) error
	ListAccess(context.Context, domain.Provider, []*domain.Resource) (domain.MapResourceAccess, error)
}

//go:generate mockery --name=resourceService --exported --with-expecter
type resourceService interface {
	Find(context.Context, domain.ListResourcesFilter) ([]*domain.Resource, error)
}

//go:generate mockery --name=auditLogger --exported --with-expecter
type auditLogger interface {
	Log(ctx context.Context, action string, data interface{}) error
}

//go:generate mockery --name=notifier --exported --with-expecter
type notifier interface {
	notifiers.Client
}

type grantCreation struct {
	AppealStatus string `validate:"required,eq=approved"`
	AccountID    string `validate:"required"`
	AccountType  string `validate:"required"`
	ResourceID   string `validate:"required"`
}

type Service struct {
	repo            repository
	providerService providerService
	resourceService resourceService

	notifier    notifier
	validator   *validator.Validate
	logger      log.Logger
	auditLogger auditLogger
}

type ServiceDeps struct {
	Repository      repository
	ProviderService providerService
	ResourceService resourceService

	Notifier    notifier
	Validator   *validator.Validate
	Logger      log.Logger
	AuditLogger auditLogger
}

func NewService(deps ServiceDeps) *Service {
	return &Service{
		repo:            deps.Repository,
		providerService: deps.ProviderService,
		resourceService: deps.ResourceService,

		notifier:    deps.Notifier,
		validator:   deps.Validator,
		logger:      deps.Logger,
		auditLogger: deps.AuditLogger,
	}
}

func (s *Service) List(ctx context.Context, filter domain.ListGrantsFilter) ([]domain.Grant, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) GetByID(ctx context.Context, id string) (*domain.Grant, error) {
	if id == "" {
		return nil, ErrEmptyIDParam
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, payload *domain.Grant) error {
	grantDetails, err := s.GetByID(ctx, payload.ID)
	if err != nil {
		return fmt.Errorf("getting grant details: %w", err)
	}

	if payload.Owner == "" {
		return ErrEmptyOwner
	}
	updatedGrant := &domain.Grant{
		ID: payload.ID,

		// Only allow updating several fields
		Owner: payload.Owner,
	}
	if err := s.repo.Update(ctx, updatedGrant); err != nil {
		return err
	}
	previousOwner := grantDetails.Owner
	grantDetails.Owner = updatedGrant.Owner
	grantDetails.UpdatedAt = updatedGrant.UpdatedAt
	*payload = *grantDetails

	if err := s.auditLogger.Log(ctx, AuditKeyUpdate, map[string]interface{}{
		"grant_id":      grantDetails.ID,
		"payload":       updatedGrant,
		"updated_grant": payload,
	}); err != nil {
		s.logger.Error("failed to record audit log", "error", err)
	}

	if previousOwner != updatedGrant.Owner {
		message := domain.NotificationMessage{
			Type: domain.NotificationTypeGrantOwnerChanged,
			Variables: map[string]interface{}{
				"grant_id":       grantDetails.ID,
				"previous_owner": previousOwner,
				"new_owner":      updatedGrant.Owner,
			},
		}
		notifications := []domain.Notification{{
			User:    updatedGrant.Owner,
			Message: message,
		}}
		if previousOwner != "" {
			notifications = append(notifications, domain.Notification{
				User:    previousOwner,
				Message: message,
			})
		}
		if errs := s.notifier.Notify(notifications); errs != nil {
			for _, err1 := range errs {
				s.logger.Error("failed to send notifications", "error", err1.Error())
			}
		}
	}

	return nil
}

func (s *Service) Prepare(ctx context.Context, appeal domain.Appeal) (*domain.Grant, error) {
	// validation
	if err := s.validator.Struct(grantCreation{
		AppealStatus: appeal.Status,
		AccountID:    appeal.AccountID,
		AccountType:  appeal.AccountType,
		ResourceID:   appeal.ResourceID,
	}); err != nil {
		return nil, fmt.Errorf("validating appeal: %w", err)
	}

	// converting aapeal into a new grant
	return appeal.ToGrant()
}

func (s *Service) Revoke(ctx context.Context, id, actor, reason string, opts ...Option) (*domain.Grant, error) {
	grant, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting grant details: %w", err)
	}

	revokedGrant := &domain.Grant{}
	*revokedGrant = *grant
	if err := grant.Revoke(actor, reason); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, grant); err != nil {
		return nil, fmt.Errorf("updating grant record in db: %w", err)
	}

	options := s.getOptions(opts...)

	if !options.skipRevokeInProvider {
		if err := s.providerService.RevokeAccess(ctx, *grant); err != nil {
			if err := s.repo.Update(ctx, grant); err != nil {
				return nil, fmt.Errorf("failed to rollback grant status: %w", err)
			}
			return nil, fmt.Errorf("removing grant in provider: %w", err)
		}
	}

	if !options.skipNotification {
		if errs := s.notifier.Notify([]domain.Notification{{
			User: grant.CreatedBy,
			Message: domain.NotificationMessage{
				Type: domain.NotificationTypeAccessRevoked,
				Variables: map[string]interface{}{
					"resource_name": fmt.Sprintf("%s (%s: %s)", grant.Resource.Name, grant.Resource.ProviderType, grant.Resource.URN),
					"role":          grant.Role,
					"account_type":  grant.AccountType,
					"account_id":    grant.AccountID,
					"requestor":     grant.Owner,
				},
			},
		}}); errs != nil {
			for _, err1 := range errs {
				s.logger.Error("failed to send notifications", "error", err1.Error())
			}
		}
	}

	if err := s.auditLogger.Log(ctx, AuditKeyRevoke, map[string]interface{}{
		"grant_id": id,
		"reason":   reason,
	}); err != nil {
		s.logger.Error("failed to record audit log", "error", err)
	}

	return grant, nil
}

func (s *Service) BulkRevoke(ctx context.Context, filter domain.RevokeGrantsFilter, actor, reason string) ([]*domain.Grant, error) {
	if filter.AccountIDs == nil || len(filter.AccountIDs) == 0 {
		return nil, fmt.Errorf("account_ids is required")
	}

	grants, err := s.List(ctx, domain.ListGrantsFilter{
		Statuses:      []string{string(domain.GrantStatusActive)},
		AccountIDs:    filter.AccountIDs,
		ProviderTypes: filter.ProviderTypes,
		ProviderURNs:  filter.ProviderURNs,
		ResourceTypes: filter.ResourceTypes,
		ResourceURNs:  filter.ResourceURNs,
	})
	if err != nil {
		return nil, fmt.Errorf("listing active grants: %w", err)
	}
	if len(grants) == 0 {
		return nil, nil
	}

	result := make([]*domain.Grant, 0)
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

	totalRequests := len(grants)
	done := make(chan *domain.Grant, totalRequests)
	resourceGrantMap := make(map[string][]*domain.Grant, 0)

	for i, grant := range grants {
		var resourceGrants []*domain.Grant
		var ok bool
		if resourceGrants, ok = resourceGrantMap[grant.ResourceID]; ok {
			resourceGrants = append(resourceGrants, &grants[i])
		} else {
			resourceGrants = []*domain.Grant{&grants[i]}
		}
		resourceGrantMap[grant.ResourceID] = resourceGrants
	}

	for _, resourceGrants := range resourceGrantMap {
		go s.expiredInActiveUserAccess(ctx, timeLimiter, done, actor, reason, resourceGrants)
	}

	var successRevoke []string
	var failedRevoke []string
	for {
		select {
		case grant := <-done:
			if grant.Status == domain.GrantStatusInactive {
				successRevoke = append(successRevoke, grant.ID)
			} else {
				failedRevoke = append(failedRevoke, grant.ID)
			}
			result = append(result, grant)
			if len(result) == totalRequests {
				s.logger.Info("successful grant revocation", "count", len(successRevoke), "ids", successRevoke)
				s.logger.Info("failed grant revocation", "count", len(failedRevoke), "ids", failedRevoke)
				return result, nil
			}
		}
	}
}

func (s *Service) expiredInActiveUserAccess(ctx context.Context, timeLimiter chan int, done chan *domain.Grant, actor string, reason string, grants []*domain.Grant) {
	for _, grant := range grants {
		<-timeLimiter

		revokedGrant := &domain.Grant{}
		*revokedGrant = *grant
		if err := revokedGrant.Revoke(actor, reason); err != nil {
			s.logger.Error("failed to revoke grant", "id", grant.ID, "error", err)
			return
		}
		if err := s.providerService.RevokeAccess(ctx, *grant); err != nil {
			done <- grant
			s.logger.Error("failed to revoke grant in provider", "id", grant.ID, "error", err)
			return
		}

		revokedGrant.Status = domain.GrantStatusInactive
		if err := s.repo.Update(ctx, revokedGrant); err != nil {
			done <- grant
			s.logger.Error("failed to update access-revoke status", "id", grant.ID, "error", err)
			return
		} else {
			done <- revokedGrant
			s.logger.Info("grant revoked", "id", grant.ID)
		}
	}
}

type ImportFromProviderCriteria struct {
	ProviderID    string `validate:"required"`
	ResourceIDs   []string
	ResourceTypes []string
	ResourceURNs  []string
}

func (s *Service) ImportFromProvider(ctx context.Context, criteria ImportFromProviderCriteria) ([]*domain.Grant, error) {
	p, err := s.providerService.GetByID(ctx, criteria.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("getting provider details: %w", err)
	}

	listResourcesFilter := domain.ListResourcesFilter{
		ProviderType: p.Type,
		ProviderURN:  p.URN,
	}
	listGrantsFilter := domain.ListGrantsFilter{
		Statuses:      []string{string(domain.GrantStatusActive)},
		ProviderTypes: []string{p.Type},
		ProviderURNs:  []string{p.URN},
	}
	if criteria.ResourceIDs != nil {
		listResourcesFilter.IDs = criteria.ResourceIDs
		listGrantsFilter.ResourceIDs = criteria.ResourceIDs
	} else {
		listResourcesFilter.ResourceTypes = criteria.ResourceTypes
		listResourcesFilter.ResourceURNs = criteria.ResourceURNs

		listGrantsFilter.ResourceTypes = criteria.ResourceTypes
		listGrantsFilter.ResourceURNs = criteria.ResourceURNs
	}
	resources, err := s.resourceService.Find(ctx, listResourcesFilter)
	if err != nil {
		return nil, fmt.Errorf("getting resources: %w", err)
	}

	resourceAccess, err := s.providerService.ListAccess(ctx, *p, resources)
	if err != nil {
		return nil, fmt.Errorf("fetching access from provider: %w", err)
	}

	resourceConfigs := make(map[string]*domain.ResourceConfig)
	for _, rc := range p.Config.Resources {
		resourceConfigs[rc.Type] = rc
	}

	resourcesMap := make(map[string]*domain.Resource)
	for _, r := range resources {
		resourcesMap[r.URN] = r
	}

	activeGrants, err := s.repo.List(ctx, listGrantsFilter)
	if err != nil {
		return nil, fmt.Errorf("getting active grants: %w", err)
	}
	// map[resourceURN]map[accounttype:accountId]map[permissionsKey]grant
	activeGrantsMap := map[string]map[string]map[string]*domain.Grant{}
	for i, g := range activeGrants {
		if activeGrantsMap[g.Resource.URN] == nil {
			activeGrantsMap[g.Resource.URN] = map[string]map[string]*domain.Grant{}
		}

		accountSignature := getAccountSignature(g.AccountType, g.AccountID)
		if activeGrantsMap[g.Resource.URN][accountSignature] == nil {
			activeGrantsMap[g.Resource.URN][accountSignature] = map[string]*domain.Grant{}
		}

		activeGrantsMap[g.Resource.URN][accountSignature][g.PermissionsKey()] = &activeGrants[i]
	}

	var newAndUpdatedGrants []*domain.Grant
	for rURN, accessEntries := range resourceAccess {
		resource, ok := resourcesMap[rURN]
		if !ok {
			continue // skip access for resources that not yet added to guardian
		}

		importedGrants := []*domain.Grant{}
		for accountSignature, accessEntries := range groupAccessEntriesByAccount(accessEntries) {
			// convert access entries to grants
			var grants []*domain.Grant
			for _, ae := range accessEntries {
				g := ae.ToGrant(*resource)
				grants = append(grants, &g)
			}

			// group grants for the same account (accountGrants) by provider role
			rc := resourceConfigs[resource.Type]
			grants = reduceGrantsByProviderRole(*rc, grants)
			for i, g := range grants {
				key := g.PermissionsKey()
				if existingGrant, ok := activeGrantsMap[rURN][accountSignature][key]; ok {
					// replace imported grant values with existing grant
					*grants[i] = *existingGrant

					// remove updated grant from active grants map
					delete(activeGrantsMap[rURN][accountSignature], key)
				}
			}

			importedGrants = append(importedGrants, grants...)
		}

		if len(importedGrants) > 0 {
			if err := s.repo.BulkUpsert(ctx, importedGrants); err != nil {
				return nil, fmt.Errorf("inserting new and updated grants into the db for %q: %w", rURN, err)
			}
			newAndUpdatedGrants = append(newAndUpdatedGrants, importedGrants...)
		}
	}

	// mark remaining active grants as inactive
	var deactivatedGrants []*domain.Grant
	for _, v := range activeGrantsMap {
		for _, v2 := range v {
			for _, g := range v2 {
				g.StatusInProvider = domain.GrantStatusInactive
				deactivatedGrants = append(deactivatedGrants, g)
			}
		}
	}
	if len(deactivatedGrants) > 0 {
		if err := s.repo.BulkUpsert(ctx, deactivatedGrants); err != nil {
			return nil, fmt.Errorf("updating grants provider status: %w", err)
		}
	}

	return newAndUpdatedGrants, nil
}

func getAccountSignature(accountType, accountID string) string {
	return fmt.Sprintf("%s:%s", accountType, accountID)
}

func groupAccessEntriesByAccount(accessEntries []domain.AccessEntry) map[string][]domain.AccessEntry {
	result := map[string][]domain.AccessEntry{}
	for _, ae := range accessEntries {
		accountSignature := getAccountSignature(ae.AccountType, ae.AccountID)
		result[accountSignature] = append(result[accountSignature], ae)
	}
	return result
}

// reduceGrantsByProviderRole reduces grants based on configured roles in the provider's resource config and returns reduced grants containing the Role according to the resource config
func reduceGrantsByProviderRole(rc domain.ResourceConfig, grants []*domain.Grant) (reducedGrants []*domain.Grant) {
	grantsGroupedByPermission := map[string]*domain.Grant{}
	var allGrantPermissions []string
	for _, g := range grants {
		// TODO: validate if permissions is empty
		allGrantPermissions = append(allGrantPermissions, g.Permissions[0])
		grantsGroupedByPermission[g.Permissions[0]] = g
	}
	sort.Strings(allGrantPermissions)

	// prioritize roles with more permissions
	sort.Slice(rc.Roles, func(i, j int) bool {
		return len(rc.Roles[i].Permissions) > len(rc.Roles[j].Permissions)
	})
	for _, role := range rc.Roles {
		rolePermissions := role.GetOrderedPermissions()
		if containing, headIndex := utils.SubsliceExists(allGrantPermissions, rolePermissions); containing {
			sampleGrant := grantsGroupedByPermission[rolePermissions[0]]
			sampleGrant.Role = role.ID
			sampleGrant.Permissions = rolePermissions
			reducedGrants = append(reducedGrants, sampleGrant)

			for _, p := range rolePermissions {
				// delete combined grants
				delete(grantsGroupedByPermission, p)
			}
			allGrantPermissions = append(allGrantPermissions[:headIndex], allGrantPermissions[headIndex+1:]...)
		}
	}

	if len(grantsGroupedByPermission) > 0 {
		// add remaining grants with non-registered provider role
		for _, g := range grantsGroupedByPermission {
			reducedGrants = append(reducedGrants, g)
		}
	}

	return
}
