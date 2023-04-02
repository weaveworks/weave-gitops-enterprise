package store

import (
	"context"
	"fmt"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/utilstest"
	"os"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/accesschecker"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
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

func TestUpsertRoleWithPolicyRules(t *testing.T) {
	// This is a sanity check test to prove that policy rules get upserted along with their roles
	g := NewGomegaWithT(t)
	ctx := context.Background()

	store, db := createStore(t)
	resourcesMap, err := utilstest.CreateAllowedResourcesMapForApplications()
	g.Expect(err).To(BeNil())
	check, err := accesschecker.NewAccessChecker(resourcesMap)
	g.Expect(err).To(BeNil())

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

	rules1, err := store.GetAccessRules(ctx)
	g.Expect(err).To(BeNil())

	g.Expect(check.HasAccess(user, obj, rules1)).To(BeTrue())

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

	rules2, err := store.GetAccessRules(ctx)
	g.Expect(err).To(BeNil())

	g.Expect(check.HasAccess(user, obj, rules2)).To(BeFalse())

}
