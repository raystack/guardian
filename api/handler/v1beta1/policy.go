package v1beta1

import (
	"context"
	"errors"

	guardianv1beta1 "github.com/raystack/guardian/api/proto/raystack/guardian/v1beta1"
	"github.com/raystack/guardian/core/policy"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
	if req.GetDryRun() {
		ctx = policy.WithDryRun(ctx)
	}

	authenticatedUser, err := s.getUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	p := s.adapter.FromPolicyProto(req.GetPolicy(), authenticatedUser)

	if err := s.policyService.Create(ctx, p); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create policy: %v", err)
	}

	policyProto, err := s.adapter.ToPolicyProto(p)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse policy: %v", err)
	}

	return &guardianv1beta1.CreatePolicyResponse{
		Policy: policyProto,
	}, nil
}

func (s *GRPCServer) UpdatePolicy(ctx context.Context, req *guardianv1beta1.UpdatePolicyRequest) (*guardianv1beta1.UpdatePolicyResponse, error) {
	if req.GetDryRun() {
		ctx = policy.WithDryRun(ctx)
	}
	authenticatedUser, err := s.getUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	p := s.adapter.FromPolicyProto(req.GetPolicy(), authenticatedUser)

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

func (s *GRPCServer) GetPolicyPreferences(ctx context.Context, req *guardianv1beta1.GetPolicyPreferencesRequest) (*guardianv1beta1.GetPolicyPreferencesResponse, error) {
	p, err := s.policyService.GetOne(ctx, req.GetId(), uint(req.GetVersion()))
	if err != nil {
		switch err {
		case policy.ErrPolicyNotFound:
			return nil, status.Error(codes.NotFound, "policy not found")
		default:
			return nil, status.Errorf(codes.Internal, "failed to retrieve policy: %v", err)
		}
	}

	appealConfigProto := s.adapter.ToPolicyAppealConfigProto(p)

	return &guardianv1beta1.GetPolicyPreferencesResponse{
		Appeal: appealConfigProto,
	}, nil
}
