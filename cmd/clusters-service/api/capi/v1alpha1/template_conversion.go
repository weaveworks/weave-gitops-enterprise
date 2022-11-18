package v1alpha1

import (
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	capiv1alpha2 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/capi/v1alpha2"
)

// ConvertTo converts this Template to the Hub version (v1alpha2).
func (src *CAPITemplate) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*capiv1alpha2.CAPITemplate)

	dst.ObjectMeta = src.ObjectMeta

	return nil
}

// ConvertFrom converts from the Hub version (v1alpha2) to this version.
func (dst *CAPITemplate) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*capiv1alpha2.CAPITemplate)

	dst.ObjectMeta = src.ObjectMeta

	return nil
}
