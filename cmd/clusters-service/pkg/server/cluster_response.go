package server

import (
	"context"
	"fmt"
	"strconv"

	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"k8s.io/apimachinery/pkg/api/errors"
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
		Type:        c.GetObjectKind().GroupVersionKind().Kind,
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
func AddCAPIClusters(ctx context.Context, kubeClient client.Client, clusters []*capiv1_proto.GitopsCluster) ([]*capiv1_proto.GitopsCluster, error) {
	capiCluster := &clusterv1.Cluster{}

	for _, cluster := range clusters {
		if cluster.CapiClusterRef != nil {
			err := kubeClient.Get(ctx, client.ObjectKey{
				Name:      cluster.GetName(),
				Namespace: cluster.GetNamespace(),
			}, capiCluster)
			if err != nil {
				if errors.IsNotFound(err) {
					continue
				}
				return nil, fmt.Errorf("failed to get capi-cluster: %w", err)
			}

			cpInitialized := false

			for _, cond := range capiCluster.Status.Conditions {
				if cond.Type == clusterv1.ControlPlaneInitializedCondition {
					var err error
					cpInitialized, err = strconv.ParseBool(string(cond.Status))
					if err != nil {
						return nil, fmt.Errorf(
							"could not parse bool from '%s'. Check the %s condition of cluster %s/%s: %w",
							cond.Status, cond.Type, capiCluster.Namespace, capiCluster.Name, err)
					}
				}
			}

			clusterStatus := &capiv1_proto.CapiClusterStatus{
				Phase:                   capiCluster.Status.Phase,
				InfrastructureReady:     capiCluster.Status.InfrastructureReady,
				ControlPlaneInitialized: cpInitialized,
				ControlPlaneReady:       capiCluster.Status.ControlPlaneReady,
				ObservedGeneration:      capiCluster.Status.ObservedGeneration,
				Conditions:              mapCapiConditions(capiCluster.Status.Conditions),
			}

			capiClusterRes := &capiv1_proto.CapiCluster{
				Name:        capiCluster.GetName(),
				Namespace:   capiCluster.GetNamespace(),
				Annotations: capiCluster.GetAnnotations(),
				Labels:      capiCluster.GetLabels(),
				Status:      clusterStatus,
			}

			if capiCluster.Spec.InfrastructureRef != nil {
				capiClusterRes.InfrastructureRef = &capiv1_proto.CapiClusterInfrastructureRef{
					ApiVersion: capiCluster.Spec.InfrastructureRef.APIVersion,
					Kind:       capiCluster.Spec.InfrastructureRef.Kind,
					Name:       capiCluster.Spec.InfrastructureRef.Name,
				}
			}

			cluster.CapiCluster = capiClusterRes
		}
	}

	return clusters, nil
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

func mapCapiConditions(conditions []clusterv1.Condition) []*capiv1_proto.Condition {
	out := []*capiv1_proto.Condition{}

	for _, c := range conditions {
		out = append(out, &capiv1_proto.Condition{
			Type:      string(c.Type),
			Status:    string(c.Status),
			Reason:    c.Reason,
			Message:   c.Message,
			Timestamp: c.LastTransitionTime.String(),
		})
	}

	return out
}
