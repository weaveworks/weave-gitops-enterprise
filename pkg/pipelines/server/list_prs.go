package server

import (
	"context"
	"fmt"
	"net/url"
	"strings"

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

	sc, err := s.clients.GetServerClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed getting server client: %w", err)
	}

	openPrs := map[string]string{}

	for _, e := range p.Spec.Environments {
		promotion := p.Spec.GetPromotion(e.Name)

		// if no PullRequest promotion is defined for the environment, skip
		if promotion == nil || promotion.Strategy.PullRequest == nil {
			continue
		}

		// getting provider token from pipeline definition
		var secret corev1.Secret
		if err := sc.Get(ctx, s.cluster, client.ObjectKey{Namespace: msg.PipelineNamespace, Name: promotion.Strategy.PullRequest.SecretRef.Name}, &secret); err != nil {
			return nil, fmt.Errorf("failed to fetch Secret: %w", err)
		}

		url, err := url.Parse(promotion.Strategy.PullRequest.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL: %w", err)
		}

		gp := git.GitProvider{
			Token:    string(secret.Data["token"]),
			Type:     promotion.Strategy.PullRequest.Type.String(),
			Hostname: url.Hostname(),
		}

		allPrs, err := s.gitProvider.ListPullRequests(ctx, gp, promotion.Strategy.PullRequest.URL)
		if err != nil {
			return nil, fmt.Errorf("failed listing pull requests: %w", err)
		}

		for _, prInfo := range allPrs {
			if prInfo.Merged {
				continue
			}

			if strings.Contains(prInfo.Description, fmt.Sprintf("%s/%s/%s", p.Namespace, p.Name, e.Name)) {
				openPrs[e.Name] = prInfo.Link
			}
		}
	}

	return &pb.ListPullRequestsResponse{
		PullRequests: openPrs,
	}, nil
}
