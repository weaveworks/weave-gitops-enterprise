package app

import (
	"context"
	"net/http"
	"os"

	grpc_runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	capiv1 "github.com/weaveworks/wks/cmd/capi-server/api/v1alpha1"
	capi_proto "github.com/weaveworks/wks/cmd/capi-server/pkg/protos"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/server"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/templates"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func NewAPIServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "capi-server",
		Long: `The capi-server servers and handles REST operations for CAPI templates.
		CAPI templates are stored in the cluster as a ConfigMap indexed by their name.`,

		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return StartServer()
		},
	}
	return cmd
}

func StartServer() error {
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
		Namespace: os.Getenv("POD_NAMESPACE"),
	}
	return RunInProcessGateway(context.Background(), "0.0.0.0:8000", library)
}

// RunInProcessGateway starts the invoke in process http gateway.
func RunInProcessGateway(ctx context.Context, addr string, library templates.Library, opts ...grpc_runtime.ServeMuxOption) error {
	mux := grpc_runtime.NewServeMux(opts...)

	capi_proto.RegisterClustersServiceHandlerServer(ctx, mux, server.NewClusterServer(library))
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

	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Errorf("Failed to listen and serve: %v", err)
		return err
	}
	return nil
}
