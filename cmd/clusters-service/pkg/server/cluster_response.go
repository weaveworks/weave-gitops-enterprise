package server

import (
	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
)

func ToClusterResponse(c *gitopsv1alpha1.GitopsCluster) *capiv1_proto.GitopsCluster {
	res := &capiv1_proto.GitopsCluster{
		Name:        c.GetName(),
		Annotations: c.Annotations,
		Labels:      c.Labels,
	}

	return res
}
