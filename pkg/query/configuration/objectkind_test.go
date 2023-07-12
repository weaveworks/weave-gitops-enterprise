package configuration

import (
	v1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	. "github.com/onsi/gomega"

	"testing"
)

// TestObjectsKinds test that default object kinds meet the expected contract
// like being in the expected flux api version. For example, flux v1 available
// kinds should be using v1 api version
func TestObjectsKinds(t *testing.T) {

	g := NewWithT(t)

	t.Run("should contain v1 kustomizations", func(t *testing.T) {
		g.Expect(KustomizationObjectKind.Gvk.GroupVersion()).To(BeIdenticalTo(v1.GroupVersion))
	})

	t.Run("should contain v1 gitrepositories", func(t *testing.T) {
		g.Expect(GitRepositoryObjectKind.Gvk.GroupVersion()).To(BeIdenticalTo(sourcev1.GroupVersion))

	})
}
