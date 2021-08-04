package v1

import (
	"time"

	"github.com/mitchellh/mapstructure"
	pb "github.com/odpf/guardian/api/proto/guardian"
	"github.com/odpf/guardian/domain"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type resourceOptions struct {
	Duration string `json:"duration"`
}

type adapter struct{}

func NewAdapter() *adapter {
	return &adapter{}
}

func (a *adapter) FromProviderProto(p *pb.Provider) (*domain.Provider, error) {
	providerConfig, err := a.FromProviderConfigProto(p.GetConfig())
	if err != nil {
		return nil, err
	}

	return &domain.Provider{
		ID:        uint(p.GetId()),
		Type:      p.GetType(),
		URN:       p.GetUrn(),
		Config:    providerConfig,
		CreatedAt: p.GetCreatedAt().AsTime(),
		UpdatedAt: p.GetUpdatedAt().AsTime(),
	}, nil
}

func (a *adapter) FromProviderConfigProto(pc *pb.ProviderConfig) (*domain.ProviderConfig, error) {
	appeal := pc.GetAppeal()
	resources := []*domain.ResourceConfig{}
	for _, r := range pc.GetResources() {
		policyProto := r.GetPolicy()
		policy := &domain.PolicyConfig{
			ID:      policyProto.GetId(),
			Version: int(policyProto.GetVersion()),
		}

		roles := []*domain.RoleConfig{}
		for _, role := range r.GetRoles() {
			permissions := []interface{}{}
			for _, p := range role.GetPermissions() {
				permissions = append(permissions, p.AsInterface())
			}

			roles = append(roles, &domain.RoleConfig{
				ID:          role.GetId(),
				Name:        role.GetName(),
				Description: role.GetDescription(),
				Permissions: permissions,
			})
		}

		resources = append(resources, &domain.ResourceConfig{
			Type:   r.GetType(),
			Policy: policy,
			Roles:  roles,
		})
	}

	return &domain.ProviderConfig{
		Type:        pc.GetType(),
		URN:         pc.GetUrn(),
		Labels:      pc.GetLabels(),
		Credentials: pc.GetCredentials().AsInterface(),
		Appeal: &domain.AppealConfig{
			AllowPermanentAccess:         appeal.GetAllowPermanentAccess(),
			AllowActiveAccessExtensionIn: appeal.GetAllowActiveAccessExtensionIn(),
		},
		Resources: resources,
	}, nil
}

func (a *adapter) ToProviderProto(p *domain.Provider) (*pb.Provider, error) {
	credentials, err := structpb.NewValue(p.Config.Credentials)
	if err != nil {
		return nil, err
	}

	appeal := &pb.ProviderConfig_AppealConfig{
		AllowPermanentAccess:         p.Config.Appeal.AllowPermanentAccess,
		AllowActiveAccessExtensionIn: p.Config.Appeal.AllowActiveAccessExtensionIn,
	}

	resources := []*pb.ProviderConfig_ResourceConfig{}
	for _, rc := range p.Config.Resources {
		policy := &pb.ProviderConfig_ResourceConfig_PolicyConfig{
			Id:      rc.Policy.ID,
			Version: int32(rc.Policy.Version),
		}

		roles := []*pb.ProviderConfig_ResourceConfig_RoleConfig{}
		for _, role := range rc.Roles {
			permissions := []*structpb.Value{}
			for _, p := range role.Permissions {
				permission, err := structpb.NewValue(p)
				if err != nil {
					return nil, err
				}
				permissions = append(permissions, permission)
			}

			roles = append(roles, &pb.ProviderConfig_ResourceConfig_RoleConfig{
				Id:          role.ID,
				Name:        role.Name,
				Description: role.Description,
				Permissions: permissions,
			})
		}

		resources = append(resources, &pb.ProviderConfig_ResourceConfig{
			Type:   rc.Type,
			Policy: policy,
			Roles:  roles,
		})
	}

	config := &pb.ProviderConfig{
		Type:        p.Config.Type,
		Urn:         p.Config.URN,
		Labels:      p.Config.Labels,
		Credentials: credentials,
		Appeal:      appeal,
		Resources:   resources,
	}

	return &pb.Provider{
		Id:        uint32(p.ID),
		Type:      p.Type,
		Urn:       p.URN,
		Config:    config,
		CreatedAt: timestamppb.New(p.CreatedAt),
		UpdatedAt: timestamppb.New(p.UpdatedAt),
	}, nil
}

func (a *adapter) FromPolicyProto(p *pb.Policy) (*domain.Policy, error) {
	steps := []*domain.Step{}
	for _, s := range p.GetSteps() {
		conditions := []*domain.Condition{}
		for _, c := range s.Conditions {
			match := &domain.MatchCondition{
				Eq: c.GetMatch().GetEq().AsInterface(),
			}

			conditions = append(conditions, &domain.Condition{
				Field: c.GetField(),
				Match: match,
			})
		}

		steps = append(steps, &domain.Step{
			Name:         s.GetName(),
			Description:  s.GetDescription(),
			Conditions:   conditions,
			AllowFailed:  s.GetAllowFailed(),
			Dependencies: s.GetDependencies(),
			Approvers:    s.GetApprovers(),
		})
	}

	return &domain.Policy{
		ID:          p.GetId(),
		Version:     uint(p.GetVersion()),
		Description: p.GetDescription(),
		Steps:       steps,
		Labels:      p.GetLabels(),
		CreatedAt:   p.GetCreatedAt().AsTime(),
		UpdatedAt:   p.GetUpdatedAt().AsTime(),
	}, nil
}

func (a *adapter) ToPolicyProto(p *domain.Policy) (*pb.Policy, error) {
	approvalSteps := []*pb.Policy_ApprovalStep{}
	for _, s := range p.Steps {
		conditions := []*pb.Policy_ApprovalStep_Condition{}
		for _, c := range s.Conditions {
			eqCondition, err := structpb.NewValue(c.Match.Eq)
			if err != nil {
				return nil, err
			}

			match := &pb.Policy_ApprovalStep_Condition_MatchCondition{
				Eq: eqCondition,
			}
			conditions = append(conditions, &pb.Policy_ApprovalStep_Condition{
				Field: c.Field,
				Match: match,
			})
		}

		approvalSteps = append(approvalSteps, &pb.Policy_ApprovalStep{
			Name:         s.Name,
			Description:  s.Description,
			Conditions:   conditions,
			AllowFailed:  s.AllowFailed,
			Dependencies: s.Dependencies,
			Approvers:    s.Approvers,
		})
	}

	return &pb.Policy{
		Id:          p.ID,
		Version:     uint32(p.Version),
		Description: p.Description,
		Steps:       approvalSteps,
		Labels:      p.Labels,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}, nil
}

func (a *adapter) FromResourceProto(r *pb.Resource) *domain.Resource {
	details := map[string]interface{}{}
	if r.GetDetails() != nil {
		details = r.GetDetails().AsMap()
	}
	return &domain.Resource{
		ID:           uint(r.GetId()),
		ProviderType: r.GetProviderType(),
		ProviderURN:  r.GetProviderUrn(),
		Type:         r.GetType(),
		URN:          r.GetUrn(),
		Name:         r.GetName(),
		Details:      details,
		Labels:       r.GetLabels(),
		CreatedAt:    r.GetCreatedAt().AsTime(),
		UpdatedAt:    r.GetUpdatedAt().AsTime(),
	}
}

func (a *adapter) ToResourceProto(r *domain.Resource) (*pb.Resource, error) {
	var detailsProto *structpb.Struct
	if r.Details != nil {
		details, err := structpb.NewStruct(r.Details)
		if err != nil {
			return nil, err
		}
		detailsProto = details
	}

	return &pb.Resource{
		Id:           uint32(r.ID),
		ProviderType: r.ProviderType,
		ProviderUrn:  r.ProviderURN,
		Type:         r.Type,
		Urn:          r.URN,
		Name:         r.Name,
		Details:      detailsProto,
		Labels:       r.Labels,
		CreatedAt:    timestamppb.New(r.CreatedAt),
		UpdatedAt:    timestamppb.New(r.UpdatedAt),
	}, nil
}

func (a *adapter) FromAppealProto(appeal *pb.Appeal) (*domain.Appeal, error) {
	expirationDate := appeal.GetOptions().GetExpirationDate().AsTime()
	options := &domain.AppealOptions{
		ExpirationDate: &expirationDate,
	}

	resource := a.FromResourceProto(appeal.GetResource())

	approvals := []*domain.Approval{}
	for _, a := range appeal.GetApprovals() {
		actor := a.GetActor()
		approvals = append(approvals, &domain.Approval{
			ID:            uint(a.GetId()),
			Name:          a.GetName(),
			AppealID:      uint(a.GetId()),
			Status:        a.GetStatus(),
			Actor:         &actor,
			PolicyID:      a.GetPolicyId(),
			PolicyVersion: uint(a.GetPolicyVersion()),
			Approvers:     a.GetApprovers(),
			CreatedAt:     appeal.GetCreatedAt().AsTime(),
			UpdatedAt:     appeal.GetUpdatedAt().AsTime(),
		})
	}

	return &domain.Appeal{
		ID:            uint(appeal.GetId()),
		ResourceID:    uint(appeal.GetResourceId()),
		PolicyID:      appeal.GetPolicyId(),
		PolicyVersion: uint(appeal.GetPolicyVersion()),
		Status:        appeal.GetStatus(),
		User:          appeal.GetUser(),
		Role:          appeal.GetRole(),
		Options:       options,
		Labels:        appeal.GetLabels(),
		Resource:      resource,
		Approvals:     approvals,
		CreatedAt:     appeal.GetCreatedAt().AsTime(),
		UpdatedAt:     appeal.GetUpdatedAt().AsTime(),
	}, nil
}

func (a *adapter) ToAppealProto(appeal *domain.Appeal) (*pb.Appeal, error) {
	expirationDate := timestamppb.New(*appeal.Options.ExpirationDate)
	options := &pb.Appeal_AppealOptions{
		ExpirationDate: expirationDate,
	}

	var resource *pb.Resource
	if appeal.Resource != nil {
		r, err := a.ToResourceProto(appeal.Resource)
		if err != nil {
			return nil, err
		}
		resource = r
	}

	approvals := []*pb.Approval{}
	for _, approval := range appeal.Approvals {
		approvalProto, err := a.ToApprovalProto(approval)
		if err != nil {
			return nil, err
		}

		approvals = append(approvals, approvalProto)
	}

	return &pb.Appeal{
		Id:            uint32(appeal.ID),
		ResourceId:    uint32(appeal.ResourceID),
		PolicyId:      appeal.PolicyID,
		PolicyVersion: uint32(appeal.PolicyVersion),
		Status:        appeal.Status,
		User:          appeal.User,
		Role:          appeal.Role,
		Options:       options,
		Labels:        appeal.Labels,
		Resource:      resource,
		Approvals:     approvals,
		CreatedAt:     timestamppb.New(appeal.CreatedAt),
		UpdatedAt:     timestamppb.New(appeal.UpdatedAt),
	}, nil
}

func (a *adapter) FromCreateAppealProto(ca *pb.CreateAppealRequest) ([]*domain.Appeal, error) {
	appeals := []*domain.Appeal{}

	for _, r := range ca.GetResources() {
		var options domain.AppealOptions

		var resOptions resourceOptions
		if err := mapstructure.Decode(r.GetOptions().AsMap(), &resOptions); err != nil {
			return nil, err
		}

		var expirationDate time.Time
		if r.GetOptions() != nil {
			if resOptions.Duration != "" {
				duration, err := time.ParseDuration(resOptions.Duration)
				if err != nil {
					return nil, err
				}
				expirationDate = time.Now().Add(duration)
			}
		}
		options.ExpirationDate = &expirationDate

		appeals = append(appeals, &domain.Appeal{
			User:       ca.GetUser(),
			ResourceID: uint(r.GetId()),
			Role:       r.GetRole(),
			Options:    &options,
		})
	}

	return appeals, nil
}

func (a *adapter) ToApprovalProto(approval *domain.Approval) (*pb.Approval, error) {
	var appealProto *pb.Appeal
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

	return &pb.Approval{
		Id:            uint32(approval.ID),
		Name:          approval.Name,
		AppealId:      uint32(approval.AppealID),
		Appeal:        appealProto,
		Status:        approval.Status,
		Actor:         actor,
		PolicyId:      approval.PolicyID,
		PolicyVersion: uint32(approval.PolicyVersion),
		Approvers:     approval.Approvers,
		CreatedAt:     timestamppb.New(approval.CreatedAt),
		UpdatedAt:     timestamppb.New(approval.UpdatedAt),
	}, nil
}
