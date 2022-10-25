package server_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	ctrl "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/mgmtfetcher"
	mgmtfetcherfake "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/mgmtfetcher/fake"
	"github.com/weaveworks/weave-gitops-enterprise/internal/grpctesting"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/pipelines/server"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestListPipelines(t *testing.T) {
	p := &ctrl.Pipeline{}
	p.Name = "my-pipeline"
	p.Namespace = "default"

	k8s, factory := grpctesting.MakeFactoryWithObjects(p)

	mgmtFetcher := mgmtfetcher.NewManagementCrossNamespacesFetcher(&mgmtfetcherfake.FakeNamespaceCache{
		Namespaces: []*corev1.Namespace{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Namespace",
				},
			},
		},
	}, kubefakes.NewFakeClientGetter(k8s), &mgmtfetcherfake.FakeAuthClientGetter{})
	pipeSrv := server.NewPipelinesServer(server.ServerOpts{
		ClustersManager:   factory,
		ManagementFetcher: mgmtFetcher,
	})

	// c := pipetesting.SetupServer(t, factory, k8s)
	ctx := auth.WithPrincipal(context.Background(), &auth.UserPrincipal{ID: "userID"})

	res, err := pipeSrv.ListPipelines(ctx, &pb.ListPipelinesRequest{})

	if err != nil {
		t.Fatal(err)
	}

	l := &ctrl.PipelineList{}
	assert.NoError(t, k8s.List(context.Background(), l))

	if len(res.Pipelines) != len(l.Items) {
		t.Fatalf("expected %v piplelines to exist on the cluster; got %v", len(l.Items), len(res.Pipelines))
	}
}
