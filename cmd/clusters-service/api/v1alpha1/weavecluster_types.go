package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WeaveClusterSpec defines the desired state of WeaveCluster
type WeaveClusterSpec struct {
	Type             string                 `json:"type,omitempty"`
	Status           string                 `json:"status,omitempty"`
	Label            string                 `json:"label,omitempty"`
	ResourceClusters []CAPIResourceTemplate `json:"resourcetemplates,omitempty"`
}

//+kubebuilder:pruning:PreserveUnknownFields

// WeaveResourceCluster describes a resource to create.
type WeaveResourceCluster struct {
	runtime.RawExtension `json:",inline"`
}

// WeaveClusterStatus defines the observed state of WeaveCluster
type WeaveClusterStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// WeaveCluster is the Schema for the WeaveClusters API
type WeaveCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WeaveClusterSpec   `json:"spec,omitempty"`
	Status WeaveClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WeaveClusterList contains a list of WeaveCluster
type WeaveClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WeaveCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WeaveCluster{}, &WeaveClusterList{})
}
