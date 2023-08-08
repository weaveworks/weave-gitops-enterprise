package testutils

import (
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	v1 "github.com/fluxcd/kustomize-controller/api/v1"
)

const (
	HelmReleaseResourceName   = "helmreleases"
	KustomizationResourceName = "kustomizations"
)

// CreateDefaultResourceKindMap utils function to allow test cases to have a default mapper to use for testing
// when operations between gvrs and gvks are present
func CreateDefaultResourceKindMap() (map[string]string, error) {
	return map[string]string{
		helmv2.HelmReleaseKind: HelmReleaseResourceName,
		v1.KustomizationKind:   KustomizationResourceName,
	}, nil
}
