package grpctesting

import (
	"context"
	"net"
	"testing"

	helm "github.com/fluxcd/helm-controller/api/v2beta1"
	ctrl "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	tfctrl "github.com/weaveworks/tf-controller/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func BuildScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	utilruntime.Must(v1.AddToScheme(scheme))
	utilruntime.Must(ctrl.AddToScheme(scheme))
	utilruntime.Must(helm.AddToScheme(scheme))
	utilruntime.Must(tfctrl.AddToScheme(scheme))

	return scheme
}

func MakeFactoryWithObjects(objects ...client.Object) (client.Client, *clustersmngrfakes.FakeClustersManager) {
	k8s := fake.NewClientBuilder().WithScheme(BuildScheme()).WithObjects(objects...).Build()

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

func Setup(t *testing.T, register func(s *grpc.Server)) *grpc.ClientConn {
	lis := bufconn.Listen(1024 * 1024)

	s := grpc.NewServer()

	register(s)

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

	return conn
}
