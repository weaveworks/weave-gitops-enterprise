// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package api

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

// TerraformClient is the client API for Terraform service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TerraformClient interface {
	ListTerraformObjects(ctx context.Context, in *ListTerraformObjectsRequest, opts ...grpc.CallOption) (*ListTerraformObjectsResponse, error)
	GetTerraformObject(ctx context.Context, in *GetTerraformObjectRequest, opts ...grpc.CallOption) (*GetTerraformObjectResponse, error)
	SyncTerraformObject(ctx context.Context, in *SyncTerraformObjectRequest, opts ...grpc.CallOption) (*SyncTerraformObjectResponse, error)
	ToggleSuspendTerraformObject(ctx context.Context, in *ToggleSuspendTerraformObjectRequest, opts ...grpc.CallOption) (*ToggleSuspendTerraformObjectResponse, error)
	GetTerraformObjectPlan(ctx context.Context, in *GetTerraformObjectPlanRequest, opts ...grpc.CallOption) (*GetTerraformObjectPlanResponse, error)
	ReplanTerraformObject(ctx context.Context, in *ReplanTerraformObjectRequest, opts ...grpc.CallOption) (*ReplanTerraformObjectResponse, error)
}

type terraformClient struct {
	cc grpc.ClientConnInterface
}

func NewTerraformClient(cc grpc.ClientConnInterface) TerraformClient {
	return &terraformClient{cc}
}

func (c *terraformClient) ListTerraformObjects(ctx context.Context, in *ListTerraformObjectsRequest, opts ...grpc.CallOption) (*ListTerraformObjectsResponse, error) {
	out := new(ListTerraformObjectsResponse)
	err := c.cc.Invoke(ctx, "/terraform.v1.Terraform/ListTerraformObjects", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *terraformClient) GetTerraformObject(ctx context.Context, in *GetTerraformObjectRequest, opts ...grpc.CallOption) (*GetTerraformObjectResponse, error) {
	out := new(GetTerraformObjectResponse)
	err := c.cc.Invoke(ctx, "/terraform.v1.Terraform/GetTerraformObject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *terraformClient) SyncTerraformObject(ctx context.Context, in *SyncTerraformObjectRequest, opts ...grpc.CallOption) (*SyncTerraformObjectResponse, error) {
	out := new(SyncTerraformObjectResponse)
	err := c.cc.Invoke(ctx, "/terraform.v1.Terraform/SyncTerraformObject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *terraformClient) ToggleSuspendTerraformObject(ctx context.Context, in *ToggleSuspendTerraformObjectRequest, opts ...grpc.CallOption) (*ToggleSuspendTerraformObjectResponse, error) {
	out := new(ToggleSuspendTerraformObjectResponse)
	err := c.cc.Invoke(ctx, "/terraform.v1.Terraform/ToggleSuspendTerraformObject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *terraformClient) GetTerraformObjectPlan(ctx context.Context, in *GetTerraformObjectPlanRequest, opts ...grpc.CallOption) (*GetTerraformObjectPlanResponse, error) {
	out := new(GetTerraformObjectPlanResponse)
	err := c.cc.Invoke(ctx, "/terraform.v1.Terraform/GetTerraformObjectPlan", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *terraformClient) ReplanTerraformObject(ctx context.Context, in *ReplanTerraformObjectRequest, opts ...grpc.CallOption) (*ReplanTerraformObjectResponse, error) {
	out := new(ReplanTerraformObjectResponse)
	err := c.cc.Invoke(ctx, "/terraform.v1.Terraform/ReplanTerraformObject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TerraformServer is the server API for Terraform service.
// All implementations must embed UnimplementedTerraformServer
// for forward compatibility
type TerraformServer interface {
	ListTerraformObjects(context.Context, *ListTerraformObjectsRequest) (*ListTerraformObjectsResponse, error)
	GetTerraformObject(context.Context, *GetTerraformObjectRequest) (*GetTerraformObjectResponse, error)
	SyncTerraformObject(context.Context, *SyncTerraformObjectRequest) (*SyncTerraformObjectResponse, error)
	ToggleSuspendTerraformObject(context.Context, *ToggleSuspendTerraformObjectRequest) (*ToggleSuspendTerraformObjectResponse, error)
	GetTerraformObjectPlan(context.Context, *GetTerraformObjectPlanRequest) (*GetTerraformObjectPlanResponse, error)
	ReplanTerraformObject(context.Context, *ReplanTerraformObjectRequest) (*ReplanTerraformObjectResponse, error)
	mustEmbedUnimplementedTerraformServer()
}

// UnimplementedTerraformServer must be embedded to have forward compatible implementations.
type UnimplementedTerraformServer struct {
}

func (UnimplementedTerraformServer) ListTerraformObjects(context.Context, *ListTerraformObjectsRequest) (*ListTerraformObjectsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListTerraformObjects not implemented")
}
func (UnimplementedTerraformServer) GetTerraformObject(context.Context, *GetTerraformObjectRequest) (*GetTerraformObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTerraformObject not implemented")
}
func (UnimplementedTerraformServer) SyncTerraformObject(context.Context, *SyncTerraformObjectRequest) (*SyncTerraformObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SyncTerraformObject not implemented")
}
func (UnimplementedTerraformServer) ToggleSuspendTerraformObject(context.Context, *ToggleSuspendTerraformObjectRequest) (*ToggleSuspendTerraformObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ToggleSuspendTerraformObject not implemented")
}
func (UnimplementedTerraformServer) GetTerraformObjectPlan(context.Context, *GetTerraformObjectPlanRequest) (*GetTerraformObjectPlanResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTerraformObjectPlan not implemented")
}
func (UnimplementedTerraformServer) ReplanTerraformObject(context.Context, *ReplanTerraformObjectRequest) (*ReplanTerraformObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReplanTerraformObject not implemented")
}
func (UnimplementedTerraformServer) mustEmbedUnimplementedTerraformServer() {}

// UnsafeTerraformServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TerraformServer will
// result in compilation errors.
type UnsafeTerraformServer interface {
	mustEmbedUnimplementedTerraformServer()
}

func RegisterTerraformServer(s grpc.ServiceRegistrar, srv TerraformServer) {
	s.RegisterService(&Terraform_ServiceDesc, srv)
}

func _Terraform_ListTerraformObjects_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListTerraformObjectsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TerraformServer).ListTerraformObjects(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/terraform.v1.Terraform/ListTerraformObjects",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TerraformServer).ListTerraformObjects(ctx, req.(*ListTerraformObjectsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Terraform_GetTerraformObject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTerraformObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TerraformServer).GetTerraformObject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/terraform.v1.Terraform/GetTerraformObject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TerraformServer).GetTerraformObject(ctx, req.(*GetTerraformObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Terraform_SyncTerraformObject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SyncTerraformObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TerraformServer).SyncTerraformObject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/terraform.v1.Terraform/SyncTerraformObject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TerraformServer).SyncTerraformObject(ctx, req.(*SyncTerraformObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Terraform_ToggleSuspendTerraformObject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ToggleSuspendTerraformObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TerraformServer).ToggleSuspendTerraformObject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/terraform.v1.Terraform/ToggleSuspendTerraformObject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TerraformServer).ToggleSuspendTerraformObject(ctx, req.(*ToggleSuspendTerraformObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Terraform_GetTerraformObjectPlan_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTerraformObjectPlanRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TerraformServer).GetTerraformObjectPlan(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/terraform.v1.Terraform/GetTerraformObjectPlan",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TerraformServer).GetTerraformObjectPlan(ctx, req.(*GetTerraformObjectPlanRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Terraform_ReplanTerraformObject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReplanTerraformObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TerraformServer).ReplanTerraformObject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/terraform.v1.Terraform/ReplanTerraformObject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TerraformServer).ReplanTerraformObject(ctx, req.(*ReplanTerraformObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Terraform_ServiceDesc is the grpc.ServiceDesc for Terraform service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Terraform_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "terraform.v1.Terraform",
	HandlerType: (*TerraformServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListTerraformObjects",
			Handler:    _Terraform_ListTerraformObjects_Handler,
		},
		{
			MethodName: "GetTerraformObject",
			Handler:    _Terraform_GetTerraformObject_Handler,
		},
		{
			MethodName: "SyncTerraformObject",
			Handler:    _Terraform_SyncTerraformObject_Handler,
		},
		{
			MethodName: "ToggleSuspendTerraformObject",
			Handler:    _Terraform_ToggleSuspendTerraformObject_Handler,
		},
		{
			MethodName: "GetTerraformObjectPlan",
			Handler:    _Terraform_GetTerraformObjectPlan_Handler,
		},
		{
			MethodName: "ReplanTerraformObject",
			Handler:    _Terraform_ReplanTerraformObject_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/terraform/terraform.proto",
}
