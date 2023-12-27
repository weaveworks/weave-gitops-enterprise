package configuration

import (
	"github.com/alecthomas/assert"
	"github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	v1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	. "github.com/onsi/gomega"
	clusterreflectorv1alpha1 "github.com/weaveworks/cluster-reflector-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"testing"
)

// TestObjectsKinds test that default object kinds meet the expected contract
// like being in the expected flux api version. For example, flux v1 available
// kinds should be using v1 api version
func TestObjectsKinds(t *testing.T) {
	g := NewWithT(t)

	t.Run("should contain v1 kustomizations", func(t *testing.T) {
		g.Expect(KustomizationObjectKind.Gvk.GroupVersion()).To(BeIdenticalTo(v1.GroupVersion))
	})

	t.Run("should contain v1 gitrepositories", func(t *testing.T) {
		g.Expect(GitRepositoryObjectKind.Gvk.GroupVersion()).To(BeIdenticalTo(sourcev1.GroupVersion))

	})
}

func TestObjectKind_Validate(t *testing.T) {
	g := NewWithT(t)

	t.Run("should return error if gvk is missing", func(t *testing.T) {
		kind := ObjectKind{}
		g.Expect(kind.Validate()).NotTo(BeNil())
	})

	t.Run("should return error if client func is missing", func(t *testing.T) {
		kind := ObjectKind{
			Gvk: schema.GroupVersionKind{
				Kind: "test",
			},
		}
		g.Expect(kind.Validate()).NotTo(BeNil())
	})

	t.Run("should return error if add to scheme func is missing", func(t *testing.T) {
		kind := ObjectKind{
			Gvk: schema.GroupVersionKind{
				Kind: "test",
			},
			NewClientObjectFunc: func() client.Object {
				return nil
			},
		}
		g.Expect(kind.Validate().Error()).To(Equal("missing add to scheme func"))
	})

	t.Run("should return error if status func is missing", func(t *testing.T) {
		kind := ObjectKind{
			Gvk: schema.GroupVersionKind{
				Kind: "test",
			},
			AddToSchemeFunc: func(*runtime.Scheme) error {
				return nil
			},
			NewClientObjectFunc: func() client.Object {
				return nil
			},
			MessageFunc: func(_ client.Object, _ ObjectKind) (string, error) {
				return "", nil
			},
		}
		g.Expect(kind.Validate().Error()).To(Equal("missing status func"))
	})

	t.Run("should return error if message func is missing", func(t *testing.T) {
		kind := ObjectKind{
			Gvk: schema.GroupVersionKind{
				Kind: "test",
			},
			AddToSchemeFunc: func(*runtime.Scheme) error {
				return nil
			},
			NewClientObjectFunc: func() client.Object {
				return nil
			},
			StatusFunc: func(_ client.Object, _ ObjectKind) (ObjectStatus, error) {
				return Success, nil
			},
		}
		g.Expect(kind.Validate().Error()).To(Equal("missing message func"))
	})
}

func TestStatusAndMessage(t *testing.T) {
	tests := []struct {
		name           string
		desiredStatus  ObjectStatus
		desiredMessage string
		obj            client.Object
		objectKind     ObjectKind
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
			objectKind: HelmReleaseObjectKind,
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
			objectKind: KustomizationObjectKind,
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
			objectKind: HelmReleaseObjectKind,
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
			objectKind: KustomizationObjectKind,
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
			objectKind: KustomizationObjectKind,
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
			objectKind: HelmReleaseObjectKind,
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
			objectKind: KustomizationObjectKind,
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
			objectKind: HelmReleaseObjectKind,
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
			objectKind: KustomizationObjectKind,
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
			objectKind: KustomizationObjectKind,
		},
		{
			name:           "AutomatedClusterDiscovery with Ready condition",
			desiredStatus:  Success,
			desiredMessage: "Applied revision: main/1234567890",
			obj: &clusterreflectorv1alpha1.AutomatedClusterDiscovery{
				Status: clusterreflectorv1alpha1.AutomatedClusterDiscoveryStatus{
					Conditions: []metav1.Condition{
						{
							Type:    "Ready",
							Status:  "True",
							Message: "Applied revision: main/1234567890",
						},
					},
				},
			},
			objectKind: AutomatedClusterDiscoveryObjectKind,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := defaultStatusFunc(tt.obj, tt.objectKind)
			assert.NoError(t, err)

			if got != tt.desiredStatus {
				t.Errorf("Status() = %v, want %v", got, tt.desiredStatus)
			}

			msg, err := defaultMessageFunc(tt.obj, tt.objectKind)
			assert.NoError(t, err)

			if msg != tt.desiredMessage {
				t.Errorf("Message() = %v, want %v", got, tt.desiredMessage)
			}
		})
	}
}
