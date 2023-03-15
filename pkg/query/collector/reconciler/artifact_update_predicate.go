package reconciler

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ArtifactUpdatePredicate triggers an update event when a HelmRepository artifact revision changes.
// i.e.: Repo information was updated.
type ArtifactUpdatePredicate struct {
	predicate.Funcs
}

func (ArtifactUpdatePredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil || e.ObjectNew == nil {
		return false
	}

	return true
}
