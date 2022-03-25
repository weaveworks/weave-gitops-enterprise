package server

import (
	"fmt"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
)

func ToClusterResponse(c *capiv1.Cluster) *capiv1_proto.Cluster {
	res := &capiv1_proto.Cluster{
		Name:   c.GetName(),
		Status: c.Spec.Status,
	}

	meta, err := ParseClusterMeta(c)
	if err != nil {
		res.Error = fmt.Sprintf("Couldn't load cluster body: %s", err.Error())
		return res
	}

	for _, o := range meta.Objects {
		res.Objects = append(res.Objects, &capiv1_proto.ClusterObject{
			Kind:        o.Kind,
			ApiVersion:  o.APIVersion,
			Name:        o.Name,
			DisplayName: o.DisplayName,
		})
	}

	return res
}

const (
	// DisplayNameAnnotation is the annotation used for labeling cluster resources
	DisplayNameAnnotation = "capi.weave.works/display-name"
)

func ParseClusterMeta(s *capiv1.Cluster) (*ClusterMeta, error) {
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

// ClusterMeta contains all the objects extracted from a WeaveCluster
type ClusterMeta struct {
	Name    string   `json:"name"`
	Type    string   `json:"type,omitempty"`
	Status  string   `json:"status,omitempty"`
	Label   string   `json:"label,omitempty"`
	Objects []Object `json:"objects,omitempty"`
}
