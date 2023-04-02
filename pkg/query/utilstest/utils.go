package utilstest

import (
	"fmt"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	HelmReleaseResourceName   = "helmreleases"
	KustomizationResourceName = "kustomizations"
)

func CreateAllowedResourcesMapForApplications() (map[string]string, error) {
	return createAllowedResourcesMap([]string{
		"helmreleases", "kustomizations",
	})
}

// Returns the corresponding rbac resource reference based on
// https://kubernetes.io/docs/reference/access-authn-authz/rbac/#referring-to-resources
// https://book.kubebuilder.io/cronjob-tutorial/gvks.html#kinds-and-resources
func gvksFromResource(resource string) ([]schema.GroupVersionKind, error) {
	switch resource {
	case HelmReleaseResourceName:
		return []schema.GroupVersionKind{
			schema.FromAPIVersionAndKind(helmv2.GroupVersion.String(), helmv2.HelmReleaseKind),
		}, nil
	case KustomizationResourceName:
		return []schema.GroupVersionKind{
			schema.FromAPIVersionAndKind(v1beta2.GroupVersion.String(), v1beta2.KustomizationKind),
		}, nil
	default:
		return []schema.GroupVersionKind{}, fmt.Errorf("cannot resolve not supported: %s", resource)
	}
}

func createAllowedResourcesMap(resources []string) (map[string]string, error) {
	var allowedResourcesByGVK = make(map[string]string)
	for _, resource := range resources {
		allowedGvks, err := gvksFromResource(resource)
		if err != nil {
			return nil, fmt.Errorf("cannot resolve resources:%v", resources)
		}
		for _, gvk := range allowedGvks {
			allowedResourcesByGVK[resource] = gvk.Kind
		}
	}
	return allowedResourcesByGVK, nil
}
