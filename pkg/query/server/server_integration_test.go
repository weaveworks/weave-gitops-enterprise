//go:build integration
// +build integration

package server_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	api "github.com/weaveworks/weave-gitops-enterprise/pkg/api/query"
	"github.com/weaveworks/weave-gitops-enterprise/test"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	namespaceTypeMeta          = typeMeta("Namespace", "v1")
	serviceAccountTypeMeta     = typeMeta("ServiceAccount", "v1")
	roleTypeMeta               = typeMeta("Role", "rbac.authorization.k8s.io/v1")
	roleBindingTypeMeta        = typeMeta("RoleBinding", "rbac.authorization.k8s.io/v1")
	clusterRoleTypeMeta        = typeMeta("ClusterRole", "rbac.authorization.k8s.io/v1")
	clusterRoleBindingTypeMeta = typeMeta("ClusterRoleBinding", "rbac.authorization.k8s.io/v1")
)

const (
	defaultTimeout  = time.Second * 5
	defaultInterval = time.Second
)

// TestQueryServer is an integration test for exercising the integration of the
// query system that includes both collecting from a cluster (using envtest) and doing queries via grpc.
// It is also used in the context of logging events per https://github.com/weaveworks/weave-gitops-enterprise/issues/2691
func TestQueryServer(t *testing.T) {
	g := NewGomegaWithT(t)
	g.SetDefaultEventuallyTimeout(defaultTimeout)
	g.SetDefaultEventuallyPollingInterval(defaultInterval)

	principal := auth.NewUserPrincipal(auth.ID("user1"), auth.Groups([]string{"group-a"}))
	defaultNamespace := "default"

	test.Create(context.Background(), t, cfg, newNamespace("flux-system"))

	testLog := testr.New(t)

	//Given a query environment
	ctx := context.Background()
	c, err := makeQueryServer(t, cfg, principal, testLog)
	g.Expect(err).To(BeNil())

	tests := []struct {
		name               string
		objects            []client.Object
		access             []client.Object
		query              string
		expectedNumObjects int
	}{
		{
			name:   "should support apps (using helm releases)",
			access: allowHelmReleaseAnyOnDefaultNamespace(principal.ID),
			objects: []client.Object{
				podinfoHelmRepository(defaultNamespace),
				podinfoHelmRelease(defaultNamespace),
			},
			query:              "kind:HelmRelease",
			expectedNumObjects: 1, // should allow only on default namespace
		},
		{
			name:   "should support helm repository",
			access: allowSourcesAnyOnDefaultNamespace(principal.ID),
			objects: []client.Object{
				podinfoHelmRepository(defaultNamespace),
			},
			query:              "kind:HelmRepository",
			expectedNumObjects: 1, // should allow only on default namespace,
		},
		{
			name:   "should support helm chart",
			access: allowSourcesAnyOnDefaultNamespace(principal.ID),
			objects: []client.Object{
				&sourcev1.HelmChart{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "podinfo",
						Namespace: defaultNamespace,
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       sourcev1.HelmChartKind,
						APIVersion: sourcev1.GroupVersion.String(),
					},
					Spec: sourcev1.HelmChartSpec{
						Chart:   "podinfo",
						Version: "v0.0.1",
						SourceRef: sourcev1.LocalHelmChartSourceReference{
							Kind: sourcev1.HelmRepositoryKind,
							Name: "podinfo",
						},
					},
				},
			},
			query:              "kind:HelmChart",
			expectedNumObjects: 1, // should allow only on default namespace,
		},
		{
			name:   "should support git repository chart",
			access: allowSourcesAnyOnDefaultNamespace(principal.ID),
			objects: []client.Object{
				&sourcev1.GitRepository{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "podinfo",
						Namespace: defaultNamespace,
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       sourcev1.GitRepositoryKind,
						APIVersion: sourcev1.GroupVersion.String(),
					},
					Spec: sourcev1.GitRepositorySpec{
						URL: "https://example.com/owner/repo",
					},
				},
			},
			query:              "kind:GitRepository",
			expectedNumObjects: 1, // should allow only on default namespace,
		},
		{
			name:   "should support oci repository",
			access: allowSourcesAnyOnDefaultNamespace(principal.ID),
			objects: []client.Object{
				&sourcev1.OCIRepository{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "podinfo",
						Namespace: defaultNamespace,
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       sourcev1.OCIRepositoryKind,
						APIVersion: sourcev1.GroupVersion.String(),
					},
					Spec: sourcev1.OCIRepositorySpec{
						URL: "oci://example.com/owner/repo",
					},
				},
			},
			query:              fmt.Sprintf("kind:%s", sourcev1.OCIRepositoryKind),
			expectedNumObjects: 1, // should allow only on default namespace,
		},
		{
			name:   "should support bucket",
			access: allowSourcesAnyOnDefaultNamespace(principal.ID),
			objects: []client.Object{
				&sourcev1.Bucket{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "podinfo",
						Namespace: defaultNamespace,
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       sourcev1.BucketKind,
						APIVersion: sourcev1.GroupVersion.String(),
					},
					Spec: sourcev1.BucketSpec{},
				},
			},
			query:              fmt.Sprintf("kind:%s", sourcev1.BucketKind),
			expectedNumObjects: 1, // should allow only on default namespace,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//When some access rules and objects ingested
			test.Create(ctx, t, cfg, tt.objects...)
			test.Create(ctx, t, cfg, tt.access...)

			//When query with expected results is successfully executed
			querySucceeded := g.Eventually(func() bool {
				query, err := c.DoQuery(ctx, &api.QueryRequest{Filters: []string{tt.query}})
				g.Expect(err).To(BeNil())
				return len(query.Objects) == tt.expectedNumObjects
			}).Should(BeTrue())
			//Then query is successfully executed
			g.Expect(querySucceeded).To(BeTrue())

		})
	}
}

func podinfoHelmRelease(defaultNamespace string) *helmv2.HelmRelease {
	return &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "podinfo",
			Namespace: defaultNamespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.String(),
		},
		Spec: helmv2.HelmReleaseSpec{
			Interval: metav1.Duration{Duration: time.Minute},
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: "podinfo",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.HelmRepositoryKind,
						Name:      "podinfo",
						Namespace: defaultNamespace,
					},
				},
			},
		},
	}
}

func podinfoHelmRepository(namespace string) *sourcev1.HelmRepository {
	return &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "podinfo",
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.HelmRepositoryKind,
			APIVersion: sourcev1.GroupVersion.String(),
		},
		Spec: sourcev1.HelmRepositorySpec{
			Interval: metav1.Duration{Duration: time.Minute},
			URL:      "http://my-url.com",
		},
	}
}

func createCollectorSecurityContext() []client.Object {

	return []client.Object{
		newServiceAccount("collector", "flux-system"),
		newClusterRole("collector",
			[]rbacv1.PolicyRule{{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			}}),
		newClusterRoleBinding("collector",
			"ClusterRole",
			"collector",
			[]rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "collector",
					Namespace: "flux-system",
				},
			}),
	}
}

func allowHelmReleaseAnyOnDefaultNamespace(username string) []client.Object {
	roleName := "helm-release-admin"
	roleBindingName := "wego-admin-helm-release-admin"

	return append(createCollectorSecurityContext(),
		newRole(roleName, "default",
			[]rbacv1.PolicyRule{{
				APIGroups: []string{"helm.toolkit.fluxcd.io"},
				Resources: []string{"helmreleases"},
				Verbs:     []string{"*"},
			}}),
		newRoleBinding(roleBindingName,
			"default",
			"Role",
			roleName,
			[]rbacv1.Subject{
				{
					Kind: "User",
					Name: username,
				},
			}))
}

func allowSourcesAnyOnDefaultNamespace(username string) []client.Object {
	roleName := "helm-release-admin"
	roleBindingName := "wego-admin-helm-release-admin"

	return append(createCollectorSecurityContext(),
		newRole(roleName, "default",
			[]rbacv1.PolicyRule{{
				APIGroups: []string{"source.toolkit.fluxcd.io"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			}}),
		newRoleBinding(roleBindingName,
			"default",
			"Role",
			roleName,
			[]rbacv1.Subject{
				{
					Kind: "User",
					Name: username,
				},
			}),
	)
}

func newRoleBinding(name, namespace, roleKind, roleName string, subjects []rbacv1.Subject) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		TypeMeta: roleBindingTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     roleKind,
			Name:     roleName,
		},
		Subjects: subjects,
	}
}

func newClusterRoleBinding(name, roleKind, roleName string, subjects []rbacv1.Subject) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: clusterRoleBindingTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     roleKind,
			Name:     roleName,
		},
		Subjects: subjects,
	}
}

func newRole(name, namespace string, rules []rbacv1.PolicyRule) *rbacv1.Role {
	return &rbacv1.Role{
		TypeMeta: roleTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Rules: rules,
	}
}

func newClusterRole(name string, rules []rbacv1.PolicyRule) *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		TypeMeta: clusterRoleTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Rules: rules,
	}
}

func newNamespace(name string) *v1.Namespace {
	return &v1.Namespace{
		TypeMeta: namespaceTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func newServiceAccount(name, namespace string) *v1.ServiceAccount {
	return &v1.ServiceAccount{
		TypeMeta: serviceAccountTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

func typeMeta(kind, apiVersion string) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       kind,
		APIVersion: apiVersion,
	}
}
