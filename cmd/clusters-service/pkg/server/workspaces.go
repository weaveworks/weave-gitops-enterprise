package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	pacv2beta1 "github.com/weaveworks/policy-agent/api/v2beta1"
	pacv2beta2 "github.com/weaveworks/policy-agent/api/v2beta2"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
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
		} else {
			return nil, fmt.Errorf("unexpected error while getting clusters client, error: %w", err)
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
	if err := clustersClient.ClusteredList(ctx, namespaces, false, opts...); err != nil {
		if !errors.As(err, &clustersmngr.ClusteredListError{}) {
			return nil, fmt.Errorf("failed to list namespaces, error: %w", err)
		}
	}

	continueToken = namespaces.GetContinue()
	listNamespaces = namespaces.Lists()

	workspaces := []*capiv1_proto.Workspace{}

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
			workspaces = append(workspaces, &capiv1_proto.Workspace{
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

func (s *server) GetWorkspace(ctx context.Context, req *capiv1_proto.GetWorkspaceRequest) (*capiv1_proto.GetWorkspaceResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	var list v1.NamespaceList
	if err := s.listWorkspaceResources(ctx, req, &list); err != nil {
		return nil, fmt.Errorf("failed to list workspace namespaces, error: %w", err)
	}

	if len(list.Items) == 0 {
		return nil, fmt.Errorf("workspace %s not found", req.Name)
	}

	var namespaces []string
	for i := range list.Items {
		namespaces = append(namespaces, list.Items[i].Name)
	}

	return &capiv1_proto.GetWorkspaceResponse{
		Name:        req.Name,
		ClusterName: req.ClusterName,
		Namespaces:  namespaces,
	}, nil
}

func (s *server) GetWorkspaceRoles(ctx context.Context, req *capiv1_proto.GetWorkspaceRequest) (*capiv1_proto.GetWorkspaceRolesResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	exists, err := s.checkWorkspaceExists(ctx, req)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("workspace %s not found", req.Name)
	}

	var list rbacv1.RoleList
	if err := s.listWorkspaceResources(ctx, req, &list); err != nil {
		return nil, fmt.Errorf("failed to list workspace roles, error: %w", err)
	}

	var roles []*capiv1_proto.WorkspaceRole
	for i := range list.Items {
		role := capiv1_proto.WorkspaceRole{
			Name:      list.Items[i].Name,
			Namespace: list.Items[i].Namespace,
			Kind:      "Role",
			Timestamp: list.Items[i].CreationTimestamp.Format(time.RFC3339),
		}
		for _, rule := range list.Items[i].Rules {
			role.Rules = append(role.Rules, &capiv1_proto.WorkspaceRoleRule{
				Groups:    rule.APIGroups,
				Resources: rule.Resources,
				Verbs:     rule.Verbs,
			})
		}

		obj := list.Items[i]
		obj.SetGroupVersionKind(rbacv1.SchemeGroupVersion.WithKind("Role"))

		manifest, err := k8sObjectToYaml(&obj)
		if err != nil {
			return nil, err
		}

		role.Manifest = manifest
		roles = append(roles, &role)
	}

	return &capiv1_proto.GetWorkspaceRolesResponse{
		Name:        req.Name,
		ClusterName: req.ClusterName,
		Objects:     roles,
	}, nil
}

func (s *server) GetWorkspaceRoleBindings(ctx context.Context, req *capiv1_proto.GetWorkspaceRequest) (*capiv1_proto.GetWorkspaceRoleBindingsResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	exists, err := s.checkWorkspaceExists(ctx, req)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("workspace %s not found", req.Name)
	}

	var list rbacv1.RoleBindingList
	if err := s.listWorkspaceResources(ctx, req, &list); err != nil {
		return nil, fmt.Errorf("failed to list workspace role bindings, error: %w", err)
	}

	var roleBindings []*capiv1_proto.WorkspaceRoleBinding
	for i := range list.Items {
		roleBinding := capiv1_proto.WorkspaceRoleBinding{
			Name:      list.Items[i].Name,
			Namespace: list.Items[i].Namespace,
			Kind:      "RoleBinding",
			Timestamp: list.Items[i].CreationTimestamp.Format(time.RFC3339),
			Role: &capiv1_proto.WorkspaceRoleBindingRoleRef{
				ApiGroup: list.Items[i].RoleRef.APIGroup,
				Kind:     list.Items[i].RoleRef.Kind,
				Name:     list.Items[i].RoleRef.Name,
			},
		}

		for _, subject := range list.Items[i].Subjects {
			roleBinding.Subjects = append(roleBinding.Subjects, &capiv1_proto.WorkspaceRoleBindingSubject{
				ApiGroup:  subject.APIGroup,
				Kind:      subject.Kind,
				Name:      subject.Name,
				Namespace: subject.Namespace,
			})
		}

		obj := list.Items[i]
		obj.SetGroupVersionKind(rbacv1.SchemeGroupVersion.WithKind("RoleBinding"))

		manifest, err := k8sObjectToYaml(&obj)
		if err != nil {
			return nil, err
		}
		roleBinding.Manifest = manifest
		roleBindings = append(roleBindings, &roleBinding)
	}

	return &capiv1_proto.GetWorkspaceRoleBindingsResponse{
		Name:        req.Name,
		ClusterName: req.ClusterName,
		Objects:     roleBindings,
	}, nil
}

func (s *server) GetWorkspaceServiceAccounts(ctx context.Context, req *capiv1_proto.GetWorkspaceRequest) (*capiv1_proto.GetWorkspaceServiceAccountsResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	exists, err := s.checkWorkspaceExists(ctx, req)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("workspace %s not found", req.Name)
	}

	var list v1.ServiceAccountList
	if err := s.listWorkspaceResources(ctx, req, &list); err != nil {
		return nil, fmt.Errorf("failed to list workspace service accounts, error: %w", err)
	}

	var serviceAccounts []*capiv1_proto.WorkspaceServiceAccount
	for i := range list.Items {
		serviceAccount := &capiv1_proto.WorkspaceServiceAccount{
			Name:      list.Items[i].Name,
			Namespace: list.Items[i].Namespace,
			Kind:      "ServiceAccount",
			Timestamp: list.Items[i].CreationTimestamp.Format(time.RFC3339),
		}

		obj := list.Items[i]
		obj.SetGroupVersionKind(rbacv1.SchemeGroupVersion.WithKind(rbacv1.ServiceAccountKind))

		manifest, err := k8sObjectToYaml(&obj)
		if err != nil {
			return nil, err
		}

		serviceAccount.Manifest = manifest
		serviceAccounts = append(serviceAccounts, serviceAccount)
	}

	return &capiv1_proto.GetWorkspaceServiceAccountsResponse{
		Name:        req.Name,
		ClusterName: req.ClusterName,
		Objects:     serviceAccounts,
	}, nil
}

func (s *server) GetWorkspacePolicies(ctx context.Context, req *capiv1_proto.GetWorkspaceRequest) (*capiv1_proto.GetWorkspacePoliciesResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	exists, err := s.checkWorkspaceExists(ctx, req)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("workspace %s not found", req.Name)
	}

	var policies []*capiv1_proto.WorkspacePolicy

	var list pacv2beta2.PolicyList
	if err := s.listWorkspaceResources(ctx, req, &list); err != nil {
		var list pacv2beta1.PolicyList
		if err := s.listWorkspaceResources(ctx, req, &list); err != nil {
			return nil, fmt.Errorf("failed to list workspace policies, error: %w", err)
		}
		for i := range list.Items {
			policies = append(policies, &capiv1_proto.WorkspacePolicy{
				Id:        list.Items[i].GetName(),
				Name:      list.Items[i].Spec.Name,
				Category:  list.Items[i].Spec.Category,
				Severity:  list.Items[i].Spec.Severity,
				Timestamp: list.Items[i].CreationTimestamp.Format(time.RFC3339),
			})
		}
	} else {
		for i := range list.Items {
			policies = append(policies, &capiv1_proto.WorkspacePolicy{
				Id:        list.Items[i].GetName(),
				Name:      list.Items[i].Spec.Name,
				Category:  list.Items[i].Spec.Category,
				Severity:  list.Items[i].Spec.Severity,
				Timestamp: list.Items[i].CreationTimestamp.Format(time.RFC3339),
			})
		}
	}

	return &capiv1_proto.GetWorkspacePoliciesResponse{
		Name:        req.Name,
		ClusterName: req.ClusterName,
		Objects:     policies,
	}, nil
}

func validateRequest(req *capiv1_proto.GetWorkspaceRequest) error {
	if req.ClusterName == "" {
		return errors.New("cluster name is required")
	}
	if req.Name == "" {
		return errors.New("workspace name is required")
	}
	return nil
}

func (s *server) checkWorkspaceExists(ctx context.Context, req *capiv1_proto.GetWorkspaceRequest) (bool, error) {
	var list v1.NamespaceList
	if err := s.listWorkspaceResources(ctx, req, &list); err != nil {
		return false, fmt.Errorf("failed to list workspace namespaces, error: %w", err)
	}
	return len(list.Items) > 0, nil
}

func (s *server) listWorkspaceResources(ctx context.Context, req *capiv1_proto.GetWorkspaceRequest, list client.ObjectList) error {
	clustersClient, err := s.clustersManager.GetImpersonatedClientForCluster(ctx, auth.Principal(ctx), req.ClusterName)
	if err != nil {
		return fmt.Errorf("error getting impersonating client: %w", err)
	}

	if clustersClient == nil {
		return fmt.Errorf("cluster %s not found", req.ClusterName)
	}

	opts := []client.ListOption{
		client.MatchingLabels{tenantLabel: req.Name},
	}
	if err := clustersClient.List(ctx, req.ClusterName, list, opts...); err != nil {
		return err
	}
	return nil
}

func k8sObjectToYaml(obj client.Object) (string, error) {
	var buf bytes.Buffer
	serializer := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)
	if err := serializer.Encode(obj, &buf); err != nil {
		return "", fmt.Errorf("failed to serialize object, error: %w", err)
	}
	return buf.String(), nil
}
