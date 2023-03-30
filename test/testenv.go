package test

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/weaveworks/weave-gitops/pkg/testutils"
)

func getRepoRoot(t *testing.T) string {
	t.Helper()
	cmdOut, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()

	if err != nil {
		t.Fatal(err)
	}

	return strings.TrimSpace(string(cmdOut))
}

// StartTestEnv starts a new Kubernetes test environment using `envtest`
func StartTestEnv(t *testing.T) *testutils.K8sTestEnv {
	t.Helper()
	envTestPath := fmt.Sprintf("%s/tools/bin/envtest", getRepoRoot(t))
	t.Setenv("KUBEBUILDER_ASSETS", envTestPath)

	k8sEnv, err := testutils.StartK8sTestEnvironment([]string{
		"../tools/testcrds",
	})
	t.Cleanup(func() {
		k8sEnv.Stop()
	})

	if err != nil {
		t.Fatal(err)
	}

	return k8sEnv
}
