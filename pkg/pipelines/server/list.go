package server

import (
	"context"
	"fmt"

	ctrl "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/pipelines/internal/convert"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *server) ListPipelines(ctx context.Context, msg *pb.ListPipelinesRequest) (*pb.ListPipelinesResponse, error) {
	namespacedLists, err := s.managementFetcher.Fetch(ctx, ctrl.PipelineKind, func() client.ObjectList {
		return &ctrl.PipelineList{}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query pipelines: %w", err)
	}

	pipelines := []*pb.Pipeline{}
	errors := []*pb.ListError{}
	for _, namespacedList := range namespacedLists {
		if namespacedList.Error != nil {
			errors = append(errors, &pb.ListError{
				Namespace: namespacedList.Namespace,
				Message:   err.Error(),
			})
		}
		pipelinesList := namespacedList.List.(*ctrl.PipelineList)
		for _, p := range pipelinesList.Items {
			pipelines = append(pipelines, convert.PipelineToProto(p))
		}
	}

	return &pb.ListPipelinesResponse{
		Pipelines: pipelines,
		Errors:    errors,
	}, nil
}
