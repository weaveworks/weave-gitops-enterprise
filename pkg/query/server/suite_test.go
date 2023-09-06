//go:build integration
// +build integration

package server_test

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-logr/logr"
	clusterctrlv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	gitopssets "github.com/weaveworks/gitopssets-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/server"
	api "github.com/weaveworks/weave-gitops-enterprise/pkg/api/query"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	queryserver "github.com/weaveworks/weave-gitops-enterprise/pkg/query/server"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"github.com/weaveworks/weave-gitops/core/nsaccess/nsaccessfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/discovery"
	typedauth "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/scheme"

	"github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var k8sClient client.Client
var cfg *rest.Config

func TestMain(m *testing.M) {
	// setup testEnvironment
	cmdOut, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	repoRoot := strings.TrimSpace(string(cmdOut))
	envTestPath := fmt.Sprintf("%s/tools/bin/envtest", repoRoot)
	os.Setenv("KUBEBUILDER_ASSETS", envTestPath)
	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("testdata", "crds"),
		},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err = testEnv.Start()
	if err != nil {
		log.Fatalf("starting test env failed: %s", err)
	}

	log.Println("environment started")

	err = sourcev1beta2.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatalf("add helm to schema failed: %s", err)
	}

	err = sourcev1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatalf("add helm to schema failed: %s", err)
	}

	err = kustomizev1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatalf("add helm to schema failed: %s", err)
	}

	err = helmv2beta1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatalf("add helm to schema failed: %s", err)
	}

	err = clusterctrlv1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatalf("add GitopsCluster to schema failed: %s", err)
	}

	err = gitopssets.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatalf("add GitopsSet to schema failed: %s", err)
	}
	_, cancel := context.WithCancel(context.Background())

	k8sClient, err = client.New(cfg, client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		log.Fatalf("cannot create kubernetes client: %s", err)
	}

	log.Println("kube client created")

	gomega.RegisterFailHandler(func(message string, skip ...int) {
		log.Println(message)
	})

	retCode := m.Run()
	log.Printf("suite ran with return code: %d", retCode)

	cancel()

	err = testEnv.Stop()
	if err != nil {
		log.Fatalf("stoping test env failed: %s", err)
	}

	log.Println("test environment stopped")
	os.Exit(retCode)
}

func makeQueryServer(t *testing.T, cfg *rest.Config, principal *auth.UserPrincipal, testLog logr.Logger) (api.QueryClient, error) {

	fetcher := &clustersmngrfakes.FakeClusterFetcher{}

	fakeCluster, err := cluster.NewSingleCluster("envtest", cfg, scheme.Scheme, kube.UserPrefixes{})
	if err != nil {
		return nil, fmt.Errorf("cannot create cluster:%w", err)
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
		testLog,
	)

	opts := server.ServerOpts{
		Logger:          testLog,
		ClustersManager: clustersManager,
	}

	enServer := server.NewClusterServer(opts)
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer(
		withClientsPoolInterceptor(clustersManager, principal),
	)

	pb.RegisterClustersServiceServer(s, enServer)

	dc, err := discovery.NewDiscoveryClientForConfig(cfg)

	opts2 := queryserver.ServerOpts{
		Logger:          testLog,
		DiscoveryClient: dc,
		ClustersManager: clustersManager,
		SkipCollection:  false,
		ObjectKinds:     configuration.SupportedObjectKinds,
		ServiceAccount: collector.ImpersonateServiceAccount{
			Name:      "collector",
			Namespace: "flux-system",
		},
	}

	qs, _, err := queryserver.NewServer(opts2)
	if err != nil {
		return nil, fmt.Errorf("cannot create query server:%w", err)
	}

	api.RegisterQueryServer(s, qs)

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	go func(tt *testing.T, log logr.Logger) {
		if err := s.Serve(lis); err != nil {
			tt.Error(err)
			log.Error(err, "serving request")
		}
	}(t, testLog)

	conn, err := grpc.DialContext(
		context.Background(),
		"bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create client:%w", err)
	}

	t.Cleanup(func() {
		//TODO review stop query server
		//stopQueryServer(context.Background())
		s.GracefulStop()
		conn.Close()
	})

	return api.NewQueryClient(conn), nil
}

func withClientsPoolInterceptor(clustersManager clustersmngr.ClustersManager, principal *auth.UserPrincipal) grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := clustersManager.UpdateClusters(ctx); err != nil {
			return nil, fmt.Errorf("failed to update clusters: %w", err)
		}
		if err := clustersManager.UpdateNamespaces(ctx); err != nil {
			return nil, fmt.Errorf("failed to update namespaces: %w", err)
		}

		clustersManager.UpdateUserNamespaces(ctx, principal)
		ctx = auth.WithPrincipal(ctx, principal)

		clusterClient, err := clustersManager.GetImpersonatedClient(ctx, principal)
		if err != nil {
			return nil, fmt.Errorf("failed to get impersonating client: %w", err)
		}

		ctx = context.WithValue(ctx, clustersmngr.ClustersClientCtxKey, clusterClient)

		return handler(ctx, req)
	})
}
