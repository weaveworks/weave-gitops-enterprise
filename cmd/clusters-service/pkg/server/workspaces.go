package server

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	pacv2beta1 "github.com/weaveworks/policy-agent/api/v2beta1"
	pacv2beta2 "github.com/weaveworks/policy-agent/api/v2beta2"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	tenantLabel = "toolkit.fluxcd.io/tenant"
)

func (s *server) ListWorkspaces(ctx context.Context, m *capiv1_proto.ListWorkspacesRequest) (*capiv1_proto.ListWorkspacesResponse, error) {
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

	opts := []client.ListOption{
		client.HasLabels{tenantLabel},
	}

	if m.Pagination != nil {
		opts = append(opts, client.Limit(m.Pagination.PageSize))
		opts = append(opts, client.Continue(m.Pagination.PageToken))
	}

	var continueToken string
	var listNamespaces map[string][]client.ObjectList

	namespaces := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &v1.NamespaceList{}
	})
	if err := clustersClient.ClusteredList(ctx, namespaces, true, opts...); err != nil {
		return nil, fmt.Errorf("failed to list service accounts, error: %v", err)
	}

	continueToken = namespaces.GetContinue()
	listNamespaces = namespaces.Lists()

	workspaces := []*capiv1_proto.WorkspaceItem{}

	for cluster, lists := range listNamespaces {
		tenants := map[string][]string{}
		for i := range lists {
			list, ok := lists[i].(*v1.NamespaceList)
			if !ok {
				continue
			}
			for _, item := range list.Items {
				name, ok := item.GetLabels()[tenantLabel]
				if !ok {
					continue
				}
				if _, ok := tenants[name]; !ok {
					tenants[name] = []string{}
				}
				tenants[name] = append(tenants[name], item.Name)
			}
		}
		for name, namespaces := range tenants {
			workspaces = append(workspaces, &capiv1_proto.WorkspaceItem{
				Name:        name,
				Namespaces:  namespaces,
				ClusterName: cluster,
			})
		}
	}
	return &capiv1_proto.ListWorkspacesResponse{
		Workspaces:    workspaces,
		Total:         int32(len(workspaces)),
		NextPageToken: continueToken,
		Errors:        respErrors,
	}, nil
}

func (s *server) GetWorkspace(ctx context.Context, m *capiv1_proto.GetWorkspaceRequest) (*capiv1_proto.GetWorkspaceResponse, error) {
	if m.ClusterName == "" {
		return nil, fmt.Errorf("required cluster name")
	}

	if m.WorkspaceName == "" {
		return nil, fmt.Errorf("required workspace name")
	}

	clustersClient, err := s.clustersManager.GetImpersonatedClientForCluster(ctx, auth.Principal(ctx), m.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	var namespaces []string
	var serviceAccounts []*capiv1_proto.WorkspaceServiceAccount
	var roles []*capiv1_proto.WorkspaceRole
	var roleBindings []*capiv1_proto.WorkspaceRoleBinding
	var policies []*capiv1_proto.WorkspacePolicy

	opts := []client.ListOption{
		client.MatchingLabels{tenantLabel: m.WorkspaceName},
	}

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var namespaceList v1.NamespaceList
		if err := clustersClient.List(gctx, m.ClusterName, &namespaceList, opts...); err != nil {
			return fmt.Errorf("failed to list workspace roles, error: %s", err)
		}
		namespaces = getNamespaces(namespaceList)
		return nil
	})

	g.Go(func() error {
		var serviceAccountList v1.ServiceAccountList
		if err := clustersClient.List(gctx, m.ClusterName, &serviceAccountList, opts...); err != nil {
			return fmt.Errorf("failed to list workspace service accounts, error: %s", err)
		}
		serviceAccounts = getServiceAccounts(serviceAccountList)
		return nil
	})

	g.Go(func() error {
		var roleList rbacv1.RoleList
		if err := clustersClient.List(gctx, m.ClusterName, &roleList, opts...); err != nil {
			return fmt.Errorf("failed to list workspace roles, error: %s", err)
		}
		roles, err = getRoles(roleList)
		return err
	})

	g.Go(func() error {
		var roleBindingList rbacv1.RoleBindingList
		if err := clustersClient.List(gctx, m.ClusterName, &roleBindingList, opts...); err != nil {
			return fmt.Errorf("failed to list workspace role bindings, error: %s", err)
		}
		roleBindings, err = getRoleBindings(roleBindingList)
		return err
	})

	g.Go(func() error {
		var policiesV2beta2List pacv2beta2.PolicyList
		if err := clustersClient.List(gctx, m.ClusterName, &policiesV2beta2List, opts...); err != nil {
			var policiesV2beta1List pacv2beta1.PolicyList
			if err := clustersClient.List(gctx, m.ClusterName, &policiesV2beta1List, opts...); err != nil {
				return fmt.Errorf("failed to list workspace policies, error: %s", err)
			}
			policies = getPoliciesV2beta1(policiesV2beta1List)
			return nil
		}
		policies = getPoliciesV2beta2(policiesV2beta2List)
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &capiv1_proto.GetWorkspaceResponse{
		Name:            m.WorkspaceName,
		ClusterName:     m.ClusterName,
		Namespaces:      namespaces,
		ServiceAccounts: serviceAccounts,
		Roles:           roles,
		RoleBindings:    roleBindings,
		Policies:        policies,
	}, nil
}

func getNamespaces(list v1.NamespaceList) []string {
	var namespaces []string
	for i := range list.Items {
		namespaces = append(namespaces, list.Items[i].Name)
	}
	return namespaces
}

func getServiceAccounts(list v1.ServiceAccountList) []*capiv1_proto.WorkspaceServiceAccount {
	var serviceAccounts []*capiv1_proto.WorkspaceServiceAccount
	for i := range list.Items {
		serviceAccounts = append(serviceAccounts, &capiv1_proto.WorkspaceServiceAccount{
			Name:      list.Items[i].Name,
			Namespace: list.Items[i].Namespace,
			Timestamp: list.Items[i].CreationTimestamp.String(),
		})
	}
	return serviceAccounts
}

func getRoles(list rbacv1.RoleList) ([]*capiv1_proto.WorkspaceRole, error) {
	var roles []*capiv1_proto.WorkspaceRole
	for i := range list.Items {
		role := capiv1_proto.WorkspaceRole{
			Name:      list.Items[i].Name,
			Namespace: list.Items[i].Namespace,
			Timestamp: list.Items[i].CreationTimestamp.String(),
		}
		for _, rule := range list.Items[i].Rules {
			role.Rules = append(role.Rules, &capiv1_proto.WorkspaceRoleRule{
				Groups:    rule.APIGroups,
				Resources: rule.Resources,
				Verbs:     rule.Verbs,
			})
		}
		yml, err := k8sObjectToYaml(&list.Items[i])
		if err != nil {
			return nil, err
		}
		role.Manifest = yml
		roles = append(roles, &role)
	}
	return roles, nil
}

func getRoleBindings(list rbacv1.RoleBindingList) ([]*capiv1_proto.WorkspaceRoleBinding, error) {
	var roleBindings []*capiv1_proto.WorkspaceRoleBinding
	for i := range list.Items {
		roleBinding := capiv1_proto.WorkspaceRoleBinding{
			Name:      list.Items[i].Name,
			Namespace: list.Items[i].Namespace,
			Timestamp: list.Items[i].CreationTimestamp.String(),
		}
		yml, err := k8sObjectToYaml(&list.Items[i])
		if err != nil {
			return nil, err
		}
		roleBinding.Manifest = yml
		roleBindings = append(roleBindings, &roleBinding)
	}
	return roleBindings, nil
}

func getPoliciesV2beta2(list pacv2beta2.PolicyList) []*capiv1_proto.WorkspacePolicy {
	var policies []*capiv1_proto.WorkspacePolicy
	for i := range list.Items {
		policies = append(policies, &capiv1_proto.WorkspacePolicy{
			Id:        list.Items[i].GetName(),
			Name:      list.Items[i].Spec.Name,
			Category:  list.Items[i].Spec.Category,
			Severity:  list.Items[i].Spec.Severity,
			Timestamp: list.Items[i].CreationTimestamp.String(),
		})
	}
	return policies
}

func getPoliciesV2beta1(list pacv2beta1.PolicyList) []*capiv1_proto.WorkspacePolicy {
	var policies []*capiv1_proto.WorkspacePolicy
	for i := range list.Items {
		policies = append(policies, &capiv1_proto.WorkspacePolicy{
			Id:        list.Items[i].GetName(),
			Name:      list.Items[i].Spec.Name,
			Category:  list.Items[i].Spec.Category,
			Severity:  list.Items[i].Spec.Severity,
			Timestamp: list.Items[i].CreationTimestamp.String(),
		})
	}
	return policies
}

func k8sObjectToYaml(obj client.Object) (string, error) {
	var buf bytes.Buffer
	serializer := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)
	if err := serializer.Encode(obj, &buf); err != nil {
		return "", fmt.Errorf("failed to serialize object, error: %v", err)
	}
	return buf.String(), nil
}
