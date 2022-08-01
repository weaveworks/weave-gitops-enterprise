package server

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/capi/v1alpha1"
	apitemplates "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
)

func TestToTemplate(t *testing.T) {
	testCases := []struct {
		name     string
		value    *capiv1.CAPITemplate
		expected *capiv1_protos.Template
		err      error
	}{
		{
			name:  "empty",
			value: &capiv1.CAPITemplate{},
			expected: &capiv1_protos.Template{
				Provider: "",
			},
		},
		{
			name: "Basics",
			value: &capiv1.CAPITemplate{
				TypeMeta: metav1.TypeMeta{
					Kind: "CAPITemplate",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
			expected: &capiv1_protos.Template{
				Name:         "foo",
				Provider:     "",
				TemplateKind: "CAPITemplate",
			},
		},
		{
			name: "Params and Objects",
			value: makeCAPITemplate(t, func(ct *capiv1.CAPITemplate) {
				ct.ObjectMeta.Name = "cluster-template-1"
				ct.Spec.Description = "this is test template 1"
				ct.Spec.ResourceTemplates = []apitemplates.ResourceTemplate{
					{
						RawExtension: rawExtension(`{
							"apiVersion": "fooversion",
							"kind": "fookind",
							"metadata": {
								"labels": {
								"name": "${CLUSTER_NAME}",
								"region": "${REGION}"
								}
							}
						}`),
					},
				}
			}),
			expected: &capiv1_protos.Template{
				Name:         "cluster-template-1",
				Description:  "this is test template 1",
				Provider:     "",
				TemplateKind: "CAPITemplate",
				Objects: []*capiv1_protos.TemplateObject{
					{
						ApiVersion: "fooversion",
						Kind:       "fookind",
						Parameters: []string{"CLUSTER_NAME", "REGION"},
					},
				},
				Parameters: []*capiv1_protos.Parameter{
					{
						Name:        "CLUSTER_NAME",
						Description: "This is used for the cluster naming.",
					},
					{
						Name: "REGION",
					},
				},
			},
		},
		{
			name: "annotations",
			value: &capiv1.CAPITemplate{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "capi.weave.works/v1alpha1",
					Kind:       "CAPITemplate",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
					Annotations: map[string]string{
						"hi": "there",
					},
				},
			},
			expected: &capiv1_protos.Template{
				Name:         "foo",
				Provider:     "",
				TemplateKind: "CAPITemplate",
				Annotations:  map[string]string{"hi": "there"},
			},
		},
		{
			name:  "With basic type errors",
			value: makeErrorTemplate(t, `"derp"`),
			expected: &capiv1_protos.Template{
				Name:         "cluster-template-1",
				TemplateKind: "CAPITemplate",
				Error:        "Couldn't load template body: failed to unmarshal resourceTemplate: json: cannot unmarshal string into Go value of type map[string]interface {}",
			},
		},
		{
			name:  "With structural errors",
			value: makeErrorTemplate(t, `{"boop":"beep"}`),
			expected: &capiv1_protos.Template{
				Name:         "cluster-template-1",
				TemplateKind: "CAPITemplate",
				Error:        "Couldn't load template body: failed to unmarshal resourceTemplate: Object 'Kind' is missing in '{\"boop\":\"beep\"}'",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result := ToTemplateResponse(tt.value)
			if diff := cmp.Diff(tt.expected, result, protocmp.Transform()); diff != "" {
				t.Fatalf("templates didn't match expected:\n%s", diff)
			}
		})
	}
}

func makeErrorTemplate(t *testing.T, rawData string) *capiv1.CAPITemplate {
	return makeCAPITemplate(t, func(ct *capiv1.CAPITemplate) {
		ct.ObjectMeta.Name = "cluster-template-1"
		ct.Spec.Description = ""
		ct.Spec.ResourceTemplates = []apitemplates.ResourceTemplate{
			{
				RawExtension: rawExtension(rawData),
			},
		}
	})
}
