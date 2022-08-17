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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func MakeClientsFactory(objects ...client.Object) (client.Client, *clustersmngrfakes.FakeClientsFactory) {
	scheme := runtime.NewScheme()
	utilruntime.Must(v1.AddToScheme(scheme))
	utilruntime.Must(ctrl.AddToScheme(scheme))

	clientsPool := &clustersmngrfakes.FakeClientsPool{}
	k8s := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

	clientsPool.ClientsReturns(map[string]client.Client{"Default": k8s})
	clientsPool.ClientReturns(k8s, nil)
	clustersClient := clustersmngr.NewClient(clientsPool, map[string][]v1.Namespace{})

	factory := &clustersmngrfakes.FakeClientsFactory{}
	factory.GetImpersonatedClientReturns(clustersClient, nil)

	return k8s, factory
}

func SetupServer(t *testing.T, fact clustersmngr.ClientsFactory) pb.PipelinesClient {
	pipeSrv := server.NewPipelinesServer(server.ServerOpts{
		ClientsFactory: fact,
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
