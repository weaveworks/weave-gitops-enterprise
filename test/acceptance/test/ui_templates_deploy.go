package acceptance

import (
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

func verifySelfServicePodinfo(podinfo Application, webUrl, bgColour, message string) {
	windowName := "podinfo"
	currentWindow, err := webDriver.Session().GetWindow()
	gomega.Expect(err).To(gomega.BeNil(), "Failed to get current/active window")

	pages.OpenWindowInBg(webDriver, webUrl, windowName)
	gomega.Expect(webDriver.NextWindow()).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to switch to '%s' window", windowName))

	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
		g.Eventually(webDriver.Title, ASSERTION_10SECONDS_TIME_OUT).Should(gomega.MatchRegexp(strings.Join([]string{podinfo.TargetNamespace, podinfo.Name}, "-")))
	}, ASSERTION_1MINUTE_TIME_OUT).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to verify '%s' window title", windowName))

	gomega.Eventually(webDriver.Find(fmt.Sprintf(`div[style*="background-color: %s"]`, bgColour))).Should(matchers.BeFound(), fmt.Sprintf("Failed to verify '%s' background colour", "podinfoApp"))
	gomega.Eventually(webDriver.Find(`div h1`).Text).Should(gomega.MatchRegexp(message), fmt.Sprintf("Failed to verify '%s' message", "podinfoApp"))
	time.Sleep(POLL_INTERVAL_1SECONDS)
	gomega.Eventually(webDriver.CloseWindow).Should(gomega.Succeed(), fmt.Sprintf("Failed to close '%s' window", windowName))
	gomega.Expect(webDriver.Session().SetWindow(currentWindow)).ShouldNot(gomega.HaveOccurred(), "Failed to switch to weave gitops enterprise dashboard")
}

var _ = ginkgo.Describe("Multi-Cluster Control Plane GitOpsTemplates for deployments", ginkgo.Label("ui", "template"), func() {
	var templateNamespaces []string

	ginkgo.BeforeEach(func() {
		gomega.Expect(webDriver.Navigate(testUiUrl)).To(gomega.Succeed())

		if !pages.ElementExist(pages.Navbar(webDriver).Title, 3) {
			loginUser()
		}
	})

	ginkgo.AfterEach(func() {
		_ = runCommandPassThrough("kubectl", "delete", "CapiTemplate", "--all")
		_ = runCommandPassThrough("kubectl", "delete", "GitOpsTemplate", "--all")
		deleteNamespace(templateNamespaces)
	})

	ginkgo.Context("[UI] GitOps Template can create resources in the management cluster", ginkgo.Label("deploy", "application"), func() {
		var existingAppCount int
		appNameSpace := "test-system"
		appTargetNamespace := "test-system"

		mgmtCluster := ClusterConfig{
			Type:      "management",
			Name:      "management",
			Namespace: "",
		}
		templateRepoPath := "./clusters/management/clusters"

		ginkgo.JustBeforeEach(func() {
			createNamespace([]string{appNameSpace, appTargetNamespace})
		})

		ginkgo.JustAfterEach(func() {
			pages.CloseOtherWindows(webDriver, enterpriseWindow)

			// Force clean the repository directory for subsequent tests
			cleanGitRepository(templateRepoPath)
			gomega.Eventually(func(g gomega.Gomega) int {
				return getApplicationCount()
			}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Equal(existingAppCount), fmt.Sprintf("There should be %d application enteries after application(s) deletion", existingAppCount))

			deleteNamespace([]string{appNameSpace, appTargetNamespace})
		})

		ginkgo.It("Verify self service MVP deployment workflow for management cluster", func() {
			repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
			existingAppCount = getApplicationCount()

			templateName := "mvp-helmrelease-template"
			templateNamespace := "default"
			templateFiles := map[string]string{
				templateName: path.Join(testDataPath, "templates/application/mvp-helmrelease-template.yaml"),
			}

			installGitOpsTemplate(templateFiles)
			pages.NavigateToPage(webDriver, "Templates")
			waitForTemplatesToAppear(len(templateFiles))
			templatesPage := pages.GetTemplatesPage(webDriver)

			ginkgo.By("And I should choose a template", func() {
				templateRow := templatesPage.GetTemplateInformation(webDriver, templateName)
				gomega.Eventually(templateRow.CreateTemplate.Click, ASSERTION_2MINUTE_TIME_OUT).Should(gomega.Succeed())
			})

			createPage := pages.GetCreateClusterPage(webDriver)
			ginkgo.By("And wait for Create resource page to be fully rendered", func() {
				pages.WaitForPageToLoad(webDriver)
				gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(".*Create new resource.*"))
			})

			// Deployment application template parameters
			podinfo := Application{
				Name:            "podinfo",
				Namespace:       appNameSpace,
				Type:            "helm_release",
				Url:             "https://stefanprodan.github.io/podinfo",
				Chart:           "podinfo",
				Version:         "6.2.3",
				TargetNamespace: appTargetNamespace,
				Values:          `"{ui: {color: \"#7c4e34\", message: \"Hello World\"}}"`,
			}
			// Ingress template parameters for deploymnet application
			ingressHost := "mvp.wge.com" // ingress host must be resolvable
			appServiceName := strings.Join([]string{podinfo.TargetNamespace, podinfo.Name}, "-")
			appServicePort := "9898" // default podinfo service
			podinfoWebUrl := fmt.Sprintf(`https://%s:%s`, ingressHost, GetEnv("UI_NODEPORT", "30080"))

			ginkgo.By("And set cluster hostname mapping in the /etc/hosts file for deploy service", func() {
				err := runCommandPassThrough(path.Join(getCheckoutRepoPath(), "test", "utils", "scripts", "hostname-to-ip.sh"), ingressHost)
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to set deployment service hostname entry in /etc/hosts file")
			})

			// Not setting the helm chart values via parameters
			var parameters = []TemplateField{
				{
					Name:  "RESOURCE_NAME",
					Value: podinfo.Name,
				},
				{
					Name:  "NAMESPACE",
					Value: podinfo.Namespace,
				},
				{
					Name:  "URL",
					Value: podinfo.Url,
				},
				{
					Name:  "CHART_NAME",
					Value: podinfo.Chart,
				},
				{
					Name:  "CHART_VERSION",
					Value: podinfo.Version,
				},
				{
					Name:  "TARGET_NAMESPACE",
					Value: podinfo.TargetNamespace,
				},
				{
					Name:  "HOST_NAME",
					Value: ingressHost,
				},
				{
					Name:  "SERVICE_NAME",
					Value: appServiceName,
				},
				{
					Name:  "SERVICE_PORT",
					Value: appServicePort,
				},
			}

			setParameterValues(createPage, parameters)

			preview := pages.GetPreview(webDriver)
			ginkgo.By("Then I should preview the PR", func() {
				gomega.Eventually(func(g gomega.Gomega) {
					g.Expect(createPage.PreviewPR.Click()).Should(gomega.Succeed())
					g.Expect(preview.Title.Text()).Should(gomega.MatchRegexp("PR Preview"))

				}, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Succeed(), "Failed to get PR preview")
			})

			ginkgo.By("Then verify all resources are labelled with template name and namespace", func() {
				// Verify resource definition preview
				gomega.Eventually(preview.GetPreviewTab("Resource Definition").Click).Should(gomega.Succeed(), "Failed to switch to 'RESOURCE DEFINITION' preview tab")
				previewText, _ := preview.Text.Text()

				re, _ := regexp.Compile(fmt.Sprintf("templates.weave.works/template-name: %s", templateName))
				matches := re.FindAllString(previewText, -1)
				gomega.Eventually(len(matches)).Should(gomega.Equal(3), "All resources should be labelled with template name")

				re, _ = regexp.Compile(fmt.Sprintf("templates.weave.works/template-namespace: %s", templateNamespace))
				matches = re.FindAllString(previewText, -1)
				gomega.Eventually(len(matches)).Should(gomega.Equal(3), "All resources should be labelled with template namespace")

				gomega.Eventually(preview.Close.Click).Should(gomega.Succeed(), "Failed to close the preview dialog")
			})

			// Pull request values
			pullRequest := PullRequest{
				Branch:  fmt.Sprintf("br-mvp-%s", strings.ToLower(randString(5))),
				Title:   "Self service MVP deployment",
				Message: "An exapmle self service MVP podinfo deployment",
			}
			_ = createGitopsPR(pullRequest)

			ginkgo.By("Then I should merge the pull request to start resource provisioning", func() {
				createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
				mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
			})

			ginkgo.By("Then force reconcile flux-system to immediately start resource provisioning", func() {
				reconcile("reconcile", "source", "git", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
				reconcile("reconcile", "", "kustomization", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
			})

			pages.NavigateToPage(webDriver, "Applications")
			applicationsPage := pages.GetApplicationsPage(webDriver)

			ginkgo.By(fmt.Sprintf("And wait for %s application to be visibe on the dashboard", podinfo.Name), func() {
				gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())

				totalAppCount := existingAppCount + 1
				gomega.Eventually(applicationsPage.CountApplications, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(totalAppCount), fmt.Sprintf("There should be %d application enteries in application table", totalAppCount))
			})

			verifyAppInformation(applicationsPage, podinfo, mgmtCluster, "Ready")

			// Edit the podinfo application bakground colour and message
			applicationInfo := applicationsPage.FindApplicationInList(podinfo.Name)
			ginkgo.By(fmt.Sprintf("And navigate to %s application page", podinfo.Name), func() {
				gomega.Eventually(applicationInfo.Name.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to %s application detail page", podinfo.Name))
			})
			verifyAppPage(podinfo)

			bgColour := "rgb(52, 87, 124)"      // default podinfo background colour '#34577c'
			message := "greetings from podinfo" // default podinfo message
			verifySelfServicePodinfo(podinfo, podinfoWebUrl, bgColour, message)

			ginkgo.By(fmt.Sprintf("Then I should start editing the %s application", podinfo.Name), func() {
				appDetailPage := pages.GetApplicationsDetailPage(webDriver, podinfo.Type)
				gomega.Eventually(appDetailPage.Edit.Click, ASSERTION_30SECONDS_TIME_OUT).Should(gomega.Succeed(), "Failed to click 'EDIT' application button")
			})

			createPage = pages.GetCreateClusterPage(webDriver)
			ginkgo.By("And wait for Edit resource page to be fully rendered", func() {
				pages.WaitForPageToLoad(webDriver)
				gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(podinfo.Name))
			})

			// Only editing setting helm chart values for the application
			parameters = []TemplateField{
				{
					Name:  "VALUES",
					Value: podinfo.Values,
				},
			}
			setParameterValues(createPage, parameters)

			_ = createGitopsPR(PullRequest{}) // accepts default pull request values
			ginkgo.By("Then I should merge the pull request to start resource provisioning", func() {
				createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
				mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
			})

			ginkgo.By("Then force reconcile flux-system to immediately start resource provisioning", func() {
				reconcile("reconcile", "source", "git", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
				reconcile("reconcile", "", "kustomization", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
			})

			bgColour = "rgb(124, 78, 52)" // edited podinfo background colour '#7c4e34'
			message = "Hello World"       // edited podinfo message
			verifySelfServicePodinfo(podinfo, podinfoWebUrl, bgColour, message)

			verifyDeleteApplication(applicationsPage, existingAppCount, podinfo.Name, templateRepoPath)
		})
	})
})
