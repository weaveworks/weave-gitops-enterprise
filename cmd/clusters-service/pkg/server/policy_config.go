package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/hashicorp/go-multierror"
	pacv2beta2 "github.com/weaveworks/policy-agent/api/v2beta2"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	core "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"google.golang.org/protobuf/types/known/anypb"
	structpb "google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	policyConfigTargetResource     = "resources"
	policyConfigTargetApplication  = "apps"
	policyConfigTargetNamespace    = "namespaces"
	policyConfigTargetWorkspace    = "workspaces"
	policyConfigConfigStatusOK     = "OK"
	policyConfigConfigStausWarning = "Warning"
)

// ListPolicyConfigs lists the policy configs
func (s *server) ListPolicyConfigs(ctx context.Context, req *capiv1_proto.ListPolicyConfigsRequest) (*capiv1_proto.ListPolicyConfigsResponse, error) {
	respErrors := []*capiv1_proto.ListError{}
	clustersClient, err := s.clustersManager.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		if merr, ok := err.(*multierror.Error); ok {
			for _, err := range merr.Errors {
				if cerr, ok := err.(*clustersmngr.ClientError); ok {
					respErrors = append(respErrors, &capiv1_proto.ListError{ClusterName: cerr.ClusterName, Message: cerr.Error()})
				}
			}
		} else {
			return nil, fmt.Errorf("unexpected error while getting clusters client, error: %w", err)
		}
	}
	policyConfigs, policyConfigsListErrors, err := s.listPolicyConfigs(ctx, clustersClient)
	if err != nil {
		return nil, err
	}

	response := capiv1_proto.ListPolicyConfigsResponse{
		Errors:        respErrors,
		PolicyConfigs: policyConfigs,
		Total:         int32(len(policyConfigs)),
	}

	response.Errors = append(response.Errors, policyConfigsListErrors...)
	return &response, nil
}

// listPolicyConfigs helper inner function to list policy configs by using cluster manager client
func (s *server) listPolicyConfigs(ctx context.Context, cl clustersmngr.Client) ([]*capiv1_proto.PolicyConfigListItem, []*capiv1_proto.ListError, error) {
	clusterListErrors := []*capiv1_proto.ListError{}

	list := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &pacv2beta2.PolicyConfigList{}
	})

	if err := cl.ClusteredList(ctx, list, false); err != nil {
		if e, ok := err.(clustersmngr.ClusteredListError); ok {
			for i := range e.Errors {
				if !strings.Contains(e.Errors[i].Error(), "no matches for kind ") {
					clusterListErrors = append(clusterListErrors, &capiv1_proto.ListError{ClusterName: e.Errors[i].Cluster, Message: e.Errors[i].Error()})
				}

			}
		} else {
			if !strings.Contains(e.Error(), "no matches for kind ") {
				return nil, clusterListErrors, fmt.Errorf("failed to list policy configs, error: %w", err)
			}

		}
	}

	policyConfigList := list.Lists()

	policyConfigs := []*capiv1_proto.PolicyConfigListItem{}
	for clusterName, objs := range policyConfigList {
		for i := range objs {
			obj, ok := objs[i].(*pacv2beta2.PolicyConfigList)
			if !ok {
				continue
			}
			for _, item := range obj.Items {
				policyConfig := capiv1_proto.PolicyConfigListItem{
					ClusterName:   clusterName,
					TotalPolicies: int32(len(item.Spec.Config)),
					Name:          item.Name,
					Match:         getPolicyConfigTargetType(item.Spec.Match),
					Status:        item.Status.Status,
					Age:           item.CreationTimestamp.Format(time.RFC3339),
				}

				policyConfigs = append(policyConfigs, &policyConfig)
			}
		}
	}
	return policyConfigs, clusterListErrors, nil
}

// getPolicyConfigTargetType gets policy configs match from policy config target
func getPolicyConfigTargetType(target pacv2beta2.PolicyConfigTarget) string {
	if target.Applications != nil {
		return policyConfigTargetApplication
	} else if target.Namespaces != nil {
		return policyConfigTargetNamespace
	} else if target.Resources != nil {
		return policyConfigTargetResource
	} else if target.Workspaces != nil {
		return policyConfigTargetWorkspace
	} else {
		return "Unknown match target"
	}
}

// GetPolicyConfig gets the policy config details
func (s *server) GetPolicyConfig(ctx context.Context, req *capiv1_proto.GetPolicyConfigRequest) (*capiv1_proto.GetPolicyConfigResponse, error) {
	clustersClient, err := s.clustersManager.GetImpersonatedClientForCluster(ctx, auth.Principal(ctx), req.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	if clustersClient == nil {
		return nil, fmt.Errorf("cluster %s not found", req.ClusterName)
	}

	policyConfig, err := s.getPolicyConfig(ctx, clustersClient, req)
	if err != nil {
		return nil, err
	}

	return policyConfig, nil
}

// getPolicyConfig gets policy config details by using cluster manager client
func (s *server) getPolicyConfig(ctx context.Context, cl clustersmngr.Client, req *capiv1_proto.GetPolicyConfigRequest) (*capiv1_proto.GetPolicyConfigResponse, error) {
	policyConfig := pacv2beta2.PolicyConfig{}
	// get policyconfig object using cluster manager client with the given cluster name and policy config name
	if err := cl.Get(ctx, req.ClusterName, types.NamespacedName{Name: req.Name}, &policyConfig); err != nil {
		return nil, fmt.Errorf("failed to get policy config, error: %w", err)
	}

	policyConfigDetails := capiv1_proto.GetPolicyConfigResponse{
		Name:          policyConfig.Name,
		ClusterName:   req.ClusterName,
		Age:           policyConfig.CreationTimestamp.Format(time.RFC3339),
		Status:        policyConfig.Status.Status,
		MatchType:     getPolicyConfigTargetType(policyConfig.Spec.Match),
		Match:         getPolicyConfigMatch(policyConfig.Spec.Match),
		TotalPolicies: int32(len(policyConfig.Spec.Config)),
	}

	policies, err := getPolicyConfigPolicies(ctx, cl, s, req.ClusterName, &policyConfig)
	if err != nil {
		return nil, err
	}
	policyConfigDetails.Policies = policies

	return &policyConfigDetails, nil
}

// getPolicyConfigMatch gets policy config match from policy config target
func getPolicyConfigMatch(target pacv2beta2.PolicyConfigTarget) *capiv1_proto.PolicyConfigMatch {
	match := &capiv1_proto.PolicyConfigMatch{}
	//check for applications, namespaces, resources and workspaces in policy config target object which is not nil
	//and set the match object accordingly
	if target.Applications != nil {
		for _, app := range target.Applications {
			newApp := &capiv1_proto.PolicyConfigApplicationMatch{
				Name:      app.Name,
				Kind:      app.Kind,
				Namespace: app.Namespace,
			}
			match.Apps = append(match.Apps, newApp)
		}
		return match

	} else if target.Namespaces != nil {
		match.Namespaces = append(match.Namespaces, target.Namespaces...)
		return match

	} else if target.Workspaces != nil {
		match.Workspaces = append(match.Workspaces, target.Workspaces...)
		return match

	} else if target.Resources != nil {
		for _, res := range target.Resources {
			newRes := &capiv1_proto.PolicyConfigResourceMatch{
				Name:      res.Name,
				Kind:      res.Kind,
				Namespace: res.Namespace,
			}
			match.Resources = append(match.Resources, newRes)
		}
		return match

	}
	return match
}

// getPolicyConfigPolicies gets policy config policies from policy config spec

func getPolicyConfigPolicies(ctx context.Context, cl clustersmngr.Client, s *server, clusterName string, item *pacv2beta2.PolicyConfig) ([]*capiv1_proto.PolicyConfigPolicy, error) {
	policies := []*capiv1_proto.PolicyConfigPolicy{}
	for policyID, policyConfig := range item.Spec.Config {

		//convert policy config parameters to structpb.Value
		params := map[string]*structpb.Value{}
		for key, value := range policyConfig.Parameters {
			v := structpb.Value{}
			if err := v.UnmarshalJSON(value.Raw); err != nil {
				return nil, err
			}
			params[key] = &v
		}

		//check if policy exist on MissingPolicies then set status to Warning else OK
		if slices.Contains(item.Status.MissingPolicies, policyID) {
			policyTarget := &capiv1_proto.PolicyConfigPolicy{
				Id:         policyID,
				Parameters: params,
				Status:     policyConfigConfigStausWarning,
			}
			policies = append(policies, policyTarget)

		} else {
			policyCRv2beta2 := pacv2beta2.Policy{}
			if err := cl.Get(ctx, clusterName, types.NamespacedName{Name: policyID}, &policyCRv2beta2); err != nil {
				return nil, fmt.Errorf("error while getting policy %s from cluster %s: %w", policyID, clusterName, err)
			}

			policy, err := policyToPolicyRespone(policyCRv2beta2, clusterName)
			if err != nil {
				return nil, err
			}
			policyTarget := &capiv1_proto.PolicyConfigPolicy{
				Id:          policyID,
				Name:        policy.Name,
				Description: policy.Description,
				Parameters:  params,
				Status:      policyConfigConfigStatusOK,
			}
			policies = append(policies, policyTarget)
		}
	}
	return policies, nil
}

func policyToPolicyRespone(policyCR pacv2beta2.Policy, clusterName string) (*core.PolicyObj, error) {
	policySpec := policyCR.Spec

	policy := &core.PolicyObj{
		Name:      policySpec.Name,
		Id:        policySpec.ID,
		Category:  policySpec.Category,
		Tags:      policySpec.Tags,
		Severity:  policySpec.Severity,
		CreatedAt: policyCR.CreationTimestamp.Format(time.RFC3339),
		Tenant:    policyCR.GetLabels()["toolkit.fluxcd.io/tenant"],
		Modes:     policyCR.Status.Modes,
	}

	var policyLabels []*core.PolicyTargetLabel
	for i := range policySpec.Targets.Labels {
		policyLabels = append(policyLabels, &core.PolicyTargetLabel{
			Values: policySpec.Targets.Labels[i],
		})
	}

	var policyParams []*core.PolicyParam
	for _, param := range policySpec.Parameters {
		policyParam := &core.PolicyParam{
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
	var policyStandards []*core.PolicyStandard
	for _, standard := range policySpec.Standards {
		policyStandards = append(policyStandards, &core.PolicyStandard{
			Id:       standard.ID,
			Controls: standard.Controls,
		})
	}

	policy.Code = policySpec.Code
	policy.Description = policySpec.Description
	policy.HowToSolve = policySpec.HowToSolve
	policy.Standards = policyStandards
	policy.Targets = &core.PolicyTargets{
		Kinds:      policySpec.Targets.Kinds,
		Namespaces: policySpec.Targets.Namespaces,
		Labels:     policyLabels,
	}
	policy.Parameters = policyParams
	policy.ClusterName = clusterName

	return policy, nil
}

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
		value := &core.PolicyParamRepeatedString{Value: arrayValue}
		anyValue, err = anypb.New(value)
	default:
		return nil, fmt.Errorf("found unsupported policy parameter type %s in policy %s", param.Type, policyID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to serialize parameter value %s in policy %s: %w", param.Name, policyID, err)
	}
	return anyValue, nil
}
