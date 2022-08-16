package v1beta1

import (
	"context"
	"errors"

	guardianv1beta1 "github.com/odpf/guardian/api/proto/odpf/guardian/v1beta1"
	"github.com/odpf/guardian/core/access"
	"github.com/odpf/guardian/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GRPCServer) ListAccesses(ctx context.Context, req *guardianv1beta1.ListAccessesRequest) (*guardianv1beta1.ListAccessesResponse, error) {
	filter := domain.ListAccessesFilter{
		Statuses:     req.GetStatuses(),
		AccountIDs:   req.GetAccountIds(),
		AccountTypes: req.GetAccountTypes(),
		ResourceIDs:  req.GetResourceIds(),
	}
	accesses, err := s.accessService.List(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list access: %v", err)
	}

	var accessProtos []*guardianv1beta1.Access
	for _, a := range accesses {
		accessProto, err := s.adapter.ToAccessProto(a)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse access %q: %v", a.ID, err)
		}
		accessProtos = append(accessProtos, accessProto)
	}

	return &guardianv1beta1.ListAccessesResponse{
		Accesses: accessProtos,
	}, nil
}

func (s *GRPCServer) GetAccess(ctx context.Context, req *guardianv1beta1.GetAccessRequest) (*guardianv1beta1.GetAccessResponse, error) {
	a, err := s.accessService.GetByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, access.ErrAccessNotFound) {
			return nil, status.Errorf(codes.NotFound, "access %q not found: %v", req.GetId(), err)
		}
		return nil, status.Errorf(codes.Internal, "failed to get access details: %v", err)
	}

	accessProto, err := s.adapter.ToAccessProto(*a)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse access: %v", err)
	}

	return &guardianv1beta1.GetAccessResponse{
		Access: accessProto,
	}, nil
}

func (s *GRPCServer) RevokeAccess(ctx context.Context, req *guardianv1beta1.RevokeAccessRequest) (*guardianv1beta1.RevokeAccessResponse, error) {
	actor, err := s.getUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get metadata: actor")
	}

	a, err := s.accessService.Revoke(ctx, req.GetId(), actor, req.GetReason())
	if err != nil {
		if errors.Is(err, access.ErrAccessNotFound) {
			return nil, status.Error(codes.NotFound, "access not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to revoke access: %v", err)
	}

	accessProto, err := s.adapter.ToAccessProto(*a)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse access: %v", err)
	}

	return &guardianv1beta1.RevokeAccessResponse{
		Access: accessProto,
	}, nil
}

func (s *GRPCServer) RevokeAccesses(ctx context.Context, req *guardianv1beta1.RevokeAccessesRequest) (*guardianv1beta1.RevokeAccessesResponse, error) {
	actor, err := s.getUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get metadata: actor")
	}

	filter := domain.RevokeAccessesFilter{
		AccountIDs:    req.GetAccountIds(),
		ProviderTypes: req.GetProviderTypes(),
		ProviderURNs:  req.GetProviderUrns(),
		ResourceTypes: req.GetResourceTypes(),
		ResourceURNs:  req.GetResourceUrns(),
	}
	accesses, err := s.accessService.BulkRevoke(ctx, filter, actor, req.GetReason())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to revoke accesses in bulk")
	}

	var accessesProto []*guardianv1beta1.Access
	for _, a := range accesses {
		accessProto, err := s.adapter.ToAccessProto(*a)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse access: %v", err)
		}
		accessesProto = append(accessesProto, accessProto)
	}

	return &guardianv1beta1.RevokeAccessesResponse{
		Accesses: accessesProto,
	}, nil
}
