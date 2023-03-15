package store

import (
	"context"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"testing"
)

func TestNewIndexerStore(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)

	tests := []struct {
		name       string
		url        string
		errPattern string
	}{
		{
			name:       "cannot create store without url",
			errPattern: "invalid url",
		},
		{
			name:       "can create store with valid arguments",
			url:        "http://meilisearch.meilisearch.svc.cluster.local:7700/",
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := newIndexerStore(tt.url, log)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(store).NotTo(BeNil())
			g.Expect(store.client).NotTo(BeNil())
		})
	}
}

func TestIndexerStore_StoreObjects(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	store, _ := newIndexerStore("http://meilisearch.meilisearch.svc.cluster.local:7700/", log)
	ctx := context.Background()

	tests := []struct {
		name       string
		objects    []models.Object
		errPattern string
	}{
		{
			name:       "cannot add objects for nil objects",
			errPattern: "invalid objects",
		},
		{
			name:       "could add objects for empty objects",
			objects:    []models.Object{},
			errPattern: "",
		},
		{
			name: "cannot add objects for empty object name",
			objects: []models.Object{
				{},
			},
			errPattern: "invalid object",
		},
		{
			name: "cannot add objects for empty objects namespace",
			objects: []models.Object{
				{
					Name: "objects",
				},
			},
			errPattern: "invalid object",
		},
		{
			name: "cannot add objects for empty objects kind",
			objects: []models.Object{
				{
					Name:      "name",
					Namespace: "namespace",
				},
			},
			errPattern: "invalid object",
		},
		{
			name: "can add objects for a valid kind",
			objects: []models.Object{
				{
					Name:      "name",
					Namespace: "namespace",
					Kind:      "ValidKind",
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
		})
	}

}
