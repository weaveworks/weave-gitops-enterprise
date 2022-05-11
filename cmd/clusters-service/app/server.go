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
	"encoding/json"
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

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	grpc_runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	"github.com/weaveworks/go-checkpoint"
	policiesv1 "github.com/weaveworks/policy-agent/api/v1"
	ent "github.com/weaveworks/weave-gitops-enterprise-credentials/pkg/entitlement"
	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/clusters"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	capi_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/server"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/version"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/utils"
	"github.com/weaveworks/weave-gitops-enterprise/common/entitlement"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/fetcher"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/handlers/agent"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/handlers/api"
	wge_version "github.com/weaveworks/weave-gitops-enterprise/pkg/version"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/nsaccess"
	core_core "github.com/weaveworks/weave-gitops/core/server"
	core_app_proto "github.com/weaveworks/weave-gitops/pkg/api/applications"
	core_core_proto "github.com/weaveworks/weave-gitops/pkg/api/core"
	core_profiles_proto "github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher/cache"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	core "github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	AuthEnabledFeatureFlag = "WEAVE_GITOPS_AUTH_ENABLED"

	defaultConfigFilename = "config"

	// Allowed login requests per second
	loginRequestRateLimit = 20
)

var (
	ErrNoIssuerURL    = errors.New("the OIDC issuer URL flag (--oidc-issuer-url) has not been set")
	ErrNoClientID     = errors.New("the OIDC client ID flag (--oidc-client-id) has not been set")
	ErrNoClientSecret = errors.New("the OIDC client secret flag (--oidc-client-secret) has not been set")
	ErrNoRedirectURL  = errors.New("the OIDC redirect URL flag (--oidc-redirect-url) has not been set")
)

func AuthEnabled() bool {
	return os.Getenv(AuthEnabledFeatureFlag) == "true"
}

func EnterprisePublicRoutes() []string {
	return append(core.PublicRoutes, "/gitops/api/agent.yaml")
}

// Options contains all the options for the `ui run` command.
type Params struct {
	dbURI                             string
	dbName                            string
	dbUser                            string
	dbPassword                        string
	dbType                            string
	dbBusyTimeout                     string
	entitlementSecretName             string
	entitlementSecretNamespace        string
	helmRepoNamespace                 string
	helmRepoName                      string
	profileCacheLocation              string
	watcherMetricsBindAddress         string
	watcherHealthzBindAddress         string
	watcherPort                       int
	AgentTemplateNatsURL              string
	AgentTemplateAlertmanagerURL      string
	htmlRootPath                      string
	OIDC                              OIDCAuthenticationOptions
	gitProviderType                   string
	gitProviderHostname               string
	capiClustersNamespace             string
	capiTemplatesNamespace            string
	injectPruneAnnotation             string
	addBasesKustomization             string
	capiTemplatesRepositoryUrl        string
	capiRepositoryPath                string
	capiRepositoryClustersPath        string
	capiTemplatesRepositoryApiUrl     string
	capiTemplatesRepositoryBaseBranch string
	runtimeNamespace                  string
	gitProviderToken                  string
	TLSCert                           string
	TLSKey                            string
	NoTLS                             bool
}

type OIDCAuthenticationOptions struct {
	IssuerURL     string
	ClientID      string
	ClientSecret  string
	RedirectURL   string
	TokenDuration time.Duration
}

func NewAPIServerCommand(log logr.Logger, tempDir string) *cobra.Command {
	p := Params{}
	cmd := &cobra.Command{
		Use:          "capi-server",
		Version:      fmt.Sprintf("Version: %s, Image Tag: %s", version.Version, wge_version.ImageTag),
		Long:         "The capi-server servers and handles REST operations for CAPI templates.",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initializeConfig(cmd)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return checkParams(p)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return StartServer(context.Background(), log, tempDir, p)
		},
	}

	cmd.Flags().StringVar(&p.dbURI, "db-uri", "/tmp/mccp.db", "URI of the database")
	cmd.Flags().StringVar(&p.dbType, "db-type", "sqlite", "database type, supported types [sqlite, postgres]")
	cmd.Flags().StringVar(&p.dbName, "db-name", "", "database name, applicable if type is postgres")
	cmd.Flags().StringVar(&p.dbUser, "db-user", "", "database user")
	cmd.Flags().StringVar(&p.dbPassword, "db-password", "", "database password")
	cmd.Flags().StringVar(&p.dbBusyTimeout, "db-busy-timeout", "5000", "How long should sqlite wait when trying to write to the database")
	cmd.Flags().StringVar(&p.entitlementSecretName, "entitlement-secret-name", ent.DefaultSecretName, "The name of the entitlement secret")
	cmd.Flags().StringVar(&p.entitlementSecretNamespace, "entitlement-secret-namespace", "flux-system", "The namespace of the entitlement secret")
	cmd.Flags().StringVar(&p.helmRepoNamespace, "helm-repo-namespace", os.Getenv("RUNTIME_NAMESPACE"), "the namespace of the Helm Repository resource to scan for profiles")
	cmd.Flags().StringVar(&p.helmRepoName, "helm-repo-name", "weaveworks-charts", "the name of the Helm Repository resource to scan for profiles")
	cmd.Flags().StringVar(&p.profileCacheLocation, "profile-cache-location", "/tmp/helm-cache", "the location where the cache Profile data lives")
	cmd.Flags().StringVar(&p.watcherHealthzBindAddress, "watcher-healthz-bind-address", ":9981", "bind address for the healthz service of the watcher")
	cmd.Flags().StringVar(&p.watcherMetricsBindAddress, "watcher-metrics-bind-address", ":9980", "bind address for the metrics service of the watcher")
	cmd.Flags().IntVar(&p.watcherPort, "watcher-port", 9443, "the port on which the watcher is running")
	cmd.Flags().StringVar(&p.AgentTemplateAlertmanagerURL, "agent-template-alertmanager-url", "http://prometheus-operator-kube-p-alertmanager.wkp-prometheus:9093/api/v2", "Value used to populate the alertmanager URL in /api/agent.yaml")
	cmd.Flags().StringVar(&p.AgentTemplateNatsURL, "agent-template-nats-url", "nats://nats-client.flux-system:4222", "Value used to populate the nats URL in /api/agent.yaml")
	cmd.Flags().StringVar(&p.htmlRootPath, "html-root-path", "/html", "Where to serve static assets from")
	cmd.Flags().StringVar(&p.gitProviderType, "git-provider-type", "", "")
	cmd.Flags().StringVar(&p.gitProviderHostname, "git-provider-hostname", "", "")
	cmd.Flags().StringVar(&p.capiClustersNamespace, "capi-clusters-namespace", "", "")
	cmd.Flags().StringVar(&p.capiTemplatesNamespace, "capi-templates-namespace", "", "")
	cmd.Flags().StringVar(&p.injectPruneAnnotation, "inject-prune-annotation", "", "")
	cmd.Flags().StringVar(&p.addBasesKustomization, "add-bases-kustomization", "enabled", "Add a kustomization to point to ./bases when creating leaf clusters")
	cmd.Flags().StringVar(&p.capiTemplatesRepositoryUrl, "capi-templates-repository-url", "", "")
	cmd.Flags().StringVar(&p.capiRepositoryPath, "capi-repository-path", "", "")
	cmd.Flags().StringVar(&p.capiRepositoryClustersPath, "capi-repository-clusters-path", "./clusters", "")
	cmd.Flags().StringVar(&p.capiTemplatesRepositoryApiUrl, "capi-templates-repository-api-url", "", "")
	cmd.Flags().StringVar(&p.capiTemplatesRepositoryBaseBranch, "capi-templates-repository-base-branch", "", "")
	cmd.Flags().StringVar(&p.runtimeNamespace, "runtime-namespace", "", "")
	cmd.Flags().StringVar(&p.gitProviderToken, "git-provider-token", "", "")

	cmd.Flags().StringVar(&p.TLSCert, "tls-cert-file", "", "filename for the TLS certficate, in-memory generated if omitted")
	cmd.Flags().StringVar(&p.TLSKey, "tls-private-key", "", "filename for the TLS key, in-memory generated if omitted")
	cmd.Flags().BoolVar(&p.NoTLS, "no-tls", false, "do not attempt to read TLS certificates")

	if AuthEnabled() {
		cmd.Flags().StringVar(&p.OIDC.IssuerURL, "oidc-issuer-url", "", "The URL of the OpenID Connect issuer")
		cmd.Flags().StringVar(&p.OIDC.ClientID, "oidc-client-id", "", "The client ID for the OpenID Connect client")
		cmd.Flags().StringVar(&p.OIDC.ClientSecret, "oidc-client-secret", "", "The client secret to use with OpenID Connect issuer")
		cmd.Flags().StringVar(&p.OIDC.RedirectURL, "oidc-redirect-url", "", "The OAuth2 redirect URL")
		cmd.Flags().DurationVar(&p.OIDC.TokenDuration, "oidc-token-duration", time.Hour, "The duration of the ID token. It should be set in the format: number + time unit (s,m,h) e.g., 20m")
	}

	return cmd
}

func checkParams(params Params) error {
	issuerURL := params.OIDC.IssuerURL
	clientID := params.OIDC.ClientID
	clientSecret := params.OIDC.ClientSecret
	redirectURL := params.OIDC.RedirectURL

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

	// Set all unset flags values to their associated env vars value if env var is present
	bindFlagValues(cmd)

	return nil
}

func bindFlagValues(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && viper.IsSet(f.Name) {
			val := viper.Get(f.Name)
			_ = cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

func StartServer(ctx context.Context, log logr.Logger, tempDir string, p Params) error {
	dbUri := p.dbURI
	if p.dbType == "sqlite" {
		var err error
		dbUri, err = utils.GetSqliteUri(dbUri, p.dbBusyTimeout)
		if err != nil {
			return err
		}
	}
	db, err := utils.Open(dbUri, p.dbType, p.dbName, p.dbUser, p.dbPassword)
	if err != nil {
		return err
	}

	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		v1.AddToScheme,
		capiv1.AddToScheme,
		sourcev1.AddToScheme,
		gitopsv1alpha1.AddToScheme,
	}

	err = schemeBuilder.AddToScheme(scheme)
	if err != nil {
		return err
	}
	kubeClientConfig, err := config.GetConfig()
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
	ns := p.capiClustersNamespace
	if ns == "" {
		return fmt.Errorf("environment variable %q cannot be empty", "CAPI_CLUSTERS_NAMESPACE")
	}

	appsConfig, err := core.DefaultApplicationsConfig(log)
	if err != nil {
		return fmt.Errorf("could not create wego default config: %w", err)
	}

	profileCache, err := cache.NewCache(p.profileCacheLocation)
	if err != nil {
		return fmt.Errorf("failed to create cacher: %w", err)
	}

	profileWatcher, err := watcher.NewWatcher(watcher.Options{
		KubeClient:         kubeClient,
		Cache:              profileCache,
		MetricsBindAddress: p.watcherMetricsBindAddress,
		HealthzBindAddress: p.watcherHealthzBindAddress,
		WatcherPort:        p.watcherPort,
	})
	if err != nil {
		return fmt.Errorf("failed to start the watcher: %w", err)
	}

	go func() {
		if err := profileWatcher.StartWatcher(log); err != nil {
			log.Error(err, "failed to start profile watcher")
			os.Exit(1)
		}
	}()

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
		policiesv1.AddToScheme,
		gitopsv1alpha1.AddToScheme,
		clusterv1.AddToScheme,
	)

	rest, clusterName, err := kube.RestConfig()
	if err != nil {
		return fmt.Errorf("could not retrieve cluster rest config: %w", err)
	}

	mcf, err := fetcher.NewMultiClusterFetcher(log, rest, clientGetter, p.capiTemplatesNamespace)
	if err != nil {
		return err
	}

	clusterClientsFactory := clustersmngr.NewClientFactory(
		mcf,
		nsaccess.NewChecker(nsaccess.DefautltWegoAppRules),
		log,
	)
	clusterClientsFactory.Start(ctx)

	return RunInProcessGateway(ctx, "0.0.0.0:8000",
		WithLog(log),
		WithProfileHelmRepository(p.helmRepoName),
		WithEntitlementSecretKey(client.ObjectKey{
			Name:      p.entitlementSecretName,
			Namespace: p.entitlementSecretNamespace,
		}),
		WithDatabase(db),
		WithKubernetesClient(kubeClient),
		WithDiscoveryClient(discoveryClient),
		WithGitProvider(git.NewGitProviderService(log)),
		WithClustersLibrary(&clusters.CRDLibrary{
			Log:          log,
			ClientGetter: clientGetter,
			Namespace:    p.capiClustersNamespace,
		}),
		WithTemplateLibrary(&templates.CRDLibrary{
			Log:          log,
			ClientGetter: clientGetter,
			Namespace:    p.capiTemplatesNamespace,
		}),
		WithApplicationsConfig(appsConfig),
		WithCoreConfig(core_core.NewCoreConfig(
			log, rest, clusterName, clusterClientsFactory,
		)),
		WithProfilesConfig(core.NewProfilesConfig(kube.ClusterConfig{
			DefaultConfig: kubeClientConfig,
			ClusterName:   "",
		}, profileCache, p.helmRepoNamespace, p.helmRepoName)),
		WithGrpcRuntimeOptions(
			[]grpc_runtime.ServeMuxOption{
				grpc_runtime.WithIncomingHeaderMatcher(CustomIncomingHeaderMatcher),
				grpc_runtime.WithMetadata(TrackEvents(log)),
				middleware.WithGrpcErrorLogging(log),
			},
		),
		WithCAPIClustersNamespace(ns),
		WithHelmRepositoryCacheDirectory(tempDir),
		WithAgentTemplate(p.AgentTemplateNatsURL, p.AgentTemplateAlertmanagerURL),
		WithHtmlRootPath(p.htmlRootPath),
		WithClientGetter(clientGetter),
		WithOIDCConfig(p.OIDC),
		WithTLSConfig(p.TLSCert, p.TLSKey, p.NoTLS),
	)
}

// RunInProcessGateway starts the invoke in process http gateway.
func RunInProcessGateway(ctx context.Context, addr string, setters ...Option) error {
	args := defaultOptions()
	for _, setter := range setters {
		setter(args)
	}
	if args.Database == nil {
		return errors.New("database is not set")
	}
	if args.KubernetesClient == nil {
		return errors.New("kubernetes client is not set")
	}
	if args.DiscoveryClient == nil {
		return errors.New("kubernetes discovery client is not set")
	}
	if args.TemplateLibrary == nil {
		return errors.New("template library is not set")
	}
	if args.GitProvider == nil {
		return errors.New("git provider is not set")
	}
	if args.ApplicationsConfig == nil {
		return errors.New("applications config is not set")
	}
	if args.CAPIClustersNamespace == "" {
		return errors.New("CAPI clusters namespace is not set")
	}
	if args.ClientGetter == nil {
		return errors.New("kubernetes client getter is not set")
	}
	if args.CoreServerConfig.ClientsFactory == nil {
		return errors.New("clients factory is not set")
	}
	if (AuthEnabled() && args.OIDC == OIDCAuthenticationOptions{}) {
		return errors.New("OIDC configuration is not set")
	}

	grpcMux := grpc_runtime.NewServeMux(args.GrpcRuntimeOptions...)

	clientset, err := kubernetes.NewForConfig(config.GetConfigOrDie())
	if err != nil {
		return errors.New("failed to create clientset")
	}
	// Add weave-gitops enterprise handlers
	clusterServer := server.NewClusterServer(
		args.Log,
		args.ClustersLibrary,
		args.TemplateLibrary,
		args.GitProvider,
		args.ClientGetter,
		args.DiscoveryClient,
		args.Database,
		args.CAPIClustersNamespace,
		args.ProfileHelmRepository,
		args.HelmRepositoryCacheDirectory,
		clientset,
	)
	if err := capi_proto.RegisterClustersServiceHandlerServer(ctx, grpcMux, clusterServer); err != nil {
		return fmt.Errorf("failed to register clusters service handler server: %w", err)
	}

	//Add weave-gitops core handlers
	wegoApplicationServer := core.NewApplicationsServer(args.ApplicationsConfig, args.ApplicationsOptions...)
	if err := core_app_proto.RegisterApplicationsHandlerServer(ctx, grpcMux, wegoApplicationServer); err != nil {
		return fmt.Errorf("failed to register application handler server: %w", err)
	}

	wegoProfilesServer := core.NewProfilesServer(args.Log, args.ProfilesConfig)
	if err := core_profiles_proto.RegisterProfilesHandlerServer(ctx, grpcMux, wegoProfilesServer); err != nil {
		return fmt.Errorf("failed to register profiles handler server: %w", err)
	}

	// Add logging middleware
	grpcHttpHandler := middleware.WithLogging(args.Log, grpcMux)

	appsServer, err := core_core.NewCoreServer(args.CoreServerConfig, core_core.WithClientGetter(args.ClientGetter))
	if err != nil {
		return fmt.Errorf("unable to create new kube client: %w", err)
	}

	if err = core_core_proto.RegisterCoreHandlerServer(ctx, grpcMux, appsServer); err != nil {
		return fmt.Errorf("could not register new app server: %w", err)
	}

	grpcHttpHandler = clustersmngr.WithClustersClient(args.CoreServerConfig.ClientsFactory, grpcHttpHandler)

	gitopsBrokerHandler := getGitopsBrokerMux(args.AgentTemplateNatsURL, args.AgentTemplateAlertmanagerURL, args.Database)

	// UI
	args.Log.Info("Attaching FileServer", "HtmlRootPath", args.HtmlRootPath)
	staticAssets := http.StripPrefix("/", http.FileServer(&spaFileSystem{http.Dir(args.HtmlRootPath)}))

	mux := http.NewServeMux()

	if AuthEnabled() {
		_, err := url.Parse(args.OIDC.IssuerURL)
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
		gitopsBrokerHandler = auth.WithAPIAuth(gitopsBrokerHandler, srv, EnterprisePublicRoutes())
	}

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
	mux.Handle("/gitops/api/", commonMiddleware(gitopsBrokerHandler))

	mux.Handle("/", staticAssets)

	s := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		args.Log.Info("Shutting down the http gateway server")
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

func getGitopsBrokerMux(agentTemplateNatsURL, agentTemplateAlertmanagerURL string, db *gorm.DB) http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/gitops/api/agent.yaml", agent.NewGetHandler(db, agentTemplateNatsURL, agentTemplateAlertmanagerURL)).Methods("GET")
	r.HandleFunc("/gitops/api/clusters", api.ListClusters(db, json.MarshalIndent)).Methods("GET")
	r.HandleFunc("/gitops/api/clusters/{id:[0-9]+}", api.FindCluster(db, json.MarshalIndent)).Methods("GET")
	r.HandleFunc("/gitops/api/clusters", api.RegisterCluster(db, validator.New(), json.Unmarshal, json.MarshalIndent, utils.Generate)).Methods("POST")
	r.HandleFunc("/gitops/api/clusters/{id:[0-9]+}", api.UpdateCluster(db, validator.New(), json.Unmarshal, json.MarshalIndent)).Methods("PUT")
	r.HandleFunc("/gitops/api/clusters/{id:[0-9]+}", api.UnregisterCluster(db)).Methods("DELETE")
	r.HandleFunc("/gitops/api/alerts", api.ListAlerts(db, json.MarshalIndent)).Methods("GET")

	return r
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
