package pipetesting

import (
	"context"
	"net"
	"testing"

	"github.com/alecthomas/assert"
	ctrl "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/pipelines/server"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	v1 "k8s.io/api/core/v1"

	helm "github.com/fluxcd/helm-controller/api/v2beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func buildScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	utilruntime.Must(v1.AddToScheme(scheme))
	utilruntime.Must(ctrl.AddToScheme(scheme))
	utilruntime.Must(helm.AddToScheme(scheme))

	return scheme
}

func MakeFactoryWithObjects(objects ...client.Object) (client.Client, *clustersmngrfakes.FakeClustersManager) {
	k8s := fake.NewClientBuilder().WithScheme(buildScheme()).WithObjects(objects...).Build()

	factory := MakeClustersManager(k8s)

	return k8s, factory
}

func MakeClustersManager(k8s client.Client) *clustersmngrfakes.FakeClustersManager {
	clientsPool := &clustersmngrfakes.FakeClientsPool{}

	clientsPool.ClientsReturns(map[string]client.Client{"Default": k8s})
	clientsPool.ClientStub = func(s string) (client.Client, error) {
		if s == "" {
			return nil, clustersmngr.ClusterNotFoundError{Cluster: ""}
		}

		return k8s, nil
	}

	nsMap := map[string][]v1.Namespace{"Default": {}}
	clustersClient := clustersmngr.NewClient(clientsPool, nsMap)

	factory := &clustersmngrfakes.FakeClustersManager{}
	factory.GetImpersonatedClientReturns(clustersClient, nil)

	return factory
}

func SetupServer(t *testing.T, fact clustersmngr.ClustersManager) pb.PipelinesClient {
	pipeSrv := server.NewPipelinesServer(server.ServerOpts{
		ClustersManager: fact,
	})

	lis := bufconn.Listen(1024 * 1024)

	s := grpc.NewServer()

	pb.RegisterPipelinesServer(s, pipeSrv)

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

	return pb.NewPipelinesClient(conn)
}

func NewNamespace(ctx context.Context, t *testing.T, k client.Client) v1.Namespace {
	ns := v1.Namespace{}
	ns.Name = "kube-test-" + rand.String(5)

	err := k.Create(ctx, &ns)
	assert.NoError(t, err, "should be able to create namespace: %s", ns.GetName())

	return ns
}
