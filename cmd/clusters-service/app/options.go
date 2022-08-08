package app

import (
	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/clusters"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	core "github.com/weaveworks/weave-gitops/core/server"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/server"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Options specifies the options that can be set
// in RunInProcessGateway.
type Options struct {
	Log                          logr.Logger
	KubernetesClient             client.Client
	DiscoveryClient              discovery.DiscoveryInterface
	ClustersLibrary              clusters.Library
	TemplateLibrary              templates.Library
	GitProvider                  git.Provider
	ApplicationsConfig           *server.ApplicationsConfig
	CoreServerConfig             core.CoreServerConfig
	ApplicationsOptions          []server.ApplicationsOption
	ProfilesConfig               server.ProfilesConfig
	ClusterFetcher               clustersmngr.ClusterFetcher
	GrpcRuntimeOptions           []runtime.ServeMuxOption
	ProfileHelmRepository        string
	HelmRepositoryCacheDirectory string
	CAPIClustersNamespace        string
	CAPIEnabled                  string
	EntitlementSecretKey         client.ObjectKey
	HtmlRootPath                 string
	ClientGetter                 kube.ClientGetter
	OIDC                         OIDCAuthenticationOptions
	TLSCert                      string
	TLSKey                       string
	NoTLS                        bool
}

type Option func(*Options)

// WithLog is used to set a logger.
func WithLog(log logr.Logger) Option {
	return func(o *Options) {
		o.Log = log
	}
}

// WithKubernetesClient is used to set a Kubernetes
// client.
func WithKubernetesClient(client client.Client) Option {
	return func(o *Options) {
		o.KubernetesClient = client
	}
}

// WithKubernetesClient is used to set a Kubernetes
// discovery client.
func WithDiscoveryClient(client discovery.DiscoveryInterface) Option {
	return func(o *Options) {
		o.DiscoveryClient = client
	}
}

// WithClustersLibrary is used to set the location that contains
// CAPI templates. Typically this will be a namespace in the cluster.
func WithClustersLibrary(clustersLibrary clusters.Library) Option {
	return func(o *Options) {
		o.ClustersLibrary = clustersLibrary
	}
}

// WithTemplateLibrary is used to set the location that contains
// CAPI templates. Typically this will be a namespace in the cluster.
func WithTemplateLibrary(templateLibrary templates.Library) Option {
	return func(o *Options) {
		o.TemplateLibrary = templateLibrary
	}
}

// WithGitProvider is used to set a git provider that makes
// API calls to GitHub or GitLab.
func WithGitProvider(gitProvider git.Provider) Option {
	return func(o *Options) {
		o.GitProvider = gitProvider
	}
}

// WithApplicationsConfig is used to set the configuration needed to work
// with Weave GitOps Core applications
func WithApplicationsConfig(appConfig *server.ApplicationsConfig) Option {
	return func(o *Options) {
		o.ApplicationsConfig = appConfig
	}
}

// WithApplicationsOptions is used to set the configuration needed to work
// with Weave GitOps Core applications
func WithApplicationsOptions(appOptions ...server.ApplicationsOption) Option {
	return func(o *Options) {
		o.ApplicationsOptions = appOptions
	}
}

// WithCoreConfig is used to set the configuration needed to work
// with Weave GitOps Core
func WithCoreConfig(coreServerConfig core.CoreServerConfig) Option {
	return func(o *Options) {
		o.CoreServerConfig = coreServerConfig
	}
}

// WithProfilesConfig is used to set the configuration needed to work
// with Weave GitOps Core profiles
func WithProfilesConfig(profilesConfig server.ProfilesConfig) Option {
	return func(o *Options) {
		o.ProfilesConfig = profilesConfig
	}
}

// WithProfileHelmRepository is used to set the name of the Flux
// HelmRepository object that will be inspected for Helm charts
// that include the profile annotation.
func WithProfileHelmRepository(name string) Option {
	return func(o *Options) {
		o.ProfileHelmRepository = name
	}
}

// WithGrpcRuntimeOptions is used to set an array of ServeMuxOption that
// will be used to configure the GRPC-Gateway.
func WithGrpcRuntimeOptions(grpcRuntimeOptions []runtime.ServeMuxOption) Option {
	return func(o *Options) {
		o.GrpcRuntimeOptions = grpcRuntimeOptions
	}
}

// WithCAPIClustersNamespace is used to set the namespace that will
// be monitored for new CAPI cluster CRs.
func WithCAPIClustersNamespace(namespace string) Option {
	return func(o *Options) {
		o.CAPIClustersNamespace = namespace
	}
}

// WithEntitlementSecretKey is used to set the key (name/namespace)
// that refers to the entitlement secret.
func WithEntitlementSecretKey(key client.ObjectKey) Option {
	return func(o *Options) {
		o.EntitlementSecretKey = key
	}
}

// WithHelmRepositoryCacheDirectory is used to set the directory
// for the Helm repository cache.
func WithHelmRepositoryCacheDirectory(cacheDir string) Option {
	return func(o *Options) {
		o.HelmRepositoryCacheDirectory = cacheDir
	}
}

// WithHtmlRootPath sets the directory on the filesystem to
// serve static assets like the frontend from.
func WithHtmlRootPath(path string) Option {
	return func(o *Options) {
		o.HtmlRootPath = path
	}
}

// WithClientGetter is used to set the client getter
// to use when querying the Kubernetes API.
func WithClientGetter(clientGetter kube.ClientGetter) Option {
	return func(o *Options) {
		o.ClientGetter = clientGetter
	}
}

// WithOIDCConfig is used to set the OIDC configuration.
func WithOIDCConfig(oidc OIDCAuthenticationOptions) Option {
	return func(o *Options) {
		o.OIDC = oidc
	}
}

func WithTLSConfig(tlsCert, tlsKey string, noTLS bool) Option {
	return func(o *Options) {
		o.TLSCert = tlsCert
		o.TLSKey = tlsKey
		o.NoTLS = noTLS
	}
}
