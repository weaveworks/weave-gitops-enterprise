package server

import (
	"context"
	"fmt"
	"time"

	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (s *server) ListPolicyValidations(ctx context.Context, m *capiv1_proto.ListPolicyValidationsRequest) (*capiv1_proto.ListPolicyValidationsResponse, error) {
	config := ctrl.GetConfigOrDie()
	clientset := kubernetes.NewForConfigOrDie(config)

	policyviolationlist := capiv1_proto.ListPolicyValidationsResponse{}

	events, err := clientset.CoreV1().Events(v1.NamespaceAll).
		List(ctx, metav1.ListOptions{
			LabelSelector: "pac.weave.works/type=Admission",
			FieldSelector: "type=Warning",
		})

	if err != nil {
		return nil, err
	}

	for _, item := range events.Items {
		// TODO: filter by cluster_id
		// if m.ClusterId != "" && m.ClusterId != getAnnotation(item.GetAnnotations(), "cluster_id")
		policyviolationlist.Violations = append(policyviolationlist.Violations, toPolicyValidation(item))

	}
	policyviolationlist.Total = int32(len(events.Items))
	return &policyviolationlist, nil
}

func (s *server) GetPolicyValidation(ctx context.Context, m *capiv1_proto.GetPolicyValidationRequest) (*capiv1_proto.GetPolicyValidationResponse, error) {
	config := ctrl.GetConfigOrDie()
	clientset := kubernetes.NewForConfigOrDie(config)

	events, err := clientset.CoreV1().Events(v1.NamespaceAll).
		List(ctx, metav1.ListOptions{
			LabelSelector: "pac.weave.works/type=Admission, pac.weave.works/id=" + m.ViolationId,
			FieldSelector: "type=Warning",
		})

	if err != nil {
		return nil, err
	}
	if len(events.Items) == 0 {
		return nil, fmt.Errorf("no policy violation found with id %s", m.ViolationId)
	}
	return &capiv1_proto.GetPolicyValidationResponse{
		Violation: toPolicyValidationDetails(events.Items[0]),
	}, nil
}

func toPolicyValidation(item v1.Event) *capiv1_proto.PolicyValidation {
	annotations := item.GetAnnotations()
	return &capiv1_proto.PolicyValidation{
		Id:        getAnnotation(item.GetLabels(), "pac.weave.works/id"),
		Name:      getAnnotation(annotations, "policy_name"),
		ClusterId: getAnnotation(annotations, "cluster_id"),
		Category:  getAnnotation(annotations, "category"),
		Severity:  getAnnotation(annotations, "severity"),
		CreatedAt: item.GetCreationTimestamp().Format(time.RFC3339),
		Message:   item.Message,
		Entity:    item.InvolvedObject.Name,
		Namespace: item.InvolvedObject.Namespace,
	}
}

func toPolicyValidationDetails(item v1.Event) *capiv1_proto.PolicyValidation {
	annotations := item.GetAnnotations()
	var violation = toPolicyValidation(item)
	violation.Description = getAnnotation(annotations, "description")
	violation.HowToSolve = getAnnotation(annotations, "how_to_solve")
	violation.ViolatingEntity = getAnnotation(annotations, "entity_manifest")
	return violation
}

func getAnnotation(annotations map[string]string, key string) string {
	value, ok := annotations[key]
	if !ok {
		return ""
	}
	return value
}
