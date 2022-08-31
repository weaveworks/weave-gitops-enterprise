package server

import (
	"context"
	"fmt"

	ctrl "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/fetcher"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/pipelines/internal/convert"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *server) GetPipeline(ctx context.Context, msg *pb.GetPipelineRequest) (*pb.GetPipelineResponse, error) {
	c, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))

	if err != nil {
		return nil, fmt.Errorf("getting impersonated client: %w", err)
	}

	p := &ctrl.Pipeline{
		ObjectMeta: v1.ObjectMeta{
			Name:      msg.Name,
			Namespace: msg.Namespace,
		},
	}

	if err := c.Get(ctx, fetcher.ManagementClusterName, client.ObjectKeyFromObject(p), p); err != nil {
		return nil, fmt.Errorf("")
	}

	appGvk := schema.GroupVersionKind{
		Kind: p.Spec.AppRef.Kind,
	}
	for _, e := range p.Spec.Environments {
		for _, t := range e.Targets {
			app := unstructured.Unstructured{}
			app.SetName(p.Spec.AppRef.Name)
			app.SetNamespace(p.Namespace)
			app.SetGroupVersionKind(appGvk)

			c.Get(ctx, t.ClusterRef.Name)
		}
	}

	return &pb.GetPipelineResponse{
		Pipeline: convert.PipelineToProto(*p),
	}, nil
}
