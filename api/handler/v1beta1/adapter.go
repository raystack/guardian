package v1beta1

import (
	"fmt"

	guardianv1beta1 "github.com/goto/guardian/api/proto/gotocompany/guardian/v1beta1"
	"github.com/goto/guardian/domain"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type adapter struct{}

func NewAdapter() *adapter {
	return &adapter{}
}

func (a *adapter) FromProviderProto(p *guardianv1beta1.Provider) (*domain.Provider, error) {
	provider := &domain.Provider{
		ID:   p.GetId(),
		Type: p.GetType(),
		URN:  p.GetUrn(),
	}

	if p.GetConfig() != nil {
		provider.Config = a.FromProviderConfigProto(p.GetConfig())
	}

	if p.GetCreatedAt() != nil {
		provider.CreatedAt = p.GetCreatedAt().AsTime()
	}
	if p.GetUpdatedAt() != nil {
		provider.UpdatedAt = p.GetUpdatedAt().AsTime()
	}

	return provider, nil
}

func (a *adapter) FromProviderConfigProto(pc *guardianv1beta1.ProviderConfig) *domain.ProviderConfig {
	providerConfig := &domain.ProviderConfig{
		Type:        pc.GetType(),
		URN:         pc.GetUrn(),
		Labels:      pc.GetLabels(),
		Credentials: pc.GetCredentials().AsInterface(),
	}

	if pc.GetAppeal() != nil {
		appealConfig := &domain.AppealConfig{}
		appealConfig.AllowPermanentAccess = pc.GetAppeal().GetAllowPermanentAccess()
		appealConfig.AllowActiveAccessExtensionIn = pc.GetAppeal().GetAllowActiveAccessExtensionIn()
		providerConfig.Appeal = appealConfig
	}

	if pc.GetResources() != nil {
		resources := []*domain.ResourceConfig{}
		for _, r := range pc.GetResources() {
			roles := []*domain.Role{}
			for _, roleProto := range r.GetRoles() {
				role := &domain.Role{
					ID:          roleProto.GetId(),
					Name:        roleProto.GetName(),
					Description: roleProto.GetDescription(),
				}

				if roleProto.Permissions != nil {
					permissions := []interface{}{}
					for _, p := range roleProto.GetPermissions() {
						permissions = append(permissions, p.AsInterface())
					}
					role.Permissions = permissions
				}

				roles = append(roles, role)
			}

			resources = append(resources, &domain.ResourceConfig{
				Type:   r.GetType(),
				Filter: r.GetFilter(),
				Policy: a.fromPolicyConfigProto(r.GetPolicy()),
				Roles:  roles,
			})
		}
		providerConfig.Resources = resources
	}

	if pc.GetParameters() != nil {
		parameters := []*domain.ProviderParameter{}
		for _, p := range pc.GetParameters() {
			parameters = append(parameters, &domain.ProviderParameter{
				Key:         p.GetKey(),
				Label:       p.GetLabel(),
				Required:    p.GetRequired(),
				Description: p.GetDescription(),
			})
		}
		providerConfig.Parameters = parameters
	}

	if pc.GetAllowedAccountTypes() != nil {
		providerConfig.AllowedAccountTypes = pc.GetAllowedAccountTypes()
	}

	return providerConfig
}

func (a *adapter) ToProviderProto(p *domain.Provider) (*guardianv1beta1.Provider, error) {
	providerProto := &guardianv1beta1.Provider{
		Id:   p.ID,
		Type: p.Type,
		Urn:  p.URN,
	}

	if p.Config != nil {
		config, err := a.ToProviderConfigProto(p.Config)
		if err != nil {
			return nil, err
		}
		providerProto.Config = config
	}

	if !p.CreatedAt.IsZero() {
		providerProto.CreatedAt = timestamppb.New(p.CreatedAt)
	}
	if !p.UpdatedAt.IsZero() {
		providerProto.UpdatedAt = timestamppb.New(p.UpdatedAt)
	}

	return providerProto, nil
}

func (a *adapter) ToProviderConfigProto(pc *domain.ProviderConfig) (*guardianv1beta1.ProviderConfig, error) {
	providerConfigProto := &guardianv1beta1.ProviderConfig{
		Type:   pc.Type,
		Urn:    pc.URN,
		Labels: pc.Labels,
	}

	if pc.Credentials != nil {
		credentials, err := structpb.NewValue(pc.Credentials)
		if err != nil {
			return nil, err
		}
		providerConfigProto.Credentials = credentials
	}

	if pc.Appeal != nil {
		providerConfigProto.Appeal = &guardianv1beta1.ProviderConfig_AppealConfig{
			AllowPermanentAccess:         pc.Appeal.AllowPermanentAccess,
			AllowActiveAccessExtensionIn: pc.Appeal.AllowActiveAccessExtensionIn,
		}
	}

	if pc.Resources != nil {
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
		providerConfigProto.Resources = resources
	}

	if pc.Parameters != nil {
		parameters := []*guardianv1beta1.ProviderConfig_ProviderParameter{}
		for _, p := range pc.Parameters {
			parameters = append(parameters, &guardianv1beta1.ProviderConfig_ProviderParameter{
				Key:         p.Key,
				Label:       p.Label,
				Required:    p.Required,
				Description: p.Description,
			})
		}
		providerConfigProto.Parameters = parameters
	}

	if pc.AllowedAccountTypes != nil {
		providerConfigProto.AllowedAccountTypes = pc.AllowedAccountTypes
	}

	return providerConfigProto, nil
}

func (a *adapter) ToProviderTypeProto(pt domain.ProviderType) *guardianv1beta1.ProviderType {
	return &guardianv1beta1.ProviderType{
		Name:          pt.Name,
		ResourceTypes: pt.ResourceTypes,
	}
}

func (a *adapter) ToRole(role *domain.Role) (*guardianv1beta1.Role, error) {
	roleProto := &guardianv1beta1.Role{
		Id:          role.ID,
		Name:        role.Name,
		Description: role.Description,
	}

	if role.Permissions != nil {
		permissions := []*structpb.Value{}
		for _, p := range role.Permissions {
			permission, err := structpb.NewValue(p)
			if err != nil {
				return nil, err
			}
			permissions = append(permissions, permission)
		}
		roleProto.Permissions = permissions
	}

	return roleProto, nil
}

func (a *adapter) FromPolicyProto(p *guardianv1beta1.Policy) *domain.Policy {
	policy := &domain.Policy{
		ID:          p.GetId(),
		Version:     uint(p.GetVersion()),
		Description: p.GetDescription(),
		Labels:      p.GetLabels(),
	}

	if p.GetSteps() != nil {
		var steps []*domain.Step
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
		policy.Steps = steps
	}

	if p.GetRequirements() != nil {
		var requirements []*domain.Requirement
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
					Expression:   r.GetOn().GetExpression(),
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
			policy.Requirements = requirements
		}
	}

	if p.GetIam() != nil {
		policy.IAM = &domain.IAMConfig{
			Provider: domain.IAMProviderType(p.GetIam().GetProvider()),
			Config:   p.GetIam().GetConfig().AsInterface(),
			Schema:   p.GetIam().GetSchema(),
		}
	}

	if p.GetAppeal() != nil {
		var durationOptions []domain.AppealDurationOption
		var questions []domain.Question
		for _, d := range p.GetAppeal().GetDurationOptions() {
			option := domain.AppealDurationOption{
				Name:  d.GetName(),
				Value: d.GetValue(),
			}
			durationOptions = append(durationOptions, option)
		}
		for _, q := range p.GetAppeal().GetQuestions() {
			question := domain.Question{
				Key:         q.GetKey(),
				Question:    q.GetQuestion(),
				Required:    q.GetRequired(),
				Description: q.GetDescription(),
			}
			questions = append(questions, question)
		}

		policy.AppealConfig = &domain.PolicyAppealConfig{
			DurationOptions:              durationOptions,
			AllowOnBehalf:                p.GetAppeal().GetAllowOnBehalf(),
			Questions:                    questions,
			AllowPermanentAccess:         p.GetAppeal().GetAllowPermanentAccess(),
			AllowActiveAccessExtensionIn: p.GetAppeal().GetAllowActiveAccessExtensionIn(),
			AllowCreatorDetailsFailure:   p.GetAppeal().GetAllowCreatorDetailsFailure(),
		}
	}

	if p.GetCreatedAt() != nil {
		policy.CreatedAt = p.GetCreatedAt().AsTime()
	}
	if p.GetUpdatedAt() != nil {
		policy.UpdatedAt = p.GetUpdatedAt().AsTime()
	}

	return policy
}

func (a *adapter) ToPolicyProto(p *domain.Policy) (*guardianv1beta1.Policy, error) {
	policyProto := &guardianv1beta1.Policy{
		Id:          p.ID,
		Version:     uint32(p.Version),
		Description: p.Description,
		Labels:      p.Labels,
	}

	if p.Steps != nil {
		var steps []*guardianv1beta1.Policy_ApprovalStep
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
		policyProto.Steps = steps
	}

	if p.Requirements != nil {
		var requirements []*guardianv1beta1.Policy_Requirement
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
					Expression:   r.On.Expression,
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
			policyProto.Requirements = requirements
		}
	}

	if p.HasIAMConfig() {
		config, err := structpb.NewValue(p.IAM.Config)
		if err != nil {
			return nil, err
		}

		policyProto.Iam = &guardianv1beta1.Policy_IAM{
			Provider: string(p.IAM.Provider),
			Config:   config,
			Schema:   p.IAM.Schema,
		}
	}

	policyProto.Appeal = a.ToPolicyAppealConfigProto(p)

	if !p.CreatedAt.IsZero() {
		policyProto.CreatedAt = timestamppb.New(p.CreatedAt)
	}
	if !p.UpdatedAt.IsZero() {
		policyProto.UpdatedAt = timestamppb.New(p.UpdatedAt)
	}

	return policyProto, nil
}

func (a *adapter) ToPolicyAppealConfigProto(p *domain.Policy) *guardianv1beta1.PolicyAppealConfig {
	if p.AppealConfig == nil {
		return nil
	}

	policyAppealConfigProto := &guardianv1beta1.PolicyAppealConfig{}
	var durationOptions []*guardianv1beta1.PolicyAppealConfig_DurationOptions
	if p.AppealConfig.DurationOptions != nil {
		for _, d := range p.AppealConfig.DurationOptions {
			durationOptions = append(durationOptions, &guardianv1beta1.PolicyAppealConfig_DurationOptions{
				Name:  d.Name,
				Value: d.Value,
			})
		}
	}
	policyAppealConfigProto.DurationOptions = durationOptions
	policyAppealConfigProto.AllowOnBehalf = p.AppealConfig.AllowOnBehalf
	policyAppealConfigProto.AllowPermanentAccess = p.AppealConfig.AllowPermanentAccess
	policyAppealConfigProto.AllowActiveAccessExtensionIn = p.AppealConfig.AllowActiveAccessExtensionIn
	policyAppealConfigProto.AllowCreatorDetailsFailure = p.AppealConfig.AllowCreatorDetailsFailure

	for _, q := range p.AppealConfig.Questions {
		policyAppealConfigProto.Questions = append(policyAppealConfigProto.Questions, &guardianv1beta1.PolicyAppealConfig_Question{
			Key:         q.Key,
			Question:    q.Question,
			Required:    q.Required,
			Description: q.Description,
		})
	}
	return policyAppealConfigProto
}

func (a *adapter) FromResourceProto(r *guardianv1beta1.Resource) *domain.Resource {
	resource := &domain.Resource{
		ID:           r.GetId(),
		ProviderType: r.GetProviderType(),
		ProviderURN:  r.GetProviderUrn(),
		Type:         r.GetType(),
		URN:          r.GetUrn(),
		Name:         r.GetName(),
		Labels:       r.GetLabels(),
		IsDeleted:    r.GetIsDeleted(),
	}

	if r.GetParentId() != "" {
		id := r.GetParentId()
		resource.ParentID = &id
	}

	if r.GetChildren() != nil {
		for _, c := range r.GetChildren() {
			resource.Children = append(resource.Children, a.FromResourceProto(c))
		}
	}

	if r.GetDetails() != nil {
		resource.Details = r.GetDetails().AsMap()
	}

	if r.GetCreatedAt() != nil {
		resource.CreatedAt = r.GetCreatedAt().AsTime()
	}
	if r.GetUpdatedAt() != nil {
		resource.UpdatedAt = r.GetUpdatedAt().AsTime()
	}

	return resource
}

func (a *adapter) ToResourceProto(r *domain.Resource) (*guardianv1beta1.Resource, error) {
	resourceProto := &guardianv1beta1.Resource{
		Id:           r.ID,
		ProviderType: r.ProviderType,
		ProviderUrn:  r.ProviderURN,
		Type:         r.Type,
		Urn:          r.URN,
		Name:         r.Name,
		Labels:       r.Labels,
		IsDeleted:    r.IsDeleted,
	}

	if r.ParentID != nil {
		resourceProto.ParentId = *r.ParentID
	}

	if r.Children != nil {
		for _, c := range r.Children {
			childProto, err := a.ToResourceProto(c)
			if err != nil {
				return nil, fmt.Errorf("failed to convert child resource to proto %q: %w", c.ID, err)
			}
			resourceProto.Children = append(resourceProto.Children, childProto)
		}
	}

	if r.Details != nil {
		details, err := structpb.NewStruct(r.Details)
		if err != nil {
			return nil, err
		}
		resourceProto.Details = details
	}

	if !r.CreatedAt.IsZero() {
		resourceProto.CreatedAt = timestamppb.New(r.CreatedAt)
	}
	if !r.UpdatedAt.IsZero() {
		resourceProto.UpdatedAt = timestamppb.New(r.UpdatedAt)
	}

	return resourceProto, nil
}

func (a *adapter) ToAppealProto(appeal *domain.Appeal) (*guardianv1beta1.Appeal, error) {
	appealProto := &guardianv1beta1.Appeal{
		Id:            appeal.ID,
		ResourceId:    appeal.ResourceID,
		PolicyId:      appeal.PolicyID,
		PolicyVersion: uint32(appeal.PolicyVersion),
		Status:        appeal.Status,
		AccountId:     appeal.AccountID,
		AccountType:   appeal.AccountType,
		CreatedBy:     appeal.CreatedBy,
		Role:          appeal.Role,
		Permissions:   appeal.Permissions,
		Options:       a.toAppealOptionsProto(appeal.Options),
		Labels:        appeal.Labels,
		Description:   appeal.Description,
	}

	if appeal.Resource != nil {
		r, err := a.ToResourceProto(appeal.Resource)
		if err != nil {
			return nil, err
		}
		appealProto.Resource = r
	}

	if appeal.Creator != nil {
		creator, err := structpb.NewValue(appeal.Creator)
		if err != nil {
			return nil, err
		}
		appealProto.Creator = creator
	}

	if appeal.Approvals != nil {
		approvals := []*guardianv1beta1.Approval{}
		for _, approval := range appeal.Approvals {
			approvalProto, err := a.ToApprovalProto(approval)
			if err != nil {
				return nil, err
			}

			approvals = append(approvals, approvalProto)
		}
		appealProto.Approvals = approvals
	}

	if appeal.Details != nil {
		details, err := structpb.NewStruct(appeal.Details)
		if err != nil {
			return nil, err
		}
		appealProto.Details = details
	}

	if !appeal.CreatedAt.IsZero() {
		appealProto.CreatedAt = timestamppb.New(appeal.CreatedAt)
	}
	if !appeal.UpdatedAt.IsZero() {
		appealProto.UpdatedAt = timestamppb.New(appeal.UpdatedAt)
	}

	grantProto, err := a.ToGrantProto(appeal.Grant)
	if err != nil {
		return nil, fmt.Errorf("parsing grant: %w", err)
	}
	appealProto.Grant = grantProto

	return appealProto, nil
}

func (a *adapter) FromCreateAppealProto(ca *guardianv1beta1.CreateAppealRequest, authenticatedUser string) ([]*domain.Appeal, error) {
	var appeals []*domain.Appeal

	for _, r := range ca.GetResources() {
		appeal := &domain.Appeal{
			AccountID:   ca.GetAccountId(),
			AccountType: ca.GetAccountType(),
			CreatedBy:   authenticatedUser,
			ResourceID:  r.GetId(),
			Role:        r.GetRole(),
			Description: ca.GetDescription(),
		}

		if r.GetOptions() != nil {
			var options *domain.AppealOptions
			if err := mapstructure.Decode(r.GetOptions().AsMap(), &options); err != nil {
				return nil, err
			}
			appeal.Options = options
		}

		if r.GetDetails() != nil {
			appeal.Details = r.GetDetails().AsMap()
		}

		appeals = append(appeals, appeal)
	}

	return appeals, nil
}

func (a *adapter) ToApprovalProto(approval *domain.Approval) (*guardianv1beta1.Approval, error) {
	approvalProto := &guardianv1beta1.Approval{
		Id:            approval.ID,
		Name:          approval.Name,
		AppealId:      approval.AppealID,
		Status:        approval.Status,
		Reason:        approval.Reason,
		PolicyId:      approval.PolicyID,
		PolicyVersion: uint32(approval.PolicyVersion),
		Approvers:     approval.Approvers,
		CreatedAt:     timestamppb.New(approval.CreatedAt),
		UpdatedAt:     timestamppb.New(approval.UpdatedAt),
	}

	if approval.Appeal != nil {
		appeal, err := a.ToAppealProto(approval.Appeal)
		if err != nil {
			return nil, err
		}
		approvalProto.Appeal = appeal
	}

	if approval.Actor != nil {
		approvalProto.Actor = *approval.Actor
	}

	if !approval.CreatedAt.IsZero() {
		approvalProto.CreatedAt = timestamppb.New(approval.CreatedAt)
	}
	if !approval.UpdatedAt.IsZero() {
		approvalProto.UpdatedAt = timestamppb.New(approval.UpdatedAt)
	}

	return approvalProto, nil
}

func (a *adapter) FromGrantProto(g *guardianv1beta1.Grant) *domain.Grant {
	if g == nil {
		return nil
	}

	grant := &domain.Grant{
		ID:               g.GetId(),
		Status:           domain.GrantStatus(g.GetStatus()),
		StatusInProvider: domain.GrantStatus(g.GetStatusInProvider()),
		AccountID:        g.GetAccountId(),
		AccountType:      g.GetAccountType(),
		ResourceID:       g.GetResourceId(),
		Role:             g.GetRole(),
		Permissions:      g.GetPermissions(),
		AppealID:         g.GetAppealId(),
		Source:           domain.GrantSource(g.Source),
		RevokedBy:        g.GetRevokedBy(),
		RevokeReason:     g.GetRevokeReason(),
		CreatedBy:        g.GetCreatedBy(),
		Owner:            g.GetOwner(),
		Resource:         a.FromResourceProto(g.GetResource()),
	}

	if g.GetExpirationDate() != nil {
		t := g.GetExpirationDate().AsTime()
		grant.ExpirationDate = &t
	}
	if g.GetRevokedAt() != nil {
		t := g.GetRevokedAt().AsTime()
		grant.RevokedAt = &t
	}
	if g.GetCreatedAt() != nil {
		grant.CreatedAt = g.GetCreatedAt().AsTime()
	}
	if g.GetUpdatedAt() != nil {
		grant.UpdatedAt = g.GetUpdatedAt().AsTime()
	}

	return grant
}

func (a *adapter) ToGrantProto(grant *domain.Grant) (*guardianv1beta1.Grant, error) {
	if grant == nil {
		return nil, nil
	}

	grantProto := &guardianv1beta1.Grant{
		Id:               grant.ID,
		Status:           string(grant.Status),
		StatusInProvider: string(grant.StatusInProvider),
		AccountId:        grant.AccountID,
		AccountType:      grant.AccountType,
		ResourceId:       grant.ResourceID,
		Role:             grant.Role,
		Permissions:      grant.Permissions,
		IsPermanent:      grant.IsPermanent,
		AppealId:         grant.AppealID,
		Source:           string(grant.Source),
		RevokedBy:        grant.RevokedBy,
		RevokeReason:     grant.RevokeReason,
		CreatedBy:        grant.CreatedBy,
		Owner:            grant.Owner,
	}

	if grant.ExpirationDate != nil {
		grantProto.ExpirationDate = timestamppb.New(*grant.ExpirationDate)
	}
	if grant.RevokedAt != nil {
		grantProto.RevokedAt = timestamppb.New(*grant.RevokedAt)
	}
	if !grant.CreatedAt.IsZero() {
		grantProto.CreatedAt = timestamppb.New(grant.CreatedAt)
	}
	if !grant.UpdatedAt.IsZero() {
		grantProto.UpdatedAt = timestamppb.New(grant.UpdatedAt)
	}
	if grant.Resource != nil {
		resourceProto, err := a.ToResourceProto(grant.Resource)
		if err != nil {
			return nil, fmt.Errorf("parsing resource: %w", err)
		}
		grantProto.Resource = resourceProto
	}
	if grant.Appeal != nil {
		appealProto, err := a.ToAppealProto(grant.Appeal)
		if err != nil {
			return nil, fmt.Errorf("parsing appeal: %w", err)
		}
		grantProto.Appeal = appealProto
	}

	return grantProto, nil
}

func (a *adapter) ToActivityProto(activity *domain.Activity) (*guardianv1beta1.ProviderActivity, error) {
	if activity == nil {
		return nil, nil
	}

	activityProto := &guardianv1beta1.ProviderActivity{
		Id:                 activity.ID,
		ProviderId:         activity.ProviderID,
		ResourceId:         activity.ResourceID,
		ProviderActivityId: activity.ProviderActivityID,
		AccountType:        activity.AccountType,
		AccountId:          activity.AccountID,
		Authorizations:     activity.Authorizations,
		RelatedPermissions: activity.RelatedPermissions,
		Type:               activity.Type,
	}

	if !activity.Timestamp.IsZero() {
		activityProto.Timestamp = timestamppb.New(activity.Timestamp)
	}

	if activity.Metadata != nil {
		metadataStruct, err := structpb.NewStruct(activity.Metadata)
		if err != nil {
			return nil, fmt.Errorf("parsing metadata: %w", err)
		}
		activityProto.Metadata = metadataStruct
	}

	if !activity.CreatedAt.IsZero() {
		activityProto.CreatedAt = timestamppb.New(activity.CreatedAt)
	}

	if activity.Provider != nil {
		providerProto, err := a.ToProviderProto(activity.Provider)
		if err != nil {
			return nil, fmt.Errorf("parsing provider: %w", err)
		}
		activityProto.Provider = providerProto
	}

	if activity.Resource != nil {
		resourceProto, err := a.ToResourceProto(activity.Resource)
		if err != nil {
			return nil, fmt.Errorf("parsing resource: %w", err)
		}
		activityProto.Resource = resourceProto
	}

	return activityProto, nil
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

	options := &domain.AppealOptions{
		Duration: o.GetDuration(),
	}

	if o.GetExpirationDate() != nil {
		expDate := o.GetExpirationDate().AsTime()
		options.ExpirationDate = &expDate
	}

	return options
}

func (a *adapter) toAppealOptionsProto(o *domain.AppealOptions) *guardianv1beta1.AppealOptions {
	if o == nil {
		return nil
	}

	optionsProto := &guardianv1beta1.AppealOptions{
		Duration: o.Duration,
	}

	if o.ExpirationDate != nil {
		optionsProto.ExpirationDate = timestamppb.New(*o.ExpirationDate)
	}

	return optionsProto
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
