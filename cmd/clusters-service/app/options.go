package app

import (
	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops/pkg/server"
	"gorm.io/gorm"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
)

// Options specifies the options that can be set
// in RunInProcessGateway.
type Options struct {
	Log                          logr.Logger
	Database                     *gorm.DB
	KubernetesClient             client.Client
	DiscoveryClient              discovery.DiscoveryInterface
	TemplateLibrary              templates.Library
	GitProvider                  git.Provider
	ApplicationsConfig           *server.ApplicationsConfig
	ProfilesConfig               server.ProfilesConfig
	GrpcRuntimeOptions           []runtime.ServeMuxOption
	HelmRepositoryCacheDirectory string
	CAPIClustersNamespace        string
	EntitlementSecretKey         client.ObjectKey
}

type Option func(*Options)

// WithLog is used to set a logger.
func WithLog(log logr.Logger) Option {
	return func(o *Options) {
		o.Log = log
	}
}

// WithDatabase is used to set a database.
func WithDatabase(database *gorm.DB) Option {
	return func(o *Options) {
		o.Database = database
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

// WithProfilesConfig is used to set the configuration needed to work
// with Weave GitOps Core profiles
func WithProfilesConfig(profilesConfig server.ProfilesConfig) Option {
	return func(o *Options) {
		o.ProfilesConfig = profilesConfig
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
