package server

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pacv2beta2 "github.com/weaveworks/policy-agent/api/v2beta2"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
)

func TestListWorkspaces(t *testing.T) {
	clusters := []struct {
		name  string
		state []runtime.Object
	}{
		{
			name: "management",
			state: []runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-a-1",
						Labels: map[string]string{
							tenantLabel: "tenant-a",
						},
					},
				},
			},
		},
		{
			name: "leaf-1",
			state: []runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-x-1",
						Labels: map[string]string{
							tenantLabel: "tenant-x",
						},
					},
				},
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-x-2",
						Labels: map[string]string{
							tenantLabel: "tenant-x",
						},
					},
				},
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-y-1",
						Labels: map[string]string{
							tenantLabel: "tenant-y",
						},
					},
				},
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-z",
					},
				},
			},
		},
	}

	tests := []struct {
		request  *capiv1_proto.ListWorkspacesRequest
		response *capiv1_proto.ListWorkspacesResponse
	}{
		{
			request: &capiv1_proto.ListWorkspacesRequest{},
			response: &capiv1_proto.ListWorkspacesResponse{
				Workspaces: []*capiv1_proto.Workspace{
					{
						Name:        "tenant-a",
						ClusterName: "management",
						Namespaces:  []string{"namespace-a-1"},
					},
					{
						Name:        "tenant-x",
						ClusterName: "leaf-1",
						Namespaces:  []string{"namespace-x-1", "namespace-x-2"},
					},
					{
						Name:        "tenant-y",
						ClusterName: "leaf-1",
						Namespaces:  []string{"namespace-y-1"},
					},
				},
			},
		},
	}

	clustersClients := map[string]client.Client{}
	for _, cluster := range clusters {
		clustersClients[cluster.name] = createClient(t, cluster.state...)
	}

	ctx := context.Background()
	s := getServer(t, clustersClients, nil)

	for _, tt := range tests {
		res, err := s.ListWorkspaces(ctx, tt.request)
		if err != nil {
			t.Error(err)
		}

		assert.Equal(t, len(tt.response.Workspaces), len(res.Workspaces), "workspaces count is not correct")

		expectedMap := map[string]*capiv1_proto.Workspace{}
		for i := range tt.response.Workspaces {
			expectedMap[tt.response.Workspaces[i].Name] = tt.response.Workspaces[i]
		}
		for i := range res.Workspaces {
			actual := res.Workspaces[i]
			expected, ok := expectedMap[actual.Name]
			if !ok {
				t.Errorf("found unexpected workspace %s", actual.Name)
			}
			assert.Equal(t, expected.Name, actual.Name, "name is not correct")
			assert.Equal(t, expected.ClusterName, actual.ClusterName, "cluster name is not correct")
			assert.Equal(t, expected.Namespaces, actual.Namespaces, "namespaces are not correct")
		}
	}
}

func TestGetWorkspace(t *testing.T) {
	clusters := []struct {
		name  string
		state []runtime.Object
	}{
		{
			name: "management",
			state: []runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-a-1",
						Labels: map[string]string{
							tenantLabel: "tenant-a",
						},
					},
				},
			},
		},
		{
			name: "leaf-1",
			state: []runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-x-1",
						Labels: map[string]string{
							tenantLabel: "tenant-x",
						},
					},
				},
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-x-2",
						Labels: map[string]string{
							tenantLabel: "tenant-x",
						},
					},
				},
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-y-1",
						Labels: map[string]string{
							tenantLabel: "tenant-y",
						},
					},
				},
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "namespace-z",
					},
				},
			},
		},
	}

	tests := []struct {
		request  *capiv1_proto.GetWorkspaceRequest
		response *capiv1_proto.GetWorkspaceResponse
		err      bool
	}{
		{
			request: &capiv1_proto.GetWorkspaceRequest{
				Name:        "tenant-a",
				ClusterName: "management",
			},
			response: &capiv1_proto.GetWorkspaceResponse{
				Name:        "tenant-a",
				ClusterName: "management",
				Namespaces:  []string{"namespace-a-1"},
			},
		},
		{
			request: &capiv1_proto.GetWorkspaceRequest{
				Name:        "tenant-x",
				ClusterName: "leaf-1",
			},
			response: &capiv1_proto.GetWorkspaceResponse{
				Name:        "tenant-x",
				ClusterName: "leaf-1",
				Namespaces:  []string{"namespace-x-1", "namespace-x-2"},
			},
		},
		{
			request: &capiv1_proto.GetWorkspaceRequest{
				Name:        "tenant-y",
				ClusterName: "leaf-1",
			},
			response: &capiv1_proto.GetWorkspaceResponse{
				Name:        "tenant-y",
				ClusterName: "leaf-1",
				Namespaces:  []string{"namespace-y-1"},
			},
		},
		{
			request: &capiv1_proto.GetWorkspaceRequest{
				Name:        uuid.NewString(),
				ClusterName: "leaf-1",
			},
			err: true,
		},
		{
			request: &capiv1_proto.GetWorkspaceRequest{
				Name:        uuid.NewString(),
				ClusterName: uuid.NewString(),
			},
			err: true,
		},
	}

	clustersClients := map[string]client.Client{}
	for _, cluster := range clusters {
		clustersClients[cluster.name] = createClient(t, cluster.state...)
	}
	s := getServer(t, clustersClients, nil)

	for _, tt := range tests {
		res, err := s.GetWorkspace(context.Background(), tt.request)
		if err != nil {
			if tt.err {
				continue
			}
			t.Fatalf("got unexpected error when getting workspace, error: %v", err)
		}
		assert.Equal(t, tt.response.Name, res.Name, "name is not correct")
		assert.Equal(t, tt.response.ClusterName, res.ClusterName, "cluster name is not correct")
		assert.Equal(t, tt.response.Namespaces, res.Namespaces, "namespaces are not correct")
	}
}

func TestListWorkspaceRoles(t *testing.T) {
	clusters := []struct {
		name  string
		state []runtime.Object
	}{
		{
			name: "management",
			state: []runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "a",
						Labels: map[string]string{
							tenantLabel: "tenant-x",
						},
					},
				},
				&rbacv1.Role{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant-x-role",
						Namespace: "a",
						Labels: map[string]string{
							tenantLabel: "tenant-x",
						},
					},
				},
			},
		},
		{
			name: "leaf-1",
			state: []runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "b",
						Labels: map[string]string{
							tenantLabel: "tenant-y",
						},
					},
				},
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "c",
						Labels: map[string]string{
							tenantLabel: "tenant-y",
						},
					},
				},
				&rbacv1.Role{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant-y-role",
						Namespace: "b",
						Labels: map[string]string{
							tenantLabel: "tenant-y",
						},
					},
				},
				&rbacv1.Role{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant-y-role",
						Namespace: "c",
						Labels: map[string]string{
							tenantLabel: "tenant-y",
						},
					},
				},
				&rbacv1.Role{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "d",
					},
				},
			},
		},
	}

	tests := []struct {
		request  *capiv1_proto.GetWorkspaceRequest
		response *capiv1_proto.GetWorkspaceRolesResponse
		err      bool
	}{
		{
			request: &capiv1_proto.GetWorkspaceRequest{
				Name:        "tenant-x",
				ClusterName: "management",
			},
			response: &capiv1_proto.GetWorkspaceRolesResponse{
				Name:        "tenant-x",
				ClusterName: "management",
				Objects: []*capiv1_proto.WorkspaceRole{
					{
						Name:      "tenant-x-role",
						Namespace: "a",
					},
				},
			},
		},
		{
			request: &capiv1_proto.GetWorkspaceRequest{
				Name:        "tenant-y",
				ClusterName: "leaf-1",
			},
			response: &capiv1_proto.GetWorkspaceRolesResponse{
				Name:        "tenant-y",
				ClusterName: "leaf-1",
				Objects: []*capiv1_proto.WorkspaceRole{
					{
						Name:      "tenant-y-role",
						Namespace: "b",
						Kind:      "Role",
					},
					{
						Name:      "tenant-y-role",
						Namespace: "c",
						Kind:      "Role",
					},
				},
			},
		},
		{
			request: &capiv1_proto.GetWorkspaceRequest{
				Name:        "tenant-x",
				ClusterName: uuid.NewString(),
			},
			err: true,
		},
		{
			request: &capiv1_proto.GetWorkspaceRequest{
				Name:        uuid.NewString(),
				ClusterName: "managament",
			},
			err: true,
		},
	}

	clustersClients := map[string]client.Client{}
	for _, cluster := range clusters {
		clustersClients[cluster.name] = createClient(t, cluster.state...)
	}
	s := getServer(t, clustersClients, nil)

	for _, tt := range tests {
		res, err := s.GetWorkspaceRoles(context.Background(), tt.request)
		if err != nil {
			if tt.err {
				continue
			}
			t.Fatalf("got unexpected error when getting workspace, error: %v", err)
		}
		assert.Equal(t, tt.response.Name, res.Name, "name is not correct")
		assert.Equal(t, tt.response.ClusterName, res.ClusterName, "cluster name is not correct")
		assert.Equal(t, len(tt.response.Objects), len(res.Objects), "object count is not correct")
	}
}

func TestListtWorkspaceRoleBindings(t *testing.T) {
	clusters := []struct {
		name  string
		state []runtime.Object
	}{
		{
			name: "management",
			state: []runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "a",
						Labels: map[string]string{
							tenantLabel: "tenant-x",
						},
					},
				},
				&rbacv1.RoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant-x-rolebinding-a",
						Namespace: "a",
						Labels: map[string]string{
							tenantLabel: "tenant-x",
						},
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.SchemeGroupVersion.Identifier(),
						Kind:     "Role",
						Name:     "role-x",
					},
					Subjects: []rbacv1.Subject{
						{
							APIGroup:  rbacv1.SchemeGroupVersion.Identifier(),
							Kind:      "ServiceAccount",
							Name:      "tenant-x-service-account",
							Namespace: "a",
						},
					},
				},
			},
		},
		{
			name: "leaf-1",
			state: []runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "b",
						Labels: map[string]string{
							tenantLabel: "tenant-y",
						},
					},
				},
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "c",
						Labels: map[string]string{
							tenantLabel: "tenant-y",
						},
					},
				},
				&rbacv1.RoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant-y-rolebinding-b",
						Namespace: "b",
						Labels: map[string]string{
							tenantLabel: "tenant-y",
						},
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.SchemeGroupVersion.Identifier(),
						Kind:     "Role",
						Name:     "role-y",
					},
					Subjects: []rbacv1.Subject{
						{
							APIGroup:  rbacv1.SchemeGroupVersion.Identifier(),
							Kind:      "ServiceAccount",
							Name:      "tenant-y-service-account",
							Namespace: "b",
						},
					},
				},
				&rbacv1.RoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant-y-rolebinding-c",
						Namespace: "c",
						Labels: map[string]string{
							tenantLabel: "tenant-y",
						},
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.SchemeGroupVersion.Identifier(),
						Kind:     "Role",
						Name:     "role-y",
					},
					Subjects: []rbacv1.Subject{
						{
							APIGroup:  rbacv1.SchemeGroupVersion.Identifier(),
							Kind:      "ServiceAccount",
							Name:      "tenant-y-service-account-",
							Namespace: "c",
						},
					},
				},
			},
		},
	}

	tests := []struct {
		request  *capiv1_proto.GetWorkspaceRequest
		response *capiv1_proto.GetWorkspaceRoleBindingsResponse
		err      bool
	}{
		{
			request: &capiv1_proto.GetWorkspaceRequest{
				Name:        "tenant-x",
				ClusterName: "management",
			},
			response: &capiv1_proto.GetWorkspaceRoleBindingsResponse{
				Name:        "tenant-x",
				ClusterName: "management",
				Objects: []*capiv1_proto.WorkspaceRoleBinding{
					{
						Name:      "tenant-x-rolebinding-a",
						Namespace: "a",
						Kind:      "RoleBinding",
					},
				},
			},
		},
		{
			request: &capiv1_proto.GetWorkspaceRequest{
				Name:        "tenant-y",
				ClusterName: "leaf-1",
			},
			response: &capiv1_proto.GetWorkspaceRoleBindingsResponse{
				Name:        "tenant-y",
				ClusterName: "leaf-1",
				Objects: []*capiv1_proto.WorkspaceRoleBinding{
					{
						Name:      "tenant-y-rolebinding-b",
						Namespace: "b",
						Kind:      "RoleBinding",
					},
					{
						Name:      "tenant-y-rolebinding-c",
						Namespace: "c",
						Kind:      "RoleBinding",
					},
				},
			},
		},
		{
			request: &capiv1_proto.GetWorkspaceRequest{
				Name:        "tenant-x",
				ClusterName: uuid.NewString(),
			},
			err: true,
		},
		{
			request: &capiv1_proto.GetWorkspaceRequest{
				Name:        uuid.NewString(),
				ClusterName: "managament",
			},
			err: true,
		},
	}

	clustersClients := map[string]client.Client{}
	for _, cluster := range clusters {
		clustersClients[cluster.name] = createClient(t, cluster.state...)
	}
	s := getServer(t, clustersClients, nil)

	for _, tt := range tests {
		res, err := s.GetWorkspaceRoleBindings(context.Background(), tt.request)
		if err != nil {
			if tt.err {
				continue
			}
			t.Fatalf("got unexpected error when getting workspace, error: %v", err)
		}
		assert.Equal(t, tt.response.Name, res.Name, "name is not correct")
		assert.Equal(t, tt.response.ClusterName, res.ClusterName, "cluster name is not correct")
		assert.Equal(t, len(tt.response.Objects), len(res.Objects), "object count is not correct")
	}
}

func TestListWorkspaceServiceAccounts(t *testing.T) {
	clusters := []struct {
		name  string
		state []runtime.Object
	}{
		{
			name: "management",
			state: []runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "a",
						Labels: map[string]string{
							tenantLabel: "tenant-x",
						},
					},
				},
				&v1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant-x-service-account-a",
						Namespace: "a",
						Labels: map[string]string{
							tenantLabel: "tenant-x",
						},
					},
				},
			},
		},
		{
			name: "leaf-1",
			state: []runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "b",
						Labels: map[string]string{
							tenantLabel: "tenant-y",
						},
					},
				},
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "c",
						Labels: map[string]string{
							tenantLabel: "tenant-y",
						},
					},
				},
				&v1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant-y-service-account-b",
						Namespace: "b",
						Labels: map[string]string{
							tenantLabel: "tenant-y",
						},
					},
				},
				&v1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tenant-y-service-account-c",
						Namespace: "c",
						Labels: map[string]string{
							tenantLabel: "tenant-y",
						},
					},
				},
				&v1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "d",
					},
				},
			},
		},
	}

	tests := []struct {
		request  *capiv1_proto.GetWorkspaceRequest
		response *capiv1_proto.GetWorkspaceServiceAccountsResponse
		err      bool
	}{
		{
			request: &capiv1_proto.GetWorkspaceRequest{
				Name:        "tenant-x",
				ClusterName: "management",
			},
			response: &capiv1_proto.GetWorkspaceServiceAccountsResponse{
				Name:        "tenant-x",
				ClusterName: "management",
				Objects: []*capiv1_proto.WorkspaceServiceAccount{
					{
						Name:      "tenant-x-service-account-a",
						Namespace: "a",
						Kind:      "ServiceAccount",
					},
				},
			},
		},
		{
			request: &capiv1_proto.GetWorkspaceRequest{
				Name:        "tenant-y",
				ClusterName: "leaf-1",
			},
			response: &capiv1_proto.GetWorkspaceServiceAccountsResponse{
				Name:        "tenant-y",
				ClusterName: "leaf-1",
				Objects: []*capiv1_proto.WorkspaceServiceAccount{
					{
						Name:      "tenant-y-service-account-b",
						Namespace: "b",
						Kind:      "ServiceAccount",
					},
					{
						Name:      "tenant-y-service-account-c",
						Namespace: "c",
						Kind:      "ServiceAccount",
					},
				},
			},
		},
		{
			request: &capiv1_proto.GetWorkspaceRequest{
				Name:        "tenant-x",
				ClusterName: uuid.NewString(),
			},
			err: true,
		},
		{
			request: &capiv1_proto.GetWorkspaceRequest{
				Name:        uuid.NewString(),
				ClusterName: "managament",
			},
			err: true,
		},
	}

	clustersClients := map[string]client.Client{}
	for _, cluster := range clusters {
		clustersClients[cluster.name] = createClient(t, cluster.state...)
	}
	s := getServer(t, clustersClients, nil)

	for _, tt := range tests {
		res, err := s.GetWorkspaceServiceAccounts(context.Background(), tt.request)
		if err != nil {
			if tt.err {
				continue
			}
			t.Fatalf("got unexpected error when getting workspace, error: %v", err)
		}
		assert.Equal(t, tt.response.Name, res.Name, "name is not correct")
		assert.Equal(t, tt.response.ClusterName, res.ClusterName, "cluster name is not correct")
		assert.Equal(t, len(tt.response.Objects), len(res.Objects), "object count is not correct")

		serviceAccountMap := map[string]*capiv1_proto.WorkspaceServiceAccount{}
		for i := range tt.response.Objects {
			serviceAccountMap[tt.response.Objects[i].Name] = tt.response.Objects[i]
		}
		for i := range res.Objects {
			actual := res.Objects[i]
			expected, ok := serviceAccountMap[actual.Name]
			if !ok {
				t.Fatalf("found unexpected workspace %s", actual.Name)
			}
			assert.Equal(t, expected.Name, actual.Name, "name is not correct")
			assert.Equal(t, expected.Namespace, actual.Namespace, "namespace is not correct")
		}
	}
}

func TestListtWorkspacePolicies(t *testing.T) {
	clusters := []struct {
		name  string
		state []runtime.Object
	}{
		{
			name: "management",
			state: []runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "a",
						Labels: map[string]string{
							tenantLabel: "tenant-x",
						},
					},
				},
				&pacv2beta2.Policy{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-x-policy-1",
						Labels: map[string]string{
							tenantLabel: "tenant-x",
						},
					},
					Spec: pacv2beta2.PolicySpec{
						Name:     "Policy-1",
						Category: "category-1",
						Severity: "low",
					},
				},
			},
		},
		{
			name: "leaf-1",
			state: []runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "b",
						Labels: map[string]string{
							tenantLabel: "tenant-y",
						},
					},
				},
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "c",
						Labels: map[string]string{
							tenantLabel: "tenant-y",
						},
					},
				},
				&pacv2beta2.Policy{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-y-policy-1",
						Labels: map[string]string{
							tenantLabel: "tenant-y",
						},
					},
					Spec: pacv2beta2.PolicySpec{
						Name:     "Policy-1",
						Category: "category-2",
						Severity: "high",
					},
				},
				&pacv2beta2.Policy{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tenant-y-policy-2",
						Labels: map[string]string{
							tenantLabel: "tenant-y",
						},
					},
					Spec: pacv2beta2.PolicySpec{
						Name:     "Policy-2",
						Category: "category-3",
						Severity: "medium",
					},
				},
				&pacv2beta2.Policy{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
				},
			},
		},
	}

	tests := []struct {
		request  *capiv1_proto.GetWorkspaceRequest
		response *capiv1_proto.GetWorkspacePoliciesResponse
		err      bool
	}{
		{
			request: &capiv1_proto.GetWorkspaceRequest{
				Name:        "tenant-x",
				ClusterName: "management",
			},
			response: &capiv1_proto.GetWorkspacePoliciesResponse{
				Name:        "tenant-x",
				ClusterName: "management",
				Objects: []*capiv1_proto.WorkspacePolicy{
					{
						Id:       "tenant-x-policy-1",
						Name:     "Policy-1",
						Category: "category-1",
						Severity: "low",
					},
				},
			},
		},
		{
			request: &capiv1_proto.GetWorkspaceRequest{
				Name:        "tenant-y",
				ClusterName: "leaf-1",
			},
			response: &capiv1_proto.GetWorkspacePoliciesResponse{
				Name:        "tenant-y",
				ClusterName: "leaf-1",
				Objects: []*capiv1_proto.WorkspacePolicy{
					{
						Id:       "tenant-y-policy-1",
						Name:     "Policy-1",
						Category: "category-2",
						Severity: "high",
					},
					{
						Id:       "tenant-y-policy-2",
						Name:     "Policy-2",
						Category: "category-3",
						Severity: "medium",
					},
				},
			},
		},
		{
			request: &capiv1_proto.GetWorkspaceRequest{
				Name:        "tenant-x",
				ClusterName: uuid.NewString(),
			},
			err: true,
		},
		{
			request: &capiv1_proto.GetWorkspaceRequest{
				Name:        uuid.NewString(),
				ClusterName: "managament",
			},
			err: true,
		},
	}

	clustersClients := map[string]client.Client{}
	for _, cluster := range clusters {
		clustersClients[cluster.name] = createClient(t, cluster.state...)
	}
	s := getServer(t, clustersClients, nil)

	for _, tt := range tests {
		res, err := s.GetWorkspacePolicies(context.Background(), tt.request)
		if err != nil {
			if tt.err {
				continue
			}
			t.Fatalf("got unexpected error when getting workspace, error: %v", err)
		}
		assert.Equal(t, tt.response.Name, res.Name, "name is not correct")
		assert.Equal(t, tt.response.ClusterName, res.ClusterName, "cluster name is not correct")
		assert.Equal(t, len(tt.response.Objects), len(res.Objects), "object count is not correct")

		policiesMap := map[string]*capiv1_proto.WorkspacePolicy{}
		for i := range tt.response.Objects {
			policiesMap[tt.response.Objects[i].Id] = tt.response.Objects[i]
		}
		for i := range res.Objects {
			actual := res.Objects[i]
			expected, ok := policiesMap[actual.Id]
			if !ok {
				t.Fatalf("found unexpected workspace %s", actual.Id)
			}
			assert.Equal(t, expected.Name, actual.Name, "name is not correct")
			assert.Equal(t, expected.Category, actual.Category, "category is not correct")
			assert.Equal(t, expected.Severity, actual.Severity, "severity are not correct")
		}
	}
}
