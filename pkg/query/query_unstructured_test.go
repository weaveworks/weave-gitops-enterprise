package query

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	store "github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

func TestQueryUnstructured(t *testing.T) {
	// Test that we can add some unstructured data to a models.Object and query against it.

	g := NewGomegaWithT(t)

	dir, err := os.MkdirTemp("", "test")
	g.Expect(err).NotTo(HaveOccurred())

	db, err := store.CreateSQLiteDB(dir)
	g.Expect(err).NotTo(HaveOccurred())

	s, err := store.NewSQLiteStore(db, logr.Discard())
	g.Expect(err).NotTo(HaveOccurred())

	idxDir, err := os.MkdirTemp("", "indexer-test")
	g.Expect(err).NotTo(HaveOccurred())

	idx, err := store.NewIndexer(s, idxDir, logr.Discard())
	g.Expect(err).NotTo(HaveOccurred())

	myStruct := struct {
		SomeField string `json:"name"`
		Location  string `json:"location"`
		Count     int    `json:"count"`
	}{
		SomeField: "someName",
		Location:  "someLocation",
		Count:     1,
	}

	b, err := json.Marshal(myStruct)
	g.Expect(err).NotTo(HaveOccurred())

	objects := []models.Object{
		{
			Cluster:      "test-cluster",
			Name:         "alpha",
			Namespace:    "namespace",
			Kind:         "ValidKind",
			APIGroup:     "example.com",
			APIVersion:   "v1",
			Unstructured: b,
			Category:     configuration.CategoryEvent,
		},
	}

	q := &qs{
		log:        logr.Discard(),
		debug:      logr.Discard(),
		r:          s,
		index:      idx,
		authorizer: allowAll,
	}

	g.Expect(s.StoreObjects(context.Background(), objects)).To(Succeed())

	g.Expect(idx.Add(context.Background(), objects)).To(Succeed())

	qy := &query{
		terms: "someName",
	}

	ctx := auth.WithPrincipal(context.Background(), &auth.UserPrincipal{
		ID: "test",
		Groups: []string{
			"group-a",
		},
	})

	result, err := q.RunQuery(ctx, qy, qy)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(result).To(HaveLen(1))
	g.Expect(result[0].Cluster).To(Equal("test-cluster"))

	// Indexing blobs can remove facets, so make sure they exist.
	facets, err := q.ListFacets(ctx)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(facets["cluster"]).To(ContainElements("test-cluster"))
}
