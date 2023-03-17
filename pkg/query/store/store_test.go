package store

import (
	"context"
	"os"
	"testing"

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

	g.Expect(seed(db, []models.Object{obj})).To(Succeed())

	objects, err := store.GetObjects(context.Background(), nil)
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
					Cluster:   "test-cluster",
					Name:      "someName",
					Namespace: "namespace",
					Kind:      "ValidKind",
				},
				{
					Cluster:   "test-cluster",
					Name:      "otherName",
					Namespace: "namespace",
					Kind:      "ValidKind",
				},
			},
			toRemove: []models.Object{
				{
					Cluster:   "test-cluster",
					Name:      "someName",
					Namespace: "namespace",
					Kind:      "ValidKind",
				},
			},
			want: []models.Object{{
				Cluster:   "test-cluster",
				Name:      "otherName",
				Namespace: "namespace",
				Kind:      "ValidKind",
			}},
		},
		{
			name: "remove multiple objects",
			seed: []models.Object{
				{
					Cluster:   "test-cluster",
					Name:      "someName",
					Namespace: "namespace",
					Kind:      "ValidKind",
				},
				{
					Cluster:   "test-cluster",
					Name:      "otherName",
					Namespace: "namespace",
					Kind:      "ValidKind",
				},
			},
			toRemove: []models.Object{
				{
					Cluster:   "test-cluster",
					Name:      "someName",
					Namespace: "namespace",
					Kind:      "ValidKind",
				},
				{
					Cluster:   "test-cluster",
					Name:      "otherName",
					Namespace: "namespace",
					Kind:      "ValidKind",
				},
			},
			want: []models.Object{},
		},
	}
	g := NewGomegaWithT(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, db := createStore(t)

			g.Expect(seed(db, tt.seed)).To(Succeed())

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
			Cluster:   "test-cluster",
			Name:      "someName",
			Namespace: "namespace",
			Kind:      "ValidKind",
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
			Cluster:   "test-cluster",
			Name:      "someName",
			Namespace: "namespace",
			Kind:      "ValidKind",
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

func TestStoreAccessRules(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("stores access rules", func(t *testing.T) {
		store, db := createStore(t)

		accessRule := models.AccessRule{
			Cluster:         "test-cluster",
			Namespace:       "namespace",
			Principal:       "someuser",
			AccessibleKinds: []string{"example.com/v1beta2/SomeKind"},
		}

		g.Expect(store.StoreAccessRules(context.Background(), []models.AccessRule{accessRule})).To(Succeed())

		sqlDB, err := db.DB()
		g.Expect(err).To(BeNil())

		var storedAccessRule models.AccessRule
		g.Expect(sqlDB.QueryRow("SELECT id FROM access_rules").Scan(&storedAccessRule.ID)).To(Succeed())

		g.Expect(storedAccessRule.ID).To(Equal(accessRule.GetID()))
	})
	t.Run("upserts access rules", func(t *testing.T) {
		store, db := createStore(t)

		accessRule := models.AccessRule{
			Cluster:         "test-cluster",
			Namespace:       "namespace",
			Principal:       "someuser",
			AccessibleKinds: []string{"example.com/v1beta2/SomeKind"},
		}

		g.Expect(store.StoreAccessRules(context.Background(), []models.AccessRule{accessRule})).To(Succeed())

		sqlDB, err := db.DB()
		g.Expect(err).To(BeNil())

		var count int64
		g.Expect(sqlDB.QueryRow("SELECT COUNT(*) FROM access_rules").Scan(&count)).To(Succeed())
		g.Expect(count).To(Equal(int64(1)))

		g.Expect(store.StoreAccessRules(context.Background(), []models.AccessRule{accessRule})).To(Succeed())

		g.Expect(sqlDB.QueryRow("SELECT COUNT(*) FROM access_rules").Scan(&count)).To(Succeed())
		g.Expect(count).To(Equal(int64(1)))
	})

}

func createStore(t *testing.T) (Store, *gorm.DB) {
	g := NewGomegaWithT(t)
	dbDir, err := os.MkdirTemp("", "db")
	g.Expect(err).To(BeNil())

	db, err := CreateSQLiteDB(dbDir)
	g.Expect(err).To(BeNil())

	store, err := NewSQLiteStore(db)
	g.Expect(err).To(BeNil())

	return store, db
}

func seed(db *gorm.DB, rows []models.Object) error {
	withID := []models.Object{}

	for _, o := range rows {
		o.ID = o.GetID()
		withID = append(withID, o)
	}
	result := db.Create(&withID)

	return result.Error
}
