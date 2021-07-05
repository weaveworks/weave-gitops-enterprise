package app

import (
	"context"
	"net/http"
	"os"
	"strings"

	grpc_runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/weaveworks/wks/common/database/utils"
	"gorm.io/gorm"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	capiv1 "github.com/weaveworks/wks/cmd/capi-server/api/v1alpha1"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/git"
	capi_proto "github.com/weaveworks/wks/cmd/capi-server/pkg/protos"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/server"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/templates"
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
			return StartServer()
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

func StartServer() error {
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
	kubeClient, err := client.New(config.GetConfigOrDie(), client.Options{Scheme: scheme})
	if err != nil {
		return err
	}
	library := &templates.CRDLibrary{
		Client:    kubeClient,
		Namespace: os.Getenv("CAPI_TEMPLATES_NAMESPACE"),
	}
	provider := git.NewGitProviderService()
	return RunInProcessGateway(context.Background(), "0.0.0.0:8000", library, provider, kubeClient, db)
}

// RunInProcessGateway starts the invoke in process http gateway.
func RunInProcessGateway(ctx context.Context, addr string, library templates.Library, provider git.Provider, client client.Client, db *gorm.DB, opts ...grpc_runtime.ServeMuxOption) error {
	mux := grpc_runtime.NewServeMux(opts...)

	capi_proto.RegisterClustersServiceHandlerServer(ctx, mux, server.NewClusterServer(library, provider, client, db))
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
