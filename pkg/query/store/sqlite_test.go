package store

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/rbac"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/utils/testutils"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"

	"github.com/go-logr/logr/testr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/metrics"
)

func TestNewSQLiteStore(t *testing.T) {
	g := NewGomegaWithT(t)
	dbDir, err := os.MkdirTemp("", "db")
	g.Expect(err).To(BeNil())

	db, err := CreateSQLiteDB(dbDir)
	g.Expect(err).To(BeNil())

	sqlDB, err := db.DB()
	g.Expect(err).To(BeNil())

	tests := []struct {
		name        string
		tableName   string
		desiredCols []string
	}{
		{
			name:        "objects table",
			tableName:   "objects",
			desiredCols: []string{"id", "cluster", "namespace", "kind", "name", "status", "message"},
		},
		{
			name:        "role_bindings table",
			tableName:   "role_bindings",
			desiredCols: []string{"id", "cluster", "namespace", "kind", "name", "role_ref_name", "role_ref_kind"},
		},
		{
			name:        "roles table",
			tableName:   "roles",
			desiredCols: []string{"id", "cluster", "namespace", "kind", "name"},
		},
		{
			name:        "subjects table",
			tableName:   "subjects",
			desiredCols: []string{"id", "namespace", "kind", "name", "role_binding_id"},
		},
		{
			name:        "policy_rules table",
			tableName:   "policy_rules",
			desiredCols: []string{"id", "role_id", "api_groups", "resources", "verbs"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cols, err := sqlDB.Query(fmt.Sprintf("PRAGMA table_info(%s)", tt.tableName))
			g.Expect(err).To(BeNil())

			var columnNames []string
			for cols.Next() {
				var index int64
				var columnName string
				var dataType interface{}
				var nullable bool
				var defaultVal interface{}
				var autoIncrement bool

				err := cols.Scan(&index, &columnName, &dataType, &nullable, &defaultVal, &autoIncrement)
				g.Expect(err).To(BeNil())

				columnNames = append(columnNames, columnName)
			}
			g.Expect(columnNames).To(ContainElements(tt.desiredCols))
		})

	}
}

func TestSQLiteStore_StoreObjects(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()
	store, db := createStore(t)
	sqlDB, err := db.DB()
	g.Expect(err).To(BeNil())
	log := testr.New(t)

	metrics.NewPrometheusServer(metrics.Options{
		ServerAddress: "localhost:8080",
	}, prometheus.Gatherers{
		prometheus.DefaultGatherer,
	})

	tests := []struct {
		name       string
		objects    []models.Object
		errPattern string
	}{
		{
			name:       "should ignore when empty objects",
			objects:    []models.Object{},
			errPattern: "",
		},
		{
			name: "should store with one object",
			objects: []models.Object{
				{
					Cluster:    "test-cluster",
					Name:       "obj-cluster-1",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
					Category:   models.CategoryAutomation,
				},
			},
			errPattern: "",
		},
		{
			name: "should store with more than one object",
			objects: []models.Object{
				{
					Cluster:    "test-cluster",
					Name:       "obj-cluster-1",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
					Category:   models.CategoryAutomation,
				},
				{
					Cluster:    "test-cluster-2",
					Name:       "obj-cluster-2",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
					Category:   models.CategorySource,
				},
			},
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.StoreObjects(ctx, tt.objects)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			var storedObjectsNum int
			g.Expect(sqlDB.QueryRow("SELECT COUNT(id) FROM objects").Scan(&storedObjectsNum)).To(Succeed())
			g.Expect(storedObjectsNum == len(tt.objects)).To(BeTrue())
		})
	}

	// Retrieve the metrics
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/metrics", nil)
	g.Expect(err).NotTo(HaveOccurred())
	resp, err := http.DefaultClient.Do(req)
	g.Expect(err).NotTo(HaveOccurred())
	b, err := io.ReadAll(resp.Body)
	g.Expect(err).NotTo(HaveOccurred())
	metrics := string(b)
	log.Info("metrics: %s", metrics)

	expMetrics := []string{
		`explorer_datastore_inflight_requests_total`,
		`explorer_datastore_latency_seconds_bucket`,
	}

	for _, expMetric := range expMetrics {
		//Contains expected value
		g.Expect(metrics).To(ContainSubstring(expMetric))
	}
}

func TestSQLiteStore_DeleteAllObjects(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()
	store, db := createStore(t)
	sqlDB, err := db.DB()
	g.Expect(err).To(BeNil())

	tests := []struct {
		name           string
		addObjects     []models.Object // objects to add before deleting
		deleteClusters []string
		errPattern     string
	}{
		{
			name:           "should do nothing for empty request",
			addObjects:     []models.Object{},
			deleteClusters: []string{},
			errPattern:     "",
		},
		{
			name: "should do nothing if no objects for cluster to delete",
			addObjects: []models.Object{
				{
					Cluster:    "test-cluster",
					Name:       "obj-cluster-1",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
					Category:   models.CategoryAutomation,
				},
			},
			deleteClusters: []string{"cluster-without-objects"},
			errPattern:     "",
		},
		{
			name: "should have deleted all for a cluster with objects",
			addObjects: []models.Object{
				{
					Cluster:    "cluster-with-objects",
					Name:       "obj-cluster-1",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
					Category:   models.CategoryAutomation,
				},
				{
					Cluster:    "cluster-with-objects",
					Name:       "obj-cluster-2",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
					Category:   models.CategoryAutomation,
				},
			},
			deleteClusters: []string{"cluster-with-objects"},
			errPattern:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g.Expect(store.StoreObjects(ctx, tt.addObjects)).To(Succeed())
			err := store.DeleteAllObjects(ctx, tt.deleteClusters)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			for _, deleteCluster := range tt.deleteClusters {
				var numResources int
				g.Expect(sqlDB.QueryRow("SELECT COUNT(id) FROM objects WHERE cluster = ?", deleteCluster).Scan(&numResources)).To(Succeed())
				g.Expect(numResources).To(Equal(0))
			}
		})
	}
}

func TestSQLiteStore_DeleteAllRoles(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()
	store, db := createStore(t)
	sqlDB, err := db.DB()
	g.Expect(err).To(BeNil())

	tests := []struct {
		name           string
		rolesToAdd     []models.Role // objects to add before deleting
		deleteClusters []string
		errPattern     string
	}{
		{
			name:           "should do nothing for empty request",
			rolesToAdd:     []models.Role{},
			deleteClusters: []string{},
			errPattern:     "",
		},
		{
			name: "should do nothing if no objects for cluster to delete",
			rolesToAdd: []models.Role{
				{
					Name:      "wego-cluster-role",
					Cluster:   "flux-system/leaf-cluster-1",
					Namespace: "",
					Kind:      "ClusterRole",
					PolicyRules: []models.PolicyRule{{
						APIGroups: strings.Join([]string{helmv2.GroupVersion.String()}, ","),
						Resources: strings.Join([]string{"helmreleases"}, ","),
						Verbs:     strings.Join([]string{"get", "list", "patch"}, ","),
					}},
				},
			},
			deleteClusters: []string{"cluster-without-objects"},
			errPattern:     "",
		},
		{
			name: "should have deleted all for a cluster with objects",
			rolesToAdd: []models.Role{
				{
					Name:      "wego-cluster-role",
					Cluster:   "flux-system/leaf-cluster-1",
					Namespace: "",
					Kind:      "ClusterRole",
					PolicyRules: []models.PolicyRule{{
						APIGroups: strings.Join([]string{helmv2.GroupVersion.String()}, ","),
						Resources: strings.Join([]string{"helmreleases"}, ","),
						Verbs:     strings.Join([]string{"get", "list", "patch"}, ","),
					}},
				},
				{
					Name:      "wego-cluster-role2",
					Cluster:   "flux-system/leaf-cluster-1",
					Namespace: "",
					Kind:      "ClusterRole",
					PolicyRules: []models.PolicyRule{{
						APIGroups: strings.Join([]string{helmv2.GroupVersion.String()}, ","),
						Resources: strings.Join([]string{"helmreleases"}, ","),
						Verbs:     strings.Join([]string{"get", "list", "patch"}, ","),
					}},
				},
			},
			deleteClusters: []string{"flux-system/leaf-cluster-1"},
			errPattern:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g.Expect(store.StoreRoles(ctx, tt.rolesToAdd)).To(Succeed())
			err := store.DeleteAllRoles(ctx, tt.deleteClusters)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			for _, deleteCluster := range tt.deleteClusters {
				var numResources int
				g.Expect(sqlDB.QueryRow("SELECT COUNT(id) FROM roles WHERE cluster = ?", deleteCluster).Scan(&numResources)).To(Succeed())
				g.Expect(numResources).To(Equal(0))
			}
		})
	}
}

func TestSQLiteStore_DeleteAllRoleBindings(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()
	store, db := createStore(t)
	sqlDB, err := db.DB()
	g.Expect(err).To(BeNil())

	tests := []struct {
		name              string
		roleBindingsToAdd []models.RoleBinding // objects to add before deleting
		deleteClusters    []string
		errPattern        string
	}{
		{
			name:              "should do nothing for empty request",
			roleBindingsToAdd: []models.RoleBinding{},
			deleteClusters:    []string{},
			errPattern:        "",
		},
		{
			name: "should do nothing if no objects for cluster to delete",
			roleBindingsToAdd: []models.RoleBinding{
				{
					Cluster:   "cluster-a",
					Name:      "binding-a",
					Namespace: "ns-a",
					Kind:      "RoleBinding",
					Subjects: []models.Subject{{
						Kind: "Group",
						Name: "group-a",
					}},
					RoleRefName: "role-a",
					RoleRefKind: "Role",
				},
			},
			deleteClusters: []string{"cluster-without-objects"},
			errPattern:     "",
		},
		{
			name: "should have deleted all for a cluster with objects",
			roleBindingsToAdd: []models.RoleBinding{
				{
					Cluster:   "cluster-a",
					Name:      "binding-a",
					Namespace: "ns-a",
					Kind:      "RoleBinding",
					Subjects: []models.Subject{{
						Kind: "Group",
						Name: "group-a",
					}},
					RoleRefName: "role-a",
					RoleRefKind: "Role",
				},
				{
					Cluster:   "cluster-a",
					Name:      "binding-b",
					Namespace: "ns-a",
					Kind:      "RoleBinding",
					Subjects: []models.Subject{{
						Kind: "Group",
						Name: "group-a",
					}},
					RoleRefName: "role-a",
					RoleRefKind: "Role",
				},
			},
			deleteClusters: []string{"cluster-a"},
			errPattern:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g.Expect(store.StoreRoleBindings(ctx, tt.roleBindingsToAdd)).To(Succeed())
			err := store.DeleteAllRoleBindings(ctx, tt.deleteClusters)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			for _, deleteCluster := range tt.deleteClusters {
				var numResources int
				g.Expect(sqlDB.QueryRow("SELECT COUNT(id) FROM role_bindings WHERE cluster = ?", deleteCluster).Scan(&numResources)).To(Succeed())
				g.Expect(numResources).To(Equal(0))
			}
		})
	}
}

func TestUpsertRoleWithPolicyRules(t *testing.T) {
	// This is a sanity check test to prove that policy rules get upserted along with their roles
	g := NewGomegaWithT(t)
	ctx := context.Background()

	store, db := createStore(t)
	resourcesMap, err := testutils.CreateDefaultResourceKindMap()
	g.Expect(err).NotTo(HaveOccurred())

	role := models.Role{
		Cluster:   "test-cluster",
		Namespace: "namespace",
		Name:      "someName",
		Kind:      "Role",
		PolicyRules: []models.PolicyRule{
			{
				APIGroups: strings.Join([]string{"example.com"}, ","),
				Resources: strings.Join([]string{"helmreleases"}, ","),
				Verbs:     strings.Join([]string{"get", "list"}, ","),
			},
		},
	}

	rb := models.RoleBinding{
		Cluster:   "test-cluster",
		Namespace: "namespace",
		Name:      "someName",
		Kind:      "RoleBinding",
		Subjects: []models.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "User",
				Name:     role.Name,
			},
		},
		RoleRefName: role.Name,
		RoleRefKind: role.Kind,
	}

	user := &auth.UserPrincipal{ID: role.Name, Groups: []string{}}

	obj := models.Object{
		Cluster:    "test-cluster",
		Namespace:  "namespace",
		Kind:       "HelmRelease",
		APIGroup:   "example.com",
		APIVersion: "",
	}

	g.Expect(store.StoreRoles(ctx, []models.Role{role})).To(Succeed())
	g.Expect(store.StoreRoleBindings(ctx, []models.RoleBinding{rb})).To(Succeed())

	sqlDB, err := db.DB()
	g.Expect(err).To(BeNil())

	var roleID string

	g.Expect(sqlDB.QueryRow("SELECT id FROM roles").Scan(&roleID)).To(Succeed())

	var id int64
	var resources1 string
	g.Expect(sqlDB.QueryRow("SELECT id, resources FROM policy_rules WHERE role_id = ?", roleID).Scan(&id, &resources1)).To(Succeed())

	g.Expect(resources1).To(Equal("helmreleases"))

	roles1, err := store.GetRoles(ctx)
	g.Expect(err).NotTo(HaveOccurred())
	rolebindings1, err := store.GetRoleBindings(ctx)
	g.Expect(err).NotTo(HaveOccurred())

	authz := rbac.NewAuthorizer(resourcesMap)
	allow := authz.ObjectAuthorizer(roles1, rolebindings1, user, obj.Cluster)
	g.Expect(allow(obj)).To(BeTrue())

	// Update the role with a new policy rule
	role.PolicyRules = []models.PolicyRule{
		{
			APIGroups: models.JoinRuleData([]string{"example.com"}),
			// Removing a kind here
			Resources: models.JoinRuleData([]string{}),
			Verbs:     models.JoinRuleData([]string{"get", "list"}),
		},
	}

	g.Expect(store.StoreRoles(ctx, []models.Role{role})).To(Succeed())

	var count1 int64
	g.Expect(sqlDB.QueryRow("SELECT COUNT(*) FROM policy_rules WHERE role_id = ?", roleID).Scan(&count1)).To(Succeed())
	g.Expect(count1).To(Equal(int64(1)))

	var resources2 string

	g.Expect(sqlDB.QueryRow("SELECT resources FROM policy_rules WHERE role_id = ?", roleID).Scan(&resources2)).To(Succeed())

	g.Expect(resources2).To(Equal(""))

	roles2, err := store.GetRoles(ctx)
	g.Expect(err).NotTo(HaveOccurred())
	rolebindings2, err := store.GetRoleBindings(ctx)
	g.Expect(err).NotTo(HaveOccurred())

	allow = authz.ObjectAuthorizer(roles2, rolebindings2, user, obj.Cluster)
	g.Expect(allow(obj)).To(BeFalse())
}
