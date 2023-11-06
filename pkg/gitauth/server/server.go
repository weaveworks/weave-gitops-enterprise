package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitlab"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	grpcStatus "google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/rand"

	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/gitauth"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/gitauth/azure"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/gitauth/bitbucket"
	gp "github.com/weaveworks/weave-gitops-enterprise/pkg/gitauth/server/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
)

const DefaultHost = "0.0.0.0"
const DefaultPort = "9001"
const (
	GitProviderCSRFHeaderName = "x-git-provider-csrf"
	GitProviderCSRFCookieName = "_git_provider_csrf"
)

var (
	ErrEmptyAccessToken = errors.New("access token is empty")
	ErrBadProvider      = errors.New("wrong provider name")
)

// RandomTokenGenerator is used to generate random (CSRF) tokens for the OAuth flow.
// The default implementation uses `uuid.NewString()` to generate a random token.
type RandomTokenGenerator func() string

type applicationServer struct {
	pb.UnimplementedGitAuthServer

	jwtClient           auth.JWTClient
	log                 logr.Logger
	ghAuthClient        auth.GithubAuthClient
	glAuthClient        auth.GitlabAuthClient
	bbAuthClient        bitbucket.AuthClient
	azureDevOpsClient   azure.AuthClient
	generateRandomToken RandomTokenGenerator
}

// An ApplicationsConfig allows for the customization of an ApplicationsServer.
// Use the DefaultConfig() to use the default dependencies.
type ApplicationsConfig struct {
	Logger                logr.Logger
	JwtClient             auth.JWTClient
	GithubAuthClient      auth.GithubAuthClient
	GitlabAuthClient      auth.GitlabAuthClient
	BitBucketServerClient bitbucket.AuthClient
	AzureDevOpsClient     azure.AuthClient
	RandomTokenGenerator  RandomTokenGenerator
}

// NewApplicationsServer creates a grpc Applications server
func NewApplicationsServer(cfg *ApplicationsConfig, setters ...ApplicationsOption) pb.GitAuthServer {
	args := &ApplicationsOptions{}

	for _, setter := range setters {
		setter(args)
	}

	return &applicationServer{
		jwtClient:           cfg.JwtClient,
		log:                 cfg.Logger,
		ghAuthClient:        cfg.GithubAuthClient,
		glAuthClient:        cfg.GitlabAuthClient,
		bbAuthClient:        cfg.BitBucketServerClient,
		azureDevOpsClient:   cfg.AzureDevOpsClient,
		generateRandomToken: cfg.RandomTokenGenerator,
	}
}

// DefaultApplicationsConfig creates a populated config with the dependencies for a Server
func DefaultApplicationsConfig(log logr.Logger) (*ApplicationsConfig, error) {
	rand.Seed(time.Now().UnixNano())
	secretKey := rand.String(20)
	envSecretKey := os.Getenv("GITOPS_JWT_ENCRYPTION_SECRET")

	if envSecretKey != "" {
		secretKey = envSecretKey
	}

	jwtClient := auth.NewJwtClient(secretKey)

	return &ApplicationsConfig{
		Logger:                log.WithName("app-server"),
		JwtClient:             jwtClient,
		GithubAuthClient:      auth.NewGithubAuthClient(http.DefaultClient),
		GitlabAuthClient:      auth.NewGitlabAuthClient(http.DefaultClient),
		BitBucketServerClient: bitbucket.NewAuthClient(http.DefaultClient),
		AzureDevOpsClient:     azure.NewAuthClient(http.DefaultClient),
		RandomTokenGenerator:  uuid.NewString,
	}, nil
}

func (s *applicationServer) GetGithubDeviceCode(ctx context.Context, msg *pb.GetGithubDeviceCodeRequest) (*pb.GetGithubDeviceCodeResponse, error) {
	res, err := s.ghAuthClient.GetDeviceCode()
	if err != nil {
		return nil, fmt.Errorf("error doing github code request: %w", err)
	}

	return &pb.GetGithubDeviceCodeResponse{
		UserCode:      res.UserCode,
		ValidationUri: res.VerificationURI,
		DeviceCode:    res.DeviceCode,
		Interval:      int32(res.Interval),
	}, nil
}

func (s *applicationServer) GetGithubAuthStatus(ctx context.Context, msg *pb.GetGithubAuthStatusRequest) (*pb.GetGithubAuthStatusResponse, error) {
	token, err := s.ghAuthClient.GetDeviceCodeAuthStatus(msg.DeviceCode)
	if err == auth.ErrAuthPending {
		return nil, grpcStatus.Error(codes.Unauthenticated, err.Error())
	} else if err != nil {
		return nil, fmt.Errorf("error getting github device code status: %w", err)
	}

	t, err := s.jwtClient.GenerateJWT(auth.ExpirationTime, gitproviders.GitProviderGitHub, token)
	if err != nil {
		return nil, fmt.Errorf("could not generate token: %w", err)
	}

	return &pb.GetGithubAuthStatusResponse{AccessToken: t}, nil
}

// Authenticate generates and returns a jwt token using git provider name and git provider token
func (s *applicationServer) Authenticate(_ context.Context, msg *pb.AuthenticateRequest) (*pb.AuthenticateResponse, error) {
	if !strings.HasPrefix(github.DefaultDomain, msg.ProviderName) &&
		!strings.HasPrefix(gitlab.DefaultDomain, msg.ProviderName) {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "%s expected github or gitlab, got %s", ErrBadProvider, msg.ProviderName)
	}

	if msg.AccessToken == "" {
		return nil, grpcStatus.Error(codes.InvalidArgument, ErrEmptyAccessToken.Error())
	}

	token, err := s.jwtClient.GenerateJWT(auth.ExpirationTime, gitproviders.GitProviderName(msg.GetProviderName()), msg.GetAccessToken())
	if err != nil {
		return nil, grpcStatus.Errorf(codes.Internal, "error generating jwt token. %s", err)
	}

	return &pb.AuthenticateResponse{Token: token}, nil
}

func (s *applicationServer) ParseRepoURL(ctx context.Context, msg *pb.ParseRepoURLRequest) (*pb.ParseRepoURLResponse, error) {
	u, err := gp.NewRepoURL(msg.Url)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "could not parse url: %s", err.Error())
	}

	return &pb.ParseRepoURLResponse{
		Name:     u.RepositoryName(),
		Owner:    u.Owner(),
		Provider: toProtoProvider(u.Provider()),
	}, nil
}

func (s *applicationServer) GetGitlabAuthURL(ctx context.Context, msg *pb.GetGitlabAuthURLRequest) (*pb.GetGitlabAuthURLResponse, error) {
	u, err := s.glAuthClient.AuthURL(ctx, msg.RedirectUri)
	if err != nil {
		return nil, fmt.Errorf("could not get gitlab auth url: %w", err)
	}

	return &pb.GetGitlabAuthURLResponse{Url: u.String()}, nil
}

func (s *applicationServer) AuthorizeGitlab(ctx context.Context, msg *pb.AuthorizeGitlabRequest) (*pb.AuthorizeGitlabResponse, error) {
	tokenState, err := s.glAuthClient.ExchangeCode(ctx, msg.RedirectUri, msg.Code)
	if err != nil {
		return nil, fmt.Errorf("could not exchange code: %w", err)
	}

	token, err := s.jwtClient.GenerateJWT(tokenState.ExpiresIn, gitproviders.GitProviderGitLab, tokenState.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("could not generate token: %w", err)
	}

	return &pb.AuthorizeGitlabResponse{Token: token}, nil
}

func (s *applicationServer) GetBitbucketServerAuthURL(ctx context.Context, msg *pb.GetBitbucketServerAuthURLRequest) (*pb.GetBitbucketServerAuthURLResponse, error) {
	// Generate a random state value
	state := s.generateRandomToken()
	// Set a gRPC header so that middleware can inspect it and issue a cookie with this value in the HTTP response
	err := grpc.SetHeader(ctx, metadata.Pairs(GitProviderCSRFHeaderName, state))
	if err != nil {
		s.log.Error(err, "Failed to set gRPC header for CSRF token")
		return nil, fmt.Errorf("failed to set state parameter for OAuth flow")
	}

	u, err := s.bbAuthClient.AuthURL(ctx, msg.RedirectUri, state)
	if err != nil {
		return nil, fmt.Errorf("failed to construct bitbucket server auth url: %w", err)
	}

	return &pb.GetBitbucketServerAuthURLResponse{Url: u.String()}, nil
}

func (s *applicationServer) AuthorizeBitbucketServer(ctx context.Context, msg *pb.AuthorizeBitbucketServerRequest) (*pb.AuthorizeBitbucketServerResponse, error) {
	err := checkCSRFToken(ctx, msg.State)
	if err != nil {
		s.log.Error(err, "Failed CSRF token check")
		return nil, fmt.Errorf("failed CSRF token check")
	}

	tokenState, err := s.bbAuthClient.ExchangeCode(ctx, msg.RedirectUri, msg.Code)
	if err != nil {
		return nil, fmt.Errorf("could not exchange code: %w", err)
	}

	token, err := s.jwtClient.GenerateJWT(tokenState.ExpiresIn, gitproviders.GitProviderBitBucketServer, tokenState.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("could not generate token: %w", err)
	}

	return &pb.AuthorizeBitbucketServerResponse{Token: token}, nil
}

func (s *applicationServer) GetAzureDevOpsAuthURL(ctx context.Context, msg *pb.GetAzureDevOpsAuthURLRequest) (*pb.GetAzureDevOpsAuthURLResponse, error) {
	// Generate a random state value
	state := s.generateRandomToken()
	// Set a gRPC header so that middleware can inspect it and issue a cookie with this value in the HTTP response
	err := grpc.SetHeader(ctx, metadata.Pairs(GitProviderCSRFHeaderName, state))
	if err != nil {
		s.log.Error(err, "Failed to set gRPC header for CSRF token")
		return nil, fmt.Errorf("failed to set state parameter for OAuth flow")
	}

	u, err := s.azureDevOpsClient.AuthURL(ctx, msg.RedirectUri, state)
	if err != nil {
		return nil, fmt.Errorf("could not get azure auth url: %w", err)
	}

	return &pb.GetAzureDevOpsAuthURLResponse{Url: u.String()}, nil
}

func (s *applicationServer) AuthorizeAzureDevOps(ctx context.Context, msg *pb.AuthorizeAzureDevOpsRequest) (*pb.AuthorizeAzureDevOpsResponse, error) {
	err := checkCSRFToken(ctx, msg.State)
	if err != nil {
		s.log.Error(err, "Failed CSRF token check")
		return nil, fmt.Errorf("failed CSRF token check")
	}

	tokenState, err := s.azureDevOpsClient.ExchangeCode(ctx, msg.RedirectUri, msg.Code)
	if err != nil {
		return nil, fmt.Errorf("could not exchange code: %w", err)
	}

	token, err := s.jwtClient.GenerateJWT(tokenState.ExpiresIn, gitproviders.GitProviderAzureDevOps, tokenState.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("could not generate token: %w", err)
	}

	return &pb.AuthorizeAzureDevOpsResponse{Token: token}, nil
}

func (s *applicationServer) ValidateProviderToken(ctx context.Context, msg *pb.ValidateProviderTokenRequest) (*pb.ValidateProviderTokenResponse, error) {
	token, err := middleware.ExtractProviderToken(ctx)
	if err != nil {
		return nil, grpcStatus.Error(codes.Unauthenticated, err.Error())
	}

	v, err := s.findValidator(msg.Provider)
	if err != nil {
		return nil, grpcStatus.Error(codes.InvalidArgument, err.Error())
	}

	if err := v.ValidateToken(ctx, token.AccessToken); err != nil {
		return nil, grpcStatus.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.ValidateProviderTokenResponse{
		Valid: true,
	}, nil
}

func toProtoProvider(p gp.GitProviderName) pb.GitProvider {
	switch p {
	case gp.GitProviderGitHub:
		return pb.GitProvider_GitHub
	case gp.GitProviderGitLab:
		return pb.GitProvider_GitLab
	case gp.GitProviderBitBucketServer:
		return pb.GitProvider_BitBucketServer
	case gp.GitProviderAzureDevOps:
		return pb.GitProvider_AzureDevOps
	}

	return pb.GitProvider_Unknown
}

func (s *applicationServer) findValidator(provider pb.GitProvider) (auth.ProviderTokenValidator, error) {
	switch provider {
	case pb.GitProvider_GitHub:
		return s.ghAuthClient, nil
	case pb.GitProvider_GitLab:
		return s.glAuthClient, nil
	case pb.GitProvider_BitBucketServer:
		return s.bbAuthClient, nil
	case pb.GitProvider_AzureDevOps:
		return s.azureDevOpsClient, nil
	}

	return nil, fmt.Errorf("unknown git provider %s", provider)
}

// checkCSRFToken inspects the incoming context for a cookie, reads the CSRF value
// and compares it to the `state` value coming back from the OAuth provider.
func checkCSRFToken(ctx context.Context, state string) error {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get(runtime.MetadataPrefix + "cookie")
		if len(values) > 0 {
			cookieState := getStateFromCookie(values[0])
			if cookieState != state {
				return fmt.Errorf("CSRF token check has failed, state parameter mismatch: %s %s", cookieState, state)
			}
		}
	}
	return nil
}

// getStateFromCookie takes a raw value of the Cookie header from the incoming request and
// constructs from that an array of cookies that can be easily inspected. If the CSRF cookie
// is found in the array, then its value gets returned. Otherwise an empty string is returned.
func getStateFromCookie(cookie string) string {
	header := http.Header{}
	header.Add("Cookie", cookie)
	req := http.Request{Header: header}
	state := ""
	for _, c := range req.Cookies() {
		if c.Name == GitProviderCSRFCookieName {
			state = c.Value
		}
	}
	return state
}
