package store

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"

	"gorm.io/gorm"
)

func TestGetObjects(t *testing.T) {
	g := NewGomegaWithT(t)

	store, db := createStore(t)

	obj := models.Object{
		Cluster:   "test-cluster",
		Name:      "someName",
		Namespace: "namespace",
		Kind:      "ValidKind",
	}

	g.Expect(SeedObjects(db, []models.Object{obj})).To(Succeed())

	iter, err := store.GetObjects(context.Background(), nil, nil)
	g.Expect(err).To(BeNil())

	objects, err := iter.All()
	g.Expect(err).To(BeNil())

	g.Expect(len(objects) > 0).To(BeTrue())
	g.Expect(objects[0].Name).To(Equal(obj.Name))

}

func TestDeleteObjects(t *testing.T) {

	tests := []struct {
		name     string
		seed     []models.Object
		toRemove []models.Object
		want     []models.Object
	}{
		{
			name: "remove one object",
			seed: []models.Object{
				{
					Cluster:    "test-cluster",
					Name:       "someName",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
				{
					Cluster:    "test-cluster",
					Name:       "otherName",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
			},
			toRemove: []models.Object{
				{
					Cluster:    "test-cluster",
					Name:       "someName",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
			},
			want: []models.Object{{
				Cluster:    "test-cluster",
				Name:       "otherName",
				Namespace:  "namespace",
				Kind:       "ValidKind",
				APIGroup:   "example.com",
				APIVersion: "v1",
			}},
		},
		{
			name: "remove multiple objects",
			seed: []models.Object{
				{
					Cluster:    "test-cluster",
					Name:       "someName",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
				{
					Cluster:    "test-cluster",
					Name:       "otherName",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
			},
			toRemove: []models.Object{
				{
					Cluster:    "test-cluster",
					Name:       "someName",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
				{
					Cluster:    "test-cluster",
					Name:       "otherName",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
				},
			},
			want: []models.Object{},
		},
	}
	g := NewGomegaWithT(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, db := createStore(t)

			g.Expect(SeedObjects(db, tt.seed)).To(Succeed())

			sqlDB, err := db.DB()
			g.Expect(err).To(BeNil())

			var count1 int64
			g.Expect(sqlDB.QueryRow("SELECT COUNT(*) FROM objects").Scan(&count1)).To(Succeed())
			g.Expect(count1).To(Equal(int64(len(tt.seed))))

			g.Expect(store.DeleteObjects(context.Background(), tt.toRemove)).To(Succeed())

			var count2 int64

			g.Expect(sqlDB.QueryRow("SELECT COUNT(*) FROM objects").Scan(&count2)).To(Succeed())

			g.Expect(count2).To(Equal(int64(len(tt.want))))
		})
	}
}

func TestStoreObjects(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("stores objects", func(t *testing.T) {
		store, db := createStore(t)

		obj := models.Object{
			Cluster:    "test-cluster",
			Name:       "someName",
			Namespace:  "namespace",
			Kind:       "ValidKind",
			APIGroup:   "example.com",
			APIVersion: "v1",
		}

		g.Expect(store.StoreObjects(context.Background(), []models.Object{obj})).To(Succeed())

		sqlDB, err := db.DB()
		g.Expect(err).To(BeNil())

		var storedObj models.Object
		g.Expect(sqlDB.QueryRow("SELECT id FROM objects").Scan(&storedObj.ID)).To(Succeed())

		g.Expect(storedObj.ID).To(Equal(obj.GetID()))
	})
	t.Run("upserts objects", func(t *testing.T) {
		store, db := createStore(t)

		obj := models.Object{
			Cluster:    "test-cluster",
			Name:       "someName",
			Namespace:  "namespace",
			Kind:       "ValidKind",
			APIGroup:   "example.com",
			APIVersion: "v1",
		}

		g.Expect(store.StoreObjects(context.Background(), []models.Object{obj})).To(Succeed())

		sqlDB, err := db.DB()
		g.Expect(err).To(BeNil())

		var count int64
		g.Expect(sqlDB.QueryRow("SELECT COUNT(*) FROM objects").Scan(&count)).To(Succeed())
		g.Expect(count).To(Equal(int64(1)))

		g.Expect(store.StoreObjects(context.Background(), []models.Object{obj})).To(Succeed())

		g.Expect(sqlDB.QueryRow("SELECT COUNT(*) FROM objects").Scan(&count)).To(Succeed())
		g.Expect(count).To(Equal(int64(1)))
	})

}

func TestGetAccessRules(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx := context.Background()

	store, _ := createStore(t)

	role := models.Role{
		Cluster:   "test-cluster",
		Namespace: "namespace",
		Name:      "someName",
		Kind:      "Role",
		PolicyRules: []models.PolicyRule{
			{
				APIGroups: strings.Join([]string{"example.com"}, ","),
				Resources: strings.Join([]string{"SomeKind"}, ","),
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
				Kind:     role.Kind,
				Name:     role.Name,
			},
		},
		RoleRefName: role.Name,
		RoleRefKind: role.Kind,
	}

	g.Expect(store.StoreRoles(ctx, []models.Role{role})).To(Succeed())
	g.Expect(store.StoreRoleBindings(ctx, []models.RoleBinding{rb})).To(Succeed())

	r, err := store.GetAccessRules(ctx)
	g.Expect(err).To(BeNil())

	g.Expect(r).To(HaveLen(1))

	expected := models.AccessRule{
		Cluster:   "test-cluster",
		Namespace: "namespace",
		Subjects: []models.Subject{{
			APIGroup:      "rbac.authorization.k8s.io",
			Kind:          "Role",
			Name:          "someName",
			RoleBindingID: rb.GetID(),
		}},
		AccessibleKinds:   []string{"example.com/SomeKind"},
		ProvidedByRole:    fmt.Sprintf("%s/%s", role.Kind, role.Name),
		ProvidedByBinding: fmt.Sprintf("%s/%s", rb.Kind, rb.Name),
	}

	diff := cmp.Diff(expected, r[0], cmpopts.IgnoreFields(models.Subject{}, "ID", "CreatedAt", "UpdatedAt"))

	if diff != "" {
		t.Errorf("GetAccessRules() mismatch (-want +got):\n%s", diff)
	}
}

func createStore(t *testing.T) (Store, *gorm.DB) {
	g := NewGomegaWithT(t)
	dbDir, err := os.MkdirTemp("", "db")
	g.Expect(err).To(BeNil())

	db, err := CreateSQLiteDB(dbDir)
	g.Expect(err).To(BeNil())

	store, err := NewSQLiteStore(db, logr.Discard())
	g.Expect(err).To(BeNil())

	return store, db
}
