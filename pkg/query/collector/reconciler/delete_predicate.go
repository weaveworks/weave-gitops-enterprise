package reconciler

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type DeletePredicate struct {
	predicate.Funcs
}

func (DeletePredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil || e.ObjectNew == nil {
		return false
	}

	return e.ObjectOld.GetGeneration() < e.ObjectNew.GetGeneration()
}
