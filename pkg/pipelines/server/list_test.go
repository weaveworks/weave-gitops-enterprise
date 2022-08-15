package server_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	ctrl "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/pipelines/internal/pipetesting"
)

func TestListPipelines(t *testing.T) {
	p := &ctrl.Pipeline{}
	p.Name = "my-pipeline"

	k8s, factory := pipetesting.MakeClientsFactory(p)

	c := pipetesting.SetupServer(t, factory)

	res, err := c.ListPipelines(context.Background(), &pb.ListPipelinesRequest{})

	if err != nil {
		t.Fatal(err)
	}

	l := &ctrl.PipelineList{}
	assert.NoError(t, k8s.List(context.Background(), l))

	if len(res.Pipelines) != len(l.Items) {
		t.Fatalf("expected %v piplelines to exist on the cluster; got %v", len(l.Items), len(res.Pipelines))
	}
}
