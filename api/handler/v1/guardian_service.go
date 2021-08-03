package v1

import (
	"context"
	"time"

	pb "github.com/odpf/guardian/api/proto/guardian"
	"github.com/odpf/guardian/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProtoAdapter interface {
	FromProviderProto(*pb.Provider) (*domain.Provider, error)
	FromProviderConfigProto(*pb.ProviderConfig) (*domain.ProviderConfig, error)
	ToProviderProto(*domain.Provider) (*pb.Provider, error)

	FromPolicyProto(*pb.Policy) (*domain.Policy, error)
	ToPolicyProto(*domain.Policy) (*pb.Policy, error)
}

type GuardianServiceServer struct {
	resourceService domain.ResourceService
	providerService domain.ProviderService
	policyService   domain.PolicyService
	appealService   domain.AppealService
	adapter         ProtoAdapter

	Now func() time.Time

	pb.UnimplementedGuardianServiceServer
}

func NewGuardianServiceServer(
	resourceService domain.ResourceService,
	providerService domain.ProviderService,
	policyService domain.PolicyService,
	appealService domain.AppealService,
	adapter ProtoAdapter,
) *GuardianServiceServer {
	return &GuardianServiceServer{
		resourceService: resourceService,
		providerService: providerService,
		policyService:   policyService,
		appealService:   appealService,
		adapter:         adapter,
	}
}

func (s *GuardianServiceServer) ListProviders(ctx context.Context, req *pb.ListProvidersRequest) (*pb.ListProvidersResponse, error) {
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

func (s *GuardianServiceServer) CreateProvider(ctx context.Context, req *pb.CreateProviderRequest) (*pb.CreateProviderResponse, error) {
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

func (s *GuardianServiceServer) UpdateProvider(ctx context.Context, req *pb.UpdateProviderRequest) (*pb.UpdateProviderResponse, error) {
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

func (s *GuardianServiceServer) ListPolicies(ctx context.Context, req *pb.ListPoliciesRequest) (*pb.ListPoliciesResponse, error) {
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

func (s *GuardianServiceServer) CreatePolicy(ctx context.Context, req *pb.CreatePolicyRequest) (*pb.CreatePolicyResponse, error) {
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

func (s *GuardianServiceServer) UpdatePolicy(ctx context.Context, req *pb.UpdatePolicyRequest) (*pb.UpdatePolicyResponse, error) {
	policy, err := s.adapter.FromPolicyProto(req.GetPolicy())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s: cannot deserialize policy", err)
	}

	policy.ID = req.GetId()
	if err := s.policyService.Update(policy); err != nil {
		return nil, status.Errorf(codes.Internal, "%s: failed to update policy", err)
	}

	policyProto, err := s.adapter.ToPolicyProto(policy)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s: failed to parse policy", err)
	}

	return &pb.UpdatePolicyResponse{
		Policy: policyProto,
	}, nil
}
