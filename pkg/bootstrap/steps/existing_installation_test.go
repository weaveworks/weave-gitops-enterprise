package steps

import (
	"testing"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetHelmReleaseProperty(t *testing.T) {

	tests := []struct {
		name        string
		helmRelease helmv2.HelmRelease
		res         bool
	}{
		{
			name: "don't ask if helmrelease doesn't exist",
			res:  false,
		},
		{
			name: "ask if helmrelease exist",
			helmRelease: helmv2.HelmRelease{
				TypeMeta: v1.TypeMeta{
					Kind:       helmv2.HelmReleaseKind,
					APIVersion: helmv2.GroupVersion.Identifier(),
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "weave-gitops-enterprise",
					Namespace: "flux-system",
				}, Spec: helmv2.HelmReleaseSpec{
					Chart: helmv2.HelmChartTemplate{
						Spec: helmv2.HelmChartTemplateSpec{
							Chart:             "test-chart",
							ReconcileStrategy: sourcev1.ReconcileStrategyChartVersion,
							SourceRef: helmv2.CrossNamespaceObjectReference{
								Kind:      sourcev1.HelmRepositoryKind,
								Name:      "test-secret-name",
								Namespace: "test-secret-namespace",
							},
							Version: "1.0.0",
						},
					},
				},
			},
			res: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := makeTestConfig(t, Config{}, &tt.helmRelease)
			res, _ := askContinueWithExistingVersion([]StepInput{}, &config)
			assert.Equal(t, tt.res, res, "invalid result")
		})
	}

}
