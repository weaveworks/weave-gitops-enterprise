package server

import (
	"context"
	"fmt"
	"time"

	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	v1 "k8s.io/api/core/v1"
	k8sFields "k8s.io/apimachinery/pkg/fields"
	k8sLabels "k8s.io/apimachinery/pkg/labels"
	sigsClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *server) ListPolicyValidations(ctx context.Context, m *capiv1_proto.ListPolicyValidationsRequest) (*capiv1_proto.ListPolicyValidationsResponse, error) {

	selector, err := k8sLabels.ValidatedSelectorFromSet(map[string]string{
		"pac.weave.works/type": "Admission"})
	if err != nil {
		return nil, fmt.Errorf("error building selector for events query: %v", err)
	}

	fields := k8sFields.OneTermEqualSelector("type", "Warning")

	policyviolationlist := capiv1_proto.ListPolicyValidationsResponse{}
	eventList, err := s.listEvents(ctx, v1.NamespaceAll, selector, fields)

	if err != nil {
		return nil, fmt.Errorf("error getting events: %v", err)
	}
	for _, item := range eventList {
		// TODO: filter by cluster_id
		// if m.ClusterId != "" && m.ClusterId != getAnnotation(item.GetAnnotations(), "cluster_id")
		policyviolationlist.Violations = append(policyviolationlist.Violations, toPolicyValidation(item))

	}
	policyviolationlist.Total = int32(len(eventList))
	return &policyviolationlist, nil
}

func (s *server) GetPolicyValidation(ctx context.Context, m *capiv1_proto.GetPolicyValidationRequest) (*capiv1_proto.GetPolicyValidationResponse, error) {
	selector, err := k8sLabels.ValidatedSelectorFromSet(map[string]string{
		"pac.weave.works/type": "Admission",
		"pac.weave.works/id":   m.ViolationId})

	if err != nil {
		return nil, fmt.Errorf("error building selector for events query: %v", err)
	}
	fields := k8sFields.OneTermEqualSelector("type", "Warning")

	eventList, err := s.listEvents(ctx, v1.NamespaceAll, selector, fields)
	if err != nil {
		return nil, fmt.Errorf("error getting events: %v", err)
	}
	if len(eventList) == 0 {
		return nil, fmt.Errorf("no policy violation found with id %s", m.ViolationId)
	}
	return &capiv1_proto.GetPolicyValidationResponse{
		Violation: toPolicyValidationDetails(eventList[0]),
	}, nil
}

func (s *server) listEvents(ctx context.Context, namespace string, selector k8sLabels.Selector, fields k8sFields.Selector) ([]v1.Event, error) {
	client, err := s.clientGetter.Client(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting client: %v", err)
	}

	eventList := v1.EventList{}
	opts := []sigsClient.ListOption{}
	opts = append(opts, sigsClient.InNamespace(namespace))
	opts = append(opts, &sigsClient.ListOptions{
		LabelSelector: selector,
		FieldSelector: fields,
	})
	err = client.List(ctx, &eventList, opts...)
	if err != nil {
		return nil, fmt.Errorf("error getting events: %v", err)
	}
	return eventList.Items, nil
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
