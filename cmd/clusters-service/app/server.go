package app

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/NYTimes/gziphandler"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-logr/logr"
	grpc_runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	"github.com/weaveworks/go-checkpoint"
	pipelinev1alpha1 "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	pacv1 "github.com/weaveworks/policy-agent/api/v1"
	pacv2beta1 "github.com/weaveworks/policy-agent/api/v2beta1"
	tfctrl "github.com/weaveworks/tf-controller/api/v1alpha1"
	ent "github.com/weaveworks/weave-gitops-enterprise-credentials/pkg/entitlement"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/namespaces"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/indexer"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/watcher"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm/watcher/cache"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/nsaccess"
	core_core "github.com/weaveworks/weave-gitops/core/server"
	core_app_proto "github.com/weaveworks/weave-gitops/pkg/api/applications"
	core_core_proto "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/featureflags"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	core "github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"google.golang.org/grpc/metadata"
	authv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeUtil "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	flaggerv1beta1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
	pd "github.com/weaveworks/progressive-delivery/pkg/server"
	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/capi/v1alpha1"
	gapiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/gitopstemplate/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/mgmtfetcher"
	capi_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	profiles_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos/profiles"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/server"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/version"
	"github.com/weaveworks/weave-gitops-enterprise/common/entitlement"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/fetcher"
	pipelines "github.com/weaveworks/weave-gitops-enterprise/pkg/pipelines/server"
	tfserver "github.com/weaveworks/weave-gitops-enterprise/pkg/terraform"
	wge_version "github.com/weaveworks/weave-gitops-enterprise/pkg/version"
	k8scache "k8s.io/client-go/tools/cache"
)

const (
	defaultConfigFilename = "config"

	// Allowed login requests per second
	loginRequestRateLimit = 20

	// resync for informers to guarantee that no event was missed
	sharedFactoryResync = 20 * time.Minute
)

var (
	ErrNoIssuerURL    = errors.New("the OIDC issuer URL flag (--oidc-issuer-url) has not been set")
	ErrNoClientID     = errors.New("the OIDC client ID flag (--oidc-client-id) has not been set")
	ErrNoClientSecret = errors.New("the OIDC client secret flag (--oidc-client-secret) has not been set")
	ErrNoRedirectURL  = errors.New("the OIDC redirect URL flag (--oidc-redirect-url) has not been set")
)

func EnterprisePublicRoutes() []string {
	return core.PublicRoutes
}

// Options contains all the options for the `ui run` command.
type Params struct {
	EntitlementSecretName             string                    `mapstructure:"entitlement-secret-name"`
	EntitlementSecretNamespace        string                    `mapstructure:"entitlement-secret-namespace"`
	HelmRepoNamespace                 string                    `mapstructure:"helm-repo-namespace"`
	HelmRepoName                      string                    `mapstructure:"helm-repo-name"`
	ProfileCacheLocation              string                    `mapstructure:"profile-cache-location"`
	WatcherMetricsBindAddress         string                    `mapstructure:"watcher-metrics-bind-address"`
	WatcherHealthzBindAddress         string                    `mapstructure:"watcher-healthz-bind-address"`
	WatcherPort                       int                       `mapstructure:"watcher-port"`
	HtmlRootPath                      string                    `mapstructure:"html-root-path"`
	OIDC                              OIDCAuthenticationOptions `mapstructure:",squash"`
	GitProviderType                   string                    `mapstructure:"git-provider-type"`
	GitProviderHostname               string                    `mapstructure:"git-provider-hostname"`
	CAPIClustersNamespace             string                    `mapstructure:"capi-clusters-namespace"`
	CAPITemplatesNamespace            string                    `mapstructure:"capi-templates-namespace"`
	InjectPruneAnnotation             string                    `mapstructure:"inject-prune-annotation"`
	AddBasesKustomization             string                    `mapstructure:"add-bases-kustomization"`
	CAPIEnabled                       bool                      `mapstructure:"capi-enabled"`
	CAPITemplatesRepositoryUrl        string                    `mapstructure:"capi-templates-repository-url"`
	CAPIRepositoryPath                string                    `mapstructure:"capi-repository-path"`
	CAPIRepositoryClustersPath        string                    `mapstructure:"capi-repository-clusters-path"`
	CAPITemplatesRepositoryApiUrl     string                    `mapstructure:"capi-templates-repository-api-url"`
	CAPITemplatesRepositoryBaseBranch string                    `mapstructure:"capi-templates-repository-base-branch"`
	RuntimeNamespace                  string                    `mapstructure:"runtime-namespace"`
	GitProviderToken                  string                    `mapstructure:"git-provider-token"`
	AuthMethods                       []string                  `mapstructure:"auth-methods"`
	TLSCert                           string                    `mapstructure:"tls-cert"`
	TLSKey                            string                    `mapstructure:"tls-key"`
	NoTLS                             bool                      `mapstructure:"no-tls"`
	DevMode                           bool                      `mapstructure:"dev-mode"`
	Cluster                           string                    `mapstructure:"cluster-name"`
	UseK8sCachedClients               bool                      `mapstructure:"use-k8s-cached-clients"`
}

type OIDCAuthenticationOptions struct {
	IssuerURL     string        `mapstructure:"oidc-issuer-url"`
	ClientID      string        `mapstructure:"oidc-client-id"`
	ClientSecret  string        `mapstructure:"oidc-client-secret"`
	RedirectURL   string        `mapstructure:"oidc-redirect-url"`
	TokenDuration time.Duration `mapstructure:"oidc-token-duration"`
}

func NewAPIServerCommand(log logr.Logger, tempDir string) *cobra.Command {
	p := &Params{}

	cmd := &cobra.Command{
		Use:          "capi-server",
		Version:      fmt.Sprintf("Version: %s, Image Tag: %s", version.Version, wge_version.ImageTag),
		Long:         "The capi-server servers and handles REST operations for CAPI templates.",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			err := initializeConfig(cmd)
			if err != nil {
				return fmt.Errorf("error initializing viper env, %w", err)
			}
			err = viper.Unmarshal(p)
			if err != nil {
				return fmt.Errorf("error unmarshalling flags and env into config struct %w", err)
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return checkParams(*p)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return StartServer(context.Background(), log, tempDir, *p)
		},
	}

	cmd.Flags().String("entitlement-secret-name", ent.DefaultSecretName, "The name of the entitlement secret")
	cmd.Flags().String("entitlement-secret-namespace", "flux-system", "The namespace of the entitlement secret")
	cmd.Flags().String("helm-repo-namespace", os.Getenv("RUNTIME_NAMESPACE"), "the namespace of the Helm Repository resource to scan for profiles")
	cmd.Flags().String("helm-repo-name", "weaveworks-charts", "the name of the Helm Repository resource to scan for profiles")
	cmd.Flags().String("profile-cache-location", "/tmp/helm-cache", "the location where the cache Profile data lives")
	cmd.Flags().String("watcher-healthz-bind-address", ":9981", "bind address for the healthz service of the watcher")
	cmd.Flags().String("watcher-metrics-bind-address", ":9980", "bind address for the metrics service of the watcher")
	cmd.Flags().Int("watcher-port", 9443, "the port on which the watcher is running")
	cmd.Flags().String("html-root-path", "/html", "Where to serve static assets from")
	cmd.Flags().String("git-provider-type", "", "")
	cmd.Flags().String("git-provider-hostname", "", "")
	cmd.Flags().Bool("capi-enabled", true, "")
	cmd.Flags().String("capi-clusters-namespace", corev1.NamespaceAll, "where to look for GitOps cluster resources, defaults to looking in all namespaces")
	cmd.Flags().String("capi-templates-namespace", "", "where to look for CAPI template resources, required")
	cmd.Flags().String("inject-prune-annotation", "", "")
	cmd.Flags().String("add-bases-kustomization", "enabled", "Add a kustomization to point to ./bases when creating leaf clusters")
	cmd.Flags().String("capi-templates-repository-url", "", "")
	cmd.Flags().String("capi-repository-path", "", "")
	cmd.Flags().String("capi-repository-clusters-path", "./clusters", "")
	cmd.Flags().String("capi-templates-repository-api-url", "", "")
	cmd.Flags().String("capi-templates-repository-base-branch", "", "")
	cmd.Flags().String("runtime-namespace", "flux-system", "Namespace hosting Gitops configuration objects (e.g. cluster-user-auth secrets)")
	cmd.Flags().String("git-provider-token", "", "")
	cmd.Flags().String("tls-cert-file", "", "filename for the TLS certficate, in-memory generated if omitted")
	cmd.Flags().String("tls-private-key", "", "filename for the TLS key, in-memory generated if omitted")
	cmd.Flags().Bool("no-tls", false, "do not attempt to read TLS certificates")
	cmd.Flags().String("cluster-name", "management", "name of the management cluster")

	cmd.Flags().StringSlice("auth-methods", []string{"oidc", "token-passthrough", "user-account"}, "Which auth methods to use, valid values are 'oidc', 'token-pass-through' and 'user-account'")
	cmd.Flags().String("oidc-issuer-url", "", "The URL of the OpenID Connect issuer")
	cmd.Flags().String("oidc-client-id", "", "The client ID for the OpenID Connect client")
	cmd.Flags().String("oidc-client-secret", "", "The client secret to use with OpenID Connect issuer")
	cmd.Flags().String("oidc-redirect-url", "", "The OAuth2 redirect URL")
	cmd.Flags().Duration("oidc-token-duration", time.Hour, "The duration of the ID token. It should be set in the format: number + time unit (s,m,h) e.g., 20m")

	cmd.Flags().Bool("dev-mode", false, "starts the server in development mode")
	cmd.Flags().Bool("use-k8s-cached-clients", true, "Enables the use of cached clients")

	return cmd
}

func checkParams(params Params) error {
	issuerURL := params.OIDC.IssuerURL
	clientID := params.OIDC.ClientID
	clientSecret := params.OIDC.ClientSecret
	redirectURL := params.OIDC.RedirectURL

	authMethods, err := auth.ParseAuthMethodArray(params.AuthMethods)
	if err != nil {
		return fmt.Errorf("could not parse auth methods while checking params: %w", err)
	}

	if !authMethods[auth.OIDC] {
		return nil
	}

	if issuerURL != "" || clientID != "" || clientSecret != "" || redirectURL != "" {
		if issuerURL == "" {
			return ErrNoIssuerURL
		}

		if clientID == "" {
			return ErrNoClientID
		}

		if clientSecret == "" {
			return ErrNoClientSecret
		}

		if redirectURL == "" {
			return ErrNoRedirectURL
		}
	}

	return nil
}

func initializeConfig(cmd *cobra.Command) error {
	// Set the base name of the config file, without the file extension.
	viper.SetConfigName(defaultConfigFilename)

	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// Align flag and env var names
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)

	// Read all env var values into viper
	viper.AutomaticEnv()

	// Read all flag values into viper
	// So they can be read from `viper.Get`, (sometimes user by weave-gitops (core))
	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		return err
	}

	return nil
}

func StartServer(ctx context.Context, log logr.Logger, tempDir string, p Params) error {

	featureflags.SetFromEnv(os.Environ())

	if p.CAPITemplatesNamespace == "" {
		return errors.New("CAPI templates namespace not set")
	}

	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		gapiv1.AddToScheme,
		sourcev1.AddToScheme,
		gitopsv1alpha1.AddToScheme,
		authv1.AddToScheme,
	}

	if p.CAPIEnabled {
		schemeBuilder = append(schemeBuilder, capiv1.AddToScheme)
	}

	err := schemeBuilder.AddToScheme(scheme)
	if err != nil {
		return err
	}
	kubeClientConfig, err := config.GetConfig()
	if err != nil {
		return err
	}
	kubernetesClientSet, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}
	kubeClient, err := client.New(kubeClientConfig, client.Options{Scheme: scheme})
	if err != nil {
		return err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(kubeClientConfig)
	if err != nil {
		return err
	}

	appsConfig, err := core.DefaultApplicationsConfig(log)
	if err != nil {
		return fmt.Errorf("could not create wego default config: %w", err)
	}

	profileCache, err := cache.NewCache(p.ProfileCacheLocation)
	if err != nil {
		return fmt.Errorf("failed to create cacher: %w", err)
	}

	profileWatcher, err := watcher.NewWatcher(watcher.Options{
		KubeClient:         kubeClient,
		Cache:              profileCache,
		MetricsBindAddress: p.WatcherMetricsBindAddress,
		HealthzBindAddress: p.WatcherHealthzBindAddress,
		WatcherPort:        p.WatcherPort,
	})
	if err != nil {
		return fmt.Errorf("failed to start the watcher: %w", err)
	}

	controllerContext := ctrl.SetupSignalHandler()

	go func() {
		if err := profileWatcher.StartWatcher(controllerContext, log); err != nil {
			log.Error(err, "failed to start profile watcher")
			os.Exit(1)
		}
	}()

	chartsCache, err := helm.NewChartIndexer(p.ProfileCacheLocation, p.Cluster)
	if err != nil {
		return fmt.Errorf("could not create charts cache: %w", err)
	}

	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	configGetter := kube.NewImpersonatingConfigGetter(kubeClientConfig, false)
	clientGetter := kube.NewDefaultClientGetter(configGetter, "",
		capiv1.AddToScheme,
		pacv1.AddToScheme,
		pacv2beta1.AddToScheme,
		gitopsv1alpha1.AddToScheme,
		clusterv1.AddToScheme,
		gapiv1.AddToScheme,
		pipelinev1alpha1.AddToScheme,
	)

	rest, clusterName, err := kube.RestConfig()
	if err != nil {
		return fmt.Errorf("could not retrieve cluster rest config: %w", err)
	}

	mcf, err := fetcher.NewMultiClusterFetcher(log, rest, clientGetter, p.CAPIClustersNamespace, p.Cluster)
	if err != nil {
		return err
	}

	clustersManagerScheme, err := kube.CreateScheme()
	if err != nil {
		return fmt.Errorf("could not create scheme: %w", err)
	}

	authMethods, err := auth.ParseAuthMethodArray(p.AuthMethods)
	if err != nil {
		return fmt.Errorf("could not parse auth methods: %w", err)
	}

	runtimeUtil.Must(pacv1.AddToScheme(clustersManagerScheme))
	runtimeUtil.Must(pacv2beta1.AddToScheme(clustersManagerScheme))
	runtimeUtil.Must(flaggerv1beta1.AddToScheme(clustersManagerScheme))
	runtimeUtil.Must(pipelinev1alpha1.AddToScheme(clustersManagerScheme))
	runtimeUtil.Must(tfctrl.AddToScheme(clustersManagerScheme))

	clientsFactory := clustersmngr.CachedClientFactory
	if !p.UseK8sCachedClients {
		log.Info("Using un-cached clients")
		clientsFactory = clustersmngr.ClientFactory
	} else {
		log.Info("Using cached clients")
	}

	clustersManager := clustersmngr.NewClustersManager(
		mcf,
		nsaccess.NewChecker(nsaccess.DefautltWegoAppRules),
		log,
		clustersManagerScheme,
		clientsFactory,
		clustersmngr.DefaultKubeConfigOptions,
	)

	indexer := indexer.NewClusterHelmIndexerTracker(chartsCache, p.Cluster, indexer.NewIndexer)
	go func() {
		err := indexer.Start(controllerContext, clustersManager, log)
		if err != nil {
			log.Error(err, "failed to start indexer")
			os.Exit(1)
		}
	}()

	clustersManager.Start(ctx)

	return RunInProcessGateway(ctx, "0.0.0.0:8000",
		WithLog(log),
		WithProfileHelmRepository(p.HelmRepoName),
		WithEntitlementSecretKey(client.ObjectKey{
			Name:      p.EntitlementSecretName,
			Namespace: p.EntitlementSecretNamespace,
		}),
		WithKubernetesClient(kubeClient),
		WithDiscoveryClient(discoveryClient),
		WithGitProvider(git.NewGitProviderService(log)),
		WithApplicationsConfig(appsConfig),
		WithCoreConfig(core_core.NewCoreConfig(
			log, rest, clusterName, clustersManager,
		)),
		WithProfilesConfig(server.NewProfilesConfig(kube.ClusterConfig{
			DefaultConfig: kubeClientConfig,
			ClusterName:   "",
		}, profileCache, p.HelmRepoNamespace, p.HelmRepoName)),
		WithGrpcRuntimeOptions(
			[]grpc_runtime.ServeMuxOption{
				grpc_runtime.WithIncomingHeaderMatcher(CustomIncomingHeaderMatcher),
				grpc_runtime.WithMetadata(TrackEvents(log)),
				middleware.WithGrpcErrorLogging(log),
			},
		),
		WithCAPIClustersNamespace(p.CAPIClustersNamespace),
		WithHelmRepositoryCacheDirectory(tempDir),
		WithHtmlRootPath(p.HtmlRootPath),
		WithClientGetter(clientGetter),
		WithAuthConfig(authMethods, p.OIDC),
		WithTLSConfig(p.TLSCert, p.TLSKey, p.NoTLS),
		WithCAPIEnabled(p.CAPIEnabled),
		WithRuntimeNamespace(p.RuntimeNamespace),
		WithDevMode(p.DevMode),
		WithClustersManager(clustersManager),
		WithChartsCache(chartsCache),
		WithKubernetesClientSet(kubernetesClientSet),
		WithManagementCluster(p.Cluster),
	)
}

// RunInProcessGateway starts the invoke in process http gateway.
func RunInProcessGateway(ctx context.Context, addr string, setters ...Option) error {
	args := defaultOptions()
	for _, setter := range setters {
		setter(args)
	}
	if args.KubernetesClient == nil {
		return errors.New("kubernetes client is not set")
	}
	if args.KubernetesClientSet == nil {
		return errors.New("kubernetes client set is not set")
	}
	if args.DiscoveryClient == nil {
		return errors.New("kubernetes discovery client is not set")
	}
	if args.GitProvider == nil {
		return errors.New("git provider is not set")
	}
	if args.ApplicationsConfig == nil {
		return errors.New("applications config is not set")
	}
	if args.ClientGetter == nil {
		return errors.New("kubernetes client getter is not set")
	}
	if args.CoreServerConfig.ClustersManager == nil {
		return errors.New("clusters manager is not set")
	}
	// TokenDuration at least should be set
	if (args.OIDC == OIDCAuthenticationOptions{}) {
		return errors.New("OIDC configuration is not set")
	}

	grpcMux := grpc_runtime.NewServeMux(args.GrpcRuntimeOptions...)

	factory := informers.NewSharedInformerFactory(args.KubernetesClientSet, sharedFactoryResync)
	namespacesCache := namespaces.NewNamespacesInformerCache(factory)
	authClientGetter := mgmtfetcher.NewUserConfigAuth(args.CoreServerConfig.RestCfg, args.Cluster)
	if args.ManagementFetcher == nil {
		args.ManagementFetcher = mgmtfetcher.NewManagementCrossNamespacesFetcher(namespacesCache, args.ClientGetter, authClientGetter)
	}

	// Add weave-gitops enterprise handlers
	clusterServer := server.NewClusterServer(
		server.ServerOpts{
			Logger:                    args.Log,
			ClustersManager:           args.CoreServerConfig.ClustersManager,
			GitProvider:               args.GitProvider,
			ClientGetter:              args.ClientGetter,
			DiscoveryClient:           args.DiscoveryClient,
			ClustersNamespace:         args.CAPIClustersNamespace,
			ProfileHelmRepositoryName: args.ProfileHelmRepository,
			HelmRepositoryCacheDir:    args.HelmRepositoryCacheDirectory,
			CAPIEnabled:               args.CAPIEnabled,
			ChartJobs:                 helm.NewJobs(),
			ChartsCache:               args.ChartsCache,
			ValuesFetcher:             helm.NewValuesFetcher(),
			RestConfig:                args.CoreServerConfig.RestCfg,
			ManagementFetcher:         args.ManagementFetcher,
			Cluster:                   args.Cluster,
		},
	)
	if err := capi_proto.RegisterClustersServiceHandlerServer(ctx, grpcMux, clusterServer); err != nil {
		return fmt.Errorf("failed to register clusters service handler server: %w", err)
	}

	//Add weave-gitops core handlers
	wegoApplicationServer := core.NewApplicationsServer(args.ApplicationsConfig, args.ApplicationsOptions...)
	if err := core_app_proto.RegisterApplicationsHandlerServer(ctx, grpcMux, wegoApplicationServer); err != nil {
		return fmt.Errorf("failed to register application handler server: %w", err)
	}

	wegoProfilesServer := server.NewProfilesServer(args.Log, args.ProfilesConfig)
	if err := profiles_proto.RegisterProfilesHandlerServer(ctx, grpcMux, wegoProfilesServer); err != nil {
		return fmt.Errorf("failed to register profiles handler server: %w", err)
	}

	// Add logging middleware
	grpcHttpHandler := middleware.WithLogging(args.Log, grpcMux)

	appsServer, err := core_core.NewCoreServer(args.CoreServerConfig)
	if err != nil {
		return fmt.Errorf("unable to create new kube client: %w", err)
	}

	if err = core_core_proto.RegisterCoreHandlerServer(ctx, grpcMux, appsServer); err != nil {
		return fmt.Errorf("could not register new app server: %w", err)
	}

	// Add progressive-delivery handlers
	if err := pd.Hydrate(ctx, grpcMux, pd.ServerOpts{
		ClustersManager: args.CoreServerConfig.ClustersManager,
		Logger:          args.Log,
	}); err != nil {
		return fmt.Errorf("failed to register progressive delivery handler server: %w", err)
	}

	if featureflags.Get("WEAVE_GITOPS_FEATURE_PIPELINES") != "" {
		if err := pipelines.Hydrate(ctx, grpcMux, pipelines.ServerOpts{
			ClustersManager:   args.ClustersManager,
			ManagementFetcher: args.ManagementFetcher,
			Cluster:           args.Cluster,
		}); err != nil {
			return fmt.Errorf("hydrating pipelines server: %w", err)
		}
	}

	if featureflags.Get("WEAVE_GITOPS_FEATURE_TERRAFORM_UI") != "" {
		if err := tfserver.Hydrate(ctx, grpcMux, tfserver.ServerOpts{
			Logger:         args.Log,
			ClientsFactory: args.ClustersManager,
			Scheme:         args.KubernetesClient.Scheme(),
		}); err != nil {
			return fmt.Errorf("hydrating terraform server: %w", err)
		}
	}

	// UI
	args.Log.Info("Attaching FileServer", "HtmlRootPath", args.HtmlRootPath)
	staticAssets := http.StripPrefix("/", http.FileServer(&spaFileSystem{http.Dir(args.HtmlRootPath)}))

	mux := http.NewServeMux()

	_, err = url.Parse(args.OIDC.IssuerURL)
	if err != nil {
		return fmt.Errorf("invalid issuer URL: %w", err)
	}

	_, err = url.Parse(args.OIDC.RedirectURL)
	if err != nil {
		return fmt.Errorf("invalid redirect URL: %w", err)
	}

	tsv, err := auth.NewHMACTokenSignerVerifier(args.OIDC.TokenDuration)
	if err != nil {
		return fmt.Errorf("could not create HMAC token signer: %w", err)
	}

	// FIXME: Slightly awkward bit of logging..
	authMethodsStrings := []string{}
	for authMethod, enabled := range args.AuthMethods {
		if enabled {
			authMethodsStrings = append(authMethodsStrings, authMethod.String())
		}
	}
	args.Log.Info("setting enabled auth methods", "enabled", authMethodsStrings)

	if args.DevMode {
		tsv.SetDevMode(args.DevMode)
	}

	authServerConfig, err := auth.NewAuthServerConfig(
		args.Log,
		auth.OIDCConfig{
			IssuerURL:     args.OIDC.IssuerURL,
			ClientID:      args.OIDC.ClientID,
			ClientSecret:  args.OIDC.ClientSecret,
			RedirectURL:   args.OIDC.RedirectURL,
			TokenDuration: args.OIDC.TokenDuration,
		},
		args.KubernetesClient,
		tsv,
		args.RuntimeNamespace,
		args.AuthMethods,
	)
	if err != nil {
		return fmt.Errorf("could not create auth server: %w", err)
	}
	srv, err := auth.NewAuthServer(
		ctx,
		authServerConfig,
	)
	if err != nil {
		return fmt.Errorf("could not create auth server: %w", err)
	}

	args.Log.Info("Registering callback route")
	if err := auth.RegisterAuthServer(mux, "/oauth2", srv, loginRequestRateLimit); err != nil {
		return fmt.Errorf("failed to register auth routes: %w", err)
	}

	// Secure `/v1` and `/gitops/api` API routes
	grpcHttpHandler = auth.WithAPIAuth(grpcHttpHandler, srv, EnterprisePublicRoutes())

	commonMiddleware := func(mux http.Handler) http.Handler {
		wrapperHandler := middleware.WithProviderToken(args.ApplicationsConfig.JwtClient, mux, args.Log)
		return entitlement.EntitlementHandler(
			ctx,
			args.Log,
			args.KubernetesClient,
			args.EntitlementSecretKey,
			entitlement.CheckEntitlementHandler(args.Log, wrapperHandler, EnterprisePublicRoutes()),
		)
	}

	mux.Handle("/v1/", commonMiddleware(grpcHttpHandler))

	staticAssetsWithGz := gziphandler.GzipHandler(staticAssets)

	mux.Handle("/", staticAssetsWithGz)

	s := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	factoryStopCh := make(chan struct{})
	factory.Start(factoryStopCh)
	k8scache.WaitForCacheSync(factoryStopCh,
		namespacesCache.CacheSync(),
	)

	go func() {
		<-ctx.Done()
		args.Log.Info("Shutting down the http gateway server")
		close(factoryStopCh)
		if err := s.Shutdown(context.Background()); err != nil {
			args.Log.Error(err, "Failed to shutdown http gateway server")
		}
	}()

	args.Log.Info("Starting to listen and serve", "address", addr)

	if err := ListenAndServe(s, args.NoTLS, args.TLSCert, args.TLSKey, args.Log); err != http.ErrServerClosed {
		args.Log.Error(err, "Failed to listen and serve")
		return err
	}
	return nil
}

func TLSConfig(hosts []string) (*tls.Config, error) {
	certPEMBlock, keyPEMBlock, err := generateKeyPair(hosts)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate TLS keys %w", err)
	}

	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate X509 key pair %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	return tlsConfig, nil
}

// Adapted from https://go.dev/src/crypto/tls/generate_cert.go
func generateKeyPair(hosts []string) ([]byte, []byte, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("Failing to generate new ecdsa key: %w", err)
	}

	// A CA is supposed to choose unique serial numbers, that is, unique for the CA.
	maxSerialNumber := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, maxSerialNumber)

	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate a random serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Weaveworks"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 365),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create certificate: %w", err)
	}

	certPEMBlock := &bytes.Buffer{}

	err = pem.Encode(certPEMBlock, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to encode cert pem: %w", err)
	}

	keyPEMBlock := &bytes.Buffer{}

	b, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to marshal ECDSA private key: %v", err)
	}

	err = pem.Encode(keyPEMBlock, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to encode key pem: %w", err)
	}

	return certPEMBlock.Bytes(), keyPEMBlock.Bytes(), nil
}

func ListenAndServe(srv *http.Server, noTLS bool, tlsCert, tlsKey string, log logr.Logger) error {
	if noTLS {
		log.Info("TLS connections disabled")
		return srv.ListenAndServe()
	}

	if tlsCert == "" && tlsKey == "" {
		log.Info("TLS cert and key not specified, generating and using in-memory keys")

		tlsConfig, err := TLSConfig([]string{"localhost", "0.0.0.0", "127.0.0.1", "weave.gitops.enterprise.com"})
		if err != nil {
			return fmt.Errorf("failed to generate a TLSConfig: %w", err)
		}

		srv.TLSConfig = tlsConfig
		// if TLSCert and TLSKey are both empty (""), ListenAndServeTLS will ignore
		// and happily use the TLSConfig supplied above
		return srv.ListenAndServeTLS("", "")
	}

	if tlsCert == "" || tlsKey == "" {
		return cmderrors.ErrNoTLSCertOrKey
	}

	log.Info("Using TLS", "cert", tlsCert, "key", tlsKey)

	return srv.ListenAndServeTLS(tlsCert, tlsKey)
}

// CustomIncomingHeaderMatcher allows the Accept header to be passed to the gRPC handlers.
// The Accept header is used by the gRPC handlers to determine whether a response other
// than `application/json` is requested.
func CustomIncomingHeaderMatcher(key string) (string, bool) {
	switch key {
	case "Accept":
		return key, true
	default:
		return grpc_runtime.DefaultHeaderMatcher(key)
	}
}

// TrackEvents tracks data for specific operations.
func TrackEvents(log logr.Logger) func(ctx context.Context, r *http.Request) metadata.MD {
	return func(ctx context.Context, r *http.Request) metadata.MD {
		var handler string
		md := make(map[string]string)
		if method, ok := grpc_runtime.RPCMethod(ctx); ok {
			md["method"] = method
			handler = method
		}
		if pattern, ok := grpc_runtime.HTTPPathPattern(ctx); ok {
			md["pattern"] = pattern
		}

		track(log, handler)

		return metadata.New(md)
	}
}

func defaultOptions() *Options {
	return &Options{
		Log: logr.Discard(),
	}
}

func track(log logr.Logger, handler string) {
	handlers := make(map[string]map[string]string)
	handlers["ListTemplates"] = map[string]string{
		"object":  "templates",
		"command": "list",
	}
	handlers["CreatePullRequest"] = map[string]string{
		"object":  "clusters",
		"command": "create",
	}
	handlers["DeleteClustersPullRequest"] = map[string]string{
		"object":  "clusters",
		"command": "delete",
	}

	for h, m := range handlers {
		if strings.HasSuffix(handler, h) {
			go checkVersionWithFlags(log, m)
		}
	}
}

func checkVersionWithFlags(log logr.Logger, flags map[string]string) {
	p := &checkpoint.CheckParams{
		Product: "weave-gitops-enterprise",
		Version: version.Version,
		Flags:   flags,
	}
	checkResponse, err := checkpoint.Check(p)
	if err != nil {
		log.Error(err, "Failed to check version")
		return
	}
	if checkResponse.Outdated {
		log.Info("There is a newer version of weave-gitops-enterprise available",
			"latest", checkResponse.CurrentVersion, "url", checkResponse.CurrentDownloadURL)
	} else {
		log.Info("The current weave-gitops-enterprise version is up to date", "current", version.Version)
	}
}

type spaFileSystem struct {
	root http.FileSystem
}

func (fs *spaFileSystem) Open(name string) (http.File, error) {
	f, err := fs.root.Open(name)
	if os.IsNotExist(err) {
		return fs.root.Open("index.html")
	}
	return f, err
}
