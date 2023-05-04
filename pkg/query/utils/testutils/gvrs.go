package testutils

import (
	"fmt"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
)

const (
	HelmReleaseResourceName   = "helmreleases"
	KustomizationResourceName = "kustomizations"
)

// Utils function to allow test cases to have a default mapper to use for testing
// when operations between gvrs and gvks are present
func CreateDefaultResourceKindMap() (map[string]string, error) {
	return createDefaultResourceKindMap([]string{
		HelmReleaseResourceName, KustomizationResourceName,
	})
}

func createDefaultResourceKindMap(resources []string) (map[string]string, error) {
	var defaultResourcesKindMap = make(map[string]string)
	for _, resource := range resources {
		kind, err := mapResourceToKind(resource)
		if err != nil {
			return nil, fmt.Errorf("cannot resolve resources:%v", resources)
		}
		defaultResourcesKindMap[resource] = kind
	}
	return defaultResourcesKindMap, nil
}

// Returns the corresponding rbac resource reference based on
// https://kubernetes.io/docs/reference/access-authn-authz/rbac/#referring-to-resources
// https://book.kubebuilder.io/cronjob-tutorial/gvks.html#kinds-and-resources
func mapResourceToKind(resource string) (string, error) {
	switch resource {
	case HelmReleaseResourceName:
		return helmv2.HelmReleaseKind, nil
	case KustomizationResourceName:
		return kustomizev1.KustomizationKind, nil
	default:
		return "", fmt.Errorf("cannot resolve not supported: %s", resource)
	}
}
