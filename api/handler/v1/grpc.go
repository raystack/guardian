package v1

import (
	"context"
	"errors"
	"time"

	pb "github.com/odpf/guardian/api/proto/odpf/guardian"
	"github.com/odpf/guardian/appeal"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/policy"
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
	adapter         ProtoAdapter

	Now func() time.Time

	pb.UnimplementedGuardianServiceServer
}

func NewGRPCServer(
	resourceService domain.ResourceService,
	providerService domain.ProviderService,
	policyService domain.PolicyService,
	appealService domain.AppealService,
	adapter ProtoAdapter,
) *GRPCServer {
	return &GRPCServer{
		resourceService: resourceService,
		providerService: providerService,
		policyService:   policyService,
		appealService:   appealService,
		adapter:         adapter,
	}
}

func (s *GRPCServer) ListProviders(ctx context.Context, req *pb.ListProvidersRequest) (*pb.ListProvidersResponse, error) {
	providers, err := s.providerService.Find()
	if err != nil {
		return nil, err
	}

	providerProtos := []*pb.Provider{}
	for _, p := range providers {
		providerProto, err := s.adapter.ToProviderProto(p)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "%s: failed to parse provider %s", err.Error(), p.URN)
		}
		providerProtos = append(providerProtos, providerProto)
	}

	return &pb.ListProvidersResponse{
		Providers: providerProtos,
	}, nil
}

func (s *GRPCServer) CreateProvider(ctx context.Context, req *pb.CreateProviderRequest) (*pb.CreateProviderResponse, error) {
	providerConfig, err := s.adapter.FromProviderConfigProto(req.GetConfig())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s: cannot deserialize provider config", err)
	}

	p := &domain.Provider{
		Type:   providerConfig.Type,
		URN:    providerConfig.URN,
		Config: providerConfig,
	}
	if err := s.providerService.Create(p); err != nil {
		return nil, status.Errorf(codes.Internal, "%s: failed to create provider", err)
	}

	providerProto, err := s.adapter.ToProviderProto(p)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s: failed to parse provider", err)
	}

	return &pb.CreateProviderResponse{
		Provider: providerProto,
	}, nil
}

func (s *GRPCServer) UpdateProvider(ctx context.Context, req *pb.UpdateProviderRequest) (*pb.UpdateProviderResponse, error) {
	id := req.GetId()
	providerConfig, err := s.adapter.FromProviderConfigProto(req.GetConfig())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s: cannot deserialize provider config", err)
	}

	p := &domain.Provider{
		ID:     uint(id),
		Type:   providerConfig.Type,
		URN:    providerConfig.URN,
		Config: providerConfig,
	}
	if err := s.providerService.Update(p); err != nil {
		return nil, status.Errorf(codes.Internal, "%s: failed to update provider", err)
	}

	providerProto, err := s.adapter.ToProviderProto(p)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s: failed to parse provider", err)
	}

	return &pb.UpdateProviderResponse{
		Provider: providerProto,
	}, nil
}

func (s *GRPCServer) ListPolicies(ctx context.Context, req *pb.ListPoliciesRequest) (*pb.ListPoliciesResponse, error) {
	policies, err := s.policyService.Find()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s: failed to get policy list", err)
	}

	policyProtos := []*pb.Policy{}
	for _, p := range policies {
		policyProto, err := s.adapter.ToPolicyProto(p)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "%s: failed to parse policy %s", err.Error(), p.ID)
		}
		policyProtos = append(policyProtos, policyProto)
	}

	return &pb.ListPoliciesResponse{
		Policies: policyProtos,
	}, nil
}

func (s *GRPCServer) CreatePolicy(ctx context.Context, req *pb.CreatePolicyRequest) (*pb.CreatePolicyResponse, error) {
	policy, err := s.adapter.FromPolicyProto(req.GetPolicy())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s: cannot deserialize policy", err)
	}

	if err := s.policyService.Create(policy); err != nil {
		return nil, status.Errorf(codes.Internal, "%s: failed to create policy", err)
	}

	policyProto, err := s.adapter.ToPolicyProto(policy)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s: failed to parse policy", err)
	}

	return &pb.CreatePolicyResponse{
		Policy: policyProto,
	}, nil
}

func (s *GRPCServer) UpdatePolicy(ctx context.Context, req *pb.UpdatePolicyRequest) (*pb.UpdatePolicyResponse, error) {
	p, err := s.adapter.FromPolicyProto(req.GetPolicy())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s: cannot deserialize policy", err)
	}

	p.ID = req.GetId()
	if err := s.policyService.Update(p); err != nil {
		if errors.Is(err, policy.ErrPolicyDoesNotExists) {
			return nil, status.Error(codes.NotFound, "policy id not found")
		} else if errors.Is(err, policy.ErrEmptyIDParam) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return nil, status.Errorf(codes.Internal, "%s: failed to update policy", err)
	}

	policyProto, err := s.adapter.ToPolicyProto(p)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s: failed to parse policy", err)
	}

	return &pb.UpdatePolicyResponse{
		Policy: policyProto,
	}, nil
}

func (s *GRPCServer) ListResources(ctx context.Context, req *pb.ListResourcesRequest) (*pb.ListResourcesResponse, error) {
	resources, err := s.resourceService.Find(map[string]interface{}{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s: failed to get resource list", err)
	}

	resourceProtos := []*pb.Resource{}
	for _, r := range resources {
		resourceProto, err := s.adapter.ToResourceProto(r)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "%s: failed to parse resource %s", err.Error(), r.Name)
		}
		resourceProtos = append(resourceProtos, resourceProto)
	}

	return &pb.ListResourcesResponse{
		Resources: resourceProtos,
	}, nil
}

func (s *GRPCServer) UpdateResource(ctx context.Context, req *pb.UpdateResourceRequest) (*pb.UpdateResourceResponse, error) {
	r := s.adapter.FromResourceProto(req.GetResource())
	r.ID = uint(req.GetId())

	if err := s.resourceService.Update(r); err != nil {
		return nil, status.Errorf(codes.Internal, "%s: failed to update resource", err)
	}

	resourceProto, err := s.adapter.ToResourceProto(r)
	if err != nil {
		if errors.Is(err, resource.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "resource not found")
		}
		return nil, status.Errorf(codes.Internal, "%s: failed to parse resource", err)
	}

	return &pb.UpdateResourceResponse{
		Resource: resourceProto,
	}, nil
}

func (s *GRPCServer) ListAppeals(ctx context.Context, req *pb.ListAppealsRequest) (*pb.ListAppealsResponse, error) {
	filters := map[string]interface{}{}
	if req.GetUser() != "" {
		filters["user"] = req.GetUser()
	}

	appeals, err := s.appealService.Find(filters)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s: failed to get appeal list", err)
	}

	appealProtos := []*pb.Appeal{}
	for _, a := range appeals {
		appealProto, err := s.adapter.ToAppealProto(a)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "%s: failed to parse appeal", err)
		}
		appealProtos = append(appealProtos, appealProto)
	}

	return &pb.ListAppealsResponse{
		Appeals: appealProtos,
	}, nil
}

func (s *GRPCServer) CreateAppeal(ctx context.Context, req *pb.CreateAppealRequest) (*pb.CreateAppealResponse, error) {
	appeals, err := s.adapter.FromCreateAppealProto(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s: cannot deserialize payload", err)
	}

	if err := s.appealService.Create(appeals); err != nil {
		if errors.Is(err, appeal.ErrAppealDuplicate) {
			return nil, status.Errorf(codes.AlreadyExists, "%s: appeal already exists", err)
		}
		return nil, status.Errorf(codes.Internal, "%s: failed to create appeal", err)
	}

	appealProtos := []*pb.Appeal{}
	for _, appeal := range appeals {
		appealProto, err := s.adapter.ToAppealProto(appeal)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "%s: failed to parse appeal", err)
		}
		appealProtos = append(appealProtos, appealProto)
	}

	return &pb.CreateAppealResponse{
		Appeals: appealProtos,
	}, nil
}

func (s *GRPCServer) ListApprovals(ctx context.Context, req *pb.ListApprovalsRequest) (*pb.ListApprovalsResponse, error) {
	approvals, err := s.appealService.GetPendingApprovals(req.GetUser())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s: failed to get approval list", err)
	}

	approvalProtos := []*pb.Approval{}
	for _, a := range approvals {
		approvalProto, err := s.adapter.ToApprovalProto(a)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "%s: failed to parse approval: %v", err, a.ID)
		}
		approvalProtos = append(approvalProtos, approvalProto)
	}

	return &pb.ListApprovalsResponse{
		Approvals: approvalProtos,
	}, nil
}

func (s *GRPCServer) GetAppeal(ctx context.Context, req *pb.GetAppealRequest) (*pb.GetAppealResponse, error) {
	id := req.GetId()
	appeal, err := s.appealService.GetByID(uint(id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s: failed to retrieve appeal", err)
	}
	if appeal == nil {
		return nil, status.Errorf(codes.NotFound, "appeal not found: %v", id)
	}

	appealProto, err := s.adapter.ToAppealProto(appeal)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s: failed to parse appeal", err)
	}

	return &pb.GetAppealResponse{
		Appeal: appealProto,
	}, nil
}

func (s *GRPCServer) UpdateApproval(ctx context.Context, req *pb.UpdateApprovalRequest) (*pb.UpdateApprovalResponse, error) {
	actor, err := s.getActor(ctx)
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
			return nil, status.Errorf(codes.InvalidArgument, "unable to process the request: %s", err)
		case appeal.ErrActionForbidden:
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		case appeal.ErrApprovalNameNotFound:
			return nil, status.Errorf(codes.NotFound, "appeal not found: %v", id)
		default:
			return nil, status.Errorf(codes.Internal, "%s: failed to update approval", err)
		}
	}

	appealProto, err := s.adapter.ToAppealProto(a)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s: failed to parse appeal", err)
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
		case appeal.ErrAppealStatusCanceled,
			appeal.ErrAppealStatusApproved,
			appeal.ErrAppealStatusRejected,
			appeal.ErrAppealStatusTerminated,
			appeal.ErrAppealStatusUnrecognized:
			return nil, status.Errorf(codes.InvalidArgument, "unable to process the request: %s", err)
		default:
			return nil, status.Errorf(codes.Internal, "%s: failed to cancel appeal", err)
		}
	}

	appealProto, err := s.adapter.ToAppealProto(a)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s: failed to parse appeal", err)
	}

	return &pb.CancelAppealResponse{
		Appeal: appealProto,
	}, nil
}

func (s *GRPCServer) RevokeAppeal(ctx context.Context, req *pb.RevokeAppealRequest) (*pb.RevokeAppealResponse, error) {
	id := req.GetId()
	actor, err := s.getActor(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get metadata: actor")
	}

	a, err := s.appealService.Revoke(uint(id), actor)
	if err != nil {
		switch err {
		case appeal.ErrRevokeAppealForbidden:
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		case appeal.ErrAppealNotFound:
			return nil, status.Errorf(codes.NotFound, "appeal not found: %v", id)
		default:
			return nil, status.Errorf(codes.Internal, "%s: failed to cancel appeal", err)
		}
	}

	appealProto, err := s.adapter.ToAppealProto(a)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s: failed to parse appeal", err)
	}

	return &pb.RevokeAppealResponse{
		Appeal: appealProto,
	}, nil
}

func (s *GRPCServer) getActor(ctx context.Context) (string, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if userEmail, ok := md["x-goog-authenticated-user-email"]; ok {
			return userEmail[0], nil
		}
	}

	return "", status.Error(codes.Internal, "failed to get request metadata")
}
