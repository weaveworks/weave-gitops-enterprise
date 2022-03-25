package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	CapiClusterRef CapiClusterRef `json:"capiclusterref,omitempty"`
	SecretRef      SecretRef      `json:"secretref,omitempty"`
	Status         string         `json:"status,omitempty"`
}

// CapiClusterRef contains enough information to locate
// the referenced Kubernetes resource object.
type CapiClusterRef struct {
	Name string `json:"name,omitempty"`
}

// SecretRef contains enough information to locate
// the referenced Kubernetes resource object.
type SecretRef struct {
	Name string `json:"name,omitempty"`
}

//+kubebuilder:pruning:PreserveUnknownFields

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Cluster is the Schema for the Clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
