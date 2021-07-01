package acceptance

import (
	"fmt"
	"sort"
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

					// Testing templates are ordered
					expected_list := make([]string, noOfTemplates)
					for i := 0; i < noOfTemplates; i++ {
						expected_list[i] = fmt.Sprintf("cluster-template-%d", i)
					}
					sort.Strings(expected_list)

					actual_list := templatesPage.GetTemplateTileList()
					for i := 0; i < noOfTemplates; i++ {
						Expect(actual_list[i]).Should(ContainSubstring(expected_list[i]))
					}
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

		Context("When Capi Template is available in the cluster", func() {
			It("Verify template parameters should be rendered dynamically and can be set for the selected template", func() {

				By("Apply/Insall CAPITemplate", func() {
					templateFiles = mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-eks-fargate.yaml")
				})

				pages.NavigateToPage(webDriver, "Templates")

				By("And User should choose a template", func() {
					templateTile := pages.GetTemplateTile(webDriver, "eks-fargate-template-0")
					Expect(templateTile.CreateTemplate.Click()).To(Succeed())
				})

				createPage := pages.GetCreateClusterPage(webDriver)
				By("And wait for Create cluster page to be fully rendered", func() {
					Eventually(createPage.CreateHeader).Should(MatchText(".*Create new cluster.*"))
					// Eventually(createPage.TemplateName).Should(MatchText(".*eks-fargate-template-0.*"))
				})

				clusterName := "my-eks-cluster"
				region := "east"
				sshKey := "abcdef1234567890"
				k8Version := "1.19"
				By("And set template parameter values", func() {
					templateParam := createPage.GetTemplateParameter("CLUSTER_NAME")
					Expect(templateParam.Label).Should(MatchText("CLUSTER_NAME.*"))
					Expect(templateParam.Feild.SendKeys(clusterName)).To(Succeed())

					templateParam = createPage.GetTemplateParameter("AWS_REGION")
					Expect(templateParam.Label).Should(MatchText("AWS_REGION.*"))
					Expect(templateParam.Feild.SendKeys(region)).To(Succeed())

					templateParam = createPage.GetTemplateParameter("AWS_SSH_KEY_NAME")
					Expect(templateParam.Label).Should(MatchText("AWS_SSH_KEY_NAME.*"))
					Expect(templateParam.Feild.SendKeys(sshKey)).To(Succeed())

					templateParam = createPage.GetTemplateParameter("KUBERNETES_VERSION")
					Expect(templateParam.Label).Should(MatchText("KUBERNETES_VERSION.*"))
					Expect(templateParam.Feild.SendKeys(k8Version)).To(Succeed())
				})

				By("Then I should preview the PR", func() {
					Expect(createPage.PreviewPR.Submit()).To(Succeed())
					preview := pages.GetPreview(webDriver)
					Eventually(preview.PreviewLabel).Should(BeFound())

					Expect(preview.PreviewText).Should(MatchText(fmt.Sprintf(`kind: Cluster[\s\w\d./:-]*name: %[1]v\s+spec:[\s\w\d./:-]*controlPlaneRef:[\s\w\d./:-]*name: %[1]v-control-plane\s+infrastructureRef:[\s\w\d./:-]*kind: AWSManagedCluster\s+name: %[1]v`, clusterName)))
					Expect(preview.PreviewText).Should((MatchText(fmt.Sprintf(`kind: AWSManagedCluster\s+metadata:\s+name: %[1]v`, clusterName))))
					Expect(preview.PreviewText).Should((MatchText(fmt.Sprintf(`kind: AWSManagedControlPlane\s+metadata:\s+name: %[1]v-control-plane\s+spec:\s+region: %[2]v\s+sshKeyName: %[3]v\s+version: "%[4]v"`, clusterName, region, sshKey, k8Version))))
					Expect(preview.PreviewText).Should((MatchText(fmt.Sprintf(`kind: AWSFargateProfile\s+metadata:\s+name: %[1]v-fargate-0`, clusterName))))
				})
			})
		})

		Context("When Capi Template is available in the cluster", func() {
			It("Verify pull request can be created for the selected capi template", func() {

				By("Apply/Insall CAPITemplate", func() {
					templateFiles = mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
				})

				pages.NavigateToPage(webDriver, "Templates")

				By("And User should choose a template", func() {
					templateTile := pages.GetTemplateTile(webDriver, "cluster-template-development-0")
					Expect(templateTile.CreateTemplate.Click()).To(Succeed())
				})

				createPage := pages.GetCreateClusterPage(webDriver)

				By("And set template parameter values", func() {
					clusterName := "my-development-cluster"
					templateParam := createPage.GetTemplateParameter("CLUSTER_NAME")
					Expect(templateParam.Feild.SendKeys(clusterName)).To(Succeed())

					namespace := "mccp-dev"
					templateParam = createPage.GetTemplateParameter("NAMESPACE")
					Expect(templateParam.Feild.SendKeys(namespace)).To(Succeed())

					k8Version := "1.19.8"
					templateParam = createPage.GetTemplateParameter("KUBERNETES_VERSION")
					Expect(templateParam.Feild.SendKeys(k8Version)).To(Succeed())
				})

				By("And press the Preview PR button", func() {
					Expect(createPage.PreviewPR.Submit()).To(Succeed())
				})

				By("And set GitOps values for pull request", func() {
					gitops := pages.GetGitOps(webDriver)
					Eventually(gitops.GitOpsLabel).Should(BeFound())

					gitops.ScrollIntoView(webDriver, gitops.GitOpsLabel)

					Expect(gitops.GitOpsFeilds[0].Label).Should(BeFound())
					Expect(gitops.GitOpsFeilds[1].Label).Should(BeFound())
					Expect(gitops.GitOpsFeilds[2].Label).Should(BeFound())
					Expect(gitops.CreatePR).Should(BeFound())
				})
			})
		})
	})
}
