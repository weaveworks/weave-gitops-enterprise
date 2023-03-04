package convert

import (
	ctrl "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
)

func PipelineToProto(p ctrl.Pipeline) *pb.Pipeline {
	r := &pb.Pipeline{
		Name:      p.Name,
		Namespace: p.Namespace,
		AppRef: &pb.AppRef{
			ApiVersion: p.Spec.AppRef.APIVersion,
			Kind:       p.Spec.AppRef.Kind,
			Name:       p.Spec.AppRef.Name,
		},
		Environments: []*pb.Environment{},
		Type:         p.GetObjectKind().GroupVersionKind().Kind,
		Promotion: &pb.Promotion{
			Strategy: &pb.Strategy{},
		},
	}

	if p.Spec.Promotion != nil {

		r.Promotion.Manual = p.Spec.Promotion.Manual

		if p.Spec.Promotion.Strategy.SecretRef != nil {
			r.Promotion.Strategy.SecretRef = &pb.LocalObjectReference{
				Name: p.Spec.Promotion.Strategy.SecretRef.Name,
			}
		}

		if p.Spec.Promotion.Strategy.PullRequest != nil {
			r.Promotion.Strategy.PullRequest = &pb.PullRequestPromotion{
				Type:   string(p.Spec.Promotion.Strategy.PullRequest.Type),
				Url:    p.Spec.Promotion.Strategy.PullRequest.URL,
				Branch: p.Spec.Promotion.Strategy.PullRequest.BaseBranch,
			}
		}

		if p.Spec.Promotion.Strategy.Notification != nil {
			r.Promotion.Strategy.Notification = &pb.Notification{}
		}
	}

	for _, e := range p.Spec.Environments {
		env := &pb.Environment{
			Name:    e.Name,
			Targets: []*pb.Target{},
		}

		for _, t := range e.Targets {
			var clusterRef pb.ClusterRef

			if t.ClusterRef != nil {
				clusterRef = pb.ClusterRef{
					Kind: t.ClusterRef.Kind,
					Name: t.ClusterRef.Name,
				}
			}

			env.Targets = append(env.Targets, &pb.Target{
				Namespace:  t.Namespace,
				ClusterRef: &clusterRef,
			})
		}

		r.Environments = append(r.Environments, env)

	}
	return r
}
