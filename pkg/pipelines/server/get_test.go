package server_test

import (
	"context"
	"testing"

	helm "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ctrl "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/internal/pipetesting"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/pipelines"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestGetPipeline(t *testing.T) {
	ctx := context.Background()

	kclient, factory := pipetesting.MakeFactoryWithObjects()
	serverClient := pipetesting.SetupServer(t, factory)

	ns := pipetesting.NewNamespace(ctx, t, kclient)

	hr := createHelmRelease(ctx, t, kclient, "app-1", ns.Name)

	envName := "env-1"

	p := &ctrl.Pipeline{
		ObjectMeta: v1.ObjectMeta{
			Name:      "pipe-1",
			Namespace: ns.Name,
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
							Namespace: ns.Name,
							ClusterRef: ctrl.CrossNamespaceClusterReference{
								APIVersion: ctrl.GroupVersion.String(),
								Kind:       "GitopsCluster",
								Name:       "cluster-1",
								Namespace:  ns.Name,
							},
						},
					},
				},
			},
		},
	}
	require.NoError(t, kclient.Create(ctx, p))

	res, err := serverClient.GetPipeline(context.Background(), &pb.GetPipelineRequest{
		Name:      p.Name,
		Namespace: ns.Name,
	})
	require.NoError(t, err)

	assert.Equal(t, p.Name, res.Pipeline.Name)
	assert.Equal(t, res.Pipeline.Status.Environments[envName].Workloads[0].Version, hr.Spec.Chart.Spec.Version)
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
	}

	require.NoError(t, k.Create(ctx, hr))

	return hr
}
