package utils

import (
	"testing"

	"github.com/alecthomas/assert"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestGetRepoPath tests the GetRepoPath function
func TestGetRepoPath(t *testing.T) {
	fakeClient, err := CreateFakeClient(t, &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "flux-system",
			Namespace: "flux-system",
		},
		Spec: kustomizev1.KustomizationSpec{
			Path: "clusters/production",
		}})
	if err != nil {
		t.Fatalf("error creating fake client: %v", err)
	}

	expectedRepoPath := "clusters/production"

	repoPath, err := getRepoPath(fakeClient, "flux-system", "flux-system")

	assert.NoError(t, err)
	assert.Equal(t, expectedRepoPath, repoPath)
}
