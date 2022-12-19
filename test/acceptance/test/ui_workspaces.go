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

func installTestWorkspaces(clusterName string) {
	ginkgo.By(fmt.Sprintf("Add workspaces to the %s cluster", clusterName), func() {
		createTenant(path.Join(testDataPath, "tenancy", "multiple-tenant.yaml"))

	})
}

func verifyFilterWorkspacesByClusterName(clusterName string, workspaceName string) {
	ginkgo.By(fmt.Sprintf("Filter Workspaces By cluster name: '%s'", clusterName), func() {

		workspacesList := pages.GetWorkspacesPage(webDriver)
		filterID := "Cluster:" + clusterName
		searchPage := pages.GetSearchPage(webDriver)
		searchPage.SelectFilter("Cluster", filterID)
		gomega.Eventually(workspacesList.CountWorkspaces()).Should(gomega.BeNumerically("=", 2), fmt.Sprintf("The number of workspaces for selected cluster:  '%s' should equal to 2", clusterName))
		// Clear the filter
		searchPage.SelectFilter("Cluster", filterID)
	})

	ginkgo.By(fmt.Sprintf("Filter Workspaces By workspace name: '%s'", workspaceName), func() {

		workspacesList := pages.GetWorkspacesPage(webDriver)
		filterID := "Name:" + workspaceName
		searchPage := pages.GetSearchPage(webDriver)
		searchPage.SelectFilter("Name", filterID)
		gomega.Eventually(workspacesList.CountWorkspaces()).Should(gomega.BeNumerically("=", 2), fmt.Sprintf("The number of workspaces for selected Name:  '%s' should equal to 2", workspaceName))
		// Clear the filter
		searchPage.SelectFilter("Name", filterID)
	})
}

func verifySearchWorkspaceByName(workspaceName string) {
	// Search by Workspace Name in the workspaces list.
	ginkgo.By(fmt.Sprintf("And search by Workspace '%s' in the workspaces list", workspaceName), func() {

		searchPage := pages.GetSearchPage(webDriver)
		searchPage.SearchName(workspaceName)
		gomega.Eventually(func(g gomega.Gomega) int {
			return pages.CountAppViolations(webDriver)
		}).Should(gomega.BeNumerically("=", 1), "Search should return '1' workspace in the list")

	})
}

var _ = ginkgo.Describe("Multi-Cluster Control Plane Workspaces", ginkgo.Label("ui", "workspaces"), func() {

	ginkgo.BeforeEach(ginkgo.OncePerOrdered, func() {
		// Delete the oidc user default roles/rolebindings because the same user is used as a tenant
		_ = runCommandPassThrough("kubectl", "delete", "-f", path.Join(testDataPath, "rbac/user-role-bindings.yaml"))

		gomega.Expect(webDriver.Navigate(testUiUrl)).To(gomega.Succeed())
		if !pages.ElementExist(pages.Navbar(webDriver).Title, 3) {
			loginUser()
		}
	})

	ginkgo.AfterEach(ginkgo.OncePerOrdered, func() {
		// Create the oidc user default roles/rolebindings afte tenant tests completed
		_ = runCommandPassThrough("kubectl", "apply", "-f", path.Join(testDataPath, "rbac/user-role-bindings.yaml"))
	})

	ginkgo.Context("[UI] Workspaces can be configured on management cluster", func() {

		workspaceName := "test-team"
		workspaceNamespaces := "test-kustomization, test-system"
		workspaceClusterName := "management"

		ginkgo.JustBeforeEach(func() {
			installTestWorkspaces("management")

		})

		ginkgo.JustAfterEach(func() {

			defer deleteTenants([]string{getTenantYamlPath()})
		})

		ginkgo.It("Verify Workspaces can be configured on management cluster and dashboard is updated accordingly", ginkgo.Label("integration", "workspaces"), func() {
			existingWorkspacesCount := getWorkspacesCount()

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
			verifyFilterWorkspacesByClusterName("management", workspaceName)
			verifySearchWorkspaceByName(workspaceName)
			gomega.Eventually(workspaceInfo.Name.Text).Should(gomega.Equal(workspaceName), "Failed to get the workspace by its name Value in the Workspaces List")

		})
	})

	ginkgo.Context("[UI] Workspaces can be configured on leaf cluster", ginkgo.Label("leaf-workspaces"), func() {
		var mgmtClusterContext string
		var leafClusterContext string
		var leafClusterkubeconfig string
		var clusterBootstrapCopnfig string
		var gitopsCluster string
		patSecret := "workspace-pat"
		bootstrapLabel := "bootstrap"
		leafClusterName := "workspaces-leaf-cluster-test"
		leafClusterNamespace := "default"

		workspaceName := "dev-team"
		workspaceNamespaces := "dev-system"
		workspaceClusterName := leafClusterName

		ginkgo.JustBeforeEach(func() {

			gomega.Expect(webDriver.Navigate(testUiUrl)).To(gomega.Succeed())
			if !pages.ElementExist(pages.Navbar(webDriver).Title, 3) {
				loginUser()
			}
			mgmtClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
			createCluster("kind", leafClusterName, "")
			leafClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
		})

		ginkgo.JustAfterEach(func() {
			useClusterContext(mgmtClusterContext)

			deleteSecret([]string{leafClusterkubeconfig, patSecret}, leafClusterNamespace)
			_ = runCommandPassThrough("kubectl", "delete", "-f", clusterBootstrapCopnfig)
			_ = runCommandPassThrough("kubectl", "delete", "-f", gitopsCluster)

			deleteCluster("kind", leafClusterName, "")
			// Delete the test workspaces
			defer deleteTenants([]string{getTenantYamlPath()})

		})

		ginkgo.It("Verify Workspaces can be configured on leaf cluster and dashboard is updated accordingly", ginkgo.Label("integration", "workspaces", "leaf-workspaces"), func() {
			existingWorkspacesCount := getWorkspacesCount()
			leafClusterkubeconfig = createLeafClusterKubeconfig(leafClusterContext, leafClusterName, leafClusterNamespace)

			// Install policy agent ,and workspaces on leaf cluster
			installPolicyAgent(leafClusterName)
			installTestWorkspaces(leafClusterName)

			useClusterContext(mgmtClusterContext)
			createPATSecret(leafClusterNamespace, patSecret)
			clusterBootstrapCopnfig = createClusterBootstrapConfig(leafClusterName, leafClusterNamespace, bootstrapLabel, patSecret)
			gitopsCluster = connectGitopsCluster(leafClusterName, leafClusterNamespace, bootstrapLabel, leafClusterkubeconfig)
			createLeafClusterSecret(leafClusterNamespace, leafClusterkubeconfig)

			waitForLeafClusterAvailability(leafClusterName, "Ready")
			addKustomizationBases("leaf", leafClusterName, leafClusterNamespace)

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

			ginkgo.By(fmt.Sprintf("And filter leaf cluster '%s' workspaces", leafClusterName), func() {
				filterID := "clusterName: " + leafClusterNamespace + `/` + leafClusterName
				searchPage := pages.GetSearchPage(webDriver)
				searchPage.SelectFilter("cluster", filterID)
			})

			ginkgo.By(fmt.Sprintf("And verify '%s' workspace Name", workspaceName), func() {
				gomega.Eventually(workspaceInfo.Name).Should(matchers.MatchText(workspaceName), fmt.Sprintf("Failed to list '%s' workspace in the Workspaces List", workspaceName))
			})

			ginkgo.By(fmt.Sprintf("And verify '%s' workspace Namespaces", workspaceName), func() {
				gomega.Eventually(workspaceInfo.Namespaces).Should(matchers.MatchText(workspaceNamespaces), fmt.Sprintf("Failed to get the expected '%s' workspace Namespaces", workspaceName))
			})

			ginkgo.By(fmt.Sprintf("And verify '%s' workspace Cluster", workspaceName), func() {
				gomega.Eventually(workspaceInfo.Cluster).Should(matchers.MatchText(workspaceClusterName), fmt.Sprintf("Failed to get the expected %[1]v workspace Cluster: %[1]v", workspaceName))
			})
			verifyFilterWorkspacesByClusterName(leafClusterName, workspaceName)
			verifySearchWorkspaceByName(workspaceName)
			gomega.Eventually(workspaceInfo.Name.Text).Should(gomega.Equal(workspaceName), "Failed to get the workspace by its name Value in the Workspaces List")

		})
	})

})
