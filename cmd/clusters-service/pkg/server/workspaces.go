package server

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	v1 "k8s.io/api/core/v1"
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
