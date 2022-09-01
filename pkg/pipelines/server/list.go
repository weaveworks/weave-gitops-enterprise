package server

import (
	"context"
	"fmt"

	ctrl "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/fetcher"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/pipelines/internal/convert"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

func (s *server) ListPipelines(ctx context.Context, msg *pb.ListPipelinesRequest) (*pb.ListPipelinesResponse, error) {
	c, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))

	if err != nil {
		return nil, fmt.Errorf("getting impersonated client: %w", err)
	}

	var res ctrl.PipelineList
	if err := c.List(ctx, fetcher.ManagementClusterName, &res); err != nil {
		return nil, fmt.Errorf("failed retrieving pipelines from API server: %w", err)
	}

	return &pb.ListPipelinesResponse{
		Pipelines: convert.PipelineToProto(res.Items),
	}, nil

}
