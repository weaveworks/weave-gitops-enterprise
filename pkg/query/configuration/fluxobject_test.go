package configuration

import (
	"testing"

	"github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestStatusAndMessage(t *testing.T) {
	tests := []struct {
		name           string
		desiredStatus  ObjectStatus
		desiredMessage string
		obj            client.Object
	}{
		{
			name:           "HelmRelease with Ready condition",
			desiredStatus:  Success,
			desiredMessage: "Helm release sync succeeded",
			obj: &v2beta1.HelmRelease{
				Status: v2beta1.HelmReleaseStatus{
					Conditions: []metav1.Condition{
						{
							Type:    "Ready",
							Status:  "True",
							Message: "Helm release sync succeeded",
						},
					},
				},
			},
		},
		{
			name:           "Kustomization with Ready condition",
			desiredStatus:  Success,
			desiredMessage: "Applied revision: main/1234567890",
			obj: &kustomizev1.Kustomization{
				Status: kustomizev1.KustomizationStatus{
					Conditions: []metav1.Condition{
						{Type: "Ready", Status: "True", Message: "Applied revision: main/1234567890"},
					},
				},
			},
		},
		{
			name:           "HelmRelease with failed Ready condition",
			desiredStatus:  Failed,
			desiredMessage: "Helm release sync failed: failed to download \"fluxcd/flux\" (hint: running `helm repo update` may help)",
			obj: &v2beta1.HelmRelease{
				Status: v2beta1.HelmReleaseStatus{
					Conditions: []metav1.Condition{
						{Type: "Ready", Status: "False", Message: "Helm release sync failed: failed to download \"fluxcd/flux\" (hint: running `helm repo update` may help)"},
					},
				},
			},
		},
		{
			name:           "Kustomization with failed Ready condition",
			desiredStatus:  Failed,
			desiredMessage: "Kustomization apply failed: failed to apply revision: main/1234567890",
			obj: &kustomizev1.Kustomization{
				Status: kustomizev1.KustomizationStatus{
					Conditions: []metav1.Condition{
						{Type: "Ready", Status: "False", Message: "Kustomization apply failed: failed to apply revision: main/1234567890"},
					},
				},
			},
		},
		{
			name:           "Kustomization with Suspended computed status",
			desiredStatus:  Suspended,
			desiredMessage: "",
			obj: &kustomizev1.Kustomization{
				Spec: kustomizev1.KustomizationSpec{
					Suspend: true,
				},
				Status: kustomizev1.KustomizationStatus{
					Conditions: []metav1.Condition{
						{Type: "CustomCondition", Status: "CustomStatus", Message: "CustomMessage"},
					},
				},
			},
		},
		{
			name:           "HelmRelease with NoStatus condition",
			desiredStatus:  NoStatus,
			desiredMessage: "",
			obj: &v2beta1.HelmRelease{
				Status: v2beta1.HelmReleaseStatus{
					Conditions: []metav1.Condition{
						{Type: "-", Status: "DoesNotMatter", Message: "CustomMessage"},
					},
				},
			},
		},
		{
			name:           "Kustomization without Ready and without NoStatus conditions",
			desiredStatus:  Failed,
			desiredMessage: "",
			obj: &kustomizev1.Kustomization{
				Status: kustomizev1.KustomizationStatus{
					Conditions: []metav1.Condition{
						{Type: "CustomCondition", Status: "CustomStatus", Message: "CustomMessage"},
					},
				},
			},
		},
		{
			name:           "HelmRelease with Ready condition and Reconciling computed status",
			desiredStatus:  Reconciling,
			desiredMessage: "Reconciling message for HelmRelease",
			obj: &v2beta1.HelmRelease{
				Status: v2beta1.HelmReleaseStatus{
					Conditions: []metav1.Condition{
						{
							Type:    "Ready",
							Status:  "Unknown",
							Reason:  "Progressing",
							Message: "Reconciling message for HelmRelease",
						},
					},
				},
			},
		},
		{
			name:           "Kustomization with Available condition and Reconciling computed status",
			desiredStatus:  Reconciling,
			desiredMessage: "Reconciling message for Kustomization",
			obj: &kustomizev1.Kustomization{
				Status: kustomizev1.KustomizationStatus{
					Conditions: []metav1.Condition{
						{
							Type:    "Available",
							Status:  "Unknown",
							Reason:  "Progressing",
							Message: "Reconciling message for Kustomization",
						},
					},
				},
			},
		},
		// TODO: Replace Kustomization with a Terraform object after Explorer starts supporting Terraform objects.
		{
			name:           "Fake Terraform object with Ready condition and PendingAction computed status",
			desiredStatus:  PendingAction,
			desiredMessage: "PendingAction message for Terraform object",
			obj: &kustomizev1.Kustomization{
				Status: kustomizev1.KustomizationStatus{
					Conditions: []metav1.Condition{
						{
							Type:    "Ready",
							Status:  "Unknown",
							Reason:  "TerraformPlannedWithChanges",
							Message: "PendingAction message for Terraform object",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaultFluxObjectStatusFunc(tt.obj); got != tt.desiredStatus {
				t.Errorf("Status() = %v, want %v", got, tt.desiredStatus)
			}

			if got := defaultFluxObjectMessageFunc(tt.obj); got != tt.desiredMessage {
				t.Errorf("Message() = %v, want %v", got, tt.desiredMessage)
			}
		})
	}
}
