package v1beta1

import (
	"context"

	guardianv1beta1 "github.com/raystack/guardian/api/proto/raystack/guardian/v1beta1"
	"github.com/raystack/guardian/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *GRPCServer) CreateNamespace(ctx context.Context, req *guardianv1beta1.CreateNamespaceRequest) (*guardianv1beta1.CreateNamespaceResponse, error) {
	metadata := map[string]any{}
	if req.GetNamespace().GetMetadata() != nil {
		metadata = req.GetNamespace().GetMetadata().AsMap()
	}
	if err := s.namespaceService.Create(ctx, &domain.Namespace{
		ID:       req.GetNamespace().GetId(),
		Name:     req.GetNamespace().GetName(),
		State:    req.GetNamespace().GetState(),
		Metadata: metadata,
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create namespace: %v", err)
	}
	return &guardianv1beta1.CreateNamespaceResponse{}, nil
}

func (s *GRPCServer) UpdateNamespace(ctx context.Context, req *guardianv1beta1.UpdateNamespaceRequest) (*guardianv1beta1.UpdateNamespaceResponse, error) {
	metadata := map[string]any{}
	if req.GetNamespace().GetMetadata() != nil {
		metadata = req.GetNamespace().GetMetadata().AsMap()
	}
	if err := s.namespaceService.Update(ctx, &domain.Namespace{
		Name:     req.GetNamespace().GetName(),
		State:    req.GetNamespace().GetState(),
		Metadata: metadata,
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update namespace: %v", err)
	}
	return &guardianv1beta1.UpdateNamespaceResponse{}, nil
}

func (s *GRPCServer) GetNamespace(ctx context.Context, req *guardianv1beta1.GetNamespaceRequest) (*guardianv1beta1.GetNamespaceResponse, error) {
	ns, err := s.namespaceService.Get(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get namespace: %v", err)
	}
	md, err := structpb.NewStruct(ns.Metadata)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get namespace: %v", err)
	}
	return &guardianv1beta1.GetNamespaceResponse{
		Namespace: &guardianv1beta1.Namespace{
			Id:       ns.ID,
			Name:     ns.Name,
			State:    ns.State,
			Metadata: md,
		},
	}, nil
}

func (s *GRPCServer) ListNamespaces(ctx context.Context, req *guardianv1beta1.ListNamespacesRequest) (*guardianv1beta1.ListNamespacesResponse, error) {
	nss, err := s.namespaceService.List(ctx, domain.NamespaceFilter{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list namespaces: %v", err)
	}
	var namespaces []*guardianv1beta1.Namespace
	for _, ns := range nss {
		md, err := structpb.NewStruct(ns.Metadata)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to list namespaces: %v", err)
		}
		namespaces = append(namespaces, &guardianv1beta1.Namespace{
			Id:       ns.ID,
			Name:     ns.Name,
			State:    ns.State,
			Metadata: md,
		})
	}
	return &guardianv1beta1.ListNamespacesResponse{
		Namespaces: namespaces,
	}, nil
}
