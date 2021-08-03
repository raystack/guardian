package v1

import (
	"github.com/golang/protobuf/ptypes"
	pb "github.com/odpf/guardian/api/proto/guardian"
	"github.com/odpf/guardian/domain"
	"google.golang.org/protobuf/types/known/structpb"
)

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

	createdAt, err := ptypes.TimestampProto(p.CreatedAt)
	if err != nil {
		return nil, err
	}
	updatedAt, err := ptypes.TimestampProto(p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &pb.Provider{
		Id:        uint32(p.ID),
		Type:      p.Type,
		Urn:       p.URN,
		Config:    config,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func (a *adapter) FromPolicyProto(p *pb.Policy) (*domain.Policy, error) {
	steps := []*domain.Step{}
	for _, s := range p.Steps {
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

	createdAt, err := ptypes.TimestampProto(p.CreatedAt)
	if err != nil {
		return nil, err
	}
	updatedAt, err := ptypes.TimestampProto(p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &pb.Policy{
		Id:          p.ID,
		Version:     uint32(p.Version),
		Description: p.Description,
		Steps:       approvalSteps,
		Labels:      p.Labels,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}
