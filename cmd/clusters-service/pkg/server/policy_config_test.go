package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
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
<<<<<<< HEAD
=======
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-1",
					},
				},
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-2",
					},
				},
>>>>>>> cd42303c (add unit tests)
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
<<<<<<< HEAD
					Status: pacv2beta2.PolicyConfigStatus{
						Status: "OK",
					},
=======
>>>>>>> cd42303c (add unit tests)
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
<<<<<<< HEAD
							"policy-3": {
=======
							"policy-x": {
>>>>>>> cd42303c (add unit tests)
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
<<<<<<< HEAD
					Status: pacv2beta2.PolicyConfigStatus{
						Status:          "Warning",
						MissingPolicies: []string{"policy-3"},
					},
=======
>>>>>>> cd42303c (add unit tests)
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
<<<<<<< HEAD
							"policy-1": {
=======
							"policy-y": {
>>>>>>> cd42303c (add unit tests)
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
<<<<<<< HEAD
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
=======
>>>>>>> cd42303c (add unit tests)
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
<<<<<<< HEAD
				PolicyConfigs: []*capiv1_proto.PolicyConfigListItem{
					{
						Name:          "policyconfig-1",
						ClusterName:   "management",
						TotalPolicies: int32(2),
						Match:         "namespaces",
						Status:        "OK",
=======
				PolicyConfigs: []*capiv1_proto.PolicyConfig{
					{
						Name:          "policyconfig-1",
						ClusterName:   "management",
						TotalPolicies: 2,
						Match:         "namespaces",
						Status:        "TBD",
>>>>>>> cd42303c (add unit tests)
						Age:           "0001-01-01T00:00:00Z",
					},
					{
						Name:          "policyconfig-2",
						ClusterName:   "management",
<<<<<<< HEAD
						TotalPolicies: int32(1),
						Match:         "apps",
						Status:        "Warning",
=======
						TotalPolicies: 1,
						Match:         "apps",
						Status:        "TBD",
>>>>>>> cd42303c (add unit tests)
						Age:           "0001-01-01T00:00:00Z",
					},
					{
						Name:          "policyconfig-3",
						ClusterName:   "management",
<<<<<<< HEAD
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
=======
						TotalPolicies: 3,
						Match:         "resources",
						Status:        "TBD",
						Age:           "0001-01-01T00:00:00Z",
					},
				},
				Total:  3,
>>>>>>> cd42303c (add unit tests)
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
<<<<<<< HEAD
=======

>>>>>>> cd42303c (add unit tests)
	}
}
