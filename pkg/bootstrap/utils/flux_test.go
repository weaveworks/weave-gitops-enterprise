package utils

import (
	"testing"
	"time"

	"github.com/alecthomas/assert"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateHelmReleaseYamlString(t *testing.T) {
	hr := helmv2.HelmRelease{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-hr",
			Namespace: "test-ns",
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:             "test-chart",
					ReconcileStrategy: sourcev1.ReconcileStrategyChartVersion,
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Name:      "test-repo",
						Namespace: "test-ns",
					},
					Version: "1.0.0",
				},
			},
			Install: &helmv2.Install{
				CRDs:            helmv2.Create,
				CreateNamespace: true,
			},
			Upgrade: &helmv2.Upgrade{
				CRDs: helmv2.Create,
			},
			Interval: v1.Duration{
				Duration: time.Hour,
			},
			Values: nil,
		},
	}

	expectedYaml := `apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: test-hr
  namespace: test-ns
spec:
  chart:
    spec:
      chart: test-chart
      reconcileStrategy: ChartVersion
      sourceRef:
        kind: HelmRepository
        name: test-repo
        namespace: test-ns
      version: 1.0.0
  install:
    crds: Create
    createNamespace: true
  interval: 1h0m0s
  upgrade:
    crds: Create
status: {}
`

	yamlString, err := CreateHelmReleaseYamlString(hr)
	assert.NoError(t, err, "error creating HelmRelease YAML string")
	assert.Equal(t, expectedYaml, yamlString, "error creating HelmRelease YAML string")
}

func TestCreateHelmRepositoryYamlString(t *testing.T) {
	helmRepo := sourcev1.HelmRepository{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-helm-repo",
			Namespace: "test-ns",
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL: "https://charts.example.com",
			SecretRef: &meta.LocalObjectReference{
				Name: "test-secret",
			},
			Interval: v1.Duration{
				Duration: time.Minute * 10,
			},
		},
	}

	expectedYaml := `apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  creationTimestamp: null
  name: test-helm-repo
  namespace: test-ns
spec:
  interval: 10m0s
  secretRef:
    name: test-secret
  url: https://charts.example.com
status: {}
`

	yamlString, err := CreateHelmRepositoryYamlString(helmRepo)
	assert.NoError(t, err, "error creating HelmRepository YAML string")
	assert.Equal(t, expectedYaml, yamlString, "doesn't match expected YAML")
}
