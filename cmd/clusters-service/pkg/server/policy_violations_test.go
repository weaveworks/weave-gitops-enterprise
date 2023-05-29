package server

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"google.golang.org/protobuf/testing/protocmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestGetPolicyViolation(t *testing.T) {
	tests := []struct {
		name         string
		ViolationId  string
		clusterState []runtime.Object
		clusterName  string
		err          error
		expected     *capiv1_proto.GetPolicyValidationResponse
	}{
		{
			name:        "get policy violation",
			ViolationId: "66101548-12c1-4f79-a09a-a12979903fba",
			clusterState: []runtime.Object{
				makeEvent(t),
			},
			clusterName: "Default",
			expected: &capiv1_proto.GetPolicyValidationResponse{
				Violation: &capiv1_proto.PolicyValidation{
					Id:              "66101548-12c1-4f79-a09a-a12979903fba",
					Name:            "Missing app Label",
					PolicyId:        "weave.policies.missing-app-label",
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
					ClusterName:     "Default",
					Occurrences: []*capiv1_proto.PolicyValidationOccurrence{
						{
							Message: "occurrence details",
						},
					},
				},
			},
			err: nil,
		},
		{
			name:        "policy violation doesn't exist",
			ViolationId: "invalid-id",
			clusterState: []runtime.Object{
				makeEvent(t),
			},
			clusterName: "Default",
			err:         errors.New("no policy violation found with id invalid-id and cluster: Default"),
		},
		{
			name:        "cluster name not specified",
			ViolationId: "66101548-12c1-4f79-a09a-a12979903fba",
			clusterState: []runtime.Object{
				makeEvent(t),
			},
			err: errRequiredClusterName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			clientsPool := &clustersmngrfakes.FakeClientsPool{}
			fakeCl := createClient(t, tt.clusterState...)
			clientsPool.ClientsReturns(map[string]client.Client{tt.clusterName: fakeCl})
			clientsPool.ClientReturns(fakeCl, nil)
			clustersClient := clustersmngr.NewClient(clientsPool, map[string][]corev1.Namespace{"Default": {
				corev1.Namespace{},
			}}, logr.Discard())

			fakeFactory := &clustersmngrfakes.FakeClustersManager{}
			fakeFactory.GetImpersonatedClientForClusterReturns(clustersClient, nil)

			s := createServer(t, serverOptions{
				clustersManager: fakeFactory,
			})

			policyViolation, err := s.GetPolicyValidation(context.Background(), &capiv1_proto.GetPolicyValidationRequest{
				ViolationId: tt.ViolationId,
				ClusterName: tt.clusterName,
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
		clusterName  string
		appName      string
		appKind      string
		namespace    string
	}{
		{
			name: "list policy violations",
			clusterState: []runtime.Object{
				makeEvent(t),
				makeEvent(t, func(e *corev1.Event) {
					e.ObjectMeta.Name = "Missing Owner Label - fake-event-2"
					e.InvolvedObject.Namespace = "weave-system"
					e.ObjectMeta.Namespace = "weave-system"
					e.Annotations["policy_name"] = "Missing Owner Label"
					e.Annotations["policy_id"] = "weave.policies.missing-app-label"
					e.Labels["pac.weave.works/id"] = "56701548-12c1-4f79-a09a-a12979903"
				}),
			},
			expected: &capiv1_proto.ListPolicyValidationsResponse{
				Violations: []*capiv1_proto.PolicyValidation{
					{
						Id:          "66101548-12c1-4f79-a09a-a12979903fba",
						Name:        "Missing app Label",
						PolicyId:    "weave.policies.missing-app-label",
						ClusterId:   "cluster-1",
						Category:    "Access Control",
						Severity:    "high",
						CreatedAt:   "0001-01-01T00:00:00Z",
						Message:     "Policy event",
						Entity:      "my-deployment",
						Namespace:   "default",
						ClusterName: "Default",
					},
					{
						Id:          "56701548-12c1-4f79-a09a-a12979903",
						Name:        "Missing Owner Label",
						PolicyId:    "weave.policies.missing-app-label",
						ClusterId:   "cluster-1",
						Category:    "Access Control",
						Severity:    "high",
						CreatedAt:   "0001-01-01T00:00:00Z",
						Message:     "Policy event",
						Entity:      "my-deployment",
						Namespace:   "weave-system",
						ClusterName: "Default",
					},
				},
				Total: int32(2),
			},
		},
		{
			name: "list application policy violations",
			clusterState: []runtime.Object{
				makeEvent(t, func(e *corev1.Event) {
					e.ObjectMeta.Name = "Missing Owner Label - fake-event-2"
					e.InvolvedObject.Namespace = "weave-system"
					e.ObjectMeta.Namespace = "weave-system"
					e.InvolvedObject.Name = "app1"
					e.InvolvedObject.Kind = "HelmRelease"
					e.Annotations["policy_name"] = "Missing Owner Label"
					e.Annotations["policy_id"] = "weave.policies.missing-app-label"
					e.Labels["pac.weave.works/id"] = "56701548-12c1-4f79-a09a-a12979904"
				}),
			},
			expected: &capiv1_proto.ListPolicyValidationsResponse{
				Violations: []*capiv1_proto.PolicyValidation{
					{
						Id:          "56701548-12c1-4f79-a09a-a12979904",
						Name:        "Missing Owner Label",
						PolicyId:    "weave.policies.missing-app-label",
						ClusterId:   "cluster-1",
						Category:    "Access Control",
						Severity:    "high",
						CreatedAt:   "0001-01-01T00:00:00Z",
						Message:     "Policy event",
						Entity:      "app1",
						Namespace:   "weave-system",
						ClusterName: "Default",
					},
				},
				Total: int32(1),
			},
			appName:   "app1",
			appKind:   "HelmRelease",
			namespace: "weave-system",
		},
		{
			name: "list policy violations with cluster filtering",
			clusterState: []runtime.Object{
				makeEvent(t),
			},
			expected:    &capiv1_proto.ListPolicyValidationsResponse{},
			clusterName: "wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientsPool := &clustersmngrfakes.FakeClientsPool{}
			fakeCl := createClient(t, tt.clusterState...)
			clientsPool.ClientsReturns(map[string]client.Client{"Default": fakeCl})
			clientsPool.ClientReturns(fakeCl, nil)
			clustersClient := clustersmngr.NewClient(clientsPool, map[string][]corev1.Namespace{"Default": {
				corev1.Namespace{},
			}}, logr.Discard())

			fakeFactory := &clustersmngrfakes.FakeClustersManager{}
			fakeFactory.GetImpersonatedClientReturns(clustersClient, nil)

			s := createServer(t, serverOptions{
				clustersManager: fakeFactory,
			})
			policyViolation, err := s.ListPolicyValidations(context.Background(), &capiv1_proto.ListPolicyValidationsRequest{
				ClusterName: tt.clusterName,
				Application: tt.appName,
				Kind:        tt.appKind,
				Namespace:   tt.namespace,
			})
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

func TestPartialPolicyValidationsConnectionErrors(t *testing.T) {
	clientsPool := &clustersmngrfakes.FakeClientsPool{}
	fakeCl := createClient(t, makeEvent(t))
	clientsPool.ClientsReturns(map[string]client.Client{"Default": fakeCl})
	clientsPool.ClientReturns(fakeCl, nil)
	clustersClient := clustersmngr.NewClient(clientsPool, map[string][]corev1.Namespace{"Default": {
		corev1.Namespace{},
	}}, logr.Discard())

	clusterErr := clustersmngr.ClientError{ClusterName: "demo", Err: errors.New("failed adding cluster client to pool: connection refused")}
	fakeFactory := &clustersmngrfakes.FakeClustersManager{}
	fakeFactory.GetImpersonatedClientStub = func(ctx context.Context, user *auth.UserPrincipal) (clustersmngr.Client, error) {
		var multi *multierror.Error
		multi = multierror.Append(multi, &clusterErr)
		return clustersClient, multi
	}

	s := createServer(t, serverOptions{
		clustersManager: fakeFactory,
	})

	policyViolation, err := s.ListPolicyValidations(context.Background(), &capiv1_proto.ListPolicyValidationsRequest{})
	if err != nil {
		t.Fatal(err)
	}

	expectValidation := &capiv1_proto.ListPolicyValidationsResponse{
		Violations: []*capiv1_proto.PolicyValidation{
			{
				Id:          "66101548-12c1-4f79-a09a-a12979903fba",
				Name:        "Missing app Label",
				PolicyId:    "weave.policies.missing-app-label",
				ClusterId:   "cluster-1",
				Category:    "Access Control",
				Severity:    "high",
				CreatedAt:   "0001-01-01T00:00:00Z",
				Message:     "Policy event",
				Entity:      "my-deployment",
				Namespace:   "default",
				ClusterName: "Default",
			},
		},
		Total:  int32(1),
		Errors: []*capiv1_proto.ListError{{Message: clusterErr.Error(), ClusterName: clusterErr.ClusterName}},
	}
	if diff := cmp.Diff(expectValidation.Violations, policyViolation.Violations, protocmp.Transform()); diff != "" {
		t.Fatalf("policy violation didn't match expected:\n%s", diff)
	}

	if diff := cmp.Diff(expectValidation.Errors, policyViolation.Errors, protocmp.Transform()); diff != "" {
		t.Fatalf("policy violation errors didn't match expected:\n%s", diff)
	}
}
