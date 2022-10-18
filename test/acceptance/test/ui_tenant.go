package acceptance

import (
	"context"
	"fmt"
	"path"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

func createTenant(tenatDefination string) string {
	tenantYaml := path.Join("/tmp", "generated-tenant.yaml")

	// Export tenants resources to output file (required to delete tenant resources after test completion)
	_, stdErr := runGitopsCommand(fmt.Sprintf(`create tenants --from-file %s --export > %s`, tenatDefination, tenantYaml))
	gomega.Expect(stdErr).Should(gomega.BeEmpty(), "gitops create tenant command failed with an error")

	// Create tenant resource using default kubeconfig
	_, stdErr = runCommandAndReturnStringOutput(fmt.Sprintf("kubectl apply -f %s", tenantYaml))
	gomega.Expect(stdErr).Should(gomega.BeEmpty(), "Failed to create tenant resources")

	return tenantYaml
}

func deleteTenants(tenantYamls []string) {
	for _, yaml := range tenantYamls {
		_ = runCommandPassThrough("kubectl", "delete", "-f", yaml)
	}
}

func DescribeTenants(gitopsTestRunner GitopsTestRunner) {
	var _ = ginkgo.Describe("Multi-Cluster Control Plane Tenancy", ginkgo.Ordered, func() {

		ginkgo.BeforeEach(ginkgo.OncePerOrdered, func() {
			// Delete the oidc user default roles/rolebindings because the same user is used as a tenant
			_ = runCommandPassThrough("kubectl", "delete", "-f", path.Join(getCheckoutRepoPath(), "test", "utils", "data", "user-role-bindings.yaml"))
		})

		ginkgo.AfterEach(ginkgo.OncePerOrdered, func() {
			// Create the oidc user default roles/rolebindings afte tenant tests completed
			_ = runCommandPassThrough("kubectl", "apply", "-f", path.Join(getCheckoutRepoPath(), "test", "utils", "data", "user-role-bindings.yaml"))
		})

		ginkgo.Context("[UI] Tenants are configured and can view/create allowed resources", ginkgo.Ordered, func() {
			existingAppCount := 0 // Tenant starts from a clean slate

			mgmtCluster := ClusterConfig{
				Type:      "management",
				Name:      "management",
				Namespace: "",
			}

			ginkgo.JustBeforeEach(func() {

			})

			ginkgo.JustAfterEach(func() {
				gomega.Expect(webDriver.Navigate(test_ui_url)).To(gomega.Succeed())
				if !pages.ElementExist(pages.Navbar(webDriver).Title, 3) {
					loginUser()
				}

				// Wait for the application to be deleted gracefully, needed when the test fails before deleting the application
				pages.NavigateToPage(webDriver, "Applications")
				applicationsPage := pages.GetApplicationsPage(webDriver)
				gomega.Eventually(func(g gomega.Gomega) int {
					return applicationsPage.CountApplications()
				}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Equal(existingAppCount), fmt.Sprintf("There should be %d application enteries after application(s) deletion", existingAppCount))
			})

			ginkgo.It("Verify tenant can install the kustomization application and dashboard is updated accordingly", ginkgo.Label("integration", "tenant", "application"), func() {
				podinfo := Application{
					Type:            "kustomization",
					Name:            "my-podinfo",
					DeploymentName:  "podinfo",
					Namespace:       "test-kustomization",
					TargetNamespace: "test-system",
					Source:          "my-podinfo",
					Path:            "./kustomize",
					SyncInterval:    "10m",
					Tenant:          "test-team",
				}

				appEvent := ApplicationEvent{
					Reason:    "ReconciliationSucceeded",
					Message:   "next run in " + podinfo.SyncInterval,
					Component: "kustomize-controller",
					Timestamp: "seconds|minutes|minute ago",
				}

				pullRequest := PullRequest{
					Branch:  "management-kustomization-apps",
					Title:   "Management Kustomization Application",
					Message: "Adding management kustomization applications",
				}

				tenantYaml := createTenant(path.Join(getCheckoutRepoPath(), "test", "utils", "data", "tenancy", "multiple-tenant.yaml"))
				defer deleteTenants([]string{tenantYaml})

				// Add GitRepository source
				sourceURL := "https://github.com/stefanprodan/podinfo"
				addSource("git", podinfo.Source, podinfo.Namespace, sourceURL, "master", "") // allowed repository

				appKustomization := fmt.Sprintf("./clusters/%s/%s-%s-kustomization.yaml", mgmtCluster.Name, podinfo.Name, podinfo.Namespace)
				defer deleteSource("git", podinfo.Source, podinfo.Namespace, "")
				defer cleanGitRepository(appKustomization)

				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
				pages.NavigateToPage(webDriver, "Applications")
				applicationsPage := pages.GetApplicationsPage(webDriver)

				ginkgo.By(`And navigate to 'Add Application' page`, func() {
					gomega.Expect(applicationsPage.AddApplication.Click()).Should(gomega.Succeed(), "Failed to click 'Add application' button")

					addApplication := pages.GetAddApplicationsPage(webDriver)
					gomega.Eventually(addApplication.ApplicationHeader.Text).Should(gomega.MatchRegexp("Applications"))
				})

				application := pages.GetAddApplication(webDriver)
				ginkgo.By(fmt.Sprintf("And select %s GitRepository", podinfo.Source), func() {
					gomega.Eventually(func(g gomega.Gomega) bool {
						g.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
						g.Eventually(application.Cluster.Click).Should(gomega.Succeed(), "Failed to click Select Cluster list")
						g.Eventually(application.SelectListItem(webDriver, mgmtCluster.Name).Click).Should(gomega.Succeed(), "Failed to select 'management' cluster from clusters list")
						g.Eventually(application.Source.Click).Should(gomega.Succeed(), "Failed to click Select Source list")
						return pages.ElementExist(application.SelectListItem(webDriver, podinfo.Source))
					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.BeTrue(), fmt.Sprintf("GitRepository %s source is not listed in source's list", podinfo.Source))

					gomega.Eventually(application.SelectListItem(webDriver, podinfo.Source).Click).Should(gomega.Succeed(), "Failed to select GitRepository source from sources list")
					gomega.Eventually(application.SourceHref.Text).Should(gomega.MatchRegexp(sourceURL), "Failed to find the source href")
				})

				AddKustomizationApp(application, podinfo)
				createGitopsPR(pullRequest)

				ginkgo.By("Then I should see see a toast with a link to the creation PR", func() {
					gitops := pages.GetGitOps(webDriver)
					gomega.Eventually(gitops.PRLinkBar, ASSERTION_1MINUTE_TIME_OUT).Should(matchers.BeFound(), "Failed to find Create PR toast")
				})

				ginkgo.By("Then I should merge the pull request to start cluster provisioning", func() {
					createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
					mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
				})

				ginkgo.By(fmt.Sprintf("And wait for %s application to be visibe on the dashboard", podinfo.Name), func() {
					gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())

					totalAppCount := existingAppCount + 1
					// gomega.Eventually(func(g gomega.Gomega) int {
					// 	return getAppHeaderCount(applicationsPage)
					// }, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Equal(totalAppCount), fmt.Sprintf("Dashboard failed to update with expected applications count: %d", totalAppCount))

					gomega.Eventually(func(g gomega.Gomega) int {
						return applicationsPage.CountApplications()
					}, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(totalAppCount), fmt.Sprintf("There should be %d application enteries in application table", totalAppCount))
				})

				verifyAppInformation(applicationsPage, podinfo, mgmtCluster, "Ready")

				applicationInfo := applicationsPage.FindApplicationInList(podinfo.Name)
				ginkgo.By(fmt.Sprintf("And navigate to %s application page", podinfo.Name), func() {
					gomega.Eventually(applicationInfo.Name.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to %s application detail page", podinfo.Name))
				})

				verifyAppPage(podinfo)
				verifyAppEvents(podinfo, appEvent)
				// verifyAppDetails(podinfo, mgmtCluster)
				// verfifyAppGraph(podinfo)

				navigatetoApplicationsPage(applicationsPage)
				verifyAppSourcePage(applicationInfo, podinfo)

				verifyDeleteApplication(applicationsPage, existingAppCount, podinfo.Name, appKustomization)
			})

			ginkgo.It("Verify tenant can install the helmrelease application and dashboard is updated accordingly", ginkgo.Label("integration", "tenant", "application"), func() {
				ginkgo.Skip("HelmReleases are always get installed in flux-system, skipping until fixed")
				tenantNamespace := "test-system"

				metallb := Application{
					Type:            "helm_release",
					Chart:           "profiles-catalog",
					SyncInterval:    "10m",
					Name:            "metallb",
					DeploymentName:  "metallb-controller",
					Namespace:       tenantNamespace,
					TargetNamespace: tenantNamespace,
					Source:          tenantNamespace + "-metallb",
					Version:         "0.0.2",
					ValuesRegex:     `namespace: ""`,
					Values:          fmt.Sprintf(`namespace: %s`, tenantNamespace),
				}

				appEvent := ApplicationEvent{
					Reason:    "info",
					Message:   "Helm install succeeded|Helm install has started",
					Component: "helm-controller",
					Timestamp: "seconds|minutes|minute ago",
				}

				pullRequest := PullRequest{
					Branch:  "management-helm-apps",
					Title:   "Management Helm Applications",
					Message: "Adding management helm applications",
				}

				tenantYaml := createTenant(path.Join(getCheckoutRepoPath(), "test", "utils", "data", "tenancy", "multiple-tenant.yaml"))
				defer deleteTenants([]string{tenantYaml})

				// Add HelmRepository source
				sourceURL := "https://raw.githubusercontent.com/weaveworks/profiles-catalog/gh-pages"
				addSource("helm", metallb.Chart, metallb.Namespace, sourceURL, "", "") // allowed helm repository

				appKustomization := fmt.Sprintf("./clusters/%s/%s-%s-helmrelease.yaml", mgmtCluster.Name, metallb.Name, tenantNamespace)
				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
				defer deleteSource("helm", metallb.Chart, tenantNamespace, "")
				defer cleanGitRepository(appKustomization)

				ginkgo.By("And wait for cluster-service to cache profiles", func() {
					gomega.Expect(waitForGitopsResources(context.Background(), "profiles", POLL_INTERVAL_5SECONDS)).To(gomega.Succeed(), "Failed to get a successful response from /v1/profiles ")
				})

				pages.NavigateToPage(webDriver, "Applications")
				applicationsPage := pages.GetApplicationsPage(webDriver)

				ginkgo.By(`And navigate to 'Add Application' page`, func() {
					gomega.Expect(applicationsPage.AddApplication.Click()).Should(gomega.Succeed(), "Failed to click 'Add application' button")

					addApplication := pages.GetAddApplicationsPage(webDriver)
					gomega.Eventually(addApplication.ApplicationHeader.Text).Should(gomega.MatchRegexp("Applications"))
				})

				application := pages.GetAddApplication(webDriver)
				createPage := pages.GetCreateClusterPage(webDriver)
				profile := createPage.GetProfileInList(metallb.Name)
				ginkgo.By(fmt.Sprintf("And select %s HelmRepository", metallb.Chart), func() {
					gomega.Eventually(func(g gomega.Gomega) bool {
						g.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
						g.Eventually(application.Cluster.Click).Should(gomega.Succeed(), "Failed to click Select Cluster list")
						g.Eventually(application.SelectListItem(webDriver, mgmtCluster.Name).Click).Should(gomega.Succeed(), "Failed to select 'management' cluster from clusters list")
						g.Eventually(application.Source.Click).Should(gomega.Succeed(), "Failed to click Select Source list")
						return pages.ElementExist(application.SelectListItem(webDriver, metallb.Chart))
					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.BeTrue(), fmt.Sprintf("HelmRepository %s source is not listed in source's list", metallb.Name))

					gomega.Eventually(application.SelectListItem(webDriver, metallb.Chart).Click).Should(gomega.Succeed(), "Failed to select HelmRepository source from sources list")
					gomega.Eventually(application.SourceHref.Text).Should(gomega.MatchRegexp(sourceURL), "Failed to find the source href")
				})

				AddHelmReleaseApp(profile, metallb)
				createGitopsPR(pullRequest)

				ginkgo.By("Then I should see see a toast with a link to the creation PR", func() {
					gitops := pages.GetGitOps(webDriver)
					gomega.Eventually(gitops.PRLinkBar, ASSERTION_1MINUTE_TIME_OUT).Should(matchers.BeFound(), "Failed to find Create PR toast")
				})

				ginkgo.By("Then I should merge the pull request to start cluster provisioning", func() {
					createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
					mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
				})

				ginkgo.By(fmt.Sprintf("And wait for %s application to be visibe on the dashboard", metallb.Name), func() {
					gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())

					totalAppCount := existingAppCount + 1
					// gomega.Eventually(func(g gomega.Gomega) int {
					// 	return getAppHeaderCount(applicationsPage)
					// }, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Equal(totalAppCount), fmt.Sprintf("Dashboard failed to update with expected applications count: %d", totalAppCount))

					gomega.Eventually(func(g gomega.Gomega) int {
						return applicationsPage.CountApplications()
					}, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(totalAppCount), fmt.Sprintf("There should be %d application enteries in application table", totalAppCount))
				})

				verifyAppInformation(applicationsPage, metallb, mgmtCluster, "Ready")

				applicationInfo := applicationsPage.FindApplicationInList(metallb.Name)
				ginkgo.By(fmt.Sprintf("And navigate to %s application page", metallb.Name), func() {
					gomega.Eventually(applicationInfo.Name.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to %s application detail page", metallb.Name))
				})

				verifyAppPage(metallb)
				verifyAppEvents(metallb, appEvent)
				verifyAppDetails(metallb, mgmtCluster)
				verfifyAppGraph(metallb)

				navigatetoApplicationsPage(applicationsPage)
				verifyAppSourcePage(applicationInfo, metallb)

				verifyDeleteApplication(applicationsPage, existingAppCount, metallb.Name, appKustomization)
			})
		})

		ginkgo.Context("[UI] Tenants are configured and can view/create allowed resources on leaf cluster", func() {
			var mgmtClusterContext string
			var leafClusterContext string
			var leafClusterkubeconfig string
			var clusterBootstrapCopnfig string
			var gitopsCluster string
			var appDir string
			existingAppCount := 0
			patSecret := "application-pat"
			bootstrapLabel := "bootstrap"

			appNameSpace := "test-system"
			appTargetNamespace := "test-system"

			leafCluster := ClusterConfig{
				Type:      "other",
				Name:      "wge-leaf-tenant-kind",
				Namespace: appNameSpace,
			}

			ginkgo.JustBeforeEach(func() {
				gomega.Expect(webDriver.Navigate(test_ui_url)).To(gomega.Succeed())
				if !pages.ElementExist(pages.Navbar(webDriver).Title, 3) {
					loginUser()
				}

				createNamespace([]string{appNameSpace, appTargetNamespace})
				appDir = path.Join("clusters", leafCluster.Namespace, leafCluster.Name, "apps")
				mgmtClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
				createCluster("kind", leafCluster.Name, "")
				leafClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
			})

			ginkgo.JustAfterEach(func() {
				useClusterContext(mgmtClusterContext)

				deleteSecret([]string{leafClusterkubeconfig, patSecret}, leafCluster.Namespace)
				_ = gitopsTestRunner.KubectlDelete([]string{}, clusterBootstrapCopnfig)
				_ = gitopsTestRunner.KubectlDelete([]string{}, gitopsCluster)

				deleteCluster("kind", leafCluster.Name, "")
				cleanGitRepository(appDir)
				deleteNamespace([]string{leafCluster.Namespace})
			})

			ginkgo.It("Verify tenant can install the kustomization application from GitRepository source on leaf cluster and management dashboard is updated accordingly", ginkgo.Label("integration", "tenant", "leaf-application"), func() {
				podinfo := Application{
					Type:            "kustomization",
					Name:            "my-podinfo",
					DeploymentName:  "podinfo",
					Namespace:       appNameSpace,
					TargetNamespace: appTargetNamespace,
					Source:          "my-podinfo",
					Path:            "./kustomize",
					SyncInterval:    "10m",
					Tenant:          "test-team",
				}

				appEvent := ApplicationEvent{
					Reason:    "ReconciliationSucceeded",
					Message:   "next run in " + podinfo.SyncInterval,
					Component: "kustomize-controller",
					Timestamp: "seconds|minutes|minute ago",
				}

				pullRequest := PullRequest{
					Branch:  "management-kustomization-leaf-cluster-apps",
					Title:   "Management Kustomization Leaf Cluster Application",
					Message: "Adding management kustomization leaf cluster applications",
				}

				sourceURL := "https://github.com/stefanprodan/podinfo"
				appKustomization := fmt.Sprintf("./clusters/%s/%s/%s-%s-kustomization.yaml", leafCluster.Namespace, leafCluster.Name, podinfo.Name, podinfo.Namespace)

				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
				leafClusterkubeconfig = createLeafClusterKubeconfig(leafClusterContext, leafCluster.Name, leafCluster.Namespace)

				// Installing policy-agent to leaf cluster
				installPolicyAgent(leafCluster.Name)
				// Installing tenant resources to leaf cluster
				_ = createTenant(path.Join(getCheckoutRepoPath(), "test", "utils", "data", "tenancy", "multiple-tenant.yaml"))

				useClusterContext(mgmtClusterContext)
				createPATSecret(leafCluster.Namespace, patSecret)
				clusterBootstrapCopnfig = createClusterBootstrapConfig(leafCluster.Name, leafCluster.Namespace, bootstrapLabel, patSecret)
				gitopsCluster = connectGitopsCuster(leafCluster.Name, leafCluster.Namespace, bootstrapLabel, leafClusterkubeconfig)
				createLeafClusterSecret(leafCluster.Namespace, leafClusterkubeconfig)

				ginkgo.By(fmt.Sprintf("And I verify %s GitopsCluster/leafCluster is bootstraped)", leafCluster.Name), func() {
					useClusterContext(leafClusterContext)
					verifyFluxControllers(GITOPS_DEFAULT_NAMESPACE)
					waitForGitRepoReady("flux-system", GITOPS_DEFAULT_NAMESPACE)
				})

				// Add GitRepository source to leaf cluster
				addSource("git", podinfo.Source, podinfo.Namespace, sourceURL, "master", "")

				useClusterContext(mgmtClusterContext)

				pages.NavigateToPage(webDriver, "Applications")
				applicationsPage := pages.GetApplicationsPage(webDriver)

				ginkgo.By("And wait for existing applications to be visibe on the dashboard", func() {
					gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())
					// gomega.Eventually(func(g gomega.Gomega) int {
					// 	return getAppHeaderCount(applicationsPage)
					// }, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Equal(existingAppCount), fmt.Sprintf("Dashboard failed to update with existing applications count: %d", existingAppCount))
				})

				ginkgo.By(`And navigate to 'Add Application' page`, func() {
					gomega.Expect(applicationsPage.AddApplication.Click()).Should(gomega.Succeed(), "Failed to click 'Add application' button")

					addApplication := pages.GetAddApplicationsPage(webDriver)
					gomega.Eventually(addApplication.ApplicationHeader.Text).Should(gomega.MatchRegexp("Applications"))
				})

				application := pages.GetAddApplication(webDriver)
				ginkgo.By(fmt.Sprintf("And select %s GitRepository for cluster %s", podinfo.Source, leafCluster.Name), func() {
					gomega.Eventually(func(g gomega.Gomega) bool {
						g.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
						g.Eventually(application.Cluster.Click).Should(gomega.Succeed(), "Failed to click Select Cluster list")
						g.Eventually(application.SelectListItem(webDriver, leafCluster.Name).Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to select %s cluster from clusters list", leafCluster.Name))
						g.Eventually(application.Source.Click).Should(gomega.Succeed(), "Failed to click Select Source list")
						return pages.ElementExist(application.SelectListItem(webDriver, podinfo.Source))
					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.BeTrue(), fmt.Sprintf("GitRepository %s source is not listed in source's list", podinfo.Source))

					gomega.Eventually(application.SelectListItem(webDriver, podinfo.Source).Click).Should(gomega.Succeed(), "Failed to select GitRepository source from sources list")
					gomega.Eventually(application.SourceHref.Text).Should(gomega.MatchRegexp(sourceURL), "Failed to find the source href")
				})

				AddKustomizationApp(application, podinfo)
				createGitopsPR(pullRequest)

				ginkgo.By("Then I should see see a toast with a link to the creation PR", func() {
					gitops := pages.GetGitOps(webDriver)
					gomega.Eventually(gitops.PRLinkBar, ASSERTION_1MINUTE_TIME_OUT).Should(matchers.BeFound(), "Failed to find Create PR toast")
				})

				ginkgo.By("Then I should merge the pull request to start cluster provisioning", func() {
					createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
					mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
				})

				ginkgo.By("And wait for leaf cluster podinfo application to be visibe on the dashboard", func() {
					gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())

					totalAppCount := existingAppCount + 1 // podinfo (leaf cluster)
					// gomega.Eventually(func(g gomega.Gomega) int {
					// 	return getAppHeaderCount(applicationsPage)
					// }, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Equal(totalAppCount), fmt.Sprintf("Dashboard failed to update with expected applications count: %d", totalAppCount))

					gomega.Eventually(func(g gomega.Gomega) int {
						return applicationsPage.CountApplications()
					}, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(totalAppCount), fmt.Sprintf("There should be %d application enteries in application table", totalAppCount))
				})

				ginkgo.By(fmt.Sprintf("And search leaf cluster '%s' app", leafCluster.Name), func() {
					searchPage := pages.GetSearchPage(webDriver)
					gomega.Eventually(searchPage.SearchBtn.Click).Should(gomega.Succeed(), "Failed to click search buttton")
					gomega.Expect(searchPage.Search.SendKeys(podinfo.Name)).Should(gomega.Succeed(), "Failed type application name in search field")
					gomega.Expect(searchPage.Search.SendKeys("\uE007")).Should(gomega.Succeed()) // send enter key code to do application search in table

					gomega.Eventually(func(g gomega.Gomega) int {
						return applicationsPage.CountApplications()
					}).Should(gomega.Equal(1), "There should be '1' application entery in application table after search")
				})

				verifyAppInformation(applicationsPage, podinfo, leafCluster, "Ready")

				applicationInfo := applicationsPage.FindApplicationInList(podinfo.Name)
				ginkgo.By(fmt.Sprintf("And navigate to %s application page", podinfo.Name), func() {
					gomega.Eventually(applicationInfo.Name.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to %s application detail page", podinfo.Name))
				})

				verifyAppPage(podinfo)
				verifyAppEvents(podinfo, appEvent)
				// verifyAppDetails(podinfo, leafCluster)
				// verfifyAppGraph(podinfo)

				navigatetoApplicationsPage(applicationsPage)
				verifyAppSourcePage(applicationInfo, podinfo)

				verifyDeleteApplication(applicationsPage, existingAppCount, podinfo.Name, appKustomization)
			})
		})
	})
}
