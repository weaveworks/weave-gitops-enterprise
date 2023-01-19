package server

import (
	"context"
	"fmt"
	"regexp"

	ctrl "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *server) ListPullRequests(ctx context.Context, msg *pb.ListPullRequestsRequest) (*pb.ListPullRequestsResponse, error) {
	c, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))

	if err != nil {
		return nil, fmt.Errorf("getting impersonated client: %w", err)
	}

	p := ctrl.Pipeline{
		ObjectMeta: v1.ObjectMeta{
			Name:      msg.PipelineName,
			Namespace: msg.PipelineNamespace,
		},
	}

	if err := c.Get(ctx, s.cluster, client.ObjectKeyFromObject(&p), &p); err != nil {
		return nil, fmt.Errorf("failed to find pipeline=%s in namespace=%s in cluster=%s: %w", msg.PipelineName, msg.PipelineNamespace, s.cluster, err)
	}
	// client.Get does not always populate TypeMeta field, without this `kind` and
	// `apiVersion` are not returned in YAML representation.
	// https://github.com/kubernetes-sigs/controller-runtime/issues/1517#issuecomment-844703142
	p.SetGroupVersionKind(ctrl.GroupVersion.WithKind(ctrl.PipelineKind))

	if p.Spec.Promotion == nil || p.Spec.Promotion.Strategy.PullRequest == nil {
		return &pb.ListPullRequestsResponse{
			PullRequests: map[string]string{},
		}, nil
	}

	sc, err := s.clients.GetServerClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed getting server client: %w", err)
	}

	// getting provider token from pipeline definition
	var secret corev1.Secret
	if err := sc.Get(ctx, s.cluster, client.ObjectKey{Namespace: msg.PipelineNamespace, Name: p.Spec.Promotion.Strategy.PullRequest.SecretRef.Name}, &secret); err != nil {
		return nil, fmt.Errorf("failed to fetch Secret: %w", err)
	}

	gp := git.GitProvider{
		Token: string(secret.Data["token"]),
		Type:  p.Spec.Promotion.Strategy.PullRequest.Type.String(),
	}

	allPrs, err := s.gitProvider.ListPullRequests(ctx, gp, p.Spec.Promotion.Strategy.PullRequest.URL)
	if err != nil {
		return nil, fmt.Errorf("failed listing pull requests: %w", err)
	}

	openPrs := map[string]string{}

	pathPattern := regexp.MustCompile("([^/]+)/([^/]+)/([^/]+)")

	for _, pr := range allPrs {
		prInfo := pr.Get()
		if prInfo.Merged {
			continue
		}

		pathMatches := pathPattern.FindStringSubmatch(prInfo.Description)
		if pathMatches == nil {
			continue
		}

		env := pathMatches[3]

		openPrs[env] = prInfo.WebURL
	}

	return &pb.ListPullRequestsResponse{
		PullRequests: openPrs,
	}, nil
}
