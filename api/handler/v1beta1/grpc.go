//go:generate mockery --name=resourceService --exported --with-expecter
//go:generate mockery --name=providerService --exported --with-expecter
//go:generate mockery --name=policyService --exported --with-expecter
//go:generate mockery --name=appealService --exported --with-expecter
//go:generate mockery --name=approvalService --exported --with-expecter

package v1beta1

import (
	"context"
	"errors"

	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/core/appeal"
	"github.com/odpf/guardian/core/policy"
	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ProtoAdapter interface {
	FromProviderProto(*guardianv1beta1.Provider) (*domain.Provider, error)
	FromProviderConfigProto(*guardianv1beta1.ProviderConfig) (*domain.ProviderConfig, error)
	ToProviderProto(*domain.Provider) (*guardianv1beta1.Provider, error)
	ToProviderConfigProto(*domain.ProviderConfig) (*guardianv1beta1.ProviderConfig, error)
	ToProviderTypeProto(domain.ProviderType) (*guardianv1beta1.ProviderType, error)
	ToRole(*domain.Role) (*guardianv1beta1.Role, error)

	FromPolicyProto(*guardianv1beta1.Policy) (*domain.Policy, error)
	ToPolicyProto(*domain.Policy) (*guardianv1beta1.Policy, error)

	FromResourceProto(*guardianv1beta1.Resource) *domain.Resource
	ToResourceProto(*domain.Resource) (*guardianv1beta1.Resource, error)

	FromAppealProto(*guardianv1beta1.Appeal) (*domain.Appeal, error)
	ToAppealProto(*domain.Appeal) (*guardianv1beta1.Appeal, error)
	FromCreateAppealProto(*guardianv1beta1.CreateAppealRequest, string) ([]*domain.Appeal, error)
	ToApprovalProto(*domain.Approval) (*guardianv1beta1.Approval, error)
}

type resourceService interface {
	Find(context.Context, map[string]interface{}) ([]*domain.Resource, error)
	GetOne(string) (*domain.Resource, error)
	BulkUpsert(context.Context, []*domain.Resource) error
	Update(context.Context, *domain.Resource) error
	Get(context.Context, *domain.ResourceIdentifier) (*domain.Resource, error)
	Delete(context.Context, string) error
	BatchDelete(context.Context, []string) error
}

type providerService interface {
	Create(context.Context, *domain.Provider) error
	Find(context.Context) ([]*domain.Provider, error)
	GetByID(context.Context, string) (*domain.Provider, error)
	GetTypes(context.Context) ([]domain.ProviderType, error)
	GetOne(ctx context.Context, pType, urn string) (*domain.Provider, error)
	Update(context.Context, *domain.Provider) error
	FetchResources(context.Context) error
	GetRoles(ctx context.Context, id, resourceType string) ([]*domain.Role, error)
	ValidateAppeal(context.Context, *domain.Appeal, *domain.Provider) error
	GrantAccess(context.Context, *domain.Appeal) error
	RevokeAccess(context.Context, *domain.Appeal) error
	Delete(context.Context, string) error
}

type policyService interface {
	Create(context.Context, *domain.Policy) error
	Find(context.Context) ([]*domain.Policy, error)
	GetOne(ctx context.Context, id string, version uint) (*domain.Policy, error)
	Update(context.Context, *domain.Policy) error
}

type appealService interface {
	GetByID(context.Context, string) (*domain.Appeal, error)
	Find(context.Context, *domain.ListAppealsFilter) ([]*domain.Appeal, error)
	Create(context.Context, []*domain.Appeal) error
	MakeAction(context.Context, domain.ApprovalAction) (*domain.Appeal, error)
	Cancel(context.Context, string) (*domain.Appeal, error)
	Revoke(ctx context.Context, id, actor, reason string) (*domain.Appeal, error)
}

type approvalService interface {
	ListApprovals(context.Context, *domain.ListApprovalsFilter) ([]*domain.Approval, error)
	BulkInsert(context.Context, []*domain.Approval) error
	AdvanceApproval(context.Context, *domain.Appeal) error
}

type GRPCServer struct {
	resourceService resourceService
	providerService providerService
	policyService   policyService
	appealService   appealService
	approvalService approvalService
	adapter         ProtoAdapter

	authenticatedUserHeaderKey string

	guardianv1beta1.UnimplementedGuardianServiceServer
}

func NewGRPCServer(
	resourceService resourceService,
	providerService providerService,
	policyService policyService,
	appealService appealService,
	approvalService approvalService,
	adapter ProtoAdapter,
	authenticatedUserHeaderKey string,
) *GRPCServer {
	return &GRPCServer{
		resourceService:            resourceService,
		providerService:            providerService,
		policyService:              policyService,
		appealService:              appealService,
		approvalService:            approvalService,
		adapter:                    adapter,
		authenticatedUserHeaderKey: authenticatedUserHeaderKey,
	}
}

func (s *GRPCServer) ListProviders(ctx context.Context, req *guardianv1beta1.ListProvidersRequest) (*guardianv1beta1.ListProvidersResponse, error) {
	providers, err := s.providerService.Find(ctx)
	if err != nil {
		return nil, err
	}

	providerProtos := []*guardianv1beta1.Provider{}
	for _, p := range providers {
		p.Config.Credentials = nil
		providerProto, err := s.adapter.ToProviderProto(p)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse provider %s: %v", p.URN, err)
		}
		providerProtos = append(providerProtos, providerProto)
	}

	return &guardianv1beta1.ListProvidersResponse{
		Providers: providerProtos,
	}, nil
}

func (s *GRPCServer) GetProvider(ctx context.Context, req *guardianv1beta1.GetProviderRequest) (*guardianv1beta1.GetProviderResponse, error) {
	p, err := s.providerService.GetByID(ctx, req.GetId())
	if err != nil {
		switch err {
		case provider.ErrRecordNotFound:
			return nil, status.Error(codes.NotFound, "provider not found")
		default:
			return nil, status.Errorf(codes.Internal, "failed to retrieve provider: %v", err)
		}
	}

	providerProto, err := s.adapter.ToProviderProto(p)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse provider %s: %v", p.URN, err)
	}

	return &guardianv1beta1.GetProviderResponse{
		Provider: providerProto,
	}, nil
}

func (s *GRPCServer) GetProviderTypes(ctx context.Context, req *guardianv1beta1.GetProviderTypesRequest) (*guardianv1beta1.GetProviderTypesResponse, error) {
	providerTypes, err := s.providerService.GetTypes(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve provider types: %v", err)
	}

	var providerTypeProtos []*guardianv1beta1.ProviderType
	for _, pt := range providerTypes {
		providerTypeProto, err := s.adapter.ToProviderTypeProto(pt)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse provider type %s: %v", pt.Name, err)
		}
		providerTypeProtos = append(providerTypeProtos, providerTypeProto)
	}

	return &guardianv1beta1.GetProviderTypesResponse{
		ProviderTypes: providerTypeProtos,
	}, nil
}

func (s *GRPCServer) CreateProvider(ctx context.Context, req *guardianv1beta1.CreateProviderRequest) (*guardianv1beta1.CreateProviderResponse, error) {
	providerConfig, err := s.adapter.FromProviderConfigProto(req.GetConfig())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot deserialize provider config: %v", err)
	}

	p := &domain.Provider{
		Type:   providerConfig.Type,
		URN:    providerConfig.URN,
		Config: providerConfig,
	}

	if err := s.providerService.Create(ctx, p); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create provider: %v", err)
	}

	providerProto, err := s.adapter.ToProviderProto(p)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse provider: %v", err)
	}

	return &guardianv1beta1.CreateProviderResponse{
		Provider: providerProto,
	}, nil
}

func (s *GRPCServer) UpdateProvider(ctx context.Context, req *guardianv1beta1.UpdateProviderRequest) (*guardianv1beta1.UpdateProviderResponse, error) {
	id := req.GetId()
	providerConfig, err := s.adapter.FromProviderConfigProto(req.GetConfig())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot deserialize provider config: %v", err)
	}

	p := &domain.Provider{
		ID:     id,
		Type:   providerConfig.Type,
		URN:    providerConfig.URN,
		Config: providerConfig,
	}

	if err := s.providerService.Update(ctx, p); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update provider: %v", err)
	}

	providerProto, err := s.adapter.ToProviderProto(p)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse provider: %v", err)
	}

	return &guardianv1beta1.UpdateProviderResponse{
		Provider: providerProto,
	}, nil
}

func (s *GRPCServer) DeleteProvider(ctx context.Context, req *guardianv1beta1.DeleteProviderRequest) (*guardianv1beta1.DeleteProviderResponse, error) {
	if err := s.providerService.Delete(ctx, req.GetId()); err != nil {
		if errors.Is(err, provider.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "provider not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete provider: %v", err)
	}

	return &guardianv1beta1.DeleteProviderResponse{}, nil
}

func (s *GRPCServer) ListRoles(ctx context.Context, req *guardianv1beta1.ListRolesRequest) (*guardianv1beta1.ListRolesResponse, error) {
	roles, err := s.providerService.GetRoles(ctx, req.GetId(), req.GetResourceType())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list roles: %v", err)
	}

	roleProtos := []*guardianv1beta1.Role{}
	for _, r := range roles {
		role, err := s.adapter.ToRole(r)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse proto: %v", err)
		}

		roleProtos = append(roleProtos, role)
	}

	return &guardianv1beta1.ListRolesResponse{
		Roles: roleProtos,
	}, nil
}

func (s *GRPCServer) ListPolicies(ctx context.Context, req *guardianv1beta1.ListPoliciesRequest) (*guardianv1beta1.ListPoliciesResponse, error) {
	policies, err := s.policyService.Find(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get policy list: %v", err)
	}

	policyProtos := []*guardianv1beta1.Policy{}
	for _, p := range policies {
		policyProto, err := s.adapter.ToPolicyProto(p)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse policy %v: %v", p.ID, err)
		}
		policyProtos = append(policyProtos, policyProto)
	}

	return &guardianv1beta1.ListPoliciesResponse{
		Policies: policyProtos,
	}, nil
}

func (s *GRPCServer) GetPolicy(ctx context.Context, req *guardianv1beta1.GetPolicyRequest) (*guardianv1beta1.GetPolicyResponse, error) {
	p, err := s.policyService.GetOne(ctx, req.GetId(), uint(req.GetVersion()))
	if err != nil {
		switch err {
		case policy.ErrPolicyNotFound:
			return nil, status.Error(codes.NotFound, "policy not found")
		default:
			return nil, status.Errorf(codes.Internal, "failed to retrieve policy: %v", err)
		}
	}

	policyProto, err := s.adapter.ToPolicyProto(p)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse policy: %v", err)
	}

	return &guardianv1beta1.GetPolicyResponse{
		Policy: policyProto,
	}, nil
}

func (s *GRPCServer) CreatePolicy(ctx context.Context, req *guardianv1beta1.CreatePolicyRequest) (*guardianv1beta1.CreatePolicyResponse, error) {
	policy, err := s.adapter.FromPolicyProto(req.GetPolicy())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot deserialize policy: %v", err)
	}

	if err := s.policyService.Create(ctx, policy); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create policy: %v", err)
	}

	policyProto, err := s.adapter.ToPolicyProto(policy)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse policy: %v", err)
	}

	return &guardianv1beta1.CreatePolicyResponse{
		Policy: policyProto,
	}, nil
}

func (s *GRPCServer) UpdatePolicy(ctx context.Context, req *guardianv1beta1.UpdatePolicyRequest) (*guardianv1beta1.UpdatePolicyResponse, error) {
	p, err := s.adapter.FromPolicyProto(req.GetPolicy())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot deserialize policy: %v", err)
	}

	p.ID = req.GetId()
	if err := s.policyService.Update(ctx, p); err != nil {
		if errors.Is(err, policy.ErrPolicyNotFound) {
			return nil, status.Error(codes.NotFound, "policy not found")
		} else if errors.Is(err, policy.ErrEmptyIDParam) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return nil, status.Errorf(codes.Internal, "failed to update policy: %v", err)
	}

	policyProto, err := s.adapter.ToPolicyProto(p)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse policy: %v", err)
	}

	return &guardianv1beta1.UpdatePolicyResponse{
		Policy: policyProto,
	}, nil
}

func (s *GRPCServer) ListUserAppeals(ctx context.Context, req *guardianv1beta1.ListUserAppealsRequest) (*guardianv1beta1.ListUserAppealsResponse, error) {
	user, err := s.getUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	filters := &domain.ListAppealsFilter{
		CreatedBy: user,
	}
	if req.GetStatuses() != nil {
		filters.Statuses = req.GetStatuses()
	}
	if req.GetRole() != "" {
		filters.Role = req.GetRole()
	}
	if req.GetProviderTypes() != nil {
		filters.ProviderTypes = req.GetProviderTypes()
	}
	if req.GetProviderUrns() != nil {
		filters.ProviderURNs = req.GetProviderUrns()
	}
	if req.GetResourceTypes() != nil {
		filters.ResourceTypes = req.GetResourceTypes()
	}
	if req.GetResourceUrns() != nil {
		filters.ResourceURNs = req.GetResourceUrns()
	}
	if req.GetOrderBy() != nil {
		filters.OrderBy = req.GetOrderBy()
	}
	appeals, err := s.listAppeals(ctx, filters)
	if err != nil {
		return nil, err
	}

	return &guardianv1beta1.ListUserAppealsResponse{
		Appeals: appeals,
	}, nil
}

func (s *GRPCServer) ListAppeals(ctx context.Context, req *guardianv1beta1.ListAppealsRequest) (*guardianv1beta1.ListAppealsResponse, error) {
	filters := &domain.ListAppealsFilter{}
	if req.GetAccountId() != "" {
		filters.AccountID = req.GetAccountId()
	}
	if req.GetStatuses() != nil {
		filters.Statuses = req.GetStatuses()
	}
	if req.GetRole() != "" {
		filters.Role = req.GetRole()
	}
	if req.GetProviderTypes() != nil {
		filters.ProviderTypes = req.GetProviderTypes()
	}
	if req.GetProviderUrns() != nil {
		filters.ProviderURNs = req.GetProviderUrns()
	}
	if req.GetResourceTypes() != nil {
		filters.ResourceTypes = req.GetResourceTypes()
	}
	if req.GetResourceUrns() != nil {
		filters.ResourceURNs = req.GetResourceUrns()
	}
	if req.GetOrderBy() != nil {
		filters.OrderBy = req.GetOrderBy()
	}
	appeals, err := s.listAppeals(ctx, filters)
	if err != nil {
		return nil, err
	}

	return &guardianv1beta1.ListAppealsResponse{
		Appeals: appeals,
	}, nil
}

func (s *GRPCServer) CreateAppeal(ctx context.Context, req *guardianv1beta1.CreateAppealRequest) (*guardianv1beta1.CreateAppealResponse, error) {
	authenticatedUser, err := s.getUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	appeals, err := s.adapter.FromCreateAppealProto(req, authenticatedUser)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot deserialize payload: %v", err)
	}

	if err := s.appealService.Create(ctx, appeals); err != nil {
		if errors.Is(err, appeal.ErrAppealDuplicate) {
			return nil, status.Errorf(codes.AlreadyExists, "appeal already exists: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to create appeal: %v", err)
	}

	appealProtos := []*guardianv1beta1.Appeal{}
	for _, appeal := range appeals {
		appealProto, err := s.adapter.ToAppealProto(appeal)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse appeal: %v", err)
		}
		appealProtos = append(appealProtos, appealProto)
	}

	return &guardianv1beta1.CreateAppealResponse{
		Appeals: appealProtos,
	}, nil
}

func (s *GRPCServer) ListUserApprovals(ctx context.Context, req *guardianv1beta1.ListUserApprovalsRequest) (*guardianv1beta1.ListUserApprovalsResponse, error) {
	user, err := s.getUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	approvals, err := s.listApprovals(ctx, &domain.ListApprovalsFilter{
		AccountID: req.GetAccountId(),
		CreatedBy: user,
		Statuses:  req.GetStatuses(),
		OrderBy:   req.GetOrderBy(),
	})
	if err != nil {
		return nil, err
	}

	return &guardianv1beta1.ListUserApprovalsResponse{
		Approvals: approvals,
	}, nil
}

func (s *GRPCServer) ListApprovals(ctx context.Context, req *guardianv1beta1.ListApprovalsRequest) (*guardianv1beta1.ListApprovalsResponse, error) {
	approvals, err := s.listApprovals(ctx, &domain.ListApprovalsFilter{
		AccountID: req.GetAccountId(),
		CreatedBy: req.GetCreatedBy(),
		Statuses:  req.GetStatuses(),
		OrderBy:   req.GetOrderBy(),
	})
	if err != nil {
		return nil, err
	}

	return &guardianv1beta1.ListApprovalsResponse{
		Approvals: approvals,
	}, nil
}

func (s *GRPCServer) GetAppeal(ctx context.Context, req *guardianv1beta1.GetAppealRequest) (*guardianv1beta1.GetAppealResponse, error) {
	id := req.GetId()
	appeal, err := s.appealService.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve appeal: %v", err)
	}
	if appeal == nil {
		return nil, status.Errorf(codes.NotFound, "appeal not found: %v", id)
	}

	appealProto, err := s.adapter.ToAppealProto(appeal)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse appeal: %v", err)
	}

	return &guardianv1beta1.GetAppealResponse{
		Appeal: appealProto,
	}, nil
}

func (s *GRPCServer) UpdateApproval(ctx context.Context, req *guardianv1beta1.UpdateApprovalRequest) (*guardianv1beta1.UpdateApprovalResponse, error) {
	actor, err := s.getUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	id := req.GetId()
	a, err := s.appealService.MakeAction(ctx, domain.ApprovalAction{
		AppealID:     id,
		ApprovalName: req.GetApprovalName(),
		Actor:        actor,
		Action:       req.GetAction().GetAction(),
		Reason:       req.GetAction().GetReason(),
	})
	if err != nil {
		switch err {
		case appeal.ErrAppealStatusCanceled,
			appeal.ErrAppealStatusApproved,
			appeal.ErrAppealStatusRejected,
			appeal.ErrAppealStatusTerminated,
			appeal.ErrAppealStatusUnrecognized,
			appeal.ErrApprovalDependencyIsPending,
			appeal.ErrAppealStatusRejected,
			appeal.ErrApprovalStatusUnrecognized,
			appeal.ErrApprovalStatusApproved,
			appeal.ErrApprovalStatusRejected,
			appeal.ErrApprovalStatusSkipped,
			appeal.ErrActionInvalidValue:
			return nil, status.Errorf(codes.InvalidArgument, "unable to process the request: %v", err)
		case appeal.ErrActionForbidden:
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		case appeal.ErrApprovalNameNotFound:
			return nil, status.Errorf(codes.NotFound, "appeal not found: %v", id)
		default:
			return nil, status.Errorf(codes.Internal, "failed to update approval: %v", err)
		}
	}

	appealProto, err := s.adapter.ToAppealProto(a)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse appeal: %v", err)
	}

	return &guardianv1beta1.UpdateApprovalResponse{
		Appeal: appealProto,
	}, nil
}

func (s *GRPCServer) CancelAppeal(ctx context.Context, req *guardianv1beta1.CancelAppealRequest) (*guardianv1beta1.CancelAppealResponse, error) {
	id := req.GetId()
	a, err := s.appealService.Cancel(ctx, id)
	if err != nil {
		switch err {
		case appeal.ErrAppealNotFound:
			return nil, status.Errorf(codes.NotFound, "appeal not found: %v", id)
		case appeal.ErrAppealStatusCanceled,
			appeal.ErrAppealStatusApproved,
			appeal.ErrAppealStatusRejected,
			appeal.ErrAppealStatusTerminated,
			appeal.ErrAppealStatusUnrecognized:
			return nil, status.Errorf(codes.InvalidArgument, "unable to process the request: %v", err)
		default:
			return nil, status.Errorf(codes.Internal, "failed to cancel appeal: %v", err)
		}
	}

	appealProto, err := s.adapter.ToAppealProto(a)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse appeal: %v", err)
	}

	return &guardianv1beta1.CancelAppealResponse{
		Appeal: appealProto,
	}, nil
}

func (s *GRPCServer) RevokeAppeal(ctx context.Context, req *guardianv1beta1.RevokeAppealRequest) (*guardianv1beta1.RevokeAppealResponse, error) {
	id := req.GetId()
	actor, err := s.getUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get metadata: actor")
	}
	reason := req.GetReason().GetReason()

	a, err := s.appealService.Revoke(ctx, id, actor, reason)
	if err != nil {
		switch err {
		case appeal.ErrAppealNotFound:
			return nil, status.Errorf(codes.NotFound, "appeal not found: %v", id)
		default:
			return nil, status.Errorf(codes.Internal, "failed to cancel appeal: %v", err)
		}
	}

	appealProto, err := s.adapter.ToAppealProto(a)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse appeal: %v", err)
	}

	return &guardianv1beta1.RevokeAppealResponse{
		Appeal: appealProto,
	}, nil
}

func (s *GRPCServer) listAppeals(ctx context.Context, filters *domain.ListAppealsFilter) ([]*guardianv1beta1.Appeal, error) {
	appeals, err := s.appealService.Find(ctx, filters)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get appeal list: %s", err)
	}

	appealProtos := []*guardianv1beta1.Appeal{}
	for _, a := range appeals {
		appealProto, err := s.adapter.ToAppealProto(a)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse appeal: %s", err)
		}
		appealProtos = append(appealProtos, appealProto)
	}

	return appealProtos, nil
}

func (s *GRPCServer) listApprovals(ctx context.Context, filters *domain.ListApprovalsFilter) ([]*guardianv1beta1.Approval, error) {
	approvals, err := s.approvalService.ListApprovals(ctx, filters)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get approval list: %s", err)
	}

	approvalProtos := []*guardianv1beta1.Approval{}
	for _, a := range approvals {
		approvalProto, err := s.adapter.ToApprovalProto(a)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse approval: %v: %s", a.ID, err)
		}
		approvalProtos = append(approvalProtos, approvalProto)
	}

	return approvalProtos, nil
}

func (s *GRPCServer) getUser(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("unable to retrieve metadata from context")
	}

	users := md.Get(s.authenticatedUserHeaderKey)
	if len(users) == 0 {
		return "", errors.New("user email not found")
	}

	currentUser := users[0]
	return currentUser, nil
}
