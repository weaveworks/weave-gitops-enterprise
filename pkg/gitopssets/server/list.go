package server

import (
	"context"
	"fmt"

	ctrl "github.com/weaveworks/gitops-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/gitopssets"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/gitopssets/internal/convert"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *server) ListGitOpsSets(ctx context.Context, msg *pb.ListGitOpsSetsRequest) (*pb.ListGitOpsSetsResponse, error) {
	namespacedLists, err := s.managementFetcher.Fetch(ctx, ctrl.GitOpsKind, func() client.ObjectList {
		return &ctrl.GitOpsList{}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query gitops sets: %w", err)
	}

	gitopssets := []*pb.GitOpsSet{}
	errors := []*pb.ListError{}
	for _, namespacedList := range namespacedLists {
		if namespacedList.Error != nil {
			errors = append(errors, &pb.ListError{
				Namespace: namespacedList.Namespace,
				Message:   err.Error(),
			})
		}
		gitOpsList := namespacedList.List.(*ctrl.GitOpsList)
		for _, gs := range gitOpsList.Items {
			gitopssets = append(gitopssets, convert.GitOpsToProto(gs))
		}
	}

	return &pb.ListGitOpsSetsResponse{
		GitOpsSets: gitopssets,
		Errors:     errors,
	}, nil
}
