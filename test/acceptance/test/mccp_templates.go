package acceptance

import (
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/wks/test/acceptance/test/pages"
)

func DescribeMCCPTemplates(mccpTestRunner MCCPTestRunner) {

	var _ = Describe("Multi-Cluster Control Plane UI", func() {

		BeforeEach(func() {

			By("Given Kubernetes cluster is setup", func() {
				//TODO - Verify that cluster is up and running using kubectl
			})

			var err error
			if webDriver == nil {

				webDriver, err = agouti.NewPage(seleniumServiceUrl, agouti.Debug, agouti.Desired(agouti.Capabilities{
					"chromeOptions": map[string][]string{
						"args": {
							// "--headless", //Uncomment to run headless
							"--disable-gpu",
							"--no-sandbox",
						}}}))
				Expect(err).NotTo(HaveOccurred())

				// Make the page bigger so we can see all the things in the screenshots
				err = webDriver.Size(1440, 900)
				Expect(err).NotTo(HaveOccurred())
			}

			By("When I navigate to MCCP UI Page", func() {

				Expect(webDriver.Navigate(GetWkpUrl())).To(Succeed())

			})
		})

		AfterEach(func() {
			TakeNextScreenshot()
		})

		It("Verify template page rendering when no capiTemplate exists", func() {
			pages.NavigateToPage(webDriver, "Templates")
			templatesPage := pages.GetTemplatesPage(webDriver)

			By("And wait for Templates page to be rendered", func() {
				Eventually(templatesPage.TemplateHeader).Should(BeVisible())
				Eventually(templatesPage.TemplateCount).Should(MatchText(`0`))

				tileCount, _ := templatesPage.TemplateTiles.Count()
				Expect(tileCount).To(Equal(0), "There should not be any template tile rendered")

			})
		})

		It("Verify template(s) are rendered from the template library.", func() {

			noOfTemplates := 5
			templateFiles := mccpTestRunner.CreateApplyCapitemplates(noOfTemplates)

			pages.NavigateToPage(webDriver, "Templates")
			templatesPage := pages.GetTemplatesPage(webDriver)

			By("And wait for Templates page to be fully rendered", func() {
				Eventually(templatesPage.TemplateHeader).Should(BeVisible())
				Eventually(templatesPage.TemplateCount).Should(MatchText(`[0-9]+`))

				count, _ := templatesPage.TemplateCount.Text()
				templateCount, _ := strconv.Atoi(count)
				tileCount, _ := templatesPage.TemplateTiles.Count()

				Eventually(templateCount).Should(Equal(noOfTemplates), "The template header count should be equal to templates created")
				Eventually(tileCount).Should(Equal(noOfTemplates), "The number of template tiles rendered should be equal to templates created")
			})

			mccpTestRunner.DeleteApplyCapiTemplates(templateFiles)
		})

		It("Verify I should be able to select a template of my choice", func() {

			// test selection with 50 capiTemplates
			templateFiles := mccpTestRunner.CreateApplyCapitemplates(50)

			pages.NavigateToPage(webDriver, "Templates")
			// templatesPage := pages.GetTemplatesPage(webDriver)

			By("And User should choose a template", func() {
				templateTile := pages.GetTemplateTile(webDriver, "cluster-template-9")

				Eventually(templateTile.Description).Should(MatchText("This is test template 9"))
				Expect(templateTile.CreateTemplate).Should(BeFound())
				Expect(templateTile.CreateTemplate.Click()).To(Succeed())
			})

			By("And wait for Create cluster page to be fully rendered", func() {
				createPage := pages.GetCreateClusterPage(webDriver)
				Eventually(createPage.CreateHeader).Should(MatchText(".*Create new cluster.*"))
			})

			mccpTestRunner.DeleteApplyCapiTemplates(templateFiles)
		})
	})
}
