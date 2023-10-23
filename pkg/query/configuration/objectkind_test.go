package configuration

import (
	v1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

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

func TestObjectKind_Validate(t *testing.T) {
	g := NewWithT(t)

	t.Run("should return error if gvk is missing", func(t *testing.T) {
		kind := ObjectKind{}
		g.Expect(kind.Validate()).NotTo(BeNil())
	})

	t.Run("should return error if client func is missing", func(t *testing.T) {
		kind := ObjectKind{
			Gvk: schema.GroupVersionKind{
				Kind: "test",
			},
		}
		g.Expect(kind.Validate()).NotTo(BeNil())
	})

	t.Run("should return error if add to scheme func is missing", func(t *testing.T) {
		kind := ObjectKind{
			Gvk: schema.GroupVersionKind{
				Kind: "test",
			},
			NewClientObjectFunc: func() client.Object {
				return nil
			},
		}
		g.Expect(kind.Validate().Error()).To(Equal("missing add to scheme func"))
	})

	t.Run("should return error if status func is missing", func(t *testing.T) {
		kind := ObjectKind{
			Gvk: schema.GroupVersionKind{
				Kind: "test",
			},
			AddToSchemeFunc: func(*runtime.Scheme) error {
				return nil
			},
			NewClientObjectFunc: func() client.Object {
				return nil
			},
			MessageFunc: func(obj client.Object) string {
				return ""
			},
		}
		g.Expect(kind.Validate().Error()).To(Equal("missing status func"))
	})

	t.Run("should return error if message func is missing", func(t *testing.T) {
		kind := ObjectKind{
			Gvk: schema.GroupVersionKind{
				Kind: "test",
			},
			AddToSchemeFunc: func(*runtime.Scheme) error {
				return nil
			},
			NewClientObjectFunc: func() client.Object {
				return nil
			},
			StatusFunc: func(obj client.Object) ObjectStatus {
				return Success
			},
		}
		g.Expect(kind.Validate().Error()).To(Equal("missing message func"))
	})
	t.Run("should return error if labels func is missing", func(t *testing.T) {
		kind := ObjectKind{
			Gvk: schema.GroupVersionKind{
				Kind: "test",
			},
			AddToSchemeFunc: func(*runtime.Scheme) error {
				return nil
			},
			NewClientObjectFunc: func() client.Object {
				return nil
			},
			StatusFunc: func(obj client.Object) ObjectStatus {
				return Success
			},
			MessageFunc: func(obj client.Object) string {
				return ""
			},
		}
		g.Expect(kind.Validate().Error()).To(Equal("missing labels func"))
	})

}
