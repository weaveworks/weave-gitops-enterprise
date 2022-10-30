package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/hashicorp/go-multierror"
	pacv1 "github.com/weaveworks/policy-agent/api/v1"
	pacv2beta1 "github.com/weaveworks/policy-agent/api/v2beta1"
	pacv2beta2 "github.com/weaveworks/policy-agent/api/v2beta2"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var errRequiredClusterName = errors.New("`clusterName` param is required")

func getPolicyParamValue(param pacv2beta2.PolicyParameters, policyID string) (*anypb.Any, error) {
	if param.Value == nil {
		return nil, nil
	}
	var anyValue *any.Any
	var err error
	switch param.Type {
	case "string":
		var strValue string
		// attempt to clean up extra quotes if not successful show as is
		unquotedValue, UnquoteErr := strconv.Unquote(string(param.Value.Raw))
		if UnquoteErr != nil {
			strValue = string(param.Value.Raw)
		} else {
			strValue = unquotedValue
		}
		value := wrapperspb.String(strValue)
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

func toPolicyResponse(policyCRD pacv2beta2.Policy, clusterName string) (*capiv1_proto.Policy, error) {
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
	var policyStandards []*capiv1_proto.PolicyStandard
	for _, standard := range policySpec.Standards {
		policyStandards = append(policyStandards, &capiv1_proto.PolicyStandard{
			Id:       standard.ID,
			Controls: standard.Controls,
		})
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
		Standards:   policyStandards,
		Targets: &capiv1_proto.PolicyTargets{
			Kinds:      policySpec.Targets.Kinds,
			Namespaces: policySpec.Targets.Namespaces,
			Labels:     policyLabels,
		},
		Parameters:  policyParams,
		CreatedAt:   policyCRD.CreationTimestamp.Format(time.RFC3339),
		ClusterName: clusterName,
		Tenant:      policyCRD.GetLabels()["toolkit.fluxcd.io/tenant"],
	}

	return policy, nil
}

func (s *server) ListPolicies(ctx context.Context, m *capiv1_proto.ListPoliciesRequest) (*capiv1_proto.ListPoliciesResponse, error) {
	respErrors := []*capiv1_proto.ListError{}
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

	opts := []client.ListOption{}
	if m.Pagination != nil {
		opts = append(opts, client.Limit(m.Pagination.PageSize))
		opts = append(opts, client.Continue(m.Pagination.PageToken))
	}

	var continueToken string
	var listsV2beta2 map[string][]client.ObjectList
	var listsV2beta1 map[string][]client.ObjectList
	var listsV1 map[string][]client.ObjectList

	if m.ClusterName == "" {
		clistV2beta2 := clustersmngr.NewClusteredList(func() client.ObjectList {
			return &pacv2beta2.PolicyList{}
		})
		clistV2beta1 := clustersmngr.NewClusteredList(func() client.ObjectList {
			return &pacv2beta1.PolicyList{}
		})
		clistV1 := clustersmngr.NewClusteredList(func() client.ObjectList {
			return &pacv1.PolicyList{}
		})

		var errsV2beta2 clustersmngr.ClusteredListError
		var errsV2beta1 clustersmngr.ClusteredListError
		var errsV1 clustersmngr.ClusteredListError

		if err := clustersClient.ClusteredList(ctx, clistV2beta2, false, opts...); err != nil {
			if !errors.As(err, &errsV2beta2) {
				return nil, fmt.Errorf("error while listing v2beta2 policies: %w", err)
			}
		}
		for _, e := range errsV2beta2.Errors {
			if !strings.Contains(e.Err.Error(), "no matches for kind \"Policy\"") {
				respErrors = append(respErrors, &capiv1_proto.ListError{ClusterName: e.Cluster, Message: e.Err.Error()})
			}
		}

		if err := clustersClient.ClusteredList(ctx, clistV2beta1, false, opts...); err != nil {
			if !errors.As(err, &errsV2beta1) {
				return nil, fmt.Errorf("error while listing v2beta1 policies: %w", err)
			}
		}
		for _, e := range errsV2beta1.Errors {
			if !strings.Contains(e.Err.Error(), "no matches for kind \"Policy\"") {
				respErrors = append(respErrors, &capiv1_proto.ListError{ClusterName: e.Cluster, Message: e.Err.Error()})
			}
		}

		if err := clustersClient.ClusteredList(ctx, clistV1, false, opts...); err != nil {
			if !errors.As(err, &errsV1) {
				return nil, fmt.Errorf("error while listing v1 policies: %w", err)
			}
		}
		for _, e := range errsV1.Errors {
			if !strings.Contains(e.Err.Error(), "no matches for kind \"Policy\"") {
				respErrors = append(respErrors, &capiv1_proto.ListError{ClusterName: e.Cluster, Message: e.Err.Error()})
			}
		}

		continueToken = clistV2beta2.GetContinue()
		listsV2beta2 = clistV2beta2.Lists()
		listsV2beta1 = clistV2beta1.Lists()
		listsV1 = clistV1.Lists()
	} else {
		listV2beta2 := &pacv2beta2.PolicyList{}
		listV2beta1 := &pacv2beta1.PolicyList{}
		listV1 := &pacv1.PolicyList{}

		policiesV2beta2, policiesV2beta1, policiesV1 := true, true, true

		if err := clustersClient.List(ctx, m.ClusterName, listV2beta2, opts...); err != nil {
			policiesV2beta2 = false
		}
		if err := clustersClient.List(ctx, m.ClusterName, listV2beta1, opts...); err != nil {
			policiesV2beta1 = false
		}
		if err := clustersClient.List(ctx, m.ClusterName, listV1, opts...); err != nil {
			policiesV1 = false
		}

		if !(policiesV2beta2 || policiesV2beta1 || policiesV1) {
			return nil, fmt.Errorf("error while listing policies for cluster %s: %w", m.ClusterName, err)
		}

		continueToken = listV2beta2.GetContinue()
		if policiesV1 {
			listsV1 = map[string][]client.ObjectList{m.ClusterName: {listV2beta1}}
		}
		if policiesV2beta1 {
			listsV2beta1 = map[string][]client.ObjectList{m.ClusterName: {listV2beta1}}
		}
		if policiesV2beta2 {
			listsV2beta2 = map[string][]client.ObjectList{m.ClusterName: {listV2beta2}}
		}
	}

	var policies []*capiv1_proto.Policy
	for clusterName, lists := range listsV2beta2 {
		for _, l := range lists {
			list, ok := l.(*pacv2beta2.PolicyList)
			if !ok {
				continue
			}
			for i := range list.Items {
				policy, err := toPolicyResponse(list.Items[i], clusterName)
				if err != nil {
					return nil, err
				}

				policies = append(policies, policy)
			}
		}
	}
	for clusterName, lists := range listsV2beta1 {
		for _, l := range lists {
			list, ok := l.(*pacv2beta1.PolicyList)
			if !ok {
				continue
			}
			for i := range list.Items {
				policy, err := toPolicyResponseV2beta1(list.Items[i], clusterName)
				if err != nil {
					return nil, err
				}

				policies = append(policies, policy)
			}
		}
	}

	for clusterName, lists := range listsV1 {
		for _, l := range lists {
			list, ok := l.(*pacv1.PolicyList)
			if !ok {
				continue
			}
			for i := range list.Items {
				policy, err := toPolicyResponseV1(list.Items[i], clusterName)
				if err != nil {
					return nil, err
				}

				policies = append(policies, policy)
			}
		}
	}

	return &capiv1_proto.ListPoliciesResponse{
		Policies:      policies,
		Total:         int32(len(policies)),
		NextPageToken: continueToken,
		Errors:        respErrors,
	}, nil
}

func (s *server) GetPolicy(ctx context.Context, m *capiv1_proto.GetPolicyRequest) (*capiv1_proto.GetPolicyResponse, error) {
	clustersClient, err := s.clustersManager.GetImpersonatedClientForCluster(ctx, auth.Principal(ctx), m.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	if m.ClusterName == "" {
		return nil, errRequiredClusterName
	}
	policyCRv2beta2 := pacv2beta2.Policy{}
	policyCRv2beta1 := pacv2beta1.Policy{}
	policyCRv1 := pacv1.Policy{}
	policiesV2beta2, policiesV2beta1, policiesV1 := true, true, true
	if err := clustersClient.Get(ctx, m.ClusterName, types.NamespacedName{Name: m.PolicyName}, &policyCRv2beta2); err != nil {
		policiesV2beta2 = false
	}
	if err := clustersClient.Get(ctx, m.ClusterName, types.NamespacedName{Name: m.PolicyName}, &policyCRv2beta1); err != nil {
		policiesV2beta1 = false
	}
	if err := clustersClient.Get(ctx, m.ClusterName, types.NamespacedName{Name: m.PolicyName}, &policyCRv1); err != nil {
		policiesV1 = false
	}
	if !(policiesV2beta2 || policiesV2beta1 || policiesV1) {
		return nil, fmt.Errorf("error while getting policy %s from cluster %s: %w", m.PolicyName, m.ClusterName, err)
	}

	var policy *capiv1_proto.Policy
	if policiesV1 {
		policy, err = toPolicyResponseV1(policyCRv1, m.ClusterName)
		if err != nil {
			return nil, err
		}
	}
	if policiesV2beta1 {
		policy, err = toPolicyResponseV2beta1(policyCRv2beta1, m.ClusterName)
		if err != nil {
			return nil, err
		}
	}
	if policiesV2beta2 {
		policy, err = toPolicyResponse(policyCRv2beta2, m.ClusterName)
		if err != nil {
			return nil, err
		}
	}

	return &capiv1_proto.GetPolicyResponse{Policy: policy}, nil
}
