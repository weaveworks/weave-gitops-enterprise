package acceptance

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"text/template"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

func createGitKustomization(repoName, nameSpace, repoURL, kustomizationName, targetNamespace string) (kustomization string) {
	contents, err := ioutil.ReadFile(path.Join(getCheckoutRepoPath(), "test", "utils", "data", "git-kustomization.yaml"))
	Expect(err).To(BeNil(), "Failed to read git-kustomization template yaml")

	t := template.Must(template.New("kustomization").Parse(string(contents)))

	type TemplateInput struct {
		GitRepoName       string
		NameSpace         string
		GitRepoURL        string
		KustomizationName string
		TargetNamespace   string
	}
	input := TemplateInput{repoName, nameSpace, repoURL, kustomizationName, targetNamespace}

	kustomization = path.Join("/tmp", kustomizationName+"-kustomization.yaml")

	f, err := os.Create(kustomization)
	Expect(err).To(BeNil(), "Failed to create kustomization manifest yaml")

	err = t.Execute(f, input)
	f.Close()
	Expect(err).To(BeNil(), "Failed to generate kustomization manifest yaml")

	return kustomization
}

func navigatetoApplicationsPage(applicationsPage *pages.ApplicationsPage) {
	By("And navigate to Applicartions page via header link", func() {
		Expect(applicationsPage.ApplicationHeader.Click()).Should(Succeed(), "Failed to navigate to Applications pages via header link")
		pages.WaitForPageToLoad(webDriver)
	})
}

func DescribeApplications(gitopsTestRunner GitopsTestRunner) {
	var _ = Describe("Multi-Cluster Control Plane Applications", func() {

		Context("[UI] When no applications are installed", func() {
			It("Verify management cluster dashboard shows bootstrap 'flux-system' application", Label("integration"), func() {
				existingAppCount := getApplicationCount()

				pages.NavigateToPage(webDriver, "Applications")
				applicationsPage := pages.GetApplicationsPage(webDriver)
				pages.WaitForPageToLoad(webDriver)

				By("And wait for Applications page to be rendered", func() {
					Eventually(applicationsPage.ApplicationHeader).Should(BeVisible())

					Eventually(func(g Gomega) string {
						g.Expect(webDriver.Refresh()).ShouldNot(HaveOccurred())
						time.Sleep(POLL_INTERVAL_1SECONDS)
						count, _ := applicationsPage.ApplicationCount.Text()
						return count

					}, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(MatchRegexp(strconv.Itoa(existingAppCount)), fmt.Sprintf("Dashboard failed to update with expected applications count: %d", existingAppCount))

					Expect(applicationsPage.CountApplications()).To(Equal(1), "There should not be any cluster in cluster table")
				})

				applicationInfo := applicationsPage.FindApplicationInList("flux-system")
				By("And verify bootstrap application Name", func() {
					Eventually(applicationInfo.Name).Should(MatchText("flux-system"), "Failed to list flux-system application in  application table")
				})

				By("And verify bootstrap application Type", func() {
					Eventually(applicationInfo.Type).Should(MatchText("Kustomization"), "Failed to have expected flux-system application type: Kustomization")
				})

				By("And verify bootstrap application Namespace", func() {
					Eventually(applicationInfo.Namespace).Should(MatchText(GITOPS_DEFAULT_NAMESPACE), fmt.Sprintf("Failed to have expected flux-system application namespace: %s", GITOPS_DEFAULT_NAMESPACE))
				})

				By("And verify bootstrap application Cluster", func() {
					Eventually(applicationInfo.Cluster).Should(MatchText("management"), "Failed to have expected flux-system application cluster: management")
				})

				By("And verify bootstrap application Source", func() {
					Eventually(applicationInfo.Source).Should(MatchText("flux-system"), "Failed to have expected flux-system application namespace: flux-system")
				})

				By("And verify bootstrap application status", func() {
					Eventually(applicationInfo.Status, ASSERTION_30SECONDS_TIME_OUT).Should(MatchText("Ready"), "Failed to have expected flux-system application status: Ready")
				})
			})
		})

		Context("[UI] Applications(s) can be installed", func() {
			appDir := "./clusters/my-cluster/podinfo"
			deploymentName := "podinfo"
			appName := "my-podinfo"
			appNameSpace := "test-kustomization"
			appTargetNamespace := "test-systems"
			appSyncInterval := "30s"

			JustAfterEach(func() {
				cleanGitRepository(appDir)
				deleteNamespace([]string{appNameSpace, appTargetNamespace})
			})

			It("Verify application can be installed  and dashboard is updated accordingly", Label("integration", "application", "browser-logs"), func() {
				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
				existingAppCount := getApplicationCount()

				pages.NavigateToPage(webDriver, "Applications")
				applicationsPage := pages.GetApplicationsPage(webDriver)

				By("Create namespaces for Kustomization deployments)", func() {
					createNamespace([]string{appNameSpace, appTargetNamespace})
				})

				By("And add Kustomization & GitRepository Source manifests pointing to podinfo repository’s master branch)", func() {
					kustomizationFile := createGitKustomization(appName, appNameSpace, "https://github.com/stefanprodan/podinfo", appName, appTargetNamespace)
					pullGitRepo(repoAbsolutePath)
					_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("mkdir -p %[2]v && cp -f %[1]v %[2]v", kustomizationFile, path.Join(repoAbsolutePath, appDir)))
					gitUpdateCommitPush(repoAbsolutePath, "Adding podinfo kustomization")
				})

				By("And wait for podinfo application to be visibe on the dashboard", func() {
					Eventually(applicationsPage.ApplicationHeader).Should(BeVisible())

					totalAppCount := existingAppCount + 1
					Eventually(func(g Gomega) string {
						g.Expect(webDriver.Refresh()).ShouldNot(HaveOccurred())
						time.Sleep(POLL_INTERVAL_1SECONDS)
						count, _ := applicationsPage.ApplicationCount.Text()
						return count

					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(MatchRegexp(strconv.Itoa(totalAppCount)), fmt.Sprintf("Dashboard failed to update with expected applications count: %d", totalAppCount))

					Eventually(func(g Gomega) int {
						return applicationsPage.CountApplications()
					}, ASSERTION_3MINUTE_TIME_OUT).Should(Equal(totalAppCount), fmt.Sprintf("There should be %d application enteries in application table", totalAppCount))
				})

				applicationInfo := applicationsPage.FindApplicationInList(appName)
				By("And verify podinfo application Name", func() {
					Eventually(applicationInfo.Name).Should(MatchText(appName), fmt.Sprintf("Failed to list %s application in  application table", appName))
				})

				By("And verify podinfo application Type", func() {
					Eventually(applicationInfo.Type).Should(MatchText("Kustomization"), fmt.Sprintf("Failed to have expected %s application type: Kustomization", appName))
				})

				By("And verify podinfo application Namespace", func() {
					Eventually(applicationInfo.Namespace).Should(MatchText(appNameSpace), fmt.Sprintf("Failed to have expected %s application namespace: %s", appName, appNameSpace))
				})

				By("And verify podinfo application Cluster", func() {
					Eventually(applicationInfo.Cluster).Should(MatchText("management"), fmt.Sprintf("Failed to have expected %s application cluster: management", appName))
				})

				By("And verify podinfo application Source", func() {
					Eventually(applicationInfo.Source).Should(MatchText(appName), fmt.Sprintf("Failed to have expected %[1]v application source: %[1]s", appName))
				})

				By("And verify podinfo application status", func() {
					Eventually(applicationInfo.Status, ASSERTION_1MINUTE_TIME_OUT).Should(MatchText("Ready"), fmt.Sprintf("Failed to have expected %s application status: Ready", appName))
				})

				By(fmt.Sprintf("And navigate to %s application page", appName), func() {
					Eventually(applicationInfo.Name.Click).Should(Succeed(), fmt.Sprintf("Failed to navigate to %s application detail page", appName))
				})

				appDetailPage := pages.GetApplicationsDetailPage(webDriver)
				By(fmt.Sprintf("And verify %s application page", appName), func() {
					Eventually(appDetailPage.Header.Text).Should(MatchRegexp(appName), fmt.Sprintf("Failed to verify dashboard application name %s", appName))
					Eventually(appDetailPage.Title.Text).Should(MatchRegexp(appName), fmt.Sprintf("Failed to verify application title %s on application page", appName))
					Eventually(appDetailPage.Sync).Should(BeEnabled(), fmt.Sprintf("Sync button is not visible/enable for %s", appName))
					Eventually(appDetailPage.Details).Should(BeEnabled(), fmt.Sprintf("Details tab button is not visible/enable for %s", appName))
					Eventually(appDetailPage.Events).Should(BeEnabled(), fmt.Sprintf("Events tab button is not visible/enable for %s", appName))
					Eventually(appDetailPage.Graph).Should(BeEnabled(), fmt.Sprintf("Graph tab button is not visible/enable for %s", appName))
				})

				By(fmt.Sprintf("And verify %s application Details", appName), func() {
					Expect(appDetailPage.Details.Click()).Should(Succeed(), fmt.Sprintf("Failed to click %s Details tab button", appName))
					pages.WaitForPageToLoad(webDriver)

					details := pages.GetApplicationDetail(webDriver)

					Eventually(details.Source.Text).Should(MatchRegexp("GitRepository/"+appName), fmt.Sprintf("Failed to verify %s Source", appName))
					Eventually(details.AppliedRevision.Text).Should(MatchRegexp("master"), fmt.Sprintf("Failed to verify %s AppliedRevision", appName))
					Eventually(details.Cluster.Text).Should(MatchRegexp("management"), fmt.Sprintf("Failed to verify %s Cluster", appName))
					Eventually(details.Path.Text).Should(MatchRegexp("./kustomize"), fmt.Sprintf("Failed to verify %s Path", appName))
					Eventually(details.Interval.Text).Should(MatchRegexp(appSyncInterval), fmt.Sprintf("Failed to verify %s AppliedRevision", appName))

					Eventually(appDetailPage.Sync.Click).Should(Succeed(), fmt.Sprintf("Failed to sync %s kustomization", appName))
					Eventually(details.LastUpdated.Text).Should(MatchRegexp("seconds ago"), fmt.Sprintf("Failed to verify %s LastUpdated", appName))

					Eventually(details.Name.Text).Should(MatchRegexp(deploymentName), fmt.Sprintf("Failed to verify %s Deployment name", appName))
					Eventually(details.Type.Text).Should(MatchRegexp("Deployment"), fmt.Sprintf("Failed to verify %s Type", appName))
					Eventually(details.Namespace.Text).Should(MatchRegexp(appTargetNamespace), fmt.Sprintf("Failed to verify %s Namespace", appName))
					Eventually(details.Status.Text, ASSERTION_2MINUTE_TIME_OUT).Should(MatchRegexp("Ready"), fmt.Sprintf("Failed to verify %s Status", appName))
					Eventually(details.Message.Text).Should(MatchRegexp("Deployment is available"), fmt.Sprintf("Failed to verify %s Message", appName))

					// Verify metadata
					Expect(details.GetMetadata("Description").Text()).Should(MatchRegexp(`Podinfo is a tiny web application made with Go`), "Failed to verify Metada description")
					verifyDashboard(details.GetMetadata("Grafana Dashboard").Find("a"), "management", "Grafana")
					Expect(details.GetMetadata("Javascript Alert").Find("a")).ShouldNot(BeFound(), "Javascript href is not sanitized")
					Expect(details.GetMetadata("Javascript Alert").Text()).Should(MatchRegexp(`javascript:alert\('hello there'\);`), "Failed to verify Javascript alert text")
				})

				By(fmt.Sprintf("And verify %s application Events", appName), func() {
					Expect(appDetailPage.Events.Click()).Should(Succeed(), fmt.Sprintf("Failed to click %s Events tab button", appName))
					pages.WaitForPageToLoad(webDriver)

					event := pages.GetApplicationEvent(webDriver, "ReconciliationSucceeded")

					Eventually(event.Reason.Text).Should(MatchRegexp("ReconciliationSucceeded"), fmt.Sprintf("Failed to verify %s Event/Reason", appName))
					Eventually(event.Message.Text).Should(MatchRegexp("next run in "+appSyncInterval), fmt.Sprintf("Failed to verify %s Event/Message", appName))
					Eventually(event.Component.Text).Should(MatchRegexp("kustomize-controller"), fmt.Sprintf("Failed to verify %s Event/Component", appName))
					Eventually(event.TimeStamp.Text).Should(MatchRegexp("seconds|minutes|minute ago"), fmt.Sprintf("Failed to verify %s Event/Timestamp", appName))
				})

				By(fmt.Sprintf("And verify %s application Grapg", appName), func() {
					Expect(appDetailPage.Graph.Click()).Should(Succeed(), fmt.Sprintf("Failed to click %s Graph tab button", appName))
					pages.WaitForPageToLoad(webDriver)

					graph := pages.GetApplicationGraph(webDriver, deploymentName, appName, appNameSpace, appTargetNamespace)

					Expect(graph.SourceGit).Should(BeFound(), fmt.Sprintf("Failed to verify %s Graph/Source", appName))
					Expect(graph.Kustomization).Should(BeFound(), fmt.Sprintf("Failed to verify %s Graph/Kustomization", appName))
					Expect(graph.Deployment).Should(BeFound(), fmt.Sprintf("Failed to verify %s Graph/Deployment", appName))
					Expect(graph.ReplicaSet).Should(BeFound(), fmt.Sprintf("Failed to verify %s Graph/ReplicaSet", appName))
					Expect(graph.Pod).Should(BeFound(), fmt.Sprintf("Failed to verify %s Graph/Pod", appName))
				})

				navigatetoApplicationsPage(applicationsPage)

				By(fmt.Sprintf("And navigate directly to %s Sources page", appName), func() {
					Expect(applicationInfo.Source.Click()).Should(Succeed(), fmt.Sprintf("Failed to navigate to %s Sources pages directly", appName))

					sourceDetailPage := pages.GetSourceDetailPage(webDriver)
					Eventually(sourceDetailPage.Header.Text).Should(MatchRegexp(appName), fmt.Sprintf("Failed to verify dashboard header source name %s ", appName))
					Eventually(sourceDetailPage.Title.Text).Should(MatchRegexp(appName), fmt.Sprintf("Failed to verify application title %s on source page", appName))
				})

				navigatetoApplicationsPage(applicationsPage)

				By("And delete the podinfo kustomization and source maifest from the repository's master branch", func() {
					cleanGitRepository(appDir)
				})

				By("And wait for podinfo application to dissappeare from the dashboard", func() {
					Eventually(applicationsPage.ApplicationCount, ASSERTION_3MINUTE_TIME_OUT).Should(MatchText(strconv.Itoa(existingAppCount)), fmt.Sprintf("Dashboard failed to update with expected applications count: %d", existingAppCount))
					Eventually(func(g Gomega) int {
						return applicationsPage.CountApplications()
					}, ASSERTION_3MINUTE_TIME_OUT).Should(Equal(existingAppCount), fmt.Sprintf("There should be %d application enteries in application table", existingAppCount))
				})
			})
		})

		Context("[UI] Applications(s) can be installed on leaf cluster", func() {
			var mgmtClusterContext string
			var leafClusterContext string
			var leafClusterkubeconfig string
			var clusterBootstrapCopnfig string
			var gitopsCluster string
			var appDir string
			var existingAppCount int
			patSecret := "application-pat"
			bootstrapLabel := "bootstrap"
			leafClusterName := "wge-leaf-application-kind"
			leafClusterNamespace := "test-system"

			deploymentName := "podinfo"
			appName := "my-podinfo"
			appNameSpace := "test-kustomization"
			appTargetNamespace := "test-system"
			appSyncInterval := "30s"

			JustBeforeEach(func() {
				existingAppCount = getApplicationCount()
				appDir = path.Join("clusters", leafClusterNamespace, leafClusterName, "apps")
				mgmtClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
				createCluster("kind", leafClusterName, "")
				createNamespace([]string{appNameSpace, appTargetNamespace})
				leafClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
			})

			JustAfterEach(func() {
				useClusterContext(mgmtClusterContext)

				deleteSecret([]string{leafClusterkubeconfig, patSecret}, leafClusterNamespace)
				_ = gitopsTestRunner.KubectlDelete([]string{}, clusterBootstrapCopnfig)
				_ = gitopsTestRunner.KubectlDelete([]string{}, gitopsCluster)

				deleteClusters("kind", []string{leafClusterName}, "")
				cleanGitRepository(appDir)
				deleteNamespace([]string{leafClusterNamespace})

			})

			It("Verify application can be installed  on leaf cluster and management dashboard is updated accordingly", Label("integration", "application", "leaf-application"), func() {
				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
				leafClusterkubeconfig = createLeafClusterKubeconfig(leafClusterContext, leafClusterName, leafClusterNamespace)

				useClusterContext(mgmtClusterContext)
				createNamespace([]string{leafClusterNamespace})

				createPATSecret(leafClusterNamespace, patSecret)
				clusterBootstrapCopnfig = createClusterBootstrapConfig(leafClusterName, leafClusterNamespace, bootstrapLabel, patSecret)
				gitopsCluster = connectGitopsCuster(leafClusterName, leafClusterNamespace, bootstrapLabel, leafClusterkubeconfig)
				createLeafClusterSecret(leafClusterNamespace, leafClusterkubeconfig)

				By("Verify GitopsCluster status after creating kubeconfig secret", func() {
					pages.NavigateToPage(webDriver, "Clusters")
					clustersPage := pages.GetClustersPage(webDriver)
					pages.WaitForPageToLoad(webDriver)
					clusterInfo := clustersPage.FindClusterInList(leafClusterName)

					Eventually(clusterInfo.Status, ASSERTION_30SECONDS_TIME_OUT).Should(MatchText("Ready"))
				})

				By("And add kustomization bases for common resources for leaf cluster)", func() {
					addKustomizationBases(leafClusterName, leafClusterNamespace)
				})

				By(fmt.Sprintf("And I verify %s GitopsCluster is bootstraped)", leafClusterName), func() {
					useClusterContext(leafClusterContext)
					verifyFluxControllers(GITOPS_DEFAULT_NAMESPACE)
					waitForGitRepoReady("flux-system", GITOPS_DEFAULT_NAMESPACE)
					useClusterContext(mgmtClusterContext)
				})

				By("And add Kustomization & GitRepository Source manifests pointing to podinfo repository’s master branch)", func() {
					kustomizationFile := createGitKustomization(appName, appNameSpace, "https://github.com/stefanprodan/podinfo", appName, appTargetNamespace)
					pullGitRepo(repoAbsolutePath)
					_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("mkdir -p %[2]v && cp -f %[1]v %[2]v", kustomizationFile, path.Join(repoAbsolutePath, appDir)))
					gitUpdateCommitPush(repoAbsolutePath, "Adding podinfo kustomization")
				})

				pages.NavigateToPage(webDriver, "Applications")
				applicationsPage := pages.GetApplicationsPage(webDriver)

				By("And wait for leaf cluster podinfo application to be visibe on the dashboard", func() {
					Eventually(applicationsPage.ApplicationHeader).Should(BeVisible())

					existingAppCount += 2                 // flux-system + clusters-bases-kustomization (leaf cluster)
					totalAppCount := existingAppCount + 1 // podinfo (leaf cluster)
					Eventually(func(g Gomega) string {
						g.Expect(webDriver.Refresh()).ShouldNot(HaveOccurred())
						time.Sleep(POLL_INTERVAL_1SECONDS)
						count, _ := applicationsPage.ApplicationCount.Text()
						return count

					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(MatchRegexp(strconv.Itoa(totalAppCount)), fmt.Sprintf("Dashboard failed to update with expected applications count: %d", totalAppCount))

					Eventually(func(g Gomega) int {
						return applicationsPage.CountApplications()
					}, ASSERTION_3MINUTE_TIME_OUT).Should(Equal(totalAppCount), fmt.Sprintf("There should be %d application enteries in application table", totalAppCount))
				})

				By(fmt.Sprintf("And search leaf cluster '%s' app", leafClusterName), func() {
					searchPage := pages.GetSearchPage(webDriver)
					Eventually(searchPage.SearchBtn.Click).Should(Succeed(), "Failed to click search buttton")
					Expect(searchPage.Search.SendKeys(appName)).Should(Succeed(), "Failed type application name in search field")
					Expect(searchPage.Search.SendKeys("\uE007")).Should(Succeed()) // send enter key code to do application search in table

					Eventually(func(g Gomega) int {
						return applicationsPage.CountApplications()
					}).Should(Equal(1), "There should be '1' application entery in application table after search")
				})

				applicationInfo := applicationsPage.FindApplicationInList(appName)
				By("And verify searched podinfo application information in application table", func() {
					Eventually(applicationInfo.Name).Should(MatchText(appName), fmt.Sprintf("Failed to list %s application in  application table", appName))
					Eventually(applicationInfo.Type).Should(MatchText("Kustomization"), fmt.Sprintf("Failed to have expected %s application type: Kustomization", appName))
					Eventually(applicationInfo.Namespace).Should(MatchText(appNameSpace), fmt.Sprintf("Failed to have expected %s application namespace: %s", appName, appNameSpace))
					Eventually(applicationInfo.Cluster).Should(MatchText(leafClusterNamespace+`/`+leafClusterName), fmt.Sprintf("Failed to have expected %s application cluster: %s", appName, leafClusterNamespace+`/`+leafClusterName))
					Eventually(applicationInfo.Source).Should(MatchText(appName), fmt.Sprintf("Failed to have expected %[1]v application source: %[1]s", appName))
					Eventually(applicationInfo.Status, ASSERTION_1MINUTE_TIME_OUT).Should(MatchText("Ready"), fmt.Sprintf("Failed to have expected %s application status: Ready", appName))
				})

				By(fmt.Sprintf("And navigate to %s application page", appName), func() {
					Eventually(applicationInfo.Name.Click).Should(Succeed(), fmt.Sprintf("Failed to navigate to %s application detail page", appName))
				})

				appDetailPage := pages.GetApplicationsDetailPage(webDriver)
				By(fmt.Sprintf("And verify %s application page", appName), func() {
					Eventually(appDetailPage.Header.Text).Should(MatchRegexp(appName), fmt.Sprintf("Failed to verify dashboard application name %s", appName))
					Eventually(appDetailPage.Title.Text).Should(MatchRegexp(appName), fmt.Sprintf("Failed to verify application title %s on application page", appName))
					Eventually(appDetailPage.Sync).Should(BeEnabled(), fmt.Sprintf("Sync button is not visible/enable for %s", appName))
					Eventually(appDetailPage.Details).Should(BeEnabled(), fmt.Sprintf("Details tab button is not visible/enable for %s", appName))
					Eventually(appDetailPage.Events).Should(BeEnabled(), fmt.Sprintf("Events tab button is not visible/enable for %s", appName))
					Eventually(appDetailPage.Graph).Should(BeEnabled(), fmt.Sprintf("Graph tab button is not visible/enable for %s", appName))
				})

				By(fmt.Sprintf("And verify %s application Details", appName), func() {
					Expect(appDetailPage.Details.Click()).Should(Succeed(), fmt.Sprintf("Failed to click %s Details tab button", appName))
					pages.WaitForPageToLoad(webDriver)

					details := pages.GetApplicationDetail(webDriver)

					Eventually(details.Source.Text).Should(MatchRegexp("GitRepository/"+appName), fmt.Sprintf("Failed to verify %s Source", appName))
					Eventually(details.AppliedRevision.Text).Should(MatchRegexp("master"), fmt.Sprintf("Failed to verify %s AppliedRevision", appName))
					Eventually(details.Cluster.Text).Should(MatchRegexp(leafClusterNamespace+`/`+leafClusterName), fmt.Sprintf("Failed to verify %s leaf Cluster", appName))
					Eventually(details.Path.Text).Should(MatchRegexp("./kustomize"), fmt.Sprintf("Failed to verify %s Path", appName))
					Eventually(details.Interval.Text).Should(MatchRegexp(appSyncInterval), fmt.Sprintf("Failed to verify %s AppliedRevision", appName))

					Eventually(appDetailPage.Sync.Click).Should(Succeed(), fmt.Sprintf("Failed to sync %s kustomization", appName))
					Eventually(details.LastUpdated.Text).Should(MatchRegexp("seconds ago"), fmt.Sprintf("Failed to verify %s LastUpdated", appName))

					Eventually(details.Name.Text).Should(MatchRegexp(deploymentName), fmt.Sprintf("Failed to verify %s Deployment name", appName))
					Eventually(details.Type.Text).Should(MatchRegexp("Deployment"), fmt.Sprintf("Failed to verify %s Type", appName))
					Eventually(details.Namespace.Text).Should(MatchRegexp(appTargetNamespace), fmt.Sprintf("Failed to verify %s Namespace", appName))
					Eventually(details.Status.Text, ASSERTION_2MINUTE_TIME_OUT).Should(MatchRegexp("Ready"), fmt.Sprintf("Failed to verify %s Status", appName))
					Eventually(details.Message.Text).Should(MatchRegexp("Deployment is available"), fmt.Sprintf("Failed to verify %s Message", appName))

					Expect(details.GetMetadata("Description").Text()).Should(MatchRegexp(`Podinfo is a tiny web application made with Go`), "Failed to verify Metada description")
					verifyDashboard(details.GetMetadata("Grafana Dashboard").Find("a"), "management", "Grafana")
					Expect(details.GetMetadata("Javascript Alert").Find("a")).ShouldNot(BeFound(), "Javascript href is not sanitized")
					Expect(details.GetMetadata("Javascript Alert").Text()).Should(MatchRegexp(`javascript:alert\('hello there'\);`), "Failed to verify Javascript alert text")
				})

				By(fmt.Sprintf("And verify %s application Events", appName), func() {
					Expect(appDetailPage.Events.Click()).Should(Succeed(), fmt.Sprintf("Failed to click %s Events tab button", appName))
					pages.WaitForPageToLoad(webDriver)

					event := pages.GetApplicationEvent(webDriver, "ReconciliationSucceeded")

					Eventually(event.Reason.Text).Should(MatchRegexp("ReconciliationSucceeded"), fmt.Sprintf("Failed to verify %s Event/Reason", appName))
					Eventually(event.Message.Text).Should(MatchRegexp("next run in "+appSyncInterval), fmt.Sprintf("Failed to verify %s Event/Message", appName))
					Eventually(event.Component.Text).Should(MatchRegexp("kustomize-controller"), fmt.Sprintf("Failed to verify %s Event/Component", appName))
					Eventually(event.TimeStamp.Text).Should(MatchRegexp("seconds|minutes|minute ago"), fmt.Sprintf("Failed to verify %s Event/Timestamp", appName))
				})

				By(fmt.Sprintf("And verify %s application Grapg", appName), func() {
					Expect(appDetailPage.Graph.Click()).Should(Succeed(), fmt.Sprintf("Failed to click %s Graph tab button", appName))
					pages.WaitForPageToLoad(webDriver)

					graph := pages.GetApplicationGraph(webDriver, deploymentName, appName, appNameSpace, appTargetNamespace)

					Expect(graph.SourceGit).Should(BeFound(), fmt.Sprintf("Failed to verify %s Graph/Source", appName))
					Expect(graph.Kustomization).Should(BeFound(), fmt.Sprintf("Failed to verify %s Graph/Kustomization", appName))
					Expect(graph.Deployment).Should(BeFound(), fmt.Sprintf("Failed to verify %s Graph/Deployment", appName))
					Expect(graph.ReplicaSet).Should(BeFound(), fmt.Sprintf("Failed to verify %s Graph/ReplicaSet", appName))
					Expect(graph.Pod).Should(BeFound(), fmt.Sprintf("Failed to verify %s Graph/Pod", appName))
				})

				navigatetoApplicationsPage(applicationsPage)

				By(fmt.Sprintf("And navigate directly to %s Sources page", appName), func() {
					Expect(applicationInfo.Source.Click()).Should(Succeed(), fmt.Sprintf("Failed to navigate to %s Sources pages directly", appName))

					sourceDetailPage := pages.GetSourceDetailPage(webDriver)
					Eventually(sourceDetailPage.Header.Text).Should(MatchRegexp(appName), fmt.Sprintf("Failed to verify dashboard header source name %s ", appName))
					Eventually(sourceDetailPage.Title.Text).Should(MatchRegexp(appName), fmt.Sprintf("Failed to verify application title %s on source page", appName))
				})

				navigatetoApplicationsPage(applicationsPage)

				By("And delete the podinfo kustomization and source maifest from the repository's master branch", func() {
					cleanGitRepository(appDir)
				})

				By("And wait for podinfo application to dissappeare from the dashboard", func() {
					Eventually(applicationsPage.ApplicationCount, ASSERTION_3MINUTE_TIME_OUT).Should(MatchText(strconv.Itoa(existingAppCount)), fmt.Sprintf("Dashboard failed to update with expected applications count: %d", existingAppCount))
					Eventually(func(g Gomega) int {
						return applicationsPage.CountApplications()
					}, ASSERTION_3MINUTE_TIME_OUT).Should(Equal(existingAppCount), fmt.Sprintf("There should be %d application enteries in application table", existingAppCount))
				})
			})
		})
	})
}
