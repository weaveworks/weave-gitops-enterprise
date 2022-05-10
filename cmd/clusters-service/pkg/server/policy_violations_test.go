package server

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"google.golang.org/protobuf/testing/protocmp"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
)

func TestGetPolicyViolation(t *testing.T) {
	tests := []struct {
		name         string
		ViolationId  string
		clusterState []runtime.Object
		event        *corev1.Event
		err          error
		expected     *capiv1_proto.GetPolicyValidationResponse
	}{
		{
			name:        "get policy violation",
			ViolationId: "weave.policies.missing-app-label",
			clusterState: []runtime.Object{
				makeEvent(t),
			},
			event: makeEvent(t),
			expected: &capiv1_proto.GetPolicyValidationResponse{
				Violation: &capiv1_proto.PolicyValidation{
					Id:              "weave.policies.missing-app-label",
					Name:            "Missing app Label",
					ClusterId:       "cluster-1",
					Category:        "Access Control",
					Severity:        "high",
					CreatedAt:       "0001-01-01T00:00:00Z",
					Message:         "Policy event",
					Entity:          "my-deployment",
					Namespace:       "default",
					Description:     "Missing app label",
					HowToSolve:      "how_to_solve",
					ViolatingEntity: `{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"nginx-deployment","namespace":"default","uid":"af912668-957b-46d4-bc7a-51e6994cba56"},"spec":{"template":{"spec":{"containers":[{"image":"nginx:latest","imagePullPolicy":"Always","name":"nginx","ports":[{"containerPort":80,"protocol":"TCP"}]}]}}}}`,
				},
			},
			err: nil,
		},
		{
			name:        "policy violation doesn't exist",
			ViolationId: "weave.policies.not-found",
			clusterState: []runtime.Object{
				makeEvent(t),
			},
			event: makeEvent(t),
			expected: &capiv1_proto.GetPolicyValidationResponse{
				Violation: &capiv1_proto.PolicyValidation{
					Id:              "weave.policies.missing-app-label",
					Name:            "Missing app Label",
					ClusterId:       "cluster-1",
					Category:        "Access Control",
					Severity:        "high",
					CreatedAt:       "0001-01-01T00:00:00Z",
					Message:         "Policy event",
					Entity:          "my-deployment",
					Namespace:       "default",
					Description:     "Missing app label",
					HowToSolve:      "how_to_solve",
					ViolatingEntity: `{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"nginx-deployment","namespace":"default","uid":"af912668-957b-46d4-bc7a-51e6994cba56"},"spec":{"template":{"spec":{"containers":[{"image":"nginx:latest","imagePullPolicy":"Always","name":"nginx","ports":[{"containerPort":80,"protocol":"TCP"}]}]}}}}`,
				},
			},
			err: errors.New("no policy violation found with id weave.policies.not-found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientSet := fakeclientset.NewSimpleClientset()
			_, err := clientSet.CoreV1().Events("default").Create(context.Background(), tt.event, v1.CreateOptions{})

			if err != nil {
				t.Fatalf("failed to create Policy violation Event to kubernets :\n%s", err)
			}

			s := createServer(t, tt.clusterState, "policies", "default", nil, nil, "", nil, clientSet)

			policyViolation, err := s.GetPolicyValidation(context.Background(), &capiv1_proto.GetPolicyValidationRequest{
				ViolationId: tt.ViolationId,
			})
			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to get policy violation:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("unexpected error while getting policy:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tt.expected, policyViolation, protocmp.Transform()); diff != "" {
					t.Fatalf("policy violation didn't match expected:\n%s", diff)
				}
			}
		})
	}
}

func TestListPolicyValidations(t *testing.T) {
	tests := []struct {
		name         string
		clusterState []runtime.Object
		events       []*corev1.Event
		err          error
		expected     *capiv1_proto.ListPolicyValidationsResponse
	}{
		{
			name:         "get policy violation",
			clusterState: []runtime.Object{},
			events: []*corev1.Event{
				makeEvent(t),
				makeEvent(t, func(e *corev1.Event) {
					e.ObjectMeta.Name = "Missing Owner Label - fake-event-2"
					e.InvolvedObject.Namespace = "weave-system"
					e.ObjectMeta.Namespace = "weave-system"
					e.Annotations["policy_name"] = "Missing Owner Label"
					e.Labels["pac.weave.works/id"] = "weave.policies.missing-owner-label"
				}),
			},
			expected: &capiv1_proto.ListPolicyValidationsResponse{
				Violations: []*capiv1_proto.PolicyValidation{
					{
						Id:        "weave.policies.missing-app-label",
						Name:      "Missing app Label",
						ClusterId: "cluster-1",
						Category:  "Access Control",
						Severity:  "high",
						CreatedAt: "0001-01-01T00:00:00Z",
						Message:   "Policy event",
						Entity:    "my-deployment",
						Namespace: "default",
					},
					{
						Id:        "weave.policies.missing-owner-label",
						Name:      "Missing Owner Label",
						ClusterId: "cluster-1",
						Category:  "Access Control",
						Severity:  "high",
						CreatedAt: "0001-01-01T00:00:00Z",
						Message:   "Policy event",
						Entity:    "my-deployment",
						Namespace: "weave-system",
					},
				},
				Total: int32(2),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientSet := fakeclientset.NewSimpleClientset()

			for _, event := range tt.events {
				_, err := clientSet.CoreV1().Events(event.InvolvedObject.Namespace).Create(context.Background(), event, v1.CreateOptions{})
				if err != nil {
					t.Fatalf("failed to create Policy violation Event to kubernets :\n%s", err)
				}
			}

			s := createServer(t, tt.clusterState, "policies", "default", nil, nil, "", nil, clientSet)

			policyViolation, err := s.ListPolicyValidations(context.Background(), &capiv1_proto.ListPolicyValidationsRequest{})
			if err != nil {
				if tt.err == nil {
					t.Fatalf("failed to list policy violation:\n%s", err)
				}
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("unexpected error while getting policy:\n%s", diff)
				}
			} else {
				if policyViolation.Total != tt.expected.Total {
					t.Fatalf("total policy violation didn't match expected:\n%s", cmp.Diff(tt.expected.Total, policyViolation.Total))
				}
				if diff := cmp.Diff(tt.expected.Violations, policyViolation.Violations, protocmp.Transform()); diff != "" {
					t.Fatalf("policy violation didn't match expected:\n%s", diff)
				}
			}
		})
	}
}
