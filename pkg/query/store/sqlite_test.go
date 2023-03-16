package store

import (
	"context"
	"os"
	"testing"

	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
)

func TestNewSQLiteStore(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	dbDir, err := os.MkdirTemp("", "db")
	g.Expect(err).To(BeNil())

	tests := []struct {
		name       string
		location   string
		errPattern string
	}{
		{
			name:       "cannot create store without location",
			errPattern: "invalid location",
		},
		{
			name:       "can create store with valid arguments",
			location:   dbDir,
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newSQLiteStore(tt.location, log)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
		})
	}
}

func TestGetObjects(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	dbDir, err := os.MkdirTemp("", "db")
	g.Expect(err).To(BeNil())
	store, _ := newSQLiteStore(dbDir, log)
	ctx := context.Background()

	object := models.Object{
		Cluster:   "test-cluster",
		Name:      "name",
		Namespace: "namespace",
		Kind:      "ValidKind",
	}

	err = store.StoreObjects(ctx, []models.Object{object})
	g.Expect(err).To(BeNil())

	tests := []struct {
		name       string
		object     models.Object
		errPattern string
	}{
		{
			name:       "can get objects",
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objects, err := store.GetObjects(context.Background())
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(len(objects) > 0).To(BeTrue())
			g.Expect(objects[0].Name == "name").To(BeTrue())
		})
	}

}
