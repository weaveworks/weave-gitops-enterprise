package server

import (
	"reflect"
	"sort"
	"testing"

	capiv1 "github.com/weaveworks/templates-controller/apis/capi/v1alpha2"
	templatesv1 "github.com/weaveworks/templates-controller/apis/core"
	apitemplate "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/git"
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
				Spec: templatesv1.TemplateSpec{
					ResourceTemplates: []templatesv1.ResourceTemplate{
						{
							Content: []templatesv1.ResourceTemplateContent{
								{
									RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "AWSCluster"
							}`),
								},
							},
						},
					},
				},
			},
			provider: "aws",
		},
		{
			name: "AWSManagedCluster",
			template: &capiv1.CAPITemplate{
				Spec: templatesv1.TemplateSpec{
					ResourceTemplates: []templatesv1.ResourceTemplate{
						{
							Content: []templatesv1.ResourceTemplateContent{
								{
									RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "AWSManagedCluster"
							}`),
								},
							},
						},
					},
				},
			},
			provider: "aws",
		},
		{
			name: "AWSManagedControlPlane",
			template: &capiv1.CAPITemplate{
				Spec: templatesv1.TemplateSpec{
					ResourceTemplates: []templatesv1.ResourceTemplate{
						{
							Content: []templatesv1.ResourceTemplateContent{
								{
									RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "AWSManagedControlPlane"
							}`),
								},
							},
						},
					},
				},
			},
			provider: "aws",
		},
		{
			name: "AzureCluster",
			template: &capiv1.CAPITemplate{
				Spec: templatesv1.TemplateSpec{
					ResourceTemplates: []templatesv1.ResourceTemplate{
						{
							Content: []templatesv1.ResourceTemplateContent{
								{
									RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "AzureCluster"
							}`),
								},
							},
						},
					},
				},
			},
			provider: "azure",
		},
		{
			name: "AzureManagedCluster",
			template: &capiv1.CAPITemplate{
				Spec: templatesv1.TemplateSpec{
					ResourceTemplates: []templatesv1.ResourceTemplate{
						{
							Content: []templatesv1.ResourceTemplateContent{
								{
									RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "AzureManagedCluster"
							}`),
								},
							},
						},
					},
				},
			},
			provider: "azure",
		},
		{
			name: "DOCluster",
			template: &capiv1.CAPITemplate{
				Spec: templatesv1.TemplateSpec{
					ResourceTemplates: []templatesv1.ResourceTemplate{
						{
							Content: []templatesv1.ResourceTemplateContent{
								{
									RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "DOCluster"
							}`),
								},
							},
						},
					},
				},
			},
			provider: "digitalocean",
		},
		{
			name: "GCPCluster",
			template: &capiv1.CAPITemplate{
				Spec: templatesv1.TemplateSpec{
					ResourceTemplates: []templatesv1.ResourceTemplate{
						{
							Content: []templatesv1.ResourceTemplateContent{
								{
									RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "GCPCluster"
							}`),
								},
							},
						},
					},
				},
			},
			provider: "gcp",
		},
		{
			name: "OpenStackCluster",
			template: &capiv1.CAPITemplate{
				Spec: templatesv1.TemplateSpec{
					ResourceTemplates: []templatesv1.ResourceTemplate{
						{
							Content: []templatesv1.ResourceTemplateContent{
								{
									RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "OpenStackCluster"
							}`),
								},
							},
						},
					},
				},
			},
			provider: "openstack",
		},
		{
			name: "PacketCluster",
			template: &capiv1.CAPITemplate{
				Spec: templatesv1.TemplateSpec{
					ResourceTemplates: []templatesv1.ResourceTemplate{
						{
							Content: []templatesv1.ResourceTemplateContent{
								{
									RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "PacketCluster"
							}`),
								},
							},
						},
					},
				},
			},
			provider: "packet",
		},
		{
			name: "VSphereCluster",
			template: &capiv1.CAPITemplate{
				Spec: templatesv1.TemplateSpec{
					ResourceTemplates: []templatesv1.ResourceTemplate{
						{
							Content: []templatesv1.ResourceTemplateContent{
								{
									RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "VSphereCluster"
							}`),
								},
							},
						},
					},
				},
			},
			provider: "vsphere",
		},
		{
			name: "FooCluster",
			template: &capiv1.CAPITemplate{
				Spec: templatesv1.TemplateSpec{
					ResourceTemplates: []templatesv1.ResourceTemplate{
						{
							Content: []templatesv1.ResourceTemplateContent{
								{
									RawExtension: rawExtension(`{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha4",
								"kind": "FooCluster"
							}`),
								},
							},
						},
					},
				},
			},
			provider: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.provider != getProvider(tt.template, apitemplate.CAPIDisplayNameAnnotation) {
				t.Fatalf("expected %s but got %s", tt.provider, getProvider(tt.template, apitemplate.CAPIDisplayNameAnnotation))
			}
		})
	}
}

func TestGetMissingFiles(t *testing.T) {

	tests := []struct {
		name          string
		originalFiles []git.CommitFile
		extraFiles    []git.CommitFile
		expected      []git.CommitFile
	}{
		{
			name: "original files with empty extra files",
			originalFiles: []git.CommitFile{
				{
					Path:    "testdata/cluster-template.yaml",
					Content: strPtr("dummy content"),
				},
				{
					Path:    "testdata/cluster-template-1.yaml",
					Content: strPtr("dummy content"),
				},
			},
			extraFiles: []git.CommitFile{},
			expected: []git.CommitFile{
				{
					Path:    "testdata/cluster-template.yaml",
					Content: strPtr(""),
				},
				{
					Path:    "testdata/cluster-template-1.yaml",
					Content: strPtr(""),
				},
			},
		},
		{
			name: "original files with files not in extra files",
			originalFiles: []git.CommitFile{
				{
					Path:    "testdata/cluster-template.yaml",
					Content: strPtr("dummy content"),
				},
				{
					Path:    "testdata/cluster-template-1.yaml",
					Content: strPtr("dummy content"),
				},
			},
			extraFiles: []git.CommitFile{

				{
					Path:    "testdata/cluster-template-2.yaml",
					Content: strPtr("dummy content"),
				},
			},
			expected: []git.CommitFile{
				{

					Path:    "testdata/cluster-template.yaml",
					Content: strPtr(""),
				},
				{

					Path:    "testdata/cluster-template-1.yaml",
					Content: strPtr(""),
				},
			},
		},
		{
			name:          "no original files",
			originalFiles: []git.CommitFile{},
			extraFiles: []git.CommitFile{
				{
					Path:    "testdata/cluster-template.yaml",
					Content: strPtr("dummy content"),
				},
				{
					Path:    "testdata/cluster-template-1.yaml",
					Content: strPtr("dummy content"),
				},
			},
			expected: []git.CommitFile{},
		},
		{
			name: "original with 1 file and extra with 2 files not in original",
			originalFiles: []git.CommitFile{
				{
					Path:    "testdata/cluster-template.yaml",
					Content: strPtr("dummy content"),
				},
			},
			extraFiles: []git.CommitFile{

				{
					Path:    "testdata/cluster-template-2.yaml",
					Content: strPtr("dummy content"),
				}, {

					Path:    "testdata/cluster-template-1.yaml",
					Content: strPtr(""),
				},
			},
			expected: []git.CommitFile{
				{

					Path:    "testdata/cluster-template.yaml",
					Content: strPtr(""),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			sortFiles(tt.expected)
			var expectedPaths []string
			expectedContents := make([]*string, len(tt.expected))
			for i := range tt.expected {
				expectedPaths = append(expectedPaths, tt.expected[i].Path)
			}

			difference := getMissingFiles(tt.originalFiles, tt.extraFiles)
			sortFiles(difference)
			var differencePaths []string
			var differenceContents []*string
			for i := range difference {
				differencePaths = append(differencePaths, difference[i].Path)
				differenceContents = append(differenceContents, difference[i].Content)
			}

			// Check paths match expected paths
			if (len(differencePaths) > 0 && len(expectedPaths) > 0) && !reflect.DeepEqual(differencePaths, expectedPaths) {
				t.Errorf("File paths not matching expected, Paths = %v, want %v", difference, tt.expected)
			}
			// Check content of files to be empty
			if (len(differencePaths) > 0 && len(expectedPaths) > 0) && !reflect.DeepEqual(differenceContents, expectedContents) {
				t.Errorf("File content not matching expected, Content= %v, want %v", difference, tt.expected)
			}

		})
	}

}

func sortFiles(files []git.CommitFile) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
}

func strPtr(s string) *string {
	return &s
}
