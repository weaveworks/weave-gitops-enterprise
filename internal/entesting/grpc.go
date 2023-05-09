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
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"github.com/weaveworks/weave-gitops/core/nsaccess/nsaccessfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	v1 "k8s.io/api/core/v1"
	typedauth "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
)

// MakeGRPCServer starts a GRPC server for clusters-service that uses `envtest`
func MakeGRPCServer(t *testing.T, cfg *rest.Config, k8sEnv *testutils.K8sTestEnv) pb.ClustersServiceClient {
	log := logr.Discard()

	fetcher := &clustersmngrfakes.FakeClusterFetcher{}

	scheme, err := kube.CreateScheme()
	if err != nil {
		t.Fatal(err)
	}

	fakeCluster, err := cluster.NewSingleCluster("Default", k8sEnv.Rest, scheme, kube.UserPrefixes{})
	if err != nil {
		t.Fatal(err)
	}

	fetcher.FetchReturns([]cluster.Cluster{fakeCluster}, nil)

	nsChecker := nsaccessfakes.FakeChecker{}
	nsChecker.FilterAccessibleNamespacesStub = func(ctx context.Context, client typedauth.AuthorizationV1Interface, n []v1.Namespace) ([]v1.Namespace, error) {
		// Pretend the user has access to everything
		return n, nil
	}

	clustersManager := clustersmngr.NewClustersManager(
		[]clustersmngr.ClusterFetcher{fetcher},
		&nsChecker,
		log,
	)

	opts := server.ServerOpts{
		Logger:          log,
		ClustersManager: clustersManager,
	}

	enServer := server.NewClusterServer(opts)
	lis := bufconn.Listen(1024 * 1024)
	principal := auth.NewUserPrincipal(auth.Token("1234"))
	s := grpc.NewServer(
		withClientsPoolInterceptor(clustersManager, cfg, principal),
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

func withClientsPoolInterceptor(clustersManager clustersmngr.ClustersManager, config *rest.Config, user *auth.UserPrincipal) grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := clustersManager.UpdateClusters(ctx); err != nil {
			return nil, fmt.Errorf("failed to update clusters: %w", err)
		}
		if err := clustersManager.UpdateNamespaces(ctx); err != nil {
			return nil, fmt.Errorf("failed to update namespaces: %w", err)
		}

		clustersManager.UpdateUserNamespaces(ctx, user)

		ctx = auth.WithPrincipal(ctx, user)

		clusterClient, err := clustersManager.GetImpersonatedClient(ctx, user)
		if err != nil {
			return nil, fmt.Errorf("failed to get impersonating client: %w", err)
		}

		ctx = context.WithValue(ctx, clustersmngr.ClustersClientCtxKey, clusterClient)

		return handler(ctx, req)
	})
}
