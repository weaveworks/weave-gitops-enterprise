package server

import (
	"context"
	"fmt"

	reflector "github.com/fluxcd/image-reflector-controller/api/v1beta1"
	ctrl "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/pipelines/internal/convert"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
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

func (s *server) ListImageAutomationObjects(ctx context.Context, msg *pb.ListImageAutomationObjectsRequest) (*pb.ListImageAutomationObjectsResponse, error) {
	c, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("getting impersonated client: %w", err)
	}

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &reflector.ImageRepositoryList{}
	})

	if err := c.ClusteredList(ctx, clist, false); err != nil {
		return nil, fmt.Errorf("fetching image update automations: %w", err)
	}

	res := []*pb.ImageRepository{}

	for _, list := range clist.Lists() {

		for _, l := range list {
			imgs, ok := l.(*reflector.ImageRepositoryList)
			if !ok {
				return nil, fmt.Errorf("type assertion error")
			}

			for _, img := range imgs.Items {
				res = append(res, &pb.ImageRepository{
					Name:      img.Name,
					Namespace: img.Namespace,
					Image:     img.Spec.Image,
				})
			}
		}
	}

	return &pb.ListImageAutomationObjectsResponse{
		ImageRepos: res,
	}, nil
}

func (s *server) ListImagePolicies(ctx context.Context, msg *pb.ListImagePoliciesRequest) (*pb.ListImagePoliciesResponse, error) {
	c, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("getting impersonated client: %w", err)
	}

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &reflector.ImagePolicyList{}
	})

	if err := c.ClusteredList(ctx, clist, false); err != nil {
		return nil, fmt.Errorf("fetching image polcies: %w", err)
	}

	res := []*pb.ImagePolicy{}

	for _, list := range clist.Lists() {

		for _, l := range list {
			pl, ok := l.(*reflector.ImagePolicyList)
			if !ok {
				return nil, fmt.Errorf("type assertion error")
			}

			for _, p := range pl.Items {
				policy := &pb.ImagePolicy{
					Policy: &pb.ImagePolicyChoice{},
					RepoRef: &pb.ImageRepoRef{
						Name:      p.Spec.ImageRepositoryRef.Name,
						Namespace: p.Spec.ImageRepositoryRef.Namespace,
					},
				}

				if p.Spec.Policy.SemVer != nil {
					policy.Policy.Semver = p.Spec.Policy.SemVer.Range
				}

				if p.Spec.Policy.Alphabetical != nil {
					policy.Policy.Alphabetical = p.Spec.Policy.Alphabetical.Order
				}

				if p.Spec.Policy.Numerical != nil {
					policy.Policy.Numerical = p.Spec.Policy.Numerical.Order
				}

				res = append(res, policy)
			}
		}
	}

	return &pb.ListImagePoliciesResponse{
		Policies: res,
	}, nil
}
