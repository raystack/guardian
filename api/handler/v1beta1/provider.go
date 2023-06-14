package v1beta1

import (
	"context"
	"errors"

	guardianv1beta1 "github.com/raystack/guardian/api/proto/raystack/guardian/v1beta1"
	"github.com/raystack/guardian/core/provider"
	"github.com/raystack/guardian/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GRPCServer) ListProviders(ctx context.Context, req *guardianv1beta1.ListProvidersRequest) (*guardianv1beta1.ListProvidersResponse, error) {
	providers, err := s.providerService.Find(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list providers: %v", err)
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
		providerTypeProtos = append(providerTypeProtos, s.adapter.ToProviderTypeProto(pt))
	}

	return &guardianv1beta1.GetProviderTypesResponse{
		ProviderTypes: providerTypeProtos,
	}, nil
}

func (s *GRPCServer) CreateProvider(ctx context.Context, req *guardianv1beta1.CreateProviderRequest) (*guardianv1beta1.CreateProviderResponse, error) {
	if req.GetDryRun() {
		ctx = provider.WithDryRun(ctx)
	}

	providerConfig := s.adapter.FromProviderConfigProto(req.GetConfig())
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
	if req.GetDryRun() {
		ctx = provider.WithDryRun(ctx)
	}

	id := req.GetId()
	providerConfig := s.adapter.FromProviderConfigProto(req.GetConfig())
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
