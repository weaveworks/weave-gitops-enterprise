package utils

import (
	"testing"

	"github.com/alecthomas/assert"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"github.com/fluxcd/source-controller/api/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type Client struct {
	Clientset kubernetes.Interface
}

// TestGetRepoUrl tests the GetRepoUrl function
func TestGetRepoUrl(t *testing.T) {
	// Initialize scheme
	s := scheme.Scheme
	s.AddKnownTypes(v1beta2.GroupVersion, &v1beta2.GitRepository{})

	// Create the fake client using NewClientBuilder
	fakeClient := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(&v1beta2.GitRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "flux-system",
				Namespace: "flux-system",
			},
			Spec: v1beta2.GitRepositorySpec{
				URL: "ssh://github.com/fluxcd/flux2-kustomize-helm-example",
			},
		}).Build()

	expectedRepoUrl := "github.com:fluxcd/flux2-kustomize-helm-example"

	// Call GetRepoUrl with the fake client
	repoUrl, err := GetRepoUrl(fakeClient, "flux-system", "flux-system")

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedRepoUrl, repoUrl)
}

// TestGetRepoBranch tests the GetRepoBranch function
func TestGetRepoBranch(t *testing.T) {
	// Initialize scheme
	s := scheme.Scheme
	s.AddKnownTypes(v1beta2.GroupVersion, &v1beta2.GitRepository{})

	// Create the fake client using NewClientBuilder
	fakeClient := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(&v1beta2.GitRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "flux-system",
				Namespace: "flux-system",
			},
			Spec: v1beta2.GitRepositorySpec{
				URL: "ssh://github.com/fluxcd/flux2-kustomize-helm-example",
				Reference: &v1beta2.GitRepositoryRef{
					Branch: "main",
				},
			},
		}).Build()

	expectedRepoBranch := "main"

	// Call GetRepoBranch with the fake client
	repoBranch, err := GetRepoBranch(fakeClient, "flux-system", "flux-system")

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedRepoBranch, repoBranch)
}

// TestGetRepoPath tests the GetRepoPath function
func TestGetRepoPath(t *testing.T) {
	// Initialize the scheme
	s := scheme.Scheme
	s.AddKnownTypes(kustomizev1.GroupVersion, &kustomizev1.Kustomization{})

	// Create the fake client using NewClientBuilder
	fakeClient := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(&kustomizev1.Kustomization{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "flux-system",
				Namespace: "flux-system",
			},
			Spec: kustomizev1.KustomizationSpec{
				Path: "\"clusters/production\"",
			},
		}).Build()

	expectedRepoPath := "clusters/production"

	// Call GetRepoPath with the fake client
	repoPath, err := GetRepoPath(fakeClient, "flux-system", "flux-system")

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedRepoPath, repoPath)
}
