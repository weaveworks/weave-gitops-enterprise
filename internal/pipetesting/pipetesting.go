package pipetesting

import (
	"context"
	"fmt"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/weaveworks/weave-gitops-enterprise/internal/grpctesting"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/pipelines/server"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"google.golang.org/grpc"

	v1 "k8s.io/api/core/v1"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/mgmtfetcher"
	mgmtfetcherfake "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/mgmtfetcher/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SetupServer(t *testing.T, fact clustersmngr.ClustersManager, c client.Client, cluster string, pipelineControllerAddress string, gitProvider git.Provider) pb.PipelinesClient {
	mgmtFetcher := mgmtfetcher.NewManagementCrossNamespacesFetcher(&mgmtfetcherfake.FakeNamespaceCache{
		Namespaces: []*v1.Namespace{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Namespace",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-ns",
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Namespace",
				},
			},
		},
	}, kubefakes.NewFakeClientGetter(c), &mgmtfetcherfake.FakeAuthClientGetter{})

	fmt.Println(pipelineControllerAddress)
	pipeSrv := server.NewPipelinesServer(server.ServerOpts{
		ClustersManager:           fact,
		ManagementFetcher:         mgmtFetcher,
		Cluster:                   cluster,
		PipelineControllerAddress: pipelineControllerAddress,
		GitProvider:               gitProvider,
	})

	conn := grpctesting.Setup(t, func(s *grpc.Server) {
		pb.RegisterPipelinesServer(s, pipeSrv)
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
