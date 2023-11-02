package v1beta1

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/sync/errgroup"

	guardianv1beta1 "github.com/goto/guardian/api/proto/gotocompany/guardian/v1beta1"
	"github.com/goto/guardian/core/resource"
	"github.com/goto/guardian/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GRPCServer) ListResources(ctx context.Context, req *guardianv1beta1.ListResourcesRequest) (*guardianv1beta1.ListResourcesResponse, error) {
	var details map[string]string
	if req.GetDetails() != nil {
		details = map[string]string{}
		for _, d := range req.GetDetails() {
			filter := strings.Split(d, ":")
			if len(filter) == 2 {
				path := filter[0]
				value := filter[1]
				details[path] = value
			}
		}
	}
	filter := domain.ListResourcesFilter{
		IsDeleted:    req.GetIsDeleted(),
		ResourceType: req.GetType(),
		ResourceURN:  req.GetUrn(),
		ProviderType: req.GetProviderType(),
		ProviderURN:  req.GetProviderUrn(),
		Name:         req.GetName(),
		Details:      details,
		Size:         req.GetSize(),
		Offset:       req.GetOffset(),
	}

	resources, total, err := s.listResources(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &guardianv1beta1.ListResourcesResponse{
		Resources: resources,
		Total:     uint32(total),
	}, nil
}

func (s *GRPCServer) listResources(ctx context.Context, filter domain.ListResourcesFilter) ([]*guardianv1beta1.Resource, int64, error) {
	eg, ctx := errgroup.WithContext(ctx)
	var resources []*domain.Resource
	var total int64

	eg.Go(func() error {
		resourceRecords, err := s.resourceService.Find(ctx, filter)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to get resource list: %s", err)
		}
		resources = resourceRecords
		return nil
	})
	eg.Go(func() error {
		totalRecord, err := s.resourceService.GetResourcesTotalCount(ctx, filter)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to get resource total count: %s", err)
		}
		total = totalRecord
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, 0, err
	}
	var resourceProtos []*guardianv1beta1.Resource
	for i, r := range resources {
		resourceProto, err := s.adapter.ToResourceProto(resources[i])
		if err != nil {
			return nil, 0, status.Errorf(codes.Internal, "failed to parse resource %v: %v", r.Name, err)
		}
		resourceProtos = append(resourceProtos, resourceProto)
	}

	return resourceProtos, total, nil
}

func (s *GRPCServer) GetResource(ctx context.Context, req *guardianv1beta1.GetResourceRequest) (*guardianv1beta1.GetResourceResponse, error) {
	r, err := s.resourceService.GetOne(ctx, req.GetId())
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

	return &guardianv1beta1.GetResourceResponse{
		Resource: resourceProto,
	}, nil
}

func (s *GRPCServer) UpdateResource(ctx context.Context, req *guardianv1beta1.UpdateResourceRequest) (*guardianv1beta1.UpdateResourceResponse, error) {
	r := s.adapter.FromResourceProto(req.GetResource())
	r.ID = req.GetId()

	if err := s.resourceService.Update(ctx, r); err != nil {
		if errors.Is(err, resource.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "resource not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update resource: %v", err)
	}

	resourceProto, err := s.adapter.ToResourceProto(r)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse resource: %v", err)
	}

	return &guardianv1beta1.UpdateResourceResponse{
		Resource: resourceProto,
	}, nil
}

func (s *GRPCServer) DeleteResource(ctx context.Context, req *guardianv1beta1.DeleteResourceRequest) (*guardianv1beta1.DeleteResourceResponse, error) {
	if err := s.resourceService.Delete(ctx, req.GetId()); err != nil {
		if errors.Is(err, resource.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "resource not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update resource: %v", err)
	}

	return &guardianv1beta1.DeleteResourceResponse{}, nil
}
