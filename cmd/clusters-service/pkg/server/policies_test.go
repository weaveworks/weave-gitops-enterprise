package server

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	policiesv1 "github.com/weaveworks/magalix-policy-agent/api/v1"
	capi_server "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	capiv1_protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Test_server_ListPolicies(t *testing.T) {
	tests := []struct {
		name         string
		clusterState []runtime.Object
		expected     *capiv1_proto.ListPoliciesResponse
		err          error
	}{
		{
			name: "list policies",
			clusterState: []runtime.Object{
				makePolicy(t),
				makePolicy(t, func(p *policiesv1.Policy) {
					p.ObjectMeta.Name = "magalix.policies.missing-app-label"
					p.Spec.Name = "Missing app Label"
					p.Spec.Severity = "medium"
				}),
			},
			expected: &capiv1_proto.ListPoliciesResponse{
				Policies: []*capiv1_proto.Policy{
					{
						Name:      "Missing app Label",
						Severity:  "medium",
						Code:      "foo",
						CreatedAt: "0001-01-01 00:00:00 +0000 UTC",
						Targets: &capiv1_proto.PolicyTargets{
							Labels: []*capi_server.PolicyTargetLabel{
								{
									Values: map[string]string{"my-label": "my-value"},
								},
							},
						},
					},
					{
						Name:      "Missing Owner Label",
						Severity:  "high",
						Code:      "foo",
						CreatedAt: "0001-01-01 00:00:00 +0000 UTC",
						Targets: &capiv1_proto.PolicyTargets{
							Labels: []*capi_server.PolicyTargetLabel{
								{
									Values: map[string]string{"my-label": "my-value"},
								},
							},
						},
					},
				},
				Total: int32(2),
			},
		},
		{
			name: "list policies with paramter type string",
			clusterState: []runtime.Object{
				makePolicy(t, func(p *policiesv1.Policy) {
					strBytes, err := json.Marshal("value")
					if err != nil {
						t.Fatal(err)
					}
					p.Spec.Parameters = append(p.Spec.Parameters, policiesv1.PolicyParameters{
						Name:    "key",
						Type:    "string",
						Default: &apiextensionsv1.JSON{Raw: strBytes},
					})
				}),
			},
			expected: &capiv1_proto.ListPoliciesResponse{
				Policies: []*capiv1_proto.Policy{
					{
						Name:     "Missing Owner Label",
						Severity: "high",
						Code:     "foo",
						Targets: &capiv1_proto.PolicyTargets{
							Labels: []*capi_server.PolicyTargetLabel{
								{
									Values: map[string]string{"my-label": "my-value"},
								},
							},
						},
						CreatedAt: "0001-01-01 00:00:00 +0000 UTC",
						Parameters: []*capi_server.PolicyParams{
							{
								Name:    "key",
								Type:    "string",
								Default: getAnyValue(t, "string", `"value"`),
							},
						},
					},
				},
				Total: int32(1),
			},
		},
		{
			name: "list policies with paramter type integer",
			clusterState: []runtime.Object{
				makePolicy(t, func(p *policiesv1.Policy) {
					intBytes, err := json.Marshal(1)
					if err != nil {
						t.Fatal(err)
					}
					p.Spec.Parameters = append(p.Spec.Parameters, policiesv1.PolicyParameters{
						Name:    "key",
						Type:    "integer",
						Default: &apiextensionsv1.JSON{Raw: intBytes},
					})
				}),
			},
			expected: &capiv1_proto.ListPoliciesResponse{
				Policies: []*capiv1_proto.Policy{
					{
						Name:     "Missing Owner Label",
						Severity: "high",
						Code:     "foo",
						Targets: &capiv1_proto.PolicyTargets{
							Labels: []*capi_server.PolicyTargetLabel{
								{
									Values: map[string]string{"my-label": "my-value"},
								},
							},
						},
						CreatedAt: "0001-01-01 00:00:00 +0000 UTC",
						Parameters: []*capi_server.PolicyParams{
							{
								Name:    "key",
								Type:    "integer",
								Default: getAnyValue(t, "integer", int32(1)),
							},
						},
					},
				},
				Total: int32(1),
			},
		},
		{
			name: "list policies with paramter type boolean",
			clusterState: []runtime.Object{
				makePolicy(t, func(p *policiesv1.Policy) {
					boolBytes, err := json.Marshal(false)
					if err != nil {
						t.Fatal(err)
					}
					p.Spec.Parameters = append(p.Spec.Parameters, policiesv1.PolicyParameters{
						Name:    "key",
						Type:    "boolean",
						Default: &apiextensionsv1.JSON{Raw: boolBytes},
					})
				}),
			},
			expected: &capiv1_proto.ListPoliciesResponse{
				Policies: []*capiv1_proto.Policy{
					{
						Name:     "Missing Owner Label",
						Severity: "high",
						Code:     "foo",
						Targets: &capiv1_proto.PolicyTargets{
							Labels: []*capi_server.PolicyTargetLabel{
								{
									Values: map[string]string{"my-label": "my-value"},
								},
							},
						},
						CreatedAt: "0001-01-01 00:00:00 +0000 UTC",
						Parameters: []*capi_server.PolicyParams{
							{
								Name:    "key",
								Type:    "boolean",
								Default: getAnyValue(t, "boolean", false),
							},
						},
					},
				},
				Total: int32(1),
			},
		},
		{
			name: "list policies with paramter type array",
			clusterState: []runtime.Object{
				makePolicy(t, func(p *policiesv1.Policy) {
					sliceBytes, err := json.Marshal([]string{"value"})
					if err != nil {
						t.Fatal(err)
					}
					p.Spec.Parameters = append(p.Spec.Parameters, policiesv1.PolicyParameters{
						Name:    "key",
						Type:    "array",
						Default: &apiextensionsv1.JSON{Raw: sliceBytes},
					})
				}),
			},
			expected: &capiv1_proto.ListPoliciesResponse{
				Policies: []*capiv1_proto.Policy{
					{
						Name:     "Missing Owner Label",
						Severity: "high",
						Code:     "foo",
						Targets: &capiv1_proto.PolicyTargets{
							Labels: []*capi_server.PolicyTargetLabel{
								{
									Values: map[string]string{"my-label": "my-value"},
								},
							},
						},
						CreatedAt: "0001-01-01 00:00:00 +0000 UTC",
						Parameters: []*capi_server.PolicyParams{
							{
								Name:    "key",
								Type:    "array",
								Default: getAnyValue(t, "array", []string{"value"}),
							},
						},
					},
				},
				Total: int32(1),
			},
		},
		{
			name: "list policies with invalid paramter type",
			clusterState: []runtime.Object{
				makePolicy(t, func(p *policiesv1.Policy) {
					strBytes, err := json.Marshal("value")
					if err != nil {
						t.Fatal(err)
					}
					p.Spec.Parameters = append(p.Spec.Parameters, policiesv1.PolicyParameters{
						Name:    "key",
						Type:    "invalid",
						Default: &apiextensionsv1.JSON{Raw: strBytes},
					})
				}),
			},
			err: errors.New("found unsupported policy paramter type invalid in policy "),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := createServer(t, tt.clusterState, "policies", "default", nil, nil, "", nil)
			listPoliciesRequest := new(capiv1_protos.ListPoliciesRequest)
			gotResponse, err := s.ListPolicies(context.Background(), listPoliciesRequest)
			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to list policies:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("unexpected error while listing policies:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tt.expected, gotResponse, protocmp.Transform()); diff != "" {
					t.Fatalf("policies didn't match expected:\n%s", diff)
				}
			}
		})
	}
}

func getAnyValue(t *testing.T, kind string, o interface{}) *anypb.Any {
	t.Helper()
	var src proto.Message
	switch kind {
	case "string":
		src = wrapperspb.String(o.(string))
	case "integer":
		src = wrapperspb.Int32(o.(int32))
	case "boolean":
		src = wrapperspb.Bool(o.(bool))
	case "array":
		src = &capiv1_proto.PolicyParamRepeatedString{Values: o.([]string)}
	}
	defaultAny, err := anypb.New(src)
	if err != nil {
		t.Fatal(err)
	}
	return defaultAny
}

func Test_server_GetPolicy(t *testing.T) {
	tests := []struct {
		name         string
		policy_name  string
		clusterState []runtime.Object
		err          error
		expected     *capiv1_proto.GetPolicyResponse
	}{
		{
			name:        "get policy",
			policy_name: "magalix.policies.missing-owner-label",
			clusterState: []runtime.Object{
				makePolicy(t),
			},
			expected: &capiv1_protos.GetPolicyResponse{
				Policy: &capiv1_protos.Policy{
					Name:     "Missing Owner Label",
					Severity: "high",
					Code:     "foo",
					Targets: &capiv1_proto.PolicyTargets{
						Labels: []*capi_server.PolicyTargetLabel{
							{
								Values: map[string]string{"my-label": "my-value"},
							},
						},
					},
					CreatedAt: "0001-01-01 00:00:00 +0000 UTC",
				},
			},
		},
		{
			name:        "policy not found",
			policy_name: "magalix.policies.not-found",
			err:         errors.New("error while getting policy magalix.policies.not-found: policies.magalix.com \"magalix.policies.not-found\" not found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := createServer(t, tt.clusterState, "policies", "default", nil, nil, "", nil)
			gotResponse, err := s.GetPolicy(context.Background(), &capiv1_proto.GetPolicyRequest{PolicyName: tt.policy_name})
			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to get policy:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("unexpected error while getting policy:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tt.expected, gotResponse, protocmp.Transform()); diff != "" {
					t.Fatalf("policy didn't match expected:\n%s", diff)
				}
			}
		})
	}
}
