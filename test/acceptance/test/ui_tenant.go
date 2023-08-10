package acceptance

import (
	"context"
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

/*
	Tenant tests are meant to be run with oidc user only. The tests treat an oidc user as tenant and add the user to tenant group with restricted permissions.
	Cluster user or wego-admin user have access to all resources which will fail these tests expected results.
*/

func getTenantYamlPath() string {
	return path.Join("/tmp", "generated-tenant.yaml")
}

func renderTenants(tenantDefinition string, gp GitProviderEnv) string {
	// render the tenant file out
	gitRepoURL := fmt.Sprintf("ssh://git@%s/%s/%s", gp.Hostname, gp.Org, gp.Repo)
	contents, err := os.ReadFile(tenantDefinition)
	gomega.Expect(err).To(gomega.BeNil(), "Failed to read GitopsCluster template yaml")
	t := template.Must(template.New(tenantDefinition).Parse(string(contents)))
	input := struct {
		MainRepoURL string
	}{
		MainRepoURL: gitRepoURL,
	}
	path := path.Join("/tmp", "rendered-tenant.yaml")
	f, err := os.Create(path)
	gomega.Expect(err).To(gomega.BeNil(), "Failed to create rendered tenant yaml")

	err = t.Execute(f, input)
	gomega.Expect(err).To(gomega.BeNil(), "Failed to render tenant yaml")

	return path
}

func createTenant(tenantDefinition string, gp GitProviderEnv) {
	tenantYaml := getTenantYamlPath()

	renderedTenantsPath := renderTenants(tenantDefinition, gp)

	// Export tenants resources to output file (required to delete tenant resources after test completion)
	_, stdErr := runGitopsCommand(fmt.Sprintf(`create tenants --from-file %s --export > %s`, renderedTenantsPath, tenantYaml))
	gomega.Expect(stdErr).Should(gomega.BeEmpty(), "gitops create tenant command failed with an error")

	// Create tenant resource using default kubeconfig
	_, stdErr = runCommandAndReturnStringOutput(fmt.Sprintf("kubectl apply -f %s", tenantYaml))
	gomega.Expect(stdErr).Should(gomega.BeEmpty(), "Failed to create tenant resources")
}

func deleteTenants(tenantYamls []string) {
	for _, yaml := range tenantYamls {
		_ = runCommandPassThrough("kubectl", "delete", "-f", yaml)
	}
}

var _ = ginkgo.Describe("Multi-Cluster Control Plane Tenancy", ginkgo.Ordered, ginkgo.Label("ui", "tenant"), func() {

	ginkgo.BeforeEach(ginkgo.OncePerOrdered, func() {
		gomega.Expect(webDriver.Navigate(testUiUrl)).To(gomega.Succeed())

		if !pages.ElementExist(pages.Navbar(webDriver).Title, 3) {
			loginUser()
		}

		// Delete the oidc user default roles/rolebindings because the same user is used as a tenant
		_ = runCommandPassThrough("kubectl", "delete", "-f", path.Join(testDataPath, "rbac/user-role-bindings.yaml"))
	})

	ginkgo.AfterEach(ginkgo.OncePerOrdered, func() {
		// Create the oidc user default roles/rolebindings afte tenant tests completed
		_ = runCommandPassThrough("kubectl", "apply", "-f", path.Join(testDataPath, "rbac/user-role-bindings.yaml"))
	})

	ginkgo.Context("Tenants are configured and can view/create allowed resources", ginkgo.Ordered, ginkgo.Label("application"), func() {
		existingAppCount := 0 // Tenant starts from a clean slate

		mgmtCluster := ClusterConfig{
			Type:      "management",
			Name:      "management",
			Namespace: "",
		}

		ginkgo.JustBeforeEach(func() {

		})

		ginkgo.JustAfterEach(func() {
			gomega.Expect(webDriver.Navigate(testUiUrl)).To(gomega.Succeed())
			if !pages.ElementExist(pages.Navbar(webDriver).Title, 3) {
				loginUser()
			}

			// Wait for the application to be deleted gracefully, needed when the test fails before deleting the application
			pages.NavigateToPage(webDriver, "Applications")
			applicationsPage := pages.GetApplicationsPage(webDriver)
			gomega.Eventually(applicationsPage.CountApplications, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Equal(existingAppCount), fmt.Sprintf("There should be %d application enteries after application(s) deletion", existingAppCount))
		})

		ginkgo.It("Verify tenant can install the kustomization application and dashboard is updated accordingly", func() {
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
				GitRepository:   "https://github.com/stefanprodan/podinfo",
			}

			appEvent := ApplicationEvent{
				Reason:    "ReconciliationSucceeded",
				Message:   "next run in " + podinfo.SyncInterval,
				Component: "kustomize-controller",
				Timestamp: "seconds|minutes|minute ago",
			}

			pullRequest := PullRequest{
				Branch:  "management-kustomization-tenant-apps",
				Title:   "Management Kustomization Tenant Application",
				Message: "Adding management kustomization Tenant applications",
			}

			defer deleteTenants([]string{getTenantYamlPath()})
			createTenant(path.Join(testDataPath, "tenancy", "multiple-tenant.yaml.tpl"), gitProviderEnv)
			copyFluxSystemGitRepo("test-kustomization")

			// Add GitRepository source
			sourceURL := "https://github.com/stefanprodan/podinfo"
			addSource("git", podinfo.Source, podinfo.Namespace, sourceURL, "master", "") // allowed repository

			appKustomization := fmt.Sprintf("./clusters/%s/%s-%s-kustomization.yaml", mgmtCluster.Name, podinfo.Name, podinfo.Namespace)
			defer deleteSource("git", podinfo.Source, podinfo.Namespace, "")
			defer cleanGitRepository(appKustomization)
			defer cleanGitRepository(fmt.Sprintf("./clusters/%s/%s-namespace.yaml", mgmtCluster.Name, podinfo.TargetNamespace))

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
				gomega.Eventually(application.SourceHref.Text).Should(gomega.MatchRegexp(sourceURL), "Failed to find the source href")
				gomega.Eventually(application.SelectListItem(webDriver, podinfo.GitRepository).Click).Should(gomega.Succeed(), "Failed to select GitRepository source from git repos list")
			})

			ginkgo.By(fmt.Sprintf("And select %s GitRepository for PR", podinfo.GitRepository), func() {
				gomega.Eventually(func(g gomega.Gomega) bool {
					g.Eventually(application.GitRepository.Click).Should(gomega.Succeed(), "Failed to click Select GitRepository list")
					return pages.ElementExist(application.SelectListItem(webDriver, podinfo.GitRepository))
				}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.BeTrue(), fmt.Sprintf("GitRepository %s is not listed in the gitrepo list", podinfo.GitRepository))
				gomega.Eventually(application.SelectListItem(webDriver, podinfo.GitRepository).Click).Should(gomega.Succeed(), "Failed to select GitRepository source from git repos list")
			})

			AddKustomizationApp(application, podinfo)
			_ = createGitopsPR(pullRequest)

			ginkgo.By("Then I should merge the pull request to start application reconciliation", func() {
				createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
				mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
			})

			ginkgo.By("Then force reconcile flux-system to immediately start application provisioning", func() {
				reconcile("reconcile", "source", "git", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
				reconcile("reconcile", "", "kustomization", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
			})

			ginkgo.By(fmt.Sprintf("And wait for %s application to be visibe on the dashboard", podinfo.Name), func() {
				gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())

				totalAppCount := existingAppCount + 1
				gomega.Eventually(applicationsPage.CountApplications, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(totalAppCount), fmt.Sprintf("There should be %d application enteries in application table", totalAppCount))
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

		ginkgo.It("Verify tenant can install the helmrelease application and dashboard is updated accordingly", func() {
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

			defer deleteTenants([]string{getTenantYamlPath()})
			createTenant(path.Join(testDataPath, "tenancy", "multiple-tenant.yaml.tpl"), gitProviderEnv)
			copyFluxSystemGitRepo("test-kustomization")

			// Add HelmRepository source
			sourceURL := "https://raw.githubusercontent.com/weaveworks/profiles-catalog/gh-pages"
			addSource("helm", metallb.Chart, metallb.Namespace, sourceURL, "", "") // allowed helm repository

			appKustomization := fmt.Sprintf("./clusters/%s/%s-%s-helmrelease.yaml", mgmtCluster.Name, metallb.Name, tenantNamespace)
			repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
			defer deleteSource("helm", metallb.Chart, tenantNamespace, "")
			defer cleanGitRepository(appKustomization)

			ginkgo.By("And wait for cluster-service to cache profiles", func() {
				gomega.Expect(waitForGitopsResources(context.Background(), Request{Path: `charts/list?repository.name=weaveworks-charts&repository.namespace=flux-system&repository.cluster.name=management`}, POLL_INTERVAL_5SECONDS, ASSERTION_15MINUTE_TIME_OUT)).To(gomega.Succeed(), "Failed to get a successful response from /v1/charts")
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
			_ = createGitopsPR(pullRequest)

			ginkgo.By("Then I should merge the pull request to start application reconciliation", func() {
				createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
				mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
			})

			ginkgo.By("Then force reconcile flux-system to immediately start application provisioning", func() {
				reconcile("reconcile", "source", "git", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
				reconcile("reconcile", "", "kustomization", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
			})

			ginkgo.By(fmt.Sprintf("And wait for %s application to be visibe on the dashboard", metallb.Name), func() {
				gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())

				totalAppCount := existingAppCount + 1

				gomega.Eventually(applicationsPage.CountApplications, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(totalAppCount), fmt.Sprintf("There should be %d application enteries in application table", totalAppCount))
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

	ginkgo.Context("Tenants are configured and can view/create allowed resources on leaf cluster", ginkgo.Ordered, ginkgo.Label("kind-leaf-cluster"), func() {
		var mgmtClusterContext string
		var leafClusterContext string
		var leafClusterkubeconfig string
		var clusterBootstrapCopnfig string
		var gitopsCluster string
		existingAppCount := 0
		patSecret := "application-pat"
		bootstrapLabel := "bootstrap"

		leafClusterNamespace := "leaf-system"
		appNameSpace := "test-system"
		appTargetNamespace := "test-system"

		leafCluster := ClusterConfig{
			Type:      "other",
			Name:      "wge-leaf-tenant-kind",
			Namespace: leafClusterNamespace,
		}

		ginkgo.JustBeforeEach(func() {
			gomega.Expect(webDriver.Navigate(testUiUrl)).To(gomega.Succeed())
			if !pages.ElementExist(pages.Navbar(webDriver).Title, 3) {
				loginUser()
			}

			createNamespace([]string{leafClusterNamespace})
			mgmtClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
			createCluster("kind", leafCluster.Name, "")
			leafClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
		})

		ginkgo.JustAfterEach(func() {
			useClusterContext(mgmtClusterContext)

			deleteSecret([]string{leafClusterkubeconfig, patSecret}, leafCluster.Namespace)
			_ = runCommandPassThrough("kubectl", "delete", "-f", clusterBootstrapCopnfig)
			_ = runCommandPassThrough("kubectl", "delete", "-f", gitopsCluster)

			deleteCluster("kind", leafCluster.Name, "")
			cleanGitRepository(path.Join("./clusters", leafCluster.Namespace))
			deleteNamespace([]string{leafCluster.Namespace})
		})

		ginkgo.It("Verify tenant can install the kustomization application from GitRepository source on leaf cluster and management dashboard is updated accordingly", ginkgo.Label("smoke", "application"), func() {
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
				Branch:  "management-kustomization-leaf-cluster-tenant-apps",
				Title:   "Management Kustomization Leaf Cluster Tenant Application",
				Message: "Adding management kustomization leaf cluster applications",
			}

			sourceURL := "https://github.com/stefanprodan/podinfo"
			appKustomization := fmt.Sprintf("./clusters/%s/%s/%s-%s-kustomization.yaml", leafCluster.Namespace, leafCluster.Name, podinfo.Name, podinfo.Namespace)

			repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
			leafClusterkubeconfig = createLeafClusterKubeconfig(leafClusterContext, leafCluster.Name, leafCluster.Namespace)
			// Installing policy-agent to leaf cluster
			installPolicyAgent(leafCluster.Name)

			useClusterContext(mgmtClusterContext)
			// Installing tenant resources to management cluster. This is an easy way to add oidc tenant user rbac
			createTenant(path.Join(testDataPath, "tenancy", "multiple-tenant.yaml.tpl"), gitProviderEnv)
			copyFluxSystemGitRepo("test-kustomization")
			createPATSecret(leafCluster.Namespace, patSecret)
			clusterBootstrapCopnfig = createClusterBootstrapConfig(leafCluster.Name, leafCluster.Namespace, bootstrapLabel, patSecret)
			gitopsCluster = connectGitopsCluster(leafCluster.Name, leafCluster.Namespace, bootstrapLabel, leafClusterkubeconfig)
			createLeafClusterSecret(leafCluster.Namespace, leafClusterkubeconfig)

			useClusterContext(leafClusterContext)
			ginkgo.By(fmt.Sprintf("And I verify %s GitopsCluster/leafCluster is bootstraped)", leafCluster.Name), func() {
				verifyFluxControllers(GITOPS_DEFAULT_NAMESPACE)
				waitForGitRepoReady("flux-system", GITOPS_DEFAULT_NAMESPACE)
			})

			// Installing tenant resources to leaf cluster after leaf-cluster is bootstrapped
			createTenant(path.Join(testDataPath, "tenancy", "multiple-tenant.yaml.tpl"), gitProviderEnv)
			copyFluxSystemGitRepo("test-kustomization")
			// Add GitRepository source to leaf cluster
			addSource("git", podinfo.Source, podinfo.Namespace, sourceURL, "master", "")

			useClusterContext(mgmtClusterContext)
			pages.NavigateToPage(webDriver, "Applications")
			applicationsPage := pages.GetApplicationsPage(webDriver)

			ginkgo.By("And wait for existing applications to be visibe on the dashboard", func() {
				gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())
			})

			ginkgo.By(`And navigate to 'Add Application' page`, func() {
				gomega.Eventually(applicationsPage.AddApplication.Click).Should(gomega.Succeed(), "Failed to click 'Add application' button")

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
			_ = createGitopsPR(pullRequest)

			ginkgo.By("Then I should merge the pull request to start application reconciliation", func() {
				createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
				mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
			})

			ginkgo.By("Then force reconcile leaf cluster flux-system for immediate application availability", func() {
				useClusterContext(leafClusterContext)
				reconcile("reconcile", "source", "git", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
				reconcile("reconcile", "", "kustomization", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
				useClusterContext(mgmtClusterContext)
			})

			ginkgo.By("And wait for leaf cluster podinfo application to be visibe on the dashboard", func() {
				gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())

				totalAppCount := existingAppCount + 1 // podinfo (leaf cluster)
				gomega.Eventually(applicationsPage.CountApplications, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(totalAppCount), fmt.Sprintf("There should be %d application enteries in application table", totalAppCount))
			})

			ginkgo.By(fmt.Sprintf("And search leaf cluster '%s' app", leafCluster.Name), func() {
				searchPage := pages.GetSearchPage(webDriver)
				searchPage.SearchName(podinfo.Name)
				gomega.Eventually(applicationsPage.CountApplications).Should(gomega.Equal(1), "There should be '1' application entery in application table after search")
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
