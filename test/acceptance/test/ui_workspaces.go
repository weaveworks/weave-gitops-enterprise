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

	ginkgo.By(fmt.Sprintf("Add workspaces to the '%s' cluster", clusterName), func() {
		createTenant(workspacesYaml, gitProviderEnv)
	})
}

func deleteWorkspaces(clusterName string) {

	ginkgo.By(fmt.Sprintf("And Finally delete workspaces from '%s' cluster", clusterName), func() {
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
		gomega.Eventually(workspacesList.CountWorkspaces()).Should(gomega.BeNumerically(">=", 1), fmt.Sprintf("The number of workspaces for selected Name:  '%s' should be equal to or greater than 1", workspaceName))
		// Clear the filter
		searchPage.SelectFilter("Name", filterID)
	})
}

func verifySearchWorkspaceByName(workspaceName string) {
	// Search by Workspace Name in the workspaces list.
	ginkgo.By(fmt.Sprintf("And search by Workspace '%s' in the workspaces list", workspaceName), func() {
		WorkspacesPage := pages.GetWorkspacesPage(webDriver)
		workspaceInfo := WorkspacesPage.FindWorkspaceInList(workspaceName)
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

func verifyWorkspaceDetailsPage(workspaceName string, workspaceNamespaces string, workspacesDetailPage *pages.WorkspaceDetailsPage) {

	// workspacesDetailPage := pages.GetWorkspaceDetailsPage(webDriver)
	ginkgo.By(fmt.Sprintf("Then verify '%s' workspace details page", workspaceName), func() {
		gomega.Eventually(workspacesDetailPage.Header.Text).Should(gomega.MatchRegexp(workspaceName), fmt.Sprintf("Failed to verify get the details page's header for '%s' workspace", workspaceName))
		gomega.Eventually(workspacesDetailPage.GoToTenantApplicationsBtn).Should(matchers.BeEnabled(), fmt.Sprintf("'Go To Tenant Applications' button is not visible/enable for workspace", workspaceName))
		gomega.Eventually(workspacesDetailPage.WorkspaceName.Text).Should(gomega.MatchRegexp(workspaceName), fmt.Sprintf("Failed to verify the '%s' workspace tenant name", workspaceName))
		gomega.Eventually(workspacesDetailPage.Namespaces.Text).Should(gomega.MatchRegexp(workspaceNamespaces), fmt.Sprintf("Failed to verify the '%s' workspace namespaces", workspaceName))
		gomega.Eventually(workspacesDetailPage.ServiceAccountsTab).Should(matchers.BeEnabled(), fmt.Sprintf("'Service Accounts' tab is not visible/enable for '%s' workspace", workspaceName))
		gomega.Eventually(workspacesDetailPage.RolesTab).Should(matchers.BeEnabled(), fmt.Sprintf("'Roles' tab is not visible/enable for  '%s' workspace", workspaceName))
		gomega.Eventually(workspacesDetailPage.RoleBindingsTab).Should(matchers.BeEnabled(), fmt.Sprintf("'Role Bindings' tab is not visible/enable for  '%s' workspace", workspaceName))
		gomega.Eventually(workspacesDetailPage.PoliciesTab).Should(matchers.BeEnabled(), fmt.Sprintf("'Policies' tab is not visible/enable for  '%s' workspace", workspaceName))
	})
}

func verifyWrokspaceServiceAccounts(workspaceName string, workspaceNamespaces string, workspacesDetailPage *pages.WorkspaceDetailsPage) {
	// workspacesDetailPage := pages.GetWorkspaceDetailsPage(webDriver)

	ginkgo.By(fmt.Sprintf("After that verify '%s' workspace Service Accounts", workspaceName), func() {
		gomega.Expect(workspacesDetailPage.ServiceAccountsTab.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to open '%s' workspace's Service Accounts tab", workspaceName))
		pages.WaitForPageToLoad(webDriver)

		serviceAccounts := pages.GetWorkspaceServiceAccounts(webDriver)

		gomega.Eventually(serviceAccounts.Name.Text).ShouldNot(gomega.BeEmpty(), fmt.Sprintf("Failed to verify '%s' workspace Service Account's Name", workspaceName))
		gomega.Expect(serviceAccounts.Name.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to open '%s' workspace's Service Accounts Name", workspaceName))
		gomega.Eventually(serviceAccounts.Namespace.Text).Should(gomega.MatchRegexp(workspaceNamespaces), fmt.Sprintf("Failed to verify '%s' workspace Service Account's Namespaces", workspaceName))
		gomega.Eventually(serviceAccounts.Age.Text).ShouldNot(gomega.BeEmpty(), fmt.Sprintf("Failed to verify '%s' workspace Service Account's Age", workspaceName))

	})
}

// func verifyWrokspaceRoles(workspaceName string, workspaceNamespaces string, workspacesDetailPage *pages.WorkspaceDetailsPage) {
// 	// workspacesDetailPage := pages.GetWorkspaceDetailsPage(webDriver)

// 	ginkgo.By(fmt.Sprintf("After that verify '%s' workspace Roles", workspaceName), func() {
// 		gomega.Expect(workspacesDetailPage.RolesTab.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to open '%s' workspace's Roles tab", workspaceName))
// 		pages.WaitForPageToLoad(webDriver)

// 		role := pages.GetWorkspaceRoles(webDriver)

// 		gomega.Eventually(role.Name.Text).ShouldNot(gomega.BeEmpty(), fmt.Sprintf("Failed to verify '%s' workspace Role 's Name", workspaceName))
// 		gomega.Expect(role.Name.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to open '%s' workspace's Roles Name", workspaceName))
// 		gomega.Eventually(role.Age.Text).ShouldNot(gomega.BeEmpty(), fmt.Sprintf("Failed to verify '%s' workspace Roles's Age", workspaceName))
// 		gomega.Eventually(role.Namespace.Text).Should(gomega.MatchRegexp(workspaceNamespaces), fmt.Sprintf("Failed to verify '%s' workspace Roles's Namespaces", workspaceName))
// 		gomega.Eventually(role.Rules.Text).ShouldNot(gomega.BeEmpty(), fmt.Sprintf("Failed to verify '%s' workspace Role's Rules", workspaceName))
// 		gomega.Eventually(role.Rules.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to open '%s' workspace's view rules button", workspaceName))
// 		gomega.Eventually(role.ViewRules.Text).Should(gomega.Equal("Rules"), "Failed to view rules")
// 		gomega.Expect(role.CloseBtn.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to close '%s' workspace's view rules button", workspaceName))
// 		gomega.Eventually(role.Age.Text).ShouldNot(gomega.BeEmpty(), fmt.Sprintf("Failed to verify '%s' workspace Roles's Age", workspaceName))
// 	})
// }

func verifyWrokspaceRoleBindings(workspaceName string, workspaceNamespaces string, workspacesDetailPage *pages.WorkspaceDetailsPage) {
	// workspacesDetailPage := pages.GetWorkspaceDetailsPage(webDriver)

	ginkgo.By(fmt.Sprintf("After that verify '%s' workspace Role Bindings", workspaceName), func() {
		gomega.Expect(workspacesDetailPage.RoleBindingsTab.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to open '%s' workspace's Roles Bindings tab", workspaceName))
		pages.WaitForPageToLoad(webDriver)

		roleBindings := pages.GetWorkspaceRoleBindings(webDriver)

		gomega.Eventually(roleBindings.Name.Text).Should(gomega.MatchRegexp(workspaceName), fmt.Sprintf("Failed to verify '%s' workspace Role Bindings's Namespaces", workspaceName))
		gomega.Expect(roleBindings.Name.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to open '%s' workspace's Roles Bindings Name", workspaceName))
		gomega.Eventually(roleBindings.RoleBindingApi.Text).Should(gomega.Equal("apiVersion"), "Failed to verify Role Bindings Manifest's apiVersion ")
		gomega.Expect(roleBindings.ManifestCloseBtn.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to Close '%s' workspace's Roles Bindings manifest", workspaceName))
		gomega.Eventually(roleBindings.Namespace.Text).Should(gomega.MatchRegexp(workspaceNamespaces), fmt.Sprintf("Failed to verify '%s' workspace Role Bindings's Namespaces", workspaceName))
		gomega.Eventually(roleBindings.Bindings.Text).ShouldNot(gomega.BeEmpty(), fmt.Sprintf("Failed to verify '%s' workspace Role Bindings's Bindings", workspaceName))
		gomega.Eventually(roleBindings.Role.Text).ShouldNot(gomega.BeEmpty(), fmt.Sprintf("Failed to verify '%s' workspace Role Bindings's Role", workspaceName))
		gomega.Eventually(roleBindings.Age.Text).ShouldNot(gomega.BeEmpty(), fmt.Sprintf("Failed to verify '%s' workspace Role Bindings's Age", workspaceName))
	})
}

func verifyWrokspacePolicies(workspaceName string, workspaceNamespaces string, workspacesDetailPage *pages.WorkspaceDetailsPage) {
	// workspacesDetailPage := pages.GetWorkspaceDetailsPage(webDriver)

	ginkgo.By(fmt.Sprintf("After that verify '%s' workspace Policies", workspaceName), func() {
		gomega.Expect(workspacesDetailPage.PoliciesTab.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to open '%s' workspace's Policies tab", workspaceName))
		pages.WaitForPageToLoad(webDriver)

		policies := pages.GetWorkspacePolicies(webDriver)

		gomega.Eventually(policies.Name.Text).Should(gomega.MatchRegexp(workspaceName), fmt.Sprintf("Failed to verify '%s' Policies Name's Namespaces", workspaceName))
		gomega.Expect(policies.Name.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to open '%s' workspace's Policies Name", workspaceName))

		// Navigate back to the policies list
		gomega.Expect(webDriver.Back()).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to navigate back to the '%s' policies list", workspaceName))

		gomega.Eventually(policies.Category.Text).ShouldNot(gomega.BeEmpty(), fmt.Sprintf("Failed to verify '%s' workspace Policies's Category", workspaceName))
		gomega.Eventually(policies.Severity.Text).ShouldNot(gomega.BeEmpty(), fmt.Sprintf("Failed to verify '%s' workspace Policies's Severity", workspaceName))
		gomega.Eventually(policies.Age.Text).ShouldNot(gomega.BeEmpty(), fmt.Sprintf("Failed to verify '%s' workspace Policies's Age", workspaceName))
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

	ginkgo.Context("Workspaces can be configured on management cluster", func() {
		var workspacesYaml string
		workspaceName := "dev-team"
		workspaceNamespaces := "dev-system"
		workspaceClusterName := "management"

		ginkgo.JustBeforeEach(func() {
			workspacesYaml = path.Join(testDataPath, "tenancy/multiple-tenant.yaml.tpl")
		})

		ginkgo.JustAfterEach(func() {
			deleteWorkspaces("management")
		})
		ginkgo.FIt("Verify Workspaces can be configured on management cluster and dashboard is updated accordingly", func() {
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

			workspaceInfo := WorkspacesPage.FindWorkspaceInList(workspaceName)
			workspacesDetailPage := pages.GetWorkspaceDetailsPage(webDriver)

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

			ginkgo.By(fmt.Sprintf("And navigate to '%s' workspace details page", workspaceName), func() {
				gomega.Eventually(workspaceInfo.Name.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to '%s' workspace details page", workspaceName))
			})
			verifyWorkspaceDetailsPage(workspaceName, workspaceNamespaces, workspacesDetailPage)
			verifyWrokspaceServiceAccounts(workspaceName, workspaceNamespaces, workspacesDetailPage)
			// verifyWrokspaceRoles(workspaceName, workspaceNamespaces, workspacesDetailPage)
			verifyWrokspaceRoleBindings(workspaceName, workspaceNamespaces, workspacesDetailPage)
			verifyWrokspacePolicies(workspaceName, workspaceNamespaces, workspacesDetailPage)

		})

	})

	ginkgo.Context("Workspaces can be configured on leaf cluster", ginkgo.Label("kind-leaf-cluster"), func() {

		var mgmtClusterContext string
		var leafClusterContext string
		var leafClusterkubeconfig string
		var clusterBootstrapCopnfig string
		var gitopsCluster string
		var workspacesYaml string
		var errStd string

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
			workspacesYaml = path.Join(testDataPath, "tenancy/multiple-tenant.yaml")
			mgmtClusterContext, errStd = runCommandAndReturnStringOutput("kubectl config current-context")
			gomega.Expect(errStd).Should(gomega.BeEmpty(), "Failed to get the management cluster context")

			createCluster("kind", leafCluster.Name, "")
			leafClusterContext, errStd = runCommandAndReturnStringOutput("kubectl config current-context")
			gomega.Expect(errStd).Should(gomega.BeEmpty(), "Failed to get the leaf cluster context")
		})

		ginkgo.JustAfterEach(func() {
			useClusterContext(mgmtClusterContext)

			deleteSecret([]string{leafClusterkubeconfig, patSecret}, leafCluster.Namespace)
			// Ignore error checking as the premature test failure may have resources not created and checking for the error in cleanup produce unncessary confusion
			_ = runCommandPassThrough("kubectl", "delete", "-f", clusterBootstrapCopnfig)
			_ = runCommandPassThrough("kubectl", "delete", "-f", gitopsCluster)

			// Delete the test workspaces from management cluster
			deleteWorkspaces("management")

			deleteCluster("kind", leafCluster.Name, "")
			cleanGitRepository(path.Join("./clusters", leafCluster.Namespace))
			deleteNamespace([]string{leafCluster.Namespace})

		})

		ginkgo.It("Verify Workspaces can be configured on leaf cluster and dashboard is updated accordingly", func() {
			// Add/Install Policy Agent on the leaf cluster
			installPolicyAgent(leafCluster.Name)
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

			useClusterContext(leafClusterContext)
			ginkgo.By(fmt.Sprintf("And I verify %s GitopsCluster is bootstraped)", leafCluster.Name), func() {
				verifyFluxControllers(GITOPS_DEFAULT_NAMESPACE)
				waitForGitRepoReady("flux-system", GITOPS_DEFAULT_NAMESPACE)
			})
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

			workspaceInfo := WorkspacesPage.FindWorkspaceInList(workspaceName)

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
