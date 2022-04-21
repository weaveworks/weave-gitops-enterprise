// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package capi_server

import (
	context "context"
	httpbody "google.golang.org/genproto/googleapis/api/httpbody"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ClustersServiceClient is the client API for ClustersService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ClustersServiceClient interface {
	ListTemplates(ctx context.Context, in *ListTemplatesRequest, opts ...grpc.CallOption) (*ListTemplatesResponse, error)
	GetTemplate(ctx context.Context, in *GetTemplateRequest, opts ...grpc.CallOption) (*GetTemplateResponse, error)
	ListTemplateParams(ctx context.Context, in *ListTemplateParamsRequest, opts ...grpc.CallOption) (*ListTemplateParamsResponse, error)
	// Returns a list of profiles within that template
	// `gitops get <template-name> --list-profiles`
	// The template annotations appear in the following form
	// capi.weave.works/profile-<n> where n is a number
	ListTemplateProfiles(ctx context.Context, in *ListTemplateProfilesRequest, opts ...grpc.CallOption) (*ListTemplateProfilesResponse, error)
	RenderTemplate(ctx context.Context, in *RenderTemplateRequest, opts ...grpc.CallOption) (*RenderTemplateResponse, error)
	ListGitopsClusters(ctx context.Context, in *ListGitopsClustersRequest, opts ...grpc.CallOption) (*ListGitopsClustersResponse, error)
	// Creates a pull request for a cluster template.
	// The template name and values will be used to
	// create a new branch for which a new pull request
	// will be created.
	CreatePullRequest(ctx context.Context, in *CreatePullRequestRequest, opts ...grpc.CallOption) (*CreatePullRequestResponse, error)
	DeleteClustersPullRequest(ctx context.Context, in *DeleteClustersPullRequestRequest, opts ...grpc.CallOption) (*DeleteClustersPullRequestResponse, error)
	ListCredentials(ctx context.Context, in *ListCredentialsRequest, opts ...grpc.CallOption) (*ListCredentialsResponse, error)
	// GetKubeconfig returns the Kubeconfig for the given
	// workload cluster.
	GetKubeconfig(ctx context.Context, in *GetKubeconfigRequest, opts ...grpc.CallOption) (*httpbody.HttpBody, error)
	// GetEnterpriseVersion returns the WeGO Enterprise version
	GetEnterpriseVersion(ctx context.Context, in *GetEnterpriseVersionRequest, opts ...grpc.CallOption) (*GetEnterpriseVersionResponse, error)
	GetConfig(ctx context.Context, in *GetConfigRequest, opts ...grpc.CallOption) (*GetConfigResponse, error)
	// ListPolicies list policies available on the management cluster
	ListPolicies(ctx context.Context, in *ListPoliciesRequest, opts ...grpc.CallOption) (*ListPoliciesResponse, error)
	// GetPolicy gets a policy on the management cluster by name
	GetPolicy(ctx context.Context, in *GetPolicyRequest, opts ...grpc.CallOption) (*GetPolicyResponse, error)
}

type clustersServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewClustersServiceClient(cc grpc.ClientConnInterface) ClustersServiceClient {
	return &clustersServiceClient{cc}
}

func (c *clustersServiceClient) ListTemplates(ctx context.Context, in *ListTemplatesRequest, opts ...grpc.CallOption) (*ListTemplatesResponse, error) {
	out := new(ListTemplatesResponse)
	err := c.cc.Invoke(ctx, "/capi_server.v1.ClustersService/ListTemplates", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clustersServiceClient) GetTemplate(ctx context.Context, in *GetTemplateRequest, opts ...grpc.CallOption) (*GetTemplateResponse, error) {
	out := new(GetTemplateResponse)
	err := c.cc.Invoke(ctx, "/capi_server.v1.ClustersService/GetTemplate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clustersServiceClient) ListTemplateParams(ctx context.Context, in *ListTemplateParamsRequest, opts ...grpc.CallOption) (*ListTemplateParamsResponse, error) {
	out := new(ListTemplateParamsResponse)
	err := c.cc.Invoke(ctx, "/capi_server.v1.ClustersService/ListTemplateParams", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clustersServiceClient) ListTemplateProfiles(ctx context.Context, in *ListTemplateProfilesRequest, opts ...grpc.CallOption) (*ListTemplateProfilesResponse, error) {
	out := new(ListTemplateProfilesResponse)
	err := c.cc.Invoke(ctx, "/capi_server.v1.ClustersService/ListTemplateProfiles", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clustersServiceClient) RenderTemplate(ctx context.Context, in *RenderTemplateRequest, opts ...grpc.CallOption) (*RenderTemplateResponse, error) {
	out := new(RenderTemplateResponse)
	err := c.cc.Invoke(ctx, "/capi_server.v1.ClustersService/RenderTemplate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clustersServiceClient) ListGitopsClusters(ctx context.Context, in *ListGitopsClustersRequest, opts ...grpc.CallOption) (*ListGitopsClustersResponse, error) {
	out := new(ListGitopsClustersResponse)
	err := c.cc.Invoke(ctx, "/capi_server.v1.ClustersService/ListGitopsClusters", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clustersServiceClient) CreatePullRequest(ctx context.Context, in *CreatePullRequestRequest, opts ...grpc.CallOption) (*CreatePullRequestResponse, error) {
	out := new(CreatePullRequestResponse)
	err := c.cc.Invoke(ctx, "/capi_server.v1.ClustersService/CreatePullRequest", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clustersServiceClient) DeleteClustersPullRequest(ctx context.Context, in *DeleteClustersPullRequestRequest, opts ...grpc.CallOption) (*DeleteClustersPullRequestResponse, error) {
	out := new(DeleteClustersPullRequestResponse)
	err := c.cc.Invoke(ctx, "/capi_server.v1.ClustersService/DeleteClustersPullRequest", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clustersServiceClient) ListCredentials(ctx context.Context, in *ListCredentialsRequest, opts ...grpc.CallOption) (*ListCredentialsResponse, error) {
	out := new(ListCredentialsResponse)
	err := c.cc.Invoke(ctx, "/capi_server.v1.ClustersService/ListCredentials", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clustersServiceClient) GetKubeconfig(ctx context.Context, in *GetKubeconfigRequest, opts ...grpc.CallOption) (*httpbody.HttpBody, error) {
	out := new(httpbody.HttpBody)
	err := c.cc.Invoke(ctx, "/capi_server.v1.ClustersService/GetKubeconfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clustersServiceClient) GetEnterpriseVersion(ctx context.Context, in *GetEnterpriseVersionRequest, opts ...grpc.CallOption) (*GetEnterpriseVersionResponse, error) {
	out := new(GetEnterpriseVersionResponse)
	err := c.cc.Invoke(ctx, "/capi_server.v1.ClustersService/GetEnterpriseVersion", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clustersServiceClient) GetConfig(ctx context.Context, in *GetConfigRequest, opts ...grpc.CallOption) (*GetConfigResponse, error) {
	out := new(GetConfigResponse)
	err := c.cc.Invoke(ctx, "/capi_server.v1.ClustersService/GetConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clustersServiceClient) ListPolicies(ctx context.Context, in *ListPoliciesRequest, opts ...grpc.CallOption) (*ListPoliciesResponse, error) {
	out := new(ListPoliciesResponse)
	err := c.cc.Invoke(ctx, "/capi_server.v1.ClustersService/ListPolicies", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clustersServiceClient) GetPolicy(ctx context.Context, in *GetPolicyRequest, opts ...grpc.CallOption) (*GetPolicyResponse, error) {
	out := new(GetPolicyResponse)
	err := c.cc.Invoke(ctx, "/capi_server.v1.ClustersService/GetPolicy", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ClustersServiceServer is the server API for ClustersService service.
// All implementations must embed UnimplementedClustersServiceServer
// for forward compatibility
type ClustersServiceServer interface {
	ListTemplates(context.Context, *ListTemplatesRequest) (*ListTemplatesResponse, error)
	GetTemplate(context.Context, *GetTemplateRequest) (*GetTemplateResponse, error)
	ListTemplateParams(context.Context, *ListTemplateParamsRequest) (*ListTemplateParamsResponse, error)
	// Returns a list of profiles within that template
	// `gitops get <template-name> --list-profiles`
	// The template annotations appear in the following form
	// capi.weave.works/profile-<n> where n is a number
	ListTemplateProfiles(context.Context, *ListTemplateProfilesRequest) (*ListTemplateProfilesResponse, error)
	RenderTemplate(context.Context, *RenderTemplateRequest) (*RenderTemplateResponse, error)
	ListGitopsClusters(context.Context, *ListGitopsClustersRequest) (*ListGitopsClustersResponse, error)
	// Creates a pull request for a cluster template.
	// The template name and values will be used to
	// create a new branch for which a new pull request
	// will be created.
	CreatePullRequest(context.Context, *CreatePullRequestRequest) (*CreatePullRequestResponse, error)
	DeleteClustersPullRequest(context.Context, *DeleteClustersPullRequestRequest) (*DeleteClustersPullRequestResponse, error)
	ListCredentials(context.Context, *ListCredentialsRequest) (*ListCredentialsResponse, error)
	// GetKubeconfig returns the Kubeconfig for the given
	// workload cluster.
	GetKubeconfig(context.Context, *GetKubeconfigRequest) (*httpbody.HttpBody, error)
	// GetEnterpriseVersion returns the WeGO Enterprise version
	GetEnterpriseVersion(context.Context, *GetEnterpriseVersionRequest) (*GetEnterpriseVersionResponse, error)
	GetConfig(context.Context, *GetConfigRequest) (*GetConfigResponse, error)
	// ListPolicies list policies available on the management cluster
	ListPolicies(context.Context, *ListPoliciesRequest) (*ListPoliciesResponse, error)
	// GetPolicy gets a policy on the management cluster by name
	GetPolicy(context.Context, *GetPolicyRequest) (*GetPolicyResponse, error)
	mustEmbedUnimplementedClustersServiceServer()
}

// UnimplementedClustersServiceServer must be embedded to have forward compatible implementations.
type UnimplementedClustersServiceServer struct {
}

func (UnimplementedClustersServiceServer) ListTemplates(context.Context, *ListTemplatesRequest) (*ListTemplatesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListTemplates not implemented")
}
func (UnimplementedClustersServiceServer) GetTemplate(context.Context, *GetTemplateRequest) (*GetTemplateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTemplate not implemented")
}
func (UnimplementedClustersServiceServer) ListTemplateParams(context.Context, *ListTemplateParamsRequest) (*ListTemplateParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListTemplateParams not implemented")
}
func (UnimplementedClustersServiceServer) ListTemplateProfiles(context.Context, *ListTemplateProfilesRequest) (*ListTemplateProfilesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListTemplateProfiles not implemented")
}
func (UnimplementedClustersServiceServer) RenderTemplate(context.Context, *RenderTemplateRequest) (*RenderTemplateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RenderTemplate not implemented")
}
func (UnimplementedClustersServiceServer) ListGitopsClusters(context.Context, *ListGitopsClustersRequest) (*ListGitopsClustersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListGitopsClusters not implemented")
}
func (UnimplementedClustersServiceServer) CreatePullRequest(context.Context, *CreatePullRequestRequest) (*CreatePullRequestResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreatePullRequest not implemented")
}
func (UnimplementedClustersServiceServer) DeleteClustersPullRequest(context.Context, *DeleteClustersPullRequestRequest) (*DeleteClustersPullRequestResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteClustersPullRequest not implemented")
}
func (UnimplementedClustersServiceServer) ListCredentials(context.Context, *ListCredentialsRequest) (*ListCredentialsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListCredentials not implemented")
}
func (UnimplementedClustersServiceServer) GetKubeconfig(context.Context, *GetKubeconfigRequest) (*httpbody.HttpBody, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetKubeconfig not implemented")
}
func (UnimplementedClustersServiceServer) GetEnterpriseVersion(context.Context, *GetEnterpriseVersionRequest) (*GetEnterpriseVersionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetEnterpriseVersion not implemented")
}
func (UnimplementedClustersServiceServer) GetConfig(context.Context, *GetConfigRequest) (*GetConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetConfig not implemented")
}
func (UnimplementedClustersServiceServer) ListPolicies(context.Context, *ListPoliciesRequest) (*ListPoliciesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListPolicies not implemented")
}
func (UnimplementedClustersServiceServer) GetPolicy(context.Context, *GetPolicyRequest) (*GetPolicyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPolicy not implemented")
}
func (UnimplementedClustersServiceServer) mustEmbedUnimplementedClustersServiceServer() {}

// UnsafeClustersServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ClustersServiceServer will
// result in compilation errors.
type UnsafeClustersServiceServer interface {
	mustEmbedUnimplementedClustersServiceServer()
}

func RegisterClustersServiceServer(s grpc.ServiceRegistrar, srv ClustersServiceServer) {
	s.RegisterService(&ClustersService_ServiceDesc, srv)
}

func _ClustersService_ListTemplates_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListTemplatesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClustersServiceServer).ListTemplates(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/capi_server.v1.ClustersService/ListTemplates",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClustersServiceServer).ListTemplates(ctx, req.(*ListTemplatesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClustersService_GetTemplate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTemplateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClustersServiceServer).GetTemplate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/capi_server.v1.ClustersService/GetTemplate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClustersServiceServer).GetTemplate(ctx, req.(*GetTemplateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClustersService_ListTemplateParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListTemplateParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClustersServiceServer).ListTemplateParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/capi_server.v1.ClustersService/ListTemplateParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClustersServiceServer).ListTemplateParams(ctx, req.(*ListTemplateParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClustersService_ListTemplateProfiles_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListTemplateProfilesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClustersServiceServer).ListTemplateProfiles(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/capi_server.v1.ClustersService/ListTemplateProfiles",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClustersServiceServer).ListTemplateProfiles(ctx, req.(*ListTemplateProfilesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClustersService_RenderTemplate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RenderTemplateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClustersServiceServer).RenderTemplate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/capi_server.v1.ClustersService/RenderTemplate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClustersServiceServer).RenderTemplate(ctx, req.(*RenderTemplateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClustersService_ListGitopsClusters_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListGitopsClustersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClustersServiceServer).ListGitopsClusters(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/capi_server.v1.ClustersService/ListGitopsClusters",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClustersServiceServer).ListGitopsClusters(ctx, req.(*ListGitopsClustersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClustersService_CreatePullRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreatePullRequestRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClustersServiceServer).CreatePullRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/capi_server.v1.ClustersService/CreatePullRequest",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClustersServiceServer).CreatePullRequest(ctx, req.(*CreatePullRequestRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClustersService_DeleteClustersPullRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteClustersPullRequestRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClustersServiceServer).DeleteClustersPullRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/capi_server.v1.ClustersService/DeleteClustersPullRequest",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClustersServiceServer).DeleteClustersPullRequest(ctx, req.(*DeleteClustersPullRequestRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClustersService_ListCredentials_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListCredentialsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClustersServiceServer).ListCredentials(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/capi_server.v1.ClustersService/ListCredentials",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClustersServiceServer).ListCredentials(ctx, req.(*ListCredentialsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClustersService_GetKubeconfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetKubeconfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClustersServiceServer).GetKubeconfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/capi_server.v1.ClustersService/GetKubeconfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClustersServiceServer).GetKubeconfig(ctx, req.(*GetKubeconfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClustersService_GetEnterpriseVersion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetEnterpriseVersionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClustersServiceServer).GetEnterpriseVersion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/capi_server.v1.ClustersService/GetEnterpriseVersion",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClustersServiceServer).GetEnterpriseVersion(ctx, req.(*GetEnterpriseVersionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClustersService_GetConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClustersServiceServer).GetConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/capi_server.v1.ClustersService/GetConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClustersServiceServer).GetConfig(ctx, req.(*GetConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClustersService_ListPolicies_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListPoliciesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClustersServiceServer).ListPolicies(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/capi_server.v1.ClustersService/ListPolicies",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClustersServiceServer).ListPolicies(ctx, req.(*ListPoliciesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClustersService_GetPolicy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPolicyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClustersServiceServer).GetPolicy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/capi_server.v1.ClustersService/GetPolicy",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClustersServiceServer).GetPolicy(ctx, req.(*GetPolicyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ClustersService_ServiceDesc is the grpc.ServiceDesc for ClustersService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ClustersService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "capi_server.v1.ClustersService",
	HandlerType: (*ClustersServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListTemplates",
			Handler:    _ClustersService_ListTemplates_Handler,
		},
		{
			MethodName: "GetTemplate",
			Handler:    _ClustersService_GetTemplate_Handler,
		},
		{
			MethodName: "ListTemplateParams",
			Handler:    _ClustersService_ListTemplateParams_Handler,
		},
		{
			MethodName: "ListTemplateProfiles",
			Handler:    _ClustersService_ListTemplateProfiles_Handler,
		},
		{
			MethodName: "RenderTemplate",
			Handler:    _ClustersService_RenderTemplate_Handler,
		},
		{
			MethodName: "ListGitopsClusters",
			Handler:    _ClustersService_ListGitopsClusters_Handler,
		},
		{
			MethodName: "CreatePullRequest",
			Handler:    _ClustersService_CreatePullRequest_Handler,
		},
		{
			MethodName: "DeleteClustersPullRequest",
			Handler:    _ClustersService_DeleteClustersPullRequest_Handler,
		},
		{
			MethodName: "ListCredentials",
			Handler:    _ClustersService_ListCredentials_Handler,
		},
		{
			MethodName: "GetKubeconfig",
			Handler:    _ClustersService_GetKubeconfig_Handler,
		},
		{
			MethodName: "GetEnterpriseVersion",
			Handler:    _ClustersService_GetEnterpriseVersion_Handler,
		},
		{
			MethodName: "GetConfig",
			Handler:    _ClustersService_GetConfig_Handler,
		},
		{
			MethodName: "ListPolicies",
			Handler:    _ClustersService_ListPolicies_Handler,
		},
		{
			MethodName: "GetPolicy",
			Handler:    _ClustersService_GetPolicy_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "capi_server.proto",
}
