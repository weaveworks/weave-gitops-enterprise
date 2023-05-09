package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/hashicorp/go-multierror"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
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
	var respErrors []*capiv1_proto.ListError
	clustersClient, err := s.clustersManager.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		if merr, ok := err.(*multierror.Error); ok {
			for _, err := range merr.Errors {
				if cerr, ok := err.(*clustersmngr.ClientError); ok {
					respErrors = append(respErrors, &capiv1_proto.ListError{ClusterName: cerr.ClusterName, Message: cerr.Error()})
				}
			}
		}
	}
	labelSelector, err := k8sLabels.ValidatedSelectorFromSet(map[string]string{
		"pac.weave.works/type": "Admission"})
	if err != nil {
		return nil, fmt.Errorf("error building selector for events query: %v", err)
	}

	fieldSelectorSet := map[string]string{
		"type": "Warning",
	}

	if m.Application != "" {
		fieldSelectorSet["involvedObject.name"] = m.Application
		fieldSelectorSet["involvedObject.kind"] = m.Kind
	}

	if m.Namespace != "" {
		fieldSelectorSet["involvedObject.namespace"] = m.Namespace
	}

	fieldSelector := k8sFields.SelectorFromSet(fieldSelectorSet)

	opts := []sigsClient.ListOption{}
	if m.Pagination != nil {
		opts = append(opts, sigsClient.Limit(m.Pagination.PageSize))
		opts = append(opts, sigsClient.Continue(m.Pagination.PageToken))
	}
	opts = append(opts, &sigsClient.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: fieldSelector,
	})
	opts = append(opts, sigsClient.InNamespace(v1.NamespaceAll))

	validationsList, err := s.listEvents(ctx, clustersClient, m.ClusterName, false, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting events: %v", err)
	}
	respErrors = append(respErrors, validationsList.Errors...)
	policyviolationlist := capiv1_proto.ListPolicyValidationsResponse{
		Total:         int32(len(validationsList.Validations)),
		Violations:    validationsList.Validations,
		Errors:        respErrors,
		NextPageToken: validationsList.Token,
	}
	return &policyviolationlist, nil
}

func (s *server) GetPolicyValidation(ctx context.Context, m *capiv1_proto.GetPolicyValidationRequest) (*capiv1_proto.GetPolicyValidationResponse, error) {
	clusterClient, err := s.clustersManager.GetImpersonatedClientForCluster(ctx, auth.Principal(ctx), m.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	if m.ClusterName == "" {
		return nil, errRequiredClusterName
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

	validationsList, err := s.listEvents(ctx, clusterClient, m.ClusterName, true, opts)
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

func (s *server) listEvents(ctx context.Context, clusterClient clustersmngr.Client, clusterName string, extraDetails bool, opts []sigsClient.ListOption) (*validationList, error) {
	respErrors := []*capiv1_proto.ListError{}
	clist := clustersmngr.NewClusteredList(func() sigsClient.ObjectList {
		return &v1.EventList{}
	})

	if err := clusterClient.ClusteredList(ctx, clist, true, opts...); err != nil {
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
		PolicyId:    getAnnotation(annotations, "policy_id"),
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
		paramsRaw := getAnnotation(annotations, "parameters")
		if paramsRaw != "" {
			parameter, err := getPolicyValidationParam([]byte(paramsRaw))
			if err != nil {
				return nil, err
			}
			policyValidation.Parameters = parameter
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

func getPolicyValidationParam(raw []byte) ([]*capiv1_proto.PolicyValidationParam, error) {
	var paramsArr []map[string]interface{}
	err := json.Unmarshal(raw, &paramsArr)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal policy validation parameter, error: %v", err)
	}

	var parameters []*capiv1_proto.PolicyValidationParam
	for i := range paramsArr {
		param := capiv1_proto.PolicyValidationParam{
			Name:     paramsArr[i]["name"].(string),
			Type:     paramsArr[i]["type"].(string),
			Required: paramsArr[i]["required"].(bool),
		}

		if val, ok := paramsArr[i]["config_ref"]; ok {
			param.ConfigRef = val.(string)
		}

		val, err := getParamValue(paramsArr[i]["value"])
		if err != nil {
			return nil, err
		}
		param.Value = val
		parameters = append(parameters, &param)
	}
	return parameters, nil
}

func getParamValue(in interface{}) (*any.Any, error) {
	if in == nil {
		return nil, nil
	}
	switch val := in.(type) {
	case string:
		value := wrapperspb.String(val)
		return anypb.New(value)
	case float64:
		value := wrapperspb.Double(val)
		return anypb.New(value)
	case bool:
		value := wrapperspb.Bool(val)
		return anypb.New(value)
	case []interface{}:
		var values []string
		for i := range val {
			values = append(values, fmt.Sprintf("%v", val[i]))
		}
		value := &capiv1_proto.PolicyParamRepeatedString{Value: values}
		return anypb.New(value)
	}
	return nil, nil
}
