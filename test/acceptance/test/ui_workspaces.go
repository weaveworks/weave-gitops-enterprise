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

func installWorkspaces(clusterName string, workspacesYaml string) {

	ginkgo.By(fmt.Sprintf("Add test workspaces to the '%s' cluster", clusterName), func() {
		createTenant(workspacesYaml)
	})
}

func deleteWorkspaces(clusterName string) {

	ginkgo.By(fmt.Sprintf("And Finally delete test workspaces from '%s' cluster", clusterName), func() {
		deleteTenants([]string{getTenantYamlPath()})
	})
}

func verifyFilterWorkspacesByClusterName(clusterName string, workspaceName string) {
	ginkgo.By(fmt.Sprintf("Filter Workspaces By cluster name: '%s'", clusterName), func() {

		workspacesList := pages.GetWorkspacesPage(webDriver)
		filterID := "Cluster: " + clusterName
		searchPage := pages.GetSearchPage(webDriver)
		searchPage.SelectFilter("Cluster", filterID)
		gomega.Eventually(workspacesList.CountWorkspaces()).Should(gomega.BeNumerically(">=", 2), fmt.Sprintf("The number of workspaces for selected cluster:  '%s' should  equal 2", clusterName))
		// Clear the filter
		searchPage.SelectFilter("Cluster", filterID)
	})

	ginkgo.By(fmt.Sprintf("Filter Workspaces By workspace name: '%s'", workspaceName), func() {

		workspacesList := pages.GetWorkspacesPage(webDriver)
		filterID := "Name: " + workspaceName
		searchPage := pages.GetSearchPage(webDriver)
		searchPage.SelectFilter("Name", filterID)
		gomega.Eventually(workspacesList.CountWorkspaces()).Should(gomega.BeNumerically(">=", 1), fmt.Sprintf("The number of workspaces for selected Name:  '%s' should be equa to or greater than 1", workspaceName))
		// Clear the filter
		searchPage.SelectFilter("Name", filterID)
	})
}

func verifySearchWorkspaceByName(workspaceName string) {
	// Search by Workspace Name in the workspaces list.
	ginkgo.By(fmt.Sprintf("And search by Workspace '%s' in the workspaces list", workspaceName), func() {
		WorkspacesPage := pages.GetWorkspacesPage(webDriver)
		workspaceInfo := WorkspacesPage.FindWorkspacInList(workspaceName)
		searchPage := pages.GetSearchPage(webDriver)
		searchPage.SearchName(workspaceName)
		gomega.Eventually(func(g gomega.Gomega) int {
			return pages.CountAppViolations(webDriver)
		}).Should(gomega.BeNumerically(">=", 1), "Search should return '1'  or greater than workspaces in the list")
		gomega.Eventually(workspaceInfo.Name.Text).Should(gomega.Equal(workspaceName), "Failed to get the workspace by its name Value in the Workspaces List")

		// Clear the search result
		gomega.Eventually(searchPage.ClearAllBtn.Click).Should(gomega.Succeed(), "Failed to clear the search result")
	})
}

var _ = ginkgo.Describe("Multi-Cluster Control Plane Workspaces", ginkgo.Label("ui", "workspaces"), func() {

	ginkgo.BeforeEach(ginkgo.OncePerOrdered, func() {
		gomega.Expect(webDriver.Navigate(testUiUrl)).To(gomega.Succeed())
		if !pages.ElementExist(pages.Navbar(webDriver).Title, 3) {
			loginUser()
		}
	})

	ginkgo.AfterEach(ginkgo.OncePerOrdered, func() {

	})

	ginkgo.Context("[UI] Workspaces can be configured on management cluster", func() {
		var workspacesYaml string
		workspaceName := "dev-team"
		workspaceNamespaces := "dev-system"
		workspaceClusterName := "management"

		ginkgo.JustBeforeEach(func() {
			workspacesYaml = path.Join(testDataPath, "tenancy/multiple-tenant.yaml")
		})

		ginkgo.JustAfterEach(func() {
			deleteWorkspaces("management")
		})

		ginkgo.It("Verify Workspaces can be configured on management cluster and dashboard is updated accordingly", ginkgo.Label("integration", "workspaces"), func() {
			existingWorkspacesCount := getWorkspacesCount()
			// Install workspaces on management cluster
			installWorkspaces("management", workspacesYaml)

			pages.NavigateToPage(webDriver, "Workspaces")
			WorkspacesPage := pages.GetWorkspacesPage(webDriver)

			ginkgo.By("And wait for workspaces to be visibe on the dashboard", func() {
				gomega.Eventually(WorkspacesPage.WorkspaceHeader).Should(matchers.BeVisible())

				logger.Info("Existing number of workspaces int the list is :", existingWorkspacesCount)
				totalWorkspacesCount := existingWorkspacesCount + 2 // should return 2 workspaces 'test-team' and 'dev-team'

				gomega.Eventually(func(g gomega.Gomega) int {
					gomega.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
					time.Sleep(POLL_INTERVAL_1SECONDS)
					return WorkspacesPage.CountWorkspaces()
				}, ASSERTION_5MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Equal(totalWorkspacesCount), fmt.Sprintf("There should be '%d' workspaces in the Workspaces list", totalWorkspacesCount))
			})

			workspaceInfo := WorkspacesPage.FindWorkspacInList(workspaceName)
			ginkgo.By(fmt.Sprintf("And verify '%s' workspace Name", workspaceName), func() {
				gomega.Eventually(workspaceInfo.Name).Should(matchers.MatchText(workspaceName), fmt.Sprintf("Failed to list '%s' workspace in the Workspaces List", workspaceName))
			})

			ginkgo.By(fmt.Sprintf("And verify '%s' workspace Namespaces", workspaceName), func() {
				gomega.Eventually(workspaceInfo.Namespaces).Should(matchers.MatchText(workspaceNamespaces), fmt.Sprintf("Failed to get the expected Namespaces for '%s' workspace ", workspaceName))
			})

			ginkgo.By(fmt.Sprintf("And verify '%s' workspace Cluster", workspaceName), func() {
				gomega.Eventually(workspaceInfo.Cluster).Should(matchers.MatchText(workspaceClusterName), fmt.Sprintf("Failed to get the expected Cluster Name for '%s' workspace", workspaceName))
			})
			verifyFilterWorkspacesByClusterName("management", workspaceName)
			verifySearchWorkspaceByName(workspaceName)

		})
	})

	ginkgo.Context("[UI] Workspaces can be configured on leaf cluster", ginkgo.Label("leaf-workspaces"), func() {

		var mgmtClusterContext string
		var leafClusterContext string
		var leafClusterkubeconfig string
		var clusterBootstrapCopnfig string
		var gitopsCluster string
		var workspacesYaml string
		var err string

		patSecret := "workspace-pat"
		bootstrapLabel := "bootstrap"

		// Just specify the leaf cluster info to create it
		leafCluster := ClusterConfig{
			Type:      "leaf",
			Name:      "workspaces-leaf-cluster-test",
			Namespace: "leaf-system",
		}

		workspaceName := "test-team"
		workspaceNamespaces := "test-kustomization, test-system"
		workspaceClusterName := leafCluster.Name

		ginkgo.JustBeforeEach(func() {

			gomega.Expect(webDriver.Navigate(testUiUrl)).To(gomega.Succeed())
			if !pages.ElementExist(pages.Navbar(webDriver).Title, 3) {
				loginUser()
			}
			workspacesYaml = path.Join(testDataPath, "tenancy/multiple-tenant.yaml")
			mgmtClusterContext, err = runCommandAndReturnStringOutput("kubectl config current-context")
			gomega.Expect(err).Should(gomega.BeNil(), "Failed to get the management clsuter context")

			createCluster("kind", leafCluster.Name, "")
			leafClusterContext, err = runCommandAndReturnStringOutput("kubectl config current-context")
			gomega.Expect(err).Should(gomega.BeNil(), "Failed to get the leaf clsuter context")

			// Add/Install Policy Agent on the leaf cluster
			installPolicyAgent(leafCluster.Name)
		})

		ginkgo.JustAfterEach(func() {
			useClusterContext(mgmtClusterContext)

			deleteSecret([]string{leafClusterkubeconfig, patSecret}, leafCluster.Namespace)
			err := runCommandPassThrough("kubectl", "delete", "-f", clusterBootstrapCopnfig)
			gomega.Expect(err).Should(gomega.BeNil(), "Failed to delete the leaf clsuter bootstrap config")

			err = runCommandPassThrough("kubectl", "delete", "-f", gitopsCluster)
			gomega.Expect(err).Should(gomega.BeNil(), "Failed to delete the gitops cluster")

			// Delete the test workspaces from management cluster
			deleteWorkspaces("management")

			deleteCluster("kind", leafCluster.Name, "")
			cleanGitRepository(path.Join("./clusters", leafCluster.Namespace))
			deleteNamespace([]string{leafCluster.Namespace})

		})

		ginkgo.It("Verify Workspaces can be configured on leaf cluster and dashboard is updated accordingly", ginkgo.Label("integration", "workspaces", "leaf-workspaces"), func() {
			ginkgo.Skip("workspaces created normally on the leaf cluster but doesn't appear in the list because of an issue in the product itself and it needs more investigation from dev team")

			// Create leaf cluster kubeconfig
			leafClusterkubeconfig = createLeafClusterKubeconfig(leafClusterContext, leafCluster.Name, leafCluster.Namespace)

			useClusterContext(mgmtClusterContext)
			existingWorkspacesCount := getWorkspacesCount()

			// Create leaf cluster namespace
			createNamespace([]string{leafCluster.Namespace})
			// Install test workspaces on management cluster
			installWorkspaces("management", workspacesYaml)
			createPATSecret(leafCluster.Namespace, patSecret)
			clusterBootstrapCopnfig = createClusterBootstrapConfig(leafCluster.Name, leafCluster.Namespace, bootstrapLabel, patSecret)
			gitopsCluster = connectGitopsCluster(leafCluster.Name, leafCluster.Namespace, bootstrapLabel, leafClusterkubeconfig)
			createLeafClusterSecret(leafCluster.Namespace, leafClusterkubeconfig)

			waitForLeafClusterAvailability(leafCluster.Name, "Ready")
			addKustomizationBases("leaf", leafCluster.Name, leafCluster.Namespace)

			ginkgo.By(fmt.Sprintf("And I verify %s GitopsCluster is bootstraped)", leafCluster.Name), func() {
				useClusterContext(leafClusterContext)
				verifyFluxControllers(GITOPS_DEFAULT_NAMESPACE)
				waitForGitRepoReady("flux-system", GITOPS_DEFAULT_NAMESPACE)
			})
			useClusterContext(leafClusterContext)
			// Installing test workspaces on leaf cluster after leaf-cluster is bootstrap completely
			installWorkspaces(leafCluster.Name, workspacesYaml)

			useClusterContext(mgmtClusterContext)
			pages.NavigateToPage(webDriver, "Workspaces")
			WorkspacesPage := pages.GetWorkspacesPage(webDriver)

			ginkgo.By("And wait for workspaces to be visibe on the dashboard", func() {
				gomega.Eventually(WorkspacesPage.WorkspaceHeader).Should(matchers.BeVisible())
				logger.Info("Existing number of workspaces int the list is :", existingWorkspacesCount)

				totalWorkspacesCount := existingWorkspacesCount + 4 //Should return 4 workspaces (2 on management cluster + 2 on leaf cluster)
				gomega.Eventually(func(g gomega.Gomega) int {
					gomega.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
					time.Sleep(POLL_INTERVAL_1SECONDS)
					return WorkspacesPage.CountWorkspaces()
				}, ASSERTION_5MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Equal(totalWorkspacesCount), fmt.Sprintf("There should be '%d' workspaces in the Workspaces list", totalWorkspacesCount))
			})

			workspaceInfo := WorkspacesPage.FindWorkspacInList(workspaceName)

			ginkgo.By(fmt.Sprintf("And filter leaf cluster '%s' workspaces", leafCluster.Name), func() {
				filterID := "clusterName: " + leafCluster.Namespace + `/` + leafCluster.Name
				searchPage := pages.GetSearchPage(webDriver)
				searchPage.SelectFilter("cluster", filterID)
			})

			ginkgo.By(fmt.Sprintf("And verify '%s' workspace Name", workspaceName), func() {
				gomega.Eventually(workspaceInfo.Name).Should(matchers.MatchText(workspaceName), fmt.Sprintf("Failed to list '%s' workspace in the Workspaces List", workspaceName))
			})

			ginkgo.By(fmt.Sprintf("And verify '%s' workspace Namespaces", workspaceName), func() {
				gomega.Eventually(workspaceInfo.Namespaces).Should(matchers.MatchText(workspaceNamespaces), fmt.Sprintf("Failed to get the expected Namespaces for '%s' workspace ", workspaceName))
			})

			ginkgo.By(fmt.Sprintf("And verify '%s' workspace Cluster", workspaceName), func() {
				gomega.Eventually(workspaceInfo.Cluster).Should(matchers.MatchText(workspaceClusterName), fmt.Sprintf("Failed to get the expected Cluster Name for '%s' workspace", workspaceName))
			})
			verifyFilterWorkspacesByClusterName(leafCluster.Name, workspaceName)
			verifySearchWorkspaceByName(workspaceName)

		})
	})

})
