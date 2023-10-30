package adapter

import (
	tfctrl "github.com/weaveworks/tf-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TerraformObjectAdapter struct {
	*tfctrl.Terraform
}

func (obj TerraformObjectAdapter) GetLastHandledReconcileRequest() string {
	return obj.Status.GetLastHandledReconcileRequest()
}

func (obj TerraformObjectAdapter) AsClientObject() client.Object {
	return obj.Terraform
}

func (obj TerraformObjectAdapter) GroupVersionKind() schema.GroupVersionKind {
	return tfctrl.GroupVersion.WithKind(tfctrl.TerraformKind)
}

func (obj TerraformObjectAdapter) SetSuspended(suspend bool) error {
	obj.Spec.Suspend = suspend
	return nil
}

func (obj TerraformObjectAdapter) DeepCopyClientObject() client.Object {
	return obj.DeepCopy()
}

func (obj TerraformObjectAdapter) GetConditions() []metav1.Condition {
	return obj.Status.Conditions
}
