package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
	pacv2beta1 "github.com/weaveworks/policy-agent/api/v2beta1"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListPolicies(t *testing.T) {
	tests := []struct {
		name         string
		clusterState []runtime.Object
		clusterName  string
		expected     *capiv1_proto.ListPoliciesResponse
		err          error
	}{
		{
			name: "list policies",
			clusterState: []runtime.Object{
				makePolicy(t),
				makePolicy(t, func(p *pacv2beta1.Policy) {
					p.ObjectMeta.Name = "weave.policies.missing-app-label"
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
						CreatedAt: "0001-01-01T00:00:00Z",
						Targets: &capiv1_proto.PolicyTargets{
							Labels: []*capiv1_proto.PolicyTargetLabel{
								{
									Values: map[string]string{"my-label": "my-value"},
								},
							},
						},
						ClusterName: "Default",
					},
					{
						Name:      "Missing Owner Label",
						Severity:  "high",
						Code:      "foo",
						CreatedAt: "0001-01-01T00:00:00Z",
						Targets: &capiv1_proto.PolicyTargets{
							Labels: []*capiv1_proto.PolicyTargetLabel{
								{
									Values: map[string]string{"my-label": "my-value"},
								},
							},
						},
						ClusterName: "Default",
					},
				},
				Total: int32(2),
			},
		},
		{
			name: "list policies with parameter type string",
			clusterState: []runtime.Object{
				makePolicy(t, func(p *pacv2beta1.Policy) {
					strBytes, err := json.Marshal("value")
					if err != nil {
						t.Fatal(err)
					}
					p.Spec.Parameters = append(p.Spec.Parameters, pacv2beta1.PolicyParameters{
						Name:  "key",
						Type:  "string",
						Value: &apiextensionsv1.JSON{Raw: strBytes},
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
							Labels: []*capiv1_proto.PolicyTargetLabel{
								{
									Values: map[string]string{"my-label": "my-value"},
								},
							},
						},
						CreatedAt: "0001-01-01T00:00:00Z",
						Parameters: []*capiv1_proto.PolicyParam{
							{
								Name:  "key",
								Type:  "string",
								Value: getAnyValue(t, "string", "value"),
							},
						},
						ClusterName: "Default",
					},
				},
				Total: int32(1),
			},
		},
		{
			name: "list policies with parameter type integer",
			clusterState: []runtime.Object{
				makePolicy(t, func(p *pacv2beta1.Policy) {
					intBytes, err := json.Marshal(1)
					if err != nil {
						t.Fatal(err)
					}
					p.Spec.Parameters = append(p.Spec.Parameters, pacv2beta1.PolicyParameters{
						Name:  "key",
						Type:  "integer",
						Value: &apiextensionsv1.JSON{Raw: intBytes},
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
							Labels: []*capiv1_proto.PolicyTargetLabel{
								{
									Values: map[string]string{"my-label": "my-value"},
								},
							},
						},
						CreatedAt: "0001-01-01T00:00:00Z",
						Parameters: []*capiv1_proto.PolicyParam{
							{
								Name:  "key",
								Type:  "integer",
								Value: getAnyValue(t, "integer", int32(1)),
							},
						},
						ClusterName: "Default",
					},
				},
				Total: int32(1),
			},
		},
		{
			name: "list policies with parameter type boolean",
			clusterState: []runtime.Object{
				makePolicy(t, func(p *pacv2beta1.Policy) {
					boolBytes, err := json.Marshal(false)
					if err != nil {
						t.Fatal(err)
					}
					p.Spec.Parameters = append(p.Spec.Parameters, pacv2beta1.PolicyParameters{
						Name:  "key",
						Type:  "boolean",
						Value: &apiextensionsv1.JSON{Raw: boolBytes},
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
							Labels: []*capiv1_proto.PolicyTargetLabel{
								{
									Values: map[string]string{"my-label": "my-value"},
								},
							},
						},
						CreatedAt: "0001-01-01T00:00:00Z",
						Parameters: []*capiv1_proto.PolicyParam{
							{
								Name:  "key",
								Type:  "boolean",
								Value: getAnyValue(t, "boolean", false),
							},
						},
						ClusterName: "Default",
					},
				},
				Total: int32(1),
			},
		},
		{
			name: "list policies with parameter type array",
			clusterState: []runtime.Object{
				makePolicy(t, func(p *pacv2beta1.Policy) {
					sliceBytes, err := json.Marshal([]string{"value"})
					if err != nil {
						t.Fatal(err)
					}
					p.Spec.Parameters = append(p.Spec.Parameters, pacv2beta1.PolicyParameters{
						Name:  "key",
						Type:  "array",
						Value: &apiextensionsv1.JSON{Raw: sliceBytes},
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
							Labels: []*capiv1_proto.PolicyTargetLabel{
								{
									Values: map[string]string{"my-label": "my-value"},
								},
							},
						},
						CreatedAt: "0001-01-01T00:00:00Z",
						Parameters: []*capiv1_proto.PolicyParam{
							{
								Name:  "key",
								Type:  "array",
								Value: getAnyValue(t, "array", []string{"value"}),
							},
						},
						ClusterName: "Default",
					},
				},
				Total: int32(1),
			},
		},
		{
			name: "list policies with cluster filtering",
			clusterState: []runtime.Object{
				makePolicy(t),
			},
			expected: &capiv1_proto.ListPoliciesResponse{
				Policies: []*capiv1_proto.Policy{
					{
						Name:      "Missing Owner Label",
						Severity:  "high",
						Code:      "foo",
						CreatedAt: "0001-01-01T00:00:00Z",
						Targets: &capiv1_proto.PolicyTargets{
							Labels: []*capiv1_proto.PolicyTargetLabel{
								{
									Values: map[string]string{"my-label": "my-value"},
								},
							},
						},
						ClusterName: "Default",
					},
				},
				Total: int32(1),
			},
			clusterName: "Default",
		},
		{
			name: "list policies with invalid cluster filtering",
			clusterState: []runtime.Object{
				makePolicy(t),
			},
			err:         errors.New("error while listing policies for cluster wrong: cluster wrong not found"),
			clusterName: "wrong",
		},
		{
			name: "list policies with invalid parameter type",
			clusterState: []runtime.Object{
				makePolicy(t, func(p *pacv2beta1.Policy) {
					strBytes, err := json.Marshal("value")
					if err != nil {
						t.Fatal(err)
					}
					p.Spec.Parameters = append(p.Spec.Parameters, pacv2beta1.PolicyParameters{
						Name:  "key",
						Type:  "invalid",
						Value: &apiextensionsv1.JSON{Raw: strBytes},
					})
				}),
			},
			err: errors.New("found unsupported policy parameter type invalid in policy "),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientsPool := &clustersmngrfakes.FakeClientsPool{}
			fakeCl := createClient(t, tt.clusterState...)
			clients := map[string]client.Client{"Default": fakeCl}
			clientsPool.ClientsReturns(clients)
			clientsPool.ClientReturns(fakeCl, nil)
			clientsPool.ClientStub = func(name string) (client.Client, error) {
				if c, found := clients[name]; found && c != nil {
					return c, nil
				}

				return nil, fmt.Errorf("cluster %s not found", name)
			}
			clustersClient := clustersmngr.NewClient(clientsPool, map[string][]v1.Namespace{})

			fakeFactory := &clustersmngrfakes.FakeClientsFactory{}
			fakeFactory.GetImpersonatedClientReturns(clustersClient, nil)

			s := createServer(t, serverOptions{
				clientsFactory: fakeFactory,
			})

			req := capiv1_proto.ListPoliciesRequest{ClusterName: tt.clusterName}
			gotResponse, err := s.ListPolicies(context.Background(), &req)
			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to list policies:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("unexpected error while listing policies:\n%s", diff)
				}
			} else {
				if !cmpPoliciesResp(t, tt.expected, gotResponse) {
					t.Fatalf("policies didn't match expected:\n%+v\n%+v", tt.expected, gotResponse)
				}
			}
		})
	}
}

func TestPartialPoliciesConnectionErrors(t *testing.T) {
	clientsPool := &clustersmngrfakes.FakeClientsPool{}
	fakeCl := createClient(t, makePolicy(t))
	clients := map[string]client.Client{"Default": fakeCl}
	clientsPool.ClientsReturns(clients)
	clientsPool.ClientReturns(fakeCl, nil)

	clustersClient := clustersmngr.NewClient(clientsPool, map[string][]v1.Namespace{})
	clusterErr := clustersmngr.ClientError{ClusterName: "demo", Err: errors.New("failed adding cluster client to pool: connection refused")}
	fakeFactory := &clustersmngrfakes.FakeClientsFactory{}
	fakeFactory.GetImpersonatedClientStub = func(ctx context.Context, user *auth.UserPrincipal) (clustersmngr.Client, error) {
		var multi *multierror.Error
		multi = multierror.Append(multi, &clusterErr)
		return clustersClient, multi
	}
	s := createServer(t, serverOptions{
		clientsFactory: fakeFactory,
	})

	req := capiv1_proto.ListPoliciesRequest{}
	gotResponse, err := s.ListPolicies(context.Background(), &req)
	if err != nil {
		t.Fatal(err)
	}

	expectedPolicy := &capiv1_proto.ListPoliciesResponse{
		Policies: []*capiv1_proto.Policy{
			{
				Name:      "Missing Owner Label",
				Severity:  "high",
				Code:      "foo",
				CreatedAt: "0001-01-01T00:00:00Z",
				Targets: &capiv1_proto.PolicyTargets{
					Labels: []*capiv1_proto.PolicyTargetLabel{
						{
							Values: map[string]string{"my-label": "my-value"},
						},
					},
				},
				ClusterName: "Default",
			},
		},
		Total:  int32(1),
		Errors: []*capiv1_proto.ListError{{Message: clusterErr.Error(), ClusterName: clusterErr.ClusterName}},
	}
	if !cmpPoliciesResp(t, expectedPolicy, gotResponse) {
		t.Fatalf("policies didn't match expected:\n%+v\n%+v", expectedPolicy, gotResponse)
	}
}

func cmpPoliciesResp(t *testing.T, pol1 *capiv1_proto.ListPoliciesResponse, pol2 *capiv1_proto.ListPoliciesResponse) bool {
	t.Helper()
	if len(pol1.Policies) != len(pol2.Policies) {
		return false
	}

	if len(pol1.Errors) != len(pol2.Errors) {
		return false
	}
	for i := range pol1.Policies {
		if !cmpPolicy(t, pol1.Policies[i], pol2.Policies[i]) {
			return false
		}
	}

	for i := range pol1.Errors {
		if pol1.Errors[i].ClusterName != pol2.Errors[i].ClusterName {
			return false
		}

		if pol1.Errors[i].Message != pol2.Errors[i].Message {
			return false
		}

	}

	return cmp.Equal(pol1.Total, pol2.Total)
}

func cmpPolicy(t *testing.T, pol1 *capiv1_proto.Policy, pol2 *capiv1_proto.Policy) bool {
	t.Helper()

	if !cmp.Equal(pol1.Id, pol2.Id, protocmp.Transform()) {
		return false
	}
	if !cmp.Equal(pol1.Targets, pol2.Targets, protocmp.Transform()) {
		return false
	}
	if !cmp.Equal(pol1.Parameters, pol2.Parameters, protocmp.Transform()) {
		return false
	}
	if !cmp.Equal(pol1.ClusterName, pol2.ClusterName, protocmp.Transform()) {
		return false
	}
	return true
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
		src = &capiv1_proto.PolicyParamRepeatedString{Value: o.([]string)}
	}
	defaultAny, err := anypb.New(src)
	if err != nil {
		t.Fatal(err)
	}
	return defaultAny
}

func TestGetPolicy(t *testing.T) {
	tests := []struct {
		name         string
		policyName   string
		clusterName  string
		clusterState []runtime.Object
		err          error
		expected     *capiv1_proto.GetPolicyResponse
	}{
		{
			name:        "get policy",
			policyName:  "weave.policies.missing-owner-label",
			clusterName: "Default",
			clusterState: []runtime.Object{
				makePolicy(t),
			},
			expected: &capiv1_proto.GetPolicyResponse{
				Policy: &capiv1_proto.Policy{
					Name:     "Missing Owner Label",
					Severity: "high",
					Code:     "foo",
					Targets: &capiv1_proto.PolicyTargets{
						Labels: []*capiv1_proto.PolicyTargetLabel{
							{
								Values: map[string]string{"my-label": "my-value"},
							},
						},
					},
					CreatedAt:   "0001-01-01T00:00:00Z",
					ClusterName: "Default",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientsPool := &clustersmngrfakes.FakeClientsPool{}
			fakeCl := createClient(t, tt.clusterState...)
			clientsPool.ClientsReturns(map[string]client.Client{tt.clusterName: fakeCl})
			clientsPool.ClientReturns(fakeCl, nil)
			clustersClient := clustersmngr.NewClient(clientsPool, map[string][]v1.Namespace{})

			fakeFactory := &clustersmngrfakes.FakeClientsFactory{}
			fakeFactory.GetImpersonatedClientForClusterReturns(clustersClient, nil)

			s := createServer(t, serverOptions{
				clientsFactory: fakeFactory,
			})

			gotResponse, err := s.GetPolicy(context.Background(), &capiv1_proto.GetPolicyRequest{
				PolicyName:  tt.policyName,
				ClusterName: tt.clusterName})
			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to get policy:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("unexpected error while getting policy:\n%s", diff)
				}
			} else {
				if !cmpPolicy(t, tt.expected.Policy, gotResponse.Policy) {
					t.Fatalf("policies didn't match expected:\n%+v\n%+v", tt.expected, gotResponse)
				}
			}
		})
	}
}
