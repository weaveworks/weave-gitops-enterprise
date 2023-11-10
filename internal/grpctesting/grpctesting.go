package grpctesting

import (
	"context"
	"net"
	"testing"

	helm "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/go-logr/logr"
	gitopssetsv1 "github.com/weaveworks/gitopssets-controller/api/v1alpha1"
	ctrl "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	tfctrl "github.com/weaveworks/tf-controller/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func BuildScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(ctrl.AddToScheme(scheme))
	utilruntime.Must(helm.AddToScheme(scheme))
	utilruntime.Must(tfctrl.AddToScheme(scheme))
	utilruntime.Must(gitopssetsv1.AddToScheme(scheme))
	utilruntime.Must(rbacv1.AddToScheme(scheme))
	utilruntime.Must(appsv1.AddToScheme(scheme))

	return scheme
}

func MakeFactoryWithObjects(objects ...client.Object) (client.Client, *clustersmngrfakes.FakeClustersManager) {
	k8s := fake.NewClientBuilder().WithScheme(BuildScheme()).WithObjects(objects...).WithStatusSubresource(&tfctrl.Terraform{}).Build()

	factory := MakeClustersManager(k8s, nil)

	return k8s, factory
}

// MakeClustersManager creates a fake ClustersManager for testing
// Pass in nsMap to give the test "principal" access to more namespaces on clusters e.g.
// "Default": {corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}}}
// will search for resources in ns: "default" on the "Default" cluster.
// By default ("nil"), no namespaces will be searched.
func MakeClustersManager(k8s client.Client, nsMap map[string][]v1.Namespace, clusters ...string) *clustersmngrfakes.FakeClustersManager {
	clientsPool := &clustersmngrfakes.FakeClientsPool{}

	clientsPool.ClientsReturns(map[string]client.Client{"Default": k8s})
	clientsPool.ClientStub = func(s string) (client.Client, error) {
		if s == "" {
			return nil, clustersmngr.ClusterNotFoundError{Cluster: ""}
		}
		if len(clusters) == 0 {
			return k8s, nil
		}

		for _, cluster := range clusters {
			if cluster == s {
				return k8s, nil
			}
		}

		return nil, clustersmngr.ClusterNotFoundError{Cluster: s}
	}

	// default no namespaces
	if nsMap == nil {
		nsMap = map[string][]v1.Namespace{"Default": {}}
	}
	clustersClient := clustersmngr.NewClient(clientsPool, nsMap, logr.Discard())

	factory := &clustersmngrfakes.FakeClustersManager{}
	factory.GetImpersonatedClientReturns(clustersClient, nil)
	factory.GetServerClientReturns(clustersClient, nil)

	return factory
}

func Setup(t *testing.T, register func(s *grpc.Server), opt ...grpc.ServerOption) *grpc.ClientConn {
	lis := bufconn.Listen(1024 * 1024)

	s := grpc.NewServer(opt...)

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
