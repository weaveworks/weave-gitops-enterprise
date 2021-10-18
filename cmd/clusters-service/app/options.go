package app

import (
	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
	"github.com/weaveworks/weave-gitops/pkg/server"
	"gorm.io/gorm"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Options struct {
	Log                logr.Logger
	Database           *gorm.DB
	KubernetesClient   client.Client
	DiscoveryClient    discovery.DiscoveryInterface
	TemplateLibrary    templates.Library
	GitProvider        git.Provider
	ApplicationsConfig *server.ApplicationsConfig
	GrpcRuntimeOptions []runtime.ServeMuxOption
}

type Option func(*Options)

func WithLog(log logr.Logger) Option {
	return func(o *Options) {
		o.Log = log
	}
}

func WithDatabase(database *gorm.DB) Option {
	return func(o *Options) {
		o.Database = database
	}
}

func WithKubernetesClient(client client.Client) Option {
	return func(o *Options) {
		o.KubernetesClient = client
	}
}

func WithDiscoveryClient(client discovery.DiscoveryInterface) Option {
	return func(o *Options) {
		o.DiscoveryClient = client
	}
}

func WithTemplateLibrary(templateLibrary templates.Library) Option {
	return func(o *Options) {
		o.TemplateLibrary = templateLibrary
	}
}

func WithGitProvider(gitProvider git.Provider) Option {
	return func(o *Options) {
		o.GitProvider = gitProvider
	}
}

func WithApplicationsConfig(appConfig *server.ApplicationsConfig) Option {
	return func(o *Options) {
		o.ApplicationsConfig = appConfig
	}
}

func WithGrpcRuntimeOptions(grpcRuntimeOptions []runtime.ServeMuxOption) Option {
	return func(o *Options) {
		o.GrpcRuntimeOptions = grpcRuntimeOptions
	}
}
