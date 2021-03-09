package acceptance

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/wks/test/acceptance/test/pages"
)

var n = 1

func TakeNextScreenshot() {
	TakeScreenShot(fmt.Sprintf("test-%v", n))
	n += 1
}

var _ = Describe("WKP UI", func() {

	BeforeEach(func() {

		By("Given Kubernetes cluster is setup", func() {
			//TODO - Verify that cluster is up and running using kubectl
		})

		var err error
		if webDriver == nil {

			webDriver, err = agouti.NewPage(seleniumServiceUrl, agouti.Debug, agouti.Desired(agouti.Capabilities{
				"chromeOptions": map[string][]string{
					"args": {
						//"--headless", //Uncomment to run headless
						"--disable-gpu",
						"--no-sandbox",
					}}}))
			Expect(err).NotTo(HaveOccurred())

			// Make the page bigger so we can see all the things in the screenshots
			err = webDriver.Size(1440, 3000)
			Expect(err).NotTo(HaveOccurred())
		}

		By("When I navigate to WKP dashboard", func() {

			Expect(webDriver.Navigate(wkpUrl)).To(Succeed())

		})
	})

	AfterEach(func() {
		TakeNextScreenshot()
		//Tear down
		//Expect(webDriver.Destroy()).To(Succeed())
	})

	It("Verify WKP Dashboard Page Structure", func() {

		var expectedWKPTitle = "Weave Kubernetes Platform"
		//var expectedClusterName = `/ gce-cluster`
		var expectedAlertInfo = "No alerts firing"
		var expectedDocLink = "/docs"
		var expectedGrafanaLink = "/grafana/d/all-nodes-resources/kubernetes-all-nodes-resources"
		var expectedComponentsLink = `https://%s.com/`

		By("Then I should see WKP UI dashboard with UI elements", func() {
			dashboardPage := pages.Dashboard(webDriver)
			By("-WKP Title and Logo", func() {
				Eventually(dashboardPage.WKPTitle).Should(HaveText(expectedWKPTitle))
			})

			By("-WKP Documentation Link", func() {
				Expect(dashboardPage.WKPDocLink.Attribute("href")).To(HaveSuffix(expectedDocLink))
			})

			By("-Cluster Name and Alert Info Text", func() {
				//Eventually(dashboardPage.ClusterName).Should(HaveText(expectedClusterName))
				Expect(dashboardPage.AlertInfo).Should(HaveText(expectedAlertInfo))
			})

			By("-Grafan Dashboard Link", func() {
				Expect(dashboardPage.GrafanaLink.Attribute("href")).To(HaveSuffix(expectedGrafanaLink))
			})

			By("-Add Components Link", func() {
				Expect(dashboardPage.AddComponentsLink.Attribute("href")).To(HavePrefix(fmt.Sprintf(expectedComponentsLink, gitProvider)))
			})

			By("-Open Git Repo Link", func() {
				Expect(dashboardPage.AddComponentsLink.Attribute("href")).To(HavePrefix(fmt.Sprintf(expectedComponentsLink, gitProvider)))
			})
		})

	})

	It("Verify Kubernetes Version", func() {

		var expectedVersion = "v1.20.0"
		dashboardPage := pages.Dashboard(webDriver)
		By("Then I should see the correct Kubernetes Version next to the cluster name", func() {
			Eventually(dashboardPage.K8SVersion).Should(HaveText(expectedVersion))
		})

	})

	It("Verify Cluster Components List", func() {

		componentsPage := pages.Components(webDriver)

		By("Then I should see components count", func() {

			Eventually(componentsPage.ClusterComponentsList).Should(HaveCount(12)) // hard coded components number???
		})

		By("And I should see following list of cluster components", func() {
			clusterComponents := []string{"External DNS", "Flux", "Flux helm operator", "Gitops repo broker", "Grafana", "Manifest loader", "Prometheus", "Scope", "Tiller", "UI"}
			for _, expectedName := range clusterComponents {
				cmp := pages.FindClusterComponent(componentsPage, expectedName)
				Expect(cmp).NotTo(Equal(nil))
				Expect(cmp.Name).To(Equal(expectedName))
				Expect(cmp.Status).To(Equal("ok"))
			}
		})
	})
})
