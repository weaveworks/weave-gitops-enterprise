package store

import (
	"context"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"os"
	"testing"
)

func TestNewInMemoryStore(t *testing.T) {
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
			store, err := newInMemoryStore(tt.location, log)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(store).NotTo(BeNil())
			g.Expect(store.db).NotTo(BeNil())
		})
	}
}

func TestStoreObject(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	dbDir, err := os.MkdirTemp("", "db")
	g.Expect(err).To(BeNil())
	store, _ := newInMemoryStore(dbDir, log)
	ctx := context.Background()

	tests := []struct {
		name       string
		object     models.Object
		errPattern string
	}{
		{
			name:       "cannot add object for nil object",
			errPattern: "invalid object",
		},
		{
			name:       "cannot add object for empty object",
			object:     models.Object{},
			errPattern: "invalid object",
		},
		{
			name:       "cannot add object for empty object name",
			object:     models.Object{},
			errPattern: "invalid object",
		},
		{
			name: "cannot add object for empty object namespace",
			object: models.Object{
				Name: "object",
			},
			errPattern: "invalid object",
		},
		{
			name: "cannot add object for empty object kind",
			object: models.Object{
				Name:      "name",
				Namespace: "namespace",
			},
			errPattern: "invalid object",
		},
		{
			name: "can add object for a valid kind",
			object: models.Object{
				Name:      "name",
				Namespace: "namespace",
				Kind:      "ValidKind",
			},
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := store.StoreObject(ctx, tt.object)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(id > 0).To(BeTrue())
		})
	}
}

func TestGetObjects(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	dbDir, err := os.MkdirTemp("", "db")
	g.Expect(err).To(BeNil())
	store, _ := newInMemoryStore(dbDir, log)
	ctx := context.Background()

	object := models.Object{
		Name:      "name",
		Namespace: "namespace",
		Kind:      "ValidKind",
	}

	_, err = store.StoreObject(ctx, object)
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
			objects, err := store.GetObjects()
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
