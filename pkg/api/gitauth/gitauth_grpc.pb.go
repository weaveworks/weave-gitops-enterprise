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

// GitAuthClient is the client API for GitAuth service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GitAuthClient interface {
	// Authenticate generates jwt token using git provider name and git provider token arguments
	Authenticate(ctx context.Context, in *AuthenticateRequest, opts ...grpc.CallOption) (*AuthenticateResponse, error)
	// GetGithubDeviceCode retrieves a temporary device code for Github authentication.
	// This code is used to start the Github device-flow.
	GetGithubDeviceCode(ctx context.Context, in *GetGithubDeviceCodeRequest, opts ...grpc.CallOption) (*GetGithubDeviceCodeResponse, error)
	// GetGithubAuthStatus gets the status of the Github device flow authentication requests.
	// Once the user has completed the Github device flow, an access token will be returned.
	// This token will expired in 15 minutes, after which the user will need to complete the flow again
	// to do Git Provider operations.
	GetGithubAuthStatus(ctx context.Context, in *GetGithubAuthStatusRequest, opts ...grpc.CallOption) (*GetGithubAuthStatusResponse, error)
	// GetGitlabAuthURL returns the URL to initiate a GitLab OAuth PKCE flow.
	// The user must browse to the returned URL to authorize the OAuth callback to the GitOps UI.
	// See the GitLab OAuth docs for more more information:
	// https://docs.gitlab.com/ee/api/oauth2.html#supported-oauth-20-flows
	GetGitlabAuthURL(ctx context.Context, in *GetGitlabAuthURLRequest, opts ...grpc.CallOption) (*GetGitlabAuthURLResponse, error)
	GetBitbucketServerAuthURL(ctx context.Context, in *GetBitbucketServerAuthURLRequest, opts ...grpc.CallOption) (*GetBitbucketServerAuthURLResponse, error)
	AuthorizeBitbucketServer(ctx context.Context, in *AuthorizeBitbucketServerRequest, opts ...grpc.CallOption) (*AuthorizeBitbucketServerResponse, error)
	// AuthorizeGitlab exchanges a GitLab code obtained via OAuth callback.
	// The returned token is useable for authentication with the GitOps server only.
	// See the GitLab OAuth docs for more more information:
	// https://docs.gitlab.com/ee/api/oauth2.html#supported-oauth-20-flows
	AuthorizeGitlab(ctx context.Context, in *AuthorizeGitlabRequest, opts ...grpc.CallOption) (*AuthorizeGitlabResponse, error)
	// ParseRepoURL returns structured data about a git repository URL
	ParseRepoURL(ctx context.Context, in *ParseRepoURLRequest, opts ...grpc.CallOption) (*ParseRepoURLResponse, error)
	// ValidateProviderToken check to see if the git provider token is still valid
	ValidateProviderToken(ctx context.Context, in *ValidateProviderTokenRequest, opts ...grpc.CallOption) (*ValidateProviderTokenResponse, error)
}

type gitAuthClient struct {
	cc grpc.ClientConnInterface
}

func NewGitAuthClient(cc grpc.ClientConnInterface) GitAuthClient {
	return &gitAuthClient{cc}
}

func (c *gitAuthClient) Authenticate(ctx context.Context, in *AuthenticateRequest, opts ...grpc.CallOption) (*AuthenticateResponse, error) {
	out := new(AuthenticateResponse)
	err := c.cc.Invoke(ctx, "/gitauth.v1.GitAuth/Authenticate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitAuthClient) GetGithubDeviceCode(ctx context.Context, in *GetGithubDeviceCodeRequest, opts ...grpc.CallOption) (*GetGithubDeviceCodeResponse, error) {
	out := new(GetGithubDeviceCodeResponse)
	err := c.cc.Invoke(ctx, "/gitauth.v1.GitAuth/GetGithubDeviceCode", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitAuthClient) GetGithubAuthStatus(ctx context.Context, in *GetGithubAuthStatusRequest, opts ...grpc.CallOption) (*GetGithubAuthStatusResponse, error) {
	out := new(GetGithubAuthStatusResponse)
	err := c.cc.Invoke(ctx, "/gitauth.v1.GitAuth/GetGithubAuthStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitAuthClient) GetGitlabAuthURL(ctx context.Context, in *GetGitlabAuthURLRequest, opts ...grpc.CallOption) (*GetGitlabAuthURLResponse, error) {
	out := new(GetGitlabAuthURLResponse)
	err := c.cc.Invoke(ctx, "/gitauth.v1.GitAuth/GetGitlabAuthURL", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitAuthClient) GetBitbucketServerAuthURL(ctx context.Context, in *GetBitbucketServerAuthURLRequest, opts ...grpc.CallOption) (*GetBitbucketServerAuthURLResponse, error) {
	out := new(GetBitbucketServerAuthURLResponse)
	err := c.cc.Invoke(ctx, "/gitauth.v1.GitAuth/GetBitbucketServerAuthURL", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitAuthClient) AuthorizeBitbucketServer(ctx context.Context, in *AuthorizeBitbucketServerRequest, opts ...grpc.CallOption) (*AuthorizeBitbucketServerResponse, error) {
	out := new(AuthorizeBitbucketServerResponse)
	err := c.cc.Invoke(ctx, "/gitauth.v1.GitAuth/AuthorizeBitbucketServer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitAuthClient) AuthorizeGitlab(ctx context.Context, in *AuthorizeGitlabRequest, opts ...grpc.CallOption) (*AuthorizeGitlabResponse, error) {
	out := new(AuthorizeGitlabResponse)
	err := c.cc.Invoke(ctx, "/gitauth.v1.GitAuth/AuthorizeGitlab", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitAuthClient) ParseRepoURL(ctx context.Context, in *ParseRepoURLRequest, opts ...grpc.CallOption) (*ParseRepoURLResponse, error) {
	out := new(ParseRepoURLResponse)
	err := c.cc.Invoke(ctx, "/gitauth.v1.GitAuth/ParseRepoURL", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitAuthClient) ValidateProviderToken(ctx context.Context, in *ValidateProviderTokenRequest, opts ...grpc.CallOption) (*ValidateProviderTokenResponse, error) {
	out := new(ValidateProviderTokenResponse)
	err := c.cc.Invoke(ctx, "/gitauth.v1.GitAuth/ValidateProviderToken", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GitAuthServer is the server API for GitAuth service.
// All implementations must embed UnimplementedGitAuthServer
// for forward compatibility
type GitAuthServer interface {
	// Authenticate generates jwt token using git provider name and git provider token arguments
	Authenticate(context.Context, *AuthenticateRequest) (*AuthenticateResponse, error)
	// GetGithubDeviceCode retrieves a temporary device code for Github authentication.
	// This code is used to start the Github device-flow.
	GetGithubDeviceCode(context.Context, *GetGithubDeviceCodeRequest) (*GetGithubDeviceCodeResponse, error)
	// GetGithubAuthStatus gets the status of the Github device flow authentication requests.
	// Once the user has completed the Github device flow, an access token will be returned.
	// This token will expired in 15 minutes, after which the user will need to complete the flow again
	// to do Git Provider operations.
	GetGithubAuthStatus(context.Context, *GetGithubAuthStatusRequest) (*GetGithubAuthStatusResponse, error)
	// GetGitlabAuthURL returns the URL to initiate a GitLab OAuth PKCE flow.
	// The user must browse to the returned URL to authorize the OAuth callback to the GitOps UI.
	// See the GitLab OAuth docs for more more information:
	// https://docs.gitlab.com/ee/api/oauth2.html#supported-oauth-20-flows
	GetGitlabAuthURL(context.Context, *GetGitlabAuthURLRequest) (*GetGitlabAuthURLResponse, error)
	GetBitbucketServerAuthURL(context.Context, *GetBitbucketServerAuthURLRequest) (*GetBitbucketServerAuthURLResponse, error)
	AuthorizeBitbucketServer(context.Context, *AuthorizeBitbucketServerRequest) (*AuthorizeBitbucketServerResponse, error)
	// AuthorizeGitlab exchanges a GitLab code obtained via OAuth callback.
	// The returned token is useable for authentication with the GitOps server only.
	// See the GitLab OAuth docs for more more information:
	// https://docs.gitlab.com/ee/api/oauth2.html#supported-oauth-20-flows
	AuthorizeGitlab(context.Context, *AuthorizeGitlabRequest) (*AuthorizeGitlabResponse, error)
	// ParseRepoURL returns structured data about a git repository URL
	ParseRepoURL(context.Context, *ParseRepoURLRequest) (*ParseRepoURLResponse, error)
	// ValidateProviderToken check to see if the git provider token is still valid
	ValidateProviderToken(context.Context, *ValidateProviderTokenRequest) (*ValidateProviderTokenResponse, error)
	mustEmbedUnimplementedGitAuthServer()
}

// UnimplementedGitAuthServer must be embedded to have forward compatible implementations.
type UnimplementedGitAuthServer struct {
}

func (UnimplementedGitAuthServer) Authenticate(context.Context, *AuthenticateRequest) (*AuthenticateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Authenticate not implemented")
}
func (UnimplementedGitAuthServer) GetGithubDeviceCode(context.Context, *GetGithubDeviceCodeRequest) (*GetGithubDeviceCodeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetGithubDeviceCode not implemented")
}
func (UnimplementedGitAuthServer) GetGithubAuthStatus(context.Context, *GetGithubAuthStatusRequest) (*GetGithubAuthStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetGithubAuthStatus not implemented")
}
func (UnimplementedGitAuthServer) GetGitlabAuthURL(context.Context, *GetGitlabAuthURLRequest) (*GetGitlabAuthURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetGitlabAuthURL not implemented")
}
func (UnimplementedGitAuthServer) GetBitbucketServerAuthURL(context.Context, *GetBitbucketServerAuthURLRequest) (*GetBitbucketServerAuthURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBitbucketServerAuthURL not implemented")
}
func (UnimplementedGitAuthServer) AuthorizeBitbucketServer(context.Context, *AuthorizeBitbucketServerRequest) (*AuthorizeBitbucketServerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AuthorizeBitbucketServer not implemented")
}
func (UnimplementedGitAuthServer) AuthorizeGitlab(context.Context, *AuthorizeGitlabRequest) (*AuthorizeGitlabResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AuthorizeGitlab not implemented")
}
func (UnimplementedGitAuthServer) ParseRepoURL(context.Context, *ParseRepoURLRequest) (*ParseRepoURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ParseRepoURL not implemented")
}
func (UnimplementedGitAuthServer) ValidateProviderToken(context.Context, *ValidateProviderTokenRequest) (*ValidateProviderTokenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateProviderToken not implemented")
}
func (UnimplementedGitAuthServer) mustEmbedUnimplementedGitAuthServer() {}

// UnsafeGitAuthServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GitAuthServer will
// result in compilation errors.
type UnsafeGitAuthServer interface {
	mustEmbedUnimplementedGitAuthServer()
}

func RegisterGitAuthServer(s grpc.ServiceRegistrar, srv GitAuthServer) {
	s.RegisterService(&GitAuth_ServiceDesc, srv)
}

func _GitAuth_Authenticate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AuthenticateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitAuthServer).Authenticate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitauth.v1.GitAuth/Authenticate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitAuthServer).Authenticate(ctx, req.(*AuthenticateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GitAuth_GetGithubDeviceCode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetGithubDeviceCodeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitAuthServer).GetGithubDeviceCode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitauth.v1.GitAuth/GetGithubDeviceCode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitAuthServer).GetGithubDeviceCode(ctx, req.(*GetGithubDeviceCodeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GitAuth_GetGithubAuthStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetGithubAuthStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitAuthServer).GetGithubAuthStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitauth.v1.GitAuth/GetGithubAuthStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitAuthServer).GetGithubAuthStatus(ctx, req.(*GetGithubAuthStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GitAuth_GetGitlabAuthURL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetGitlabAuthURLRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitAuthServer).GetGitlabAuthURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitauth.v1.GitAuth/GetGitlabAuthURL",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitAuthServer).GetGitlabAuthURL(ctx, req.(*GetGitlabAuthURLRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GitAuth_GetBitbucketServerAuthURL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBitbucketServerAuthURLRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitAuthServer).GetBitbucketServerAuthURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitauth.v1.GitAuth/GetBitbucketServerAuthURL",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitAuthServer).GetBitbucketServerAuthURL(ctx, req.(*GetBitbucketServerAuthURLRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GitAuth_AuthorizeBitbucketServer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AuthorizeBitbucketServerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitAuthServer).AuthorizeBitbucketServer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitauth.v1.GitAuth/AuthorizeBitbucketServer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitAuthServer).AuthorizeBitbucketServer(ctx, req.(*AuthorizeBitbucketServerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GitAuth_AuthorizeGitlab_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AuthorizeGitlabRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitAuthServer).AuthorizeGitlab(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitauth.v1.GitAuth/AuthorizeGitlab",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitAuthServer).AuthorizeGitlab(ctx, req.(*AuthorizeGitlabRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GitAuth_ParseRepoURL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ParseRepoURLRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitAuthServer).ParseRepoURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitauth.v1.GitAuth/ParseRepoURL",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitAuthServer).ParseRepoURL(ctx, req.(*ParseRepoURLRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GitAuth_ValidateProviderToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateProviderTokenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitAuthServer).ValidateProviderToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitauth.v1.GitAuth/ValidateProviderToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitAuthServer).ValidateProviderToken(ctx, req.(*ValidateProviderTokenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// GitAuth_ServiceDesc is the grpc.ServiceDesc for GitAuth service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GitAuth_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "gitauth.v1.GitAuth",
	HandlerType: (*GitAuthServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Authenticate",
			Handler:    _GitAuth_Authenticate_Handler,
		},
		{
			MethodName: "GetGithubDeviceCode",
			Handler:    _GitAuth_GetGithubDeviceCode_Handler,
		},
		{
			MethodName: "GetGithubAuthStatus",
			Handler:    _GitAuth_GetGithubAuthStatus_Handler,
		},
		{
			MethodName: "GetGitlabAuthURL",
			Handler:    _GitAuth_GetGitlabAuthURL_Handler,
		},
		{
			MethodName: "GetBitbucketServerAuthURL",
			Handler:    _GitAuth_GetBitbucketServerAuthURL_Handler,
		},
		{
			MethodName: "AuthorizeBitbucketServer",
			Handler:    _GitAuth_AuthorizeBitbucketServer_Handler,
		},
		{
			MethodName: "AuthorizeGitlab",
			Handler:    _GitAuth_AuthorizeGitlab_Handler,
		},
		{
			MethodName: "ParseRepoURL",
			Handler:    _GitAuth_ParseRepoURL_Handler,
		},
		{
			MethodName: "ValidateProviderToken",
			Handler:    _GitAuth_ValidateProviderToken_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/gitauth/gitauth.proto",
}
