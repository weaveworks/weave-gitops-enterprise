package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes/any"
	policiesv1 "github.com/weaveworks/policy-agent/api/v1"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/apimachinery/pkg/types"
)

func getPolicyParamValue(param policiesv1.PolicyParameters, policyID string) (*anypb.Any, error) {
	if param.Value == nil {
		return nil, nil
	}
	var anyValue *any.Any
	var err error
	switch param.Type {
	case "string":
		value := wrapperspb.String(string(param.Value.Raw))
		anyValue, err = anypb.New(value)
	case "integer":
		intValue, convErr := strconv.Atoi(string(param.Value.Raw))
		if convErr != nil {
			err = convErr
			break
		}
		value := wrapperspb.Int32(int32(intValue))
		anyValue, err = anypb.New(value)
	case "boolean":
		boolValue, convErr := strconv.ParseBool(string(param.Value.Raw))
		if convErr != nil {
			err = convErr
			break
		}
		value := wrapperspb.Bool(boolValue)
		anyValue, err = anypb.New(value)
	case "array":
		var arrayValue []string
		convErr := json.Unmarshal(param.Value.Raw, &arrayValue)
		if convErr != nil {
			err = convErr
			break
		}
		value := &capiv1_proto.PolicyParamRepeatedString{Value: arrayValue}
		anyValue, err = anypb.New(value)
	default:
		return nil, fmt.Errorf("found unsupported policy parameter type %s in policy %s", param.Type, policyID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to serialize parameter value %s in policy %s: %w", param.Name, policyID, err)
	}
	return anyValue, nil
}

func toPolicyResponse(policyCRD policiesv1.Policy) (*capiv1_proto.Policy, error) {
	policySpec := policyCRD.Spec

	var policyLabels []*capiv1_proto.PolicyTargetLabel
	for i := range policySpec.Targets.Labels {
		policyLabels = append(policyLabels, &capiv1_proto.PolicyTargetLabel{
			Values: policySpec.Targets.Labels[i],
		})
	}

	var policyParams []*capiv1_proto.PolicyParam
	for _, param := range policySpec.Parameters {
		policyParam := &capiv1_proto.PolicyParam{
			Name:     param.Name,
			Required: param.Required,
			Type:     param.Type,
		}
		value, err := getPolicyParamValue(param, policySpec.ID)
		if err != nil {
			return nil, err
		}
		policyParam.Value = value
		policyParams = append(policyParams, policyParam)
	}
	policy := &capiv1_proto.Policy{
		Name:        policySpec.Name,
		Id:          policySpec.ID,
		Code:        policySpec.Code,
		Description: policySpec.Description,
		HowToSolve:  policySpec.HowToSolve,
		Category:    policySpec.Category,
		Tags:        policySpec.Tags,
		Severity:    policySpec.Severity,
		Controls:    policySpec.Controls,
		Targets: &capiv1_proto.PolicyTargets{
			Kinds:      policySpec.Targets.Kinds,
			Namespaces: policySpec.Targets.Namespaces,
			Labels:     policyLabels,
		},
		Parameters: policyParams,
		CreatedAt:  policyCRD.CreationTimestamp.Format(time.RFC3339),
	}

	return policy, nil
}

func (s *server) ListPolicies(ctx context.Context, m *capiv1_proto.ListPoliciesRequest) (*capiv1_proto.ListPoliciesResponse, error) {
	client, err := s.clientGetter.Client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load Kubernetes client: %w", err)
	}

	list := policiesv1.PolicyList{}
	err = client.List(ctx, &list)
	if err != nil {
		return nil, fmt.Errorf("error while listing policies: %w", err)
	}

	var policies []*capiv1_proto.Policy
	for i := range list.Items {
		policy, err := toPolicyResponse(list.Items[i])
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}
	return &capiv1_proto.ListPoliciesResponse{
		Policies: policies,
		Total:    int32(len(policies)),
	}, nil
}

func (s *server) GetPolicy(ctx context.Context, m *capiv1_proto.GetPolicyRequest) (*capiv1_proto.GetPolicyResponse, error) {
	client, err := s.clientGetter.Client(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load Kubernetes client: %w", err)
	}

	policyCR := policiesv1.Policy{}
	err = client.Get(ctx, types.NamespacedName{Name: m.PolicyName}, &policyCR)
	if err != nil {
		return nil, fmt.Errorf("error while getting policy %s: %w", m.PolicyName, err)
	}

	policy, err := toPolicyResponse(policyCR)
	if err != nil {
		return nil, err
	}
	return &capiv1_proto.GetPolicyResponse{Policy: policy}, nil
}
