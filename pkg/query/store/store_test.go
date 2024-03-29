package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
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

func TestGetObjectsWithPagination(t *testing.T) {
	g := NewGomegaWithT(t)

	store, db := createStore(t)

	testObjects := []models.Object{
		{
			Cluster:   "test-cluster",
			Name:      "someName-1",
			Namespace: "namespace",
			Kind:      "ValidKind",
		},
		{
			Cluster:   "test-cluster",
			Name:      "someName-2",
			Namespace: "namespace",
			Kind:      "ValidKind",
		},
		{
			Cluster:   "test-cluster",
			Name:      "someName-3",
			Namespace: "namespace",
			Kind:      "ValidKind",
		},
		{
			Cluster:   "test-cluster",
			Name:      "someName-4",
			Namespace: "namespace",
			Kind:      "ValidKind",
		},
		{
			Cluster:   "test-cluster",
			Name:      "someName-5",
			Namespace: "namespace",
			Kind:      "ValidKind",
		},
		{
			Cluster:   "test-cluster",
			Name:      "someName-6",
			Namespace: "namespace",
			Kind:      "ValidKind",
		},
		{
			Cluster:   "test-cluster",
			Name:      "someName-7",
			Namespace: "namespace",
			Kind:      "ValidKind",
		},
	}

	g.Expect(SeedObjects(db, testObjects)).To(Succeed())

	iter, err := store.GetObjects(context.Background(), nil, nil)
	g.Expect(err).To(BeNil())

	objects, err := iter.All()
	g.Expect(err).To(BeNil())
	g.Expect(len(objects)).To(Equal(len(testObjects)))
	g.Expect(objects[0].Name).To(Equal(testObjects[0].Name))
	g.Expect(objects[3].Name).To(Equal(testObjects[3].Name))

	// With pagination and without an offset.
	iter, err = store.GetObjects(context.Background(), nil, nil)
	g.Expect(err).To(BeNil())

	objects, err = iter.Page(3, 0)
	g.Expect(err).To(BeNil())
	g.Expect(len(objects)).To(Equal(3))
	g.Expect(objects[0].Name).To(Equal(testObjects[0].Name))

	objects, err = iter.Page(3, 0)
	g.Expect(err).To(BeNil())
	g.Expect(len(objects)).To(Equal(3))
	g.Expect(objects[0].Name).To(Equal(testObjects[3].Name))

	objects, err = iter.Page(3, 0)
	g.Expect(err).To(BeNil())
	g.Expect(len(objects)).To(Equal(1))
	g.Expect(objects[0].Name).To(Equal(testObjects[6].Name))

	// With pagination and with an initial offset.
	iter, err = store.GetObjects(context.Background(), nil, nil)
	g.Expect(err).To(BeNil())

	objects, err = iter.Page(2, 2)
	g.Expect(err).To(BeNil())
	g.Expect(len(objects)).To(Equal(2))
	g.Expect(objects[0].Name).To(Equal(testObjects[2].Name))

	objects, err = iter.Page(2, 0)
	g.Expect(err).To(BeNil())
	g.Expect(len(objects)).To(Equal(2))
	g.Expect(objects[0].Name).To(Equal(testObjects[4].Name))

	objects, err = iter.Page(2, 0)
	g.Expect(err).To(BeNil())
	g.Expect(len(objects)).To(Equal(1))
	g.Expect(objects[0].Name).To(Equal(testObjects[6].Name))
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
					Category:   configuration.CategoryAutomation,
				},
				{
					Cluster:    "test-cluster",
					Name:       "otherName",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
					Category:   configuration.CategoryAutomation,
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
					Category:   configuration.CategoryAutomation,
				},
			},
			want: []models.Object{{
				Cluster:    "test-cluster",
				Name:       "otherName",
				Namespace:  "namespace",
				Kind:       "ValidKind",
				APIGroup:   "example.com",
				APIVersion: "v1",
				Category:   configuration.CategoryAutomation,
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
					Category:   configuration.CategoryAutomation,
				},
				{
					Cluster:    "test-cluster",
					Name:       "otherName",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
					Category:   configuration.CategoryAutomation,
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
					Category:   configuration.CategoryAutomation,
				},
				{
					Cluster:    "test-cluster",
					Name:       "otherName",
					Namespace:  "namespace",
					Kind:       "ValidKind",
					APIGroup:   "example.com",
					APIVersion: "v1",
					Category:   configuration.CategoryAutomation,
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
			Category:   configuration.CategoryAutomation,
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
			Category:   configuration.CategoryAutomation,
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
		AccessibleKinds:         []string{"example.com/SomeKind"},
		ProvidedByRole:          fmt.Sprintf("%s/%s", role.Kind, role.Name),
		ProvidedByBinding:       fmt.Sprintf("%s/%s", rb.Kind, rb.Name),
		AccessibleResourceNames: []string{},
	}

	diff := cmp.Diff(expected, r[0], cmpopts.IgnoreFields(models.Subject{}, "ID", "CreatedAt", "UpdatedAt"))

	if diff != "" {
		t.Errorf("GetAccessRules() mismatch (-want +got):\n%s", diff)
	}
}

func TestStoreUnstructured(t *testing.T) {
	g := NewGomegaWithT(t)
	input := []byte(`{"myKey": "someValue"}`)
	rawMsg := json.RawMessage(input)
	store, db := createStore(t)

	obj := models.Object{
		Cluster:      "test-cluster",
		Name:         "someName",
		Namespace:    "namespace",
		Kind:         "ValidKind",
		APIGroup:     "example.com",
		APIVersion:   "v1",
		Category:     configuration.CategoryAutomation,
		Unstructured: rawMsg,
	}

	g.Expect(store.StoreObjects(context.Background(), []models.Object{obj})).To(Succeed())

	sqlDB, err := db.DB()
	g.Expect(err).To(BeNil())

	result := []byte{}

	g.Expect(sqlDB.QueryRow("SELECT unstructured FROM objects").Scan(&result)).To(Succeed())

	g.Expect(result).To(Equal(input))
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
