package acceptance

import (
	"fmt"
	"os/exec"
	"path"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

func AuthenticateWithGitProvider(webDriver *agouti.Page, gitProvider string) bool {
	if gitProvider == "github" {
		authenticate := pages.AuthenticateWithGithub(webDriver)

		if pages.ElementExist(authenticate.AuthenticateGithub) {
			log.Info("Found, authing...")
			Expect(authenticate.AuthenticateGithub.Click()).To(Succeed())
			Eventually(authenticate.AccessCode).Should(BeVisible())
			accessCode, _ := authenticate.AccessCode.Text()
			Expect(authenticate.AuthroizeButton.Click()).To(Succeed())
			accessCode = strings.Replace(accessCode, "-", "", 1)

			// Move to device activation window
			TakeScreenShot("application_authentication")
			Expect(webDriver.NextWindow()).ShouldNot(HaveOccurred(), "Failed to switch to github authentication window")
			TakeScreenShot("github_authentication")

			activate := pages.ActivateDeviceGithub(webDriver)

			if pages.ElementExist(activate.Username) {
				Eventually(activate.Username).Should(BeVisible())
				Expect(activate.Username.SendKeys(GITHUB_USER)).To(Succeed())
				Expect(activate.Password.SendKeys(GITHUB_PASSWORD)).To(Succeed())
				Expect(activate.Signin.Click()).To(Succeed())
			} else {
				log.Info("Login not found, assuming already logged in")
				TakeScreenShot("login_skipped")
			}

			if pages.ElementExist(activate.AuthCode) {
				Eventually(activate.AuthCode).Should(BeVisible())
				// Generate 6 digit authentication OTP for MFA
				authCode, _ := runCommandAndReturnStringOutput("totp-cli instant")
				Expect(activate.AuthCode.SendKeys(authCode)).To(Succeed())
			} else {
				log.Info("OTP not found, assuming already logged in")
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
			Eventually(authenticate.AuthroizeButton).ShouldNot(BeFound())
			return true
		}
	}
	return false
}

func DescribeApplications(gitopsTestRunner GitopsTestRunner) {
	var _ = Describe("Gitops application UI Tests", func() {

		GITOPS_BIN_PATH := GetGitopsBinPath()
		var repoAbsolutePath string

		BeforeEach(func() {

			By("Given I have a gitops binary installed on my local machine", func() {
				Expect(FileExists(GITOPS_BIN_PATH)).To(BeTrue(), fmt.Sprintf("%s can not be found.", GITOPS_BIN_PATH))
			})

			By("Given Kubernetes cluster is setup", func() {
				gitopsTestRunner.CheckClusterService(GetCapiEndpointUrl())
			})

			InitializeWebdriver(GetWGEUrl())
		})

		AfterEach(func() {

		})

		Context("[CLI] When Wego core and enterprise are installed in the cluster", func() {
			appName := "nginx"
			appNamespace := "my-nginx"
			appPath := "nginx-app"
			kustomizationFile := "../../utils/data/nginx.yaml"
			kustomizationCommitMsg := "edit nginx kustomization repo file"

			JustBeforeEach(func() {
				By("And cluster repo does not already exist", func() {
					gitopsTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
					_ = deleteDirectory([]string{path.Join("/tmp/", CLUSTER_REPOSITORY)})
				})

			})

			JustAfterEach(func() {
				SusspendGitopsApplication(appName, GITOPS_DEFAULT_NAMESPACE)
				DeleteGitopsApplication(appName, GITOPS_DEFAULT_NAMESPACE)
				DeleteGitopsDeploySecret(GITOPS_DEFAULT_NAMESPACE)

				_ = gitopsTestRunner.KubectlDelete([]string{}, kustomizationFile)
				gitopsTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
				_ = deleteDirectory([]string{path.Join("/tmp/", CLUSTER_REPOSITORY)})
			})

			It("@application Verify application's status and history can be monitored.", func() {
				By("When I create a private repository for cluster configs", func() {
					repoAbsolutePath = gitopsTestRunner.InitAndCreateEmptyRepo(CLUSTER_REPOSITORY, true)
					testFile := createTestFile("README.md", "# gitops-capi-template")

					gitopsTestRunner.GitAddCommitPush(repoAbsolutePath, testFile)
				})

				By("When I install gitops/wego to my active cluster", func() {
					InstallAndVerifyGitops(GITOPS_DEFAULT_NAMESPACE, GetGitRepositoryURL(repoAbsolutePath))
				})

				By("And I add the kustomization file for application deployment", func() {
					_ = runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("mkdir -p %[1]v && cp %[2]v %[1]v", path.Join(repoAbsolutePath, appPath), kustomizationFile))
					GitUpdateCommitPush(repoAbsolutePath, kustomizationCommitMsg)
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

					Eventually(addApp.Name).Should(BeVisible())
					Expect(addApp.Name.SendKeys(appName)).To(Succeed())
					Expect(addApp.SourceRepoUrl.SendKeys(GetGitRepositoryURL(repoAbsolutePath))).To(Succeed())
					Expect(addApp.ConfigRepoUrl.SendKeys(GetGitRepositoryURL(repoAbsolutePath))).To(Succeed())
					Expect(addApp.Path.SendKeys(appPath)).To(Succeed())
					Expect(addApp.AutoMerge.Check()).To(Succeed())
				})

				By(`And authenticate with Github`, func() {
					if AuthenticateWithGitProvider(webDriver, "github") {
						Eventually(addApp.GitCredentials).Should(BeVisible())
					}
					addApp = pages.GetAddApplicationForm(webDriver)
					Expect(addApp.Submit.Click()).To(Succeed(), "Failed to click application add Submit button")
					pages.WaitForAuthenticationAlert(webDriver, "Application added successfully!")
					Expect(addApp.ViewApplication.Click()).To(Succeed())
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
					Eventually(appDetails.URL).Should(MatchText(fmt.Sprintf(`ssh://git.+%s/%s.+`, GITHUB_ORG, CLUSTER_REPOSITORY)))
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

				By(`Then authenticate with Github`, func() {
					if AuthenticateWithGitProvider(webDriver, "github") {
						pages.WaitForAuthenticationAlert(webDriver, "Authentication Successful")
					}
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
