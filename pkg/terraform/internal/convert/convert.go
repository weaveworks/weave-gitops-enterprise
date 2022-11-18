package convert

import (
	"fmt"

	tfctrl "github.com/weaveworks/tf-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/terraform"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ToPBTerraformObject(clusterName string, tf *tfctrl.Terraform) pb.TerraformObject {

	inv := []*pb.ResourceRef{}

	if tf.Status.Inventory != nil {
		for _, r := range tf.Status.Inventory.Entries {
			inv = append(inv, &pb.ResourceRef{
				Name:       r.Name,
				Type:       r.Type,
				Identifier: r.Identifier,
			})
		}
	}

	conditions := []*pb.Condition{}

	if tf.Status.Conditions != nil {
		for _, r := range tf.Status.Conditions {
			conditions = append(conditions, &pb.Condition{
				Type:      r.Type,
				Status:    fmt.Sprint(r.Status),
				Reason:    r.Reason,
				Message:   r.Message,
				Timestamp: r.LastTransitionTime.String(),
			})
		}
	}

	return pb.TerraformObject{
		Name:        tf.Name,
		Namespace:   tf.Namespace,
		ClusterName: clusterName,

		SourceRef: &pb.SourceRef{
			ApiVersion: tf.Spec.SourceRef.APIVersion,
			Kind:       tf.Spec.SourceRef.Kind,
			Name:       tf.Spec.SourceRef.Name,
			Namespace:  tf.Spec.SourceRef.Namespace,
		},

		AppliedRevision:      tf.Status.LastAppliedRevision,
		Path:                 tf.Spec.Path,
		Interval:             durationToInterval(tf.Spec.Interval),
		DriftDetectionResult: tf.HasDrift(),
		Inventory:            inv,
		Conditions:           conditions,
		Suspended:            tf.Spec.Suspend,
		Type:                 tf.GetObjectKind().GroupVersionKind().Kind,
	}
}

func durationToInterval(duration metav1.Duration) *pb.Interval {
	return &pb.Interval{
		Hours:   int64(duration.Hours()),
		Minutes: int64(duration.Minutes()) % 60,
		Seconds: int64(duration.Seconds()) % 60,
	}
}
