package profiles

import (
	"testing"

	"github.com/alecthomas/assert"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/utils"
)

func TestConstructCAPIValues(t *testing.T) {

	// Mocked values for our utilities
	mockBranch := "main"
	mockUrl := "https://github.com/test-repo"
	mockPath := "test-repo"

	// Set up the mocked functions
	utils.GetRepoBranchFunc = func() (string, error) {
		return mockBranch, nil
	}

	utils.GetRepoUrlFunc = func() (string, error) {
		return mockUrl, nil
	}

	utils.GetRepoPathFunc = func() (string, error) {
		return mockPath, nil
	}

	// Reset functions to original versions after test
	defer func() {
		utils.GetRepoBranchFunc = utils.GetRepoBranch
		utils.GetRepoUrlFunc = utils.GetRepoUrl
		utils.GetRepoPathFunc = utils.GetRepoPath
	}()

	templatesNamespace := "default"
	clustersNamespace := "default"

	values, err := constructCAPIValues(templatesNamespace, clustersNamespace)
	if err != nil {
		t.Error(err)
	}

	expectedValues := map[string]interface{}{
		"repositoryURL":          "https///github.com/test-repo",
		"repositoryPath":         "test-repo/clusters",
		"repositoryClustersPath": "test-repo",
		"baseBranch":             "main",
		"templates": map[string]interface{}{
			"namespace": "default",
		},
		"clusters": map[string]interface{}{
			"namespace": "default",
		},
	}

	assert.Equal(t, expectedValues, values)
}
