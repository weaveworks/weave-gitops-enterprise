package acceptance

import (
	"context"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

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
			gomega.Eventually(func(g gomega.Gomega) {
				g.Eventually(createPage.GetTemplateParameter(webDriver, parameters[i].Name).ListBox.Click).Should(gomega.Succeed())
				g.Eventually(pages.GetOption(webDriver, parameters[i].Option).Click).Should(gomega.Succeed())
				g.Expect(createPage.GetTemplateParameter(webDriver, parameters[i].Name).ListBox).Should(matchers.MatchText(parameters[i].Option))
			}, ASSERTION_30SECONDS_TIME_OUT).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to select %s parameter option: %s", parameters[i].Name, parameters[i].Option))
		} else {
			field := createPage.GetTemplateParameter(webDriver, parameters[i].Name).Field
			pages.ClearFieldValue(field)
			gomega.Expect(field.SendKeys(parameters[i].Value)).To(gomega.Succeed())
		}
	}
}

func DescribeTemplates(gitopsTestRunner GitopsTestRunner) {
	var _ = ginkgo.Describe("Multi-Cluster Control Plane GitOpsTemplates", func() {
		templateFiles := []string{}

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

		ginkgo.Context("[UI] When no GitOps Templates are available in the cluster", func() {
			ginkgo.It("Verify template page renders no GitOpsTemplate", ginkgo.Label("integration"), func() {
				ginkgo.By("And wait for  good looking response from /v1/templates", func() {
					gomega.Expect(waitForGitopsResources(context.Background(), Request{Path: "templates"}, POLL_INTERVAL_15SECONDS)).To(gomega.Succeed(), "Failed to get a successful response from /v1/templates")
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

		ginkgo.Context("[UI] When GitOps Templates are available in the cluster", func() {
			ginkgo.It("Verify template(s) are rendered from the template library.", func() {
				awsTemplateCount := 2
				eksFargateTemplateCount := 2
				azureTemplateCount := 3
				capdTemplateCount := 5
				totalTemplateCount := awsTemplateCount + eksFargateTemplateCount + azureTemplateCount + capdTemplateCount

				ordered_template_list := func() []string {
					expected_list := make([]string, totalTemplateCount)
					for i := 0; i < awsTemplateCount; i++ {
						expected_list[i] = fmt.Sprintf("capa-cluster-template-%d", i)
					}
					for i := 0; i < eksFargateTemplateCount; i++ {
						expected_list[i] = fmt.Sprintf("capa-cluster-template-eks-fargate-%d", i)
					}
					for i := 0; i < capdTemplateCount; i++ {
						expected_list[i] = fmt.Sprintf("capd-cluster-template-%d", i)
					}
					for i := 0; i < azureTemplateCount; i++ {
						expected_list[i] = fmt.Sprintf("capz-cluster-template-%d", i)
					}

					sort.Strings(expected_list)
					return expected_list
				}()

				ginkgo.By("Apply/Install CAPITemplate", func() {
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(5, "templates/cluster/docker/cluster-template.yaml")...)
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(3, "templates/cluster/azure/cluster-template-e2e.yaml")...)
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(2, "templates/cluster/aws/cluster-template-ec2.yaml")...)
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(2, "templates/cluster/aws/cluster-template-eks-fargate.yaml")...)
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

			ginkgo.It("Verify I should be able to select a template of my choice", func() {

				// test selection with 50 capiTemplates
				templateFiles = gitopsTestRunner.CreateApplyCapitemplates(50, "templates/miscellaneous/templated-cluster-template.yaml")

				navigateToTemplatesGrid(webDriver)
				templatesPage := pages.GetTemplatesPage(webDriver)

				ginkgo.By("And I should choose a template - grid view", func() {
					templateTile := pages.GetTemplateTile(webDriver, "capg-cluster-template-9")

					gomega.Eventually(templateTile.Description).Should(matchers.MatchText("This is the std. CAPG template 9"))
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
					templateRow := templatesPage.GetTemplateInformation(webDriver, "capg-cluster-template-10")
					gomega.Eventually(templateRow.Type).Should(matchers.MatchText(""))
					gomega.Eventually(templateRow.Namespace).Should(matchers.MatchText("default"))
					gomega.Eventually(templateRow.Provider).Should(matchers.MatchText(""))
					gomega.Eventually(templateRow.Description).Should(matchers.MatchText("This is the std. CAPG template 10"))
					gomega.Expect(templateRow.CreateTemplate).Should(matchers.BeFound())
					gomega.Expect(templateRow.CreateTemplate.Click()).To(gomega.Succeed())
				})

				ginkgo.By("And wait for Create cluster page to be fully rendered - table view", func() {
					createPage := pages.GetCreateClusterPage(webDriver)
					gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(".*Create new resource.*"))
				})
			})

			ginkgo.It("Verify UI shows message related to an invalid template(s) when valid templates are not available", func() {

				ginkgo.By("Apply/Install invalid CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "templates/miscellaneous/invalid-cluster-template.yaml")
				})

				navigateToTemplatesGrid(webDriver)
				templatesPage := pages.GetTemplatesPage(webDriver)

				ginkgo.By("And User should see message informing user of the invalid template in the cluster - grid view", func() {
					templateTile := pages.GetTemplateTile(webDriver, "invalid-cluster-template-0")
					gomega.Eventually(templateTile.ErrorHeader).Should(matchers.BeFound())
					gomega.Expect(templateTile.ErrorDescription).Should(matchers.BeFound())
					gomega.Expect(templateTile.CreateTemplate).ShouldNot(matchers.BeEnabled())
				})

				ginkgo.By("And I should change the templates view to 'table'", func() {
					gomega.Expect(templatesPage.SelectView("table").Click()).To(gomega.Succeed())
				})

				ginkgo.By("And User should see message informing user of the invalid template in the cluster - table view", func() {
					templateRow := templatesPage.GetTemplateInformation(webDriver, "invalid-cluster-template-0")
					gomega.Eventually(templateRow.Type).Should(matchers.MatchText(""))
					gomega.Eventually(templateRow.Namespace).Should(matchers.MatchText("default"))
					gomega.Eventually(templateRow.Provider).Should(matchers.MatchText(""))
					gomega.Eventually(templateRow.Description).Should(matchers.MatchText("Couldn't load template body"))
					gomega.Expect(templateRow.CreateTemplate).ShouldNot(matchers.BeEnabled())
				})
			})

			ginkgo.It("Verify UI shows message related to an invalid template(s) and renders the available valid template(s)", func() {

				noOfValidTemplates := 3
				ginkgo.By("Apply/Install valid CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(noOfValidTemplates, "templates/cluster/aws/cluster-template-eks-fargate.yaml")
				})

				noOfInvalidTemplates := 1
				ginkgo.By("Apply/Install invalid CAPITemplate", func() {
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(noOfInvalidTemplates, "templates/miscellaneous/invalid-cluster-template.yaml")...)
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
					templateTile := pages.GetTemplateTile(webDriver, "invalid-cluster-template-0")
					gomega.Eventually(templateTile.ErrorHeader).Should(matchers.BeFound())
					gomega.Expect(templateTile.ErrorDescription).Should(matchers.BeFound())
					gomega.Expect(templateTile.CreateTemplate).ShouldNot(matchers.BeEnabled())
				})
			})

			ginkgo.It("Verify template parameters should be rendered dynamically and can be set for the selected template", ginkgo.Label("integration"), func() {

				ginkgo.By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "templates/cluster/aws/cluster-template-eks-fargate.yaml")
				})

				navigateToTemplatesGrid(webDriver)
				templatesPage := pages.GetTemplatesPage(webDriver)

				ginkgo.By("And I should change the templates view to 'table'", func() {
					gomega.Expect(templatesPage.SelectView("table").Click()).To(gomega.Succeed())
				})

				ginkgo.By("And I should choose a template - table view", func() {
					templateRow := templatesPage.GetTemplateInformation(webDriver, "capa-cluster-template-eks-fargate-0")
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

		ginkgo.Context("[UI] When GitOps Template are available in the management cluster for resource creation", func() {
			clusterPath := "./clusters/management/clusters"
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

			ginkgo.It("Verify pull request for cluster can be created to the management cluster", ginkgo.Label("integration", "git"), func() {
				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)

				ginkgo.By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "templates/cluster/docker/cluster-template.yaml")
				})

				navigateToTemplatesGrid(webDriver)

				templateName := "capd-cluster-template-0"
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
					gomega.Eventually(preview.GetPreviewTab("Resource Definition").Click).Should(gomega.Succeed(), "Failed to switch to 'RESOURCE DEFINITION' preview tab")
					gomega.Eventually(preview.Text).Should(matchers.MatchText(`kind: Cluster[\s\w\d./:-]*metadata:[\s\w\d./:-]*labels:[\s\w\d./:-]*cni: calico`))
					gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`kind: GitopsCluster[\s\w\d./:-]*metadata:[\s\w\d./:-]*labels:[\s\w\d./:-]*templates.weave.works/template-name: %s`, templateName)))
					gomega.Eventually(preview.Text).Should(matchers.MatchText(`kind: GitopsCluster[\s\w\d./:-]*metadata:[\s\w\d./:-]*labels:[\s\w\d./:-]*templates.weave.works/template-namespace: default`))
					gomega.Eventually(preview.Text).Should(matchers.MatchText(`kind: GitopsCluster[\s\w\d./:-]*metadata:[\s\w\d./:-]*labels:[\s\w\d./:-]*weave.works/flux: bootstrap`))
					gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`kind: GitopsCluster[\s\w\d./:-]*metadata:[\s\w\d./:-]*name: %s[\s\w\d./:-]*namespace: %s[\s\w\d./:-]*capiClusterRef`, leafCluster.Name, leafCluster.Namespace)))

					// Verify profiles preview
					gomega.Eventually(preview.GetPreviewTab("Profiles").Click).Should(gomega.Succeed(), "Failed to switch to 'PROFILES' preview tab")
					gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`kind: HelmRepository[\s\w\d./:-]*name: %s[\s\w\d./:-]*namespace: %s[\s\w\d./:-]*url: %s`, certManager.Chart, GITOPS_DEFAULT_NAMESPACE, sourceHRUrl)))
					gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`kind: HelmRelease[\s\w\d./:-]*name: %s[\s\w\d./:-]*namespace: %s[\s\w\d./:-]*spec`, certManager.Name, certManager.Namespace)))
					// Need to enable/update this check when profiles will eventually move out from annotations
					// gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`chart: %s[\s\w\d./:-]*sourceRef:[\s\w\d./:-]*name: %s[\s\w\d./:-]*version: %s[\s\w\d./:-]*targetNamespace: %s[\s\w\d./:-]*installCRDs: true`, certManager.Name, certManager.Chart, certManager.Version, certManager.TargetNamespace)))

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

			ginkgo.It("Verify pull request can not be created by using exiting repository branch", ginkgo.Label("integration", "git"), func() {
				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
				// Checkout repo main branch in case of test failure
				defer checkoutRepoBranch(repoAbsolutePath, "main")

				branchName := "ui-test-branch"
				ginkgo.By("And create new git repository branch", func() {
					_ = createGitRepoBranch(repoAbsolutePath, branchName)
				})

				ginkgo.By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "templates/cluster/docker/cluster-template.yaml")
				})

				navigateToTemplatesGrid(webDriver)

				ginkgo.By("And User should choose a template", func() {
					templateTile := pages.GetTemplateTile(webDriver, "capd-cluster-template-0")
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
				messages := pages.GetMessages(webDriver)
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

					if pages.ElementExist(messages.Error) {
						gomega.Expect(messages.Close.Click()).To(gomega.Succeed())
					}

					gomega.Expect(gitops.CreatePR.Click()).To(gomega.Succeed(), "Failed to click 'CREATE PULL REQUEST' button")
				})

				ginkgo.By("Then I should not see pull request error creation message", func() {
					gomega.Eventually(messages.Error).Should(matchers.MatchText(fmt.Sprintf(`unable to create pull request.+unable to create new branch "%s"`, branchName)))
				})
			})
		})
	})
}
