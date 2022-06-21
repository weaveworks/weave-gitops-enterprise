package acceptance

import (
	"fmt"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

func navigatetoApplicationsPage(applicationsPage *pages.ApplicationsPage) {
	By("And navigate to Applicartions page via header link", func() {
		Expect(applicationsPage.ApplicationHeader.Click()).Should(Succeed(), "Failed to navigate to Applications pages via header link")
		pages.WaitForPageToLoad(webDriver)
	})
}

func DescribeApplications(gitopsTestRunner GitopsTestRunner) {
	var _ = Describe("Multi-Cluster Control Plane Applications", func() {

		Context("[UI] When no applications are installed", func() {
			FIt("Verify management cluster dashboard shows only bootstrap 'flux-system' application", Label("integration", "application"), func() {
				pages.NavigateToPage(webDriver, "Applications")
				applicationsPage := pages.GetApplicationsPage(webDriver)
				pages.WaitForPageToLoad(webDriver)

				By("And wait for Applications page to be rendered", func() {
					Eventually(applicationsPage.ApplicationHeader).Should(BeVisible())
					Eventually(applicationsPage.ApplicationCount).Should(MatchText(`1`))
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
					Eventually(applicationInfo.Status).Should(MatchText("Ready"), "Failed to have expected flux-system application status: Ready")
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
				existingAppCount := 1

				pages.NavigateToPage(webDriver, "Applications")
				applicationsPage := pages.GetApplicationsPage(webDriver)
				By("And wait for Applications page to be fully rendered", func() {
					pages.WaitForPageToLoad(webDriver)
					existingAppCount = applicationsPage.CountApplications()
				})

				By("Create namespaces for Kustomization deployments)", func() {
					createNamespace([]string{appNameSpace, appTargetNamespace})
				})

				By("Create a GitRepository manifest pointing to podinfo repository’s master branch", func() {
					pullGitRepo(repoAbsolutePath)

					err := runCommandPassThrough("sh", "-c",
						fmt.Sprintf(`cd %[1]v && mkdir -p %[2]v &&
						flux create source git %[3]v --url=https://github.com/stefanprodan/podinfo --branch=master --interval=%[4]v --namespace=%[5]v --export > %[2]v/%[3]v-source.yaml`,
							repoAbsolutePath, appDir, appName, appSyncInterval, appNameSpace))
					Expect(err).ShouldNot(HaveOccurred(), fmt.Sprintf("Failed to generate Gitrepository %s source manifest", appName))

					gitUpdateCommitPush(repoAbsolutePath, "")

				})

				By("Create a Kustomization that applies the podinfo deployment", func() {
					pullGitRepo(repoAbsolutePath)

					err := runCommandPassThrough("sh", "-c",
						fmt.Sprintf(`cd %[1]v && mkdir -p %[2]v &&
						flux create kustomization %[3]v --target-namespace=%[4]v --source=%[3]v --path="./kustomize" --prune=true --interval=%[5]v --namespace=%[6]v --export > %[2]v/%[3]v-kustomization.yaml`,
							repoAbsolutePath, appDir, appName, appTargetNamespace, appSyncInterval, appNameSpace))
					Expect(err).ShouldNot(HaveOccurred(), fmt.Sprintf("Failed to generate Kustomization %s manifest", appName))

					gitUpdateCommitPush(repoAbsolutePath, "")

				})

				By("And wait for podinfo application to be visibe on the dashboard", func() {
					Eventually(applicationsPage.ApplicationHeader).Should(BeVisible())

					totalAppCount := existingAppCount + 1
					Eventually(applicationsPage.ApplicationCount, ASSERTION_3MINUTE_TIME_OUT).Should(MatchText(strconv.Itoa(totalAppCount)), fmt.Sprintf("Dashboard failed to update with expected applications count: %d", totalAppCount))
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
					Eventually(applicationInfo.Status).Should(MatchText("Ready"), fmt.Sprintf("Failed to have expected %s application status: Ready", appName))
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

					details := pages.GetApplicationDetail(webDriver, deploymentName)

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
					Eventually(details.Status.Text, ASSERTION_1MINUTE_TIME_OUT).Should(MatchRegexp("Ready"), fmt.Sprintf("Failed to verify %s Status", appName))
					Eventually(details.Message.Text).Should(MatchRegexp("Deployment is available"), fmt.Sprintf("Failed to verify %s Message", appName))

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

					graph := pages.GetApplicationGraph(webDriver, appNameSpace, appTargetNamespace)

					Expect(graph.SourceGit).Should(BeVisible(), fmt.Sprintf("Failed to verify %s Graph/Source", appName))
					Expect(graph.Kustomization).Should(BeVisible(), fmt.Sprintf("Failed to verify %s Graph/Kustomization", appName))
					Expect(graph.Deployment).Should(BeVisible(), fmt.Sprintf("Failed to verify %s Graph/Deployment", appName))
					Expect(graph.ReplicaSet).Should(BeFound(), fmt.Sprintf("Failed to verify %s Graph/ReplicaSet", appName))
					Expect(graph.Pod.At(0)).Should(BeFound(), fmt.Sprintf("Failed to verify %s Graph/Pod", appName))
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
