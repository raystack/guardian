package v1beta1

import (
	"context"
	"errors"

	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/core/grant"
	"github.com/odpf/guardian/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GRPCServer) ListGrants(ctx context.Context, req *guardianv1beta1.ListGrantsRequest) (*guardianv1beta1.ListGrantsResponse, error) {
	filter := domain.ListGrantsFilter{
		Statuses:      req.GetStatuses(),
		AccountIDs:    req.GetAccountIds(),
		AccountTypes:  req.GetAccountTypes(),
		ResourceIDs:   req.GetResourceIds(),
		Roles:         req.GetRoles(),
		ProviderTypes: req.GetProviderTypes(),
		ProviderURNs:  req.GetProviderUrns(),
		ResourceTypes: req.GetResourceTypes(),
		ResourceURNs:  req.GetResourceUrns(),
		CreatedBy:     req.GetCreatedBy(),
		OrderBy:       req.GetOrderBy(),
	}
	grants, err := s.listGrants(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &guardianv1beta1.ListGrantsResponse{
		Grants: grants,
	}, nil
}

func (s *GRPCServer) ListUserGrants(ctx context.Context, req *guardianv1beta1.ListUserGrantsRequest) (*guardianv1beta1.ListUserGrantsResponse, error) {
	user, err := s.getUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get metadata: user")
	}

	filter := domain.ListGrantsFilter{
		Statuses:      req.GetStatuses(),
		AccountIDs:    req.GetAccountIds(),
		AccountTypes:  req.GetAccountTypes(),
		ResourceIDs:   req.GetResourceIds(),
		Roles:         req.GetRoles(),
		ProviderTypes: req.GetProviderTypes(),
		ProviderURNs:  req.GetProviderUrns(),
		ResourceTypes: req.GetResourceTypes(),
		ResourceURNs:  req.GetResourceUrns(),
		OrderBy:       req.GetOrderBy(),
		CreatedBy:     user,
	}
	grants, err := s.listGrants(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &guardianv1beta1.ListUserGrantsResponse{
		Grants: grants,
	}, nil
}

func (s *GRPCServer) GetGrant(ctx context.Context, req *guardianv1beta1.GetGrantRequest) (*guardianv1beta1.GetGrantResponse, error) {
	a, err := s.grantService.GetByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, grant.ErrGrantNotFound) {
			return nil, status.Errorf(codes.NotFound, "grant %q not found: %v", req.GetId(), err)
		}
		return nil, status.Errorf(codes.Internal, "failed to get grant details: %v", err)
	}

	grantProto, err := s.adapter.ToGrantProto(a)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse grant: %v", err)
	}

	return &guardianv1beta1.GetGrantResponse{
		Grant: grantProto,
	}, nil
}

func (s *GRPCServer) RevokeGrant(ctx context.Context, req *guardianv1beta1.RevokeGrantRequest) (*guardianv1beta1.RevokeGrantResponse, error) {
	actor, err := s.getUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get metadata: actor")
	}

	a, err := s.grantService.Revoke(ctx, req.GetId(), actor, req.GetReason())
	if err != nil {
		if errors.Is(err, grant.ErrGrantNotFound) {
			return nil, status.Error(codes.NotFound, "grant not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to revoke grant: %v", err)
	}

	grantProto, err := s.adapter.ToGrantProto(a)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse grant: %v", err)
	}

	return &guardianv1beta1.RevokeGrantResponse{
		Grant: grantProto,
	}, nil
}

func (s *GRPCServer) RevokeGrants(ctx context.Context, req *guardianv1beta1.RevokeGrantsRequest) (*guardianv1beta1.RevokeGrantsResponse, error) {
	actor, err := s.getUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get metadata: actor")
	}

	filter := domain.RevokeGrantsFilter{
		AccountIDs:    req.GetAccountIds(),
		ProviderTypes: req.GetProviderTypes(),
		ProviderURNs:  req.GetProviderUrns(),
		ResourceTypes: req.GetResourceTypes(),
		ResourceURNs:  req.GetResourceUrns(),
	}
	grants, err := s.grantService.BulkRevoke(ctx, filter, actor, req.GetReason())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to revoke grants in bulk")
	}

	var grantsProto []*guardianv1beta1.Grant
	for _, a := range grants {
		grantProto, err := s.adapter.ToGrantProto(a)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse grant: %v", err)
		}
		grantsProto = append(grantsProto, grantProto)
	}

	return &guardianv1beta1.RevokeGrantsResponse{
		Grants: grantsProto,
	}, nil
}

func (s *GRPCServer) listGrants(ctx context.Context, filter domain.ListGrantsFilter) ([]*guardianv1beta1.Grant, error) {
	grants, err := s.grantService.List(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list grants: %v", err)
	}

	var grantProtos []*guardianv1beta1.Grant
	for i, a := range grants {
		grantProto, err := s.adapter.ToGrantProto(&grants[i])
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse grant %q: %v", a.ID, err)
		}
		grantProtos = append(grantProtos, grantProto)
	}

	return grantProtos, nil
}
