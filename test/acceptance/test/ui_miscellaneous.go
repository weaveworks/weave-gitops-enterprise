package acceptance

import (
	"path"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

var _ = ginkgo.Describe("Multi-Cluster Control Plane miscellaneous UI tests", func() {

	ginkgo.BeforeEach(func() {
		gomega.Expect(webDriver.Navigate(test_ui_url)).To(gomega.Succeed())

		if !pages.ElementExist(pages.Navbar(webDriver).Title, 3) {
			loginUser()
		}
	})

	ginkgo.AfterEach(func() {

	})

	ginkgo.Context("[UI] When entitlement is available in the cluster", func() {
		DEPLOYMENT_APP := "my-mccp-cluster-service"

		checkEntitlement := func(typeEntitelment string, beFound bool) {
			checkOutput := func() bool {
				if !pages.ElementExist(pages.GetClustersPage(webDriver).Version) {
					gomega.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
				}
				loginUser()
				messages := pages.GetMessages(webDriver)
				switch typeEntitelment {
				case "expired":
					if errMsg, _ := messages.Warning.Text(); strings.Contains(errMsg, "Your entitlement for Weave GitOps Enterprise has expired") {
						return true
					}
				case "invalid", "missing":
					if errMsg, _ := messages.Error.Text(); strings.Contains(errMsg, "No entitlement was found for Weave GitOps Enterprise") {
						return true
					}
				}
				return false
			}

			gomega.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())

			if beFound {
				gomega.Eventually(checkOutput, ASSERTION_2MINUTE_TIME_OUT).Should(gomega.BeTrue())
			} else {
				gomega.Eventually(checkOutput, ASSERTION_2MINUTE_TIME_OUT).Should(gomega.BeFalse())
			}

		}

		ginkgo.JustAfterEach(func() {
			ginkgo.By("When I apply the valid entitlement", func() {
				gomega.Expect(gitopsTestRunner.KubectlApply([]string{}, path.Join(testDataPath, "entitlement/entitlement-secret.yaml")), "Failed to create/configure entitlement")
			})

			ginkgo.By("Then I restart the cluster service pod for valid entitlemnt to take effect", func() {
				gomega.Expect(gitopsTestRunner.RestartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
			})

			ginkgo.By("And the Cluster service is healthy", func() {
				CheckClusterService(wge_endpoint_url)
			})

			ginkgo.By("And I should not see the error or warning message for valid entitlement", func() {
				checkEntitlement("expired", false)
				checkEntitlement("missing", false)
			})
		})

		ginkgo.It("Verify cluster service acknowledges the entitlement presences", ginkgo.Label("integration"), func() {

			ginkgo.By("When I delete the entitlement", func() {
				gomega.Expect(gitopsTestRunner.KubectlDelete([]string{}, path.Join(testDataPath, "entitlement/entitlement-secret.yaml")), "Failed to delete entitlement secret")
			})

			ginkgo.By("Then I restart the cluster service pod for missing entitlemnt to take effect", func() {
				gomega.Expect(gitopsTestRunner.RestartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE)).ShouldNot(gomega.HaveOccurred(), "Failed restart deployment successfully")
			})

			ginkgo.By("And I should see the error message for missing entitlement", func() {
				checkEntitlement("missing", true)

			})

			ginkgo.By("When I apply the expired entitlement", func() {
				gomega.Expect(gitopsTestRunner.KubectlApply([]string{}, path.Join(testDataPath, "entitlement/entitlement-secret-expired.yaml")), "Failed to create/configure entitlement")
			})

			ginkgo.By("Then I restart the cluster service pod for expired entitlemnt to take effect", func() {
				gomega.Expect(gitopsTestRunner.RestartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
			})

			ginkgo.By("And I should see the warning message for expired entitlement", func() {
				checkEntitlement("expired", true)
			})

			ginkgo.By("When I apply the invalid entitlement", func() {
				gomega.Expect(gitopsTestRunner.KubectlApply([]string{}, path.Join(testDataPath, "entitlement/entitlement-secret-invalid.yaml")), "Failed to create/configure entitlement")
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
