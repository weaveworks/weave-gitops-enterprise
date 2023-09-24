package utils

import (
	"testing"

	"github.com/alecthomas/assert"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8s_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type Client struct {
	Client k8s_client.Client
}

// TestGetRepoUrl tests the GetRepoUrl function
func TestGetRepoUrl(t *testing.T) {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		v1.AddToScheme,
		sourcev1.AddToScheme,
	}
	err := schemeBuilder.AddToScheme(scheme)
	if err != nil {
		t.Fatal(err)
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(&sourcev1.GitRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "flux-system",
				Namespace: "flux-system",
			},
			Spec: sourcev1.GitRepositorySpec{
				URL: "ssh://github.com/fluxcd/flux2-kustomize-helm-example",
			},
		}).Build()

	expectedRepoUrl := "github.com:fluxcd/flux2-kustomize-helm-example"

	repoUrl, err := GetRepoUrl(fakeClient, "flux-system", "flux-system")

	assert.NoError(t, err)
	assert.Equal(t, expectedRepoUrl, repoUrl)
}

// TestGetRepoBranch tests the GetRepoBranch function
func TestGetRepoBranch(t *testing.T) {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		v1.AddToScheme,
		sourcev1.AddToScheme,
	}
	err := schemeBuilder.AddToScheme(scheme)
	if err != nil {
		t.Fatal(err)
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(&sourcev1.GitRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "flux-system",
				Namespace: "flux-system",
			},
			Spec: sourcev1.GitRepositorySpec{
				URL: "ssh://github.com/fluxcd/flux2-kustomize-helm-example",
				Reference: &sourcev1.GitRepositoryRef{
					Branch: "main",
				},
			},
		}).Build()

	expectedRepoBranch := "main"

	repoBranch, err := GetRepoBranch(fakeClient, "flux-system", "flux-system")

	assert.NoError(t, err)
	assert.Equal(t, expectedRepoBranch, repoBranch)
}

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

	repoPath, err := GetRepoPath(fakeClient, "flux-system", "flux-system")

	assert.NoError(t, err)
	assert.Equal(t, expectedRepoPath, repoPath)
}
