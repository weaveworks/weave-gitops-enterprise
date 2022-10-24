package app

import (
	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/mgmtfetcher"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/server"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	core "github.com/weaveworks/weave-gitops/core/server"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	core_server "github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Options specifies the options that can be set
// in RunInProcessGateway.
type Options struct {
	Log                          logr.Logger
	KubernetesClient             client.Client
	DiscoveryClient              discovery.DiscoveryInterface
	GitProvider                  git.Provider
	ApplicationsConfig           *core_server.ApplicationsConfig
	CoreServerConfig             core.CoreServerConfig
	ApplicationsOptions          []core_server.ApplicationsOption
	ProfilesConfig               server.ProfilesConfig
	ClusterFetcher               clustersmngr.ClusterFetcher
	GrpcRuntimeOptions           []runtime.ServeMuxOption
	RuntimeNamespace             string
	ProfileHelmRepository        string
	HelmRepositoryCacheDirectory string
	CAPIClustersNamespace        string
	CAPIEnabled                  bool
	EntitlementSecretKey         client.ObjectKey
	HtmlRootPath                 string
	ClientGetter                 kube.ClientGetter
	AuthMethods                  map[auth.AuthMethod]bool
	OIDC                         OIDCAuthenticationOptions
	TLSCert                      string
	TLSKey                       string
	NoTLS                        bool
	DevMode                      bool
	ClustersManager              clustersmngr.ClustersManager
	ChartsCache                  *helm.HelmChartIndexer
	KubernetesClientSet          kubernetes.Interface
	ManagementFetcher            *mgmtfetcher.ManagementCrossNamespacesFetcher
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

// WithGitProvider is used to set a git provider that makes
// API calls to GitHub or GitLab.
func WithGitProvider(gitProvider git.Provider) Option {
	return func(o *Options) {
		o.GitProvider = gitProvider
	}
}

// WithApplicationsConfig is used to set the configuration needed to work
// with Weave GitOps Core applications
func WithApplicationsConfig(appConfig *core_server.ApplicationsConfig) Option {
	return func(o *Options) {
		o.ApplicationsConfig = appConfig
	}
}

// WithApplicationsOptions is used to set the configuration needed to work
// with Weave GitOps Core applications
func WithApplicationsOptions(appOptions ...core_server.ApplicationsOption) Option {
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

// WithRuntimeNamespace set the namespace that holds any authentication
// secrets (e.g. cluster-user-auth or oidc-auth).
func WithRuntimeNamespace(RuntimeNamespace string) Option {
	return func(o *Options) {
		o.RuntimeNamespace = RuntimeNamespace
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

// WithAuthConfig is used to set the auth configuration including OIDC
func WithAuthConfig(authMethods map[auth.AuthMethod]bool, oidc OIDCAuthenticationOptions) Option {
	return func(o *Options) {
		o.AuthMethods = authMethods
		o.OIDC = oidc
	}
}

// WithTLSConfig is used to set the TLS configuration.
func WithTLSConfig(tlsCert, tlsKey string, noTLS bool) Option {
	return func(o *Options) {
		o.TLSCert = tlsCert
		o.TLSKey = tlsKey
		o.NoTLS = noTLS
	}
}

// WithCAPIEnabled is enabled/disable CAPI support
// If the CAPI CRDS are not installed in the cluster and CAPI is
// enabled, the system will error on certain routes
func WithCAPIEnabled(capiEnabled bool) Option {
	return func(o *Options) {
		o.CAPIEnabled = capiEnabled
	}
}

// WithDevMode starts the server in development mode
func WithDevMode(devMode bool) Option {
	return func(o *Options) {
		o.DevMode = devMode
	}
}

// WithClustersManager defines the clusters manager that will be use for cross-cluster queries.
func WithClustersManager(factory clustersmngr.ClustersManager) Option {
	return func(o *Options) {
		o.ClustersManager = factory
	}
}

// WithClustersCache defines the clusters cache that will be use for cross-cluster queries.
func WithChartsCache(chartCache *helm.HelmChartIndexer) Option {
	return func(o *Options) {
		o.ChartsCache = chartCache
	}
}

// WithKubernetesClientSet defines the kuberntes client set that will be used for
func WithKubernetesClientSet(kubernetesClientSet kubernetes.Interface) Option {
	return func(o *Options) {
		o.KubernetesClientSet = kubernetesClientSet
	}
}

// WithManagemetFetcher defines the mangement fetcher to be used
func WithManagemetFetcher(fetcher *mgmtfetcher.ManagementCrossNamespacesFetcher) Option {
	return func(o *Options) {
		o.ManagementFetcher = fetcher
	}
}
