package acceptance

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"text/template"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

func createGitKustomization(repoName, nameSpace, repoURL, kustomizationName, targetNamespace string) (kustomization string) {
	contents, err := ioutil.ReadFile(path.Join(getCheckoutRepoPath(), "test", "utils", "data", "git-kustomization.yaml"))
	gomega.Expect(err).To(gomega.BeNil(), "Failed to read git-kustomization template yaml")

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
	gomega.Expect(err).To(gomega.BeNil(), "Failed to create kustomization manifest yaml")

	err = t.Execute(f, input)
	f.Close()
	gomega.Expect(err).To(gomega.BeNil(), "Failed to generate kustomization manifest yaml")

	return kustomization
}

func navigatetoApplicationsPage(applicationsPage *pages.ApplicationsPage) {
	ginkgo.By("And navigate to Applicartions page via header link", func() {
		gomega.Expect(applicationsPage.ApplicationHeader.Click()).Should(gomega.Succeed(), "Failed to navigate to Applications pages via header link")
		pages.WaitForPageToLoad(webDriver)
	})
}

func verifyAppInformation(applicationsPage *pages.ApplicationsPage, appName, appType, appNamespace, cluster, clusterNamespace, source, status string) {
	ginkgo.By(fmt.Sprintf("And verify %s application information in application table for cluster: %s", appName, cluster), func() {
		applicationInfo := applicationsPage.FindApplicationInList(appName)
		gomega.Eventually(applicationInfo.Name).Should(matchers.MatchText(appName), fmt.Sprintf("Failed to list %s application in  application table", appName))
		gomega.Eventually(applicationInfo.Type).Should(matchers.MatchText(appType), fmt.Sprintf("Failed to have expected %s application type: %s", appName, appType))
		gomega.Eventually(applicationInfo.Namespace).Should(matchers.MatchText(appNamespace), fmt.Sprintf("Failed to have expected %s application namespace: %s", appName, appNamespace))
		gomega.Eventually(applicationInfo.Cluster).Should(matchers.MatchText(path.Join(clusterNamespace, cluster)), fmt.Sprintf("Failed to have expected %s application cluster: %s", appName, path.Join(clusterNamespace, cluster)))
		gomega.Eventually(applicationInfo.Source).Should(matchers.MatchText(source), fmt.Sprintf("Failed to have expected %s application source: %s", appName, source))
		gomega.Eventually(applicationInfo.Status, ASSERTION_1MINUTE_TIME_OUT).Should(matchers.MatchText(status), fmt.Sprintf("Failed to have expected %s application status: %s", appName, status))
	})
}

func DescribeApplications(gitopsTestRunner GitopsTestRunner) {
	var _ = ginkgo.Describe("Multi-Cluster Control Plane Applications", func() {

		ginkgo.Context("[UI] When no applications are installed", func() {
			ginkgo.It("Verify management cluster dashboard shows bootstrap 'flux-system' application", ginkgo.Label("integration"), func() {
				existingAppCount := getApplicationCount()

				pages.NavigateToPage(webDriver, "Applications")

				ginkgo.By("And wait for  good looking response from /v1/objects", func() {
					gomega.Expect(waitForGitopsResources(context.Background(), "objects?kind=Kustomization", POLL_INTERVAL_15SECONDS)).To(gomega.Succeed(), "Failed to get a successful response from /v1/objects")
				})

				applicationsPage := pages.GetApplicationsPage(webDriver)
				pages.WaitForPageToLoad(webDriver)

				ginkgo.By("And wait for Applications page to be rendered", func() {
					gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())

					gomega.Eventually(func(g gomega.Gomega) string {
						g.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
						time.Sleep(POLL_INTERVAL_1SECONDS)
						count, _ := applicationsPage.ApplicationCount.Text()
						return count

					}, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.MatchRegexp(strconv.Itoa(existingAppCount)), fmt.Sprintf("Dashboard failed to update with expected applications count: %d", existingAppCount))

					gomega.Expect(applicationsPage.CountApplications()).To(gomega.Equal(1), "There should not be any cluster in cluster table")
				})

				verifyAppInformation(applicationsPage, "flux-system", "Kustomization", GITOPS_DEFAULT_NAMESPACE, "management", "", "flux-system", "Ready")
			})
		})

		ginkgo.Context("[UI] Applications(s) can be installed", func() {
			appDir := "./clusters/my-cluster/podinfo"
			deploymentName := "podinfo"
			appName := "my-podinfo"
			appNameSpace := "test-kustomization"
			appTargetNamespace := "test-systems"
			appSyncInterval := "30s"

			ginkgo.JustAfterEach(func() {
				cleanGitRepository(appDir)
				deleteNamespace([]string{appNameSpace, appTargetNamespace})
			})

			ginkgo.It("Verify application can be installed  and dashboard is updated accordingly", ginkgo.Label("integration", "application", "browser-logs"), func() {
				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
				existingAppCount := getApplicationCount()

				pages.NavigateToPage(webDriver, "Applications")
				applicationsPage := pages.GetApplicationsPage(webDriver)

				ginkgo.By("Create namespaces for Kustomization deployments)", func() {
					createNamespace([]string{appNameSpace, appTargetNamespace})
				})

				ginkgo.By("And add Kustomization & GitRepository Source manifests pointing to podinfo repository’s master branch)", func() {
					kustomizationFile := createGitKustomization(appName, appNameSpace, "https://github.com/stefanprodan/podinfo", appName, appTargetNamespace)
					pullGitRepo(repoAbsolutePath)
					_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("mkdir -p %[2]v && cp -f %[1]v %[2]v", kustomizationFile, path.Join(repoAbsolutePath, appDir)))
					gitUpdateCommitPush(repoAbsolutePath, "Adding podinfo kustomization")
				})

				ginkgo.By("And wait for podinfo application to be visibe on the dashboard", func() {
					gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())

					totalAppCount := existingAppCount + 1
					gomega.Eventually(func(g gomega.Gomega) string {
						g.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
						time.Sleep(POLL_INTERVAL_1SECONDS)
						count, _ := applicationsPage.ApplicationCount.Text()
						return count

					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.MatchRegexp(strconv.Itoa(totalAppCount)), fmt.Sprintf("Dashboard failed to update with expected applications count: %d", totalAppCount))

					gomega.Eventually(func(g gomega.Gomega) int {
						return applicationsPage.CountApplications()
					}, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(totalAppCount), fmt.Sprintf("There should be %d application enteries in application table", totalAppCount))
				})

				verifyAppInformation(applicationsPage, appName, "Kustomization", appNameSpace, "management", "", appName, "Ready")

				applicationInfo := applicationsPage.FindApplicationInList(appName)
				ginkgo.By(fmt.Sprintf("And navigate to %s application page", appName), func() {
					gomega.Eventually(applicationInfo.Name.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to %s application detail page", appName))
				})

				appDetailPage := pages.GetApplicationsDetailPage(webDriver)
				ginkgo.By(fmt.Sprintf("And verify %s application page", appName), func() {
					gomega.Eventually(appDetailPage.Header.Text).Should(gomega.MatchRegexp(appName), fmt.Sprintf("Failed to verify dashboard application name %s", appName))
					gomega.Eventually(appDetailPage.Title.Text).Should(gomega.MatchRegexp(appName), fmt.Sprintf("Failed to verify application title %s on application page", appName))
					gomega.Eventually(appDetailPage.Sync).Should(matchers.BeEnabled(), fmt.Sprintf("Sync button is not visible/enable for %s", appName))
					gomega.Eventually(appDetailPage.Details).Should(matchers.BeEnabled(), fmt.Sprintf("Details tab button is not visible/enable for %s", appName))
					gomega.Eventually(appDetailPage.Events).Should(matchers.BeEnabled(), fmt.Sprintf("Events tab button is not visible/enable for %s", appName))
					gomega.Eventually(appDetailPage.Graph).Should(matchers.BeEnabled(), fmt.Sprintf("Graph tab button is not visible/enable for %s", appName))
				})

				ginkgo.By(fmt.Sprintf("And verify %s application Details", appName), func() {
					gomega.Expect(appDetailPage.Details.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to click %s Details tab button", appName))
					pages.WaitForPageToLoad(webDriver)

					details := pages.GetApplicationDetail(webDriver)

					gomega.Eventually(details.Source.Text).Should(gomega.MatchRegexp("GitRepository/"+appName), fmt.Sprintf("Failed to verify %s Source", appName))
					gomega.Eventually(details.AppliedRevision.Text).Should(gomega.MatchRegexp("master"), fmt.Sprintf("Failed to verify %s AppliedRevision", appName))
					gomega.Eventually(details.Cluster.Text).Should(gomega.MatchRegexp("management"), fmt.Sprintf("Failed to verify %s Cluster", appName))
					gomega.Eventually(details.Path.Text).Should(gomega.MatchRegexp("./kustomize"), fmt.Sprintf("Failed to verify %s Path", appName))
					gomega.Eventually(details.Interval.Text).Should(gomega.MatchRegexp(appSyncInterval), fmt.Sprintf("Failed to verify %s AppliedRevision", appName))

					gomega.Eventually(appDetailPage.Sync.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to sync %s kustomization", appName))
					gomega.Eventually(details.LastUpdated.Text).Should(gomega.MatchRegexp("seconds ago"), fmt.Sprintf("Failed to verify %s LastUpdated", appName))

					gomega.Eventually(details.Name.Text).Should(gomega.MatchRegexp(deploymentName), fmt.Sprintf("Failed to verify %s Deployment name", appName))
					gomega.Eventually(details.Type.Text).Should(gomega.MatchRegexp("Deployment"), fmt.Sprintf("Failed to verify %s Type", appName))
					gomega.Eventually(details.Namespace.Text).Should(gomega.MatchRegexp(appTargetNamespace), fmt.Sprintf("Failed to verify %s Namespace", appName))
					gomega.Eventually(details.Status.Text, ASSERTION_2MINUTE_TIME_OUT).Should(gomega.MatchRegexp("Ready"), fmt.Sprintf("Failed to verify %s Status", appName))
					gomega.Eventually(details.Message.Text).Should(gomega.MatchRegexp("Deployment is available"), fmt.Sprintf("Failed to verify %s Message", appName))

					// Verify metadata
					gomega.Expect(details.GetMetadata("Description").Text()).Should(gomega.MatchRegexp(`Podinfo is a tiny web application made with Go`), "Failed to verify Metada description")
					verifyDashboard(details.GetMetadata("Grafana Dashboard").Find("a"), "management", "Grafana")
					gomega.Expect(details.GetMetadata("Javascript Alert").Find("a")).ShouldNot(matchers.BeFound(), "Javascript href is not sanitized")
					gomega.Expect(details.GetMetadata("Javascript Alert").Text()).Should(gomega.MatchRegexp(`javascript:alert\('hello there'\);`), "Failed to verify Javascript alert text")
				})

				ginkgo.By(fmt.Sprintf("And verify %s application Events", appName), func() {
					gomega.Expect(appDetailPage.Events.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to click %s Events tab button", appName))
					pages.WaitForPageToLoad(webDriver)

					event := pages.GetApplicationEvent(webDriver, "ReconciliationSucceeded")

					gomega.Eventually(event.Reason.Text).Should(gomega.MatchRegexp("ReconciliationSucceeded"), fmt.Sprintf("Failed to verify %s Event/Reason", appName))
					gomega.Eventually(event.Message.Text).Should(gomega.MatchRegexp("next run in "+appSyncInterval), fmt.Sprintf("Failed to verify %s Event/Message", appName))
					gomega.Eventually(event.Component.Text).Should(gomega.MatchRegexp("kustomize-controller"), fmt.Sprintf("Failed to verify %s Event/Component", appName))
					gomega.Eventually(event.TimeStamp.Text).Should(gomega.MatchRegexp("seconds|minutes|minute ago"), fmt.Sprintf("Failed to verify %s Event/Timestamp", appName))
				})

				ginkgo.By(fmt.Sprintf("And verify %s application Grapg", appName), func() {
					gomega.Expect(appDetailPage.Graph.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to click %s Graph tab button", appName))
					pages.WaitForPageToLoad(webDriver)

					graph := pages.GetApplicationGraph(webDriver, deploymentName, appName, appNameSpace, appTargetNamespace)

					gomega.Expect(graph.SourceGit).Should(matchers.BeFound(), fmt.Sprintf("Failed to verify %s Graph/Source", appName))
					gomega.Expect(graph.Kustomization).Should(matchers.BeFound(), fmt.Sprintf("Failed to verify %s Graph/Kustomization", appName))
					gomega.Expect(graph.Deployment).Should(matchers.BeFound(), fmt.Sprintf("Failed to verify %s Graph/Deployment", appName))
					gomega.Expect(graph.ReplicaSet).Should(matchers.BeFound(), fmt.Sprintf("Failed to verify %s Graph/ReplicaSet", appName))
					gomega.Expect(graph.Pod).Should(matchers.BeFound(), fmt.Sprintf("Failed to verify %s Graph/Pod", appName))
				})

				navigatetoApplicationsPage(applicationsPage)

				ginkgo.By(fmt.Sprintf("And navigate directly to %s Sources page", appName), func() {
					gomega.Expect(applicationInfo.Source.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to %s Sources pages directly", appName))

					sourceDetailPage := pages.GetSourceDetailPage(webDriver)
					gomega.Eventually(sourceDetailPage.Header.Text).Should(gomega.MatchRegexp(appName), fmt.Sprintf("Failed to verify dashboard header source name %s ", appName))
					gomega.Eventually(sourceDetailPage.Title.Text).Should(gomega.MatchRegexp(appName), fmt.Sprintf("Failed to verify application title %s on source page", appName))
				})

				navigatetoApplicationsPage(applicationsPage)

				ginkgo.By("And delete the podinfo kustomization and source maifest from the repository's master branch", func() {
					cleanGitRepository(appDir)
				})

				ginkgo.By("And wait for podinfo application to dissappeare from the dashboard", func() {
					gomega.Eventually(applicationsPage.ApplicationCount, ASSERTION_3MINUTE_TIME_OUT).Should(matchers.MatchText(strconv.Itoa(existingAppCount)), fmt.Sprintf("Dashboard failed to update with expected applications count: %d", existingAppCount))
					gomega.Eventually(func(g gomega.Gomega) int {
						return applicationsPage.CountApplications()
					}, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(existingAppCount), fmt.Sprintf("There should be %d application enteries in application table", existingAppCount))
				})
			})
		})

		ginkgo.Context("[UI] Applications(s) can be installed on leaf cluster", func() {
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

			ginkgo.JustBeforeEach(func() {
				existingAppCount = getApplicationCount()
				appDir = path.Join("clusters", leafClusterNamespace, leafClusterName, "apps")
				mgmtClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
				createCluster("kind", leafClusterName, "")
				createNamespace([]string{appNameSpace, appTargetNamespace})
				leafClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
			})

			ginkgo.JustAfterEach(func() {
				useClusterContext(mgmtClusterContext)

				deleteSecret([]string{leafClusterkubeconfig, patSecret}, leafClusterNamespace)
				_ = gitopsTestRunner.KubectlDelete([]string{}, clusterBootstrapCopnfig)
				_ = gitopsTestRunner.KubectlDelete([]string{}, gitopsCluster)

				deleteCluster("kind", leafClusterName, "")
				cleanGitRepository(appDir)
				deleteNamespace([]string{leafClusterNamespace})

			})

			ginkgo.It("Verify application can be installed  on leaf cluster and management dashboard is updated accordingly", ginkgo.Label("integration", "application", "leaf-application"), func() {
				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
				leafClusterkubeconfig = createLeafClusterKubeconfig(leafClusterContext, leafClusterName, leafClusterNamespace)

				useClusterContext(mgmtClusterContext)
				createNamespace([]string{leafClusterNamespace})

				createPATSecret(leafClusterNamespace, patSecret)
				clusterBootstrapCopnfig = createClusterBootstrapConfig(leafClusterName, leafClusterNamespace, bootstrapLabel, patSecret)
				gitopsCluster = connectGitopsCuster(leafClusterName, leafClusterNamespace, bootstrapLabel, leafClusterkubeconfig)
				createLeafClusterSecret(leafClusterNamespace, leafClusterkubeconfig)

				ginkgo.By("Verify GitopsCluster status after creating kubeconfig secret", func() {
					pages.NavigateToPage(webDriver, "Clusters")
					clustersPage := pages.GetClustersPage(webDriver)
					pages.WaitForPageToLoad(webDriver)
					clusterInfo := clustersPage.FindClusterInList(leafClusterName)

					gomega.Eventually(clusterInfo.Status, ASSERTION_30SECONDS_TIME_OUT).Should(matchers.MatchText("Ready"))
				})

				ginkgo.By("And add kustomization bases for common resources for leaf cluster)", func() {
					addKustomizationBases("leaf", leafClusterName, leafClusterNamespace)
				})

				ginkgo.By(fmt.Sprintf("And I verify %s GitopsCluster is bootstraped)", leafClusterName), func() {
					useClusterContext(leafClusterContext)
					verifyFluxControllers(GITOPS_DEFAULT_NAMESPACE)
					waitForGitRepoReady("flux-system", GITOPS_DEFAULT_NAMESPACE)
					useClusterContext(mgmtClusterContext)
				})

				ginkgo.By("And add Kustomization & GitRepository Source manifests pointing to podinfo repository’s master branch)", func() {
					kustomizationFile := createGitKustomization(appName, appNameSpace, "https://github.com/stefanprodan/podinfo", appName, appTargetNamespace)
					pullGitRepo(repoAbsolutePath)
					_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("mkdir -p %[2]v && cp -f %[1]v %[2]v", kustomizationFile, path.Join(repoAbsolutePath, appDir)))
					gitUpdateCommitPush(repoAbsolutePath, "Adding podinfo kustomization")
				})

				pages.NavigateToPage(webDriver, "Applications")
				applicationsPage := pages.GetApplicationsPage(webDriver)

				ginkgo.By("And wait for leaf cluster podinfo application to be visibe on the dashboard", func() {
					gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())

					existingAppCount += 2                 // flux-system + clusters-bases-kustomization (leaf cluster)
					totalAppCount := existingAppCount + 1 // podinfo (leaf cluster)
					gomega.Eventually(func(g gomega.Gomega) string {
						g.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
						time.Sleep(POLL_INTERVAL_1SECONDS)
						count, _ := applicationsPage.ApplicationCount.Text()
						return count

					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.MatchRegexp(strconv.Itoa(totalAppCount)), fmt.Sprintf("Dashboard failed to update with expected applications count: %d", totalAppCount))

					gomega.Eventually(func(g gomega.Gomega) int {
						return applicationsPage.CountApplications()
					}, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(totalAppCount), fmt.Sprintf("There should be %d application enteries in application table", totalAppCount))
				})

				ginkgo.By(fmt.Sprintf("And search leaf cluster '%s' app", leafClusterName), func() {
					searchPage := pages.GetSearchPage(webDriver)
					gomega.Eventually(searchPage.SearchBtn.Click).Should(gomega.Succeed(), "Failed to click search buttton")
					gomega.Expect(searchPage.Search.SendKeys(appName)).Should(gomega.Succeed(), "Failed type application name in search field")
					gomega.Expect(searchPage.Search.SendKeys("\uE007")).Should(gomega.Succeed()) // send enter key code to do application search in table

					gomega.Eventually(func(g gomega.Gomega) int {
						return applicationsPage.CountApplications()
					}).Should(gomega.Equal(1), "There should be '1' application entery in application table after search")
				})

				verifyAppInformation(applicationsPage, appName, "Kustomization", appNameSpace, leafClusterName, leafClusterNamespace, appName, "Ready")

				applicationInfo := applicationsPage.FindApplicationInList(appName)
				ginkgo.By(fmt.Sprintf("And navigate to %s application page", appName), func() {
					gomega.Eventually(applicationInfo.Name.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to %s application detail page", appName))
				})

				appDetailPage := pages.GetApplicationsDetailPage(webDriver)
				ginkgo.By(fmt.Sprintf("And verify %s application page", appName), func() {
					gomega.Eventually(appDetailPage.Header.Text).Should(gomega.MatchRegexp(appName), fmt.Sprintf("Failed to verify dashboard application name %s", appName))
					gomega.Eventually(appDetailPage.Title.Text).Should(gomega.MatchRegexp(appName), fmt.Sprintf("Failed to verify application title %s on application page", appName))
					gomega.Eventually(appDetailPage.Sync).Should(matchers.BeEnabled(), fmt.Sprintf("Sync button is not visible/enable for %s", appName))
					gomega.Eventually(appDetailPage.Details).Should(matchers.BeEnabled(), fmt.Sprintf("Details tab button is not visible/enable for %s", appName))
					gomega.Eventually(appDetailPage.Events).Should(matchers.BeEnabled(), fmt.Sprintf("Events tab button is not visible/enable for %s", appName))
					gomega.Eventually(appDetailPage.Graph).Should(matchers.BeEnabled(), fmt.Sprintf("Graph tab button is not visible/enable for %s", appName))
				})

				ginkgo.By(fmt.Sprintf("And verify %s application Details", appName), func() {
					gomega.Expect(appDetailPage.Details.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to click %s Details tab button", appName))
					pages.WaitForPageToLoad(webDriver)

					details := pages.GetApplicationDetail(webDriver)

					gomega.Eventually(details.Source.Text).Should(gomega.MatchRegexp("GitRepository/"+appName), fmt.Sprintf("Failed to verify %s Source", appName))
					gomega.Eventually(details.AppliedRevision.Text).Should(gomega.MatchRegexp("master"), fmt.Sprintf("Failed to verify %s AppliedRevision", appName))
					gomega.Eventually(details.Cluster.Text).Should(gomega.MatchRegexp(leafClusterNamespace+`/`+leafClusterName), fmt.Sprintf("Failed to verify %s leaf Cluster", appName))
					gomega.Eventually(details.Path.Text).Should(gomega.MatchRegexp("./kustomize"), fmt.Sprintf("Failed to verify %s Path", appName))
					gomega.Eventually(details.Interval.Text).Should(gomega.MatchRegexp(appSyncInterval), fmt.Sprintf("Failed to verify %s AppliedRevision", appName))

					gomega.Eventually(appDetailPage.Sync.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to sync %s kustomization", appName))
					gomega.Eventually(details.LastUpdated.Text).Should(gomega.MatchRegexp("seconds ago"), fmt.Sprintf("Failed to verify %s LastUpdated", appName))

					gomega.Eventually(details.Name.Text).Should(gomega.MatchRegexp(deploymentName), fmt.Sprintf("Failed to verify %s Deployment name", appName))
					gomega.Eventually(details.Type.Text).Should(gomega.MatchRegexp("Deployment"), fmt.Sprintf("Failed to verify %s Type", appName))
					gomega.Eventually(details.Namespace.Text).Should(gomega.MatchRegexp(appTargetNamespace), fmt.Sprintf("Failed to verify %s Namespace", appName))
					gomega.Eventually(details.Status.Text, ASSERTION_2MINUTE_TIME_OUT).Should(gomega.MatchRegexp("Ready"), fmt.Sprintf("Failed to verify %s Status", appName))
					gomega.Eventually(details.Message.Text).Should(gomega.MatchRegexp("Deployment is available"), fmt.Sprintf("Failed to verify %s Message", appName))

					gomega.Expect(details.GetMetadata("Description").Text()).Should(gomega.MatchRegexp(`Podinfo is a tiny web application made with Go`), "Failed to verify Metada description")
					verifyDashboard(details.GetMetadata("Grafana Dashboard").Find("a"), "management", "Grafana")
					gomega.Expect(details.GetMetadata("Javascript Alert").Find("a")).ShouldNot(matchers.BeFound(), "Javascript href is not sanitized")
					gomega.Expect(details.GetMetadata("Javascript Alert").Text()).Should(gomega.MatchRegexp(`javascript:alert\('hello there'\);`), "Failed to verify Javascript alert text")
				})

				ginkgo.By(fmt.Sprintf("And verify %s application Events", appName), func() {
					gomega.Expect(appDetailPage.Events.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to click %s Events tab button", appName))
					pages.WaitForPageToLoad(webDriver)

					event := pages.GetApplicationEvent(webDriver, "ReconciliationSucceeded")

					gomega.Eventually(event.Reason.Text).Should(gomega.MatchRegexp("ReconciliationSucceeded"), fmt.Sprintf("Failed to verify %s Event/Reason", appName))
					gomega.Eventually(event.Message.Text).Should(gomega.MatchRegexp("next run in "+appSyncInterval), fmt.Sprintf("Failed to verify %s Event/Message", appName))
					gomega.Eventually(event.Component.Text).Should(gomega.MatchRegexp("kustomize-controller"), fmt.Sprintf("Failed to verify %s Event/Component", appName))
					gomega.Eventually(event.TimeStamp.Text).Should(gomega.MatchRegexp("seconds|minutes|minute ago"), fmt.Sprintf("Failed to verify %s Event/Timestamp", appName))
				})

				ginkgo.By(fmt.Sprintf("And verify %s application Grapg", appName), func() {
					gomega.Expect(appDetailPage.Graph.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to click %s Graph tab button", appName))
					pages.WaitForPageToLoad(webDriver)

					graph := pages.GetApplicationGraph(webDriver, deploymentName, appName, appNameSpace, appTargetNamespace)

					gomega.Expect(graph.SourceGit).Should(matchers.BeFound(), fmt.Sprintf("Failed to verify %s Graph/Source", appName))
					gomega.Expect(graph.Kustomization).Should(matchers.BeFound(), fmt.Sprintf("Failed to verify %s Graph/Kustomization", appName))
					gomega.Expect(graph.Deployment).Should(matchers.BeFound(), fmt.Sprintf("Failed to verify %s Graph/Deployment", appName))
					gomega.Expect(graph.ReplicaSet).Should(matchers.BeFound(), fmt.Sprintf("Failed to verify %s Graph/ReplicaSet", appName))
					gomega.Expect(graph.Pod).Should(matchers.BeFound(), fmt.Sprintf("Failed to verify %s Graph/Pod", appName))
				})

				navigatetoApplicationsPage(applicationsPage)

				ginkgo.By(fmt.Sprintf("And navigate directly to %s Sources page", appName), func() {
					gomega.Expect(applicationInfo.Source.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to %s Sources pages directly", appName))

					sourceDetailPage := pages.GetSourceDetailPage(webDriver)
					gomega.Eventually(sourceDetailPage.Header.Text).Should(gomega.MatchRegexp(appName), fmt.Sprintf("Failed to verify dashboard header source name %s ", appName))
					gomega.Eventually(sourceDetailPage.Title.Text).Should(gomega.MatchRegexp(appName), fmt.Sprintf("Failed to verify application title %s on source page", appName))
				})

				navigatetoApplicationsPage(applicationsPage)

				ginkgo.By("And delete the podinfo kustomization and source maifest from the repository's master branch", func() {
					cleanGitRepository(appDir)
				})

				ginkgo.By("And wait for podinfo application to dissappeare from the dashboard", func() {
					gomega.Eventually(applicationsPage.ApplicationCount, ASSERTION_3MINUTE_TIME_OUT).Should(matchers.MatchText(strconv.Itoa(existingAppCount)), fmt.Sprintf("Dashboard failed to update with expected applications count: %d", existingAppCount))
					gomega.Eventually(func(g gomega.Gomega) int {
						return applicationsPage.CountApplications()
					}, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(existingAppCount), fmt.Sprintf("There should be %d application enteries in application table", existingAppCount))
				})
			})
		})
	})
}
