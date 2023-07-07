package cleaner

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/configuration"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store/storefakes"
)

func TestObjectCleaner(t *testing.T) {
	g := NewWithT(t)
	s := storefakes.FakeStore{}

	cfg := configuration.BucketObjectKind
	cfg.RetentionPolicy = configuration.RetentionPolicy(60 * time.Second)

	objs := []models.Object{
		{
			Cluster:    "cluster1",
			Kind:       cfg.Gvk.Kind,
			Name:       "name1",
			APIGroup:   cfg.Gvk.Group,
			APIVersion: cfg.Gvk.Version,
			// Deleted 1 hour ago, our retention policy is 60s, so this should be deleted.
			KubernetesDeletedAt: time.Now().Add(-time.Hour),
		},
		{
			Cluster:    "cluster1",
			Kind:       cfg.Gvk.Kind,
			Name:       "name2",
			APIGroup:   cfg.Gvk.Group,
			APIVersion: cfg.Gvk.Version,
			// Deleted 10s ago, our retention policy is 60s, so this should not be deleted.
			KubernetesDeletedAt: time.Now().Add(10 * -time.Second),
		},
	}

	iter := storefakes.FakeIterator{}
	iter.AllReturnsOnCall(0, objs, nil)
	// Pretend the first object is deleted on the second call
	iter.AllReturnsOnCall(1, objs[1:2], nil)
	s.GetAllObjectsReturns(&iter, nil)

	index := storefakes.FakeIndexWriter{}

	oc := objectCleaner{
		log:    logr.Discard(),
		store:  &s,
		idx:    &index,
		config: []configuration.ObjectKind{cfg},
	}

	// Skipping starting the cleaner here to avoid dealing with async and time stuff.
	g.Expect(oc.removeOldObjects(context.Background())).To(Succeed())

	g.Expect(s.DeleteObjectsCallCount()).To(Equal(1))
	_, result := s.DeleteObjectsArgsForCall(0)
	g.Expect(result).To(Equal([]models.Object{objs[0]}))

	g.Expect(index.RemoveCallCount()).To(Equal(1))
	_, idxResult := index.RemoveArgsForCall(0)
	g.Expect(idxResult).To(Equal([]models.Object{objs[0]}))

	// Call it again, make sure it doesn't delete anything.
	// The `iter` mock will return only the second object, which is not old enough to be deleted.
	g.Expect(oc.removeOldObjects(context.Background())).To(Succeed())
	g.Expect(s.DeleteObjectsCallCount()).To(Equal(1))

}
