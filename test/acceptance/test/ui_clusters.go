package acceptance

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

func ClusterStatusFromList(clustersPage *pages.ClustersPage, clusterName string) *agouti.Selection {
	return pages.FindClusterInList(clustersPage, clusterName).Status
}

func DescribeClusters(gitopsTestRunner GitopsTestRunner) {

	var _ = Describe("Multi-Cluster Control Plane Clusters", func() {

		BeforeEach(func() {
			Expect(webDriver.Navigate(test_ui_url)).To(Succeed())
		})

		It("@integration Verify Weave Gitops Enterprise version", func() {
			By("And I verify the version", func() {
				clustersPage := pages.GetClustersPage(webDriver)
				Eventually(clustersPage.Version).Should(BeFound())
				Expect(clustersPage.Version.Text()).Should(MatchRegexp(enterpriseChartVersion()), "Expected Weave Gitops enterprise version is not found")
			})
		})

		It("Verify page structure first time with no cluster configured", func() {
			if GetEnv("ACCEPTANCE_TESTS_DATABASE_TYPE", "") == "postgres" {
				Skip("This test case runs only with sqlite")
			}

			By("And wego enterprise state is reset", func() {
				gitopsTestRunner.ResetControllers("enterprise")
				gitopsTestRunner.VerifyWegoPodsRunning()
				Eventually(webDriver.Refresh).ShouldNot(HaveOccurred())
			})

			clustersPage := pages.GetClustersPage(webDriver)
			By("Then I should see the correct clusters count next to the clusters header", func() {
				Eventually(clustersPage.ClusterCount).Should(MatchText(`[0-9]+`))
			})

			// By("And should have 'Connect a cluster' button", func() {
			// 	Eventually(clustersPage.ConnectClusterButton).Should(HaveText(expectedConnectClusterLabel))
			// })

			By("And should have clusters list table", func() {
				Eventually(clustersPage.ClustersListSection).Should(BeFound())
			})

			By("And should have No clusters configured text", func() {
				Eventually(clustersPage.NoClusterConfigured).Should(HaveText("No clusters configured"))
			})

			By("And should have support email", func() {
				Expect(clustersPage.SupportEmailLink.Attribute("href")).To(HaveSuffix("mailto:support@weave.works"))
			})

			By("And should have No alerts firing message", func() {
				Expect(webDriver.Navigate(test_ui_url + "/clusters/alerts")).To(Succeed())
				Eventually(clustersPage.NoFiringAlertMessage).Should(HaveText("No alerts firing"))
			})

		})

		It("Verify that clusters table have correct column headers ", func() {

			clustersPage := pages.GetClustersPage(webDriver)

			By("Then I should see clusters table with Name column", func() {
				Eventually(clustersPage.HeaderName).Should(HaveText("Name"))
			})

			By("And with Status column", func() {
				Eventually(clustersPage.HeaderStatus).Should(HaveText("Status"))
			})
		})
	})
}
