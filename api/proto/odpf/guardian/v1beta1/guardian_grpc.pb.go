// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package guardianv1beta1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// GuardianServiceClient is the client API for GuardianService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GuardianServiceClient interface {
	ListProviders(ctx context.Context, in *ListProvidersRequest, opts ...grpc.CallOption) (*ListProvidersResponse, error)
	GetProvider(ctx context.Context, in *GetProviderRequest, opts ...grpc.CallOption) (*GetProviderResponse, error)
	GetProviderTypes(ctx context.Context, in *GetProviderTypesRequest, opts ...grpc.CallOption) (*GetProviderTypesResponse, error)
	CreateProvider(ctx context.Context, in *CreateProviderRequest, opts ...grpc.CallOption) (*CreateProviderResponse, error)
	UpdateProvider(ctx context.Context, in *UpdateProviderRequest, opts ...grpc.CallOption) (*UpdateProviderResponse, error)
	ListRoles(ctx context.Context, in *ListRolesRequest, opts ...grpc.CallOption) (*ListRolesResponse, error)
	ListPolicies(ctx context.Context, in *ListPoliciesRequest, opts ...grpc.CallOption) (*ListPoliciesResponse, error)
	GetPolicy(ctx context.Context, in *GetPolicyRequest, opts ...grpc.CallOption) (*GetPolicyResponse, error)
	CreatePolicy(ctx context.Context, in *CreatePolicyRequest, opts ...grpc.CallOption) (*CreatePolicyResponse, error)
	UpdatePolicy(ctx context.Context, in *UpdatePolicyRequest, opts ...grpc.CallOption) (*UpdatePolicyResponse, error)
	ListResources(ctx context.Context, in *ListResourcesRequest, opts ...grpc.CallOption) (*ListResourcesResponse, error)
	GetResource(ctx context.Context, in *GetResourceRequest, opts ...grpc.CallOption) (*GetResourceResponse, error)
	UpdateResource(ctx context.Context, in *UpdateResourceRequest, opts ...grpc.CallOption) (*UpdateResourceResponse, error)
	ListUserAppeals(ctx context.Context, in *ListUserAppealsRequest, opts ...grpc.CallOption) (*ListUserAppealsResponse, error)
	ListAppeals(ctx context.Context, in *ListAppealsRequest, opts ...grpc.CallOption) (*ListAppealsResponse, error)
	GetAppeal(ctx context.Context, in *GetAppealRequest, opts ...grpc.CallOption) (*GetAppealResponse, error)
	CancelAppeal(ctx context.Context, in *CancelAppealRequest, opts ...grpc.CallOption) (*CancelAppealResponse, error)
	RevokeAppeal(ctx context.Context, in *RevokeAppealRequest, opts ...grpc.CallOption) (*RevokeAppealResponse, error)
	CreateAppeal(ctx context.Context, in *CreateAppealRequest, opts ...grpc.CallOption) (*CreateAppealResponse, error)
	ListUserApprovals(ctx context.Context, in *ListUserApprovalsRequest, opts ...grpc.CallOption) (*ListUserApprovalsResponse, error)
	ListApprovals(ctx context.Context, in *ListApprovalsRequest, opts ...grpc.CallOption) (*ListApprovalsResponse, error)
	UpdateApproval(ctx context.Context, in *UpdateApprovalRequest, opts ...grpc.CallOption) (*UpdateApprovalResponse, error)
}

type guardianServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGuardianServiceClient(cc grpc.ClientConnInterface) GuardianServiceClient {
	return &guardianServiceClient{cc}
}

func (c *guardianServiceClient) ListProviders(ctx context.Context, in *ListProvidersRequest, opts ...grpc.CallOption) (*ListProvidersResponse, error) {
	out := new(ListProvidersResponse)
	err := c.cc.Invoke(ctx, "/odpf.guardian.v1beta1.GuardianService/ListProviders", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guardianServiceClient) GetProvider(ctx context.Context, in *GetProviderRequest, opts ...grpc.CallOption) (*GetProviderResponse, error) {
	out := new(GetProviderResponse)
	err := c.cc.Invoke(ctx, "/odpf.guardian.v1beta1.GuardianService/GetProvider", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guardianServiceClient) GetProviderTypes(ctx context.Context, in *GetProviderTypesRequest, opts ...grpc.CallOption) (*GetProviderTypesResponse, error) {
	out := new(GetProviderTypesResponse)
	err := c.cc.Invoke(ctx, "/odpf.guardian.v1beta1.GuardianService/GetProviderTypes", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guardianServiceClient) CreateProvider(ctx context.Context, in *CreateProviderRequest, opts ...grpc.CallOption) (*CreateProviderResponse, error) {
	out := new(CreateProviderResponse)
	err := c.cc.Invoke(ctx, "/odpf.guardian.v1beta1.GuardianService/CreateProvider", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guardianServiceClient) UpdateProvider(ctx context.Context, in *UpdateProviderRequest, opts ...grpc.CallOption) (*UpdateProviderResponse, error) {
	out := new(UpdateProviderResponse)
	err := c.cc.Invoke(ctx, "/odpf.guardian.v1beta1.GuardianService/UpdateProvider", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guardianServiceClient) ListRoles(ctx context.Context, in *ListRolesRequest, opts ...grpc.CallOption) (*ListRolesResponse, error) {
	out := new(ListRolesResponse)
	err := c.cc.Invoke(ctx, "/odpf.guardian.v1beta1.GuardianService/ListRoles", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guardianServiceClient) ListPolicies(ctx context.Context, in *ListPoliciesRequest, opts ...grpc.CallOption) (*ListPoliciesResponse, error) {
	out := new(ListPoliciesResponse)
	err := c.cc.Invoke(ctx, "/odpf.guardian.v1beta1.GuardianService/ListPolicies", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guardianServiceClient) GetPolicy(ctx context.Context, in *GetPolicyRequest, opts ...grpc.CallOption) (*GetPolicyResponse, error) {
	out := new(GetPolicyResponse)
	err := c.cc.Invoke(ctx, "/odpf.guardian.v1beta1.GuardianService/GetPolicy", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guardianServiceClient) CreatePolicy(ctx context.Context, in *CreatePolicyRequest, opts ...grpc.CallOption) (*CreatePolicyResponse, error) {
	out := new(CreatePolicyResponse)
	err := c.cc.Invoke(ctx, "/odpf.guardian.v1beta1.GuardianService/CreatePolicy", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guardianServiceClient) UpdatePolicy(ctx context.Context, in *UpdatePolicyRequest, opts ...grpc.CallOption) (*UpdatePolicyResponse, error) {
	out := new(UpdatePolicyResponse)
	err := c.cc.Invoke(ctx, "/odpf.guardian.v1beta1.GuardianService/UpdatePolicy", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guardianServiceClient) ListResources(ctx context.Context, in *ListResourcesRequest, opts ...grpc.CallOption) (*ListResourcesResponse, error) {
	out := new(ListResourcesResponse)
	err := c.cc.Invoke(ctx, "/odpf.guardian.v1beta1.GuardianService/ListResources", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guardianServiceClient) GetResource(ctx context.Context, in *GetResourceRequest, opts ...grpc.CallOption) (*GetResourceResponse, error) {
	out := new(GetResourceResponse)
	err := c.cc.Invoke(ctx, "/odpf.guardian.v1beta1.GuardianService/GetResource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guardianServiceClient) UpdateResource(ctx context.Context, in *UpdateResourceRequest, opts ...grpc.CallOption) (*UpdateResourceResponse, error) {
	out := new(UpdateResourceResponse)
	err := c.cc.Invoke(ctx, "/odpf.guardian.v1beta1.GuardianService/UpdateResource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guardianServiceClient) ListUserAppeals(ctx context.Context, in *ListUserAppealsRequest, opts ...grpc.CallOption) (*ListUserAppealsResponse, error) {
	out := new(ListUserAppealsResponse)
	err := c.cc.Invoke(ctx, "/odpf.guardian.v1beta1.GuardianService/ListUserAppeals", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guardianServiceClient) ListAppeals(ctx context.Context, in *ListAppealsRequest, opts ...grpc.CallOption) (*ListAppealsResponse, error) {
	out := new(ListAppealsResponse)
	err := c.cc.Invoke(ctx, "/odpf.guardian.v1beta1.GuardianService/ListAppeals", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guardianServiceClient) GetAppeal(ctx context.Context, in *GetAppealRequest, opts ...grpc.CallOption) (*GetAppealResponse, error) {
	out := new(GetAppealResponse)
	err := c.cc.Invoke(ctx, "/odpf.guardian.v1beta1.GuardianService/GetAppeal", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guardianServiceClient) CancelAppeal(ctx context.Context, in *CancelAppealRequest, opts ...grpc.CallOption) (*CancelAppealResponse, error) {
	out := new(CancelAppealResponse)
	err := c.cc.Invoke(ctx, "/odpf.guardian.v1beta1.GuardianService/CancelAppeal", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guardianServiceClient) RevokeAppeal(ctx context.Context, in *RevokeAppealRequest, opts ...grpc.CallOption) (*RevokeAppealResponse, error) {
	out := new(RevokeAppealResponse)
	err := c.cc.Invoke(ctx, "/odpf.guardian.v1beta1.GuardianService/RevokeAppeal", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guardianServiceClient) CreateAppeal(ctx context.Context, in *CreateAppealRequest, opts ...grpc.CallOption) (*CreateAppealResponse, error) {
	out := new(CreateAppealResponse)
	err := c.cc.Invoke(ctx, "/odpf.guardian.v1beta1.GuardianService/CreateAppeal", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guardianServiceClient) ListUserApprovals(ctx context.Context, in *ListUserApprovalsRequest, opts ...grpc.CallOption) (*ListUserApprovalsResponse, error) {
	out := new(ListUserApprovalsResponse)
	err := c.cc.Invoke(ctx, "/odpf.guardian.v1beta1.GuardianService/ListUserApprovals", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guardianServiceClient) ListApprovals(ctx context.Context, in *ListApprovalsRequest, opts ...grpc.CallOption) (*ListApprovalsResponse, error) {
	out := new(ListApprovalsResponse)
	err := c.cc.Invoke(ctx, "/odpf.guardian.v1beta1.GuardianService/ListApprovals", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guardianServiceClient) UpdateApproval(ctx context.Context, in *UpdateApprovalRequest, opts ...grpc.CallOption) (*UpdateApprovalResponse, error) {
	out := new(UpdateApprovalResponse)
	err := c.cc.Invoke(ctx, "/odpf.guardian.v1beta1.GuardianService/UpdateApproval", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GuardianServiceServer is the server API for GuardianService service.
// All implementations must embed UnimplementedGuardianServiceServer
// for forward compatibility
type GuardianServiceServer interface {
	ListProviders(context.Context, *ListProvidersRequest) (*ListProvidersResponse, error)
	GetProvider(context.Context, *GetProviderRequest) (*GetProviderResponse, error)
	GetProviderTypes(context.Context, *GetProviderTypesRequest) (*GetProviderTypesResponse, error)
	CreateProvider(context.Context, *CreateProviderRequest) (*CreateProviderResponse, error)
	UpdateProvider(context.Context, *UpdateProviderRequest) (*UpdateProviderResponse, error)
	ListRoles(context.Context, *ListRolesRequest) (*ListRolesResponse, error)
	ListPolicies(context.Context, *ListPoliciesRequest) (*ListPoliciesResponse, error)
	GetPolicy(context.Context, *GetPolicyRequest) (*GetPolicyResponse, error)
	CreatePolicy(context.Context, *CreatePolicyRequest) (*CreatePolicyResponse, error)
	UpdatePolicy(context.Context, *UpdatePolicyRequest) (*UpdatePolicyResponse, error)
	ListResources(context.Context, *ListResourcesRequest) (*ListResourcesResponse, error)
	GetResource(context.Context, *GetResourceRequest) (*GetResourceResponse, error)
	UpdateResource(context.Context, *UpdateResourceRequest) (*UpdateResourceResponse, error)
	ListUserAppeals(context.Context, *ListUserAppealsRequest) (*ListUserAppealsResponse, error)
	ListAppeals(context.Context, *ListAppealsRequest) (*ListAppealsResponse, error)
	GetAppeal(context.Context, *GetAppealRequest) (*GetAppealResponse, error)
	CancelAppeal(context.Context, *CancelAppealRequest) (*CancelAppealResponse, error)
	RevokeAppeal(context.Context, *RevokeAppealRequest) (*RevokeAppealResponse, error)
	CreateAppeal(context.Context, *CreateAppealRequest) (*CreateAppealResponse, error)
	ListUserApprovals(context.Context, *ListUserApprovalsRequest) (*ListUserApprovalsResponse, error)
	ListApprovals(context.Context, *ListApprovalsRequest) (*ListApprovalsResponse, error)
	UpdateApproval(context.Context, *UpdateApprovalRequest) (*UpdateApprovalResponse, error)
	mustEmbedUnimplementedGuardianServiceServer()
}

// UnimplementedGuardianServiceServer must be embedded to have forward compatible implementations.
type UnimplementedGuardianServiceServer struct {
}

func (UnimplementedGuardianServiceServer) ListProviders(context.Context, *ListProvidersRequest) (*ListProvidersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListProviders not implemented")
}
func (UnimplementedGuardianServiceServer) GetProvider(context.Context, *GetProviderRequest) (*GetProviderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetProvider not implemented")
}
func (UnimplementedGuardianServiceServer) GetProviderTypes(context.Context, *GetProviderTypesRequest) (*GetProviderTypesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetProviderTypes not implemented")
}
func (UnimplementedGuardianServiceServer) CreateProvider(context.Context, *CreateProviderRequest) (*CreateProviderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateProvider not implemented")
}
func (UnimplementedGuardianServiceServer) UpdateProvider(context.Context, *UpdateProviderRequest) (*UpdateProviderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateProvider not implemented")
}
func (UnimplementedGuardianServiceServer) ListRoles(context.Context, *ListRolesRequest) (*ListRolesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListRoles not implemented")
}
func (UnimplementedGuardianServiceServer) ListPolicies(context.Context, *ListPoliciesRequest) (*ListPoliciesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListPolicies not implemented")
}
func (UnimplementedGuardianServiceServer) GetPolicy(context.Context, *GetPolicyRequest) (*GetPolicyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPolicy not implemented")
}
func (UnimplementedGuardianServiceServer) CreatePolicy(context.Context, *CreatePolicyRequest) (*CreatePolicyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreatePolicy not implemented")
}
func (UnimplementedGuardianServiceServer) UpdatePolicy(context.Context, *UpdatePolicyRequest) (*UpdatePolicyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdatePolicy not implemented")
}
func (UnimplementedGuardianServiceServer) ListResources(context.Context, *ListResourcesRequest) (*ListResourcesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListResources not implemented")
}
func (UnimplementedGuardianServiceServer) GetResource(context.Context, *GetResourceRequest) (*GetResourceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetResource not implemented")
}
func (UnimplementedGuardianServiceServer) UpdateResource(context.Context, *UpdateResourceRequest) (*UpdateResourceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateResource not implemented")
}
func (UnimplementedGuardianServiceServer) ListUserAppeals(context.Context, *ListUserAppealsRequest) (*ListUserAppealsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListUserAppeals not implemented")
}
func (UnimplementedGuardianServiceServer) ListAppeals(context.Context, *ListAppealsRequest) (*ListAppealsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAppeals not implemented")
}
func (UnimplementedGuardianServiceServer) GetAppeal(context.Context, *GetAppealRequest) (*GetAppealResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAppeal not implemented")
}
func (UnimplementedGuardianServiceServer) CancelAppeal(context.Context, *CancelAppealRequest) (*CancelAppealResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CancelAppeal not implemented")
}
func (UnimplementedGuardianServiceServer) RevokeAppeal(context.Context, *RevokeAppealRequest) (*RevokeAppealResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RevokeAppeal not implemented")
}
func (UnimplementedGuardianServiceServer) CreateAppeal(context.Context, *CreateAppealRequest) (*CreateAppealResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAppeal not implemented")
}
func (UnimplementedGuardianServiceServer) ListUserApprovals(context.Context, *ListUserApprovalsRequest) (*ListUserApprovalsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListUserApprovals not implemented")
}
func (UnimplementedGuardianServiceServer) ListApprovals(context.Context, *ListApprovalsRequest) (*ListApprovalsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListApprovals not implemented")
}
func (UnimplementedGuardianServiceServer) UpdateApproval(context.Context, *UpdateApprovalRequest) (*UpdateApprovalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateApproval not implemented")
}
func (UnimplementedGuardianServiceServer) mustEmbedUnimplementedGuardianServiceServer() {}

// UnsafeGuardianServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GuardianServiceServer will
// result in compilation errors.
type UnsafeGuardianServiceServer interface {
	mustEmbedUnimplementedGuardianServiceServer()
}

func RegisterGuardianServiceServer(s grpc.ServiceRegistrar, srv GuardianServiceServer) {
	s.RegisterService(&GuardianService_ServiceDesc, srv)
}

func _GuardianService_ListProviders_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListProvidersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardianServiceServer).ListProviders(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/odpf.guardian.v1beta1.GuardianService/ListProviders",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardianServiceServer).ListProviders(ctx, req.(*ListProvidersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuardianService_GetProvider_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetProviderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardianServiceServer).GetProvider(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/odpf.guardian.v1beta1.GuardianService/GetProvider",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardianServiceServer).GetProvider(ctx, req.(*GetProviderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuardianService_GetProviderTypes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetProviderTypesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardianServiceServer).GetProviderTypes(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/odpf.guardian.v1beta1.GuardianService/GetProviderTypes",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardianServiceServer).GetProviderTypes(ctx, req.(*GetProviderTypesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuardianService_CreateProvider_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateProviderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardianServiceServer).CreateProvider(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/odpf.guardian.v1beta1.GuardianService/CreateProvider",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardianServiceServer).CreateProvider(ctx, req.(*CreateProviderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuardianService_UpdateProvider_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateProviderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardianServiceServer).UpdateProvider(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/odpf.guardian.v1beta1.GuardianService/UpdateProvider",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardianServiceServer).UpdateProvider(ctx, req.(*UpdateProviderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuardianService_ListRoles_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRolesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardianServiceServer).ListRoles(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/odpf.guardian.v1beta1.GuardianService/ListRoles",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardianServiceServer).ListRoles(ctx, req.(*ListRolesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuardianService_ListPolicies_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListPoliciesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardianServiceServer).ListPolicies(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/odpf.guardian.v1beta1.GuardianService/ListPolicies",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardianServiceServer).ListPolicies(ctx, req.(*ListPoliciesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuardianService_GetPolicy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPolicyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardianServiceServer).GetPolicy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/odpf.guardian.v1beta1.GuardianService/GetPolicy",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardianServiceServer).GetPolicy(ctx, req.(*GetPolicyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuardianService_CreatePolicy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreatePolicyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardianServiceServer).CreatePolicy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/odpf.guardian.v1beta1.GuardianService/CreatePolicy",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardianServiceServer).CreatePolicy(ctx, req.(*CreatePolicyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuardianService_UpdatePolicy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdatePolicyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardianServiceServer).UpdatePolicy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/odpf.guardian.v1beta1.GuardianService/UpdatePolicy",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardianServiceServer).UpdatePolicy(ctx, req.(*UpdatePolicyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuardianService_ListResources_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListResourcesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardianServiceServer).ListResources(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/odpf.guardian.v1beta1.GuardianService/ListResources",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardianServiceServer).ListResources(ctx, req.(*ListResourcesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuardianService_GetResource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetResourceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardianServiceServer).GetResource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/odpf.guardian.v1beta1.GuardianService/GetResource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardianServiceServer).GetResource(ctx, req.(*GetResourceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuardianService_UpdateResource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateResourceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardianServiceServer).UpdateResource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/odpf.guardian.v1beta1.GuardianService/UpdateResource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardianServiceServer).UpdateResource(ctx, req.(*UpdateResourceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuardianService_ListUserAppeals_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListUserAppealsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardianServiceServer).ListUserAppeals(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/odpf.guardian.v1beta1.GuardianService/ListUserAppeals",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardianServiceServer).ListUserAppeals(ctx, req.(*ListUserAppealsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuardianService_ListAppeals_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListAppealsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardianServiceServer).ListAppeals(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/odpf.guardian.v1beta1.GuardianService/ListAppeals",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardianServiceServer).ListAppeals(ctx, req.(*ListAppealsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuardianService_GetAppeal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAppealRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardianServiceServer).GetAppeal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/odpf.guardian.v1beta1.GuardianService/GetAppeal",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardianServiceServer).GetAppeal(ctx, req.(*GetAppealRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuardianService_CancelAppeal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CancelAppealRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardianServiceServer).CancelAppeal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/odpf.guardian.v1beta1.GuardianService/CancelAppeal",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardianServiceServer).CancelAppeal(ctx, req.(*CancelAppealRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuardianService_RevokeAppeal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RevokeAppealRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardianServiceServer).RevokeAppeal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/odpf.guardian.v1beta1.GuardianService/RevokeAppeal",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardianServiceServer).RevokeAppeal(ctx, req.(*RevokeAppealRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuardianService_CreateAppeal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateAppealRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardianServiceServer).CreateAppeal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/odpf.guardian.v1beta1.GuardianService/CreateAppeal",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardianServiceServer).CreateAppeal(ctx, req.(*CreateAppealRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuardianService_ListUserApprovals_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListUserApprovalsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardianServiceServer).ListUserApprovals(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/odpf.guardian.v1beta1.GuardianService/ListUserApprovals",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardianServiceServer).ListUserApprovals(ctx, req.(*ListUserApprovalsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuardianService_ListApprovals_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListApprovalsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardianServiceServer).ListApprovals(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/odpf.guardian.v1beta1.GuardianService/ListApprovals",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardianServiceServer).ListApprovals(ctx, req.(*ListApprovalsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuardianService_UpdateApproval_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateApprovalRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuardianServiceServer).UpdateApproval(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/odpf.guardian.v1beta1.GuardianService/UpdateApproval",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuardianServiceServer).UpdateApproval(ctx, req.(*UpdateApprovalRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// GuardianService_ServiceDesc is the grpc.ServiceDesc for GuardianService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GuardianService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "odpf.guardian.v1beta1.GuardianService",
	HandlerType: (*GuardianServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListProviders",
			Handler:    _GuardianService_ListProviders_Handler,
		},
		{
			MethodName: "GetProvider",
			Handler:    _GuardianService_GetProvider_Handler,
		},
		{
			MethodName: "GetProviderTypes",
			Handler:    _GuardianService_GetProviderTypes_Handler,
		},
		{
			MethodName: "CreateProvider",
			Handler:    _GuardianService_CreateProvider_Handler,
		},
		{
			MethodName: "UpdateProvider",
			Handler:    _GuardianService_UpdateProvider_Handler,
		},
		{
			MethodName: "ListRoles",
			Handler:    _GuardianService_ListRoles_Handler,
		},
		{
			MethodName: "ListPolicies",
			Handler:    _GuardianService_ListPolicies_Handler,
		},
		{
			MethodName: "GetPolicy",
			Handler:    _GuardianService_GetPolicy_Handler,
		},
		{
			MethodName: "CreatePolicy",
			Handler:    _GuardianService_CreatePolicy_Handler,
		},
		{
			MethodName: "UpdatePolicy",
			Handler:    _GuardianService_UpdatePolicy_Handler,
		},
		{
			MethodName: "ListResources",
			Handler:    _GuardianService_ListResources_Handler,
		},
		{
			MethodName: "GetResource",
			Handler:    _GuardianService_GetResource_Handler,
		},
		{
			MethodName: "UpdateResource",
			Handler:    _GuardianService_UpdateResource_Handler,
		},
		{
			MethodName: "ListUserAppeals",
			Handler:    _GuardianService_ListUserAppeals_Handler,
		},
		{
			MethodName: "ListAppeals",
			Handler:    _GuardianService_ListAppeals_Handler,
		},
		{
			MethodName: "GetAppeal",
			Handler:    _GuardianService_GetAppeal_Handler,
		},
		{
			MethodName: "CancelAppeal",
			Handler:    _GuardianService_CancelAppeal_Handler,
		},
		{
			MethodName: "RevokeAppeal",
			Handler:    _GuardianService_RevokeAppeal_Handler,
		},
		{
			MethodName: "CreateAppeal",
			Handler:    _GuardianService_CreateAppeal_Handler,
		},
		{
			MethodName: "ListUserApprovals",
			Handler:    _GuardianService_ListUserApprovals_Handler,
		},
		{
			MethodName: "ListApprovals",
			Handler:    _GuardianService_ListApprovals_Handler,
		},
		{
			MethodName: "UpdateApproval",
			Handler:    _GuardianService_UpdateApproval_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "odpf/guardian/v1beta1/guardian.proto",
}
