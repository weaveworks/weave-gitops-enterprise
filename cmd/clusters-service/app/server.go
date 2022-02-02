package app

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/go-logr/logr"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	grpc_runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/weaveworks/go-checkpoint"
	ent "github.com/weaveworks/weave-gitops-enterprise-credentials/pkg/entitlement"
	core_app_proto "github.com/weaveworks/weave-gitops/pkg/api/applications"
	core_profiles_proto "github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher/cache"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	core "github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/klog/v2/klogr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	capi_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/server"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/version"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/utils"
	"github.com/weaveworks/weave-gitops-enterprise/common/entitlement"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/handlers/agent"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/handlers/api"
)

// Options contains all the options for the `ui run` command.
type Params struct {
	dbURI                        string
	dbName                       string
	dbUser                       string
	dbPassword                   string
	dbType                       string
	dbBusyTimeout                string
	entitlementSecretName        string
	entitlementSecretNamespace   string
	helmRepoNamespace            string
	helmRepoName                 string
	profileCacheLocation         string
	watcherMetricsBindAddress    string
	watcherHealthzBindAddress    string
	watcherPort                  int
	AgentTemplateNatsURL         string
	AgentTemplateAlertmanagerURL string
}

func NewAPIServerCommand(log logr.Logger, tempDir string) *cobra.Command {
	p := Params{}
	cmd := &cobra.Command{
		Use:          "capi-server",
		Long:         "The capi-server servers and handles REST operations for CAPI templates.",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initializeConfig(cmd)
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
	cmd.Flags().StringVar(&p.entitlementSecretNamespace, "entitlement-secret-namespace", ent.DefaultSecretNamespace, "The namespace of the entitlement secret")
	cmd.Flags().StringVar(&p.helmRepoNamespace, "helm-repo-namespace", os.Getenv("RUNTIME_NAMESPACE"), "the namespace of the Helm Repository resource to scan for profiles")
	cmd.Flags().StringVar(&p.helmRepoName, "helm-repo-name", "weaveworks-charts", "the name of the Helm Repository resource to scan for profiles")
	cmd.Flags().StringVar(&p.profileCacheLocation, "profile-cache-location", "/tmp/helm-cache", "the location where the cache Profile data lives")
	cmd.Flags().StringVar(&p.watcherHealthzBindAddress, "watcher-healthz-bind-address", ":9981", "bind address for the healthz service of the watcher")
	cmd.Flags().StringVar(&p.watcherMetricsBindAddress, "watcher-metrics-bind-address", ":9980", "bind address for the metrics service of the watcher")
	cmd.Flags().IntVar(&p.watcherPort, "watcher-port", 9443, "the port on which the watcher is running")
	cmd.Flags().StringVar(&p.AgentTemplateAlertmanagerURL, "agent-template-alertmanager-url", "http://prometheus-operator-kube-p-alertmanager.wkp-prometheus:9093/api/v2", "Value used to populate the alertmanager URL in /api/agent.yaml")
	cmd.Flags().StringVar(&p.AgentTemplateNatsURL, "agent-template-nats-url", "nats://nats-client.wkp-gitops-repo-broker:4222", "Value used to populate the nats URL in /api/agent.yaml")

	return cmd
}

func initializeConfig(cmd *cobra.Command) error {
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
		sourcev1beta1.AddToScheme,
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
	ns := os.Getenv("CAPI_CLUSTERS_NAMESPACE")
	if ns == "" {
		return fmt.Errorf("environment variable %q cannot be empty", "CAPI_CLUSTERS_NAMESPACE")
	}

	appsConfig, err := core.DefaultApplicationsConfig()
	if err != nil {
		return fmt.Errorf("could not create wego default config: %w", err)
	}
	// Override logger to ensure consistency
	appsConfig.Logger = log

	// Setup the flux binary needed by some weave-gitops code endpoints like adding apps
	flux.New(osys.New(), &runner.CLIRunner{}).SetupBin()

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
		if err := profileWatcher.StartWatcher(); err != nil {
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
		WithTemplateLibrary(&templates.CRDLibrary{
			Log:       log,
			Client:    kubeClient,
			Namespace: os.Getenv("CAPI_TEMPLATES_NAMESPACE"),
		}),
		WithApplicationsConfig(appsConfig),
		WithProfilesConfig(core.NewProfilesConfig(kubeClient, profileCache, p.helmRepoNamespace, p.helmRepoName)),
		WithGrpcRuntimeOptions(
			[]grpc_runtime.ServeMuxOption{
				grpc_runtime.WithIncomingHeaderMatcher(CustomIncomingHeaderMatcher),
				grpc_runtime.WithMetadata(TrackEvents(log)),
				middleware.WithGrpcErrorLogging(klogr.New()),
			},
		),
		WithCAPIClustersNamespace(ns),
		WithHelmRepositoryCacheDirectory(tempDir),
		WithAgentTemplate(p.AgentTemplateNatsURL, p.AgentTemplateAlertmanagerURL),
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

	grpcMux := grpc_runtime.NewServeMux(args.GrpcRuntimeOptions...)

	clusterServer := server.NewClusterServer(
		args.Log,
		args.TemplateLibrary,
		args.GitProvider,
		args.KubernetesClient,
		args.DiscoveryClient,
		args.Database,
		args.CAPIClustersNamespace,
		args.ProfileHelmRepository,
		args.HelmRepositoryCacheDirectory,
	)
	if err := capi_proto.RegisterClustersServiceHandlerServer(ctx, grpcMux, clusterServer); err != nil {
		return fmt.Errorf("failed to register clusters service handler server: %w", err)
	}

	//Add weave-gitops core handlers
	wegoApplicationServer := core.NewApplicationsServer(args.ApplicationsConfig, args.ApplicationsOptions...)
	if err := core_app_proto.RegisterApplicationsHandlerServer(ctx, grpcMux, wegoApplicationServer); err != nil {
		return fmt.Errorf("failed to register application handler server: %w", err)
	}

	wegoProfilesServer := core.NewProfilesServer(args.ProfilesConfig)
	if err := core_profiles_proto.RegisterProfilesHandlerServer(ctx, grpcMux, wegoProfilesServer); err != nil {
		return fmt.Errorf("failed to register profiles handler server: %w", err)
	}

	grpcHttpHandler := middleware.WithLogging(args.Log, grpcMux)

	mux := http.NewServeMux()
	mux.Handle("/v1/", grpcHttpHandler)

	httpHandler := middleware.WithProviderToken(args.ApplicationsConfig.JwtClient, mux, args.Log)
	httpHandler = entitlement.EntitlementHandler(ctx, args.Log, args.KubernetesClient, args.EntitlementSecretKey, entitlement.CheckEntitlementHandler(args.Log, httpHandler))

	gitopsBrokerHandler := getGitopsBrokerMux(args.AgentTemplateNatsURL, args.AgentTemplateAlertmanagerURL, args.Database)
	mux.Handle("/gitops/", gitopsBrokerHandler)

	// UI
	var log = logrus.New()
	assetFS := getAssets()
	assetHandler := http.FileServer(http.FS(assetFS))
	redirector := createRedirector(assetFS, log)

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		extension := filepath.Ext(req.URL.Path)

		if extension == "" {
			redirector(w, req)
			return
		}
		assetHandler.ServeHTTP(w, req)
	}))

	s := &http.Server{
		Addr:    addr,
		Handler: httpHandler,
	}

	go func() {
		<-ctx.Done()
		args.Log.Info("Shutting down the http gateway server")
		if err := s.Shutdown(context.Background()); err != nil {
			args.Log.Error(err, "Failed to shutdown http gateway server")
		}
	}()

	args.Log.Info("Starting to listen and serve", "address", addr)

	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		args.Log.Error(err, "Failed to listen and serve")
		return err
	}
	return nil
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

func getGitopsBrokerMux(agentTemplateNatsURL, agentTemplateAlertmanagerURL string, db *gorm.DB) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/agent.yaml", agent.NewGetHandler(db, agentTemplateNatsURL, agentTemplateAlertmanagerURL)).Methods("GET")
	r.HandleFunc("/clusters", api.ListClusters(db, json.MarshalIndent)).Methods("GET")
	r.HandleFunc("/clusters/{id:[0-9]+}", api.FindCluster(db, json.MarshalIndent)).Methods("GET")
	r.HandleFunc("/clusters", api.RegisterCluster(db, validator.New(), json.Unmarshal, json.MarshalIndent, utils.Generate)).Methods("POST")
	r.HandleFunc("/clusters/{id:[0-9]+}", api.UpdateCluster(db, validator.New(), json.Unmarshal, json.MarshalIndent)).Methods("PUT")
	r.HandleFunc("/clusters/{id:[0-9]+}", api.UnregisterCluster(db)).Methods("DELETE")
	r.HandleFunc("/alerts", api.ListAlerts(db, json.MarshalIndent)).Methods("GET")

	return r
}

//go:embed dist/*
var static embed.FS

func getAssets() fs.FS {
	f, err := fs.Sub(static, "dist")

	if err != nil {
		panic(err)
	}

	return f
}

// A redirector ensures that index.html always gets served.
// The JS router will take care of actual navigation once the index.html page lands.
func createRedirector(fsys fs.FS, log logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		indexPage, err := fsys.Open("index.html")

		if err != nil {
			log.Error(err, "could not open index.html page")
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		stat, err := indexPage.Stat()
		if err != nil {
			log.Error(err, "could not get index.html stat")
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		bt := make([]byte, stat.Size())
		_, err = indexPage.Read(bt)

		if err != nil {
			log.Error(err, "could not read index.html")
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		_, err = w.Write(bt)

		if err != nil {
			log.Error(err, "error writing index.html")
			w.WriteHeader(http.StatusInternalServerError)

			return
		}
	}
}
