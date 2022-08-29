package convert

import (
	"testing"

	ctrl "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
	"gotest.tools/v3/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPipelineToProto(t *testing.T) {
	name := "test"
	ns := "test-ns"
	cluster := "cluster"

	list := []ctrl.Pipeline{{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: ctrl.PipelineSpec{
			AppRef: ctrl.LocalAppReference{
				APIVersion: "pipelines.weave.works/v1alpha1",
				Kind:       "HelmRelease",
				Name:       name,
			},
			Environments: []ctrl.Environment{{
				Name: "env",
				Targets: []ctrl.Target{{
					Namespace: ns,
					ClusterRef: ctrl.CrossNamespaceClusterReference{
						Name: cluster,
						Kind: "GitopsCluster",
					},
				}},
			}},
		},
	}}

	expected := &pb.Pipeline{
		Name:      name,
		Namespace: ns,
		AppRef: &pb.AppRef{
			APIVersion: "pipelines.weave.works/v1alpha1",
			Kind:       "HelmRelease",
			Name:       name,
		},
		Environments: []*pb.Environment{{
			Name: "env",
			Targets: []*pb.Target{{
				Namespace: ns,
				ClusterRef: &pb.ClusterRef{
					Name: cluster,
					Kind: "GitopsCluster",
				},
			}},
		}},
	}

	converted := PipelineToProto(list, cluster)[0]

	assert.Equal(t, expected.Name, converted.Name)
	assert.Equal(t, expected.Namespace, converted.Namespace)

	assert.Equal(t, expected.AppRef.APIVersion, converted.AppRef.APIVersion)
	assert.Equal(t, expected.AppRef.Kind, converted.AppRef.Kind)
	assert.Equal(t, expected.AppRef.Name, converted.AppRef.Name)

	expEnv := expected.Environments[0]
	convEnv := converted.Environments[0]
	assert.Equal(t, expEnv.Name, convEnv.Name)

	expTarget := expEnv.Targets[0]
	convTarget := convEnv.Targets[0]
	assert.Equal(t, expTarget.Namespace, convTarget.Namespace)
	assert.Equal(t, expTarget.ClusterRef.Name, convTarget.ClusterRef.Name)
}
