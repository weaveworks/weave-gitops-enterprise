package server

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/capi/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/capi"
	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
)

func TestToTemplate(t *testing.T) {
	testCases := []struct {
		name     string
		value    string
		expected *capiv1_protos.Template
		err      error
	}{
		{
			name:  "empty",
			value: "",
			expected: &capiv1_protos.Template{
				Provider: "",
			},
		},
		{
			name: "Basics",
			value: `
apiVersion: capi.weave.works/v1alpha1
kind: CAPITemplate
metadata:
  name: foo
`,
			expected: &capiv1_protos.Template{
				Name:     "foo",
				Provider: "",
			},
		},
		{
			name: "Params and Objects",
			value: makeTemplate(t, func(ct *capiv1.CAPITemplate) {
				ct.ObjectMeta.Name = "cluster-template-1"
				ct.Spec.Description = "this is test template 1"
				ct.Spec.ResourceTemplates = []capiv1.ResourceTemplate{
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
				Name:        "cluster-template-1",
				Description: "this is test template 1",
				Provider:    "",
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
			value: `
apiVersion: capi.weave.works/v1alpha1
kind: CAPITemplate
metadata:
  annotations:
    hi: there
  name: foo
`,
			expected: &capiv1_protos.Template{
				Name:        "foo",
				Provider:    "",
				Annotations: map[string]string{"hi": "there"},
			},
		},
		{
			name:  "With basic type errors",
			value: makeErrorTemplate(t, `"derp"`),
			expected: &capiv1_protos.Template{
				Name:  "cluster-template-1",
				Error: "Couldn't load template body: failed to unmarshal resourceTemplate: json: cannot unmarshal string into Go value of type map[string]interface {}",
			},
		},
		{
			name:  "With structural errors",
			value: makeErrorTemplate(t, `{ "boop": "beep" }`),
			expected: &capiv1_protos.Template{
				Name:  "cluster-template-1",
				Error: "Couldn't load template body: failed to unmarshal resourceTemplate: Object 'Kind' is missing in '{\"boop\":\"beep\"}'",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result := ToTemplateResponse(mustParseBytes(t, tt.value))
			if diff := cmp.Diff(tt.expected, result, protocmp.Transform()); diff != "" {
				t.Fatalf("templates didn't match expected:\n%s", diff)
			}
		})
	}
}

func makeErrorTemplate(t *testing.T, rawData string) string {
	return makeTemplate(t, func(ct *capiv1.CAPITemplate) {
		ct.ObjectMeta.Name = "cluster-template-1"
		ct.Spec.Description = ""
		ct.Spec.ResourceTemplates = []capiv1.ResourceTemplate{
			{
				RawExtension: rawExtension(rawData),
			},
		}
	})
}

func mustParseBytes(t *testing.T, data string) *capiv1.CAPITemplate {
	t.Helper()
	parsed, err := capi.ParseBytes([]byte(data), "no-key-provided")
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
