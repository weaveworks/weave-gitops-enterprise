package server_test

import (
	"context"
	"fmt"
	"testing"

	helm "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ctrl "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/internal/grpctesting"
	"github.com/weaveworks/weave-gitops-enterprise/internal/pipetesting"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetPipeline(t *testing.T) {
	ctx := context.Background()

	cluster := types.NamespacedName{
		Name: "management",
	}

	kclient := fake.NewClientBuilder().WithScheme(grpctesting.BuildScheme()).Build()

	pipelineNamespace := pipetesting.NewNamespace(ctx, t, kclient)
	targetNamespace := pipetesting.NewNamespace(ctx, t, kclient)

	factory := grpctesting.MakeClustersManager(kclient, "management", fmt.Sprintf("%s/cluster-1", pipelineNamespace.Name))
	serverClient := pipetesting.SetupServer(t, factory, kclient, cluster)

	hr := createHelmRelease(ctx, t, kclient, "app-1", targetNamespace.Name)

	envName := "env-1"

	t.Run("cluster ref is not set", func(t *testing.T) {
		p := newPipeline("pipe-1", pipelineNamespace.Name, targetNamespace.Name, envName, hr)
		require.NoError(t, kclient.Create(ctx, p))

		res, err := serverClient.GetPipeline(context.Background(), &pb.GetPipelineRequest{
			Name:      p.Name,
			Namespace: pipelineNamespace.Name,
		})
		require.NoError(t, err)

		assert.Equal(t, p.Name, res.Pipeline.Name)
		assert.Equal(t, res.Pipeline.Status.Environments[envName].TargetsStatuses[0].Workloads[0].Version, hr.Spec.Chart.Spec.Version)
		assert.Equal(t, res.Pipeline.Status.Environments[envName].TargetsStatuses[0].Workloads[0].LastAppliedRevision, hr.Status.LastAppliedRevision)
		assert.Equal(t, res.Pipeline.Status.Environments[envName].TargetsStatuses[0].Namespace, targetNamespace.Name)
	})

	t.Run("cluster ref is set", func(t *testing.T) {
		p := newPipeline("pipe-2", pipelineNamespace.Name, targetNamespace.Name, envName, hr)
		p.Spec.Environments[0].Targets[0].ClusterRef = &ctrl.CrossNamespaceClusterReference{
			APIVersion: ctrl.GroupVersion.String(),
			Kind:       "GitopsCluster",
			Name:       "cluster-1",
			Namespace:  pipelineNamespace.Name,
		}
		require.NoError(t, kclient.Create(ctx, p))

		res, err := serverClient.GetPipeline(context.Background(), &pb.GetPipelineRequest{
			Name:      p.Name,
			Namespace: pipelineNamespace.Name,
		})
		require.NoError(t, err)

		assert.Equal(t, p.Name, res.Pipeline.Name)
		assert.Equal(t, res.Pipeline.Status.Environments[envName].TargetsStatuses[0].Workloads[0].Version, hr.Spec.Chart.Spec.Version)
		assert.Equal(t, res.Pipeline.Status.Environments[envName].TargetsStatuses[0].Namespace, targetNamespace.Name)
	})

	t.Run("cluster ref without Namespace", func(t *testing.T) {
		p := newPipeline("pipe-3", pipelineNamespace.Name, targetNamespace.Name, envName, hr)
		p.Spec.Environments[0].Targets[0].ClusterRef = &ctrl.CrossNamespaceClusterReference{
			APIVersion: ctrl.GroupVersion.String(),
			Kind:       "GitopsCluster",
			Name:       "cluster-1",
		}
		require.NoError(t, kclient.Create(ctx, p))

		res, err := serverClient.GetPipeline(context.Background(), &pb.GetPipelineRequest{
			Name:      p.Name,
			Namespace: pipelineNamespace.Name,
		})
		require.NoError(t, err)

		assert.Equal(t, p.Name, res.Pipeline.Name)
		assert.Equal(t, res.Pipeline.Status.Environments[envName].TargetsStatuses[0].Workloads[0].Version, hr.Spec.Chart.Spec.Version)
		assert.Equal(t, res.Pipeline.Status.Environments[envName].TargetsStatuses[0].Namespace, targetNamespace.Name)
	})

}

func newPipeline(name string, pNamespace string, tNamespace string, envName string, hr *helm.HelmRelease) *ctrl.Pipeline {
	return &ctrl.Pipeline{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: pNamespace,
		},
		Spec: ctrl.PipelineSpec{
			AppRef: ctrl.LocalAppReference{
				APIVersion: hr.GroupVersionKind().GroupVersion().String(),
				Kind:       hr.Kind,
				Name:       hr.Name,
			},
			Environments: []ctrl.Environment{
				{
					Name: envName,
					Targets: []ctrl.Target{
						{
							Namespace: tNamespace,
						},
					},
				},
			},
		},
	}
}

func createHelmRelease(ctx context.Context, t *testing.T, k client.Client, name string, ns string) *helm.HelmRelease {
	hr := &helm.HelmRelease{
		TypeMeta: v1.TypeMeta{
			APIVersion: helm.GroupVersion.String(),
			Kind:       "HelmRelease",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: helm.HelmReleaseSpec{
			Chart: helm.HelmChartTemplate{
				Spec: helm.HelmChartTemplateSpec{
					Version: "0.1.2",
				},
			},
		},
		Status: helm.HelmReleaseStatus{
			LastAppliedRevision: "0.1.2",
		},
	}

	require.NoError(t, k.Create(ctx, hr))

	return hr
}
