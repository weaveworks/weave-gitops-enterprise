package server

import (
	"testing"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
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
