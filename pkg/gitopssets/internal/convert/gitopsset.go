package convert

import (
	"bytes"
	"fmt"

	ctrl "github.com/weaveworks/gitopssets-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/gitopssets"
	k8s_json "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

func GitOpsToProto(clusterName string, gs ctrl.GitOpsSet) (*pb.GitOpsSet, error) {
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
	var buf bytes.Buffer

	serializer := k8s_json.NewSerializer(k8s_json.DefaultMetaFactory, nil, nil, false)
	if err := serializer.Encode(&gs, &buf); err != nil {
		return nil, err
	}

	fmt.Println("GITOPSSETTOPROTO")
	fmt.Println(buf.String())

	return &pb.GitOpsSet{
		Name:               gs.Name,
		Namespace:          gs.Namespace,
		Labels:             gs.Labels,
		Annotations:        gs.Annotations,
		ClusterName:        clusterName,
		Inventory:          inv,
		Conditions:         conditions,
		Type:               gs.GetObjectKind().GroupVersionKind().Kind,
		Suspended:          gs.Spec.Suspend,
		ObservedGeneration: gs.Status.ObservedGeneration,
		Yaml:               buf.String(),
		ObjectRef: &pb.ObjectRef{
			Kind:        gs.GetObjectKind().GroupVersionKind().Kind,
			Name:        gs.Name,
			Namespace:   gs.Namespace,
			ClusterName: clusterName,
		},
	}, nil
}
