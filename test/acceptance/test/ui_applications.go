package acceptance

import (
	"context"
	"fmt"
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

func createGitKustomization(kustomizationName, kustomizationNameSpace, kustomizationPath, repoName, sourceNameSpace, targetNamespace string) (kustomization string) {
	contents, err := os.ReadFile(path.Join(testDataPath, "kustomization/git-kustomization.yaml"))
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

func installPolicyConfig(clusterName string, policyConfigYaml string) {
	ginkgo.By(fmt.Sprintf("Add/Install Policy config to the %s cluster", clusterName), func() {
		err := runCommandPassThrough("kubectl", "apply", "-f", policyConfigYaml)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to install policy Config on cluster '%s'", clusterName))
	})
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

		if app.TargetNamespace != GITOPS_DEFAULT_NAMESPACE {
			gomega.Eventually(application.CreateTargetNamespace.Check).Should(gomega.Succeed(), "Failed to select 'Create target namespace for kustomization'")
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
		gomega.Eventually(profile.Version.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to select expeted %s profile version", app.Version))
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
			gomega.Eventually(applicationInfo.Tenant).Should(matchers.MatchText(app.Tenant), fmt.Sprintf("Failed to have expected %s tenant", app.Tenant))
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
			gomega.Eventually(details.Kind.Text).Should(gomega.MatchRegexp(cases.Title(language.English, cases.NoLower).String(app.Type)), fmt.Sprintf("Failed to verify %s kind", app.Name))
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
		}, ASSERTION_3MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.MatchRegexp("^Ready"), fmt.Sprintf("Failed to verify %s Status", app.Name))

		msgRegex := fmt.Sprintf(`ReplicaSet "%s.+" has successfully progressed|Deployment has minimum availability`, app.DeploymentName)
		gomega.Eventually(details.Message.Text).Should(gomega.MatchRegexp(msgRegex), fmt.Sprintf("Failed to verify %s Message", app.Name))
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

func verifyAppViolationsList(violatingApp Application, violationsData ApplicationViolations) {

	// Declare application details page variable
	appDetailPage := pages.GetApplicationsDetailPage(webDriver, violatingApp.Type)

	ginkgo.By(fmt.Sprintf("And open  '%s' application Violations tab", violatingApp.Name), func() {

		gomega.Eventually(appDetailPage.Violations.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to click '%s' Violations tab button", violatingApp.Name))
		pages.WaitForPageToLoad(webDriver)

	})
	appViolationsList := pages.GetApplicationViolationsList(webDriver, violationsData.ViolationMessage)
	ginkgo.By("And check violations are visible in the Application Violations List", func() {

		gomega.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
		pages.WaitForPageToLoad(webDriver)
		gomega.Expect((webDriver.URL())).Should(gomega.ContainSubstring("/violations?clusterName="))

		// Checking that application violation are visible.
		gomega.Eventually(appDetailPage.Violations).Should(matchers.BeVisible())

		// Checking the Violation Message in the application violation list.
		gomega.Eventually(appViolationsList.ViolationMessage.Text).Should(gomega.Equal("MESSAGE"), "Failed to get Violation Message title in violations List page")
		gomega.Eventually(appViolationsList.ViolationMessageValue.Text).Should(gomega.Equal(violationsData.ViolationMessage), fmt.Sprintf("Failed to list '%s' violation in '%s' vioilations list", violationsData.ViolationMessage, violatingApp.Name))
		// Checking the Severity in the application violation list.
		gomega.Eventually(appViolationsList.Severity.Text).Should(gomega.Equal("SEVERITY"), "Failed to get the Severity title in App violations List")
		gomega.Expect(appViolationsList.SeverityIcon).ShouldNot(gomega.BeNil(), "Failed to get the Severity icon in App violations List")
		gomega.Eventually(appViolationsList.SeverityValue.Text).Should(gomega.Equal(violationsData.ViolationSeverity), "Failed to get the Severity Value in App violations List")
		// Checking the Violated policy in the application violation list.
		gomega.Eventually(appViolationsList.ViolatedPolicy.Text).Should(gomega.Equal("VIOLATED POLICY"), "Failed to get the Violated Policy title in App violations List")
		gomega.Eventually(appViolationsList.ViolatedPolicyValue.Text).Should(gomega.Equal(violationsData.PolicyName), "Failed to get the Violated Policy Value in App violations List")
		// Checking the Violation Time in the application violation list.
		gomega.Eventually(appViolationsList.ViolationTime.Text).Should(gomega.Equal("VIOLATION TIME"), "Failed to get the Violation time title in App violations List")
		gomega.Expect(appViolationsList.ViolationTimeValue.Text()).NotTo(gomega.BeEmpty(), "Failed to get violation time value in App violations List")
	})

	// Checking that the Violations can be filtered in the application violation list.
	ginkgo.By("And Violations can be filtered by Severity", func() {

		filterID := "severity: medium"
		searchPage := pages.GetSearchPage(webDriver)
		searchPage.SelectFilter("severity", filterID)
		gomega.Eventually(func(g gomega.Gomega) int {
			return pages.CountAppViolations(webDriver)
		}).Should(gomega.BeNumerically(">=", 2), "The number of selected violations for medium severity should be equal or greater than 2")
		gomega.Expect(pages.CountAppViolations(webDriver)).Should(gomega.Equal(pages.AppViolationOccurrances(webDriver, "severity", "medium")), "The application violations list contains severity other then the filtered medium severity")
		// Clear the filter
		searchPage.SelectFilter("severity", filterID)
	})

	// Checking that you can search by violated policy name in the application violation list.
	ginkgo.By(fmt.Sprintf("And search by violated policy name in '%s' app violations list", violatingApp.Name), func() {

		searchPage := pages.GetSearchPage(webDriver)
		searchPage.SearchName(violationsData.PolicyName)
		gomega.Eventually(func(g gomega.Gomega) int {
			return pages.CountAppViolations(webDriver)
		}).Should(gomega.BeNumerically(">=", 1), "There should be at least '1' Violation Message in the list after search")
		gomega.Eventually(appViolationsList.ViolationMessageValue.Text).Should(gomega.Equal(violationsData.ViolationMessage), "Failed to get the Violation Message Value in App violations List")

	})

}

func verifyAppViolationsDetailsPage(clusterName string, violatingApp Application, violationsData ApplicationViolations) {

	ginkgo.By(fmt.Sprintf("Verify '%s' Application Violation Details", violationsData.PolicyName), func() {

		appViolationsList := pages.GetApplicationViolationsList(webDriver, violationsData.ViolationMessage)
		gomega.Eventually(appViolationsList.ViolationMessageValue.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to violation details page of violation '%s'", violationsData.ViolationMessage))

		gomega.Expect(webDriver.URL()).Should(gomega.ContainSubstring("/clusters/violations/details?clusterName"))

		appViolationsDetialsPage := pages.GetApplicationViolationsDetailsPage(webDriver)

		gomega.Eventually(appViolationsDetialsPage.ViolationHeader.Text).Should(gomega.Equal(violationsData.ViolationMessage), "Failed to get violation header on App violations details page")
		gomega.Eventually(appViolationsDetialsPage.PolicyName.Text).Should(gomega.Equal("Policy Name :"), "Failed to get policy name field on App violations details page")
		gomega.Eventually(appViolationsDetialsPage.PolicyNameValue.Text).Should(gomega.Equal(violationsData.PolicyName), "Failed to get policy name value on App violations details page")

		// Click policy name from app violations details page to navigate to policy details page
		gomega.Eventually(appViolationsDetialsPage.PolicyNameValue.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to '%s' policy detail page", appViolationsDetialsPage.PolicyNameValue))
		gomega.Expect(webDriver.URL()).Should(gomega.ContainSubstring("/policies/details?"))

		// Navigate back to the app violations list
		gomega.Expect(webDriver.Back()).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to navigate back to the '%s' app violations list", violatingApp.Name))

		gomega.Eventually(appViolationsDetialsPage.ClusterName.Text).Should(gomega.Equal("Cluster :"), "Failed to get cluster field on App violations details page")
		gomega.Eventually(appViolationsDetialsPage.ClusterNameValue.Text).Should(gomega.MatchRegexp(clusterName), "Failed to get cluster name value on App violations details page")

		gomega.Eventually(appViolationsDetialsPage.ViolationTime.Text).Should(gomega.Equal("Violation Time :"), "Failed to get violation time field on App violations details page")
		gomega.Expect(appViolationsDetialsPage.ViolationTimeValue.Text()).NotTo(gomega.BeEmpty(), "Failed to get violation time value on App violations details page")

		gomega.Eventually(appViolationsDetialsPage.Severity.Text).Should(gomega.Equal("Severity :"), "Failed to get severity field on App violations details page")
		gomega.Expect(appViolationsDetialsPage.SeverityIcon).NotTo(gomega.BeNil(), "Failed to get severity icon value on App violations details page")
		gomega.Eventually(appViolationsDetialsPage.SeverityValue.Text).Should(gomega.Equal(violationsData.ViolationSeverity), "Failed to get severity value on App violations details page")

		gomega.Eventually(appViolationsDetialsPage.Category.Text).Should(gomega.Equal("Category :"), "Failed to get category field on App violations details page")
		gomega.Eventually(appViolationsDetialsPage.CategoryValue.Text).Should(gomega.Equal(violationsData.ViolationCategory), "Failed to get category value on App violations details page")

		gomega.Eventually(appViolationsDetialsPage.Occurrences.Text).Should(gomega.MatchRegexp("Occurrences"), "Failed to get Occurrences field on App violations details page")
		gomega.Eventually(appViolationsDetialsPage.OccurrencesCount.Text).Should(gomega.Equal("( 1 )"), "Failed to get Occurrences count on App violations details page")
		gomega.Expect(appViolationsDetialsPage.OccurrencesValue.Text()).NotTo(gomega.BeEmpty(), "Failed to get Occurrences value on App violations details page")

		gomega.Eventually(appViolationsDetialsPage.Description.Text).Should(gomega.Equal("Description:"), "Failed to get description field on App violations details page")
		gomega.Expect(appViolationsDetialsPage.DescriptionValue.Text()).NotTo(gomega.BeEmpty(), "Failed to get description value on App violations details page")

		gomega.Eventually(appViolationsDetialsPage.HowToSolve.Text).Should(gomega.Equal("How to solve:"), "Failed to get how to resolve field on App violations details page")
		gomega.Expect(appViolationsDetialsPage.HowToSolveValue.Text()).NotTo(gomega.BeEmpty(), "Failed to get how to resolve value on App violations details page")

		gomega.Eventually(appViolationsDetialsPage.ViolatingEntity.Text).Should(gomega.Equal("Violating Entity:"), "Failed to get violating entity field on App violations details page")
		gomega.Expect(appViolationsDetialsPage.ViolatingEntityValue.Text()).NotTo(gomega.BeEmpty(), "Failed to get violating entity value on App violations details page")
	})

}

func verifyPolicyConfigInAppViolationsDetails(policyName string, violationMsg string) {

	ginkgo.By("Navigate back to Violations list", func() {

		gomega.Eventually(webDriver.Back).ShouldNot(gomega.HaveOccurred(), "Failed to navigate back to violations list")
		pages.WaitForPageToLoad(webDriver)

	})
	appViolationsList := pages.GetApplicationViolationsList(webDriver, violationMsg)

	gomega.Eventually(appViolationsList.ViolationMessageValue.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to violation details page of violation '%s'", violationMsg))
	pages.WaitForPageToLoad(webDriver)
	ViolationsDetialsPage := pages.GetApplicationViolationsDetailsPage(webDriver)

	ginkgo.By(fmt.Sprintf("And verify policy config parameters values  for '%s'", policyName), func() {
		parameter := ViolationsDetialsPage.GetPolicyConfigViolationsParameters("replica_count")
		gomega.Expect(parameter.ParameterName.Text()).Should(gomega.MatchRegexp(`replica_count`), "Failed to verify `replica_count` parameter 'Name'")
		gomega.Expect(parameter.ParameterValue.Text()).Should(gomega.MatchRegexp(`4`), "Failed to verify `replica_count` parameter 'Value'")
		gomega.Expect(parameter.PolicyConfigName.Text()).Should(gomega.MatchRegexp(`policy-config-001`), "Failed to verify `replica_count` parameter Policy Config 'Name'")

		parameter = ViolationsDetialsPage.GetPolicyConfigViolationsParameters("exclude_namespaces")
		gomega.Expect(parameter.ParameterName.Text()).Should(gomega.MatchRegexp(`exclude_namespaces`), "Failed to verify `exclude_namespaces` parameter 'Name'")
		gomega.Expect(parameter.ParameterValue.Text()).Should(gomega.MatchRegexp(`undefined`), "Failed to verify `exclude_namespaces` parameter 'Value'")
		gomega.Expect(parameter.PolicyConfigName.Text()).Should(gomega.MatchRegexp(`-`), "Failed to verify `exclude_namespaces` parameter 'Value'")

		parameter = ViolationsDetialsPage.GetPolicyConfigViolationsParameters("exclude_label_key")
		gomega.Expect(parameter.ParameterName.Text()).Should(gomega.MatchRegexp(`exclude_label_key`), "Failed to verify `exclude_label_key` parameter'Name'")
		gomega.Expect(parameter.ParameterValue.Text()).Should(gomega.MatchRegexp(`undefined`), "Failed to verify `exclude_label_key` parameter 'Value'")
		gomega.Expect(parameter.PolicyConfigName.Text()).Should(gomega.MatchRegexp(`-`), "Failed to verify `exclude_label_key` parameter Policy Config 'Name'")

		parameter = ViolationsDetialsPage.GetPolicyConfigViolationsParameters("exclude_label_value")
		gomega.Expect(parameter.ParameterName.Text()).Should(gomega.MatchRegexp(`exclude_label_value`), "Failed to verify `exclude_label_value` parameter 'Name'")
		gomega.Expect(parameter.ParameterValue.Text()).Should(gomega.MatchRegexp(`undefined`), "Failed to verify `exclude_label_value` parameter 'Value'")
		gomega.Expect(parameter.PolicyConfigName.Text()).Should(gomega.MatchRegexp(`-`), "Failed to verify `exclude_label_value` parameter Policy Config 'Name'")
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

var _ = ginkgo.Describe("Multi-Cluster Control Plane Applications", ginkgo.Label("ui", "application"), func() {

	ginkgo.BeforeEach(func() {
		gomega.Expect(webDriver.Navigate(testUiUrl)).To(gomega.Succeed())

		if !pages.ElementExist(pages.Navbar(webDriver).Title, 3) {
			loginUser()
		}
	})

	ginkgo.Context("When no applications are installed", func() {

		ginkgo.It("Verify management cluster dashboard shows bootstrap 'flux-system' application", func() {
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
				gomega.Expect(waitForGitopsResources(context.Background(), Request{"objects", []byte(`{"kind": "Kustomization"}`)}, POLL_INTERVAL_15SECONDS)).To(gomega.Succeed(), "Failed to get a successful response from /v1/objects")
			})

			applicationsPage := pages.GetApplicationsPage(webDriver)
			pages.WaitForPageToLoad(webDriver)

			ginkgo.By("And wait for Applications page to be rendered", func() {
				gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())
				gomega.Eventually(applicationsPage.CountApplications, ASSERTION_1MINUTE_TIME_OUT).Should(gomega.Equal(1), "There should not be any application in application's table except flux-system")
			})

			verifyAppInformation(applicationsPage, fluxSystem, mgmtCluster, "Ready")
		})
	})

	ginkgo.Context("Applications(s) can be installed on management cluster", func() {

		var existingAppCount int
		var downloadedResourcesPath string
		appNameSpace := "test-kustomization"
		appTargetNamespace := "test-system"

		mgmtCluster := ClusterConfig{
			Type:      "management",
			Name:      "management",
			Namespace: "",
		}

		ginkgo.JustBeforeEach(func() {
			downloadedResourcesPath = path.Join(os.Getenv("HOME"), "Downloads", "resources.zip")
			// Application target namespace is created by the kustomization 'Add Application' UI
			createNamespace([]string{appNameSpace, appTargetNamespace})
			_ = deleteFile([]string{downloadedResourcesPath})
		})

		ginkgo.JustAfterEach(func() {
			pages.CloseOtherWindows(webDriver, enterpriseWindow)
			// Wait for the application to be deleted gracefully, needed when the test fails before deleting the application
			gomega.Eventually(func(g gomega.Gomega) int {
				return getApplicationCount()
			}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Equal(existingAppCount), fmt.Sprintf("There should be %d application enteries after application(s) deletion", existingAppCount))

			deleteNamespace([]string{appNameSpace, appTargetNamespace})
			_ = deleteFile([]string{downloadedResourcesPath})
		})

		ginkgo.It("Verify application with annotations/metadata can be installed  and dashboard is updated accordingly", func() {

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
				gomega.Eventually(applicationsPage.CountApplications, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(totalAppCount), fmt.Sprintf("There should be %d application enteries in application table", totalAppCount))
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

		ginkgo.It("Verify application can be installed from HelmRepository source and dashboard is updated accordingly", func() {
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
			appKustomization := fmt.Sprintf("./clusters/%s/%s-%s-helmrelease.yaml", mgmtCluster.Name, metallb.Name, metallb.TargetNamespace)

			repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
			existingAppCount = getApplicationCount()

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

			preview := pages.GetPreview(webDriver)
			ginkgo.By("Then I should preview the PR", func() {
				gomega.Eventually(func(g gomega.Gomega) {
					g.Expect(createPage.PreviewPR.Click()).Should(gomega.Succeed())
					g.Expect(preview.Title.Text()).Should(gomega.MatchRegexp("PR Preview"))

				}, ASSERTION_1MINUTE_TIME_OUT).Should(gomega.Succeed(), "Failed to get PR preview")
			})

			ginkgo.By("Then verify preview tab lists", func() {
				// Verify profiles preview
				gomega.Eventually(preview.GetPreviewTab("Helm Releases").Click).Should(gomega.Succeed(), "Failed to switch to 'PROFILES' preview tab")
				gomega.Eventually(preview.Path.At(0)).Should(matchers.MatchText(path.Join("clusters/management", strings.Join([]string{metallb.Name, metallb.TargetNamespace, "helmrelease.yaml"}, "-"))))
				gomega.Eventually(preview.Text.At(0)).Should(matchers.MatchText(fmt.Sprintf(`kind: HelmRelease[\s\w\d./:-]*name: %s[\s\w\d./:-]*namespace: %s[\s\w\d./:-]*spec`, metallb.Name, metallb.Namespace)))
				gomega.Eventually(preview.Text.At(0)).Should(matchers.MatchText(fmt.Sprintf(`chart: %s[\s\w\d./:-]*sourceRef:[\s\w\d./:-]*name: %s[\s\w\d./:-]*version: %s[\s\w\d./:-]*targetNamespace: %s[\s\w\d./:-]*prometheus[\s\w\d./:-]*namespace: %s`, metallb.Name, metallb.Chart, metallb.Version, metallb.TargetNamespace, metallb.TargetNamespace)))

				// Verify kustomization tab view is disabled because no kustomization is part of pull request
				gomega.Expect(preview.GetPreviewTab("Kustomizations").Attribute("class")).Should(gomega.MatchRegexp("Mui-disabled"), "'KUSTOMIZATIONS' preview tab should be disabled")
			})

			ginkgo.By("And verify downloaded preview resources", func() {
				// verify download prview resources
				gomega.Eventually(func(g gomega.Gomega) {
					g.Expect(preview.Download.Click()).Should(gomega.Succeed())
					_, err := os.Stat(downloadedResourcesPath)
					g.Expect(err).Should(gomega.Succeed())
				}, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_3SECONDS).ShouldNot(gomega.HaveOccurred(), "Failed to click 'Download' preview resources")
				gomega.Eventually(preview.Close.Click).Should(gomega.Succeed(), "Failed to close the preview dialog")

				fileList, _ := getArchiveFileList(downloadedResourcesPath)
				previewResources := []string{
					path.Join("clusters/management", strings.Join([]string{metallb.Name, metallb.TargetNamespace, "helmrelease.yaml"}, "-")),
				}
				gomega.Expect(len(fileList)).Should(gomega.Equal(len(previewResources)), "Failed to verify expected number of downloaded preview resources")
				gomega.Expect(fileList).Should(gomega.ContainElements(previewResources), "Failed to verify downloaded preview resources files")
			})

			prUrl := createGitopsPR(pullRequest)
			ginkgo.By("Then I should merge the pull request to start application reconciliation", func() {
				createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
				gomega.Expect(createPRUrl).Should(gomega.Equal(prUrl))

			})

			ginkgo.By("And the manifests are present in the cluster config repository", func() {
				mergePullRequest(gitProviderEnv, repoAbsolutePath, prUrl)
				pullGitRepo(repoAbsolutePath)

				_, err := os.Stat(path.Join(repoAbsolutePath, "clusters/management", strings.Join([]string{metallb.Name, metallb.TargetNamespace, "helmrelease.yaml"}, "-")))
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "helmrelease kustomization yaml can not be found.")
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

		ginkgo.It("Verify application can be installed from GitRepository source and dashboard is updated accordingly", func() {
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

			// tartget namespace is created by the kustomization, hence deleting it beforehand to avoid namespace creation errors
			deleteNamespace([]string{appTargetNamespace})

			sourceURL := "https://github.com/stefanprodan/podinfo"
			appKustomization := fmt.Sprintf("./clusters/%s/%s-%s-kustomization.yaml", mgmtCluster.Name, podinfo.Name, podinfo.Namespace)

			defer deleteSource("git", podinfo.Source, podinfo.Namespace, "")
			defer cleanGitRepository(appKustomization)
			defer cleanGitRepository(fmt.Sprintf("./clusters/%s/%s-namespace.yaml", mgmtCluster.Name, podinfo.TargetNamespace))

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

			createPage := pages.GetCreateClusterPage(webDriver)
			preview := pages.GetPreview(webDriver)
			ginkgo.By("Then I should preview the PR", func() {
				gomega.Eventually(func(g gomega.Gomega) {
					g.Expect(createPage.PreviewPR.Click()).Should(gomega.Succeed())
					g.Expect(preview.Title.Text()).Should(gomega.MatchRegexp("PR Preview"))

				}, ASSERTION_1MINUTE_TIME_OUT).Should(gomega.Succeed(), "Failed to get PR preview")
			})

			ginkgo.By("Then verify preview tab lists", func() {
				// Verify kustomizations preview resources.zip
				gomega.Eventually(preview.GetPreviewTab("Kustomizations").Click).Should(gomega.Succeed(), "Failed to switch to 'KUSTOMIZATION' preview tab")
				gomega.Eventually(preview.Path.At(0)).Should(matchers.MatchText(path.Join("clusters/management", strings.Join([]string{podinfo.TargetNamespace, "namespace.yaml"}, "-"))))
				gomega.Eventually(preview.Text.At(0)).Should(matchers.MatchText(fmt.Sprintf(`kind: Namespace[\s\w\d./:-]*name: %s`, podinfo.TargetNamespace)))
				gomega.Eventually(preview.Path.At(1)).Should(matchers.MatchText(path.Join("clusters/management", strings.Join([]string{podinfo.Name, podinfo.Namespace, "kustomization.yaml"}, "-"))))
				gomega.Eventually(preview.Text.At(1)).Should(matchers.MatchText(fmt.Sprintf(`kind: Kustomization[\s\w\d./:-]*name: %s[\s\w\d./:-]*namespace: %s[\s\w\d./:-]*spec`, podinfo.Name, podinfo.Namespace)))
				gomega.Eventually(preview.Text.At(1)).Should(matchers.MatchText(fmt.Sprintf(`path: %s`, podinfo.Path)))
				gomega.Eventually(preview.Text.At(1)).Should(matchers.MatchText(fmt.Sprintf(`sourceRef:[\s\w\d./:-]*kind: GitRepository[\s\w\d./:-]*name: %s[\s\w\d./:-]*namespace: %s[\s\w\d./:-]*targetNamespace: %s`, podinfo.Source, podinfo.Namespace, podinfo.TargetNamespace)))

				// Verify helmrelease tab view is disabled because no helmrelease is part of pull request
				gomega.Expect(preview.GetPreviewTab("Helm Releases").Attribute("class")).Should(gomega.MatchRegexp("Mui-disabled"), "'Helmrelease' preview tab should be disabled")
			})

			ginkgo.By("And verify downloaded preview resources", func() {
				// verify download prview resources
				gomega.Eventually(func(g gomega.Gomega) {
					g.Expect(preview.Download.Click()).Should(gomega.Succeed())
					_, err := os.Stat(downloadedResourcesPath)
					g.Expect(err).Should(gomega.Succeed())
				}, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_3SECONDS).ShouldNot(gomega.HaveOccurred(), "Failed to click 'Download' preview resources")
				gomega.Eventually(preview.Close.Click).Should(gomega.Succeed(), "Failed to close the preview dialog")

				fileList, _ := getArchiveFileList(downloadedResourcesPath)
				previewResources := []string{
					path.Join("clusters/management", strings.Join([]string{podinfo.TargetNamespace, "namespace.yaml"}, "-")),
					path.Join("clusters/management", strings.Join([]string{podinfo.Name, podinfo.Namespace, "kustomization.yaml"}, "-")),
				}
				gomega.Expect(len(fileList)).Should(gomega.Equal(len(previewResources)), "Failed to verify expected number of downloaded preview resources")
				gomega.Expect(fileList).Should(gomega.ContainElements(previewResources), "Failed to verify downloaded preview resources files")
			})

			prUrl := createGitopsPR(pullRequest)

			ginkgo.By("Then I should merge the pull request to start application reconciliation", func() {
				createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
				gomega.Expect(createPRUrl).Should(gomega.Equal(prUrl))

			})

			ginkgo.By("And the manifests are present in the cluster config repository", func() {
				mergePullRequest(gitProviderEnv, repoAbsolutePath, prUrl)
				pullGitRepo(repoAbsolutePath)

				_, err := os.Stat(path.Join(repoAbsolutePath, "clusters/management", strings.Join([]string{podinfo.TargetNamespace, "namespace.yaml"}, "-")))
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "target namespace.yaml can not be found.")

				_, err = os.Stat(path.Join(repoAbsolutePath, "clusters/management", strings.Join([]string{podinfo.Name, podinfo.Namespace, "kustomization.yaml"}, "-")))
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Kustomization kustomization.yaml can not be found.")
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
			verifyAppDetails(podinfo, mgmtCluster)
			verfifyAppGraph(podinfo)

			navigatetoApplicationsPage(applicationsPage)
			verifyAppSourcePage(applicationInfo, podinfo)

			verifyDeleteApplication(applicationsPage, existingAppCount, podinfo.Name, appKustomization)
		})
	})

	ginkgo.Context("Applications(s) can be installed on leaf cluster", ginkgo.Label("kind-leaf-cluster"), func() {
		var mgmtClusterContext string
		var leafClusterContext string
		var leafClusterkubeconfig string
		var clusterBootstrapCopnfig string
		var gitopsCluster string
		var existingAppCount int
		var downloadedResourcesPath string
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
			downloadedResourcesPath = path.Join(os.Getenv("HOME"), "Downloads", "resources.zip")
			existingAppCount = getApplicationCount()
			mgmtClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
			createCluster("kind", leafCluster.Name, "")
			createNamespace([]string{appNameSpace, appTargetNamespace})
			leafClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")

			_ = deleteFile([]string{downloadedResourcesPath})
		})

		ginkgo.JustAfterEach(func() {
			useClusterContext(mgmtClusterContext)

			deleteSecret([]string{leafClusterkubeconfig, patSecret}, leafCluster.Namespace)
			_ = runCommandPassThrough("kubectl", "delete", "-f", clusterBootstrapCopnfig)
			_ = runCommandPassThrough("kubectl", "delete", "-f", gitopsCluster)

			deleteCluster("kind", leafCluster.Name, "")
			cleanGitRepository(path.Join("./clusters", leafCluster.Namespace))
			deleteNamespace([]string{leafCluster.Namespace})
			_ = deleteFile([]string{downloadedResourcesPath})

		})

		ginkgo.It("Verify application can be installed from GitRepository source on leaf cluster and management dashboard is updated accordingly", func() {
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
			gitopsCluster = connectGitopsCluster(leafCluster.Name, leafCluster.Namespace, bootstrapLabel, leafClusterkubeconfig)
			createLeafClusterSecret(leafCluster.Namespace, leafClusterkubeconfig)

			waitForLeafClusterAvailability(leafCluster.Name, "Ready")
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

			createPage := pages.GetCreateClusterPage(webDriver)
			preview := pages.GetPreview(webDriver)
			ginkgo.By("Then I should preview the PR", func() {
				gomega.Eventually(func(g gomega.Gomega) {
					g.Expect(createPage.PreviewPR.Click()).Should(gomega.Succeed())
					g.Expect(preview.Title.Text()).Should(gomega.MatchRegexp("PR Preview"))
				}, ASSERTION_1MINUTE_TIME_OUT).Should(gomega.Succeed(), "Failed to get PR preview")
			})

			ginkgo.By("Then verify preview tab lists", func() {
				// Verify kustomizations preview resources.zip
				gomega.Eventually(preview.GetPreviewTab("Kustomizations").Click).Should(gomega.Succeed(), "Failed to switch to 'KUSTOMIZATION' preview tab")
				gomega.Eventually(preview.Path.At(0)).Should(matchers.MatchText(path.Join("clusters", leafCluster.Namespace, leafCluster.Name, strings.Join([]string{podinfo.TargetNamespace, "namespace.yaml"}, "-"))))
				gomega.Eventually(preview.Text.At(0)).Should(matchers.MatchText(fmt.Sprintf(`kind: Namespace[\s\w\d./:-]*name: %s`, podinfo.TargetNamespace)))
				gomega.Eventually(preview.Path.At(1)).Should(matchers.MatchText(path.Join("clusters", leafCluster.Namespace, leafCluster.Name, strings.Join([]string{podinfo.Name, podinfo.Namespace, "kustomization.yaml"}, "-"))))
				gomega.Eventually(preview.Text.At(1)).Should(matchers.MatchText(fmt.Sprintf(`kind: Kustomization[\s\w\d./:-]*name: %s[\s\w\d./:-]*namespace: %s[\s\w\d./:-]*spec`, podinfo.Name, podinfo.Namespace)))
				gomega.Eventually(preview.Text.At(1)).Should(matchers.MatchText(fmt.Sprintf(`path: %s`, podinfo.Path)))
				gomega.Eventually(preview.Text.At(1)).Should(matchers.MatchText(fmt.Sprintf(`sourceRef:[\s\w\d./:-]*kind: GitRepository[\s\w\d./:-]*name: %s[\s\w\d./:-]*namespace: %s[\s\w\d./:-]*targetNamespace: %s`, podinfo.Source, podinfo.Namespace, podinfo.TargetNamespace)))

				// Verify helmrelease tab view is disabled because no helmrelease is part of pull request
				gomega.Expect(preview.GetPreviewTab("Helm Releases").Attribute("class")).Should(gomega.MatchRegexp("Mui-disabled"), "'Helmrelease' preview tab should be disabled")
			})

			ginkgo.By("And verify downloaded preview resources", func() {
				// verify download prview resources
				gomega.Eventually(func(g gomega.Gomega) {
					g.Expect(preview.Download.Click()).Should(gomega.Succeed())
					_, err := os.Stat(downloadedResourcesPath)
					g.Expect(err).Should(gomega.Succeed())
				}, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_3SECONDS).ShouldNot(gomega.HaveOccurred(), "Failed to click 'Download' preview resources")
				gomega.Eventually(preview.Close.Click).Should(gomega.Succeed(), "Failed to close the preview dialog")

				fileList, _ := getArchiveFileList(downloadedResourcesPath)
				previewResources := []string{
					path.Join("clusters", leafCluster.Namespace, leafCluster.Name, strings.Join([]string{podinfo.TargetNamespace, "namespace.yaml"}, "-")),
					path.Join("clusters", leafCluster.Namespace, leafCluster.Name, strings.Join([]string{podinfo.Name, podinfo.Namespace, "kustomization.yaml"}, "-")),
				}
				gomega.Expect(len(fileList)).Should(gomega.Equal(len(previewResources)), "Failed to verify expected number of downloaded preview resources")
				gomega.Expect(fileList).Should(gomega.ContainElements(previewResources), "Failed to verify downloaded preview resources files")
			})

			_ = createGitopsPR(pullRequest)

			ginkgo.By("Then I should merge the pull request to start application reconciliation", func() {
				createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
				mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
			})

			ginkgo.By("Then force reconcile leaf cluster flux-system to immediately start application provisioning", func() {
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
			verifyAppDetails(podinfo, leafCluster)
			verfifyAppGraph(podinfo)

			navigatetoApplicationsPage(applicationsPage)
			verifyAppSourcePage(applicationInfo, podinfo)

			verifyDeleteApplication(applicationsPage, existingAppCount, podinfo.Name, appKustomization)
		})

		ginkgo.It("Verify application can be installed from HelmRepository source on leaf cluster and management dashboard is updated accordingly", func() {
			metallb := Application{
				Type:            "helm_release",
				Chart:           "profiles-catalog",
				SyncInterval:    "10m",
				Name:            "metallb",
				DeploymentName:  "metallb-controller",
				Namespace:       GITOPS_DEFAULT_NAMESPACE, // HelmRelease application always get installed in flux-system namespace
				TargetNamespace: appTargetNamespace,
				Source:          GITOPS_DEFAULT_NAMESPACE + "-metallb",
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
			appKustomization := fmt.Sprintf("./clusters/%s/%s/%s-%s-helmrelease.yaml", leafCluster.Namespace, leafCluster.Name, metallb.Name, metallb.TargetNamespace)

			repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
			leafClusterkubeconfig = createLeafClusterKubeconfig(leafClusterContext, leafCluster.Name, leafCluster.Namespace)

			useClusterContext(mgmtClusterContext)
			createNamespace([]string{leafCluster.Namespace})

			createPATSecret(leafCluster.Namespace, patSecret)
			clusterBootstrapCopnfig = createClusterBootstrapConfig(leafCluster.Name, leafCluster.Namespace, bootstrapLabel, patSecret)
			gitopsCluster = connectGitopsCluster(leafCluster.Name, leafCluster.Namespace, bootstrapLabel, leafClusterkubeconfig)
			createLeafClusterSecret(leafCluster.Namespace, leafClusterkubeconfig)

			waitForLeafClusterAvailability(leafCluster.Name, "Ready")
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
				gomega.Expect(waitForGitopsResources(context.Background(), Request{Path: `charts/list?repository.name=weaveworks-charts&repository.namespace=flux-system&repository.cluster.name=management`}, POLL_INTERVAL_5SECONDS, ASSERTION_15MINUTE_TIME_OUT)).To(gomega.Succeed(), "Failed to get a successful response from /v1/charts ")
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
					g.Eventually(application.SelectListItem(webDriver, leafCluster.Name).Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to select %s cluster from clusters list", leafCluster.Name))
					g.Eventually(application.Source.Click).Should(gomega.Succeed(), "Failed to click Select Source list")
					return pages.ElementExist(application.SelectListItem(webDriver, metallb.Chart))
				}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.BeTrue(), fmt.Sprintf("HelmRepository %s source is not listed in source's list", metallb.Name))

				gomega.Eventually(application.SelectListItem(webDriver, metallb.Chart).Click).Should(gomega.Succeed(), "Failed to select HelmRepository source from sources list")
				gomega.Eventually(application.SourceHref.Text).Should(gomega.MatchRegexp(sourceURL), "Failed to find the source href")
			})

			AddHelmReleaseApp(profile, metallb)

			preview := pages.GetPreview(webDriver)
			ginkgo.By("Then I should preview the PR", func() {
				gomega.Eventually(func(g gomega.Gomega) {
					g.Expect(createPage.PreviewPR.Click()).Should(gomega.Succeed())
					g.Expect(preview.Title.Text()).Should(gomega.MatchRegexp("PR Preview"))

				}, ASSERTION_1MINUTE_TIME_OUT).Should(gomega.Succeed(), "Failed to get PR preview")
			})

			ginkgo.By("Then verify preview tab lists", func() {
				// Verify profiles preview
				gomega.Eventually(preview.GetPreviewTab("Helm Releases").Click).Should(gomega.Succeed(), "Failed to switch to 'PROFILES' preview tab")
				gomega.Eventually(preview.Path.At(0)).Should(matchers.MatchText(path.Join("clusters", leafCluster.Namespace, leafCluster.Name, strings.Join([]string{metallb.Name, metallb.TargetNamespace, "helmrelease.yaml"}, "-"))))
				gomega.Eventually(preview.Text.At(0)).Should(matchers.MatchText(fmt.Sprintf(`kind: HelmRelease[\s\w\d./:-]*name: %s[\s\w\d./:-]*namespace: %s[\s\w\d./:-]*spec`, metallb.Name, metallb.Namespace)))
				gomega.Eventually(preview.Text.At(0)).Should(matchers.MatchText(fmt.Sprintf(`chart: %s[\s\w\d./:-]*sourceRef:[\s\w\d./:-]*name: %s[\s\w\d./:-]*version: %s[\s\w\d./:-]*targetNamespace: %s[\s\w\d./:-]*prometheus[\s\w\d./:-]*namespace: %s`, metallb.Name, metallb.Chart, metallb.Version, metallb.TargetNamespace, appNameSpace)))

				// Verify kustomization tab view is disabled because no kustomization is part of pull request
				gomega.Expect(preview.GetPreviewTab("Kustomizations").Attribute("class")).Should(gomega.MatchRegexp("Mui-disabled"), "'KUSTOMIZATIONS' preview tab should be disabled")
			})

			ginkgo.By("And verify downloaded preview resources", func() {
				// verify download prview resources
				gomega.Eventually(func(g gomega.Gomega) {
					g.Expect(preview.Download.Click()).Should(gomega.Succeed())
					_, err := os.Stat(downloadedResourcesPath)
					g.Expect(err).Should(gomega.Succeed())
				}, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_3SECONDS).ShouldNot(gomega.HaveOccurred(), "Failed to click 'Download' preview resources")
				gomega.Eventually(preview.Close.Click).Should(gomega.Succeed(), "Failed to close the preview dialog")

				fileList, _ := getArchiveFileList(downloadedResourcesPath)
				previewResources := []string{
					path.Join("clusters", leafCluster.Namespace, leafCluster.Name, strings.Join([]string{metallb.Name, metallb.TargetNamespace, "helmrelease.yaml"}, "-")),
				}
				gomega.Expect(len(fileList)).Should(gomega.Equal(len(previewResources)), "Failed to verify expected number of downloaded preview resources")
				gomega.Expect(fileList).Should(gomega.ContainElements(previewResources), "Failed to verify downloaded preview resources files")
			})

			_ = createGitopsPR(pullRequest)

			ginkgo.By("Then I should merge the pull request to start application reconciliation", func() {
				createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
				mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
			})

			ginkgo.By("Then force reconcile leaf cluster flux-system to immediately start application provisioning", func() {
				useClusterContext(leafClusterContext)
				reconcile("reconcile", "source", "git", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
				reconcile("reconcile", "", "kustomization", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
				useClusterContext(mgmtClusterContext)
			})

			ginkgo.By(fmt.Sprintf("And wait for %s application to be visibe on the dashboard", metallb.Name), func() {
				gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())

				totalAppCount := existingAppCount + 1 // metallb (leaf cluster)
				gomega.Eventually(applicationsPage.CountApplications, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(totalAppCount), fmt.Sprintf("There should be %d application enteries in application table", totalAppCount))
			})

			ginkgo.By(fmt.Sprintf("And search leaf cluster '%s' app", leafCluster.Name), func() {
				searchPage := pages.GetSearchPage(webDriver)
				searchPage.SearchName(metallb.Name)
				gomega.Eventually(applicationsPage.CountApplications).Should(gomega.Equal(1), "There should be '1' application entery in application table after search")
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
	ginkgo.Context("Application violations are available for management cluster", ginkgo.Label("violation"), func() {
		// Count of existing applications before deploying new application
		var existingAppCount int
		var downloadedResourcesPath string
		var policiesYaml string
		var policyConfigYaml string

		// Just specify the violated application info to create it
		appNameSpace := "test-kustomization"
		appTargetNamespace := "test-system"

		mgmtCluster := ClusterConfig{
			Type:      "management",
			Name:      "management",
			Namespace: "",
		}

		ginkgo.JustBeforeEach(func() {
			downloadedResourcesPath = path.Join(os.Getenv("HOME"), "Downloads", "resources.zip")
			policiesYaml = path.Join(testDataPath, "policies/policies.yaml")
			policyConfigYaml = path.Join(testDataPath, "policies/policy-config.yaml")

			// Application target namespace is created by the kustomization 'Add Application' UI
			createNamespace([]string{appNameSpace, appTargetNamespace})
			_ = deleteFile([]string{downloadedResourcesPath})

			// Add/Install test Policies,Policy Config on the management cluster
			installTestPolicies(mgmtCluster.Name, policiesYaml)
			installPolicyConfig(mgmtCluster.Name, policyConfigYaml)
		})

		ginkgo.JustAfterEach(func() {
			// Wait for the application to be deleted gracefully, needed when the test fails before deleting the application
			gomega.Eventually(func(g gomega.Gomega) int {
				return getApplicationCount()
			}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Equal(existingAppCount), fmt.Sprintf("There should be %d application enteries after application(s) deletion", existingAppCount))

			// Delete the Policy config and test policies
			_ = runCommandPassThrough("kubectl", "delete", "-f", policyConfigYaml)
			_ = runCommandPassThrough("kubectl", "delete", "-f", policiesYaml)

			deleteNamespace([]string{appNameSpace, appTargetNamespace})
			_ = deleteFile([]string{downloadedResourcesPath})
		})

		ginkgo.It("Verify application violations for management cluster", func() {
			// Podinfo application details
			podinfo := Application{
				Type:            "kustomization",
				Name:            "app-violations-podinfo",
				DeploymentName:  "podinfo",
				Namespace:       appNameSpace,
				TargetNamespace: appTargetNamespace,
				Source:          "app-violations-podinfo",
				Path:            "./kustomize",
				SyncInterval:    "30s",
			}

			// App Violations data
			appViolations := ApplicationViolations{
				PolicyName:               "Container Image Pull Policy acceptance test",
				ViolationMessage:         `Container Image Pull Policy acceptance test in deployment podinfo (1 occurrences)`,
				ViolationSeverity:        "Medium",
				ViolationCategory:        "weave.categories.software-supply-chain",
				ConfigPolicy:             "Containers Minimum Replica Count acceptance test",
				PolicyConfigViolationMsg: `Containers Minimum Replica Count acceptance test in deployment podinfo (1 occurrences)`,
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
			// Declare application page variable
			applicationsPage := pages.GetApplicationsPage(webDriver)

			ginkgo.By("And add Kustomization & GitRepository Source manifests pointing to podinfo repository’s master branch)", func() {

				pullGitRepo(repoAbsolutePath)
				err := runCommandPassThrough("sh", "-c", fmt.Sprintf("mkdir -p %[2]v && cp -f %[1]v %[2]v", appKustomization, path.Join(repoAbsolutePath, appDir)))
				gomega.Expect(err).Should(gomega.BeNil(), "Failed to add kustomization file for '%s'", podinfo.Name)
				gitUpdateCommitPush(repoAbsolutePath, "Adding podinfo kustomization")
			})

			ginkgo.By("And wait for podinfo application to be visibe on the dashboard", func() {
				gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())

				totalAppCount := existingAppCount + 1
				gomega.Eventually(applicationsPage.CountApplications, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(totalAppCount), fmt.Sprintf("There should be %d application enteries in application table", totalAppCount))
			})

			verifyAppInformation(applicationsPage, podinfo, mgmtCluster, "Ready")

			applicationInfo := applicationsPage.FindApplicationInList(podinfo.Name)
			ginkgo.By(fmt.Sprintf("And navigate to %s application page", podinfo.Name), func() {
				gomega.Eventually(applicationInfo.Name.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to %s application detail page", podinfo.Name))
			})

			verifyAppViolationsList(podinfo, appViolations)
			verifyAppViolationsDetailsPage(mgmtCluster.Name, podinfo, appViolations)
			verifyPolicyConfigInAppViolationsDetails(appViolations.ConfigPolicy, appViolations.PolicyConfigViolationMsg)
			verifyDeleteApplication(applicationsPage, existingAppCount, podinfo.Name, appDir)

		})
	})

	ginkgo.Context("Application violations are available for leaf cluster", ginkgo.Label("violation", "kind-leaf-cluster"), func() {
		var mgmtClusterContext string
		var leafClusterContext string
		var leafClusterkubeconfig string
		var clusterBootstrapCopnfig string
		var gitopsCluster string
		var existingAppCount int
		var policiesYaml string
		var policyConfigYaml string
		patSecret := "application-violations-pat"
		bootstrapLabel := "bootstrap"

		// Just specify the violated application info to create it
		appNameSpace := "test-kustomization"
		appTargetNamespace := "test-system"

		// Just specify the leaf cluster info to create it
		leafCluster := ClusterConfig{
			Type:      "leaf",
			Name:      "app-violations-leaf-cluster-test",
			Namespace: "test-system",
		}

		ginkgo.JustBeforeEach(func() {
			policiesYaml = path.Join(testDataPath, "policies/policies.yaml")
			policyConfigYaml = path.Join(testDataPath, "policies/policy-config.yaml")

			// Get the count of existing applications before deploying new application
			existingAppCount = getApplicationCount()
			mgmtClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")

			createCluster("kind", leafCluster.Name, "")
			// Create App namespace
			createNamespace([]string{appNameSpace, appTargetNamespace})
			leafClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")

			// Add/Install Policy Agent on the leaf cluster
			installPolicyAgent(leafCluster.Name)
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

		ginkgo.It("Verify application violations for leaf cluster", func() {
			// Podinfo application details
			podinfo := Application{
				Type:            "kustomization",
				Name:            "app-violations-podinfo",
				DeploymentName:  "podinfo",
				Namespace:       appNameSpace,
				TargetNamespace: appTargetNamespace,
				Source:          "app-violations-podinfo",
				Path:            "./kustomize",
				SyncInterval:    "10m",
			}

			pullRequest := PullRequest{
				Branch:  "Leaf-cluster-apps-kustomization-" + randString(5),
				Title:   "Leaf Cluster Application Kustomization PR",
				Message: "Adding leaf cluster applications kustomization",
			}

			sourceURL := "https://github.com/stefanprodan/podinfo"
			appKustomization := fmt.Sprintf("./clusters/%s/%s/%s-%s-kustomization.yaml", leafCluster.Namespace, leafCluster.Name, podinfo.Name, podinfo.Namespace)

			repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)

			// App Violations data
			appViolations := ApplicationViolations{
				PolicyName:               "Container Image Pull Policy acceptance test",
				ViolationMessage:         `Container Image Pull Policy acceptance test in deployment podinfo (1 occurrences)`,
				ViolationSeverity:        "Medium",
				ViolationCategory:        "weave.categories.software-supply-chain",
				ConfigPolicy:             "Containers Minimum Replica Count acceptance test",
				PolicyConfigViolationMsg: `Containers Minimum Replica Count acceptance test in deployment podinfo (1 occurrences)`,
			}
			useClusterContext(mgmtClusterContext)
			// Create leaf cluster namespace
			createNamespace([]string{leafCluster.Namespace})

			// Create leaf cluster kubeconfig
			leafClusterkubeconfig = createLeafClusterKubeconfig(leafClusterContext, leafCluster.Name, leafCluster.Namespace)

			createPATSecret(leafCluster.Namespace, patSecret)
			clusterBootstrapCopnfig = createClusterBootstrapConfig(leafCluster.Name, leafCluster.Namespace, bootstrapLabel, patSecret)
			gitopsCluster = connectGitopsCluster(leafCluster.Name, leafCluster.Namespace, bootstrapLabel, leafClusterkubeconfig)
			createLeafClusterSecret(leafCluster.Namespace, leafClusterkubeconfig)

			// Declare application page variable
			applicationsPage := pages.GetApplicationsPage(webDriver)

			// First let the leaf cluster to bootstrap and be Ready before installing policies.'Containers Minimum Replica Count acceptance test' policy will prevent the bootstrap if deployed first
			waitForLeafClusterAvailability(leafCluster.Name, "Ready")
			addKustomizationBases(leafCluster.Type, leafCluster.Name, leafCluster.Namespace)

			ginkgo.By(fmt.Sprintf("And verify '%s' leafCluster is bootstraped", leafCluster.Name), func() {
				useClusterContext(leafClusterContext)
				verifyFluxControllers(GITOPS_DEFAULT_NAMESPACE)
				waitForGitRepoReady("flux-system", GITOPS_DEFAULT_NAMESPACE)
			})

			// Add/Install test Policies and Policy Config on the leaf cluster
			installTestPolicies(leafCluster.Name, policiesYaml)
			installPolicyConfig(leafCluster.Name, policyConfigYaml)

			// Add GitRepository source to leaf cluster
			addSource("git", podinfo.Source, podinfo.Namespace, sourceURL, "master", "")
			useClusterContext(mgmtClusterContext)

			pages.NavigateToPage(webDriver, "Applications")

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
			_ = createGitopsPR(pullRequest)

			ginkgo.By("Then merge the pull request to start application reconciliation", func() {
				createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
				mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
			})

			ginkgo.By("Then force reconcile leaf cluster flux-system for immediate application availability", func() {
				useClusterContext(leafClusterContext)
				reconcile("reconcile", "source", "git", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
				reconcile("reconcile", "", "kustomization", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
				useClusterContext(mgmtClusterContext)
			})

			ginkgo.By(fmt.Sprintf("And wait for leaf cluster %s application to be visibe on the dashboard", podinfo.Name), func() {
				gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())

				totalAppCount := existingAppCount + 1 // podinfo (leaf cluster)
				gomega.Eventually(applicationsPage.CountApplications, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(totalAppCount), fmt.Sprintf("There should be %d application enteries in application table, but found %d", totalAppCount, existingAppCount))
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

			verifyAppViolationsList(podinfo, appViolations)
			verifyAppViolationsDetailsPage(leafCluster.Name, podinfo, appViolations)
			verifyPolicyConfigInAppViolationsDetails(appViolations.ConfigPolicy, appViolations.PolicyConfigViolationMsg)
			verifyDeleteApplication(applicationsPage, existingAppCount, podinfo.Name, appKustomization)

		})

	})
})
