package entesting

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/go-logr/logr"
	pb "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/server"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"github.com/weaveworks/weave-gitops/core/nsaccess/nsaccessfakes"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

// MakeGRPCServer starts a GRPC server for clusters-service that uses `envtest`
func MakeGRPCServer(t *testing.T, cfg *rest.Config, k8sEnv *testutils.K8sTestEnv) pb.ClustersServiceClient {
	log := logr.Discard()

	fetcher := &clustersmngrfakes.FakeClusterFetcher{}
	fetcher.FetchReturns([]clustersmngr.Cluster{restConfigToCluster(k8sEnv.Rest)}, nil)

	nsChecker := nsaccessfakes.FakeChecker{}
	nsChecker.FilterAccessibleNamespacesStub = func(ctx context.Context, c *rest.Config, n []v1.Namespace) ([]v1.Namespace, error) {
		// Pretend the user has access to everything
		return n, nil
	}

	clientsFactory := clustersmngr.NewClientFactory(
		fetcher,
		&nsChecker,
		log,
		nil,
	)

	opts := server.ServerOpts{
		Logger:         log,
		ClientsFactory: clientsFactory,
	}

	enServer := server.NewClusterServer(opts)
	lis := bufconn.Listen(1024 * 1024)
	principal := &auth.UserPrincipal{}
	s := grpc.NewServer(
		withClientsPoolInterceptor(clientsFactory, cfg, principal),
	)

	pb.RegisterClustersServiceServer(s, enServer)

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	go func(tt *testing.T) {
		if err := s.Serve(lis); err != nil {
			tt.Error(err)
		}
	}(t)

	conn, err := grpc.DialContext(
		context.Background(),
		"bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		s.GracefulStop()
		conn.Close()
	})

	return pb.NewClustersServiceClient(conn)
}

func restConfigToCluster(cfg *rest.Config) clustersmngr.Cluster {
	return clustersmngr.Cluster{
		Name:      "Default",
		Server:    cfg.Host,
		TLSConfig: cfg.TLSClientConfig,
	}
}

func withClientsPoolInterceptor(clientsFactory clustersmngr.ClientsFactory, config *rest.Config, user *auth.UserPrincipal) grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := clientsFactory.UpdateClusters(ctx); err != nil {
			return nil, fmt.Errorf("failed to update clusters: %w", err)
		}
		if err := clientsFactory.UpdateNamespaces(ctx); err != nil {
			return nil, fmt.Errorf("failed to update namespaces: %w", err)
		}

		clientsFactory.UpdateUserNamespaces(ctx, user)

		ctx = auth.WithPrincipal(ctx, user)

		clusterClient, err := clientsFactory.GetImpersonatedClient(ctx, user)
		if err != nil {
			return nil, fmt.Errorf("failed to get impersonating client: %w", err)
		}

		ctx = context.WithValue(ctx, clustersmngr.ClustersClientCtxKey, clusterClient)

		return handler(ctx, req)
	})
}
