package store

import (
	"context"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
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

func TestAdd(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	dbDir, err := os.MkdirTemp("", "db")
	g.Expect(err).To(BeNil())
	store, _ := newInMemoryStore(dbDir, log)
	ctx := context.Background()

	tests := []struct {
		name       string
		document   Document
		errPattern string
	}{
		{
			name:       "cannot add document for nil document",
			errPattern: "invalid document",
		},
		{
			name:       "cannot add document for empty document",
			document:   Document{},
			errPattern: "invalid document",
		},
		{
			name:       "cannot add document for empty document name",
			document:   Document{},
			errPattern: "invalid document",
		},
		{
			name: "cannot add document for empty document namespace",
			document: Document{
				Name: "document",
			},
			errPattern: "invalid document",
		},
		{
			name: "cannot add document for empty document kind",
			document: Document{
				Name:      "name",
				Namespace: "namespace",
			},
			errPattern: "invalid document",
		},
		{
			name: "can add document for a valid kind",
			document: Document{
				Name:      "name",
				Namespace: "namespace",
				Kind:      "ValidKind",
			},
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := store.Add(ctx, tt.document)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			g.Expect(id > 0).To(BeTrue())
		})
	}

}
