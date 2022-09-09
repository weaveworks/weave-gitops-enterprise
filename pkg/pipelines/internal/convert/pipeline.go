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
	}

	for _, e := range p.Spec.Environments {
		env := &pb.Environment{
			Name:    e.Name,
			Targets: []*pb.Target{},
		}

		for _, t := range e.Targets {
			env.Targets = append(env.Targets, &pb.Target{
				Namespace: t.Namespace,
				ClusterRef: &pb.ClusterRef{
					Name: t.ClusterRef.Name,
					Kind: t.ClusterRef.Kind,
				}})
		}

		r.Environments = append(r.Environments, env)

	}

	return r
}
