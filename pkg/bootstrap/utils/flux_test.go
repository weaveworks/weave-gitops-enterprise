package utils

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/alecthomas/assert"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops-enterprise/test/utils"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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
					ReconcileStrategy: sourcev1beta2.ReconcileStrategyChartVersion,
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
	helmRepo := sourcev1beta2.HelmRepository{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-helm-repo",
			Namespace: "test-ns",
		},
		Spec: sourcev1beta2.HelmRepositorySpec{
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

func TestGetHelmReleaseProperty(t *testing.T) {

	tests := []struct {
		name        string
		property    string
		expectedOut string
		err         bool
	}{
		{
			name:     "property doesn't exist",
			property: "test",
			err:      true,
		},
		{
			name:        "property exist",
			property:    "version",
			expectedOut: "1.0.0",
			err:         false,
		},
		{
			name:        "property exist",
			property:    "domain",
			expectedOut: "testdomain.com",
			err:         false,
		},
	}

	values := map[string]interface{}{
		"ingress": map[string]interface{}{
			"annotations": map[string]string{
				"external-dns.alpha.kubernetes.io/hostname": "testdomain.com",
			},
			"className": "public-nginx",
			"enabled":   true,
			"hosts": []map[string]interface{}{
				{
					"host": "testdomain.com",
					"paths": []map[string]string{
						{
							"path":     "/",
							"pathType": "ImplementationSpecific",
						},
					},
				},
			},
		},
	}
	valuesBytes, err := json.Marshal(values)
	if err != nil {
		t.Fatalf("failed to marshal values: %v", err)
	}

	testHR := &helmv2.HelmRelease{
		TypeMeta: v1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.Identifier(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "wego",
			Namespace: "flux-system",
		}, Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:             "test-chart",
					ReconcileStrategy: sourcev1beta2.ReconcileStrategyChartVersion,
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1beta2.HelmRepositoryKind,
						Name:      "test-secret-name",
						Namespace: "test-secret-namespace",
					},
					Version: "1.0.0",
				},
			},
			Values: &apiextensionsv1.JSON{Raw: valuesBytes},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := utils.CreateFakeClient(t, testHR)
			prop, err := GetHelmReleaseProperty(client, "wego", "flux-system", tt.property)
			if err != nil {
				if tt.err {
					return
				}
				t.Fatalf("error getting property: %v", err)
			}
			assert.Equal(t, tt.expectedOut, prop, "invalid property")
		})
	}

}

func TestGetHelmReleaseValues(t *testing.T) {
	values := map[string]interface{}{
		"ingress": map[string]interface{}{
			"annotations": map[string]string{
				"external-dns.alpha.kubernetes.io/hostname": "testdomain.com",
			},
			"className": "public-nginx",
			"enabled":   true,
			"hosts": []map[string]interface{}{
				{
					"host": "testdomain.com",
					"paths": []map[string]string{
						{
							"path":     "/",
							"pathType": "ImplementationSpecific",
						},
					},
				},
			},
		},
	}
	valuesBytes, err := json.Marshal(values)
	if err != nil {
		t.Fatalf("failed to marshal values: %v", err)
	}

	testHR := &helmv2.HelmRelease{
		TypeMeta: v1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.Identifier(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "wego",
			Namespace: "flux-system",
		}, Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:             "test-chart",
					ReconcileStrategy: sourcev1beta2.ReconcileStrategyChartVersion,
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1beta2.HelmRepositoryKind,
						Name:      "test-secret-name",
						Namespace: "test-secret-namespace",
					},
					Version: "1.0.0",
				},
			},
			Values: &apiextensionsv1.JSON{Raw: valuesBytes},
		},
	}

	tests := []struct {
		name           string
		helmRelease    helmv2.HelmRelease
		expectedValues []byte
		err            bool
	}{
		{
			name:           "values exist",
			helmRelease:    *testHR,
			expectedValues: valuesBytes,
			err:            false,
		},
		{
			name:        "values doesn't exist",
			helmRelease: helmv2.HelmRelease{},
			err:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := utils.CreateFakeClient(t, &tt.helmRelease)
			values, err := GetHelmReleaseValues(client, "wego", "flux-system")
			if err != nil {
				if tt.err {
					return
				}
				t.Fatalf("error getting helmrelease: %v", err)
			}
			assert.Equal(t, tt.expectedValues, values, "invalid values")
		})
	}

}
