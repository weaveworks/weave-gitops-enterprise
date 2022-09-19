package pipetesting

import (
	"context"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/weaveworks/weave-gitops-enterprise/internal/grpctesting"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/pipelines/server"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"google.golang.org/grpc"

	v1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SetupServer(t *testing.T, fact clustersmngr.ClustersManager) pb.PipelinesClient {
	pipeSrv := server.NewPipelinesServer(server.ServerOpts{
		ClustersManager: fact,
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
