package acceptance

import (
	"context"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	"github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

type TemplateField struct {
	Name   string
	Value  string
	Option string
}

func navigateToTemplatesGrid(webDriver *agouti.Page) {
	pages.NavigateToPage(webDriver, "Templates")
	pages.WaitForPageToLoad(webDriver)
	gomega.Expect(pages.GetTemplatesPage(webDriver).SelectView("grid").Click()).To(gomega.Succeed())
}

func setParameterValues(createPage *pages.CreateCluster, parameters []TemplateField) {
	for i := 0; i < len(parameters); i++ {
		if parameters[i].Option != "" {
			selectOption := func() bool {
				_ = createPage.GetTemplateParameter(webDriver, parameters[i].Name).ListBox.Click()
				time.Sleep(POLL_INTERVAL_100MILLISECONDS)
				visible, _ := pages.GetOption(webDriver, parameters[i].Option).Visible()
				return visible
			}
			gomega.Eventually(selectOption, ASSERTION_DEFAULT_TIME_OUT).Should(gomega.BeTrue(), fmt.Sprintf("Failed to select parameter option '%s'", parameters[i].Name))
			gomega.Expect(pages.GetOption(webDriver, parameters[i].Option).Click()).To(gomega.Succeed())
		} else {
			field := createPage.GetTemplateParameter(webDriver, parameters[i].Name).Field
			pages.ClearFieldValue(field)
			gomega.Expect(field.SendKeys(parameters[i].Value)).To(gomega.Succeed())
		}
	}
}

func selectCredentials(createPage *pages.CreateCluster, credentialName string, credentialCount int) {
	selectCredential := func() bool {
		gomega.Eventually(createPage.Credentials.Click).Should(gomega.Succeed())
		// Credentials are not filtered for selected template
		if cnt, _ := pages.GetCredentials(webDriver).Count(); cnt > 0 {
			gomega.Eventually(pages.GetCredentials(webDriver).Count).Should(gomega.Equal(credentialCount), fmt.Sprintf(`Credentials count in the cluster should be '%d' including 'None`, credentialCount))
			gomega.Expect(pages.GetCredential(webDriver, credentialName).Click()).To(gomega.Succeed())
		}

		credentialText, _ := createPage.Credentials.Text()
		return strings.Contains(credentialText, credentialName)
	}
	gomega.Eventually(selectCredential, ASSERTION_30SECONDS_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.BeTrue())
}

func DescribeTemplates(gitopsTestRunner GitopsTestRunner) {
	var _ = ginkgo.Describe("Multi-Cluster Control Plane Templates", func() {
		templateFiles := []string{}
		clusterPath := "./clusters/management/clusters"

		ginkgo.BeforeEach(func() {
			gomega.Expect(webDriver.Navigate(test_ui_url)).To(gomega.Succeed())

			if !pages.ElementExist(pages.Navbar(webDriver).Title, 3) {
				loginUser()
			}
		})

		ginkgo.AfterEach(func() {
			gitopsTestRunner.DeleteApplyCapiTemplates(templateFiles)
			// Reset/empty the templateFiles list
			templateFiles = []string{}
		})

		ginkgo.Context("[UI] When no Capi Templates are available in the cluster", func() {
			ginkgo.It("Verify template page renders no capiTemplate", ginkgo.Label("integration"), func() {
				ginkgo.By("And wait for  good looking response from /v1/templates", func() {
					gomega.Expect(waitForGitopsResources(context.Background(), "templates", POLL_INTERVAL_15SECONDS)).To(gomega.Succeed(), "Failed to get a successful response from /v1/templates")
				})

				navigateToTemplatesGrid(webDriver)
				templatesPage := pages.GetTemplatesPage(webDriver)

				ginkgo.By("And wait for Templates page to be rendered", func() {
					gomega.Eventually(templatesPage.TemplateHeader).Should(matchers.BeVisible())

					tileCount, _ := templatesPage.TemplateTiles.Count()
					gomega.Expect(tileCount).To(gomega.Equal(0), "There should not be any template tile rendered")

				})
			})
		})

		ginkgo.Context("[UI] When Capi Templates are available in the cluster", func() {
			ginkgo.It("Verify template(s) are rendered from the template library.", func() {
				awsTemplateCount := 2
				eksFargateTemplateCount := 2
				azureTemplateCount := 3
				capdTemplateCount := 5
				totalTemplateCount := awsTemplateCount + eksFargateTemplateCount + azureTemplateCount + capdTemplateCount

				ordered_template_list := func() []string {
					expected_list := make([]string, totalTemplateCount)
					for i := 0; i < 2; i++ {
						expected_list[i] = fmt.Sprintf("aws-cluster-template-%d", i)
					}
					for i := 0; i < 3; i++ {
						expected_list[i] = fmt.Sprintf("azure-capi-quickstart-template-%d", i)
					}
					for i := 0; i < 5; i++ {
						expected_list[i] = fmt.Sprintf("cluster-template-development-%d", i)
					}
					for i := 0; i < 2; i++ {
						expected_list[i] = fmt.Sprintf("eks-fargate-template-%d", i)
					}
					sort.Strings(expected_list)
					return expected_list
				}()

				ginkgo.By("Apply/Install CAPITemplate", func() {
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(5, "capi-template-capd.yaml")...)
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(3, "capi-server-v1-template-azure.yaml")...)
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(2, "capi-server-v1-template-aws.yaml")...)
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(2, "capi-server-v1-template-eks-fargate.yaml")...)
				})

				templatesPage := pages.GetTemplatesPage(webDriver)
				navigateToTemplatesGrid(webDriver)

				ginkgo.By("And wait for Templates page to be fully rendered", func() {
					gomega.Eventually(templatesPage.TemplateHeader).Should(matchers.BeVisible())
					tileCount, _ := templatesPage.TemplateTiles.Count()
					gomega.Eventually(tileCount).Should(gomega.Equal(totalTemplateCount), "The number of template tiles rendered should be equal to number of templates created")
				})

				ginkgo.By("And I should change the templates view to 'table'", func() {
					gomega.Expect(templatesPage.SelectView("table").Click()).To(gomega.Succeed())
					gomega.Eventually(templatesPage.CountTemplateRows()).Should(gomega.Equal(totalTemplateCount), "The number of rows rendered should be equal to number of templates created")

				})

				ginkgo.By("And templates are ordered - table view", func() {
					actual_list := templatesPage.GetTemplateTableList()
					for i := 0; i < totalTemplateCount; i++ {
						gomega.Expect(actual_list[i]).Should(gomega.ContainSubstring(ordered_template_list[i]))
					}
				})

				ginkgo.By("And templates can be filtered by provider - table view", func() {
					filterID := "provider: aws"
					searchPage := pages.GetSearchPage(webDriver)
					searchPage.SelectFilter("provider", filterID)
					gomega.Eventually(templatesPage.CountTemplateRows()).Should(gomega.Equal(4), "The number of selected template tiles rendered should be equal to number of aws templates created")
				})

				ginkgo.By("And I should change the templates view to 'grid'", func() {
					gomega.Expect(templatesPage.SelectView("grid").Click()).To(gomega.Succeed())
					tileCount, _ := templatesPage.TemplateTiles.Count()
					gomega.Eventually(tileCount).Should(gomega.Equal(totalTemplateCount), "The number of template tiles rendered should be equal to number of templates created")
				})

				ginkgo.By("And templates are ordered - grid view", func() {
					actual_list := templatesPage.GetTemplateTileList()
					for i := 0; i < totalTemplateCount; i++ {
						gomega.Expect(actual_list[i]).Should(gomega.ContainSubstring(ordered_template_list[i]))
					}
				})

				ginkgo.By("And templates can be filtered by provider - grid view", func() {
					gomega.Expect(templatesPage.SelectView("grid").Click()).To(gomega.Succeed())
					// Select cluster provider by selecting from the popup list
					gomega.Expect(templatesPage.TemplateProvider.Click()).To(gomega.Succeed())
					gomega.Expect(templatesPage.SelectProvider("aws").Click()).To(gomega.Succeed())

					tileCount, _ := templatesPage.TemplateTiles.Count()
					gomega.Eventually(tileCount).Should(gomega.Equal(awsTemplateCount+eksFargateTemplateCount), "The number of aws provider template tiles rendered should be equal to number of aws templates created")

					// Select cluster provider by typing the provider name
					gomega.Expect(templatesPage.TemplateProvider.Click()).To(gomega.Succeed())
					gomega.Expect(templatesPage.TemplateProvider.SendKeys("\uE003")).To(gomega.Succeed()) // sending back space key
					gomega.Expect(templatesPage.TemplateProvider.SendKeys("azure")).To(gomega.Succeed())
					gomega.Expect(templatesPage.TemplateProviderPopup.At(0).Click()).To(gomega.Succeed())

					tileCount, _ = templatesPage.TemplateTiles.Count()
					gomega.Eventually(tileCount).Should(gomega.Equal(azureTemplateCount), "The number of azure provider template tiles rendered should be equal to number of azure templates created")
				})
			})
		})

		ginkgo.Context("[UI] When Capi Templates are available in the cluster", func() {
			ginkgo.It("Verify I should be able to select a template of my choice", func() {

				// test selection with 50 capiTemplates
				templateFiles = gitopsTestRunner.CreateApplyCapitemplates(50, "capi-server-v1-capitemplate.yaml")

				navigateToTemplatesGrid(webDriver)
				templatesPage := pages.GetTemplatesPage(webDriver)

				ginkgo.By("And I should choose a template - grid view", func() {
					templateTile := pages.GetTemplateTile(webDriver, "cluster-template-9")

					gomega.Eventually(templateTile.Description).Should(matchers.MatchText("This is test template 9"))
					gomega.Expect(templateTile.CreateTemplate).Should(matchers.BeFound())
					gomega.Expect(templateTile.CreateTemplate.Click()).To(gomega.Succeed())
				})

				ginkgo.By("And wait for Create cluster page to be fully rendered - grid view", func() {
					createPage := pages.GetCreateClusterPage(webDriver)
					gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(".*Create new resource.*"))
				})

				ginkgo.By("And I should wait for the table to be fully loaded - table view by default", func() {
					pages.NavigateToPage(webDriver, "Templates")
					pages.WaitForPageToLoad(webDriver)
				})

				ginkgo.By("And I should choose a template from the default table view", func() {
					templateRow := templatesPage.GetTemplateRow(webDriver, "cluster-template-10")
					gomega.Eventually(templateRow.Provider).Should(matchers.MatchText(""))
					gomega.Eventually(templateRow.Description).Should(matchers.MatchText("This is test template 10"))
					gomega.Expect(templateRow.CreateTemplate).Should(matchers.BeFound())
					gomega.Expect(templateRow.CreateTemplate.Click()).To(gomega.Succeed())
				})

				ginkgo.By("And wait for Create cluster page to be fully rendered - table view", func() {
					createPage := pages.GetCreateClusterPage(webDriver)
					gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(".*Create new resource.*"))
				})
			})
		})

		ginkgo.Context("[UI] When only invalid Capi Template(s) are available in the cluster", func() {
			ginkgo.It("Verify UI shows message related to an invalid template(s)", func() {

				ginkgo.By("Apply/Install invalid CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-invalid-capitemplate.yaml")
				})

				navigateToTemplatesGrid(webDriver)
				templatesPage := pages.GetTemplatesPage(webDriver)

				ginkgo.By("And User should see message informing user of the invalid template in the cluster - grid view", func() {
					templateTile := pages.GetTemplateTile(webDriver, "cluster-invalid-template-0")
					gomega.Eventually(templateTile.ErrorHeader).Should(matchers.BeFound())
					gomega.Expect(templateTile.ErrorDescription).Should(matchers.BeFound())
					gomega.Expect(templateTile.CreateTemplate).ShouldNot(matchers.BeEnabled())
				})

				ginkgo.By("And I should change the templates view to 'table'", func() {
					gomega.Expect(templatesPage.SelectView("table").Click()).To(gomega.Succeed())
				})

				ginkgo.By("And User should see message informing user of the invalid template in the cluster - table view", func() {
					templateRow := templatesPage.GetTemplateRow(webDriver, "cluster-invalid-template-0")
					gomega.Eventually(templateRow.Provider).Should(matchers.MatchText(""))
					gomega.Eventually(templateRow.Description).Should(matchers.MatchText("Couldn't load template body"))
					gomega.Expect(templateRow.CreateTemplate).ShouldNot(matchers.BeEnabled())
				})
			})
		})

		ginkgo.Context("[UI] When both valid and invalid Capi Templates are available in the cluster", func() {
			ginkgo.It("Verify UI shows message related to an invalid template(s) and renders the available valid template(s)", func() {

				noOfValidTemplates := 3
				ginkgo.By("Apply/Install valid CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(noOfValidTemplates, "capi-server-v1-template-eks-fargate.yaml")
				})

				noOfInvalidTemplates := 1
				ginkgo.By("Apply/Install invalid CAPITemplate", func() {
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(noOfInvalidTemplates, "capi-server-v1-invalid-capitemplate.yaml")...)
				})

				navigateToTemplatesGrid(webDriver)
				templatesPage := pages.GetTemplatesPage(webDriver)
				ginkgo.By("And wait for Templates page to be fully rendered", func() {
					gomega.Expect(templatesPage.SelectView("grid").Click()).To(gomega.Succeed())
					gomega.Eventually(templatesPage.TemplateHeader).Should(matchers.BeVisible())
					tileCount, _ := templatesPage.TemplateTiles.Count()
					gomega.Eventually(tileCount).Should(gomega.Equal(noOfValidTemplates+noOfInvalidTemplates), "The number of template tiles rendered should be equal to number of templates created")
				})

				ginkgo.By("And User should see message informing user of the invalid template in the cluster", func() {
					templateTile := pages.GetTemplateTile(webDriver, "cluster-invalid-template-0")
					gomega.Eventually(templateTile.ErrorHeader).Should(matchers.BeFound())
					gomega.Expect(templateTile.ErrorDescription).Should(matchers.BeFound())
					gomega.Expect(templateTile.CreateTemplate).ShouldNot(matchers.BeEnabled())
				})
			})
		})

		ginkgo.Context("[UI] When Capi Template is available in the cluster", func() {
			ginkgo.It("Verify template parameters should be rendered dynamically and can be set for the selected template", ginkgo.Label("integration"), func() {

				ginkgo.By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-eks-fargate.yaml")
				})

				navigateToTemplatesGrid(webDriver)
				templatesPage := pages.GetTemplatesPage(webDriver)

				ginkgo.By("And I should change the templates view to 'table'", func() {
					gomega.Expect(templatesPage.SelectView("table").Click()).To(gomega.Succeed())
				})

				ginkgo.By("And I should choose a template - table view", func() {
					templateRow := templatesPage.GetTemplateRow(webDriver, "eks-fargate-template-0")
					gomega.Expect(templateRow.CreateTemplate.Click()).To(gomega.Succeed())
				})

				createPage := pages.GetCreateClusterPage(webDriver)
				ginkgo.By("And wait for Create cluster page to be fully rendered", func() {
					pages.WaitForPageToLoad(webDriver)
					gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(".*Create new resource.*"))
				})

				clusterName := "my-eks-cluster"
				region := "east"

				var parameters = []TemplateField{
					{
						Name:   "CLUSTER_NAME",
						Value:  clusterName,
						Option: "",
					},
				}

				setParameterValues(createPage, parameters)

				ginkgo.By("Then missing required parameters should get focus when previewing PR", func() {
					gomega.Eventually(createPage.PreviewPR.Click).Should(gomega.Succeed())
					gomega.Eventually(createPage.GetTemplateParameter(webDriver, "AWS_REGION").Focused).Should(matchers.BeFound(), "Missing required parameter 'AWS_REGION' failed to get focus")
				})

				parameters = []TemplateField{
					{
						Name:   "AWS_REGION",
						Value:  region,
						Option: "",
					},
				}

				setParameterValues(createPage, parameters)

				ginkgo.By("Then I should preview the PR", func() {
					preview := pages.GetPreview(webDriver)
					gomega.Eventually(func(g gomega.Gomega) {
						g.Expect(createPage.PreviewPR.Click()).Should(gomega.Succeed())
						g.Expect(preview.Title.Text()).Should(gomega.MatchRegexp("PR Preview"))

					}, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Succeed(), "Failed to get PR preview")

					gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`kind: Cluster[\s\w\d./:-]*name: %[1]v\s+namespace: default\s+spec:[\s\w\d./:-]*controlPlaneRef:[\s\w\d./:-]*name: %[1]v-control-plane\s+infrastructureRef:[\s\w\d./:-]*kind: AWSManagedCluster\s+name: %[1]v`, clusterName)))
					gomega.Eventually(preview.Text).Should((matchers.MatchText(fmt.Sprintf(`kind: AWSManagedCluster\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v`, clusterName))))
					gomega.Eventually(preview.Text).Should((matchers.MatchText(fmt.Sprintf(`kind: AWSManagedControlPlane\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v-control-plane\s+namespace: default\s+spec:\s+region: %[2]v\s+sshKeyName: null\s+version: null`, clusterName, region))))
					gomega.Eventually(preview.Text).Should((matchers.MatchText(fmt.Sprintf(`kind: AWSFargateProfile\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v-fargate-0`, clusterName))))

					gomega.Eventually(preview.Close.Click).Should(gomega.Succeed())
				})
			})
		})

		ginkgo.Context("[UI] When Capi Template is available in the cluster", func() {
			var downloadedResourcesPath string

			ginkgo.JustBeforeEach(func() {
				downloadedResourcesPath = path.Join(os.Getenv("HOME"), "Downloads", "resources.zip")
				_ = deleteFile([]string{downloadedResourcesPath})
			})

			ginkgo.JustAfterEach(func() {
				// Force clean the repository directory for subsequent tests
				cleanGitRepository(clusterPath)
				_ = deleteFile([]string{downloadedResourcesPath})
			})

			ginkgo.It("Verify pull request can be created for capi template to the management cluster", ginkgo.Label("integration", "git", "browser-logs"), func() {
				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)

				ginkgo.By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-template-capd.yaml")
				})

				navigateToTemplatesGrid(webDriver)

				templateName := "cluster-template-development-0"
				ginkgo.By("And User should choose a template", func() {
					templateTile := pages.GetTemplateTile(webDriver, templateName)
					gomega.Expect(templateTile.CreateTemplate.Click()).To(gomega.Succeed())
				})

				createPage := pages.GetCreateClusterPage(webDriver)
				ginkgo.By("And wait for Create cluster page to be fully rendered", func() {
					pages.WaitForPageToLoad(webDriver)
					gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(".*Create new resource.*"))
				})

				// Parameter values
				leafCluster := ClusterConfig{
					Type:      "capi",
					Name:      "quick-capd-cluster",
					Namespace: "quick-capi",
				}
				k8Version := "1.22.0"
				controlPlaneMachineCount := "3"
				workerMachineCount := "3"

				var parameters = []TemplateField{
					{
						Name:   "CLUSTER_NAME",
						Value:  leafCluster.Name,
						Option: "",
					},
					{
						Name:   "NAMESPACE",
						Value:  leafCluster.Namespace,
						Option: "",
					},
					{
						Name:   "CONTROL_PLANE_MACHINE_COUNT",
						Value:  "",
						Option: controlPlaneMachineCount,
					},
					{
						Name:   "KUBERNETES_VERSION",
						Value:  "",
						Option: k8Version,
					},
					{
						Name:   "WORKER_MACHINE_COUNT",
						Value:  workerMachineCount,
						Option: "",
					},
					{
						Name:   "INSTALL_CRDS",
						Value:  "",
						Option: "true",
					},
				}

				setParameterValues(createPage, parameters)

				sourceHRUrl := "https://raw.githubusercontent.com/weaveworks/profiles-catalog/gh-pages"
				certManager := Application{
					DefaultApp:      true,
					Type:            "helm_release",
					Chart:           "weaveworks-charts",
					Name:            "cert-manager",
					Namespace:       GITOPS_DEFAULT_NAMESPACE,
					TargetNamespace: "cert-manager",
					Version:         "0.0.8",
					Values:          `installCRDs: \${INSTALL_CRDS}`,
					Layer:           "layer-0",
				}
				profile := createPage.GetProfileInList(certManager.Name)
				AddHelmReleaseApp(profile, certManager)

				podinfo := Application{
					Name:            "podinfo",
					Namespace:       GITOPS_DEFAULT_NAMESPACE,
					TargetNamespace: "test-system",
					Source:          "flux-system",
					Path:            "apps/podinfo",
				}
				gomega.Expect(createPage.AddApplication.Click()).Should(gomega.Succeed(), "Failed to click 'Add application' button")
				application := pages.GetAddApplication(webDriver, 1)
				AddKustomizationApp(application, podinfo)

				postgres := Application{
					Name:            "postgres",
					Namespace:       GITOPS_DEFAULT_NAMESPACE,
					TargetNamespace: "test-system",
					Path:            "apps/postgres",
				}
				gomega.Expect(createPage.AddApplication.Click()).Should(gomega.Succeed(), "Failed to click 'Add application' button")
				application = pages.GetAddApplication(webDriver, 2)
				AddKustomizationApp(application, postgres)
				gomega.Expect(application.RemoveApplication.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to remove application no. %d", 2))

				pages.ScrollWindow(webDriver, 0, 500)
				preview := pages.GetPreview(webDriver)
				ginkgo.By("Then I should preview the PR", func() {
					gomega.Eventually(func(g gomega.Gomega) {
						g.Expect(createPage.PreviewPR.Click()).Should(gomega.Succeed())
						g.Expect(preview.Title.Text()).Should(gomega.MatchRegexp("PR Preview"))

					}, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Succeed(), "Failed to get PR preview")
				})

				ginkgo.By("Then verify preview tab lists", func() {
					// Verify cluster definition preview
					gomega.Eventually(preview.GetPreviewTab("Cluster Definition").Click).Should(gomega.Succeed(), "Failed to switch to 'CLUSTER DEFINATION' preview tab")
					gomega.Eventually(preview.Text).Should(matchers.MatchText(`kind: Cluster[\s\w\d./:-]*metadata:[\s\w\d./:-]*labels:[\s\w\d./:-]*cni: calico`))
					gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`kind: GitopsCluster[\s\w\d./:-]*metadata:[\s\w\d./:-]*labels:[\s\w\d./:-]*templates.weave.works/template-name: %s`, templateName)))
					gomega.Eventually(preview.Text).Should(matchers.MatchText(`kind: GitopsCluster[\s\w\d./:-]*metadata:[\s\w\d./:-]*labels:[\s\w\d./:-]*templates.weave.works/template-namespace: default`))
					gomega.Eventually(preview.Text).Should(matchers.MatchText(`kind: GitopsCluster[\s\w\d./:-]*metadata:[\s\w\d./:-]*labels:[\s\w\d./:-]*weave.works/flux: bootstrap`))
					gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`kind: GitopsCluster[\s\w\d./:-]*metadata:[\s\w\d./:-]*name: %s[\s\w\d./:-]*namespace: %s[\s\w\d./:-]*capiClusterRef`, leafCluster.Name, leafCluster.Namespace)))

					// Verify profiles preview
					gomega.Eventually(preview.GetPreviewTab("Profiles").Click).Should(gomega.Succeed(), "Failed to switch to 'PROFILES' preview tab")
					gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`kind: HelmRepository[\s\w\d./:-]*name: %s[\s\w\d./:-]*namespace: %s[\s\w\d./:-]*url: %s`, certManager.Chart, GITOPS_DEFAULT_NAMESPACE, sourceHRUrl)))
					gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`kind: HelmRelease[\s\w\d./:-]*name: %s[\s\w\d./:-]*namespace: %s[\s\w\d./:-]*spec`, certManager.Name, certManager.Namespace)))
					gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`chart: %s[\s\w\d./:-]*sourceRef:[\s\w\d./:-]*name: %s[\s\w\d./:-]*version: %s[\s\w\d./:-]*targetNamespace: %s[\s\w\d./:-]*installCRDs: true`, certManager.Name, certManager.Chart, certManager.Version, certManager.TargetNamespace)))

					// Verify kustomizations preview
					gomega.Eventually(preview.GetPreviewTab("Kustomizations").Click).Should(gomega.Succeed(), "Failed to switch to 'KUSTOMIZATION' preview tab")
					gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`kind: Namespace[\s\w\d./:-]*name: %s`, podinfo.TargetNamespace)))
					gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`kind: Kustomization[\s\w\d./:-]*name: %s[\s\w\d./:-]*namespace: %s[\s\w\d./:-]*spec`, podinfo.Name, podinfo.Namespace)))
					gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`sourceRef:[\s\w\d./:-]*kind: GitRepository[\s\w\d./:-]*name: %s[\s\w\d./:-]*namespace: %s[\s\w\d./:-]*targetNamespace: %s`, podinfo.Source, podinfo.Namespace, podinfo.TargetNamespace)))
				})

				ginkgo.By("And verify downloaded preview resources", func() {
					// verify download prview resources
					gomega.Eventually(func(g gomega.Gomega) {
						g.Expect(preview.Download.Click()).Should(gomega.Succeed())
						_, err := os.Stat(downloadedResourcesPath)
						g.Expect(err).Should(gomega.Succeed())
					}, ASSERTION_1MINUTE_TIME_OUT).ShouldNot(gomega.HaveOccurred(), "Failed to click 'Download' preview resources")
					gomega.Eventually(preview.Close.Click).Should(gomega.Succeed())
					fileList, _ := getArchiveFileList(path.Join(os.Getenv("HOME"), "Downloads", "resources.zip"))

					previewResources := []string{
						"cluster_definition.yaml",
						path.Join("clusters", leafCluster.Namespace, leafCluster.Name, "clusters-bases-kustomization.yaml"),
						path.Join("clusters", leafCluster.Namespace, leafCluster.Name, "profiles.yaml"),
						path.Join("clusters", leafCluster.Namespace, leafCluster.Name, strings.Join([]string{podinfo.TargetNamespace, "namespace.yaml"}, "-")),
						path.Join("clusters", leafCluster.Namespace, leafCluster.Name, strings.Join([]string{podinfo.Name, podinfo.Namespace, "kustomization.yaml"}, "-")),
					}
					gomega.Expect(len(fileList)).Should(gomega.Equal(len(previewResources)), "Failed to verify expected number of downloaded preview resources")
					gomega.Expect(fileList).Should(gomega.ContainElements(previewResources), "Failed to verify downloaded preview resources files")
				})

				pullRequest := PullRequest{
					Branch:  "ui-feature-capd",
					Title:   "My first pull request",
					Message: "First capd capi template",
				}

				prUrl := createGitopsPR(pullRequest)
				var createPRUrl string

				ginkgo.By("And I should veriyfy the pull request in the cluster config repository", func() {
					createPRUrl = verifyPRCreated(gitProviderEnv, repoAbsolutePath)
					gomega.Expect(createPRUrl).Should(gomega.Equal(prUrl))
				})

				ginkgo.By("And the manifests are present in the cluster config repository", func() {
					mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
					pullGitRepo(repoAbsolutePath)
					_, err := os.Stat(path.Join(repoAbsolutePath, clusterPath, leafCluster.Namespace, leafCluster.Name+".yaml"))
					gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Cluster config can not be found.")

					_, err = os.Stat(path.Join(repoAbsolutePath, "clusters", leafCluster.Namespace, leafCluster.Name, "profiles.yaml"))
					gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "profiles.yaml can not be found.")

					_, err = os.Stat(path.Join(repoAbsolutePath, "clusters", leafCluster.Namespace, leafCluster.Name, strings.Join([]string{podinfo.TargetNamespace, "namespace.yaml"}, "-")))
					gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "target namespace.yaml can not be found.")

					_, err = os.Stat(path.Join(repoAbsolutePath, "clusters", leafCluster.Namespace, leafCluster.Name, strings.Join([]string{podinfo.Name, podinfo.Namespace, "kustomization.yaml"}, "-")))
					gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "podinfo kustomization.yaml are found when expected to be deleted.")

					_, err = os.Stat(path.Join(repoAbsolutePath, "clusters", leafCluster.Namespace, leafCluster.Name, strings.Join([]string{postgres.Name, postgres.Namespace, "kustomization.yaml"}, "-")))
					gomega.Expect(err).Should(gomega.MatchError(os.ErrNotExist), "postgress kustomization is found when expected to be deleted.")
				})
			})

			ginkgo.It("Verify pull request can not be created by using exiting repository branch", ginkgo.Label("integration", "git", "browser-logs"), func() {
				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)

				branchName := "ui-test-branch"
				ginkgo.By("And create new git repository branch", func() {
					_ = createGitRepoBranch(repoAbsolutePath, branchName)
				})

				ginkgo.By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-template-capd.yaml")
				})

				navigateToTemplatesGrid(webDriver)

				ginkgo.By("And User should choose a template", func() {
					templateTile := pages.GetTemplateTile(webDriver, "cluster-template-development-0")
					gomega.Expect(templateTile.CreateTemplate.Click()).To(gomega.Succeed())
				})

				createPage := pages.GetCreateClusterPage(webDriver)
				ginkgo.By("And wait for Create cluster page to be fully rendered", func() {
					pages.WaitForPageToLoad(webDriver)
					gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(".*Create new resource.*"))
				})

				// Parameter values
				clusterName := "quick-capd-cluster2"
				namespace := "quick-capi"
				k8Version := "1.22.0"
				controlPlaneMachineCount := "2"
				workerMachineCount := "2"

				var parameters = []TemplateField{
					{
						Name:   "CLUSTER_NAME",
						Value:  clusterName,
						Option: "",
					},
					{
						Name:   "NAMESPACE",
						Value:  namespace,
						Option: "",
					},
					{
						Name:   "CONTROL_PLANE_MACHINE_COUNT",
						Value:  "",
						Option: controlPlaneMachineCount,
					},
					{
						Name:   "KUBERNETES_VERSION",
						Value:  "",
						Option: k8Version,
					},
					{
						Name:   "WORKER_MACHINE_COUNT",
						Value:  workerMachineCount,
						Option: "",
					},
					{
						Name:   "INSTALL_CRDS",
						Value:  "",
						Option: "true",
					},
				}

				setParameterValues(createPage, parameters)

				//Pull request values
				prTitle := "My first pull request"
				prCommit := "First capd capi template"

				gitops := pages.GetGitOps(webDriver)
				ginkgo.By("And set GitOps values for pull request", func() {
					gomega.Eventually(gitops.GitOpsLabel).Should(matchers.BeFound())

					pages.ScrollWindow(webDriver, 0, 4000)

					pages.ClearFieldValue(gitops.BranchName)
					gomega.Expect(gitops.BranchName.SendKeys(branchName)).To(gomega.Succeed())
					pages.ClearFieldValue(gitops.PullRequestTile)
					gomega.Expect(gitops.PullRequestTile.SendKeys(prTitle)).To(gomega.Succeed())
					pages.ClearFieldValue(gitops.CommitMessage)
					gomega.Expect(gitops.CommitMessage.SendKeys(prCommit)).To(gomega.Succeed())

					AuthenticateWithGitProvider(webDriver, gitProviderEnv.Type, gitProviderEnv.Hostname)
					gomega.Eventually(gitops.GitCredentials).Should(matchers.BeVisible())

					if pages.ElementExist(gitops.ErrorBar) {
						gomega.Expect(gitops.ErrorBar.Click()).To(gomega.Succeed())
					}

					gomega.Expect(gitops.CreatePR.Click()).To(gomega.Succeed())
				})

				ginkgo.By("Then I should not see pull request to be created", func() {
					gomega.Eventually(gitops.ErrorBar).Should(matchers.MatchText(fmt.Sprintf(`unable to create pull request.+unable to create new branch "%s"`, branchName)))
				})
			})
		})

		ginkgo.Context("[UI] When no infrastructure provider credentials are available in the management cluster", func() {
			ginkgo.It("Verify no credentials exists in management cluster", ginkgo.Label("integration", "git"), func() {
				ginkgo.By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-template-capd.yaml")
				})

				navigateToTemplatesGrid(webDriver)

				ginkgo.By("And User should choose a template", func() {
					templateTile := pages.GetTemplateTile(webDriver, "cluster-template-development-0")
					gomega.Expect(templateTile.CreateTemplate.Click()).To(gomega.Succeed())
				})

				createPage := pages.GetCreateClusterPage(webDriver)
				ginkgo.By("And wait for Create cluster page to be fully rendered", func() {
					pages.WaitForPageToLoad(webDriver)
					gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(".*Create new resource.*"))
				})

				ginkgo.By("Then no infrastructure provider identity can be selected", func() {
					selectCredentials(createPage, "None", 1)
				})
			})
		})

		ginkgo.Context("[UI] When infrastructure provider credentials are available in the management cluster", func() {

			ginkgo.JustAfterEach(func() {
				gitopsTestRunner.DeleteIPCredentials("AWS")
				gitopsTestRunner.DeleteIPCredentials("AZURE")
			})

			ginkgo.It("Verify matching selected credential can be used for cluster creation", ginkgo.Label("integration", "git"), func() {
				ginkgo.By("Apply/Install CAPITemplates", func() {
					eksTemplateFile := gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-aws.yaml")
					azureTemplateFiles := gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-azure.yaml")
					templateFiles = append(azureTemplateFiles, eksTemplateFile...)
				})

				ginkgo.By("And create infrastructure provider credentials)", func() {
					gitopsTestRunner.CreateIPCredentials("AWS")
					gitopsTestRunner.CreateIPCredentials("AZURE")
				})

				navigateToTemplatesGrid(webDriver)

				ginkgo.By("And User should choose a template", func() {
					templateTile := pages.GetTemplateTile(webDriver, "aws-cluster-template-0")
					gomega.Expect(templateTile.CreateTemplate.Click()).To(gomega.Succeed())
				})

				createPage := pages.GetCreateClusterPage(webDriver)
				ginkgo.By("And wait for Create cluster page to be fully rendered", func() {
					pages.WaitForPageToLoad(webDriver)
					gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(".*Create new resource.*"))
				})

				ginkgo.By("Then AWS test-role-identity credential can be selected", func() {
					selectCredentials(createPage, "test-role-identity", 4)
				})

				// AWS template parameter values
				awsClusterName := "my-aws-cluster"
				awsRegion := "eu-west-3"
				awsK8version := "1.19.8"
				awsSshKeyName := "my-aws-ssh-key"
				awsNamespace := "default"
				awsControlMAchineType := "t4g.large"
				awsNodeMAchineType := "t3.micro"

				var parameters = []TemplateField{
					{
						Name:   "CLUSTER_NAME",
						Value:  awsClusterName,
						Option: "",
					},
					{
						Name:   "AWS_REGION",
						Value:  awsRegion,
						Option: "",
					},
					{
						Name:   "AWS_SSH_KEY_NAME",
						Value:  awsSshKeyName,
						Option: "",
					},
					{
						Name:   "NAMESPACE",
						Value:  awsNamespace,
						Option: "",
					},
					{
						Name:   "CONTROL_PLANE_MACHINE_COUNT",
						Value:  "2",
						Option: "",
					},
					{
						Name:   "KUBERNETES_VERSION",
						Value:  awsK8version,
						Option: "",
					},
					{
						Name:   "AWS_CONTROL_PLANE_MACHINE_TYPE",
						Value:  awsControlMAchineType,
						Option: "",
					},
					{
						Name:   "WORKER_MACHINE_COUNT",
						Value:  "3",
						Option: "",
					},
					{
						Name:   "AWS_NODE_MACHINE_TYPE",
						Value:  awsNodeMAchineType,
						Option: "",
					},
				}

				setParameterValues(createPage, parameters)

				ginkgo.By("Then I should see PR preview containing identity reference added in the template", func() {
					preview := pages.GetPreview(webDriver)
					gomega.Eventually(func(g gomega.Gomega) {
						g.Expect(createPage.PreviewPR.Click()).Should(gomega.Succeed())
						g.Expect(preview.Title.Text()).Should(gomega.MatchRegexp("PR Preview"))

					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Succeed(), "Failed to get PR preview")

					gomega.Eventually(preview.Title).Should(matchers.MatchText("PR Preview"))

					gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`kind: AWSCluster\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %s[\s\w\d-.:/]+identityRef:[\s\w\d-.:/]+kind: AWSClusterRoleIdentity\s+name: test-role-identity`, awsClusterName)))
					gomega.Eventually(preview.Close.Click).Should(gomega.Succeed())
				})

			})
		})

		ginkgo.Context("[UI] When infrastructure provider credentials are available in the management cluster", func() {

			ginkgo.JustAfterEach(func() {
				gitopsTestRunner.DeleteIPCredentials("AWS")
			})

			ginkgo.It("Verify user can not use wrong credentials for infrastructure provider", ginkgo.Label("integration", "git"), func() {
				ginkgo.By("Apply/Install CAPITemplates", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-azure.yaml")
				})

				ginkgo.By("And create infrastructure provider credentials)", func() {
					gitopsTestRunner.CreateIPCredentials("AWS")
				})

				navigateToTemplatesGrid(webDriver)

				ginkgo.By("And User should choose a template", func() {
					templateTile := pages.GetTemplateTile(webDriver, "azure-capi-quickstart-template-0")
					gomega.Expect(templateTile.CreateTemplate.Click()).To(gomega.Succeed())
				})

				createPage := pages.GetCreateClusterPage(webDriver)
				ginkgo.By("And wait for Create cluster page to be fully rendered", func() {
					pages.WaitForPageToLoad(webDriver)
					gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(".*Create new resource.*"))
				})

				ginkgo.By("Then AWS aws-test-identity credential can be selected", func() {
					selectCredentials(createPage, "aws-test-identity", 3)
				})

				// Azure template parameter values
				azureClusterName := "my-azure-cluster"
				azureK8version := "1.21.2"
				azureNamespace := "default"
				azureControlMAchineType := "Standard_D2s_v3"
				azureNodeMAchineType := "Standard_D4_v4"

				var parameters = []TemplateField{
					{
						Name:   "CLUSTER_NAME",
						Value:  azureClusterName,
						Option: "",
					},
					{
						Name:   "NAMESPACE",
						Value:  azureNamespace,
						Option: "",
					},
					{
						Name:   "CONTROL_PLANE_MACHINE_COUNT",
						Value:  "2",
						Option: "",
					},
					{
						Name:   "KUBERNETES_VERSION",
						Value:  "",
						Option: azureK8version,
					},
					{
						Name:   "AZURE_CONTROL_PLANE_MACHINE_TYPE",
						Value:  "",
						Option: azureControlMAchineType,
					},
					{
						Name:   "WORKER_MACHINE_COUNT",
						Value:  "3",
						Option: "",
					},
					{
						Name:   "AZURE_NODE_MACHINE_TYPE",
						Value:  "",
						Option: azureNodeMAchineType,
					},
				}

				setParameterValues(createPage, parameters)

				ginkgo.By("Then I should see PR preview without identity reference added to the template", func() {
					preview := pages.GetPreview(webDriver)
					gomega.Eventually(func(g gomega.Gomega) {
						g.Expect(createPage.PreviewPR.Click()).Should(gomega.Succeed())
						g.Expect(preview.Title.Text()).Should(gomega.MatchRegexp("PR Preview"))

					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Succeed(), "Failed to get PR preview")

					gomega.Eventually(preview.Title).Should(matchers.MatchText("PR Preview"))

					gomega.Eventually(preview.Text).ShouldNot(matchers.MatchText(`kind: AWSCluster[\s\w\d-.:/]+identityRef:`), "Identity reference should not be found in preview pull request AzureCluster object")
					gomega.Eventually(preview.Close.Click).Should(gomega.Succeed())
				})

			})
		})

		ginkgo.Context("[UI] When leaf cluster pull request is available in the management cluster", func() {
			var clusterBootstrapCopnfig string
			var clusterResourceSet string
			var crsConfigmap string
			var downloadedKubeconfigPath string
			var capdCluster ClusterConfig

			clusterNamespace := map[string]string{
				GitProviderGitLab: "capi-test-system",
				GitProviderGitHub: "default",
			}

			bootstrapLabel := "bootstrap"
			patSecret := "capi-pat"

			ginkgo.JustBeforeEach(func() {
				capdCluster = ClusterConfig{"capd", "ui-end-to-end-capd-cluster-" + strings.ToLower(RandString(6)), clusterNamespace[gitProviderEnv.Type]}
				downloadedKubeconfigPath = path.Join(os.Getenv("HOME"), "Downloads", fmt.Sprintf("%s.kubeconfig", capdCluster.Name))
				_ = deleteFile([]string{downloadedKubeconfigPath})

				createNamespace([]string{capdCluster.Namespace})
				createPATSecret(capdCluster.Namespace, patSecret)
				clusterBootstrapCopnfig = createClusterBootstrapConfig(capdCluster.Name, capdCluster.Namespace, bootstrapLabel, patSecret)
				clusterResourceSet = createClusterResourceSet(capdCluster.Name, capdCluster.Namespace)
				crsConfigmap = createCRSConfigmap(capdCluster.Name, capdCluster.Namespace)
			})

			ginkgo.JustAfterEach(func() {
				_ = deleteFile([]string{downloadedKubeconfigPath})
				deleteSecret([]string{patSecret}, capdCluster.Namespace)
				_ = gitopsTestRunner.KubectlDelete([]string{}, clusterBootstrapCopnfig)
				_ = gitopsTestRunner.KubectlDelete([]string{}, crsConfigmap)
				_ = gitopsTestRunner.KubectlDelete([]string{}, clusterResourceSet)

				reconcile("suspend", "source", "git", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
				// Force clean the repository directory for subsequent tests
				cleanGitRepository(clusterPath)
				cleanGitRepository(path.Join("./clusters", capdCluster.Namespace, capdCluster.Name))
				// Force delete capicluster incase delete PR fails to delete to free resources
				removeGitopsCapiClusters([]ClusterConfig{capdCluster})
				reconcile("resume", "source", "git", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
			})

			ginkgo.It("Verify leaf CAPD cluster can be provisioned and kubeconfig is available for cluster operations", ginkgo.Label("smoke", "integration", "capd", "git", "browser-logs"), func() {
				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)

				ginkgo.By("Add Application/Kustomization manifests to management cluster's repository main branch)", func() {
					pullGitRepo(repoAbsolutePath)
					postgres := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "postgres-manifest.yaml")
					_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("mkdir -p %[2]v && cp -f %[1]v %[2]v", postgres, path.Join(repoAbsolutePath, "apps/postgres")))

					podinfo := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "podinfo-manifest.yaml")
					_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("mkdir -p %[2]v && cp -f %[1]v %[2]v", podinfo, path.Join(repoAbsolutePath, "apps/podinfo")))

					gitUpdateCommitPush(repoAbsolutePath, "Adding postgres kustomization")
				})

				ginkgo.By("Then I Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-template-capd.yaml")
				})

				navigateToTemplatesGrid(webDriver)

				ginkgo.By("And wait for cluster-service to cache profiles", func() {
					gomega.Expect(waitForGitopsResources(context.Background(), "profiles", POLL_INTERVAL_5SECONDS, ASSERTION_15MINUTE_TIME_OUT)).To(gomega.Succeed(), "Failed to get a successful response from /v1/profiles ")
				})

				ginkgo.By("And User should choose a template", func() {
					templateTile := pages.GetTemplateTile(webDriver, "cluster-template-development-0")
					gomega.Expect(templateTile.CreateTemplate.Click()).To(gomega.Succeed())
				})

				createPage := pages.GetCreateClusterPage(webDriver)
				ginkgo.By("And wait for Create cluster page to be fully rendered", func() {
					pages.WaitForPageToLoad(webDriver)
					gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(".*Create new resource.*"))
				})

				// Parameter values
				leafCluster := ClusterConfig{
					Type:      "capi",
					Name:      capdCluster.Name,
					Namespace: capdCluster.Namespace,
				}
				k8Version := "1.23.3"
				controlPlaneMachineCount := "1"
				workerMachineCount := "1"

				var parameters = []TemplateField{
					{
						Name:   "CLUSTER_NAME",
						Value:  leafCluster.Name,
						Option: "",
					},
					{
						Name:   "CONTROL_PLANE_MACHINE_COUNT",
						Value:  "",
						Option: controlPlaneMachineCount,
					},
					{
						Name:   "INSTALL_CRDS",
						Value:  "",
						Option: "true",
					},
					{
						Name:   "KUBERNETES_VERSION",
						Value:  "",
						Option: k8Version,
					},
					{
						Name:   "NAMESPACE",
						Value:  leafCluster.Namespace,
						Option: "",
					},
					{
						Name:   "WORKER_MACHINE_COUNT",
						Value:  workerMachineCount,
						Option: "",
					},
				}

				setParameterValues(createPage, parameters)
				pages.ScrollWindow(webDriver, 0, 500)

				certManager := Application{
					DefaultApp:      true,
					Type:            "helm_release",
					Name:            "cert-manager",
					Namespace:       GITOPS_DEFAULT_NAMESPACE,
					TargetNamespace: "cert-manager",
					Version:         "0.0.8",
					Values:          `installCRDs: \${INSTALL_CRDS}`,
					Layer:           "layer-0",
				}
				profile := createPage.GetProfileInList(certManager.Name)
				AddHelmReleaseApp(profile, certManager)

				pages.ScrollWindow(webDriver, 0, 1500)
				policyAgent := Application{
					Type:            "helm_release",
					Name:            "weave-policy-agent",
					Namespace:       GITOPS_DEFAULT_NAMESPACE,
					TargetNamespace: "policy-system",
					Version:         "0.6.3",
					ValuesRegex:     `accountId: "",clusterId: ""`,
					Values:          fmt.Sprintf(`accountId: "weaveworks",clusterId: "%s"`, leafCluster.Name),
					Layer:           "layer-1",
				}
				profile = createPage.GetProfileInList(policyAgent.Name)
				AddHelmReleaseApp(profile, policyAgent)

				postgres := Application{
					Type:            "kustomization",
					Name:            "postgres",
					Namespace:       GITOPS_DEFAULT_NAMESPACE,
					TargetNamespace: GITOPS_DEFAULT_NAMESPACE,
					Path:            "apps/postgres",
				}
				gomega.Expect(createPage.AddApplication.Click()).Should(gomega.Succeed(), "Failed to click 'Add application' button")
				application := pages.GetAddApplication(webDriver, 1)
				AddKustomizationApp(application, postgres)

				podinfo := Application{
					Type:            "kustomization",
					Name:            "podinfo",
					Namespace:       GITOPS_DEFAULT_NAMESPACE,
					TargetNamespace: GITOPS_DEFAULT_NAMESPACE,
					Path:            "apps/podinfo",
				}
				gomega.Expect(createPage.AddApplication.Click()).Should(gomega.Succeed(), "Failed to click 'Add application' button")
				application = pages.GetAddApplication(webDriver, 2)
				AddKustomizationApp(application, podinfo)

				pages.ScrollWindow(webDriver, 0, 500)
				ginkgo.By("Then I should preview the PR", func() {
					preview := pages.GetPreview(webDriver)
					gomega.Eventually(func(g gomega.Gomega) {
						g.Expect(createPage.PreviewPR.Click()).Should(gomega.Succeed())
						g.Expect(preview.Title.Text()).Should(gomega.MatchRegexp("PR Preview"))

					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Succeed(), "Failed to get PR preview")

					gomega.Eventually(preview.GetPreviewTab("Cluster Definition").Click).Should(gomega.Succeed(), "Failed to switch to 'CLUSTER DEFINATION' preview tab")
					gomega.Eventually(preview.Text).Should(matchers.MatchText(`kind: Cluster[\s\w\d./:-]*metadata:[\s\w\d./:-]*labels:[\s\w\d./:-]*cni: calico`))
					gomega.Eventually(preview.Text).Should(matchers.MatchText(`kind: GitopsCluster[\s\w\d./:-]*metadata:[\s\w\d./:-]*labels:[\s\w\d./:-]*weave.works/flux: bootstrap`))
					gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`kind: GitopsCluster[\s\w\d./:-]*metadata:[\s\w\d./:-]*name: %s[\s\w\d./:-]*namespace: %s[\s\w\d./:-]*capiClusterRef`, leafCluster.Name, leafCluster.Namespace)))

					gomega.Eventually(preview.GetPreviewTab("Profiles").Click).Should(gomega.Succeed(), "Failed to switch to 'PROFILES' preview tab")
					gomega.Eventually(preview.GetPreviewTab("Kustomizations").Click).Should(gomega.Succeed(), "Failed to switch to 'KUSTOMIZATION' preview tab")

					gomega.Eventually(preview.Close.Click).Should(gomega.Succeed())
				})

				// Pull request values
				pullRequest := PullRequest{
					Branch:  fmt.Sprintf("br-%s", leafCluster.Name),
					Title:   "CAPD pull request",
					Message: "CAPD capi template",
				}
				_ = createGitopsPR(pullRequest)

				ginkgo.By("Then I should merge the pull request to start cluster provisioning", func() {
					createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
					mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
				})

				ginkgo.By("Then force reconcile flux-system to immediately start cluster provisioning", func() {
					reconcile("reconcile", "source", "git", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
					reconcile("reconcile", "", "kustomization", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
				})

				waitForGitRepoReady("flux-system", GITOPS_DEFAULT_NAMESPACE)
				waitForLeafClusterAvailability(leafCluster.Name, "^Ready")

				ginkgo.By("And I wait for the cluster to have connectivity", func() {
					// Describe GitopsCluster to check conditions
					_ = runCommandPassThrough("kubectl", "describe", "gitopsclusters.gitops.weave.works")
					waitForResourceState("ClusterConnectivity", "true", "gitopscluster", capdCluster.Namespace, "", "", ASSERTION_3MINUTE_TIME_OUT)
				})

				clusterInfo := pages.GetClustersPage(webDriver).FindClusterInList(leafCluster.Name)
				verifyDashboard(clusterInfo.GetDashboard("prometheus"), leafCluster.Name, "Prometheus")

				clustersPage := pages.GetClustersPage(webDriver)
				ginkgo.By("And I should download the kubeconfig for the CAPD capi cluster", func() {
					clusterInfo := clustersPage.FindClusterInList(leafCluster.Name)
					gomega.Expect(clusterInfo.Name.Click()).To(gomega.Succeed())
					clusterStatus := pages.GetClusterStatus(webDriver)
					gomega.Eventually(clusterStatus.Phase, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(matchers.HaveText(`"Provisioned"`))

					gomega.Eventually(func(g gomega.Gomega) {
						g.Expect(clusterStatus.KubeConfigButton.Click()).To(gomega.Succeed())
						_, err := os.Stat(downloadedKubeconfigPath)
						g.Expect(err).Should(gomega.Succeed())
					}, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_3SECONDS).ShouldNot(gomega.HaveOccurred(), "Failed to download kubeconfig for capd cluster")
				})

				ginkgo.By("And I verify cluster infrastructure for the CAPD capi cluster", func() {
					clusterInfra := pages.GetClusterInfrastructure(webDriver)
					gomega.Expect(clusterInfra.Kind.Text()).To(gomega.MatchRegexp(`DockerCluster`), "Failed to verify CAPD infarstructure provider")
				})

				ginkgo.By(fmt.Sprintf("And verify that %s capd cluster kubeconfig is correct", leafCluster.Name), func() {
					verifyCapiClusterKubeconfig(downloadedKubeconfigPath, leafCluster.Name)
				})

				// Add user roles and permissions for multi-cluster queries
				addKustomizationBases("capi", leafCluster.Name, leafCluster.Namespace)

				ginkgo.By(fmt.Sprintf("And I verify %s capd cluster is healthy and profiles are installed)", leafCluster.Name), func() {
					verifyCapiClusterHealth(downloadedKubeconfigPath, []Application{postgres, podinfo, certManager, policyAgent})
				})

				existingAppCount := getApplicationCount()

				pages.NavigateToPage(webDriver, "Applications")
				applicationsPage := pages.GetApplicationsPage(webDriver)
				pages.WaitForPageToLoad(webDriver)

				ginkgo.By(fmt.Sprintf("And filter capi cluster '%s' application", leafCluster.Name), func() {
					totalAppCount := existingAppCount + 6 // flux-system, clusters-bases-kustomization, cert-manager, policy-agent, postgres, podinfo
					gomega.Eventually(applicationsPage.CountApplications, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(totalAppCount), fmt.Sprintf("There should be %d application enteries in application table", totalAppCount))

					filterID := "clusterName: " + leafCluster.Namespace + `/` + leafCluster.Name
					searchPage := pages.GetSearchPage(webDriver)
					searchPage.SelectFilter("cluster", filterID)
					gomega.Eventually(applicationsPage.CountApplications).Should(gomega.Equal(6), "There should be 6 application enteries in application table")
				})

				verifyAppInformation(applicationsPage, certManager, leafCluster, "Ready")
				verifyAppInformation(applicationsPage, policyAgent, leafCluster, "Ready")
				verifyAppInformation(applicationsPage, postgres, leafCluster, "Ready")
				verifyAppInformation(applicationsPage, podinfo, leafCluster, "Ready")

				ginkgo.By("Then I should select the cluster to create the delete pull request", func() {
					pages.NavigateToPage(webDriver, "Clusters")
					gomega.Eventually(clustersPage.FindClusterInList(leafCluster.Name).Status, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(matchers.BeFound())
					clusterInfo := clustersPage.FindClusterInList(leafCluster.Name)
					gomega.Expect(clusterInfo.Checkbox.Click()).To(gomega.Succeed())

					gomega.Eventually(webDriver.FindByXPath(`//button[@id="delete-cluster"][@disabled]`)).ShouldNot(matchers.BeFound())
					gomega.Expect(clustersPage.PRDeleteClusterButton.Click()).To(gomega.Succeed())

					deletePR := pages.GetDeletePRPopup(webDriver)
					gomega.Expect(deletePR.PRDescription.SendKeys("Delete CAPD capi cluster, it is not required any more")).To(gomega.Succeed())

					AuthenticateWithGitProvider(webDriver, gitProviderEnv.Type, gitProviderEnv.Hostname)
					gomega.Eventually(deletePR.GitCredentials).Should(matchers.BeVisible())

					gomega.Expect(deletePR.DeleteClusterButton.Click()).To(gomega.Succeed())
				})

				ginkgo.By("Then I should see a toast with a link to the deletion PR", func() {
					gitops := pages.GetGitOps(webDriver)
					gomega.Eventually(gitops.PRLinkBar, ASSERTION_1MINUTE_TIME_OUT).Should(matchers.BeFound())
				})

				var deletePRUrl string
				ginkgo.By("And I should veriyfy the delete pull request in the cluster config repository", func() {
					deletePRUrl = verifyPRCreated(gitProviderEnv, repoAbsolutePath)
				})

				ginkgo.By("Then I should merge the delete pull request to delete cluster", func() {
					mergePullRequest(gitProviderEnv, repoAbsolutePath, deletePRUrl)
				})

				ginkgo.By("And the delete pull request manifests are not present in the cluster config repository", func() {
					pullGitRepo(repoAbsolutePath)
					_, err := os.Stat(path.Join(repoAbsolutePath, clusterPath, leafCluster.Namespace, leafCluster.Name+".yaml"))
					gomega.Expect(err).Should(gomega.MatchError(os.ErrNotExist), "Cluster config is found when expected to be deleted.")

					_, err = os.Stat(path.Join(repoAbsolutePath, "clusters", leafCluster.Namespace, leafCluster.Name))
					gomega.Expect(err).Should(gomega.MatchError(os.ErrNotExist), "Cluster kustomizations are found when expected to be deleted.")
				})

				ginkgo.By(fmt.Sprintf("Then I should see the '%s' cluster deleted", leafCluster.Name), func() {
					clusterFound := func() error {
						return runCommandPassThrough("kubectl", "get", "cluster", leafCluster.Name, "-n", capdCluster.Namespace)
					}
					gomega.Eventually(clusterFound, ASSERTION_5MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.HaveOccurred())
				})
			})

			ginkgo.It("Verify CAPI cluster resource can be modified/edited", ginkgo.Label("integration", "capd", "git"), func() {
				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)

				ginkgo.By("Add Application/Kustomization manifests to management cluster's repository main branch)", func() {
					pullGitRepo(repoAbsolutePath)
					postgres := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "postgres-manifest.yaml")
					_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("mkdir -p %[2]v && cp -f %[1]v %[2]v", postgres, path.Join(repoAbsolutePath, "apps/postgres")))

					podinfo := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "podinfo-manifest.yaml")
					_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("mkdir -p %[2]v && cp -f %[1]v %[2]v", podinfo, path.Join(repoAbsolutePath, "apps/podinfo")))

					gitUpdateCommitPush(repoAbsolutePath, "Adding postgres kustomization")
				})

				ginkgo.By("Then I Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-template-capd.yaml")
				})

				navigateToTemplatesGrid(webDriver)

				ginkgo.By("And wait for cluster-service to cache profiles", func() {
					gomega.Expect(waitForGitopsResources(context.Background(), "profiles", POLL_INTERVAL_5SECONDS, ASSERTION_15MINUTE_TIME_OUT)).To(gomega.Succeed(), "Failed to get a successful response from /v1/profiles ")
				})

				ginkgo.By("And User should choose a template", func() {
					templateTile := pages.GetTemplateTile(webDriver, "cluster-template-development-0")
					gomega.Expect(templateTile.CreateTemplate.Click()).To(gomega.Succeed())
				})

				createPage := pages.GetCreateClusterPage(webDriver)
				ginkgo.By("And wait for Create cluster page to be fully rendered", func() {
					pages.WaitForPageToLoad(webDriver)
					gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(".*Create new resource.*"))
				})

				// Parameter values
				leafCluster := ClusterConfig{
					Type:      "capi",
					Name:      capdCluster.Name,
					Namespace: capdCluster.Namespace,
				}
				k8Version := "1.23.3"
				controlPlaneMachineCount := "1"
				workerMachineCount := "1"

				var parameters = []TemplateField{
					{
						Name:   "CLUSTER_NAME",
						Value:  leafCluster.Name,
						Option: "",
					},
					{
						Name:   "CONTROL_PLANE_MACHINE_COUNT",
						Value:  "",
						Option: controlPlaneMachineCount,
					},
					{
						Name:   "INSTALL_CRDS",
						Value:  "",
						Option: "true",
					},
					{
						Name:   "KUBERNETES_VERSION",
						Value:  "",
						Option: k8Version,
					},
					{
						Name:   "NAMESPACE",
						Value:  leafCluster.Namespace,
						Option: "",
					},
					{
						Name:   "WORKER_MACHINE_COUNT",
						Value:  workerMachineCount,
						Option: "",
					},
				}

				setParameterValues(createPage, parameters)
				pages.ScrollWindow(webDriver, 0, 500)

				certManager := Application{
					DefaultApp:      true,
					Type:            "helm_release",
					Name:            "cert-manager",
					Namespace:       GITOPS_DEFAULT_NAMESPACE,
					TargetNamespace: "cert-manager",
					Version:         "0.0.8",
					Values:          `installCRDs: \${INSTALL_CRDS}`,
					Layer:           "layer-0",
				}
				profile := createPage.GetProfileInList(certManager.Name)
				AddHelmReleaseApp(profile, certManager)

				metallb := Application{
					Type:            "helm_release",
					Name:            "metallb",
					Namespace:       GITOPS_DEFAULT_NAMESPACE,
					TargetNamespace: GITOPS_DEFAULT_NAMESPACE,
					Version:         "0.0.3",
					ValuesRegex:     `namespace: ""`,
					Values:          fmt.Sprintf(`namespace: "%s"`, clusterNamespace),
					Layer:           "layer-0",
				}
				profile = createPage.GetProfileInList(metallb.Name)
				AddHelmReleaseApp(profile, metallb)

				postgres := Application{
					Type:            "kustomization",
					Name:            "postgres",
					Namespace:       GITOPS_DEFAULT_NAMESPACE,
					TargetNamespace: GITOPS_DEFAULT_NAMESPACE,
					Path:            "apps/postgres",
				}
				gomega.Expect(createPage.AddApplication.Click()).Should(gomega.Succeed(), "Failed to click 'Add application' button")
				application := pages.GetAddApplication(webDriver, 1)
				AddKustomizationApp(application, postgres)

				pages.ScrollWindow(webDriver, 0, 500)
				// Pull request values
				pullRequest := PullRequest{
					Branch:  "br-" + leafCluster.Name,
					Title:   "CAPD pull request",
					Message: "CAPD capi template",
				}
				_ = createGitopsPR(pullRequest)

				ginkgo.By("Then I should merge the pull request to start cluster provisioning", func() {
					createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
					mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
				})

				ginkgo.By("Then force reconcile flux-system to immediately start cluster provisioning", func() {
					reconcile("reconcile", "source", "git", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
					reconcile("reconcile", "", "kustomization", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
				})

				waitForGitRepoReady("flux-system", GITOPS_DEFAULT_NAMESPACE)
				waitForLeafClusterAvailability(leafCluster.Name, "^Ready")

				clustersPage := pages.GetClustersPage(webDriver)
				ginkgo.By("And I wait for the cluster to have connectivity", func() {
					// Describe GitopsCluster to check conditions
					_ = runCommandPassThrough("kubectl", "describe", "gitopsclusters.gitops.weave.works")
					waitForResourceState("ClusterConnectivity", "true", "gitopscluster", capdCluster.Namespace, "", "", ASSERTION_3MINUTE_TIME_OUT)
				})

				ginkgo.By("And I should download the kubeconfig for the CAPD capi cluster", func() {
					clusterInfo := clustersPage.FindClusterInList(leafCluster.Name)
					gomega.Expect(clusterInfo.Name.Click()).To(gomega.Succeed())
					clusterStatus := pages.GetClusterStatus(webDriver)
					gomega.Eventually(clusterStatus.Phase, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(matchers.HaveText(`"Provisioned"`))

					gomega.Eventually(func(g gomega.Gomega) {
						g.Expect(clusterStatus.KubeConfigButton.Click()).To(gomega.Succeed())
						_, err := os.Stat(downloadedKubeconfigPath)
						g.Expect(err).Should(gomega.Succeed())
					}, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_3SECONDS).ShouldNot(gomega.HaveOccurred(), "Failed to download kubeconfig for capd cluster")
				})

				ginkgo.By(fmt.Sprintf("And verify that %s capd cluster kubeconfig is correct", leafCluster.Name), func() {
					verifyCapiClusterKubeconfig(downloadedKubeconfigPath, leafCluster.Name)
				})

				// Add user roles and permissions for multi-cluster queries
				addKustomizationBases("capi", leafCluster.Name, leafCluster.Namespace)

				ginkgo.By(fmt.Sprintf("And I verify %s capd cluster is healthy and profiles are installed)", leafCluster.Name), func() {
					verifyCapiClusterHealth(downloadedKubeconfigPath, []Application{certManager, metallb})
				})

				existingAppCount := getApplicationCount()

				pages.NavigateToPage(webDriver, "Applications")
				applicationsPage := pages.GetApplicationsPage(webDriver)
				pages.WaitForPageToLoad(webDriver)

				ginkgo.By(fmt.Sprintf("And filter capi cluster '%s' application", leafCluster.Name), func() {
					totalAppCount := existingAppCount + 5 // flux-system, clusters-bases-kustomization, cert-manager, metallb, postgres
					gomega.Eventually(applicationsPage.CountApplications, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(totalAppCount), fmt.Sprintf("There should be %d application enteries in application table", totalAppCount))

					filterID := "clusterName: " + leafCluster.Namespace + `/` + leafCluster.Name
					searchPage := pages.GetSearchPage(webDriver)
					searchPage.SelectFilter("cluster", filterID)
					gomega.Eventually(applicationsPage.CountApplications).Should(gomega.Equal(5), "There should be 5 application enteries in application table")
				})

				verifyAppInformation(applicationsPage, metallb, leafCluster, "Ready")
				verifyAppInformation(applicationsPage, certManager, leafCluster, "Ready")
				verifyAppInformation(applicationsPage, postgres, leafCluster, "Ready")

				pages.NavigateToPage(webDriver, "Clusters")
				ginkgo.By(fmt.Sprintf("Then I should start editing the %s cluster", leafCluster.Name), func() {
					clusterInfo := clustersPage.FindClusterInList(leafCluster.Name)
					gomega.Eventually(clusterInfo.EditCluster.Click, ASSERTION_30SECONDS_TIME_OUT).Should(gomega.Succeed(), "Failed to click 'EDIT CLUSTER' button")
				})

				createPage = pages.GetCreateClusterPage(webDriver)
				ginkgo.By("And wait for Create/Edit cluster page to be fully rendered", func() {
					pages.WaitForPageToLoad(webDriver)
					gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(fmt.Sprintf(`.*%s.*`, leafCluster.Name)))
				})

				workerMachineCount = "2"
				parameters = []TemplateField{
					{
						Name:   "WORKER_MACHINE_COUNT",
						Value:  workerMachineCount,
						Option: "",
					},
				}

				setParameterValues(createPage, parameters)
				pages.ScrollWindow(webDriver, 0, 500)

				ginkgo.By("And edit cluster to remove profile/application", func() {
					// Deleting metallb profile
					profile = createPage.GetProfileInList(metallb.Name)
					gomega.Eventually(profile.Checkbox).Should(matchers.BeSelected(), fmt.Sprintf("Profile %s should already be selected", metallb.Name))
					gomega.Eventually(profile.Checkbox.Uncheck).Should(gomega.Succeed(), fmt.Sprintf("Failed to unselect the %s profile", metallb.Name))

					// Deleting postgres application
					application = pages.GetAddApplication(webDriver, 1)
					gomega.Expect(application.RemoveApplication.Click()).Should(gomega.Succeed(), fmt.Sprintf("Failed to remove application no. %d", 1))
				})

				// Now edit cluster to add a new application podinfo
				podinfo := Application{
					Type:            "kustomization",
					Name:            "podinfo",
					Namespace:       GITOPS_DEFAULT_NAMESPACE,
					TargetNamespace: "test-system",
					Path:            "apps/podinfo",
				}
				gomega.Expect(createPage.AddApplication.Click()).Should(gomega.Succeed(), "Failed to click 'Add application' button")
				application = pages.GetAddApplication(webDriver, 1)
				AddKustomizationApp(application, podinfo)

				pages.ScrollWindow(webDriver, 0, 500)
				_ = createGitopsPR(PullRequest{}) // accepts default pull request values

				ginkgo.By("Then I should merge the pull request to start cluster provisioning", func() {
					createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
					mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
				})

				ginkgo.By("Then force reconcile flux-system to immediately start cluster modification take effect", func() {
					reconcile("reconcile", "source", "git", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
					reconcile("reconcile", "", "kustomization", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
				})

				ginkgo.By(fmt.Sprintf("And I verify %s capd cluster is healthy and edited as expected)", leafCluster.Name), func() {
					// 1 controlPlaneMachine and 2 workerMachine nodes
					gomega.Eventually(func(g gomega.Gomega) int {
						stdOut, _ := runCommandAndReturnStringOutput(fmt.Sprintf(`kubectl --kubeconfig=%s get nodes | grep %s | wc -l`, downloadedKubeconfigPath, leafCluster.Name))
						count, _ := strconv.Atoi(strings.Trim(stdOut, " "))
						return count
					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Equal(3), "Failed to verify number of expected nodes")

					verifyCapiClusterHealth(downloadedKubeconfigPath, []Application{podinfo, certManager})
					// Verify metallb has been removed after editing leaf cluster
					cmd := fmt.Sprintf("kubectl --kubeconfig=%s get deploy %s -n %s", downloadedKubeconfigPath, metallb.TargetNamespace+"-metallb-controller", metallb.TargetNamespace)
					logger.Trace(cmd)
					gomega.Eventually(func(g gomega.Gomega) {
						g.Expect(runCommandPassThrough(cmd)).Should(gomega.HaveOccurred())
					}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Succeed(), metallb.Name+" application failed to uninstall")
				})

				pages.NavigateToPage(webDriver, "Applications")
				pages.WaitForPageToLoad(webDriver)

				ginkgo.By(fmt.Sprintf("And filter capi cluster '%s' application", leafCluster.Name), func() {
					totalAppCount := existingAppCount + 4 // flux-system, clusters-bases-kustomization, cert-manager, podinfo
					gomega.Eventually(applicationsPage.CountApplications, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(totalAppCount), fmt.Sprintf("There should be %d application enteries in application table", totalAppCount))

					filterID := "clusterName: " + leafCluster.Namespace + `/` + leafCluster.Name
					searchPage := pages.GetSearchPage(webDriver)
					searchPage.SelectFilter("cluster", filterID)
					gomega.Eventually(applicationsPage.CountApplications).Should(gomega.Equal(4), "There should be 4 application enteries in application table")
				})

				verifyAppInformation(applicationsPage, certManager, leafCluster, "Ready")
				verifyAppInformation(applicationsPage, podinfo, leafCluster, "Ready")
			})
		})

		ginkgo.Context("[UI] When entitlement is available in the cluster", func() {
			DEPLOYMENT_APP := "my-mccp-cluster-service"

			checkEntitlement := func(typeEntitelment string, beFound bool) {
				checkOutput := func() bool {
					if !pages.ElementExist(pages.GetClustersPage(webDriver).Version) {
						gomega.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
					}
					loginUser()
					found, _ := pages.GetEntitelment(webDriver, typeEntitelment).Visible()
					return found

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
					gomega.Expect(gitopsTestRunner.KubectlApply([]string{}, path.Join(getCheckoutRepoPath(), "test", "utils", "scripts", "entitlement-secret.yaml")), "Failed to create/configure entitlement")
				})

				ginkgo.By("Then I restart the cluster service pod for valid entitlemnt to take effect", func() {
					gomega.Expect(gitopsTestRunner.RestartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
				})

				ginkgo.By("And the Cluster service is healthy", func() {
					CheckClusterService(capi_endpoint_url)
				})

				ginkgo.By("And I should not see the error or warning message for valid entitlement", func() {
					checkEntitlement("expired", false)
					checkEntitlement("missing", false)
				})
			})

			ginkgo.It("Verify cluster service acknowledges the entitlement presences", ginkgo.Label("integration"), func() {

				ginkgo.By("When I delete the entitlement", func() {
					gomega.Expect(gitopsTestRunner.KubectlDelete([]string{}, path.Join(getCheckoutRepoPath(), "test", "utils", "scripts", "entitlement-secret.yaml")), "Failed to delete entitlement secret")
				})

				ginkgo.By("Then I restart the cluster service pod for missing entitlemnt to take effect", func() {
					gomega.Expect(gitopsTestRunner.RestartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE)).ShouldNot(gomega.HaveOccurred(), "Failed restart deployment successfully")
				})

				ginkgo.By("And I should see the error message for missing entitlement", func() {
					checkEntitlement("missing", true)

				})

				ginkgo.By("When I apply the expired entitlement", func() {
					gomega.Expect(gitopsTestRunner.KubectlApply([]string{}, path.Join(getCheckoutRepoPath(), "test", "utils", "data", "entitlement-secret-expired.yaml")), "Failed to create/configure entitlement")
				})

				ginkgo.By("Then I restart the cluster service pod for expired entitlemnt to take effect", func() {
					gomega.Expect(gitopsTestRunner.RestartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
				})

				ginkgo.By("And I should see the warning message for expired entitlement", func() {
					checkEntitlement("expired", true)
				})

				ginkgo.By("When I apply the invalid entitlement", func() {
					gomega.Expect(gitopsTestRunner.KubectlApply([]string{}, path.Join(getCheckoutRepoPath(), "test", "utils", "data", "entitlement-secret-invalid.yaml")), "Failed to create/configure entitlement")
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
