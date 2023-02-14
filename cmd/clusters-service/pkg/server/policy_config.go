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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
<<<<<<< HEAD
	policyConfigTargetResource    = "resources"
	policyConfigTargetApplication = "apps"
	policyConfigTargetNamespace   = "namespaces"
	policyConfigTargetWorkspace   = "workspaces"
=======
	resource    = "Resources"
	application = "Applications"
	namespace   = "Namespaces"
	workspace   = "Workspaces"
>>>>>>> 788e5f29 (update naming convention)
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
<<<<<<< HEAD
					Status:        item.Status.Status,
=======
					Status:        "TBD",
>>>>>>> 788e5f29 (update naming convention)
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
