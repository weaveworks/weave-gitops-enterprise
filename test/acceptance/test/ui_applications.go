package acceptance

import (
	"fmt"
	"os/exec"
	"path"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

func AuthenticateWithGitProvider(webDriver *agouti.Page, gitProvider string) {
	if gitProvider == GitProviderGitHub {
		authenticate := pages.AuthenticateWithGithub(webDriver)

		if pages.ElementExist(authenticate.AuthenticateGithub) {
			Expect(authenticate.AuthenticateGithub.Click()).To(Succeed())
			AuthenticateWithGitHub(webDriver)

			// Sometimes authentication failed to get the github device code, it may require revalidation with new access code
			if pages.ElementExist(authenticate.AuthorizationError) {
				logger.Info("Error getting github device code, requires revalidating...")
				Expect(authenticate.Close.Click()).To(Succeed())
				Eventually(authenticate.AuthenticateGithub.Click).Should(Succeed())
				AuthenticateWithGitHub(webDriver)
			}

			Eventually(authenticate.AuthroizeButton).ShouldNot(BeFound())
		}
	} else if gitProvider == GitProviderGitLab {
		authenticate := pages.AuthenticateWithGitlab(webDriver)

		if pages.ElementExist(authenticate.AuthenticateGitlab) {
			Expect(authenticate.AuthenticateGitlab.Click()).To(Succeed())

			if !pages.ElementExist(authenticate.Username) {
				if pages.ElementExist(authenticate.CheckBrowser) {
					setGitlabBrowserCompatibility(webDriver)
					Eventually(authenticate.CheckBrowser, ASSERTION_30SECONDS_TIME_OUT).ShouldNot(BeFound())
					TakeScreenShot("gitlab_browser_compatibility")
				}

				if pages.ElementExist(authenticate.AcceptCookies, 10) {
					Eventually(authenticate.AcceptCookies.Click).Should(Succeed())
				}
			}

			TakeScreenShot("gitlab_cookies_accepted")
			if pages.ElementExist(authenticate.Username) {
				Eventually(authenticate.Username).Should(BeVisible())
				Expect(authenticate.Username.SendKeys(gitProviderEnv.Username)).To(Succeed())
				Expect(authenticate.Password.SendKeys(gitProviderEnv.Password)).To(Succeed())
				Expect(authenticate.Signin.Click()).To(Succeed())
			} else {
				logger.Info("Login not found, assuming already logged in")
			}

			if pages.ElementExist(authenticate.Authorize) {
				Expect(authenticate.Authorize.Click()).To(Succeed())
			}
		}
	}
}

func setGitlabBrowserCompatibility(webDriver *agouti.Page) {
	// opening the gitlab in a separate window not controlled by webdriver seems to redirect gitlab to login
	pages.OpenNewWindow(webDriver, `http://`+gitProviderEnv.Hostname+`/users/sign_in`, "gitlab")
	// Make sure weave-gitops-enterprise application window is still active window
	Expect(webDriver.SwitchToWindow(WGE_WINDOW_NAME)).ShouldNot(HaveOccurred(), "Failed to switch to wego application window")
}

func AuthenticateWithGitHub(webDriver *agouti.Page) {

	authenticate := pages.AuthenticateWithGithub(webDriver)

	Eventually(authenticate.AccessCode).Should(BeVisible())
	accessCode, _ := authenticate.AccessCode.Text()
	Expect(authenticate.AuthroizeButton.Click()).To(Succeed())
	accessCode = strings.Replace(accessCode, "-", "", 1)
	logger.Info(accessCode)

	// Move to device activation window
	TakeScreenShot("application_authentication")
	Expect(webDriver.NextWindow()).ShouldNot(HaveOccurred(), "Failed to switch to github authentication window")
	TakeScreenShot("github_authentication")

	activate := pages.ActivateDeviceGithub(webDriver)

	if pages.ElementExist(activate.Username) {
		Eventually(activate.Username).Should(BeVisible())
		Expect(activate.Username.SendKeys(gitProviderEnv.Username)).To(Succeed())
		Expect(activate.Password.SendKeys(gitProviderEnv.Password)).To(Succeed())
		Expect(activate.Signin.Click()).To(Succeed())
	} else {
		logger.Info("Login not found, assuming already logged in")
		TakeScreenShot("login_skipped")
	}

	if pages.ElementExist(activate.AuthCode) {
		Eventually(activate.AuthCode).Should(BeVisible())
		// Generate 6 digit authentication OTP for MFA
		authCode, _ := runCommandAndReturnStringOutput("totp-cli instant")
		Expect(activate.AuthCode.SendKeys(authCode)).To(Succeed())
	} else {
		logger.Info("OTP not found, assuming already logged in")
		TakeScreenShot("otp_skipped")
	}

	Eventually(activate.Continue).Should(BeVisible())
	Expect(activate.UserCode.At(0).SendKeys(accessCode)).To(Succeed())
	Expect(activate.Continue.Click()).To(Succeed())

	Eventually(activate.AuthroizeWeaveworks).Should(BeEnabled())
	Expect(activate.AuthroizeWeaveworks.Click()).To(Succeed())

	Eventually(activate.ConnectedMessage).Should(BeVisible())
	Expect(webDriver.CloseWindow()).ShouldNot(HaveOccurred())

	// Device is connected, now move back to application window
	Expect(webDriver.SwitchToWindow(WGE_WINDOW_NAME)).ShouldNot(HaveOccurred(), "Failed to switch to wego application window")
}

func DescribeApplications(gitopsTestRunner GitopsTestRunner) {
	var _ = Describe("Gitops application UI Tests", func() {

		BeforeEach(func() {

			By("Given I have a gitops binary installed on my local machine", func() {
				Expect(fileExists(gitops_bin_path)).To(BeTrue(), fmt.Sprintf("%s can not be found.", gitops_bin_path))
			})

			By("Given Kubernetes cluster is setup", func() {
				gitopsTestRunner.CheckClusterService(capi_endpoint_url)
			})

			initializeWebdriver(test_ui_url)
		})

		AfterEach(func() {

		})

		Context("[CLI] When Wego core and enterprise are installed in the cluster", func() {
			appName := "nginx"
			appNamespace := "my-nginx"
			appPath := "nginx-app"
			kustomizationFile := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "nginx.yaml")
			kustomizationCommitMsg := "edit nginx kustomization repo file"

			JustAfterEach(func() {
				susspendGitopsApplication(appName, GITOPS_DEFAULT_NAMESPACE)
				deleteGitopsApplication(appName, GITOPS_DEFAULT_NAMESPACE)
				deleteGitopsDeploySecret(GITOPS_DEFAULT_NAMESPACE)

				_ = gitopsTestRunner.KubectlDelete([]string{}, kustomizationFile)
			})

			It("@application @git Verify application's status and history can be monitored.", func() {
				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)

				By("When I create a private repository for cluster configs", func() {
					initAndCreateEmptyRepo(gitProviderEnv, true)
				})

				By("When I install gitops/wego to my active cluster", func() {
					installAndVerifyGitops(GITOPS_DEFAULT_NAMESPACE, getGitRepositoryURL(repoAbsolutePath))
				})

				By("And I add the kustomization file for application deployment", func() {
					pullGitRepo(repoAbsolutePath)
					_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("mkdir -p %[1]v && cp %[2]v %[1]v", path.Join(repoAbsolutePath, appPath), kustomizationFile))
					gitUpdateCommitPush(repoAbsolutePath, kustomizationCommitMsg)
				})

				pages.NavigateToPage(webDriver, "Applications")
				applicationsPage := pages.GetApplicationPage(webDriver)
				addApp := pages.GetAddApplicationForm(webDriver)

				By("And wait for Application page to be rendered", func() {
					applicationsPage.WaitForPageToLoad(webDriver, 0)
					Eventually(applicationsPage.ApplicationHeader).Should(BeVisible())
					Eventually(applicationsPage.AddApplication).Should(BeVisible())
				})

				By(fmt.Sprintf("Then I add application '%s' to weave gitops cluster", appName), func() {
					Expect(applicationsPage.AddApplication.Click()).To(Succeed())

					Expect(pages.ElementExist(addApp.Name)).To(BeTrue(), "Application name field doesn't exist")
					Expect(addApp.Name.SendKeys(appName)).To(Succeed())
					Expect(addApp.SourceRepoUrl.SendKeys(getGitRepositoryURL(repoAbsolutePath))).To(Succeed())
					Expect(addApp.ConfigRepoUrl.SendKeys(getGitRepositoryURL(repoAbsolutePath))).To(Succeed())
					Expect(addApp.Path.SendKeys(appPath)).To(Succeed())
					Expect(addApp.AutoMerge.Check()).To(Succeed())
				})

				By(`And authenticate with Git provider`, func() {
					AuthenticateWithGitProvider(webDriver, gitProviderEnv.Type)
					Eventually(addApp.GitCredentials).Should(BeVisible())

					addApp = pages.GetAddApplicationForm(webDriver)
					Expect(addApp.Submit.Click()).To(Succeed(), "Failed to click application add Submit button")
					pages.WaitForAuthenticationAlert(webDriver, "Application added successfully!")
					Eventually(addApp.ViewApplication.Click).Should(Succeed())
				})

				By("Then I should see gitops add command linked the repo to the cluster", func() {
					verifyWegoAddCommand(appName, GITOPS_DEFAULT_NAMESPACE)
				})

				By("And I should see workload is deployed to the cluster", func() {
					Expect(waitForResource("deploy", appName, appNamespace, "", ASSERTION_5MINUTE_TIME_OUT)).To(Succeed())
					Expect(waitForResource("pods", "", appNamespace, "", ASSERTION_5MINUTE_TIME_OUT)).To(Succeed())
					command := exec.Command("sh", "-c", fmt.Sprintf("kubectl wait --for=condition=Ready --timeout=60s -n %s --all pods --selector='app!=wego-app'", appNamespace))
					session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
					Eventually(session, ASSERTION_5MINUTE_TIME_OUT).Should(gexec.Exit())
				})

				By("And wait for Application page to be rendered", func() {
					Expect(webDriver.Refresh()).ShouldNot(HaveOccurred())
					applicationsPage.WaitForPageToLoad(webDriver, 1)
					Eventually(applicationsPage.ApplicationHeader).Should(BeVisible())
					Eventually(applicationsPage.AddApplication).Should(BeVisible())
				})

				By(fmt.Sprintf(`When I click on the application %s`, appName), func() {
					appRow := pages.GetApplicationRow(applicationsPage, appName)
					Expect(appRow.Application.Click()).To(Succeed())
				})

				appDetails := pages.GetApplicationDetails(webDriver)
				By(fmt.Sprintf(`Then %s details should be rendered`, appName), func() {
					appDetails.WaitForPageToLoad(webDriver)

					Eventually(appDetails.Name).Should(MatchText(appName))
					Eventually(appDetails.DeploymentType).Should(MatchText("Kustomize"))
					Eventually(appDetails.URL).Should(MatchText(fmt.Sprintf(`ssh://git.+%s/%s.+`, gitProviderEnv.Org, gitProviderEnv.Repo)))
					Eventually(appDetails.Path).Should(MatchText(appPath))
				})

				By(fmt.Sprintf(`And %s source/git status is available`, appName), func() {
					sourceCondition := pages.GetApplicationConditions(webDriver, "Source Conditions")
					Eventually(sourceCondition.Type).Should(MatchText("Ready"))
					Eventually(sourceCondition.Status).Should(MatchText("True"))
					Eventually(sourceCondition.Reason).Should(MatchText("GitOperationSucceed"))
					Eventually(sourceCondition.Message).Should(MatchText(`Fetched revision: main/[\w\d].+`))
				})

				By(fmt.Sprintf(`And %s automation/kustomization status is available`, appName), func() {
					sourceCondition := pages.GetApplicationConditions(webDriver, "Automation Conditions")
					Eventually(sourceCondition.Type).Should(MatchText("Ready"))
					Eventually(sourceCondition.Status, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(MatchText("True"))
					Eventually(sourceCondition.Reason).Should(MatchText("ReconciliationSucceeded"))
					Eventually(sourceCondition.Message).Should(MatchText(`Applied revision: main/[\w\d].+`))
				})

				By(`Then authenticate with Git provider`, func() {
					AuthenticateWithGitProvider(webDriver, gitProviderEnv.Type)
					Expect(webDriver.Refresh()).ShouldNot(HaveOccurred())
				})

				By(fmt.Sprintf(`And verify %s application commit history`, appName), func() {
					commits := pages.GetCommits(webDriver)
					commitFound := false
					for j := 0; j < len(commits); j++ {
						if msg, _ := commits[j].Message.Text(); msg == kustomizationCommitMsg {
							commitFound = true
							break
						}
					}
					Expect(commitFound).Should(BeTrue(), fmt.Sprintf(`'%s' commit message was not found in '%s' application's commit history`, kustomizationCommitMsg, appName))
				})
			})
		})
	})
}
