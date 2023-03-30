package server

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	structpb "google.golang.org/protobuf/types/known/structpb"

	"sigs.k8s.io/controller-runtime/pkg/client"

	pacv2beta2 "github.com/weaveworks/policy-agent/api/v2beta2"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestListPolicyConfigs(t *testing.T) {
	clusters := []struct {
		name  string
		state []runtime.Object
	}{
		{
			name: "management",
			state: []runtime.Object{
				&pacv2beta2.PolicyConfig{
					TypeMeta: metav1.TypeMeta{
						APIVersion: pacv2beta2.GroupVersion.Identifier(),
						Kind:       pacv2beta2.PolicyConfigKind,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "policyconfig-1",
					},
					Spec: pacv2beta2.PolicyConfigSpec{
						Config: map[string]pacv2beta2.PolicyConfigConfig{
							"policy-1": {
								Parameters: map[string]apiextensionsv1.JSON{
									"param-1": {Raw: []byte{}},
									"param-2": {Raw: []byte{}},
								},
							},
							"policy-2": {
								Parameters: map[string]apiextensionsv1.JSON{
									"param-1": {Raw: []byte{}},
									"param-2": {Raw: []byte{}},
								},
							},
						},
						Match: pacv2beta2.PolicyConfigTarget{
							Namespaces: []string{"namespace-1", "namespace-2"},
						},
					},
					Status: pacv2beta2.PolicyConfigStatus{
						Status: "OK",
					},
				},
				&pacv2beta2.PolicyConfig{
					TypeMeta: metav1.TypeMeta{
						APIVersion: pacv2beta2.GroupVersion.Identifier(),
						Kind:       pacv2beta2.PolicyConfigKind,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "policyconfig-2",
					},
					Spec: pacv2beta2.PolicyConfigSpec{
						Config: map[string]pacv2beta2.PolicyConfigConfig{
							"policy-3": {
								Parameters: map[string]apiextensionsv1.JSON{
									"param-a": {Raw: []byte{}},
									"param-b": {Raw: []byte{}},
								},
							},
						},
						Match: pacv2beta2.PolicyConfigTarget{
							Applications: []pacv2beta2.PolicyTargetApplication{
								{
									Kind:      "Kustomization",
									Name:      "app-a",
									Namespace: "namespace-1",
								},
							},
						},
					},
					Status: pacv2beta2.PolicyConfigStatus{
						Status:          "Warning",
						MissingPolicies: []string{"policy-3"},
					},
				},
				&pacv2beta2.PolicyConfig{
					TypeMeta: metav1.TypeMeta{
						APIVersion: pacv2beta2.GroupVersion.Identifier(),
						Kind:       pacv2beta2.PolicyConfigKind,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "policyconfig-3",
					},
					Spec: pacv2beta2.PolicyConfigSpec{
						Config: map[string]pacv2beta2.PolicyConfigConfig{
							"policy-1": {
								Parameters: map[string]apiextensionsv1.JSON{
									"param-i":  {Raw: []byte{}},
									"param-ii": {Raw: []byte{}},
								},
							},
							"policy-x": {
								Parameters: map[string]apiextensionsv1.JSON{
									"param-i":  {Raw: []byte{}},
									"param-ii": {Raw: []byte{}},
								},
							},
							"policy-z": {
								Parameters: map[string]apiextensionsv1.JSON{
									"param-i":  {Raw: []byte{}},
									"param-ii": {Raw: []byte{}},
								},
							},
						},
						Match: pacv2beta2.PolicyConfigTarget{
							Resources: []pacv2beta2.PolicyTargetResource{
								{
									Kind:      "Deployment",
									Name:      "dep-i",
									Namespace: "namespace-1",
								},
							},
						},
					},
					Status: pacv2beta2.PolicyConfigStatus{
						Status:          "Warning",
						MissingPolicies: []string{"policy-x", "policy-z"},
					},
				},
				&pacv2beta2.PolicyConfig{
					TypeMeta: metav1.TypeMeta{
						APIVersion: pacv2beta2.GroupVersion.Identifier(),
						Kind:       pacv2beta2.PolicyConfigKind,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "policyconfig-4",
					},
					Spec: pacv2beta2.PolicyConfigSpec{
						Config: map[string]pacv2beta2.PolicyConfigConfig{
							"policy-1": {
								Parameters: map[string]apiextensionsv1.JSON{
									"param-1": {Raw: []byte{}},
									"param-2": {Raw: []byte{}},
								},
							},
							"policy-2": {
								Parameters: map[string]apiextensionsv1.JSON{
									"param-1": {Raw: []byte{}},
									"param-2": {Raw: []byte{}},
								},
							},
						},
						Match: pacv2beta2.PolicyConfigTarget{
							Workspaces: []string{"tenant-1", "tenant-2"},
						},
					},
					Status: pacv2beta2.PolicyConfigStatus{
						Status: "OK",
					},
				},
			},
		},
		{
			name: "leaf",
			state: []runtime.Object{
				&pacv2beta2.PolicyConfig{
					TypeMeta: metav1.TypeMeta{
						APIVersion: pacv2beta2.GroupVersion.Identifier(),
						Kind:       pacv2beta2.PolicyConfigKind,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "policyconfig-5",
					},
					Spec: pacv2beta2.PolicyConfigSpec{
						Config: map[string]pacv2beta2.PolicyConfigConfig{
							"policy-1": {
								Parameters: map[string]apiextensionsv1.JSON{
									"param-1": {Raw: []byte{}},
									"param-2": {Raw: []byte{}},
								},
							},
							"policy-2": {
								Parameters: map[string]apiextensionsv1.JSON{
									"param-1": {Raw: []byte{}},
									"param-2": {Raw: []byte{}},
								},
							},
						},
						Match: pacv2beta2.PolicyConfigTarget{
							Namespaces: []string{"namespace-1", "namespace-2"},
						},
					},
					Status: pacv2beta2.PolicyConfigStatus{
						Status: "OK",
					},
				},
			},
		},
	}

	tests := []struct {
		request  *capiv1_proto.ListPolicyConfigsRequest
		response *capiv1_proto.ListPolicyConfigsResponse
		err      bool
	}{
		{
			request: &capiv1_proto.ListPolicyConfigsRequest{},
			response: &capiv1_proto.ListPolicyConfigsResponse{
				PolicyConfigs: []*capiv1_proto.PolicyConfigListItem{
					{
						Name:          "policyconfig-1",
						ClusterName:   "management",
						TotalPolicies: int32(2),
						Match:         "namespaces",
						Status:        "OK",
						Age:           "0001-01-01T00:00:00Z",
					},
					{
						Name:          "policyconfig-2",
						ClusterName:   "management",
						TotalPolicies: int32(1),
						Match:         "apps",
						Status:        "Warning",
						Age:           "0001-01-01T00:00:00Z",
					},
					{
						Name:          "policyconfig-3",
						ClusterName:   "management",
						TotalPolicies: int32(3),
						Match:         "resources",
						Status:        "Warning",
						Age:           "0001-01-01T00:00:00Z",
					},
					{
						Name:          "policyconfig-4",
						ClusterName:   "management",
						TotalPolicies: int32(2),
						Match:         "workspaces",
						Status:        "OK",
						Age:           "0001-01-01T00:00:00Z",
					},
					{
						Name:          "policyconfig-5",
						ClusterName:   "leaf",
						TotalPolicies: int32(2),
						Match:         "namespaces",
						Status:        "OK",
						Age:           "0001-01-01T00:00:00Z",
					},
				},
				Total:  5,
				Errors: []*capiv1_proto.ListError{},
			},
			err: false,
		},
	}

	clustersClients := map[string]client.Client{}
	for _, cluster := range clusters {
		clustersClients[cluster.name] = createClient(t, cluster.state...)
	}

	s := getServer(t, clustersClients, nil)

	for _, tt := range tests {
		res, err := s.ListPolicyConfigs(context.Background(), tt.request)
		if err != nil {
			if tt.err {
				continue
			}
			t.Fatalf("got unexpected error when getting policy config, error: %v", err)
		}
		assert.ElementsMatch(t, tt.response.PolicyConfigs, res.PolicyConfigs, "policy configs do not match expected configs")
		assert.Equal(t, tt.response.Total, res.Total, "total config number is not correct")
	}
}

// TestGetPolicyConfig executes unittests for GetPolicyConfig
func TestGetPolicyConfig(t *testing.T) {

	clusters := []struct {
		name  string
		state []runtime.Object
	}{
		{
			name: "management",
			state: []runtime.Object{
				&pacv2beta2.PolicyConfig{
					TypeMeta: metav1.TypeMeta{
						APIVersion: pacv2beta2.GroupVersion.Identifier(),
						Kind:       pacv2beta2.PolicyConfigKind,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "policyconfig-1",
					},
					Spec: pacv2beta2.PolicyConfigSpec{
						Config: map[string]pacv2beta2.PolicyConfigConfig{
							"policy-1": {
								Parameters: map[string]apiextensionsv1.JSON{
									"param-1": {Raw: []byte(`"val"`)},
									"param-2": {Raw: []byte("2")},
								},
							},
							"policy-2": {
								Parameters: map[string]apiextensionsv1.JSON{
									"param-1": {Raw: []byte(`"val"`)},
									"param-2": {Raw: []byte("2")},
								},
							},
						},
						Match: pacv2beta2.PolicyConfigTarget{
							Namespaces: []string{"namespace-1", "namespace-2"},
						},
					},
					Status: pacv2beta2.PolicyConfigStatus{
						Status:          "OK",
						MissingPolicies: []string{},
					},
				},
				&pacv2beta2.PolicyConfig{
					TypeMeta: metav1.TypeMeta{
						APIVersion: pacv2beta2.GroupVersion.Identifier(),
						Kind:       pacv2beta2.PolicyConfigKind,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "policyconfig-2",
					},
					Spec: pacv2beta2.PolicyConfigSpec{
						Config: map[string]pacv2beta2.PolicyConfigConfig{
							"policy-3": {
								Parameters: map[string]apiextensionsv1.JSON{
									"param-a": {Raw: []byte(`"val"`)},
									"param-b": {Raw: []byte("2")},
								},
							},
						},
						Match: pacv2beta2.PolicyConfigTarget{
							Applications: []pacv2beta2.PolicyTargetApplication{
								{
									Kind:      "Kustomization",
									Name:      "app-a",
									Namespace: "namespace-1",
								},
							},
						},
					},
					Status: pacv2beta2.PolicyConfigStatus{
						Status:          "Warning",
						MissingPolicies: []string{"policy-3"},
					},
				},
				&pacv2beta2.PolicyConfig{
					TypeMeta: metav1.TypeMeta{
						APIVersion: pacv2beta2.GroupVersion.Identifier(),
						Kind:       pacv2beta2.PolicyConfigKind,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "policyconfig-3",
					},
					Spec: pacv2beta2.PolicyConfigSpec{
						Config: map[string]pacv2beta2.PolicyConfigConfig{
							"policy-1": {
								Parameters: map[string]apiextensionsv1.JSON{
									"param-1": {Raw: []byte(`"val"`)},
									"param-2": {Raw: []byte("2")},
								},
							},
							"policy-3": {
								Parameters: map[string]apiextensionsv1.JSON{
									"param-a": {Raw: []byte(`"val"`)},
									"param-b": {Raw: []byte("2")},
								},
							},
							"policy-4": {
								Parameters: map[string]apiextensionsv1.JSON{
									"param-i":  {Raw: []byte(`"val"`)},
									"param-ii": {Raw: []byte("2")},
								},
							},
						},
						Match: pacv2beta2.PolicyConfigTarget{
							Resources: []pacv2beta2.PolicyTargetResource{
								{
									Kind:      "Deployment",
									Name:      "dep-i",
									Namespace: "namespace-1",
								},
							},
						},
					},
					Status: pacv2beta2.PolicyConfigStatus{
						Status:          "Warning",
						MissingPolicies: []string{"policy-3", "policy-4"},
					},
				},
				&pacv2beta2.PolicyConfig{
					TypeMeta: metav1.TypeMeta{
						APIVersion: pacv2beta2.GroupVersion.Identifier(),
						Kind:       pacv2beta2.PolicyConfigKind,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "policyconfig-4",
					},
					Spec: pacv2beta2.PolicyConfigSpec{
						Config: map[string]pacv2beta2.PolicyConfigConfig{
							"policy-1": {
								Parameters: map[string]apiextensionsv1.JSON{
									"param-1": {Raw: []byte(`"val"`)},
									"param-2": {Raw: []byte("2")},
								},
							},
						},
						Match: pacv2beta2.PolicyConfigTarget{
							Workspaces: []string{"tenant-1", "tenant-2"},
						},
					},
					Status: pacv2beta2.PolicyConfigStatus{
						Status:          "OK",
						MissingPolicies: []string{},
					},
				},
				makePolicy(t, func(p *pacv2beta2.Policy) {
					p.ObjectMeta.Name = "policy-1"
					p.Spec.ID = "policy-1"
					p.Spec.Description = "this is policy 1"
				}),
				makePolicy(t, func(p *pacv2beta2.Policy) {
					p.ObjectMeta.Name = "policy-2"
					p.Spec.ID = "policy-2"
					p.Spec.Description = "this is policy 2"
				}),
			},
		},
	}

	tests := []struct {
		request  *capiv1_proto.GetPolicyConfigRequest
		response *capiv1_proto.GetPolicyConfigResponse
		err      bool
	}{
		{
			request: &capiv1_proto.GetPolicyConfigRequest{
				Name:        "policyconfig-1",
				ClusterName: "management",
			},
			response: &capiv1_proto.GetPolicyConfigResponse{
				Name:          "policyconfig-1",
				ClusterName:   "management",
				TotalPolicies: int32(2),
				Status:        "OK",
				Age:           "0001-01-01T00:00:00Z",
				MatchType:     "namespaces",
				Match: &capiv1_proto.PolicyConfigMatch{
					Namespaces: []string{"namespace-1", "namespace-2"},
				},
				Policies: []*capiv1_proto.PolicyConfigPolicy{
					{
						Id:          "policy-1",
						Name:        "Missing Owner Label",
						Description: "this is policy 1",
						Parameters: map[string]*structpb.Value{
							"param-1": {
								Kind: &structpb.Value_StringValue{
									StringValue: "val",
								},
							},
							"param-2": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 2,
								},
							},
						},
						Status: "OK",
					},
					{
						Id:          "policy-2",
						Name:        "Missing Owner Label",
						Description: "this is policy 2",
						Parameters: map[string]*structpb.Value{
							"param-1": {
								Kind: &structpb.Value_StringValue{
									StringValue: "val",
								},
							},
							"param-2": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 2,
								},
							},
						},
						Status: "OK",
					},
				},
			},
		},
		{
			request: &capiv1_proto.GetPolicyConfigRequest{
				Name:        "policyconfig-2",
				ClusterName: "management",
			},
			response: &capiv1_proto.GetPolicyConfigResponse{
				Name:          "policyconfig-2",
				ClusterName:   "management",
				TotalPolicies: int32(1),
				Status:        "Warning",
				Age:           "0001-01-01T00:00:00Z",
				MatchType:     "apps",
				Match: &capiv1_proto.PolicyConfigMatch{
					Apps: []*capiv1_proto.PolicyConfigApplicationMatch{
						{
							Name:      "app-a",
							Namespace: "namespace-1",
							Kind:      "Kustomization",
						},
					},
				},
				Policies: []*capiv1_proto.PolicyConfigPolicy{
					{
						Id:          "policy-3",
						Name:        "",
						Description: "",
						Parameters: map[string]*structpb.Value{
							"param-a": {
								Kind: &structpb.Value_StringValue{
									StringValue: "val",
								},
							},
							"param-b": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 2,
								},
							},
						},
						Status: "Warning",
					},
				},
			},
		},
		{
			request: &capiv1_proto.GetPolicyConfigRequest{
				Name:        "policyconfig-3",
				ClusterName: "management",
			},
			response: &capiv1_proto.GetPolicyConfigResponse{
				Name:          "policyconfig-3",
				ClusterName:   "management",
				TotalPolicies: int32(3),
				Status:        "Warning",
				Age:           "0001-01-01T00:00:00Z",
				MatchType:     "resources",
				Match: &capiv1_proto.PolicyConfigMatch{
					Resources: []*capiv1_proto.PolicyConfigResourceMatch{
						{
							Name:      "dep-i",
							Namespace: "namespace-1",
							Kind:      "Deployment",
						},
					},
				},
				Policies: []*capiv1_proto.PolicyConfigPolicy{
					{
						Id:          "policy-1",
						Name:        "Missing Owner Label",
						Description: "this is policy 1",
						Status:      "OK",
						Parameters: map[string]*structpb.Value{
							"param-1": {
								Kind: &structpb.Value_StringValue{
									StringValue: "val",
								},
							},
							"param-2": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 2,
								},
							},
						},
					},
					{
						Id:          "policy-3",
						Name:        "",
						Description: "",
						Status:      "Warning",
						Parameters: map[string]*structpb.Value{
							"param-a": {
								Kind: &structpb.Value_StringValue{
									StringValue: "val",
								},
							},
							"param-b": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 2,
								},
							},
						},
					},
					{
						Id:          "policy-4",
						Name:        "",
						Description: "",
						Status:      "Warning",
						Parameters: map[string]*structpb.Value{
							"param-i": {
								Kind: &structpb.Value_StringValue{
									StringValue: "val",
								},
							},
							"param-ii": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 2,
								},
							},
						},
					},
				},
			},
		},
		{
			request: &capiv1_proto.GetPolicyConfigRequest{
				Name:        "policyconfig-4",
				ClusterName: "management",
			},
			response: &capiv1_proto.GetPolicyConfigResponse{
				Name:          "policyconfig-4",
				ClusterName:   "management",
				TotalPolicies: int32(1),
				Status:        "OK",
				Age:           "0001-01-01T00:00:00Z",
				MatchType:     "workspaces",
				Match: &capiv1_proto.PolicyConfigMatch{
					Workspaces: []string{"tenant-1", "tenant-2"},
				},
				Policies: []*capiv1_proto.PolicyConfigPolicy{
					{
						Id:          "policy-1",
						Name:        "Missing Owner Label",
						Description: "this is policy 1",
						Parameters: map[string]*structpb.Value{
							"param-1": {
								Kind: &structpb.Value_StringValue{
									StringValue: "val",
								},
							},
							"param-2": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 2,
								},
							},
						},
						Status: "OK",
					},
				},
			},
		},
		{
			request: &capiv1_proto.GetPolicyConfigRequest{
				Name:        uuid.NewString(),
				ClusterName: uuid.NewString(),
			},
			err: true,
		},
	}

	clustersClients := map[string]client.Client{}
	for _, cluster := range clusters {
		clustersClients[cluster.name] = createClient(t, cluster.state...)
	}

	s := getServer(t, clustersClients, nil)

	for _, tt := range tests {
		res, err := s.GetPolicyConfig(context.Background(), tt.request)
		if err != nil {
			if tt.err {
				continue
			}
			t.Fatalf("unexpected error: %v", err)
		}
		assert.Equal(t, tt.response.Name, res.Name, "policy config name is not equal")
		assert.Equal(t, tt.response.ClusterName, res.ClusterName, "policy config cluster name is not equal")
		assert.Equal(t, tt.response.TotalPolicies, res.TotalPolicies, "policy config total policies is not equal")
		assert.Equal(t, tt.response.Status, res.Status, "policy config status is not equal")
		assert.Equal(t, tt.response.Age, res.Age, "policy config age is not equal")
		assert.Equal(t, tt.response.MatchType, res.MatchType, "policy config matchType is not equal")
		assert.Equal(t, tt.response.Match.Namespaces, res.Match.Namespaces, "policy config match namespaces is not equal")
		assert.Equal(t, tt.response.Match.Apps, res.Match.Apps, "policy config match apps is not equal")
		assert.Equal(t, tt.response.Match.Resources, res.Match.Resources, "policy config match resources is not equal")
		assert.Equal(t, tt.response.Match.Workspaces, res.Match.Workspaces, "policy config match workspaces is not equal")

		// create a map from result policies to compare
		resPoliciesMap := map[string]*capiv1_proto.PolicyConfigPolicy{}
		for _, policy := range res.Policies {
			resPoliciesMap[policy.Id] = policy
		}

		// for each policy check if the parameters are equal
		for _, policy := range tt.response.Policies {
			// check if the policy exists in the result
			_, ok := resPoliciesMap[policy.Id]
			if !ok {
				assert.Fail(t, "policy is not found in the result")
			} else {
				assert.Equal(t, policy.Id, resPoliciesMap[policy.Id].Id, "policy id is not equal")
				assert.Equal(t, policy.Name, resPoliciesMap[policy.Id].Name, "policy name is not equal")
				assert.Equal(t, policy.Description, resPoliciesMap[policy.Id].Description, "policy description is not equal")
				assert.Equal(t, policy.Status, resPoliciesMap[policy.Id].Status, "policy status is not equal")
				for param, paramVal := range policy.Parameters {
					// check if the parameter exists in the result
					_, ok := resPoliciesMap[policy.Id].Parameters[param]
					if !ok {
						assert.Fail(t, "policy parameter is not found in the result")
					} else {
						assert.Equal(t, paramVal.AsInterface(), resPoliciesMap[policy.Id].Parameters[param].AsInterface(), "policy parameters are not equal")
					}

				}
			}

		}

	}
}
