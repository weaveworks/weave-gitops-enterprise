package acceptance

import (
	"fmt"
	"regexp"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func DescribeMiscellaneousCli(gitopsTestRunner GitopsTestRunner) {
	var _ = ginkgo.Describe("Gitops miscellaneous CLI tests", ginkgo.Label("cli"), func() {

		templateFiles := []string{}
		var stdOut string
		var stdErr string

		ginkgo.BeforeEach(func() {

		})

		ginkgo.AfterEach(func() {
			gitopsTestRunner.DeleteApplyCapiTemplates(templateFiles)
			templateFiles = []string{}
		})

		ginkgo.Context("[CLI] When entitlement is available in the cluster", func() {
			var resourceName string
			DEPLOYMENT_APP := "my-mccp-cluster-service"

			checkEntitlement := func(typeEntitelment string, beFound bool) {
				checkOutput := func() bool {
					cmd := fmt.Sprintf(`get %s`, resourceName)
					stdOut, stdErr = runGitopsCommand(cmd, ASSERTION_1MINUTE_TIME_OUT)

					msg := stdErr + " " + stdOut

					if typeEntitelment == "expired" {
						re := regexp.MustCompile(`Your entitlement for Weave GitOps Enterprise has expired`)
						return re.MatchString(msg)
					}
					re := regexp.MustCompile(`No entitlement was found for Weave GitOps Enterprise`)
					return re.MatchString(msg)

				}

				matcher := gomega.BeFalse
				if beFound {
					matcher = gomega.BeTrue
				}

				resourceName = "templates"
				logger.Infof("Running 'gitops get %s --endpoint %s'", resourceName, capi_endpoint_url)
				gomega.Eventually(checkOutput, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(matcher())
				resourceName = "credentials"
				logger.Infof("Running 'gitops get %s --endpoint %s'", resourceName, capi_endpoint_url)
				gomega.Eventually(checkOutput, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(matcher())
				resourceName = "clusters"
				logger.Infof("Running 'gitops get %s --endpoint %s'", resourceName, capi_endpoint_url)
				gomega.Eventually(checkOutput, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(matcher())
			}

			ginkgo.JustBeforeEach(func() {

				templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "templates/cluster/docker/cluster-template.yaml")
				gitopsTestRunner.CreateIPCredentials("AWS")
			})

			ginkgo.JustAfterEach(func() {
				ginkgo.By("When I apply the valid entitlement", func() {
					gomega.Expect(gitopsTestRunner.KubectlApply([]string{}, "../../utils/data/entitlement/entitlement-secret.yaml"), "Failed to create/configure entitlement")
				})

				ginkgo.By("Then I restart the cluster service pod for valid entitlemnt to take effect", func() {
					gomega.Expect(gitopsTestRunner.RestartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
				})

				ginkgo.By("And the Cluster service is healthy", func() {
					CheckClusterService(capi_endpoint_url)
				})

				ginkgo.By("And I should not see the error or warning message for valid entitlement", func() {
					checkEntitlement("expired", false)
					checkEntitlement("invalid", false)
				})

				gitopsTestRunner.DeleteIPCredentials("AWS")
				// Login to the dashbord because the logout automatically when the cluster service restarts for entitlement checking
				loginUser()
			})

			ginkgo.It("Verify cluster service acknowledges the entitlement presences", func() {

				ginkgo.By("When I delete the entitlement", func() {
					gomega.Expect(gitopsTestRunner.KubectlDelete([]string{}, "../../utils/data/entitlement/entitlement-secret.yaml"), "Failed to delete entitlement secret")
				})

				ginkgo.By("Then I restart the cluster service pod for missing entitlemnt to take effect", func() {
					gomega.Expect(gitopsTestRunner.RestartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE)).ShouldNot(gomega.HaveOccurred(), "Failed restart deployment successfully")
				})

				ginkgo.By("And I should see the error message for missing entitlement", func() {
					checkEntitlement("missing", true)
				})

				ginkgo.By("When I apply the expired entitlement", func() {
					gomega.Expect(gitopsTestRunner.KubectlApply([]string{}, "../../utils/data/entitlement/entitlement-secret-expired.yaml"), "Failed to create/configure entitlement")
				})

				ginkgo.By("Then I restart the cluster service pod for expired entitlemnt to take effect", func() {
					gomega.Expect(gitopsTestRunner.RestartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
				})

				ginkgo.By("And I should see the warning message for expired entitlement", func() {
					checkEntitlement("expired", true)
				})

				ginkgo.By("When I apply the invalid entitlement", func() {
					gomega.Expect(gitopsTestRunner.KubectlApply([]string{}, "../../utils/data/entitlement/entitlement-secret-invalid.yaml"), "Failed to create/configure entitlement")
				})

				ginkgo.By("Then I restart the cluster service pod for invalid entitlemnt to take effect", func() {
					gomega.Expect(gitopsTestRunner.RestartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
				})

				ginkgo.By("And I should see the error message for invalid entitlement", func() {
					checkEntitlement("invalid", true)
				})
			})
		})

	})
}
