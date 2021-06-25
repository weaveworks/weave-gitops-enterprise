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

		templateFiles := []string{}
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
			mccpTestRunner.DeleteApplyCapiTemplates(templateFiles)
		})

		Context("When no Capi Templates are available in the cluster", func() {
			It("Verify template page renders no capiTemplate", func() {
				pages.NavigateToPage(webDriver, "Templates")
				templatesPage := pages.GetTemplatesPage(webDriver)

				By("And wait for Templates page to be rendered", func() {
					Eventually(templatesPage.TemplateHeader).Should(BeVisible())
					Eventually(templatesPage.TemplateCount).Should(MatchText(`0`))

					tileCount, _ := templatesPage.TemplateTiles.Count()
					Expect(tileCount).To(Equal(0), "There should not be any template tile rendered")

				})
			})
		})

		Context("When Capi Templates are available in the cluster", func() {
			It("Verify template(s) are rendered from the template library.", func() {

				noOfTemplates := 5
				templateFiles = mccpTestRunner.CreateApplyCapitemplates(noOfTemplates, "capi-server-v1-capitemplate.yaml")

				pages.NavigateToPage(webDriver, "Templates")
				templatesPage := pages.GetTemplatesPage(webDriver)

				By("And wait for Templates page to be fully rendered", func() {
					Eventually(templatesPage.TemplateHeader).Should(BeVisible())
					Eventually(templatesPage.TemplateCount).Should(MatchText(`[0-9]+`))

					count, _ := templatesPage.TemplateCount.Text()
					templateCount, _ := strconv.Atoi(count)
					tileCount, _ := templatesPage.TemplateTiles.Count()

					Eventually(templateCount).Should(Equal(noOfTemplates), "The template header count should be equal to templates created")
					Eventually(tileCount).Should(Equal(noOfTemplates), "The number of template tiles rendered should be equal to number of templates created")
				})
			})
		})

		Context("When Capi Templates are available in the cluster", func() {
			It("Verify I should be able to select a template of my choice", func() {

				// test selection with 50 capiTemplates
				templateFiles = mccpTestRunner.CreateApplyCapitemplates(50, "capi-server-v1-capitemplate.yaml")

				pages.NavigateToPage(webDriver, "Templates")

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
			})

			Context("When only invalid Capi Template(s) are available in the cluster", func() {
				XIt("Verify UI shows message related to an invalid template(s)", func() {

					By("Apply/Insall invalid CAPITemplate", func() {
						templateFiles = mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-invalid-capitemplate.yaml")
					})

					pages.NavigateToPage(webDriver, "Templates")

					By("And User should see message informing user of the invalid template in the cluster", func() {
						// TODO
					})

				})
			})

			Context("When both valid and invalid Capi Templates are available in the cluster", func() {
				XIt("Verify UI shows message related to an invalid template(s) and renders the available valid template(s)", func() {

					noOfTemplates := 3
					By("Apply/Insall valid CAPITemplate", func() {
						templateFiles = mccpTestRunner.CreateApplyCapitemplates(noOfTemplates, "capi-server-v1-template-eks-fargate.yaml")
					})

					By("Apply/Insall invalid CAPITemplate", func() {
						templateFiles = mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-invalid-capitemplate.yaml")
					})

					pages.NavigateToPage(webDriver, "Templates")
					templatesPage := pages.GetTemplatesPage(webDriver)

					By("And wait for Templates page to be fully rendered", func() {
						Eventually(templatesPage.TemplateHeader).Should(BeVisible())

						count, _ := templatesPage.TemplateCount.Text()
						templateCount, _ := strconv.Atoi(count)
						tileCount, _ := templatesPage.TemplateTiles.Count()

						Eventually(templateCount).Should(Equal(noOfTemplates), "The template header count should be equal to templates created")
						Eventually(tileCount).Should(Equal(noOfTemplates), "The number of template tiles rendered should be equal to number of templates created")
					})

					By("And User should see message informing user of the invalid template in the cluster", func() {
						// TODO
					})
				})
			})
		})
	})
}
