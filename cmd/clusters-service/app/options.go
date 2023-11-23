package app

import (
	"github.com/alexedwards/scs/v2"
	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/mgmtfetcher"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/estimation"
	gitauth "github.com/weaveworks/weave-gitops-enterprise/pkg/gitauth/server"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/monitoring"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/monitoring/metrics"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/monitoring/profiling"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	core "github.com/weaveworks/weave-gitops/core/server"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Options specifies the options that can be set
// in RunInProcessGateway.
type Options struct {
	Log                       logr.Logger
	KubernetesClient          client.Client
	DiscoveryClient           discovery.DiscoveryInterface
	GitProvider               git.Provider
	ApplicationsConfig        *gitauth.ApplicationsConfig
	CoreServerConfig          core.CoreServerConfig
	ApplicationsOptions       []gitauth.ApplicationsOption
	ClusterFetcher            clustersmngr.ClusterFetcher
	GrpcRuntimeOptions        []runtime.ServeMuxOption
	RuntimeNamespace          string
	ProfileHelmRepository     types.NamespacedName
	CAPIClustersNamespace     string
	CAPIEnabled               bool
	EntitlementSecretKey      client.ObjectKey
	RoutePrefix               string
	HtmlRootPath              string
	ClientGetter              kube.ClientGetter
	AuthMethods               map[auth.AuthMethod]bool
	NoAuthUser                string
	SessionManager            auth.SessionManager
	OIDC                      OIDCAuthenticationOptions
	TLSCert                   string
	TLSKey                    string
	NoTLS                     bool
	ClustersManager           clustersmngr.ClustersManager
	ChartsCache               *helm.HelmChartIndexer
	KubernetesClientSet       kubernetes.Interface
	ManagementFetcher         *mgmtfetcher.ManagementCrossNamespacesFetcher
	Cluster                   string
	Estimator                 estimation.Estimator
	UIConfig                  string
	PipelineControllerAddress string
	CollectorServiceAccount   collector.ImpersonateServiceAccount
	MonitoringOptions         monitoring.Options
	ExplorerCleanerDisabled   bool
	ExplorerEnabledFor        []string
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
func WithApplicationsConfig(appConfig *gitauth.ApplicationsConfig) Option {
	return func(o *Options) {
		o.ApplicationsConfig = appConfig
	}
}

// WithApplicationsOptions is used to set the configuration needed to work
// with Weave GitOps Core applications
func WithApplicationsOptions(appOptions ...gitauth.ApplicationsOption) Option {
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
func WithProfileHelmRepository(repo types.NamespacedName) Option {
	return func(o *Options) {
		o.ProfileHelmRepository = repo
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
func WithAuthConfig(authMethods map[auth.AuthMethod]bool, oidc OIDCAuthenticationOptions, noAuthUser string, sessionManager *scs.SessionManager) Option {
	return func(o *Options) {
		o.AuthMethods = authMethods
		o.OIDC = oidc
		o.NoAuthUser = noAuthUser
		o.SessionManager = sessionManager
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

// WithManagementCluster is used to set the management cluster name
func WithManagementCluster(cluster string) Option {
	return func(o *Options) {
		o.Cluster = cluster
	}
}

func WithTemplateCostEstimator(estimator estimation.Estimator) Option {
	return func(o *Options) {
		o.Estimator = estimator
	}
}

func WithUIConfig(uiConfig string) Option {
	return func(o *Options) {
		o.UIConfig = uiConfig
	}
}

func WithPipelineControllerAddress(address string) Option {
	return func(o *Options) {
		o.PipelineControllerAddress = address
	}
}

// WithCollectorServiceAccount configures the service account to use for explorer collector
func WithCollectorServiceAccount(name, namespace string) Option {
	return func(o *Options) {
		o.CollectorServiceAccount = collector.ImpersonateServiceAccount{
			Name:      name,
			Namespace: namespace,
		}
	}
}

// WithMonitoring configures monitoring server
func WithMonitoring(enabled bool, address string, metricsEnabled bool, profilingEnabled bool, log logr.Logger) Option {
	return func(o *Options) {
		o.MonitoringOptions = monitoring.Options{
			Enabled:       enabled,
			ServerAddress: address,
			Log:           log,
			MetricsOptions: metrics.Options{
				Enabled: metricsEnabled,
			},
			ProfilingOptions: profiling.Options{
				Enabled: profilingEnabled,
			},
		}
	}
}

// WithExplorerCleanerDisabled configures the object cleaner
func WithExplorerCleanerDisabled(disabled bool) Option {
	return func(o *Options) {
		o.ExplorerCleanerDisabled = disabled
	}
}

// ExplorerEnabledFor turns on and off the explorer for different parts of the UI
func WithExplorerEnabledFor(enabledFor []string) Option {
	return func(o *Options) {
		o.ExplorerEnabledFor = append(o.ExplorerEnabledFor, enabledFor...)
	}
}
