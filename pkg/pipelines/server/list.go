package server

import (
	"context"
	"fmt"

	ctrl "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/pipelines/internal/convert"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *server) ListPipelines(ctx context.Context, msg *pb.ListPipelinesRequest) (*pb.ListPipelinesResponse, error) {
	c, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))

	if err != nil {
		return nil, fmt.Errorf("getting impersonated client: %w", err)
	}

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &ctrl.PipelineList{}
	})

	if err := c.ClusteredList(ctx, clist, false); err != nil {
		return nil, fmt.Errorf("listing pipelines: %w", err)
	}

	result := []*pb.Pipeline{}

	for cluster, lists := range clist.Lists() {
		for _, l := range lists {
			list, ok := l.(*ctrl.PipelineList)
			if !ok {
				continue
			}
			result = append(result, convert.PipelineToProto(list.Items, cluster)...)
		}

	}

	return &pb.ListPipelinesResponse{Pipelines: result}, nil

}
