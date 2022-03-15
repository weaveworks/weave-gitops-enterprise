package server

import (
	"fmt"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	processor "sigs.k8s.io/cluster-api/cmd/clusterctl/client/yamlprocessor"
)

func ToClusterResponse(c *capiv1.WeaveCluster) *capiv1_proto.Cluster {
	res := &capiv1_proto.Cluster{
		Name:   c.GetName(),
		Type:   c.Spec.Type,
		Status: c.Spec.Status,
		Label:  c.Spec.Label,
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

func ParseClusterMeta(s *capiv1.WeaveCluster) (*ClusterMeta, error) {
	proc := processor.NewSimpleProcessor()
	variables := map[string]bool{}
	var objects []Object
	for _, v := range s.Spec.ResourceClusters {
		tv, err := proc.GetVariables(v.RawExtension.Raw)
		if err != nil {
			return nil, fmt.Errorf("failed to get parameters processing cluster: %w", err)
		}
		for _, n := range tv {
			variables[n] = true
		}
		var uv unstructured.Unstructured
		if err := uv.UnmarshalJSON(v.RawExtension.Raw); err != nil {
			return nil, fmt.Errorf("failed to unmarshal resourceCluster: %w", err)
		}
		objects = append(objects, Object{
			Kind:        uv.GetKind(),
			APIVersion:  uv.GetAPIVersion(),
			DisplayName: uv.GetAnnotations()[DisplayNameAnnotation],
		})
	}

	return &ClusterMeta{
		Name:    s.ObjectMeta.Name,
		Type:    s.Spec.Type,
		Status:  s.Spec.Status,
		Label:   s.Spec.Label,
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
