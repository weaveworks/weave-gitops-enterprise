package templates

import (
	"testing"

	"github.com/alecthomas/assert"
	"github.com/google/go-cmp/cmp"
	capiv1 "github.com/weaveworks/templates-controller/apis/capi/v1alpha2"
	templatesv1 "github.com/weaveworks/templates-controller/apis/core"
	gapiv1 "github.com/weaveworks/templates-controller/apis/gitops/v1alpha2"
	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"google.golang.org/protobuf/testing/protocmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetProfilesFromTemplate(t *testing.T) {
	t.Run("base case", func(t *testing.T) {
		annotations := map[string]string{
			"capi.weave.works/profile-0": "{\"name\": \"k8s-rbac-permissions\", \"version\": \"0.0.8\",  \"values\": \"adminGroups: weaveworks\"}",
			"capi.weave.works/profile-1": "{\"name\": \"external-dns\", \"version\": \"0.0.8\", \"editable\": true }",
			"capi.weave.works/profile-2": "{\"name\": \"cert-manager\", \"version\": \"2.0.1\"}",
		}

		expected := []*capiv1_protos.TemplateProfile{
			{Name: "cert-manager", Version: "2.0.1", Required: true},
			{Name: "external-dns", Version: "0.0.8", Editable: true, Required: true},
			{Name: "k8s-rbac-permissions", Version: "0.0.8", Values: "adminGroups: weaveworks", Required: true},
		}

		tm := makeCAPITemplate(t, func(c *capiv1.CAPITemplate) {
			c.Annotations = annotations
		})
		result, err := GetProfilesFromTemplate(tm)

		assert.NoError(t, err)

		if diff := cmp.Diff(expected, result, protocmp.Transform()); diff != "" {
			t.Fatalf("template params didn't match expected:\n%s", diff)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		annotations := map[string]string{
			"capi.weave.works/profile-0": "{\"version\": \"0.0.8\",  \"values\": \"adminGroups: weaveworks\"}",
		}
		tm := makeCAPITemplate(t, func(c *capiv1.CAPITemplate) {
			c.Annotations = annotations
		})
		_, err := GetProfilesFromTemplate(tm)
		assert.Error(t, err)
		assert.Regexp(t, "profile name is required", err.Error())
	})

	t.Run("bad json", func(t *testing.T) {
		annotations := map[string]string{
			"capi.weave.works/profile-0": "{\"name\": \"k8s-rbac-permissions\", \"version\": \"0.0.8\",  \"values\": \"adminGroups: weaveworks\"",
		}
		tm := makeCAPITemplate(t, func(c *capiv1.CAPITemplate) {
			c.Annotations = annotations
		})
		_, err := GetProfilesFromTemplate(tm)
		assert.Error(t, err)
		assert.Regexp(t, "failed to unmarshal profiles: unexpected end of JSON input", err.Error())
	})

	t.Run("profiles in template.spec.profiles overrides profiles specified in the annotations", func(t *testing.T) {
		// base annotations
		annotations := map[string]string{
			"capi.weave.works/profile-0": "{\"name\": \"k8s-rbac-permissions\", \"version\": \"0.0.8\",  \"values\": \"adminGroups: weaveworks\"}",
			"capi.weave.works/profile-1": "{\"name\": \"external-dns\", \"version\": \"0.0.7\", \"editable\": true }",
		}

		// profiles in template.spec.profiles
		profiles := []templatesv1.Chart{
			{Chart: "cert-manager", Version: "2.0.1", SourceRef: corev1.ObjectReference{Name: "charts", Namespace: "default"}},
			{Chart: "external-dns", Version: "0.0.8", Editable: true},
		}

		expected := []*capiv1_protos.TemplateProfile{
			// spec
			{Name: "cert-manager", Version: "2.0.1", SourceRef: &capiv1_protos.SourceRef{Name: "charts", Namespace: "default"}},
			{Name: "external-dns", Version: "0.0.8", Editable: true, SourceRef: &capiv1_protos.SourceRef{Name: "", Namespace: ""}},
			// annotations
			{Name: "k8s-rbac-permissions", Version: "0.0.8", Values: "adminGroups: weaveworks", Required: true},
		}

		tm := makeCAPITemplate(t, func(c *capiv1.CAPITemplate) {
			c.Annotations = annotations
			c.Spec.Charts.Items = profiles
		})

		result, err := GetProfilesFromTemplate(tm)
		assert.NoError(t, err)
		if diff := cmp.Diff(expected, result, protocmp.Transform()); diff != "" {
			t.Fatalf("template params didn't match expected:\n%s", diff)
		}
	})

	t.Run("All the fields are loaded properly from template.profiles", func(t *testing.T) {
		profiles := []templatesv1.Chart{
			{
				Chart:   "k8s-rbac-permissions",
				Version: "0.0.8",
				HelmReleaseTemplate: templatesv1.HelmReleaseTemplateSpec{
					Content: &templatesv1.HelmReleaseTemplate{
						RawExtension: runtime.RawExtension{
							Raw: []byte(`{ "spec": { "interval": "${INTERVAL}" } }`),
						},
					},
				},
				Values: &templatesv1.HelmReleaseValues{
					RawExtension: runtime.RawExtension{
						Raw: []byte(`{ "adminGroups": "weaveworks" }`),
					},
				},
				Layer:           "layer-foo",
				TargetNamespace: "foo-ns",
				Editable:        true,
				Required:        true,
				SourceRef: corev1.ObjectReference{
					Name:      "charts",
					Namespace: "default",
				},
			},
		}

		expected := []*capiv1_protos.TemplateProfile{
			{
				Name:            "k8s-rbac-permissions",
				Version:         "0.0.8",
				Editable:        true,
				ProfileTemplate: "spec:\n  interval: ${INTERVAL}\n",
				Values:          "adminGroups: weaveworks\n",
				Layer:           "layer-foo",
				Namespace:       "foo-ns",
				Required:        true,
				SourceRef:       &capiv1_protos.SourceRef{Name: "charts", Namespace: "default"},
			},
		}

		tm := makeCAPITemplate(t, func(c *capiv1.CAPITemplate) {
			c.Spec.Charts.Items = profiles
		})

		result, err := GetProfilesFromTemplate(tm)
		// no error
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})
}

func TestTemplateHasRequiredProfiles(t *testing.T) {
	// no profiles
	t.Run("no profiles", func(t *testing.T) {
		tm := makeCAPITemplate(t)
		hasRequiredProfiles, err := TemplateHasRequiredProfiles(tm)
		assert.NoError(t, err)
		assert.False(t, hasRequiredProfiles)
	})

	t.Run("annotations", func(t *testing.T) {
		tm := makeCAPITemplate(t, func(c *capiv1.CAPITemplate) {
			c.SetAnnotations(map[string]string{
				"capi.weave.works/profile-0": `{"name": "demo-profile", "version": "0.0.1" }`,
			})
		})
		hasRequiredPrfiles, err := TemplateHasRequiredProfiles(tm)
		assert.NoError(t, err)
		assert.True(t, hasRequiredPrfiles)
	})
}

func TestProfileAnnotations(t *testing.T) {
	// Create a gitopstemplate with a profile annotation
	gitopstemplate := &gapiv1.GitOpsTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
			Annotations: map[string]string{
				"capi.weave.works/profile-0": "test-profile-0",
				"capi.weave.works/profile-1": "test-profile-1",
				// misspelled profile annotation
				"capi.weave.works/profiles-1": "test-profile",
				"other-annot":                 "test-profile",
			},
		},
	}

	// Get the profile annotations
	profileAnnotations := ProfileAnnotations(gitopstemplate)

	// Check the profile annotations
	expectedProfileAnnotations := map[string]string{
		"capi.weave.works/profile-0": "test-profile-0",
		"capi.weave.works/profile-1": "test-profile-1",
	}

	// use cmp.Diff to compare the two maps
	if diff := cmp.Diff(expectedProfileAnnotations, profileAnnotations); diff != "" {
		t.Fatalf("Annotations did not match:\n%s", diff)
	}
}

// FIXME: try and share this with the other tests
func makeCAPITemplate(t *testing.T, opts ...func(*capiv1.CAPITemplate)) *capiv1.CAPITemplate {
	t.Helper()
	basicRaw := `
	{
		"apiVersion":"fooversion",
		"kind":"fookind",
		"metadata":{
		   "name":"${CLUSTER_NAME}",
		   "annotations":{
			  "capi.weave.works/display-name":"ClusterName"
		   }
		}
	 }`
	ct := &capiv1.CAPITemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       capiv1.Kind,
			APIVersion: "capi.weave.works/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-template-1",
			Namespace: "default",
		},
		Spec: templatesv1.TemplateSpec{
			Description: "this is test template 1",
			Params: []templatesv1.TemplateParam{
				{
					Name:        "CLUSTER_NAME",
					Description: "This is used for the cluster naming.",
				},
			},
			ResourceTemplates: []templatesv1.ResourceTemplate{
				{
					Content: []templatesv1.ResourceTemplateContent{
						{
							RawExtension: rawExtension(basicRaw),
						},
					},
				},
			},
		},
	}
	for _, o := range opts {
		o(ct)
	}
	return ct
}

func rawExtension(s string) runtime.RawExtension {
	return runtime.RawExtension{
		Raw: []byte(s),
	}
}
