package testutils

import (
	"github.com/fluxcd/helm-controller/api/v2beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewHelmRelease creates a test helm release out of the parameters.It uses a decorator pattern to add custom configuration.
func NewHelmRelease(name string, namespace string, opts ...func(*v2beta1.HelmRelease)) *v2beta1.HelmRelease {
	helmRelease := &v2beta1.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v2beta1.GroupVersion.Version,
			Kind:       v2beta1.HelmReleaseKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	for _, opt := range opts {
		opt(helmRelease)
	}

	return helmRelease
}
