package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	pacv2beta2 "github.com/weaveworks/policy-agent/api/v2beta2"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	structpb "google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	policyConfigTargetResource    = "resources"
	policyConfigTargetApplication = "apps"
	policyConfigTargetNamespace   = "namespaces"
	policyConfigTargetWorkspace   = "workspaces"
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
	clustersClient, err := s.clustersManager.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("unexpected error while getting clusters client, error: %w", err)
	}

	policyConfig, err := s.getPolicyConfig(ctx, clustersClient, req)
	if err != nil {
		return nil, err
	}

	return policyConfig, nil
}

// getPolicyConfigDetails helper inner function to get policy config details by using cluster manager client
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
		Match:         getPolicyConfigMatch(policyConfig.Spec.Match),
		TotalPolicies: int32(len(policyConfig.Spec.Config)),
	}

	policyConfigDetails.Policies, _ = getPolicyConfigPolicies(ctx, s, req.ClusterName, policyConfig.Spec.Config)
	return &policyConfigDetails, nil
}

// getPolicyConfigMatch gets policy config match from policy config target
func getPolicyConfigMatch(target pacv2beta2.PolicyConfigTarget) *capiv1_proto.PolicyConfigMatch {
	match := &capiv1_proto.PolicyConfigMatch{}
	//check for applications, namespaces, resources and workspaces in policy config target object which is not nil
	//and set the match object accordingly
	if target.Applications != nil {
		match.Apps = []*capiv1_proto.PolicyConfigApplicationMatch{}
		for _, app := range target.Applications {
			newApp := &capiv1_proto.PolicyConfigApplicationMatch{
				Name:      app.Name,
				Kind:      app.Kind,
				Namespace: app.Namespace,
			}
			match.Apps = append(match.Apps, newApp)
		}
	} else if target.Namespaces != nil {
		match.Namespaces = []string{}
		for _, ns := range target.Namespaces {
			match.Namespaces = append(match.Namespaces, ns)
		}

	} else if target.Workspaces != nil {
		match.Workspaces = []string{}
		for _, ns := range target.Workspaces {
			match.Workspaces = append(match.Workspaces, ns)
		}
	} else if target.Resources != nil {
		match.Resources = []*capiv1_proto.PolicyConfigResourceMatch{}
		for _, res := range target.Resources {
			newRes := &capiv1_proto.PolicyConfigResourceMatch{
				Name:      res.Name,
				Kind:      res.Kind,
				Namespace: res.Namespace,
			}
			match.Resources = append(match.Resources, newRes)
		}
	}
	return match
}

// getPolicyConfigPolicies gets policy config policies from policy config spec

func getPolicyConfigPolicies(ctx context.Context, s *server, clusterName string, config map[string]pacv2beta2.PolicyConfigConfig) ([]*capiv1_proto.PolicyConfigConfig, error) {
	policies := []*capiv1_proto.PolicyConfigConfig{}
	for policyID, policyConfig := range config {

		// Get policy from policy manager
		policy, err := s.GetPolicy(ctx, &capiv1_proto.GetPolicyRequest{ClusterName: clusterName, PolicyName: policyID})
		if err != nil {
			return nil, err
		}

		params := map[string]*structpb.Value{}
		for key, value := range policyConfig.Parameters {
			v := structpb.Value{}
			if err := v.UnmarshalJSON(value.Raw); err != nil {
				return nil, err
			}
			params[key] = &v
		}

		policyTarget := &capiv1_proto.PolicyConfigConfig{
			Id:          policyID,
			Name:        policy.Policy.Name,
			Description: policy.Policy.Description,
			Parameters:  params,
		}

		policies = append(policies, policyTarget)
	}

	return policies, nil
}
