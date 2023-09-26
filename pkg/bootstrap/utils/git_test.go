package utils

import (
	"testing"

	"github.com/alecthomas/assert"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TestGetRepoPath tests the GetRepoPath function
func TestGetRepoPath(t *testing.T) {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		v1.AddToScheme,
		kustomizev1.AddToScheme,
	}
	err := schemeBuilder.AddToScheme(scheme)
	if err != nil {
		t.Fatal(err)
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(&kustomizev1.Kustomization{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "flux-system",
				Namespace: "flux-system",
			},
			Spec: kustomizev1.KustomizationSpec{
				Path: "clusters/production",
			},
		}).Build()

	expectedRepoPath := "clusters/production"

	repoPath, err := getRepoPath(fakeClient, "flux-system", "flux-system")

	assert.NoError(t, err)
	assert.Equal(t, expectedRepoPath, repoPath)
}
