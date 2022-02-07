package server

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
)

func TestGetProvider(t *testing.T) {
	tests := []struct {
		name     string
		template *capiv1.CAPITemplate
		provider string
	}{
		{
			name: "AWSCluster",
			template: &capiv1.CAPITemplate{
				Spec: capiv1.CAPITemplateSpec{
					ResourceTemplates: []capiv1.CAPIResourceTemplate{
						{
							RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "AWSCluster"
							}`),
						},
					},
				},
			},
			provider: "aws",
		},
		{
			name: "AWSManagedCluster",
			template: &capiv1.CAPITemplate{
				Spec: capiv1.CAPITemplateSpec{
					ResourceTemplates: []capiv1.CAPIResourceTemplate{
						{
							RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "AWSManagedCluster"
							}`),
						},
					},
				},
			},
			provider: "aws",
		},
		{
			name: "AWSManagedControlPlane",
			template: &capiv1.CAPITemplate{
				Spec: capiv1.CAPITemplateSpec{
					ResourceTemplates: []capiv1.CAPIResourceTemplate{
						{
							RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "AWSManagedControlPlane"
							}`),
						},
					},
				},
			},
			provider: "aws",
		},
		{
			name: "AzureCluster",
			template: &capiv1.CAPITemplate{
				Spec: capiv1.CAPITemplateSpec{
					ResourceTemplates: []capiv1.CAPIResourceTemplate{
						{
							RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "AzureCluster"
							}`),
						},
					},
				},
			},
			provider: "azure",
		},
		{
			name: "AzureManagedCluster",
			template: &capiv1.CAPITemplate{
				Spec: capiv1.CAPITemplateSpec{
					ResourceTemplates: []capiv1.CAPIResourceTemplate{
						{
							RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "AzureManagedCluster"
							}`),
						},
					},
				},
			},
			provider: "azure",
		},
		{
			name: "DOCluster",
			template: &capiv1.CAPITemplate{
				Spec: capiv1.CAPITemplateSpec{
					ResourceTemplates: []capiv1.CAPIResourceTemplate{
						{
							RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "DOCluster"
							}`),
						},
					},
				},
			},
			provider: "digitalocean",
		},
		{
			name: "GCPCluster",
			template: &capiv1.CAPITemplate{
				Spec: capiv1.CAPITemplateSpec{
					ResourceTemplates: []capiv1.CAPIResourceTemplate{
						{
							RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "GCPCluster"
							}`),
						},
					},
				},
			},
			provider: "gcp",
		},
		{
			name: "OpenStackCluster",
			template: &capiv1.CAPITemplate{
				Spec: capiv1.CAPITemplateSpec{
					ResourceTemplates: []capiv1.CAPIResourceTemplate{
						{
							RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "OpenStackCluster"
							}`),
						},
					},
				},
			},
			provider: "openstack",
		},
		{
			name: "PacketCluster",
			template: &capiv1.CAPITemplate{
				Spec: capiv1.CAPITemplateSpec{
					ResourceTemplates: []capiv1.CAPIResourceTemplate{
						{
							RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "PacketCluster"
							}`),
						},
					},
				},
			},
			provider: "packet",
		},
		{
			name: "VSphereCluster",
			template: &capiv1.CAPITemplate{
				Spec: capiv1.CAPITemplateSpec{
					ResourceTemplates: []capiv1.CAPIResourceTemplate{
						{
							RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "VSphereCluster"
							}`),
						},
					},
				},
			},
			provider: "vsphere",
		},
		{
			name: "FooCluster",
			template: &capiv1.CAPITemplate{
				Spec: capiv1.CAPITemplateSpec{
					ResourceTemplates: []capiv1.CAPIResourceTemplate{
						{
							RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "FooCluster"
							}`),
						},
					},
				},
			},
			provider: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.provider != getProvider(tt.template) {
				t.Fatalf("expected %s but got %s", tt.provider, getProvider(tt.template))
			}
		})
	}
}

func TestGenerateProfileFiles(t *testing.T) {
	c := createClient(t, makeTestHelmRepository("base"))
	file, err := generateProfileFiles(
		context.TODO(),
		"testing",
		"test-ns",
		"",
		"cluster-foo",
		c,
		[]*capiv1_protos.ProfileValues{
			{
				Name:    "foo",
				Version: "0.0.1",
				Values:  base64.StdEncoding.EncodeToString([]byte("foo: bar")),
			},
		},
	)
	assert.NoError(t, err)
	expected := `apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: HelmRepository
metadata:
  creationTimestamp: null
  name: testing
  namespace: test-ns
spec:
  interval: 10m0s
  url: base/charts
status: {}
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: cluster-foo-foo
  namespace: wego-system
spec:
  chart:
    spec:
      chart: foo
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta1
        kind: HelmRepository
        name: testing
        namespace: test-ns
      version: 0.0.1
  interval: 1m0s
  values:
    foo: bar
status: {}
`
	assert.Equal(t, expected, *file.Content)
}

func TestGenerateProfileFilesWithLayers(t *testing.T) {
	c := createClient(t, makeTestHelmRepository("base"))
	file, err := generateProfileFiles(
		context.TODO(),
		"testing",
		"test-ns",
		"",
		"cluster-foo",
		c,
		[]*capiv1_protos.ProfileValues{
			{
				Name:    "foo",
				Version: "0.0.1",
				Values:  base64.StdEncoding.EncodeToString([]byte("foo: bar")),
			},
			{
				Name:    "bar",
				Version: "0.0.1",
				Layer:   "testing",
				Values:  base64.StdEncoding.EncodeToString([]byte("foo: bar")),
			},
		},
	)
	assert.NoError(t, err)
	expected := `apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: HelmRepository
metadata:
  creationTimestamp: null
  name: testing
  namespace: test-ns
spec:
  interval: 10m0s
  url: base/charts
status: {}
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  labels:
    weave.works/applied-layer: testing
  name: cluster-foo-bar
  namespace: wego-system
spec:
  chart:
    spec:
      chart: bar
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta1
        kind: HelmRepository
        name: testing
        namespace: test-ns
      version: 0.0.1
  interval: 1m0s
  values:
    foo: bar
status: {}
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  creationTimestamp: null
  name: cluster-foo-foo
  namespace: wego-system
spec:
  chart:
    spec:
      chart: foo
      sourceRef:
        apiVersion: source.toolkit.fluxcd.io/v1beta1
        kind: HelmRepository
        name: testing
        namespace: test-ns
      version: 0.0.1
  dependsOn:
  - name: cluster-foo-bar
  interval: 1m0s
  values:
    foo: bar
status: {}
`
	assert.Equal(t, expected, *file.Content)
}

func TestGetProfilesFromTemplate(t *testing.T) {
	annotations := map[string]string{
		"capi.weave.works/profile-0": "{\"name\": \"profile-a\", \"version\": \"v0.0.1\" }",
	}

	expected := []*capiv1_protos.TemplateProfile{
		{
			Name:    "profile-a",
			Version: "v0.0.1",
		},
	}

	result, err := getProfilesFromTemplate(annotations)
	assert.NoError(t, err)

	assert.Equal(t, result, expected)
}
