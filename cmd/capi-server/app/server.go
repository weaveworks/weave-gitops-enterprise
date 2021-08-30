package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-logr/logr"
	grpc_runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/weaveworks/go-checkpoint"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/utils"
	"github.com/weaveworks/weave-gitops/pkg/middleware"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/klog/v2/klogr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/capi-server/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/capi-server/pkg/git"
	capi_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/capi-server/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/capi-server/pkg/server"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/capi-server/pkg/templates"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/capi-server/pkg/version"
	wego_proto "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	wego_server "github.com/weaveworks/weave-gitops/pkg/server"
)

func NewAPIServerCommand() *cobra.Command {
	var dbURI string
	var dbName string
	var dbUser string
	var dbPassword string
	var dbType string
	var dbBusyTimeout string

	cmd := &cobra.Command{
		Use:          "capi-server",
		Long:         "The capi-server servers and handles REST operations for CAPI templates.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return StartServer(context.Background())
		},
	}

	cmd.Flags().StringVar(&dbURI, "db-uri", "/tmp/mccp.db", "URI of the database")
	cmd.Flags().StringVar(&dbType, "db-type", "sqlite", "database type, supported types [sqlite, postgres]")
	cmd.Flags().StringVar(&dbName, "db-name", os.Getenv("DB_NAME"), "database name, applicable if type is postgres")
	cmd.Flags().StringVar(&dbUser, "db-user", os.Getenv("DB_USER"), "database user")
	cmd.Flags().StringVar(&dbPassword, "db-password", os.Getenv("DB_PASSWORD"), "database password")
	cmd.Flags().StringVar(&dbBusyTimeout, "db-busy-timeout", "5000", "How long should sqlite wait when trying to write to the database")

	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()
	viper.BindPFlags(cmd.Flags())

	return cmd
}

func StartServer(ctx context.Context) error {
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
	library := &templates.CRDLibrary{
		Client:    kubeClient,
		Namespace: os.Getenv("CAPI_TEMPLATES_NAMESPACE"),
	}
	ns := os.Getenv("CAPI_CLUSTERS_NAMESPACE")
	if ns == "" {
		return fmt.Errorf("environment variable %q cannot be empty", "CAPI_CLUSTERS_NAMESPACE")
	}
	provider := git.NewGitProviderService()
	kube, err := kube.NewKubeHTTPClient()
	if err != nil {
		return fmt.Errorf("could not create kube http client: %w", err)
	}
	return RunInProcessGateway(ctx, "0.0.0.0:8000", library, provider, kubeClient, discoveryClient, db, ns, kube,
		grpc_runtime.WithIncomingHeaderMatcher(CustomIncomingHeaderMatcher),
		grpc_runtime.WithMetadata(TrackEvents),
		middleware.WithGrpcErrorLogging(klogr.New()),
	)
}

// RunInProcessGateway starts the invoke in process http gateway.
func RunInProcessGateway(ctx context.Context, addr string, library templates.Library, provider git.Provider, client client.Client, discoveryClient discovery.DiscoveryInterface, db *gorm.DB, ns string, kube kube.Kube, opts ...grpc_runtime.ServeMuxOption) error {
	mux := grpc_runtime.NewServeMux(opts...)

	capi_proto.RegisterClustersServiceHandlerServer(ctx, mux, server.NewClusterServer(library, provider, client, discoveryClient, db, ns))

	//Add weave-gitops core handlers
	wegoServer := wego_server.NewApplicationsServer(&wego_server.ApplicationsConfig{
		Logger:     logr.Discard(), // TODO: Wire in logger
		KubeClient: kube,
	})
	wego_proto.RegisterApplicationsHandlerServer(ctx, mux, wegoServer)

	s := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		log.Infof("Shutting down the http gateway server")
		if err := s.Shutdown(context.Background()); err != nil {
			log.Errorf("Failed to shutdown http gateway server: %v", err)
		}
	}()

	log.Infof("Starting to listen and serve on %v", addr)

	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Errorf("Failed to listen and serve: %v", err)
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
func TrackEvents(ctx context.Context, r *http.Request) metadata.MD {
	var handler string
	md := make(map[string]string)
	if method, ok := grpc_runtime.RPCMethod(ctx); ok {
		md["method"] = method
		handler = method
	}
	if pattern, ok := grpc_runtime.HTTPPathPattern(ctx); ok {
		md["pattern"] = pattern
	}

	track(handler)

	return metadata.New(md)
}

func track(handler string) {
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
			go checkVersionWithFlags(m)
		}
	}
}

func checkVersionWithFlags(flags map[string]string) {
	p := &checkpoint.CheckParams{
		Product: "weave-gitops-enterprise",
		Version: version.Version,
		Flags:   flags,
	}
	checkResponse, err := checkpoint.Check(p)
	if err != nil {
		log.Debugf("Failed to check version: %v.", err)
		return
	}
	if checkResponse.Outdated {
		log.Infof("weave-gitops-enterprise version %s is available; please update at %s.",
			checkResponse.CurrentVersion, checkResponse.CurrentDownloadURL)
	} else {
		log.Debug("weave-gitops-enterprise version is up to date.")
	}
}
