package convert

import (
	"fmt"

	ctrl "github.com/weaveworks/gitopssets-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/gitopssets"
)

func GitOpsToProto(gs ctrl.GitOpsSet) *pb.GitOpsSet {
	inv := []*pb.ResourceRef{}

	if gs.Status.Inventory != nil {
		for _, r := range gs.Status.Inventory.Entries {
			inv = append(inv, &pb.ResourceRef{
				ID:      r.ID,
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

	return &pb.GitOpsSet{
		Name:       gs.Name,
		Namespace:  gs.Namespace,
		Inventory:  inv,
		Conditions: conditions,
		Suspended:  gs.Spec.Suspend,

	}
}
