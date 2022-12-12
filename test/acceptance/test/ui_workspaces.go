package acceptance

import (
	"fmt"
	"path"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

func installTestWorkspaces(clusterName string, workspacesYaml string) {
	ginkgo.By(fmt.Sprintf("Add/Install test Policies to the %s cluster", clusterName), func() {
		err := runCommandPassThrough("go", " run", " main.go", "create", "tenants", "--from-file", workspacesYaml, "--prune")
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to install workspaces to cluster '%s'", clusterName))
	})
}

func DescribeWorkspaces(gitopsTestRunner GitopsTestRunner) {
	var _ = ginkgo.Describe("Multi-Cluster Control Plane Workspaces", func() {

		ginkgo.BeforeEach(func() {
			gomega.Expect(webDriver.Navigate(test_ui_url)).To(gomega.Succeed())

			if !pages.ElementExist(pages.Navbar(webDriver).Title, 3) {
				loginUser()
			}
		})

		ginkgo.Context("[UI] Workspaces can be configured on management cluster", func() {
			var workspacesYaml string

			workspaceName := "bar-tenant"
			workspaceNamespaces := "foobar-ns"
			workspaceClusterName := "management"

			ginkgo.JustBeforeEach(func() {
				workspacesYaml = path.Join(getCheckoutRepoPath(), "pkg", "tenancy", "testdata", "example.yaml")
			})

			ginkgo.JustAfterEach(func() {
				_ = gitopsTestRunner.KubectlDelete([]string{}, workspacesYaml)
			})

			ginkgo.FIt("Verify Workspaces can be configured on management cluster and dashboard is updated accordingly", ginkgo.Label("integration", "policy"), func() {
				existingWorkspacesCount := getWorkspacesCount()
				installTestWorkspaces("management", workspacesYaml)

				pages.NavigateToPage(webDriver, "Workspaces")
				WorkspacesPage := pages.GetWorkspacesPage(webDriver)

				ginkgo.By("And wait for workspaces to be visibe on the dashboard", func() {
					gomega.Eventually(WorkspacesPage.WorkspaceHeader).Should(matchers.BeVisible())

					totalWorkspacesCount := existingWorkspacesCount + 2
					gomega.Eventually(func(g gomega.Gomega) int {
						gomega.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
						time.Sleep(POLL_INTERVAL_1SECONDS)
						return WorkspacesPage.CountWorkspaces()
					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_3SECONDS).Should(gomega.Equal(totalWorkspacesCount), fmt.Sprintf("There should be '%d' workspaces in Workspaces table but found '%d'", totalWorkspacesCount, existingWorkspacesCount))

				})

				workspaceInfo := WorkspacesPage.FindWorkspacInList(workspaceName)
				ginkgo.By(fmt.Sprintf("And verify '%s' workspace Name", workspaceName), func() {
					gomega.Eventually(workspaceInfo.Name).Should(matchers.MatchText(workspaceName), fmt.Sprintf("Failed to list '%s' workspace in the Workspaces List", workspaceName))
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' workspace Namespaces", workspaceName), func() {
					gomega.Eventually(workspaceInfo.Namespaces).Should(matchers.MatchText(workspaceNamespaces), fmt.Sprintf("Failed to get the expected '%s' workspace Namespaces", workspaceName))
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' workspace Cluster", workspaceName), func() {
					gomega.Eventually(workspaceInfo.Cluster).Should(matchers.MatchText(workspaceClusterName), fmt.Sprintf("Failed to get the expected %[1]v workspace Cluster: %[1]v", workspaceName))
				})

			})
		})
	})
}
