package store

import (
	"context"
	"os"
	"testing"

	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/sqlite"
)

func TestGetObjects(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	dbDir, err := os.MkdirTemp("", "db")
	g.Expect(err).To(BeNil())
	store, _, err := sqlite.NewStore(dbDir, log)
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
