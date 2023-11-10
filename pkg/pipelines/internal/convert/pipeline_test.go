package convert

import (
	"testing"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/stretchr/testify/assert"
	ctrl "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
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
				APIVersion: "helm.toolkit.fluxcd.io/v2beta1",
				Kind:       "HelmRelease",
				Name:       name,
			},
			Environments: []ctrl.Environment{{
				Name: "env",
				Targets: []ctrl.Target{{
					Namespace: ns,
					ClusterRef: &ctrl.CrossNamespaceClusterReference{
						Name: cluster,
						Kind: "GitopsCluster",
					},
				}},
			}},
			Promotion: &ctrl.Promotion{
				Manual: true,
				Strategy: ctrl.Strategy{
					PullRequest: &ctrl.PullRequestPromotion{
						Type:       ctrl.Github,
						URL:        "https://github.com/weaveworks/pipeline-controller",
						BaseBranch: "main",
					},
					SecretRef: &meta.LocalObjectReference{
						Name: "secret",
					},
				},
			},
		},
	}}

	expected := &pb.Pipeline{
		Name:      name,
		Namespace: ns,
		AppRef: &pb.AppRef{
			ApiVersion: "helm.toolkit.fluxcd.io/v2beta1",
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
		Promotion: &pb.Promotion{
			Manual: true,
			Strategy: &pb.Strategy{
				PullRequest: &pb.PullRequestPromotion{
					Type:   "github",
					Url:    "https://github.com/weaveworks/pipeline-controller",
					Branch: "main",
				},
				SecretRef: &pb.LocalObjectReference{
					Name: "secret",
				},
			},
		},
	}

	converted := PipelineToProto(list[0])

	assert.Equal(t, expected.Name, converted.Name)
	assert.Equal(t, expected.Namespace, converted.Namespace)

	assert.Equal(t, expected.AppRef.ApiVersion, converted.AppRef.ApiVersion)
	assert.Equal(t, expected.AppRef.Kind, converted.AppRef.Kind)
	assert.Equal(t, expected.AppRef.Name, converted.AppRef.Name)
	assert.Equal(t, expected.Promotion.Manual, converted.Promotion.Manual)
	assert.Equal(t, expected.Promotion.Strategy.PullRequest.Type, converted.Promotion.Strategy.PullRequest.Type)
	assert.Equal(t, expected.Promotion.Strategy.PullRequest.Url, converted.Promotion.Strategy.PullRequest.Url)
	assert.Equal(t, expected.Promotion.Strategy.PullRequest.Branch, converted.Promotion.Strategy.PullRequest.Branch)
	assert.Equal(t, expected.Promotion.Strategy.SecretRef.Name, converted.Promotion.Strategy.SecretRef.Name)

	expEnv := expected.Environments[0]
	convEnv := converted.Environments[0]
	assert.Equal(t, expEnv.Name, convEnv.Name)

	expTarget := expEnv.Targets[0]
	convTarget := convEnv.Targets[0]
	assert.Equal(t, expTarget.Namespace, convTarget.Namespace)
	assert.Equal(t, expTarget.ClusterRef.Name, convTarget.ClusterRef.Name)
}
