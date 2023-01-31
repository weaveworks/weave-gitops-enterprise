package convert

import (
	"encoding/json"
	"fmt"

	ctrl "github.com/weaveworks/gitopssets-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/gitopssets"
)

func GitOpsToProto(clusterName string, gs ctrl.GitOpsSet) *pb.GitOpsSet {
	inv := []*pb.ResourceRef{}
	if gs.Status.Inventory != nil {
		for _, r := range gs.Status.Inventory.Entries {
			inv = append(inv, &pb.ResourceRef{
				Id:      r.ID,
				Version: r.Version,
			})
		}
	}

	conditions := []*pb.Condition{}
	if gs.Status.Conditions != nil {
		for _, r := range gs.Status.Conditions {
			conditions = append(conditions, &pb.Condition{
				Type:      r.Type,
				Status:    fmt.Sprint(r.Status),
				Reason:    r.Reason,
				Message:   r.Message,
				Timestamp: r.LastTransitionTime.String(),
			})
		}
	}

	generators := []string{}
	for _, gsg := range gs.Spec.Generators  {	
		listJSON, _ := json.Marshal(gsg.List)
		generators = append(generators, string(listJSON))
	}	    

	return &pb.GitOpsSet{
		Name:       gs.Name,
		Namespace:  gs.Namespace,
		SourceRef: &pb.SourceRef{
		    ApiVersion: gs.APIVersion,
			Kind:       gs.GetObjectKind().GroupVersionKind().Kind,
			Name:       gs.Name,
			Namespace:  gs.Namespace,
		},
		Labels: gs.Labels,
		Annotations:         gs.Annotations,
		ClusterName: clusterName,
		Inventory:  inv,
		Conditions: conditions,
		Generators: generators,
		Type:         gs.GetObjectKind().GroupVersionKind().Kind,
		Suspended: gs.Spec.Suspend,
		ObservedGeneration:      gs.Status.ObservedGeneration,
	}
}