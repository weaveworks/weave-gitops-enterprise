package steps

import (
	"testing"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops-enterprise/test/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_getRepoPath(t *testing.T) {
	fakeClient := utils.CreateFakeClient(t, &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "flux-system",
			Namespace: "flux-system",
		},
		Spec: kustomizev1.KustomizationSpec{
			Path: "clusters/production",
		}})

	expectedRepoPath := "clusters/production"

	repoPath, err := getRepoPath(fakeClient, "flux-system", "flux-system")

	assert.NoError(t, err)
	assert.Equal(t, expectedRepoPath, repoPath)
}
