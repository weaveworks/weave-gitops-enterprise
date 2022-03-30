package server

import (
	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
)

func ToClusterResponse(c *gitopsv1alpha1.GitopsCluster) *capiv1_proto.Cluster {
	res := &capiv1_proto.Cluster{
		Name:   c.GetName(),
		Status: c.Spec.Status,
	}

	return res
}

const (
	// DisplayNameAnnotation is the annotation used for labeling cluster resources
	DisplayNameAnnotation = "capi.weave.works/display-name"
)

func ParseClusterMeta(s *gitopsv1alpha1.GitopsCluster) (*ClusterMeta, error) {
	var objects []Object

	return &ClusterMeta{
		Name:    s.ObjectMeta.Name,
		Status:  s.Spec.Status,
		Objects: objects,
	}, nil
}

// Object contains the details of the object rendered from a cluster
type Object struct {
	Kind        string `json:"kind"`
	APIVersion  string `json:"apiVersion"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

// ClusterMeta contains all the objects extracted from a Cluster
type ClusterMeta struct {
	Name    string   `json:"name"`
	Type    string   `json:"type,omitempty"`
	Status  string   `json:"status,omitempty"`
	Label   string   `json:"label,omitempty"`
	Objects []Object `json:"objects,omitempty"`
}
