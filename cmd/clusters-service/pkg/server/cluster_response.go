package server

import (
	"context"
	"fmt"

	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"google.golang.org/protobuf/types/known/anypb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

// AddCAPIClusters returns a list of capi-cluster CRs given a list of clusters
func AddCAPIClusters(ctx context.Context, kubeClient client.Client, clusters []*capiv1_proto.GitopsCluster) (*anypb.Any, error) {
	capiClusters := []string{}
	capiCluster := &clusterv1.Cluster{}

	for _, cluster := range clusters {
		if cluster.CapiClusterRef != nil {
			err := kubeClient.Get(ctx, client.ObjectKey{
				Name:      cluster.GetName(),
				Namespace: cluster.GetNamespace(),
			}, capiCluster)
			if err != nil {
				return nil, err
			}

			capiClusterStr := fmt.Sprint(capiCluster)
			capiClusters = append(capiClusters, capiClusterStr)
		}
	}

	value := &capiv1_proto.CapiClusterRepeatedString{Value: capiClusters}
	capiClustersAny, err := anypb.New(value)
	if err != nil {
		return nil, err
	}
	return capiClustersAny, nil
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
