package adapters

import (
	"fmt"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"

	"github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FluxObject interface {
	GetConditions() []metav1.Condition
}

type ObjectStatus string

const (
	Success ObjectStatus = "Success"
	Failed  ObjectStatus = "Failed"
)

// TODO can we generlise it?
func ToFluxObject(obj client.Object) (FluxObject, error) {
	switch t := obj.(type) {
	case *v2beta1.HelmRelease:
		return t, nil
	case *kustomizev1.Kustomization:
		return t, nil
	case *sourcev1.HelmRepository:
		return t, nil
	case *sourcev1.HelmChart:
		return t, nil
	case *sourcev1.Bucket:
		return t, nil
	case *sourcev1.GitRepository:
		return t, nil
	case *sourcev1.OCIRepository:
		return t, nil
	}

	return nil, fmt.Errorf("unknown object type: %T", obj)
}

func Status(fo FluxObject) ObjectStatus {
	for _, c := range fo.GetConditions() {
		if c.Type == "Ready" || c.Type == "Available" {
			if c.Status == "True" {
				return Success
			}

			return Failed
		}
	}

	return Failed
}

func Message(fo FluxObject) string {
	for _, c := range fo.GetConditions() {
		if c.Type == "Ready" || c.Type == "Available" {
			return c.Message
		}
	}

	return ""
}
