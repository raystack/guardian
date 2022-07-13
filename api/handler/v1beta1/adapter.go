package v1beta1

import (
	"time"

	"github.com/mitchellh/mapstructure"
	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/domain"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type adapter struct{}

func NewAdapter() *adapter {
	return &adapter{}
}

func (a *adapter) FromProviderProto(p *guardianv1beta1.Provider) (*domain.Provider, error) {
	providerConfig, err := a.FromProviderConfigProto(p.GetConfig())
	if err != nil {
		return nil, err
	}

	return &domain.Provider{
		ID:        p.GetId(),
		Type:      p.GetType(),
		URN:       p.GetUrn(),
		Config:    providerConfig,
		CreatedAt: p.GetCreatedAt().AsTime(),
		UpdatedAt: p.GetUpdatedAt().AsTime(),
	}, nil
}

func (a *adapter) FromProviderConfigProto(pc *guardianv1beta1.ProviderConfig) (*domain.ProviderConfig, error) {
	var appealConfig *domain.AppealConfig
	if pc.GetAppeal() != nil {
		appealConfig = &domain.AppealConfig{}
		appealConfig.AllowPermanentAccess = pc.GetAppeal().GetAllowPermanentAccess()
		appealConfig.AllowActiveAccessExtensionIn = pc.GetAppeal().GetAllowActiveAccessExtensionIn()
	}

	resources := []*domain.ResourceConfig{}
	for _, r := range pc.GetResources() {
		roles := []*domain.Role{}
		for _, role := range r.GetRoles() {
			permissions := []interface{}{}
			for _, p := range role.GetPermissions() {
				permissions = append(permissions, p.AsInterface())
			}

			roles = append(roles, &domain.Role{
				ID:          role.GetId(),
				Name:        role.GetName(),
				Description: role.GetDescription(),
				Permissions: permissions,
			})
		}

		resources = append(resources, &domain.ResourceConfig{
			Type:   r.GetType(),
			Policy: a.fromPolicyConfigProto(r.GetPolicy()),
			Roles:  roles,
		})
	}

	var allowedAccountTypes []string
	if pc.GetAllowedAccountTypes() != nil {
		allowedAccountTypes = pc.GetAllowedAccountTypes()
	}

	return &domain.ProviderConfig{
		Type:                pc.GetType(),
		URN:                 pc.GetUrn(),
		Labels:              pc.GetLabels(),
		Credentials:         pc.GetCredentials().AsInterface(),
		Appeal:              appealConfig,
		Resources:           resources,
		AllowedAccountTypes: allowedAccountTypes,
	}, nil
}

func (a *adapter) ToProviderProto(p *domain.Provider) (*guardianv1beta1.Provider, error) {
	config, err := a.ToProviderConfigProto(p.Config)
	if err != nil {
		return nil, err
	}

	return &guardianv1beta1.Provider{
		Id:        p.ID,
		Type:      p.Type,
		Urn:       p.URN,
		Config:    config,
		CreatedAt: timestamppb.New(p.CreatedAt),
		UpdatedAt: timestamppb.New(p.UpdatedAt),
	}, nil
}

func (a *adapter) ToProviderConfigProto(pc *domain.ProviderConfig) (*guardianv1beta1.ProviderConfig, error) {
	credentials, err := structpb.NewValue(pc.Credentials)
	if err != nil {
		return nil, err
	}

	var appeal *guardianv1beta1.ProviderConfig_AppealConfig
	if pc.Appeal != nil {
		appeal = &guardianv1beta1.ProviderConfig_AppealConfig{
			AllowPermanentAccess:         pc.Appeal.AllowPermanentAccess,
			AllowActiveAccessExtensionIn: pc.Appeal.AllowActiveAccessExtensionIn,
		}
	}

	resources := []*guardianv1beta1.ProviderConfig_ResourceConfig{}
	for _, rc := range pc.Resources {
		roles := []*guardianv1beta1.Role{}
		for _, role := range rc.Roles {
			roleProto, err := a.ToRole(role)
			if err != nil {
				return nil, err
			}
			roles = append(roles, roleProto)
		}

		resources = append(resources, &guardianv1beta1.ProviderConfig_ResourceConfig{
			Type:   rc.Type,
			Policy: a.toPolicyConfigProto(rc.Policy),
			Roles:  roles,
		})
	}

	var allowedAccountTypes []string
	if pc.AllowedAccountTypes != nil {
		allowedAccountTypes = pc.AllowedAccountTypes
	}

	return &guardianv1beta1.ProviderConfig{
		Type:                pc.Type,
		Urn:                 pc.URN,
		Labels:              pc.Labels,
		Credentials:         credentials,
		Appeal:              appeal,
		Resources:           resources,
		AllowedAccountTypes: allowedAccountTypes,
	}, nil
}

func (a *adapter) ToProviderTypeProto(pt domain.ProviderType) (*guardianv1beta1.ProviderType, error) {
	return &guardianv1beta1.ProviderType{
		Name:          pt.Name,
		ResourceTypes: pt.ResourceTypes,
	}, nil
}

func (a *adapter) ToRole(role *domain.Role) (*guardianv1beta1.Role, error) {
	permissions := []*structpb.Value{}
	for _, p := range role.Permissions {
		permission, err := structpb.NewValue(p)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	return &guardianv1beta1.Role{
		Id:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Permissions: permissions,
	}, nil
}

func (a *adapter) FromPolicyProto(p *guardianv1beta1.Policy) (*domain.Policy, error) {
	var steps []*domain.Step
	if p.GetSteps() != nil {
		for _, s := range p.GetSteps() {
			steps = append(steps, &domain.Step{
				Name:            s.GetName(),
				Description:     s.GetDescription(),
				When:            s.GetWhen(),
				Strategy:        domain.ApprovalStepStrategy(s.GetStrategy()),
				RejectionReason: s.GetRejectionReason(),
				ApproveIf:       s.GetApproveIf(),
				AllowFailed:     s.GetAllowFailed(),
				Approvers:       s.GetApprovers(),
			})
		}
	}

	var requirements []*domain.Requirement
	if p.GetRequirements() != nil {
		for _, r := range p.GetRequirements() {
			var on *domain.RequirementTrigger
			if r.GetOn() != nil {
				var conditions []*domain.Condition
				if r.GetOn().GetConditions() != nil {
					for _, c := range r.GetOn().GetConditions() {
						conditions = append(conditions, a.fromConditionProto(c))
					}
				}

				on = &domain.RequirementTrigger{
					ProviderType: r.GetOn().GetProviderType(),
					ProviderURN:  r.GetOn().GetProviderUrn(),
					ResourceType: r.GetOn().GetResourceType(),
					ResourceURN:  r.GetOn().GetResourceUrn(),
					Role:         r.GetOn().GetRole(),
					Conditions:   conditions,
				}
			}

			var additionalAppeals []*domain.AdditionalAppeal
			if r.GetAppeals() != nil {
				for _, aa := range r.GetAppeals() {
					var resource *domain.ResourceIdentifier
					if aa.GetResource() != nil {
						resource = &domain.ResourceIdentifier{
							ProviderType: aa.GetResource().GetProviderType(),
							ProviderURN:  aa.GetResource().GetProviderUrn(),
							Type:         aa.GetResource().GetType(),
							URN:          aa.GetResource().GetUrn(),
							ID:           aa.GetResource().GetId(),
						}
					}

					additionalAppeals = append(additionalAppeals, &domain.AdditionalAppeal{
						Resource: resource,
						Role:     aa.GetRole(),
						Options:  a.fromAppealOptionsProto(aa.GetOptions()),
						Policy:   a.fromPolicyConfigProto(aa.GetPolicy()),
					})
				}
			}

			requirements = append(requirements, &domain.Requirement{
				On:      on,
				Appeals: additionalAppeals,
			})
		}
	}

	var iam *domain.IAMConfig
	if p.GetIam() != nil {
		iam = &domain.IAMConfig{
			Provider:      domain.IAMProviderType(p.GetIam().GetProvider()),
			Config:        p.GetIam().GetConfig().AsInterface(),
			Schema:        p.GetIam().GetSchema(),
			AccountStatus: p.GetIam().GetAccountStatus(),
		}
	}

	return &domain.Policy{
		ID:           p.GetId(),
		Version:      uint(p.GetVersion()),
		Description:  p.GetDescription(),
		Steps:        steps,
		Requirements: requirements,
		Labels:       p.GetLabels(),
		IAM:          iam,
		CreatedAt:    p.GetCreatedAt().AsTime(),
		UpdatedAt:    p.GetUpdatedAt().AsTime(),
	}, nil
}

func (a *adapter) ToPolicyProto(p *domain.Policy) (*guardianv1beta1.Policy, error) {
	var steps []*guardianv1beta1.Policy_ApprovalStep
	if p.Steps != nil {
		for _, s := range p.Steps {
			steps = append(steps, &guardianv1beta1.Policy_ApprovalStep{
				Name:            s.Name,
				Description:     s.Description,
				When:            s.When,
				Strategy:        string(s.Strategy),
				RejectionReason: s.RejectionReason,
				ApproveIf:       s.ApproveIf,
				AllowFailed:     s.AllowFailed,
				Approvers:       s.Approvers,
			})
		}
	}

	var requirements []*guardianv1beta1.Policy_Requirement
	if p.Requirements != nil {
		for _, r := range p.Requirements {
			var on *guardianv1beta1.Policy_Requirement_RequirementTrigger
			if r.On != nil {
				var conditions []*guardianv1beta1.Condition
				if r.On.Conditions != nil {
					for _, c := range r.On.Conditions {
						condition, err := a.toConditionProto(c)
						if err != nil {
							return nil, err
						}
						conditions = append(conditions, condition)
					}
				}

				on = &guardianv1beta1.Policy_Requirement_RequirementTrigger{
					ProviderType: r.On.ProviderType,
					ProviderUrn:  r.On.ProviderURN,
					ResourceType: r.On.ResourceType,
					ResourceUrn:  r.On.ResourceURN,
					Role:         r.On.Role,
					Conditions:   conditions,
				}
			}

			var additionalAppeals []*guardianv1beta1.Policy_Requirement_AdditionalAppeal
			if r.Appeals != nil {
				for _, aa := range r.Appeals {
					var resource *guardianv1beta1.Policy_Requirement_AdditionalAppeal_ResourceIdentifier
					if aa.Resource != nil {
						resource = &guardianv1beta1.Policy_Requirement_AdditionalAppeal_ResourceIdentifier{
							ProviderType: aa.Resource.ProviderType,
							ProviderUrn:  aa.Resource.ProviderURN,
							Type:         aa.Resource.Type,
							Urn:          aa.Resource.URN,
							Id:           aa.Resource.ID,
						}
					}

					additionalAppeals = append(additionalAppeals, &guardianv1beta1.Policy_Requirement_AdditionalAppeal{
						Resource: resource,
						Role:     aa.Role,
						Options:  a.toAppealOptionsProto(aa.Options),
						Policy:   a.toPolicyConfigProto(aa.Policy),
					})
				}
			}

			requirements = append(requirements, &guardianv1beta1.Policy_Requirement{
				On:      on,
				Appeals: additionalAppeals,
			})
		}
	}

	var iam *guardianv1beta1.Policy_IAM
	if p.HasIAMConfig() {
		config, err := structpb.NewValue(p.IAM.Config)
		if err != nil {
			return nil, err
		}

		iam = &guardianv1beta1.Policy_IAM{
			Provider:      string(p.IAM.Provider),
			Config:        config,
			Schema:        p.IAM.Schema,
			AccountStatus: p.IAM.AccountStatus,
		}
	}

	return &guardianv1beta1.Policy{
		Id:           p.ID,
		Version:      uint32(p.Version),
		Description:  p.Description,
		Steps:        steps,
		Requirements: requirements,
		Labels:       p.Labels,
		Iam:          iam,
		CreatedAt:    timestamppb.New(p.CreatedAt),
		UpdatedAt:    timestamppb.New(p.UpdatedAt),
	}, nil
}

func (a *adapter) FromResourceProto(r *guardianv1beta1.Resource) *domain.Resource {
	details := map[string]interface{}{}
	if r.GetDetails() != nil {
		details = r.GetDetails().AsMap()
	}
	return &domain.Resource{
		ID:           r.GetId(),
		ProviderType: r.GetProviderType(),
		ProviderURN:  r.GetProviderUrn(),
		Type:         r.GetType(),
		URN:          r.GetUrn(),
		Name:         r.GetName(),
		Details:      details,
		Labels:       r.GetLabels(),
		CreatedAt:    r.GetCreatedAt().AsTime(),
		UpdatedAt:    r.GetUpdatedAt().AsTime(),
		IsDeleted:    r.GetIsDeleted(),
	}
}

func (a *adapter) ToResourceProto(r *domain.Resource) (*guardianv1beta1.Resource, error) {
	var detailsProto *structpb.Struct
	if r.Details != nil {
		details, err := structpb.NewStruct(r.Details)
		if err != nil {
			return nil, err
		}
		detailsProto = details
	}

	return &guardianv1beta1.Resource{
		Id:           r.ID,
		ProviderType: r.ProviderType,
		ProviderUrn:  r.ProviderURN,
		Type:         r.Type,
		Urn:          r.URN,
		Name:         r.Name,
		Details:      detailsProto,
		Labels:       r.Labels,
		CreatedAt:    timestamppb.New(r.CreatedAt),
		UpdatedAt:    timestamppb.New(r.UpdatedAt),
		IsDeleted:    r.IsDeleted,
	}, nil
}

func (a *adapter) FromAppealProto(appeal *guardianv1beta1.Appeal) (*domain.Appeal, error) {
	resource := a.FromResourceProto(appeal.GetResource())

	approvals := []*domain.Approval{}
	for _, a := range appeal.GetApprovals() {
		var actor *string
		if a.GetActor() != "" {
			actorStr := a.GetActor()
			actor = &actorStr
		}

		approvals = append(approvals, &domain.Approval{
			ID:            a.GetId(),
			Name:          a.GetName(),
			AppealID:      a.GetId(),
			Status:        a.GetStatus(),
			Actor:         actor,
			Reason:        a.GetReason(),
			PolicyID:      a.GetPolicyId(),
			PolicyVersion: uint(a.GetPolicyVersion()),
			Approvers:     a.GetApprovers(),
			CreatedAt:     appeal.GetCreatedAt().AsTime(),
			UpdatedAt:     appeal.GetUpdatedAt().AsTime(),
		})
	}

	details := map[string]interface{}{}
	if appeal.GetDetails() != nil {
		details = appeal.GetDetails().AsMap()
	}

	return &domain.Appeal{
		ID:            appeal.GetId(),
		ResourceID:    appeal.GetResourceId(),
		PolicyID:      appeal.GetPolicyId(),
		PolicyVersion: uint(appeal.GetPolicyVersion()),
		Status:        appeal.GetStatus(),
		AccountID:     appeal.GetAccountId(),
		AccountType:   appeal.GetAccountType(),
		CreatedBy:     appeal.GetCreatedBy(),
		Creator:       appeal.GetCreator().AsInterface(),
		Role:          appeal.GetRole(),
		Options:       a.fromAppealOptionsProto(appeal.GetOptions()),
		Labels:        appeal.GetLabels(),
		RevokedBy:     appeal.GetRevokedBy(),
		RevokedAt:     appeal.GetRevokedAt().AsTime(),
		RevokeReason:  appeal.GetRevokeReason(),
		Resource:      resource,
		Approvals:     approvals,
		CreatedAt:     appeal.GetCreatedAt().AsTime(),
		UpdatedAt:     appeal.GetUpdatedAt().AsTime(),
		Details:       details,
	}, nil
}

func (a *adapter) ToAppealProto(appeal *domain.Appeal) (*guardianv1beta1.Appeal, error) {
	var resource *guardianv1beta1.Resource
	if appeal.Resource != nil {
		r, err := a.ToResourceProto(appeal.Resource)
		if err != nil {
			return nil, err
		}
		resource = r
	}

	creator, err := structpb.NewValue(appeal.Creator)
	if err != nil {
		return nil, err
	}

	approvals := []*guardianv1beta1.Approval{}
	for _, approval := range appeal.Approvals {
		approvalProto, err := a.ToApprovalProto(approval)
		if err != nil {
			return nil, err
		}

		approvals = append(approvals, approvalProto)
	}

	var detailsProto *structpb.Struct
	if appeal.Details != nil {
		details, err := structpb.NewStruct(appeal.Details)
		if err != nil {
			return nil, err
		}
		detailsProto = details
	}

	return &guardianv1beta1.Appeal{
		Id:            appeal.ID,
		ResourceId:    appeal.ResourceID,
		PolicyId:      appeal.PolicyID,
		PolicyVersion: uint32(appeal.PolicyVersion),
		Status:        appeal.Status,
		AccountId:     appeal.AccountID,
		AccountType:   appeal.AccountType,
		CreatedBy:     appeal.CreatedBy,
		Creator:       creator,
		Role:          appeal.Role,
		Options:       a.toAppealOptionsProto(appeal.Options),
		Labels:        appeal.Labels,
		RevokedBy:     appeal.RevokedBy,
		RevokedAt:     timestamppb.New(appeal.RevokedAt),
		RevokeReason:  appeal.RevokeReason,
		Resource:      resource,
		Approvals:     approvals,
		CreatedAt:     timestamppb.New(appeal.CreatedAt),
		UpdatedAt:     timestamppb.New(appeal.UpdatedAt),
		Details:       detailsProto,
	}, nil
}

func (a *adapter) FromCreateAppealProto(ca *guardianv1beta1.CreateAppealRequest, authenticatedUser string) ([]*domain.Appeal, error) {
	var appeals []*domain.Appeal

	for _, r := range ca.GetResources() {
		var options *domain.AppealOptions
		if r.GetOptions() != nil {
			if err := mapstructure.Decode(r.GetOptions().AsMap(), &options); err != nil {
				return nil, err
			}
		}

		appeals = append(appeals, &domain.Appeal{
			AccountID:   ca.GetAccountId(),
			AccountType: ca.GetAccountType(),
			CreatedBy:   authenticatedUser,
			ResourceID:  r.GetId(),
			Role:        r.GetRole(),
			Options:     options,
			Details:     r.GetDetails().AsMap(),
		})
	}

	return appeals, nil
}

func (a *adapter) ToApprovalProto(approval *domain.Approval) (*guardianv1beta1.Approval, error) {
	var appealProto *guardianv1beta1.Appeal
	if approval.Appeal != nil {
		appeal, err := a.ToAppealProto(approval.Appeal)
		if err != nil {
			return nil, err
		}
		appealProto = appeal
	}

	var actor string
	if approval.Actor != nil {
		actor = *approval.Actor
	}

	return &guardianv1beta1.Approval{
		Id:            approval.ID,
		Name:          approval.Name,
		AppealId:      approval.AppealID,
		Appeal:        appealProto,
		Status:        approval.Status,
		Actor:         actor,
		Reason:        approval.Reason,
		PolicyId:      approval.PolicyID,
		PolicyVersion: uint32(approval.PolicyVersion),
		Approvers:     approval.Approvers,
		CreatedAt:     timestamppb.New(approval.CreatedAt),
		UpdatedAt:     timestamppb.New(approval.UpdatedAt),
	}, nil
}

func (a *adapter) fromConditionProto(c *guardianv1beta1.Condition) *domain.Condition {
	if c == nil {
		return nil
	}

	var match *domain.MatchCondition
	if c.GetMatch() != nil {
		match = &domain.MatchCondition{
			Eq: c.GetMatch().GetEq(),
		}
	}

	return &domain.Condition{
		Field: c.GetField(),
		Match: match,
	}
}

func (a *adapter) toConditionProto(c *domain.Condition) (*guardianv1beta1.Condition, error) {
	if c == nil {
		return nil, nil
	}

	var match *guardianv1beta1.Condition_MatchCondition
	if c.Match != nil {
		eq, err := structpb.NewValue(c.Match.Eq)
		if err != nil {
			return nil, err
		}

		match = &guardianv1beta1.Condition_MatchCondition{
			Eq: eq,
		}
	}

	return &guardianv1beta1.Condition{
		Field: c.Field,
		Match: match,
	}, nil
}

func (a *adapter) fromAppealOptionsProto(o *guardianv1beta1.AppealOptions) *domain.AppealOptions {
	if o == nil {
		return nil
	}

	var expirationDate time.Time
	if o.GetExpirationDate() != nil {
		expirationDate = o.GetExpirationDate().AsTime()
	}

	return &domain.AppealOptions{
		Duration:       o.GetDuration(),
		ExpirationDate: &expirationDate,
	}
}

func (a *adapter) toAppealOptionsProto(o *domain.AppealOptions) *guardianv1beta1.AppealOptions {
	if o == nil {
		return nil
	}

	var expirationDate *timestamppb.Timestamp
	if o.ExpirationDate != nil {
		expirationDate = timestamppb.New(*o.ExpirationDate)
	}

	return &guardianv1beta1.AppealOptions{
		Duration:       o.Duration,
		ExpirationDate: expirationDate,
	}
}

func (a *adapter) fromPolicyConfigProto(c *guardianv1beta1.PolicyConfig) *domain.PolicyConfig {
	if c == nil {
		return nil
	}

	return &domain.PolicyConfig{
		ID:      c.GetId(),
		Version: int(c.GetVersion()),
	}
}

func (a *adapter) toPolicyConfigProto(c *domain.PolicyConfig) *guardianv1beta1.PolicyConfig {
	if c == nil {
		return nil
	}

	return &guardianv1beta1.PolicyConfig{
		Id:      c.ID,
		Version: int32(c.Version),
	}
}
