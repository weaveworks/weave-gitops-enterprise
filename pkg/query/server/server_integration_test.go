//go:build integration
// +build integration

package server_test

import (
	"context"
	"fmt"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	api "github.com/weaveworks/weave-gitops-enterprise/pkg/api/query"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops-enterprise/test"
	l "github.com/weaveworks/weave-gitops/core/logger"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"go.uber.org/zap/zapcore"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"testing"
	"time"
)

var (
	roleTypeMeta        = typeMeta("Role", "rbac.authorization.k8s.io/v1")
	roleBindingTypeMeta = typeMeta("RoleBinding", "rbac.authorization.k8s.io/v1")
)

const (
	defaultTimeout  = time.Second * 5
	defaultInterval = time.Second
)

// TestQueryServer_IntegrationTest is an integration test for excercising the integration of the
// query system that includes both collecting from a cluster (using env teset) and doing queries via grpc.
// It is also used in the context of logging events per https://github.com/weaveworks/weave-gitops-enterprise/issues/2691
func TestQueryServer_IntegrationTest(t *testing.T) {
	g := NewGomegaWithT(t)
	g.SetDefaultEventuallyTimeout(defaultTimeout)
	g.SetDefaultEventuallyPollingInterval(defaultInterval)

	principal := auth.NewUserPrincipal(auth.ID("user1"), auth.Groups([]string{"group-a"}))
	defaultNamespace := "default"

	testLog := testr.New(t)
	tests := []struct {
		name               string
		logLevel           string
		objects            []client.Object
		access             []client.Object
		principal          *auth.UserPrincipal
		expectedEvents     []string
		expectedNumObjects int
		nonExpectedEvents  []string
	}{
		{
			name:               "should provide query infra events as baseline",
			objects:            []client.Object{},
			access:             []client.Object{},
			logLevel:           "info",
			principal:          principal,
			nonExpectedEvents:  []string{},
			expectedNumObjects: 0,
			expectedEvents: []string{
				// collector infra events
				"objects collector started",
				"role collector started",
				"collectors started", //potential duplicate
				"watcher started",    //potential duplicate
				"watching cluster",
				// query infra events
				"access checker created",
				"query service created",
				"query server created",
				//TODO add stop events
			},
		},
		{
			name:   "should be verbose with debug logging level",
			access: allowHelmReleaseAnyOnDefaultNamespace(principal.ID),
			objects: []client.Object{
				podinfoHelmRepository(defaultNamespace),
				podinfoHelmRelease(defaultNamespace),
				podinfoHelmRelease("kube-public"),
			},
			logLevel:           "debug",
			principal:          principal,
			nonExpectedEvents:  []string{},
			expectedNumObjects: 1, // should allow only on default namespace
			expectedEvents: []string{
				//object collection events
				"object transaction received",
				"storing object",
				"objects stored",
				"objects processed",
				"rolebinding stored",
				"role stored",
				"roles processed",
				//query execution events
				"query received",
				"objects retrieved",
				"query processed",
				//unauthorised access events
				"unauthorised access",
			},
		},
		{
			name:   "should be silent with info logging level",
			access: allowHelmReleaseAnyOnDefaultNamespace(principal.ID),
			objects: []client.Object{
				podinfoHelmRepository(defaultNamespace),
				podinfoHelmRelease(defaultNamespace),
				podinfoHelmRelease("kube-public"),
			},
			logLevel:           "info",
			principal:          principal,
			expectedEvents:     []string{},
			expectedNumObjects: 1, // should allow only on default namespace
			nonExpectedEvents: []string{
				//object collection events
				"object transaction received",
				"storing object",
				"objects stored",
				"objects processed",
				"rolebinding stored",
				"role stored",
				"roles processed",
				//query execution events
				"query received",
				"objects retrieved",
				"query processed",
				//unauthorised access events
				"unauthorised access",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//Given a query environment
			ctx := context.Background()
			appLog, loggerPath := newLoggerWithLevel(t, tt.logLevel)
			c, err := makeQueryServer(t, cfg, tt.principal, appLog, testLog)
			g.Expect(err).To(BeNil())
			//And some access rules and objects ingested
			test.Create(ctx, t, cfg, tt.objects...)
			test.Create(ctx, t, cfg, tt.access...)

			//When query with expected results is successfully executed
			querySucceeded := g.Eventually(func() bool {
				clause := api.QueryClause{
					Key:     "namespace",
					Value:   defaultNamespace,
					Operand: string(store.OperandEqual),
				}
				query, err := c.DoQuery(ctx, &api.QueryRequest{
					Query: []*api.QueryClause{&clause},
				})
				g.Expect(err).To(BeNil())
				return len(query.Objects) == tt.expectedNumObjects
			}).Should(BeTrue())
			//Then query is successfully executed
			g.Expect(querySucceeded).To(BeTrue())

			//When query without expected results is successfully executed
			querySucceeded = g.Eventually(func() bool {
				clause := api.QueryClause{
					Key:     "namespace",
					Value:   "kube-public",
					Operand: string(store.OperandEqual),
				}
				query, err := c.DoQuery(ctx, &api.QueryRequest{
					Query: []*api.QueryClause{&clause},
				})
				g.Expect(err).To(BeNil())
				return len(query.Objects) == 0
			}).Should(BeTrue())
			//Then query is successfully executed
			g.Expect(querySucceeded).To(BeTrue())

			//And logging events are found
			g.Expect(assertLogs(loggerPath, tt.expectedEvents, tt.nonExpectedEvents)).To(Succeed())

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

func allowHelmReleaseAnyOnDefaultNamespace(username string) []client.Object {
	roleName := "helm-release-admin"
	roleBindingName := "wego-admin-helm-release-admin"

	return []client.Object{
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
			}),
	}
}

func assertLogs(loggerPath string, expectedEvents []string, nonExpectedEvents []string) error {
	//get logs
	logs, err := os.ReadFile(loggerPath)
	if err != nil {
		return fmt.Errorf("cannot read logger file: %w", err)

	}
	logss := string(logs)
	logLines := strings.Split(logss, "\n")
	for _, event := range expectedEvents {
		found := false
		for _, logLine := range logLines {
			if strings.Contains(logLine, event) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("event not found: %s", event)
		}
	}

	for _, nonExpectedEvent := range nonExpectedEvents {
		if strings.Contains(logss, nonExpectedEvent) {
			return fmt.Errorf("non expected event found:%s", nonExpectedEvent)
		}
	}

	return nil
}

func newLoggerWithLevel(t *testing.T, logLevel string) (logr.Logger, string) {
	g := NewGomegaWithT(t)

	file, err := os.CreateTemp(os.TempDir(), "query-server-log")
	g.Expect(err).ShouldNot(HaveOccurred())

	name := file.Name()
	g.Expect(err).ShouldNot(HaveOccurred())

	level, err := zapcore.ParseLevel(logLevel)
	cfg := l.BuildConfig(
		l.WithLogLevel(level),
		l.WithMode(false),
		l.WithOutAndErrPaths("stdout", "stderr"),
		l.WithOutAndErrPaths(name, name),
	)

	log, err := l.NewFromConfig(cfg)
	g.Expect(err).NotTo(HaveOccurred())

	t.Cleanup(func() {
		err := os.Remove(file.Name())
		if err != nil {
			t.Fatal(err)
		}
	})

	return log, name
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

func typeMeta(kind, apiVersion string) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       kind,
		APIVersion: apiVersion,
	}
}
