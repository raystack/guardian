package v1beta1

import (
	"context"
	"errors"

	guardianv1beta1 "github.com/goto/guardian/api/proto/gotocompany/guardian/v1beta1"
	"github.com/goto/guardian/core/grant"
	"github.com/goto/guardian/core/provider"
	"github.com/goto/guardian/domain"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GRPCServer) ListGrants(ctx context.Context, req *guardianv1beta1.ListGrantsRequest) (*guardianv1beta1.ListGrantsResponse, error) {
	filter := domain.ListGrantsFilter{
		Q:             req.GetQ(),
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
		Owner:         req.GetOwner(),
		OrderBy:       req.GetOrderBy(),
		Size:          int(req.GetSize()),
		Offset:        int(req.GetOffset()),
	}
	grants, total, err := s.listGrants(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &guardianv1beta1.ListGrantsResponse{
		Grants: grants,
		Total:  int32(total),
	}, nil
}

func (s *GRPCServer) ListUserGrants(ctx context.Context, req *guardianv1beta1.ListUserGrantsRequest) (*guardianv1beta1.ListUserGrantsResponse, error) {
	user, err := s.getUser(ctx)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get metadata: user")
	}

	filter := domain.ListGrantsFilter{
		Q:             req.GetQ(),
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
		Size:          int(req.GetSize()),
		Offset:        int(req.GetOffset()),
		Owner:         user,
	}
	grants, total, err := s.listGrants(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &guardianv1beta1.ListUserGrantsResponse{
		Grants: grants,
		Total:  int32(total),
	}, nil
}

func (s *GRPCServer) GetGrant(ctx context.Context, req *guardianv1beta1.GetGrantRequest) (*guardianv1beta1.GetGrantResponse, error) {
	a, err := s.grantService.GetByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, grant.ErrGrantNotFound) {
			return nil, status.Errorf(codes.NotFound, "grant %q not found: %v", req.GetId(), err)
		}
		return nil, s.internalError(ctx, "failed to get grant details: %v", err)
	}

	grantProto, err := s.adapter.ToGrantProto(a)
	if err != nil {
		return nil, s.internalError(ctx, "failed to parse grant: %v", err)
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
		return nil, s.internalError(ctx, "failed to revoke grant: %v", err)
	}

	grantProto, err := s.adapter.ToGrantProto(a)
	if err != nil {
		return nil, s.internalError(ctx, "failed to parse grant: %v", err)
	}

	return &guardianv1beta1.RevokeGrantResponse{
		Grant: grantProto,
	}, nil
}

func (s *GRPCServer) UpdateGrant(ctx context.Context, req *guardianv1beta1.UpdateGrantRequest) (*guardianv1beta1.UpdateGrantResponse, error) {
	g := &domain.Grant{
		ID:    req.GetId(),
		Owner: req.GetOwner(),
	}
	if err := s.grantService.Update(ctx, g); err != nil {
		switch {
		case errors.Is(err, grant.ErrGrantNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, grant.ErrEmptyOwner):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, s.internalError(ctx, "failed to update grant: %v", err)
		}
	}

	grantProto, err := s.adapter.ToGrantProto(g)
	if err != nil {
		return nil, s.internalError(ctx, "failed to parse grant: %v", err)
	}

	return &guardianv1beta1.UpdateGrantResponse{
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
		return nil, s.internalError(ctx, "failed to revoke grants in bulk")
	}

	var grantsProto []*guardianv1beta1.Grant
	for _, a := range grants {
		grantProto, err := s.adapter.ToGrantProto(a)
		if err != nil {
			return nil, s.internalError(ctx, "failed to parse grant: %v", err)
		}
		grantsProto = append(grantsProto, grantProto)
	}

	return &guardianv1beta1.RevokeGrantsResponse{
		Grants: grantsProto,
	}, nil
}

func (s *GRPCServer) listGrants(ctx context.Context, filter domain.ListGrantsFilter) ([]*guardianv1beta1.Grant, int64, error) {
	eg, ctx := errgroup.WithContext(ctx)
	var grants []domain.Grant
	var total int64

	eg.Go(func() error {
		grantRecords, err := s.grantService.List(ctx, filter)
		if err != nil {
			return s.internalError(ctx, "failed to get grant list: %s", err)
		}
		grants = grantRecords
		return nil
	})
	eg.Go(func() error {
		totalRecord, err := s.grantService.GetGrantsTotalCount(ctx, filter)
		if err != nil {
			return s.internalError(ctx, "failed to get grant total count: %s", err)
		}
		total = totalRecord
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, 0, err
	}

	var grantProtos []*guardianv1beta1.Grant
	for i, a := range grants {
		grantProto, err := s.adapter.ToGrantProto(&grants[i])
		if err != nil {
			return nil, 0, s.internalError(ctx, "failed to parse grant %q: %v", a.ID, err)
		}
		grantProtos = append(grantProtos, grantProto)
	}

	return grantProtos, total, nil
}

func (s *GRPCServer) ImportGrantsFromProvider(ctx context.Context, req *guardianv1beta1.ImportGrantsFromProviderRequest) (*guardianv1beta1.ImportGrantsFromProviderResponse, error) {
	grants, err := s.grantService.ImportFromProvider(ctx, grant.ImportFromProviderCriteria{
		ProviderID:    req.GetProviderId(),
		ResourceIDs:   req.GetResourceIds(),
		ResourceTypes: req.GetResourceTypes(),
		ResourceURNs:  req.GetResourceUrns(),
	})
	if err != nil {
		switch {
		case errors.Is(err, provider.ErrRecordNotFound):
			return nil, status.Errorf(codes.NotFound, "provider with id %q not found: %v", req.GetProviderId(), err)
		case errors.Is(err, grant.ErrEmptyImportedGrants):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return nil, s.internalError(ctx, "failed to import access: %v", err)
	}

	grantsProto := []*guardianv1beta1.Grant{}
	for _, g := range grants {
		grantProto, err := s.adapter.ToGrantProto(g)
		if err != nil {
			return nil, s.internalError(ctx, "failed to parse appeal proto %q: %v", g.ID, err)
		}
		grantsProto = append(grantsProto, grantProto)
	}

	return &guardianv1beta1.ImportGrantsFromProviderResponse{
		Grants: grantsProto,
	}, nil
}

func (s *GRPCServer) ListUserRoles(ctx context.Context, req *guardianv1beta1.ListUserRolesRequest) (*guardianv1beta1.ListUserRolesResponse, error) {
	user, err := s.getUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get metadata: user")
	}

	roles, err := s.grantService.ListUserRoles(ctx, user)
	if err != nil {
		return nil, s.internalError(ctx, "Internal Error")
	}
	return &guardianv1beta1.ListUserRolesResponse{
		Roles: roles,
	}, nil
}
