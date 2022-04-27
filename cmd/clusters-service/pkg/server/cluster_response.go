package server

import (
	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ToClusterResponse(c *gitopsv1alpha1.GitopsCluster) *capiv1_proto.GitopsCluster {
	res := &capiv1_proto.GitopsCluster{
		Name:        c.GetName(),
		Namespace:   c.GetNamespace(),
		Annotations: c.Annotations,
		Labels:      c.Labels,
		Conditions:  mapConditions(c.Status.Conditions),
	}

	if c.Spec.CAPIClusterRef != nil {
		res.CapiClusterRef = &capiv1_proto.GitopsClusterRef{Name: c.Spec.CAPIClusterRef.Name}
	}

	if c.Spec.SecretRef != nil {
		res.SecretRef = &capiv1_proto.GitopsClusterRef{Name: c.Spec.SecretRef.Name}
	}

	return res
}

func mapConditions(conditions []metav1.Condition) []*capiv1_proto.Condition {
	out := []*capiv1_proto.Condition{}

	for _, c := range conditions {
		out = append(out, &capiv1_proto.Condition{
			Type:      c.Type,
			Status:    string(c.Status),
			Reason:    c.Reason,
			Message:   c.Message,
			Timestamp: c.LastTransitionTime.String(),
		})
	}

	return out
}
