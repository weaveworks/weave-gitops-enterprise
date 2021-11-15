package acceptance

import (
	"fmt"
	"os/exec"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	. "github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

func DescribeApplications(gitopsTestRunner GitopsTestRunner) {
	var _ = Describe("Gitops upgrade Tests", func() {

		GITOPS_BIN_PATH := GetGitopsBinPath()
		var repoAbsolutePath string

		var session *gexec.Session
		var err error

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
			appName := "my-app"
			appPath := "namespace-app"

			JustBeforeEach(func() {
				By("And cluster repo does not already exist", func() {
					gitopsTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
					_ = deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})
				})

			})

			JustAfterEach(func() {
				SusspendGitopsApplication(appName, GITOPS_DEFAULT_NAMESPACE)
				DeleteGitopsApplication(appName, GITOPS_DEFAULT_NAMESPACE)
				DeleteGitopsDeploySecret(GITOPS_DEFAULT_NAMESPACE)

				gitopsTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
				_ = deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})
			})

			XIt("@application Verify application's status and history can be monitored", func() {

				By("When I create a private repository for cluster configs", func() {
					repoAbsolutePath = gitopsTestRunner.InitAndCreateEmptyRepo(CLUSTER_REPOSITORY, true)
					testFile := createTestFile("README.md", "# gitops-capi-template")

					gitopsTestRunner.GitAddCommitPush(repoAbsolutePath, testFile)
				})

				By("When I install gitops/wego to my active cluster", func() {
					InstallAndVerifyGitops(GITOPS_DEFAULT_NAMESPACE, GetGitRepositoryURL(repoAbsolutePath))
				})

				By("And I add the kustomization file for application deployment", func() {
					_ = runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("mkdir -p %[1]v && cp ../../utils/data/test_kustomization.yaml %[1]v", path.Join(repoAbsolutePath, appPath)))
					GitUpdateCommitPush(repoAbsolutePath)
				})

				addCommand := fmt.Sprintf("add app . --path=./%s  --name=%s  --auto-merge=true", appPath, appName)
				By(fmt.Sprintf("And I run gitops add app command ' %s 'in namespace %s from dir %s", addCommand, GITOPS_DEFAULT_NAMESPACE, repoAbsolutePath), func() {
					command := exec.Command("sh", "-c", fmt.Sprintf("cd %s && %s %s", repoAbsolutePath, GITOPS_BIN_PATH, addCommand))
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
					Eventually(session).Should(gexec.Exit())
					Expect(string(session.Err.Contents())).Should(BeEmpty())
				})

				By(fmt.Sprintf("And wait for %s/%s GitRepository resource to be available in cluster", GITOPS_DEFAULT_NAMESPACE, appName), func() {
					repoExists := func() bool {
						cmd := fmt.Sprintf(`kubectl get GitRepository %s -n %s`, appName, GITOPS_DEFAULT_NAMESPACE)
						out, _ := runCommandAndReturnStringOutput(cmd)

						return out != ""
					}
					Eventually(repoExists, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(BeTrue(), fmt.Sprintf("%s/%s Gitrepository resource does not exist in the cluster", GITOPS_DEFAULT_NAMESPACE, appName))

				})

				pages.NavigateToPage(webDriver, "Applications")
				applicationsPage := pages.GetApplicationPage(webDriver)

				By("And wait for Application page to be rendered", func() {
					applicationsPage.WaitForPageToLoad(webDriver, 1)
					Eventually(applicationsPage.ApplicationHeader).Should(BeVisible())
					Eventually(applicationsPage.AddApplication).Should(BeVisible())
					Eventually(applicationsPage.ApplicationCount).Should(MatchText(`1`))

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
			})
		})
	})
}
