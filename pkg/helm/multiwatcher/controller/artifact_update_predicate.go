package controller

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
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

	oldSource, ok := e.ObjectOld.(*sourcev1beta2.HelmRepository)
	if !ok {
		return false
	}

	newSource, ok := e.ObjectNew.(*sourcev1beta2.HelmRepository)
	if !ok {
		return false
	}

	if filterAnnotation(oldSource) != filterAnnotation(newSource) {
		return true
	}

	if oldSource.GetArtifact() == nil && newSource.GetArtifact() != nil {
		return true
	}

	// There is no way that the old artifact is newer here. We just care that they are of a different revision.
	// Kubernetes takes care of setting old and new accordingly.
	if oldArtifact, newArtifact := oldSource.GetArtifact(), newSource.GetArtifact(); oldArtifact != nil && newArtifact != nil {
		if oldArtifact.Revision != newArtifact.Revision {
			return true
		}

		if oldArtifact.URL != newArtifact.URL {
			return true
		}
	}

	return false
}

func filterAnnotation(hr *sourcev1beta2.HelmRepository) string {
	return hr.GetAnnotations()[HelmVersionFilterAnnotation]
}
