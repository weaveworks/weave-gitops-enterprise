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
