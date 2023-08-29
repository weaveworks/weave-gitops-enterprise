package acceptance

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Application struct {
	DefaultApp      bool
	Type            string
	Chart           string
	Url             string
	Source          string
	Path            string
	SyncInterval    string
	Name            string
	Namespace       string
	Description     string
	Tenant          string
	TargetNamespace string
	DeploymentName  string
	Version         string
	ValuesRegex     string
	Values          string
	Layer           string
	GitRepository   string
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

type ApplicationViolations struct {
	PolicyName               string
	ViolationMessage         string
	ViolationSeverity        string
	ViolationCategory        string
	ConfigPolicy             string
	PolicyConfigViolationMsg string
}

func navigatetoApplicationsPage(applicationsPage *pages.ApplicationsPage) {
	ginkgo.By("And back to Applications page via header link", func() {
		gomega.Expect(applicationsPage.ApplicationHeaderLink.Click()).Should(gomega.Succeed(), "Failed to navigate to Applications pages via header link")
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
		gomega.Eventually(profile.Name.Click, ASSERTION_1MINUTE_TIME_OUT).Should(gomega.Succeed(), fmt.Sprintf("Failed to find %s profile", app.Name))

		if app.DefaultApp {
			gomega.Eventually(profile.Checkbox).Should(matchers.BeSelected(), fmt.Sprintf("Default profile %s is not selected as default", app.Name))
		} else {
			gomega.Eventually(profile.Checkbox).ShouldNot(matchers.BeSelected(), fmt.Sprintf("Profile %s should not be selected as default", app.Name))
			gomega.Eventually(profile.Checkbox.Check).Should(gomega.Succeed(), fmt.Sprintf("Failed to select the %s profile", app.Name))
		}

		// Work around some flakiness in selecting the version
		gomega.Eventually(profile.Version.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to select expected %s profile version", app.Version))
		gomega.Eventually(pages.GetOption(webDriver, app.Version).Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to select %s version: %s", app.Name, app.Version))
		gomega.Eventually(pages.GetOption(webDriver, app.Version)).ShouldNot(matchers.BeFound(), fmt.Sprintf("Failed to verify that %s version selector is no longer present on the page", app.Name))

		if app.Layer != "" {
			gomega.Eventually(profile.Layer.Text).Should(gomega.MatchRegexp(app.Layer), fmt.Sprintf("Failed to verify expeted %s profile layer", app.Layer))
		}

		gomega.Expect(profile.Namespace.SendKeys(app.TargetNamespace)).To(gomega.Succeed())

		gomega.Eventually(profile.Values.Click).Should(gomega.Succeed())
		valuesYaml := pages.GetValuesYaml(webDriver)
		gomega.Eventually(valuesYaml.Title.Text, ASSERTION_30SECONDS_TIME_OUT).Should(gomega.MatchRegexp(app.Name))
		gomega.Eventually(valuesYaml.TextArea.Text, ASSERTION_30SECONDS_TIME_OUT).Should(gomega.MatchRegexp(strings.Split(app.ValuesRegex, ",")[0]))

		if app.DefaultApp {
			// Default profiles values are updated via annotation template parameters
			gomega.Eventually(valuesYaml.Cancel.Click).Should(gomega.Succeed())
		} else {
			// Update values.yaml for the profile
			text, _ := valuesYaml.TextArea.Text()
			for i, val := range strings.Split(app.Values, ",") {
				text = strings.ReplaceAll(text, strings.Split(app.ValuesRegex, ",")[i], val)
			}

			gomega.Expect(valuesYaml.TextArea.Clear()).To(gomega.Succeed())
			gomega.Expect(valuesYaml.TextArea.SendKeys(text)).To(gomega.Succeed(), fmt.Sprintf("Failed to change values.yaml for %s profile", app.Name))

			gomega.Eventually(valuesYaml.Save.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to save values.yaml for %s profile", app.Name))
		}
	})
}

func verifyAppInformation(applicationsPage *pages.ApplicationsPage, app Application, cluster ClusterConfig, status string) {

	ginkgo.By(fmt.Sprintf("And verify %s application information in application table for cluster: %s", app.Name, cluster.Name), func() {
		applicationInfo := applicationsPage.FindApplicationInList(app.Name)

		if app.Type == "helm_release" {
			gomega.Eventually(applicationInfo.Type, ASSERTION_1MINUTE_TIME_OUT).Should(matchers.MatchText("HelmRelease"), fmt.Sprintf("Failed to have expected %s application type: %s", app.Name, app.Type))
		} else {
			gomega.Eventually(applicationInfo.Type, ASSERTION_1MINUTE_TIME_OUT).Should(matchers.MatchText("Kustomization"), fmt.Sprintf("Failed to have expected %s application type: %s", app.Name, app.Type))
		}

		gomega.Eventually(applicationInfo.Name).Should(matchers.MatchText(app.Name), fmt.Sprintf("Failed to list %s application in  application table", app.Name))
		gomega.Eventually(applicationInfo.Namespace).Should(matchers.MatchText(app.Namespace), fmt.Sprintf("Failed to have expected %s application namespace: %s", app.Name, app.Namespace))
		gomega.Eventually(applicationInfo.Cluster).Should(matchers.MatchText(path.Join(cluster.Namespace, cluster.Name)), fmt.Sprintf("Failed to have expected %s application cluster: %s", app.Name, path.Join(cluster.Namespace, cluster.Name)))
		gomega.Eventually(applicationInfo.Source).Should(matchers.MatchText(app.Source), fmt.Sprintf("Failed to have expected %s application source: %s", app.Name, app.Source))
		gomega.Eventually(applicationInfo.Status, ASSERTION_2MINUTE_TIME_OUT).Should(matchers.MatchText(status), fmt.Sprintf("Failed to have expected %s application status: %s", app.Name, status))

		if app.Tenant != "" {
			// namespaces can take 30s to be refreshed
			// at 30s resolution
			// UI polling is 10s
			// So worst case update should be 1m10s
			gomega.Eventually(applicationInfo.Tenant, ASSERTION_2MINUTE_TIME_OUT).Should(matchers.MatchText(app.Tenant), fmt.Sprintf("Failed to have expected %s tenant", app.Tenant))
		}
	})
}

func verifyAppPage(app Application) {
	appDetailPage := pages.GetApplicationsDetailPage(webDriver, app.Type)

	ginkgo.By(fmt.Sprintf("And verify %s application page", app.Name), func() {
		gomega.Eventually(appDetailPage.Header.Text).Should(gomega.MatchRegexp(app.Name), fmt.Sprintf("Failed to verify dashboard application name %s", app.Name))
		// There is currently no Application Title on the Application Page body..
		// gomega.Eventually(appDetailPage.Title.Text).Should(gomega.MatchRegexp(app.Name), fmt.Sprintf("Failed to verify application title %s on application page", app.Name))
		gomega.Eventually(appDetailPage.Sync).Should(matchers.BeEnabled(), fmt.Sprintf("Sync button is not visible/enable for %s", app.Name))
		gomega.Eventually(appDetailPage.Details).Should(matchers.BeEnabled(), fmt.Sprintf("Details tab button is not visible/enable for %s", app.Name))
		gomega.Eventually(appDetailPage.Events).Should(matchers.BeEnabled(), fmt.Sprintf("Events tab button is not visible/enable for %s", app.Name))
		gomega.Eventually(appDetailPage.Graph).Should(matchers.BeEnabled(), fmt.Sprintf("Graph tab button is not visible/enable for %s", app.Name))
		gomega.Eventually(appDetailPage.Violations).Should(matchers.BeEnabled(), fmt.Sprintf("Violations tab button is not visible/enable for %s", app.Name))
	})
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

func verifyAppSourcePage(applicationInfo *pages.ApplicationInformation, app Application) {
	ginkgo.By(fmt.Sprintf("And navigate directly to %s Sources page", app.Name), func() {
		gomega.Expect(applicationInfo.Source.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to %s Sources pages directly", app.Name))

		sourceDetailPage := pages.GetSourceDetailPage(webDriver)
		gomega.Eventually(sourceDetailPage.Header.Text).Should(gomega.MatchRegexp(app.Source), fmt.Sprintf("Failed to verify dashboard header source name %s ", app.Source))
		gomega.Eventually(sourceDetailPage.Title.Text).Should(gomega.MatchRegexp(app.Source), fmt.Sprintf("Failed to verify application title %s on source page", app.Source))
	})
}

func verifyDeleteApplication(applicationsPage *pages.ApplicationsPage, existingAppCount int, appName, appKustomization string) {
	pages.NavigateToPage(webDriver, "Applications")

	if appKustomization != "" {
		ginkgo.By(fmt.Sprintf("And delete the %s kustomization and source maifest from the repository's master branch", appName), func() {
			cleanGitRepository(appKustomization)
		})
	}

	ginkgo.By("Then force reconcile flux-system to immediately start application deletion take effect", func() {
		reconcile("reconcile", "source", "git", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
		reconcile("reconcile", "", "kustomization", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
	})

	ginkgo.By(fmt.Sprintf("And wait for %s application to dissappeare from the dashboard", appName), func() {
		gomega.Eventually(func(g gomega.Gomega) int {
			g.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
			time.Sleep(POLL_INTERVAL_1SECONDS)
			return applicationsPage.CountApplications()
		}, ASSERTION_3MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Equal(existingAppCount), fmt.Sprintf("There should be %d application enteries in application table", existingAppCount))
	})
}

func createGitopsPR(pullRequest PullRequest) (prUrl string) {
	ginkgo.By("And set GitOps values for pull request", func() {
		gitops := pages.GetGitOps(webDriver)
		gomega.Eventually(gitops.GitOpsLabel).Should(matchers.BeFound())

		if pullRequest.Branch != "" {
			pages.ClearFieldValue(gitops.BranchName)
			gomega.Expect(gitops.BranchName.SendKeys(pullRequest.Branch)).To(gomega.Succeed())
		}
		if pullRequest.Title != "" {
			pages.ClearFieldValue(gitops.PullRequestTitle)
			gomega.Expect(gitops.PullRequestTitle.SendKeys(pullRequest.Title)).To(gomega.Succeed())
		}
		if pullRequest.Message != "" {
			pages.ClearFieldValue(gitops.CommitMessage)
			gomega.Expect(gitops.CommitMessage.SendKeys(pullRequest.Message)).To(gomega.Succeed())
		}

		authenticateWithGitProvider(webDriver, gitProviderEnv.Type, gitProviderEnv.Hostname)
		gomega.Eventually(gitops.GitCredentials).Should(matchers.BeVisible())
	})

	gitops := pages.GetGitOps(webDriver)
	messages := pages.GetMessages(webDriver)
	ginkgo.By("Then I should see a toast with a link to the creation PR", func() {
		gomega.Eventually(func(g gomega.Gomega) {
			g.Expect(gitops.CreatePR.Click()).Should(gomega.Succeed())
			g.Eventually(messages.Success, ASSERTION_30SECONDS_TIME_OUT).Should(matchers.MatchText("PR created successfully"))
		}, ASSERTION_1MINUTE_TIME_OUT).ShouldNot(gomega.HaveOccurred(), "Failed to create pull request")
	})

	prUrl, _ = messages.Success.Find("a").Attribute("href")
	return prUrl
}
