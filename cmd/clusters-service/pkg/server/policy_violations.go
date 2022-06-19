package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	v1 "k8s.io/api/core/v1"
	k8sFields "k8s.io/apimachinery/pkg/fields"
	k8sLabels "k8s.io/apimachinery/pkg/labels"
	sigsClient "sigs.k8s.io/controller-runtime/pkg/client"
)

type validationList struct {
	Validations []*capiv1_proto.PolicyValidation
	Token       string
	Errors      []*capiv1_proto.ListError
}

func (s *server) ListPolicyValidations(ctx context.Context, m *capiv1_proto.ListPolicyValidationsRequest) (*capiv1_proto.ListPolicyValidationsResponse, error) {

	selector, err := k8sLabels.ValidatedSelectorFromSet(map[string]string{
		"pac.weave.works/type": "Admission"})
	if err != nil {
		return nil, fmt.Errorf("error building selector for events query: %v", err)
	}

	fields := k8sFields.OneTermEqualSelector("type", "Warning")

	opts := []sigsClient.ListOption{}
	if m.Pagination != nil {
		opts = append(opts, sigsClient.Limit(m.Pagination.PageSize))
		opts = append(opts, sigsClient.Continue(m.Pagination.PageToken))
	}
	opts = append(opts, &sigsClient.ListOptions{
		LabelSelector: selector,
		FieldSelector: fields,
	})
	opts = append(opts, sigsClient.InNamespace(v1.NamespaceAll))

	validationsList, err := s.listEvents(ctx, m.ClusterName, false, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting events: %v", err)
	}
	policyviolationlist := capiv1_proto.ListPolicyValidationsResponse{
		Total:         int32(len(validationsList.Validations)),
		Violations:    validationsList.Validations,
		Errors:        validationsList.Errors,
		NextPageToken: validationsList.Token,
	}
	return &policyviolationlist, nil
}

func (s *server) GetPolicyValidation(ctx context.Context, m *capiv1_proto.GetPolicyValidationRequest) (*capiv1_proto.GetPolicyValidationResponse, error) {
	if m.ClusterName == "" {
		return nil, requiredClusterNameErr
	}

	selector, err := k8sLabels.ValidatedSelectorFromSet(map[string]string{
		"pac.weave.works/type": "Admission",
		"pac.weave.works/id":   m.ViolationId})

	if err != nil {
		return nil, fmt.Errorf("error building selector for events query: %v", err)
	}
	opts := []sigsClient.ListOption{}

	fields := k8sFields.OneTermEqualSelector("type", "Warning")
	opts = append(opts, &sigsClient.ListOptions{
		LabelSelector: selector,
		FieldSelector: fields,
	})
	opts = append(opts, sigsClient.InNamespace(v1.NamespaceAll))

	validationsList, err := s.listEvents(ctx, m.ClusterName, true, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting events: %v", err)
	}
	if len(validationsList.Errors) > 0 {
		return nil, fmt.Errorf("error getting events: %s", validationsList.Errors[0].Message)
	}
	if len(validationsList.Validations) == 0 {
		return nil, fmt.Errorf("no policy violation found with id %s and cluster: %s", m.ViolationId, m.ClusterName)
	}
	return &capiv1_proto.GetPolicyValidationResponse{
		Violation: validationsList.Validations[0],
	}, nil
}

func (s *server) listEvents(ctx context.Context, clusterName string, extraDetails bool, opts []sigsClient.ListOption) (*validationList, error) {
	clustersClient, err := s.clientsFactory.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %s", err)
	}
	clist := clustersmngr.NewClusteredList(func() sigsClient.ObjectList {
		return &v1.EventList{}
	})

	respErrors := []*capiv1_proto.ListError{}
	if err := clustersClient.ClusteredList(ctx, clist, true, opts...); err != nil {
		var errs clustersmngr.ClusteredListError
		if !errors.As(err, &errs) {
			return nil, fmt.Errorf("error while listing events: %w", err)
		}

		for _, e := range errs.Errors {
			respErrors = append(respErrors, &capiv1_proto.ListError{ClusterName: e.Cluster, Message: e.Err.Error()})
		}
	}

	var validations []*capiv1_proto.PolicyValidation
	for listClusterName, lists := range clist.Lists() {
		if clusterName != "" && listClusterName != clusterName {
			continue
		}
		for _, l := range lists {
			list, ok := l.(*v1.EventList)
			if !ok {
				continue
			}
			for i := range list.Items {
				validation, err := toPolicyValidation(list.Items[i], listClusterName, extraDetails)
				if err != nil {
					return nil, fmt.Errorf("error while getting policy violation event details: %w", err)
				}
				validations = append(validations, validation)
			}
		}
	}

	return &validationList{
		Validations: validations,
		Token:       clist.GetContinue(),
		Errors:      respErrors,
	}, nil
}

func toPolicyValidation(item v1.Event, clusterName string, extraDetails bool) (*capiv1_proto.PolicyValidation, error) {
	annotations := item.GetAnnotations()
	policyValidation := &capiv1_proto.PolicyValidation{
		Id:          getAnnotation(item.GetLabels(), "pac.weave.works/id"),
		Name:        getAnnotation(annotations, "policy_name"),
		ClusterId:   getAnnotation(annotations, "cluster_id"),
		Category:    getAnnotation(annotations, "category"),
		Severity:    getAnnotation(annotations, "severity"),
		CreatedAt:   item.GetCreationTimestamp().Format(time.RFC3339),
		Message:     item.Message,
		Entity:      item.InvolvedObject.Name,
		Namespace:   item.InvolvedObject.Namespace,
		ClusterName: clusterName,
	}
	if extraDetails {
		policyValidation.Description = getAnnotation(annotations, "description")
		policyValidation.HowToSolve = getAnnotation(annotations, "how_to_solve")
		policyValidation.ViolatingEntity = getAnnotation(annotations, "entity_manifest")
		err := json.Unmarshal([]byte(getAnnotation(annotations, "occurrences")), &policyValidation.Occurrences)
		if err != nil {
			return nil, fmt.Errorf("failed to get occurrences from event: %w", err)
		}
	}

	return policyValidation, nil
}

func getAnnotation(annotations map[string]string, key string) string {
	value, ok := annotations[key]
	if !ok {
		return ""
	}
	return value
}
