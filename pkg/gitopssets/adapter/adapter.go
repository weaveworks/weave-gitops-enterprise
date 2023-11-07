package adapter

import (
	ctrl "github.com/weaveworks/gitopssets-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GitOpsSetAdapter struct {
	*ctrl.GitOpsSet
}

func (obj GitOpsSetAdapter) GetLastHandledReconcileRequest() string {
	return obj.Status.GetLastHandledReconcileRequest()
}

func (obj GitOpsSetAdapter) AsClientObject() client.Object {
	return obj.GitOpsSet
}

func (obj GitOpsSetAdapter) GroupVersionKind() schema.GroupVersionKind {
	return ctrl.GroupVersion.WithKind("GitOpsSet")
}

func (obj GitOpsSetAdapter) SetSuspended(suspend bool) error {
	obj.Spec.Suspend = suspend
	return nil
}

func (obj GitOpsSetAdapter) DeepCopyClientObject() client.Object {
	return obj.DeepCopy()
}

func (obj GitOpsSetAdapter) GetConditions() []metav1.Condition {
	return obj.Status.Conditions
}
