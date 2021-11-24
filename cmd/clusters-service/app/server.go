package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/go-logr/logr"
	grpc_runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/weaveworks/go-checkpoint"
	ent "github.com/weaveworks/weave-gitops-enterprise-credentials/pkg/entitlement"
	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	capi_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/server"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/version"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/utils"
	"github.com/weaveworks/weave-gitops-enterprise/common/entitlement"
	wego_proto "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	wego_server "github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"google.golang.org/grpc/metadata"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/klog/v2/klogr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func NewAPIServerCommand(log logr.Logger, tempDir string) *cobra.Command {
	var dbURI string
	var dbName string
	var dbUser string
	var dbPassword string
	var dbType string
	var dbBusyTimeout string
	var entitlementSecretName string
	var entitlementSecretNamespace string
	var profileHelmRepository string

	cmd := &cobra.Command{
		Use:          "capi-server",
		Long:         "The capi-server servers and handles REST operations for CAPI templates.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return StartServer(context.Background(), log, tempDir)
		},
	}

	cmd.Flags().StringVar(&dbURI, "db-uri", "/tmp/mccp.db", "URI of the database")
	cmd.Flags().StringVar(&dbType, "db-type", "sqlite", "database type, supported types [sqlite, postgres]")
	cmd.Flags().StringVar(&dbName, "db-name", os.Getenv("DB_NAME"), "database name, applicable if type is postgres")
	cmd.Flags().StringVar(&dbUser, "db-user", os.Getenv("DB_USER"), "database user")
	cmd.Flags().StringVar(&dbPassword, "db-password", os.Getenv("DB_PASSWORD"), "database password")
	cmd.Flags().StringVar(&dbBusyTimeout, "db-busy-timeout", "5000", "How long should sqlite wait when trying to write to the database")
	cmd.Flags().StringVar(&entitlementSecretName, "entitlement-secret-name", ent.DefaultSecretName, "The name of the entitlement secret")
	cmd.Flags().StringVar(&entitlementSecretNamespace, "entitlement-secret-namespace", ent.DefaultSecretNamespace, "The namespace of the entitlement secret")
	cmd.Flags().StringVar(&profileHelmRepository, "profile-helm-repository", "weaveworks-charts", "The name of the Flux `HelmRepository` object in the current namespace that references the profiles")

	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()
	viper.BindPFlags(cmd.Flags())

	return cmd
}

func StartServer(ctx context.Context, log logr.Logger, tempDir string) error {
	dbUri := viper.GetString("db-uri")
	dbType := viper.GetString("db-type")
	if dbType == "sqlite" {
		var err error
		dbUri, err = utils.GetSqliteUri(dbUri, viper.GetString("db-busy-timeout"))
		if err != nil {
			return err
		}
	}
	dbName := viper.GetString("db-name")
	dbUser := viper.GetString("db-user")
	dbPassword := viper.GetString("db-password")
	db, err := utils.Open(dbUri, dbType, dbName, dbUser, dbPassword)
	if err != nil {
		return err
	}

	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		v1.AddToScheme,
		capiv1.AddToScheme,
		sourcev1beta1.AddToScheme,
	}
	schemeBuilder.AddToScheme(scheme)
	kubeClientConfig := config.GetConfigOrDie()
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

	appsConfig, err := wego_server.DefaultConfig()
	if err != nil {
		return fmt.Errorf("could not create wego default config: %w", err)
	}
	// Override logger to ensure consistency
	appsConfig.Logger = log

	// Setup the flux binary needed by some weave-gitops code endpoints like adding apps
	flux.New(osys.New(), &runner.CLIRunner{}).SetupBin()

	return RunInProcessGateway(ctx, "0.0.0.0:8000",
		WithLog(log),
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
		WithGrpcRuntimeOptions(
			[]grpc_runtime.ServeMuxOption{
				grpc_runtime.WithIncomingHeaderMatcher(CustomIncomingHeaderMatcher),
				grpc_runtime.WithMetadata(TrackEvents(log)),
				middleware.WithGrpcErrorLogging(klogr.New()),
			},
		),
		WithCAPIClustersNamespace(ns),
		WithHelmRepositoryCacheDirectory(tempDir),
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
		return errors.New("Kubernetes client is not set")
	}
	if args.DiscoveryClient == nil {
		return errors.New("Kubernetes discovery client is not set")
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

	mux := grpc_runtime.NewServeMux(args.GrpcRuntimeOptions...)

	capi_proto.RegisterClustersServiceHandlerServer(ctx, mux, server.NewClusterServer(args.Log, args.TemplateLibrary, args.GitProvider, args.KubernetesClient, args.DiscoveryClient, args.Database, args.CAPIClustersNamespace, args.ProfileHelmRepository, args.HelmRepositoryCacheDirectory))

	//Add weave-gitops core handlers
	wegoServer := wego_server.NewApplicationsServer(args.ApplicationsConfig)
	wego_proto.RegisterApplicationsHandlerServer(ctx, mux, wegoServer)

	httpHandler := middleware.WithLogging(args.Log, mux)
	httpHandler = middleware.WithProviderToken(args.ApplicationsConfig.JwtClient, httpHandler, args.Log)
	httpHandler = entitlement.EntitlementHandler(ctx, args.Log, args.KubernetesClient, args.EntitlementSecretKey, entitlement.CheckEntitlementHandler(args.Log, httpHandler))

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
		Log:                   logr.Discard(),
		ProfileHelmRepository: viper.GetString("profile-helm-repository"),
		EntitlementSecretKey: client.ObjectKey{
			Name:      viper.GetString("entitlement-secret-name"),
			Namespace: viper.GetString("entitlement-secret-namespace"),
		},
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
