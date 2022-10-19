package acceptance

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Application struct {
	Type            string
	Chart           string
	Source          string
	Path            string
	SyncInterval    string
	Name            string
	Namespace       string
	Tenant          string
	TargetNamespace string
	DeploymentName  string
	Version         string
	ValuesRegex     string
	Values          string
	Layer           string
}

type ApplicationEvent struct {
	Reason    string
	Message   string
	Component string
	Timestamp string
}

type PullRequest struct {
	Branch      string
	Title       string
	Message     string
	Description string
}

func createGitKustomization(kustomizationName, kustomizationNameSpace, kustomizationPath, repoName, sourceNameSpace, targetNamespace string) (kustomization string) {
	contents, err := ioutil.ReadFile(path.Join(getCheckoutRepoPath(), "test", "utils", "data", "git-kustomization.yaml"))
	gomega.Expect(err).To(gomega.BeNil(), "Failed to read git-kustomization template yaml")

	t := template.Must(template.New("kustomization").Parse(string(contents)))

	type TemplateInput struct {
		KustomizationName      string
		KustomizationNameSpace string
		KustomizationPath      string
		GitRepoName            string
		SourceNameSpace        string
		TargetNamespace        string
	}
	input := TemplateInput{kustomizationName, kustomizationNameSpace, kustomizationPath, repoName, sourceNameSpace, targetNamespace}

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

func AddKustomizationApp(application *pages.AddApplication, app Application) {
	ginkgo.By(fmt.Sprintf("And add %s application from %s GitRepository path", app.Name, app.Path), func() {
		gomega.Expect(application.Name.SendKeys(app.Name)).Should(gomega.Succeed(), "Failed to input application name")
		gomega.Expect(application.Namespace.SendKeys(app.Namespace)).Should(gomega.Succeed(), "Failed to input application namespace")
		gomega.Expect(application.TargetNamespace.SendKeys(app.TargetNamespace)).Should(gomega.Succeed(), "Failed to input application target namespace")
		gomega.Expect(application.Path.SendKeys(app.Path)).Should(gomega.Succeed(), "Failed to input application path")

		if source, _ := application.Source.Attribute("value"); source != "" {
			gomega.Expect(source).Should(gomega.MatchRegexp(app.Source), "Application source GitRepository is incorrect")
		}
	})
}

func AddHelmReleaseApp(profile *pages.ProfileInformation, app Application) {
	ginkgo.By(fmt.Sprintf("And add %s profile/application from %s HelmRepository", app.Name, app.Chart), func() {
		gomega.Eventually(profile.Name.Click, ASSERTION_2MINUTE_TIME_OUT).Should(gomega.Succeed(), fmt.Sprintf("Failed to find %s profile", app.Name))
		gomega.Eventually(profile.Checkbox.Check).Should(gomega.Succeed(), fmt.Sprintf("Failed to select the %s profile", app.Name))

		gomega.Eventually(profile.Version.Click).Should(gomega.Succeed())
		gomega.Eventually(pages.GetOption(webDriver, app.Version).Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to select %s version: %s", app.Name, app.Version))

		if app.Layer != "" {
			gomega.Eventually(profile.Layer.Text).Should(gomega.MatchRegexp(app.Layer))
		}

		gomega.Expect(profile.Namespace.SendKeys(app.TargetNamespace)).To(gomega.Succeed())

		gomega.Eventually(profile.Values.Click).Should(gomega.Succeed())
		valuesYaml := pages.GetValuesYaml(webDriver)
		gomega.Eventually(valuesYaml.Title.Text).Should(gomega.MatchRegexp(app.Name))
		gomega.Eventually(valuesYaml.TextArea.Text).Should(gomega.MatchRegexp(strings.Split(app.ValuesRegex, ",")[0]))

		text, _ := valuesYaml.TextArea.Text()
		for i, val := range strings.Split(app.Values, ",") {
			text = strings.ReplaceAll(text, strings.Split(app.ValuesRegex, ",")[i], val)
		}

		gomega.Expect(valuesYaml.TextArea.Clear()).To(gomega.Succeed())
		gomega.Expect(valuesYaml.TextArea.SendKeys(text)).To(gomega.Succeed(), fmt.Sprintf("Failed to change values.yaml for %s profile", app.Name))

		gomega.Eventually(valuesYaml.Save.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to save values.yaml for %s profile", app.Name))
	})
}

func verifyAppInformation(applicationsPage *pages.ApplicationsPage, app Application, cluster ClusterConfig, status string) {

	ginkgo.By(fmt.Sprintf("And verify %s application information in application table for cluster: %s", app.Name, cluster.Name), func() {
		applicationInfo := applicationsPage.FindApplicationInList(app.Name)

		if app.Type == "helm_release" {
			gomega.Eventually(applicationInfo.Type).Should(matchers.MatchText("HelmRelease"), fmt.Sprintf("Failed to have expected %s application type: %s", app.Name, app.Type))
		} else {
			gomega.Eventually(applicationInfo.Type).Should(matchers.MatchText("Kustomization"), fmt.Sprintf("Failed to have expected %s application type: %s", app.Name, app.Type))
		}

		gomega.Eventually(applicationInfo.Name).Should(matchers.MatchText(app.Name), fmt.Sprintf("Failed to list %s application in  application table", app.Name))
		gomega.Eventually(applicationInfo.Namespace).Should(matchers.MatchText(app.Namespace), fmt.Sprintf("Failed to have expected %s application namespace: %s", app.Name, app.Namespace))
		gomega.Eventually(applicationInfo.Cluster).Should(matchers.MatchText(path.Join(cluster.Namespace, cluster.Name)), fmt.Sprintf("Failed to have expected %s application cluster: %s", app.Name, path.Join(cluster.Namespace, cluster.Name)))
		gomega.Eventually(applicationInfo.Source).Should(matchers.MatchText(app.Source), fmt.Sprintf("Failed to have expected %s application source: %s", app.Name, app.Source))
		gomega.Eventually(applicationInfo.Status, ASSERTION_2MINUTE_TIME_OUT).Should(matchers.MatchText(status), fmt.Sprintf("Failed to have expected %s application status: %s", app.Name, status))

		if app.Tenant != "" {
			gomega.Eventually(applicationInfo.Tenant).Should(matchers.MatchText(app.Tenant), fmt.Sprintf("Failed to have expected %s tenant", app.Tenant))
		}
	})
}

func verifyAppPage(app Application) {
	appDetailPage := pages.GetApplicationsDetailPage(webDriver, app.Type)

	ginkgo.By(fmt.Sprintf("And verify %s application page", app.Name), func() {
		gomega.Eventually(appDetailPage.Header.Text).Should(gomega.MatchRegexp(app.Name), fmt.Sprintf("Failed to verify dashboard application name %s", app.Name))
		gomega.Eventually(appDetailPage.Title.Text).Should(gomega.MatchRegexp(app.Name), fmt.Sprintf("Failed to verify application title %s on application page", app.Name))
		gomega.Eventually(appDetailPage.Sync).Should(matchers.BeEnabled(), fmt.Sprintf("Sync button is not visible/enable for %s", app.Name))
		gomega.Eventually(appDetailPage.Details).Should(matchers.BeEnabled(), fmt.Sprintf("Details tab button is not visible/enable for %s", app.Name))
		gomega.Eventually(appDetailPage.Events).Should(matchers.BeEnabled(), fmt.Sprintf("Events tab button is not visible/enable for %s", app.Name))
		gomega.Eventually(appDetailPage.Graph).Should(matchers.BeEnabled(), fmt.Sprintf("Graph tab button is not visible/enable for %s", app.Name))
		gomega.Eventually(appDetailPage.Violations).Should(matchers.BeEnabled(), fmt.Sprintf("Violations tab button is not visible/enable for %s", app.Name))
	})
}

func verifyAppDetails(app Application, cluster ClusterConfig) {
	appDetailPage := pages.GetApplicationsDetailPage(webDriver, app.Type)

	ginkgo.By(fmt.Sprintf("And verify %s application Details", app.Name), func() {
		gomega.Expect(appDetailPage.Details.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to click %s Details tab button", app.Name))
		pages.WaitForPageToLoad(webDriver)

		details := pages.GetApplicationDetail(webDriver)

		if app.Type == "helm_release" {
			gomega.Eventually(details.Source.Text).Should(gomega.MatchRegexp("HelmChart/"+app.Namespace+"-"+app.Name), fmt.Sprintf("Failed to verify %s Source", app.Name))
			gomega.Eventually(details.Chart.Text).Should(gomega.MatchRegexp(app.Name), fmt.Sprintf("Failed to verify %s Chart", app.Name))
			gomega.Eventually(details.ChartVersion.Text).Should(gomega.MatchRegexp(app.Version), fmt.Sprintf("Failed to verify %s Chart Version", app.Name))

			gomega.Eventually(func(g gomega.Gomega) (string, error) {
				g.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
				time.Sleep(POLL_INTERVAL_1SECONDS)
				return details.AppliedRevision.Text()
			}, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.MatchRegexp(app.Version), fmt.Sprintf("Failed to verify %s Last Applied Version", app.Name))

			gomega.Eventually(details.AttemptedRevision.Text, ASSERTION_30SECONDS_TIME_OUT).Should(gomega.MatchRegexp(app.Version), fmt.Sprintf("Failed to verify %s Last Attempted Version", app.Name))

		} else {
			gomega.Eventually(details.Source.Text).Should(gomega.MatchRegexp("GitRepository/"+app.Name), fmt.Sprintf("Failed to verify %s Source", app.Name))
			gomega.Eventually(details.AppliedRevision.Text).Should(gomega.MatchRegexp("master"), fmt.Sprintf("Failed to verify %s AppliedRevision", app.Name))
			gomega.Eventually(details.Path.Text).Should(gomega.MatchRegexp(app.Path), fmt.Sprintf("Failed to verify %s Path", app.Name))
		}

		if cluster.Type == "management" {
			gomega.Eventually(details.Cluster.Text).Should(gomega.MatchRegexp(cluster.Name), fmt.Sprintf("Failed to verify %s Cluster", app.Name))
		} else {
			gomega.Eventually(details.Cluster.Text).Should(gomega.MatchRegexp(cluster.Namespace+"/"+cluster.Name), fmt.Sprintf("Failed to verify %s Cluster", app.Name))
		}

		if app.Tenant != "" {
			gomega.Eventually(details.Tenant.Text).Should(gomega.MatchRegexp(app.Tenant), fmt.Sprintf("Failed to verify %s Tenant", app.Tenant))
		}

		gomega.Eventually(details.Interval.Text).Should(gomega.MatchRegexp(app.SyncInterval), fmt.Sprintf("Failed to verify %s AppliedRevision", app.Name))
		gomega.Eventually(appDetailPage.Sync.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to sync %s kustomization", app.Name))
		gomega.Eventually(details.LastUpdated.Text).Should(gomega.MatchRegexp("seconds ago"), fmt.Sprintf("Failed to verify %s LastUpdated", app.Name))

		gomega.Eventually(details.Name.Text).Should(gomega.MatchRegexp(app.DeploymentName), fmt.Sprintf("Failed to verify %s Deployment name", app.Name))
		gomega.Eventually(details.Type.Text).Should(gomega.MatchRegexp("Deployment"), fmt.Sprintf("Failed to verify %s Type", app.Name))
		gomega.Eventually(details.Namespace.Text).Should(gomega.MatchRegexp(app.TargetNamespace), fmt.Sprintf("Failed to verify %s Namespace", app.Name))

		gomega.Eventually(func(g gomega.Gomega) (string, error) {
			g.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
			time.Sleep(POLL_INTERVAL_1SECONDS)
			return details.Status.Text()
		}, ASSERTION_3MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.MatchRegexp("Ready"), fmt.Sprintf("Failed to verify %s Status", app.Name))
		gomega.Eventually(details.Message.Text).Should(gomega.MatchRegexp("Deployment is available"), fmt.Sprintf("Failed to verify %s Message", app.Name))
	})
}

func verifyAppAnnotations(app Application) {
	details := pages.GetApplicationDetail(webDriver)

	gomega.Expect(details.GetMetadata("Description").Text()).Should(gomega.MatchRegexp(`Podinfo is a tiny web application made with Go`), "Failed to verify Metada description")
	verifyDashboard(details.GetMetadata("Grafana Dashboard").Find("a"), "management", "Grafana")
	gomega.Expect(details.GetMetadata("Javascript Alert").Find("a")).ShouldNot(matchers.BeFound(), "Javascript href is not sanitized")
	gomega.Expect(details.GetMetadata("Javascript Alert").Text()).Should(gomega.MatchRegexp(`javascript:alert\('hello there'\);`), "Failed to verify Javascript alert text")
}

func verifyAppEvents(app Application, appEvent ApplicationEvent) {
	appDetailPage := pages.GetApplicationsDetailPage(webDriver, app.Type)

	ginkgo.By(fmt.Sprintf("And verify %s application Events", app.Name), func() {
		gomega.Expect(appDetailPage.Events.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to click %s Events tab button", app.Name))
		pages.WaitForPageToLoad(webDriver)

		event := pages.GetApplicationEvent(webDriver, appEvent.Reason)

		gomega.Eventually(event.Reason.Text).Should(gomega.MatchRegexp(cases.Title(language.English, cases.NoLower).String(appEvent.Reason)), fmt.Sprintf("Failed to verify %s Event/Reason", app.Name))
		gomega.Eventually(event.Message.Text).Should(gomega.MatchRegexp(appEvent.Message), fmt.Sprintf("Failed to verify %s Event/Message", app.Name))
		gomega.Eventually(event.Component.Text).Should(gomega.MatchRegexp(appEvent.Component), fmt.Sprintf("Failed to verify %s Event/Component", app.Name))
		gomega.Eventually(event.TimeStamp.Text).Should(gomega.MatchRegexp(appEvent.Timestamp), fmt.Sprintf("Failed to verify %s Event/Timestamp", app.Name))
	})
}

func verfifyAppGraph(app Application) {
	appDetailPage := pages.GetApplicationsDetailPage(webDriver, app.Type)

	ginkgo.By(fmt.Sprintf("And verify %s application Grapg", app.Name), func() {
		gomega.Expect(appDetailPage.Graph.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to click %s Graph tab button", app.Name))
		pages.WaitForPageToLoad(webDriver)

		graph := pages.GetApplicationGraph(webDriver)

		if app.Type == "helm_release" {
			gomega.Expect(graph.HelmRepository.FirstByXPath(fmt.Sprintf(`//div[.="%s"]`, app.Chart))).Should(matchers.BeFound(), fmt.Sprintf("Failed to verify %s Graph/HelmRepository", app.Name))
			gomega.Expect(graph.HelmRelease.FirstByXPath(fmt.Sprintf(`//div[.="%s"]`, app.Namespace))).Should(matchers.BeFound(), fmt.Sprintf("Failed to verify %s Graph/Helmrelease", app.Name))
		} else {
			gomega.Expect(graph.GitRepository.FirstByXPath(fmt.Sprintf(`//div[.="%s"]`, app.Chart))).Should(matchers.BeFound(), fmt.Sprintf("Failed to verify %s Graph/GitRepository", app.Name))
			gomega.Expect(graph.Kustomization.FirstByXPath(fmt.Sprintf(`//div[.="%s"]`, app.Namespace))).Should(matchers.BeFound(), fmt.Sprintf("Failed to verify %s Graph/Kustomization", app.Name))
		}

		gomega.Expect(graph.Deployment.FirstByXPath(fmt.Sprintf(`//div[.="%s"]`, app.Name))).Should(matchers.BeFound(), fmt.Sprintf("Failed to verify %s Graph/Deployment", app.Name))
		gomega.Expect(graph.Deployment.FirstByXPath(fmt.Sprintf(`//div[.="%s"]`, app.TargetNamespace))).Should(matchers.BeFound(), fmt.Sprintf("Failed to verify %s Graph/Deployment namespace", app.Name))
		gomega.Expect(graph.ReplicaSet.FirstByXPath(fmt.Sprintf(`//div[.="%s"]`, app.Name))).Should(matchers.BeFound(), fmt.Sprintf("Failed to verify %s Graph/ReplicaSet", app.Name))
		gomega.Expect(graph.ReplicaSet.FirstByXPath(fmt.Sprintf(`//div[.="%s"]`, app.TargetNamespace))).Should(matchers.BeFound(), fmt.Sprintf("Failed to verify %s Graph/ReplicaSet namespace", app.Name))
		gomega.Expect(graph.Pod.FirstByXPath(fmt.Sprintf(`//div[.="%s"]`, app.Name))).Should(matchers.BeFound(), fmt.Sprintf("Failed to verify %s Graph/Pod", app.Name))
		gomega.Expect(graph.Pod.FirstByXPath(fmt.Sprintf(`//div[.="%s"]`, app.TargetNamespace))).Should(matchers.BeFound(), fmt.Sprintf("Failed to verify %s Graph/Pod namespace", app.Name))
	})
}

func verifyAppSourcePage(applicationInfo *pages.ApplicationInformation, app Application) {
	ginkgo.By(fmt.Sprintf("And navigate directly to %s Sources page", app.Name), func() {
		gomega.Expect(applicationInfo.Source.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to %s Sources pages directly", app.Name))

		sourceDetailPage := pages.GetSourceDetailPage(webDriver)
		gomega.Eventually(sourceDetailPage.Header.Text).Should(gomega.MatchRegexp(app.Source), fmt.Sprintf("Failed to verify dashboard header source name %s ", app.Source))
		gomega.Eventually(sourceDetailPage.Title.Text).Should(gomega.MatchRegexp(app.Source), fmt.Sprintf("Failed to verify application title %s on source page", app.Source))
	})
}

func verifyDeleteApplication(applicationsPage *pages.ApplicationsPage, existingAppCount int, appName, appKustomization string) {
	navigatetoApplicationsPage(applicationsPage)

	if appKustomization != "" {
		ginkgo.By(fmt.Sprintf("And delete the %s kustomization and source maifest from the repository's master branch", appName), func() {
			cleanGitRepository(appKustomization)
		})
	}

	ginkgo.By(fmt.Sprintf("And wait for %s application to dissappeare from the dashboard", appName), func() {
		gomega.Eventually(func(g gomega.Gomega) int {
			g.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
			time.Sleep(POLL_INTERVAL_1SECONDS)
			return applicationsPage.CountApplications()
		}, ASSERTION_3MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Equal(existingAppCount), fmt.Sprintf("There should be %d application enteries in application table", existingAppCount))
	})
}

func createGitopsPR(pullRequest PullRequest) {
	ginkgo.By("And set GitOps values for pull request", func() {
		gitops := pages.GetGitOps(webDriver)
		gomega.Eventually(gitops.GitOpsLabel).Should(matchers.BeFound())

		pages.ClearFieldValue(gitops.BranchName)
		gomega.Expect(gitops.BranchName.SendKeys(pullRequest.Branch)).To(gomega.Succeed())
		pages.ClearFieldValue(gitops.PullRequestTile)
		gomega.Expect(gitops.PullRequestTile.SendKeys(pullRequest.Title)).To(gomega.Succeed())
		pages.ClearFieldValue(gitops.CommitMessage)
		gomega.Expect(gitops.CommitMessage.SendKeys(pullRequest.Message)).To(gomega.Succeed())

		AuthenticateWithGitProvider(webDriver, gitProviderEnv.Type, gitProviderEnv.Hostname)
		gomega.Eventually(gitops.GitCredentials).Should(matchers.BeVisible())
		gomega.Eventually(gitops.CreatePR.Click()).Should(gomega.Succeed(), "Failed to create pull request")
	})
}

func DescribeApplications(gitopsTestRunner GitopsTestRunner) {
	var _ = ginkgo.Describe("Multi-Cluster Control Plane Applications", func() {

		ginkgo.BeforeEach(func() {
			gomega.Expect(webDriver.Navigate(test_ui_url)).To(gomega.Succeed())

			if !pages.ElementExist(pages.Navbar(webDriver).Title, 3) {
				loginUser()
			}
		})

		ginkgo.Context("[UI] When no applications are installed", func() {
			ginkgo.It("Verify management cluster dashboard shows bootstrap 'flux-system' application", ginkgo.Label("integration"), func() {
				fluxSystem := Application{
					Type:      "Kustomization",
					Chart:     "weaveworks-charts",
					Name:      "flux-system",
					Namespace: GITOPS_DEFAULT_NAMESPACE,
					Source:    "flux-system",
				}

				mgmtCluster := ClusterConfig{
					Type:      "management",
					Name:      "management",
					Namespace: "",
				}

				pages.NavigateToPage(webDriver, "Applications")

				ginkgo.By("And wait for  good looking response from /v1/objects", func() {
					gomega.Expect(waitForGitopsResources(context.Background(), "objects?kind=Kustomization", POLL_INTERVAL_15SECONDS)).To(gomega.Succeed(), "Failed to get a successful response from /v1/objects")
				})

				applicationsPage := pages.GetApplicationsPage(webDriver)
				pages.WaitForPageToLoad(webDriver)

				ginkgo.By("And wait for Applications page to be rendered", func() {
					gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())
					gomega.Expect(applicationsPage.CountApplications()).To(gomega.Equal(1), "There should not be any cluster in cluster table")
				})

				verifyAppInformation(applicationsPage, fluxSystem, mgmtCluster, "Ready")
			})
		})

		ginkgo.Context("[UI] Applications(s) can be installed", func() {

			var existingAppCount int
			appNameSpace := "test-kustomization"
			appTargetNamespace := "test-system"

			mgmtCluster := ClusterConfig{
				Type:      "management",
				Name:      "management",
				Namespace: "",
			}

			ginkgo.JustBeforeEach(func() {
				createNamespace([]string{appNameSpace, appTargetNamespace})
			})

			ginkgo.JustAfterEach(func() {
				// Wait for the application to be deleted gracefully, needed when the test fails before deleting the application
				gomega.Eventually(func(g gomega.Gomega) int {
					return getApplicationCount()
				}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Equal(existingAppCount), fmt.Sprintf("There should be %d application enteries after application(s) deletion", existingAppCount))

				deleteNamespace([]string{appNameSpace, appTargetNamespace})
			})

			ginkgo.It("Verify application with annotations/metadata can be installed  and dashboard is updated accordingly", ginkgo.Label("integration", "application", "browser-logs"), func() {

				podinfo := Application{
					Type:            "kustomization",
					Name:            "my-podinfo",
					DeploymentName:  "podinfo",
					Namespace:       appNameSpace,
					TargetNamespace: appTargetNamespace,
					Source:          "my-podinfo",
					Path:            "./kustomize",
					SyncInterval:    "30s",
				}

				sourceURL := "https://github.com/stefanprodan/podinfo"
				addSource("git", podinfo.Source, podinfo.Namespace, sourceURL, "master", "")

				appDir := fmt.Sprintf("./clusters/%s/podinfo", mgmtCluster.Name)
				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
				existingAppCount = getApplicationCount()

				appKustomization := createGitKustomization(podinfo.Name, podinfo.Namespace, podinfo.Path, podinfo.Source, podinfo.Namespace, podinfo.TargetNamespace)
				defer deleteSource("git", podinfo.Source, podinfo.Namespace, "")
				defer cleanGitRepository(appDir)

				pages.NavigateToPage(webDriver, "Applications")
				applicationsPage := pages.GetApplicationsPage(webDriver)

				ginkgo.By("And add Kustomization & GitRepository Source manifests pointing to podinfo repository’s master branch)", func() {

					pullGitRepo(repoAbsolutePath)
					_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("mkdir -p %[2]v && cp -f %[1]v %[2]v", appKustomization, path.Join(repoAbsolutePath, appDir)))
					gitUpdateCommitPush(repoAbsolutePath, "Adding podinfo kustomization")
				})

				ginkgo.By("And wait for podinfo application to be visibe on the dashboard", func() {
					gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())

					totalAppCount := existingAppCount + 1

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
				verifyAppDetails(podinfo, mgmtCluster)
				verifyAppAnnotations(podinfo)

				navigatetoApplicationsPage(applicationsPage)
				verifyAppSourcePage(applicationInfo, podinfo)

				verifyDeleteApplication(applicationsPage, existingAppCount, podinfo.Name, appDir)
			})

			ginkgo.It("Verify application can be installed from HelmRepository source and dashboard is updated accordingly", ginkgo.Label("integration", "application", "browser-logs"), func() {
				metallb := Application{
					Type:            "helm_release",
					Chart:           "weaveworks-charts",
					SyncInterval:    "10m",
					Name:            "metallb",
					DeploymentName:  "metallb-controller",
					Namespace:       GITOPS_DEFAULT_NAMESPACE, // HelmRelease application always get installed in flux-system namespace
					TargetNamespace: appNameSpace,
					Source:          GITOPS_DEFAULT_NAMESPACE + "-metallb",
					Version:         "0.0.2",
					ValuesRegex:     `namespace: ""`,
					Values:          fmt.Sprintf(`namespace: %s`, appNameSpace),
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
				sourceURL := "https://raw.githubusercontent.com/weaveworks/profiles-catalog/gh-pages"
				appKustomization := fmt.Sprintf("./clusters/%s/%s-%s-helmrelease.yaml", mgmtCluster.Name, metallb.Name, appNameSpace)

				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
				existingAppCount = getApplicationCount()

				defer cleanGitRepository(appKustomization)

				ginkgo.By("And wait for cluster-service to cache profiles", func() {
					gomega.Expect(waitForGitopsResources(context.Background(), "profiles", POLL_INTERVAL_5SECONDS, ASSERTION_15MINUTE_TIME_OUT)).To(gomega.Succeed(), "Failed to get a successful response from /v1/profiles ")
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

			ginkgo.It("Verify application can be installed from GitRepository source and dashboard is updated accordingly", ginkgo.Label("integration", "application", "browser-logs"), func() {
				podinfo := Application{
					Type:            "kustomization",
					Name:            "my-podinfo",
					DeploymentName:  "podinfo",
					Namespace:       appNameSpace,
					TargetNamespace: appTargetNamespace,
					Source:          "my-podinfo",
					Path:            "./kustomize",
					SyncInterval:    "10m",
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

				sourceURL := "https://github.com/stefanprodan/podinfo"
				appKustomization := fmt.Sprintf("./clusters/%s/%s-%s-kustomization.yaml", mgmtCluster.Name, podinfo.Name, podinfo.Namespace)

				defer deleteSource("git", podinfo.Source, podinfo.Namespace, "")
				defer cleanGitRepository(appKustomization)

				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
				existingAppCount = getApplicationCount()

				pages.NavigateToPage(webDriver, "Applications")
				applicationsPage := pages.GetApplicationsPage(webDriver)

				addSource("git", podinfo.Source, podinfo.Namespace, sourceURL, "master", "")
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
				verifyAppDetails(podinfo, mgmtCluster)
				verfifyAppGraph(podinfo)

				navigatetoApplicationsPage(applicationsPage)
				verifyAppSourcePage(applicationInfo, podinfo)

				verifyDeleteApplication(applicationsPage, existingAppCount, podinfo.Name, appKustomization)
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

			appNameSpace := "test-kustomization"
			appTargetNamespace := "test-system"

			leafCluster := ClusterConfig{
				Type:      "other",
				Name:      "wge-leaf-application-kind",
				Namespace: "test-system",
			}

			ginkgo.JustBeforeEach(func() {
				existingAppCount = getApplicationCount()
				appDir = path.Join("clusters", leafCluster.Namespace, leafCluster.Name, "apps")
				mgmtClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
				createCluster("kind", leafCluster.Name, "")
				createNamespace([]string{appNameSpace, appTargetNamespace})
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

			ginkgo.It("Verify application can be installed from GitRepository source on leaf cluster and management dashboard is updated accordingly", ginkgo.Label("integration", "application", "leaf-application"), func() {
				podinfo := Application{
					Type:            "kustomization",
					Name:            "my-podinfo",
					DeploymentName:  "podinfo",
					Namespace:       appNameSpace,
					TargetNamespace: appTargetNamespace,
					Source:          "my-podinfo",
					Path:            "./kustomize",
					SyncInterval:    "10m",
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

				useClusterContext(mgmtClusterContext)
				createNamespace([]string{leafCluster.Namespace})

				createPATSecret(leafCluster.Namespace, patSecret)
				clusterBootstrapCopnfig = createClusterBootstrapConfig(leafCluster.Name, leafCluster.Namespace, bootstrapLabel, patSecret)
				gitopsCluster = connectGitopsCuster(leafCluster.Name, leafCluster.Namespace, bootstrapLabel, leafClusterkubeconfig)
				createLeafClusterSecret(leafCluster.Namespace, leafClusterkubeconfig)

				ginkgo.By("Verify GitopsCluster status after creating kubeconfig secret", func() {
					pages.NavigateToPage(webDriver, "Clusters")
					clustersPage := pages.GetClustersPage(webDriver)
					pages.WaitForPageToLoad(webDriver)
					clusterInfo := clustersPage.FindClusterInList(leafCluster.Name)

					gomega.Eventually(clusterInfo.Status, ASSERTION_30SECONDS_TIME_OUT).Should(matchers.MatchText("Ready"))
				})

				addKustomizationBases("leaf", leafCluster.Name, leafCluster.Namespace)

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

					existingAppCount += 2 // flux-system + clusters-bases-kustomization (leaf cluster)
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

					gomega.Eventually(func(g gomega.Gomega) int {
						return applicationsPage.CountApplications()
					}, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(totalAppCount), fmt.Sprintf("There should be %d application enteries in application table", totalAppCount))
				})

				ginkgo.By(fmt.Sprintf("And search leaf cluster '%s' app", leafCluster.Name), func() {
					searchPage := pages.GetSearchPage(webDriver)
					gomega.Eventually(searchPage.SearchBtn.Click).Should(gomega.Succeed(), "Failed to click search buttton")
					gomega.Expect(searchPage.Search.SendKeys(podinfo.Name)).Should(gomega.Succeed(), "Failed to type violation name in search field")
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
				verifyAppDetails(podinfo, leafCluster)
				verfifyAppGraph(podinfo)

				navigatetoApplicationsPage(applicationsPage)
				verifyAppSourcePage(applicationInfo, podinfo)

				verifyDeleteApplication(applicationsPage, existingAppCount, podinfo.Name, appKustomization)
			})

			ginkgo.It("Verify application can be installed from HelmRepository source on leaf cluster and management dashboard is updated accordingly", ginkgo.Label("integration", "application", "leaf-application"), func() {
				ginkgo.Skip("Test is waiting for #1282 to be fixed. Can't get profile from leaf clusters")

				metallb := Application{
					Type:            "helm_release",
					Chart:           "profiles-catalog",
					SyncInterval:    "10m",
					Name:            "metallb",
					DeploymentName:  "metallb-controller",
					Namespace:       appNameSpace,
					TargetNamespace: appTargetNamespace,
					Source:          appNameSpace + "-metallb",
					Version:         "0.0.2",
					ValuesRegex:     `namespace: ""`,
					Values:          fmt.Sprintf(`namespace: %s`, appNameSpace),
				}

				appEvent := ApplicationEvent{
					Reason:    "info",
					Message:   "Helm install has started|Helm install succeeded",
					Component: "helm-controller",
					Timestamp: "seconds|minutes|minute ago",
				}

				pullRequest := PullRequest{
					Branch:  "management-helm-leaf-cluster-apps",
					Title:   "Management Helm Leaf Cluster Application",
					Message: "Adding management helm leaf cluster applications",
				}

				sourceURL := "https://raw.githubusercontent.com/weaveworks/profiles-catalog/gh-pages"
				appKustomization := fmt.Sprintf("./clusters/%s/%s/%s-%s-kustomization.yaml", leafCluster.Namespace, leafCluster.Name, metallb.Name, metallb.Namespace)

				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
				leafClusterkubeconfig = createLeafClusterKubeconfig(leafClusterContext, leafCluster.Name, leafCluster.Namespace)

				useClusterContext(mgmtClusterContext)
				createNamespace([]string{leafCluster.Namespace})

				createPATSecret(leafCluster.Namespace, patSecret)
				clusterBootstrapCopnfig = createClusterBootstrapConfig(leafCluster.Name, leafCluster.Namespace, bootstrapLabel, patSecret)
				gitopsCluster = connectGitopsCuster(leafCluster.Name, leafCluster.Namespace, bootstrapLabel, leafClusterkubeconfig)
				createLeafClusterSecret(leafCluster.Namespace, leafClusterkubeconfig)

				ginkgo.By("Verify GitopsCluster status after creating kubeconfig secret", func() {
					pages.NavigateToPage(webDriver, "Clusters")
					clustersPage := pages.GetClustersPage(webDriver)
					pages.WaitForPageToLoad(webDriver)
					clusterInfo := clustersPage.FindClusterInList(leafCluster.Name)

					gomega.Eventually(clusterInfo.Status, ASSERTION_30SECONDS_TIME_OUT).Should(matchers.MatchText("Ready"))
				})

				addKustomizationBases("leaf", leafCluster.Name, leafCluster.Namespace)

				ginkgo.By(fmt.Sprintf("And I verify %s GitopsCluster/leafCluster is bootstraped)", leafCluster.Name), func() {
					useClusterContext(leafClusterContext)
					verifyFluxControllers(GITOPS_DEFAULT_NAMESPACE)
					waitForGitRepoReady("flux-system", GITOPS_DEFAULT_NAMESPACE)
				})

				// Add HelmRepository source to leaf cluster
				addSource("helm", metallb.Chart, metallb.Namespace, sourceURL, "", "")
				useClusterContext(mgmtClusterContext)

				ginkgo.By("And wait for cluster-service to cache profiles", func() {
					gomega.Expect(waitForGitopsResources(context.Background(), "profiles", POLL_INTERVAL_5SECONDS, ASSERTION_15MINUTE_TIME_OUT)).To(gomega.Succeed(), "Failed to get a successful response from /v1/profiles ")
				})

				pages.NavigateToPage(webDriver, "Applications")
				applicationsPage := pages.GetApplicationsPage(webDriver)

				ginkgo.By("And wait for existing applications to be visibe on the dashboard", func() {
					gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())

					existingAppCount += 2 // flux-system + clusters-bases-kustomization (leaf cluster)
				})

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
						g.Eventually(application.SelectListItem(webDriver, leafCluster.Name).Click).Should(gomega.Succeed(), "Failed to select 'management' cluster from clusters list")
						g.Eventually(application.Source.Click).Should(gomega.Succeed(), "Failed to click Select Source list")
						return pages.ElementExist(application.SelectListItem(webDriver, metallb.Chart))
					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.BeTrue(), fmt.Sprintf("HelmRepository %s source is not listed in source's list", metallb.Name))

					gomega.Expect(pages.ClickElement(webDriver, application.SelectListItem(webDriver, metallb.Chart), -250, 0)).Should(gomega.Succeed(), "Failed to select HelmRepository source from sources list")
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

					totalAppCount := existingAppCount + 1 // metallb (leaf cluster)

					gomega.Eventually(func(g gomega.Gomega) int {
						return applicationsPage.CountApplications()
					}, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(totalAppCount), fmt.Sprintf("There should be %d application enteries in application table", totalAppCount))
				})

				ginkgo.By(fmt.Sprintf("And search leaf cluster '%s' app", leafCluster.Name), func() {
					searchPage := pages.GetSearchPage(webDriver)
					gomega.Eventually(searchPage.SearchBtn.Click).Should(gomega.Succeed(), "Failed to click search buttton")
					gomega.Expect(searchPage.Search.SendKeys(metallb.Name)).Should(gomega.Succeed(), "Failed to type violations name in search field")
					gomega.Expect(searchPage.Search.SendKeys("\uE007")).Should(gomega.Succeed()) // send enter key code to do application search in table

					gomega.Eventually(func(g gomega.Gomega) int {
						return applicationsPage.CountApplications()
					}).Should(gomega.Equal(1), "There should be '1' application entery in application table after search")
				})

				verifyAppInformation(applicationsPage, metallb, leafCluster, "Ready")

				applicationInfo := applicationsPage.FindApplicationInList(metallb.Name)
				ginkgo.By(fmt.Sprintf("And navigate to %s application page", metallb.Name), func() {
					gomega.Eventually(applicationInfo.Name.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to %s application detail page", metallb.Name))
				})

				verifyAppPage(metallb)
				verifyAppEvents(metallb, appEvent)
				verifyAppDetails(metallb, leafCluster)
				verfifyAppGraph(metallb)

				navigatetoApplicationsPage(applicationsPage)
				verifyAppSourcePage(applicationInfo, metallb)

				verifyDeleteApplication(applicationsPage, existingAppCount, metallb.Name, appKustomization)
			})

		})

		// Application Violations tests
		ginkgo.Context("[UI] Application violations are available for management cluster", func() {
			// Count of existing applications before deploying new application
			var existingAppCount int

			// Just specify policies yaml path
			policiesYaml := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "policies.yaml")

			// Just specify the Violating policy info to create it
			policyName := "Container Running As Root acceptance test"
			violationMsg := `Container Running As Root acceptance test in deployment podinfo \(1 occurrences\)`
			violationSeverity := "High"
			violationCategory := "weave.categories.pod-security"

			// Just specify the violated application info to create it
			appNameSpace := "test-kustomization"
			appTargetNamespace := "test-system"

			mgmtCluster := ClusterConfig{
				Type:      "management",
				Name:      "management",
				Namespace: "",
			}

			ginkgo.JustBeforeEach(func() {
				createNamespace([]string{appNameSpace, appTargetNamespace})

				// Add/Install test Policies to the management cluster
				installTestPolicies("management", policiesYaml)
			})

			ginkgo.JustAfterEach(func() {
				// Wait for the application to be deleted gracefully, needed when the test fails before deleting the application
				gomega.Eventually(func(g gomega.Gomega) int {
					return getApplicationCount()
				}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Equal(existingAppCount), fmt.Sprintf("There should be %d application enteries after application(s) deletion", existingAppCount))

				_ = gitopsTestRunner.KubectlDelete([]string{}, policiesYaml)
				deleteNamespace([]string{appNameSpace, appTargetNamespace})
			})

			ginkgo.FIt("Verify application violations Details page", ginkgo.Label("integration", "application", "violation"), func() {
				// Podinfo application details
				podinfo := Application{
					Type:            "kustomization",
					Name:            "app-violations-podinfo",
					DeploymentName:  "podinfo",
					Namespace:       appNameSpace,
					TargetNamespace: appTargetNamespace,
					Source:          "flux-system",
					Path:            "./apps/podinfo",
					SyncInterval:    "30s",
				}

				appDir := fmt.Sprintf("./clusters/%s/podinfo", mgmtCluster.Name)
				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
				existingAppCount = getApplicationCount()

				appKustomization := createGitKustomization(podinfo.Name, podinfo.Namespace, podinfo.Path, podinfo.Source, GITOPS_DEFAULT_NAMESPACE, podinfo.TargetNamespace)
				defer cleanGitRepository(appDir)

				pages.NavigateToPage(webDriver, "Applications")
				// Declare application page variable
				applicationsPage := pages.GetApplicationsPage(webDriver)

				ginkgo.By("Add Application/Kustomization manifests to management cluster's repository main branch)", func() {
					pullGitRepo(repoAbsolutePath)
					podinfo := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "podinfo-app-violations-manifest.yaml")
					createCommand := fmt.Sprintf("mkdir -p %[2]v && cp -f %[1]v %[2]v", podinfo, path.Join(repoAbsolutePath, "apps/podinfo"))
					err := runCommandPassThrough("sh", "-c", createCommand)
					gomega.Expect(err).Should(gomega.BeNil(), "Failed to run '%s'", createCommand)
					gitUpdateCommitPush(repoAbsolutePath, "Adding podinfo kustomization")
				})

				ginkgo.By("Install kustomization Application on management cluster", func() {
					pullGitRepo(repoAbsolutePath)
					command := fmt.Sprintf("mkdir -p %[2]v && cp -f %[1]v %[2]v", appKustomization, path.Join(repoAbsolutePath, appDir))
					err := runCommandPassThrough("sh", "-c", command)
					gomega.Expect(err).Should(gomega.BeNil(), "Failed to run '%s'", command)
					gitUpdateCommitPush(repoAbsolutePath, "Adding podinfo kustomization")
				})

				ginkgo.By("And wait for podinfo application to be visibe on the dashboard", func() {
					gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())

					totalAppCount := existingAppCount + 1

					gomega.Eventually(func(g gomega.Gomega) int {
						return applicationsPage.CountApplications()
					}, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(totalAppCount), fmt.Sprintf("There should be '%d' application enteries in application table", totalAppCount))
				})

				verifyAppInformation(applicationsPage, podinfo, mgmtCluster, "Ready")

				applicationInfo := applicationsPage.FindApplicationInList(podinfo.Name)

				ginkgo.By(fmt.Sprintf("And navigate to '%s' application page", podinfo.Name), func() {
					gomega.Eventually(applicationInfo.Name.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to '%s' application detail page", podinfo.Name))
					time.Sleep(POLL_INTERVAL_1SECONDS)
				})

				ginkgo.By(fmt.Sprintf("And open  '%s' application Violations tab", podinfo.Name), func() {
					// Declare application details page variable
					appDetailPage := pages.GetApplicationsDetailPage(webDriver, podinfo.Type)
					gomega.Eventually(appDetailPage.Violations.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to click '%s' Violations tab button", podinfo.Name))
					pages.WaitForPageToLoad(webDriver)
					time.Sleep(POLL_INTERVAL_3SECONDS)

				})

				ginkgo.By("And check violations are visible in the Application Violations List", func() {
					gomega.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
					gomega.Expect((webDriver.URL())).Should(gomega.ContainSubstring("/violations?clusterName="))
					appViolationsMsg := pages.GetAppViolationsMsgInList(webDriver)
					gomega.Eventually(appViolationsMsg.AppViolationsMsg.Text, ASSERTION_30SECONDS_TIME_OUT).Should(gomega.MatchRegexp(violationMsg), fmt.Sprintf("Failed to list '%s' violation in '%s' vioilations list", violationMsg, podinfo.Name))
				})

				// Application violation tab test
				ginkgo.By(fmt.Sprintf("Opening  %s application Violations tab", podinfo.Name), func() {
					appDetailPage := pages.GetApplicationsDetailPage(webDriver, podinfo.Type)
					gomega.Eventually(appDetailPage.Violations.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to click %s Violations tab button", podinfo.Name))
					// Checking that application violation are visible.
					gomega.Eventually(appDetailPage.Violations).Should(matchers.BeVisible())
					pages.WaitForPageToLoad(webDriver)
					ApplicationViolationsList := pages.GetApplicationViolationsList(webDriver)
					// Checking the Violation Message in the application violation list.
					gomega.Eventually(ApplicationViolationsList.ViolationMessage.Text).Should(gomega.MatchRegexp("MESSAGE"), "Failed to get Violation Message title in violations List page")
					gomega.Eventually(ApplicationViolationsList.ViolationMessageValue.Text).Should(gomega.MatchRegexp(violationMsg), "Failed to get the Violation Message Value in App violations List")
					// Checking the Severity in the application violation list.
					gomega.Eventually(ApplicationViolationsList.Severity.Text).Should(gomega.MatchRegexp("SEVERITY"), "Failed to get the Severity title in App violations List")
					gomega.Expect(ApplicationViolationsList.SeverityIcon).ShouldNot(gomega.BeNil(), "Failed to get the Severity icon in App violations List")
					gomega.Eventually(ApplicationViolationsList.SeverityValue.Text).Should(gomega.MatchRegexp(violationSeverity), "Failed to get the Severity Value in App violations List")
					// Checking the Violated Poilicy in the application violation list.
					gomega.Eventually(ApplicationViolationsList.ViolatedPolicy.Text).Should(gomega.MatchRegexp("VIOLATED POLICY"), "Failed to get the Violated Policy title in App violations List")
					gomega.Eventually(ApplicationViolationsList.ViolatedPolicyValue.Text).Should(gomega.MatchRegexp(violationMsg), "Failed to get the Violated Policy Value in App violations List")
					// Checking the Violation Time in the application violation list.
					gomega.Eventually(ApplicationViolationsList.ViolationTime.Text).Should(gomega.MatchRegexp("VIOLATION TIME"), "Failed to get the Violation time title in App violations List")
					gomega.Expect(ApplicationViolationsList.ViolationTimeValue.Text()).NotTo(gomega.BeEmpty(), "Failed to get violation time value in App violations List")
					gomega.Expect(ApplicationViolationsList.ViolationTimeValueSorting).ShouldNot(gomega.BeNil(), "Failed to get the Violation Time sorting icon in App violations List")
					pages.WaitForPageToLoad(webDriver)
				})
				// Checking that the Violations can be filtered in the application violation list.
				ginkgo.By("And Violations can be filtered by Severity", func() {
					ApplicationViolationsList := pages.GetApplicationViolationsList(webDriver)
					filterID := "severity: high"
					searchPage := pages.GetSearchPage(webDriver)
					searchPage.SelectFilter("severity", filterID)
					gomega.Eventually(ApplicationViolationsList.CountViolations).Should(gomega.Equal(2), "The number of selected violations for high severity should be equal two")
					// Clear the filter
					searchPage.SelectFilter("severity", filterID)
				})
				// Checking that you can search by violated policy name in the application violation list.
				ginkgo.By(fmt.Sprintf("And search by violated policy name in '%s' app violations list", podinfo.Name), func() {
					ApplicationViolationsList := pages.GetApplicationViolationsList(webDriver)
					searchPage := pages.GetSearchPage(webDriver)
					gomega.Eventually(searchPage.SearchBtn.Click).Should(gomega.Succeed(), "Failed to click search buttton")
					gomega.Expect(searchPage.Search.SendKeys(policyName)).Should(gomega.Succeed(), "Failed to type violated policy name in search field")
					gomega.Expect(searchPage.Search.SendKeys("\uE007")).Should(gomega.Succeed()) // Send enter key code to do violations search in table
					gomega.Eventually(func(g gomega.Gomega) int {
						return ApplicationViolationsList.CountViolations()
					}).Should(gomega.Equal(1), "There should be '1' Violation Message in the list after search")
					gomega.Eventually(ApplicationViolationsList.ViolationMessageValue.Text).Should(gomega.MatchRegexp(violationMsg), "Failed to get the Violation Message Value in App violations List")
					pages.WaitForPageToLoad(webDriver)
				})

				ginkgo.By(fmt.Sprintf("Verify '%s' Application Violation Details", policyName), func() {

					appViolationsMsg := pages.GetAppViolationsMsgInList(webDriver)

					gomega.Eventually(appViolationsMsg.AppViolationsMsg.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to violation detail page of violation '%s'", violationMsg))

					gomega.Expect(webDriver.URL()).Should(gomega.ContainSubstring("/clusters/violations/details?clusterName"))

					appViolationsDetialsPage := pages.GetApplicationViolationsDetailsPage(webDriver)

					gomega.Eventually(appViolationsDetialsPage.ViolationHeader.Text).Should(gomega.MatchRegexp(violationMsg), "Failed to get violation header on App violations details page")
					gomega.Eventually(appViolationsDetialsPage.PolicyName.Text).Should(gomega.MatchRegexp("Policy Name :"), "Failed to get policy name field on App violations details page")
					gomega.Eventually(appViolationsDetialsPage.PolicyNameValue.Text).Should(gomega.MatchRegexp(policyName), "Failed to get policy name value on App violations details page")

					// Click policy name from app violations details page to navigate to policy details page
					gomega.Eventually(appViolationsDetialsPage.PolicyNameValue.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to '%s' policy detail page", appViolationsDetialsPage.PolicyNameValue))
					gomega.Expect(webDriver.URL()).Should(gomega.ContainSubstring("/policies/details?"))

					// Navigate back to the app violations list
					gomega.Expect(webDriver.Back()).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to navigate back to the '%s' app violations list", podinfo.Name))

					gomega.Eventually(appViolationsDetialsPage.ClusterName.Text).Should(gomega.MatchRegexp("Cluster Name :"), "Failed to get cluster name field on App violations details page")
					gomega.Eventually(appViolationsDetialsPage.ClusterNameValue.Text).Should(gomega.MatchRegexp(mgmtCluster.Name), "Failed to get cluster name value on App violations details page")

					gomega.Eventually(appViolationsDetialsPage.ViolationTime.Text).Should(gomega.MatchRegexp("Violation Time :"), "Failed to get violation time field on App violations details page")
					gomega.Expect(appViolationsDetialsPage.ViolationTimeValue.Text()).NotTo(gomega.BeEmpty(), "Failed to get violation time value on App violations details page")

					gomega.Eventually(appViolationsDetialsPage.Severity.Text).Should(gomega.MatchRegexp("Severity :"), "Failed to get severity field on App violations details page")
					gomega.Expect(appViolationsDetialsPage.SeverityIcon).NotTo(gomega.BeNil(), "Failed to get severity icon value on App violations details page")
					gomega.Eventually(appViolationsDetialsPage.SeverityValue.Text).Should(gomega.MatchRegexp(violationSeverity), "Failed to get severity value on App violations details page")

					gomega.Eventually(appViolationsDetialsPage.Category.Text).Should(gomega.MatchRegexp("Category :"), "Failed to get category field on App violations details page")
					gomega.Eventually(appViolationsDetialsPage.CategoryValue.Text).Should(gomega.MatchRegexp(violationCategory), "Failed to get category value on App violations details page")

					gomega.Eventually(appViolationsDetialsPage.Occurrences.Text).Should(gomega.MatchRegexp("Occurrences"), "Failed to get Occurrences field on App violations details page")
					gomega.Eventually(appViolationsDetialsPage.OccurrencesCount.Text).Should(gomega.MatchRegexp("( 1 )"), "Failed to get Occurrences count on App violations details page")
					gomega.Expect(appViolationsDetialsPage.OccurrencesValue.Text()).NotTo(gomega.BeEmpty(), "Failed to get Occurrences value on App violations details page")

					gomega.Eventually(appViolationsDetialsPage.Description.Text).Should(gomega.MatchRegexp("Description:"), "Failed to get description field on App violations details page")
					gomega.Expect(appViolationsDetialsPage.DescriptionValue.Text()).NotTo(gomega.BeEmpty(), "Failed to get description value on App violations details page")

					gomega.Eventually(appViolationsDetialsPage.HowToSolve.Text).Should(gomega.MatchRegexp("How to solve:"), "Failed to get how to resolve field on App violations details page")
					gomega.Expect(appViolationsDetialsPage.HowToSolveValue.Text()).NotTo(gomega.BeEmpty(), "Failed to get how to resolve value on App violations details page")

					gomega.Eventually(appViolationsDetialsPage.ViolatingEntity.Text).Should(gomega.MatchRegexp("Violating Entity:"), "Failed to get violating entity field on App violations details page")
					gomega.Expect(appViolationsDetialsPage.ViolatingEntityValue.Text()).NotTo(gomega.BeEmpty(), "Failed to get violating entity value on App violations details page")
				})

				ginkgo.By(fmt.Sprintf("And finally delete the '%s' application", podinfo.Name), func() {
					verifyDeleteApplication(applicationsPage, existingAppCount, podinfo.Name, appDir)
				})
			})
		})
	})
}
