package v1

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/odpf/guardian/api/proto/odpf/guardian"
	"github.com/odpf/guardian/appeal"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/policy"
	"github.com/odpf/guardian/provider"
	"github.com/odpf/guardian/resource"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ProtoAdapter interface {
	FromProviderProto(*pb.Provider) (*domain.Provider, error)
	FromProviderConfigProto(*pb.ProviderConfig) (*domain.ProviderConfig, error)
	ToProviderProto(*domain.Provider) (*pb.Provider, error)
	ToProviderConfigProto(*domain.ProviderConfig) (*pb.ProviderConfig, error)
	ToRole(*domain.Role) (*pb.Role, error)

	FromPolicyProto(*pb.Policy) (*domain.Policy, error)
	ToPolicyProto(*domain.Policy) (*pb.Policy, error)

	FromResourceProto(*pb.Resource) *domain.Resource
	ToResourceProto(*domain.Resource) (*pb.Resource, error)

	FromAppealProto(*pb.Appeal) (*domain.Appeal, error)
	ToAppealProto(*domain.Appeal) (*pb.Appeal, error)
	FromCreateAppealProto(*pb.CreateAppealRequest) ([]*domain.Appeal, error)
	ToApprovalProto(*domain.Approval) (*pb.Approval, error)
}

type GRPCServer struct {
	resourceService domain.ResourceService
	providerService domain.ProviderService
	policyService   domain.PolicyService
	appealService   domain.AppealService
	approvalService domain.ApprovalService
	adapter         ProtoAdapter

	authenticatedUserHeaderKey string

	pb.UnimplementedGuardianServiceServer
}

func NewGRPCServer(
	resourceService domain.ResourceService,
	providerService domain.ProviderService,
	policyService domain.PolicyService,
	appealService domain.AppealService,
	approvalService domain.ApprovalService,
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

func (s *GRPCServer) ListProviders(ctx context.Context, req *pb.ListProvidersRequest) (*pb.ListProvidersResponse, error) {
	providers, err := s.providerService.Find()
	if err != nil {
		return nil, err
	}

	providerProtos := []*pb.Provider{}
	for _, p := range providers {
		p.Config.Credentials = nil
		providerProto, err := s.adapter.ToProviderProto(p)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse provider %s: %v", p.URN, err)
		}
		providerProtos = append(providerProtos, providerProto)
	}

	return &pb.ListProvidersResponse{
		Providers: providerProtos,
	}, nil
}

func (s *GRPCServer) GetProvider(ctx context.Context, req *pb.GetProviderRequest) (*pb.Provider, error) {
	p, err := s.providerService.GetByID(uint(req.GetId()))
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

	return providerProto, nil
}

func (s *GRPCServer) CreateProvider(ctx context.Context, req *pb.CreateProviderRequest) (*pb.CreateProviderResponse, error) {
	providerConfig, err := s.adapter.FromProviderConfigProto(req.GetConfig())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot deserialize provider config: %v", err)
	}

	p := &domain.Provider{
		Type:   providerConfig.Type,
		URN:    providerConfig.URN,
		Config: providerConfig,
	}
	if err := s.providerService.Create(p); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create provider: %v", err)
	}

	providerProto, err := s.adapter.ToProviderProto(p)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse provider: %v", err)
	}

	return &pb.CreateProviderResponse{
		Provider: providerProto,
	}, nil
}

func (s *GRPCServer) UpdateProvider(ctx context.Context, req *pb.UpdateProviderRequest) (*pb.UpdateProviderResponse, error) {
	id := req.GetId()
	providerConfig, err := s.adapter.FromProviderConfigProto(req.GetConfig())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot deserialize provider config: %v", err)
	}

	p := &domain.Provider{
		ID:     uint(id),
		Type:   providerConfig.Type,
		URN:    providerConfig.URN,
		Config: providerConfig,
	}
	if err := s.providerService.Update(p); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update provider: %v", err)
	}

	providerProto, err := s.adapter.ToProviderProto(p)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse provider: %v", err)
	}

	return &pb.UpdateProviderResponse{
		Provider: providerProto,
	}, nil
}

func (s *GRPCServer) ListRoles(ctx context.Context, req *pb.ListRolesRequest) (*pb.ListRolesResponse, error) {
	roles, err := s.providerService.GetRoles(uint(req.GetId()), req.GetResourceType())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list roles: %v", err)
	}

	roleProtos := []*pb.Role{}
	for _, r := range roles {
		role, err := s.adapter.ToRole(r)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse proto: %v", err)
		}

		roleProtos = append(roleProtos, role)
	}

	return &pb.ListRolesResponse{
		Roles: roleProtos,
	}, nil
}

func (s *GRPCServer) ListPolicies(ctx context.Context, req *pb.ListPoliciesRequest) (*pb.ListPoliciesResponse, error) {
	policies, err := s.policyService.Find()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get policy list: %v", err)
	}

	policyProtos := []*pb.Policy{}
	for _, p := range policies {
		policyProto, err := s.adapter.ToPolicyProto(p)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse policy %v: %v", p.ID, err)
		}
		policyProtos = append(policyProtos, policyProto)
	}

	return &pb.ListPoliciesResponse{
		Policies: policyProtos,
	}, nil
}

func (s *GRPCServer) GetPolicy(ctx context.Context, req *pb.GetPolicyRequest) (*pb.Policy, error) {
	p, err := s.policyService.GetOne(req.GetId(), uint(req.GetVersion()))
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

	return policyProto, nil
}

func (s *GRPCServer) CreatePolicy(ctx context.Context, req *pb.CreatePolicyRequest) (*pb.CreatePolicyResponse, error) {
	policy, err := s.adapter.FromPolicyProto(req.GetPolicy())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot deserialize policy: %v", err)
	}

	if err := s.policyService.Create(policy); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create policy: %v", err)
	}

	policyProto, err := s.adapter.ToPolicyProto(policy)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse policy: %v", err)
	}

	return &pb.CreatePolicyResponse{
		Policy: policyProto,
	}, nil
}

func (s *GRPCServer) UpdatePolicy(ctx context.Context, req *pb.UpdatePolicyRequest) (*pb.UpdatePolicyResponse, error) {
	p, err := s.adapter.FromPolicyProto(req.GetPolicy())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot deserialize policy: %v", err)
	}

	p.ID = req.GetId()
	if err := s.policyService.Update(p); err != nil {
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

	return &pb.UpdatePolicyResponse{
		Policy: policyProto,
	}, nil
}

func (s *GRPCServer) ListResources(ctx context.Context, req *pb.ListResourcesRequest) (*pb.ListResourcesResponse, error) {
	var details map[string]string
	fmt.Printf("req.GetDetailsPaths(): %v\n", req.GetDetailsPaths())
	fmt.Printf("req.GetDetailsValues(): %v\n", req.GetDetailsValues())
	if len(req.GetDetailsPaths()) == len(req.GetDetailsValues()) {
		details = map[string]string{}
		paths := req.GetDetailsPaths()
		for i, v := range req.GetDetailsValues() {
			details[paths[i]] = v
		}
	}
	resources, err := s.resourceService.Find(map[string]interface{}{
		"is_deleted":    req.GetIsDeleted(),
		"type":          req.GetType(),
		"urn":           req.GetUrn(),
		"provider_type": req.GetProviderType(),
		"provider_urn":  req.GetProviderUrn(),
		"name":          req.GetName(),
		"details":       details,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get resource list: %v", err)
	}

	resourceProtos := []*pb.Resource{}
	for _, r := range resources {
		resourceProto, err := s.adapter.ToResourceProto(r)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse resource %v: %v", r.Name, err)
		}
		resourceProtos = append(resourceProtos, resourceProto)
	}

	return &pb.ListResourcesResponse{
		Resources: resourceProtos,
	}, nil
}

func (s *GRPCServer) GetResource(ctx context.Context, req *pb.GetResourceRequest) (*pb.Resource, error) {
	r, err := s.resourceService.GetOne(uint(req.GetId()))
	if err != nil {
		switch err {
		case resource.ErrRecordNotFound:
			return nil, status.Error(codes.NotFound, "resource not found")
		default:
			return nil, status.Errorf(codes.Internal, "failed to retrieve resource: %v", err)
		}
	}

	resourceProto, err := s.adapter.ToResourceProto(r)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse resource: %v", err)
	}

	return resourceProto, nil
}

func (s *GRPCServer) UpdateResource(ctx context.Context, req *pb.UpdateResourceRequest) (*pb.UpdateResourceResponse, error) {
	r := s.adapter.FromResourceProto(req.GetResource())
	r.ID = uint(req.GetId())

	if err := s.resourceService.Update(r); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update resource: %v", err)
	}

	resourceProto, err := s.adapter.ToResourceProto(r)
	if err != nil {
		if errors.Is(err, resource.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "resource not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to parse resource: %v", err)
	}

	return &pb.UpdateResourceResponse{
		Resource: resourceProto,
	}, nil
}

func (s *GRPCServer) ListUserAppeals(ctx context.Context, req *pb.ListUserAppealsRequest) (*pb.ListUserAppealsResponse, error) {
	user, err := s.getUser(ctx)
	if err != nil {
		return nil, err
	}

	filters := map[string]interface{}{
		"user": user,
	}
	appeals, err := s.listAppeals(filters)
	if err != nil {
		return nil, err
	}

	return &pb.ListUserAppealsResponse{
		Appeals: appeals,
	}, nil
}

func (s *GRPCServer) ListAppeals(ctx context.Context, req *pb.ListAppealsRequest) (*pb.ListAppealsResponse, error) {
	filters := map[string]interface{}{}
	if req.GetUser() != "" {
		filters["user"] = req.GetUser()
	}
	appeals, err := s.listAppeals(filters)
	if err != nil {
		return nil, err
	}

	return &pb.ListAppealsResponse{
		Appeals: appeals,
	}, nil
}

func (s *GRPCServer) CreateAppeal(ctx context.Context, req *pb.CreateAppealRequest) (*pb.CreateAppealResponse, error) {
	appeals, err := s.adapter.FromCreateAppealProto(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot deserialize payload: %v", err)
	}

	if err := s.appealService.Create(appeals); err != nil {
		if errors.Is(err, appeal.ErrAppealDuplicate) {
			return nil, status.Errorf(codes.AlreadyExists, "appeal already exists: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to create appeal: %v", err)
	}

	appealProtos := []*pb.Appeal{}
	for _, appeal := range appeals {
		appealProto, err := s.adapter.ToAppealProto(appeal)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse appeal: %v", err)
		}
		appealProtos = append(appealProtos, appealProto)
	}

	return &pb.CreateAppealResponse{
		Appeals: appealProtos,
	}, nil
}

func (s *GRPCServer) ListUserApprovals(ctx context.Context, req *pb.ListUserApprovalsRequest) (*pb.ListUserApprovalsResponse, error) {
	user, err := s.getUser(ctx)
	if err != nil {
		return nil, err
	}

	approvals, err := s.listApprovals(&domain.ListApprovalsFilter{
		User:     user,
		Statuses: req.GetStatuses(),
	})
	if err != nil {
		return nil, err
	}

	return &pb.ListUserApprovalsResponse{
		Approvals: approvals,
	}, nil
}

func (s *GRPCServer) ListApprovals(ctx context.Context, req *pb.ListApprovalsRequest) (*pb.ListApprovalsResponse, error) {
	approvals, err := s.listApprovals(&domain.ListApprovalsFilter{
		User:     req.GetUser(),
		Statuses: req.GetStatuses(),
	})
	if err != nil {
		return nil, err
	}

	return &pb.ListApprovalsResponse{
		Approvals: approvals,
	}, nil
}

func (s *GRPCServer) GetAppeal(ctx context.Context, req *pb.GetAppealRequest) (*pb.GetAppealResponse, error) {
	id := req.GetId()
	appeal, err := s.appealService.GetByID(uint(id))
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

	return &pb.GetAppealResponse{
		Appeal: appealProto,
	}, nil
}

func (s *GRPCServer) UpdateApproval(ctx context.Context, req *pb.UpdateApprovalRequest) (*pb.UpdateApprovalResponse, error) {
	actor, err := s.getUser(ctx)
	if err != nil {
		return nil, err
	}

	id := req.GetId()
	a, err := s.appealService.MakeAction(domain.ApprovalAction{
		AppealID:     uint(id),
		ApprovalName: req.GetApprovalName(),
		Actor:        actor,
		Action:       req.GetAction().GetAction(),
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

	return &pb.UpdateApprovalResponse{
		Appeal: appealProto,
	}, nil
}

func (s *GRPCServer) CancelAppeal(ctx context.Context, req *pb.CancelAppealRequest) (*pb.CancelAppealResponse, error) {
	id := req.GetId()
	a, err := s.appealService.Cancel(uint(id))
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

	return &pb.CancelAppealResponse{
		Appeal: appealProto,
	}, nil
}

func (s *GRPCServer) RevokeAppeal(ctx context.Context, req *pb.RevokeAppealRequest) (*pb.RevokeAppealResponse, error) {
	id := req.GetId()
	actor, err := s.getUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get metadata: actor")
	}
	reason := req.GetReason().GetReason()

	a, err := s.appealService.Revoke(uint(id), actor, reason)
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

	return &pb.RevokeAppealResponse{
		Appeal: appealProto,
	}, nil
}

func (s *GRPCServer) listAppeals(filters map[string]interface{}) ([]*pb.Appeal, error) {
	appeals, err := s.appealService.Find(filters)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get appeal list: %s", err)
	}

	appealProtos := []*pb.Appeal{}
	for _, a := range appeals {
		appealProto, err := s.adapter.ToAppealProto(a)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse appeal: %s", err)
		}
		appealProtos = append(appealProtos, appealProto)
	}

	return appealProtos, nil
}

func (s *GRPCServer) listApprovals(filters *domain.ListApprovalsFilter) ([]*pb.Approval, error) {
	approvals, err := s.approvalService.ListApprovals(filters)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get approval list: %s", err)
	}

	approvalProtos := []*pb.Approval{}
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
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		users := md.Get(s.authenticatedUserHeaderKey)
		if len(users) > 0 {
			return users[0], nil
		}
	}

	return "", status.Error(codes.Unauthenticated, "user email not found")
}
