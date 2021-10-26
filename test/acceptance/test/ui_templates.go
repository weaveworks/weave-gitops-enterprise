package acceptance

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	. "github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

type TemplateField struct {
	Name   string
	Value  string
	Option string
}

func setParameterValues(createPage *pages.CreateCluster, paramSection map[string][]TemplateField) {
	for section, parameters := range paramSection {
		By(fmt.Sprintf("And set template section %s parameter values", section), func() {
			templateSection := createPage.GetTemplateSection(webDriver, section)
			Expect(templateSection.Name).Should(HaveText(section))

			for i := 0; i < len(parameters); i++ {
				paramSet := false
				for j := 0; j < len(templateSection.Fields); j++ {
					val, _ := templateSection.Fields[j].Label.Text()
					if strings.Contains(val, parameters[i].Name) {
						By("And set template parameter values", func() {
							if parameters[i].Option != "" {
								Expect(templateSection.Fields[j].ListBox.Click()).To(Succeed())
								Expect(pages.GetParameterOption(webDriver, parameters[i].Option).Click()).To(Succeed())
							} else {
								Expect(templateSection.Fields[j].Field.SendKeys(parameters[i].Value)).To(Succeed())
							}
						})
						paramSet = true
					}
				}
				Expect(paramSet).Should(BeTrue(), fmt.Sprintf("Parameter '%s' isn't found in section '%s' ", parameters[i].Name, section))
			}
		})
	}
}

func DescribeTemplates(gitopsTestRunner GitopsTestRunner) {
	var _ = Describe("Multi-Cluster Control Plane Templates", func() {

		GITOPS_BIN_PATH := GetGitopsBinPath()

		templateFiles := []string{}

		BeforeEach(func() {

			By("Given Kubernetes cluster is setup", func() {
				gitopsTestRunner.CheckClusterService()
			})
			initializeWebdriver()
		})

		AfterEach(func() {
			gitopsTestRunner.DeleteApplyCapiTemplates(templateFiles)
			// Reset/empty the templateFiles list
			templateFiles = []string{}
		})

		Context("[UI] When no Capi Templates are available in the cluster", func() {
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

		Context("[UI] When Capi Templates are available in the cluster", func() {
			It("Verify template(s) are rendered from the template library.", func() {
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

				By("Apply/Install CAPITemplate", func() {
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(5, "capi-server-v1-template-capd.yaml")...)
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(3, "capi-server-v1-template-azure.yaml")...)
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(2, "capi-server-v1-template-aws.yaml")...)
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(2, "capi-server-v1-template-eks-fargate.yaml")...)
				})

				pages.NavigateToPage(webDriver, "Templates")
				templatesPage := pages.GetTemplatesPage(webDriver)

				By("And wait for Templates page to be fully rendered", func() {
					templatesPage.WaitForPageToLoad(webDriver)
					Eventually(templatesPage.TemplateHeader).Should(BeVisible())
					Eventually(templatesPage.TemplateCount).Should(MatchText(`[0-9]+`))

					count, _ := templatesPage.TemplateCount.Text()
					templateCount, _ := strconv.Atoi(count)
					tileCount, _ := templatesPage.TemplateTiles.Count()

					Eventually(templateCount).Should(Equal(totalTemplateCount), "The template header count should be equal to templates created")
					Eventually(tileCount).Should(Equal(totalTemplateCount), "The number of template tiles rendered should be equal to number of templates created")
				})

				By("And I should change the templates view to 'table'", func() {
					Expect(templatesPage.SelectView("table").Click()).To(Succeed())
					rowCount, _ := templatesPage.TemplatesTable.Count()
					Eventually(rowCount).Should(Equal(totalTemplateCount), "The number of template tiles rendered should be equal to number of templates created")

				})

				By("And templates are ordered - table view", func() {
					actual_list := templatesPage.GetTemplateTableList()
					for i := 0; i < totalTemplateCount; i++ {
						Expect(actual_list[i]).Should(ContainSubstring(ordered_template_list[i]))
					}
				})

				By("And templates can be filtered by provider - table view", func() {
					// Select cluster provider by selecting from the popup list
					Expect(templatesPage.TemplateProvider.Click()).To(Succeed())
					Expect(templatesPage.SelectProvider("aws").Click()).To(Succeed())

					rowCount, _ := templatesPage.TemplatesTable.Count()
					Eventually(rowCount).Should(Equal(4), "The number of selected template tiles rendered should be equal to number of aws templates created")

					Expect(templatesPage.TemplateProvider.Click()).To(Succeed())
					Expect(templatesPage.TemplateProvider.SendKeys("\uE003")).To(Succeed()) // sending back space key

					rowCount, _ = templatesPage.TemplateTiles.Count()
					Eventually(rowCount).Should(Equal(totalTemplateCount), "The number of template tiles rendered should be equal to number of templates created")

				})

				By("And I should change the templates view to 'grid'", func() {
					Expect(templatesPage.SelectView("grid").Click()).To(Succeed())
					tileCount, _ := templatesPage.TemplateTiles.Count()
					Eventually(tileCount).Should(Equal(totalTemplateCount), "The number of template tiles rendered should be equal to number of templates created")
				})

				By("And templates are ordered - grid view", func() {
					actual_list := templatesPage.GetTemplateTileList()
					for i := 0; i < totalTemplateCount; i++ {
						Expect(actual_list[i]).Should(ContainSubstring(ordered_template_list[i]))
					}
				})

				By("And templates can be filtered by provider - grid view", func() {
					// Select cluster provider by selecting from the popup list
					Expect(templatesPage.TemplateProvider.Click()).To(Succeed())
					Expect(templatesPage.SelectProvider("aws").Click()).To(Succeed())

					tileCount, _ := templatesPage.TemplateTiles.Count()
					Eventually(tileCount).Should(Equal(awsTemplateCount+eksFargateTemplateCount), "The number of aws provider template tiles rendered should be equal to number of aws templates created")

					// Select cluster provider by typing the provider name
					Expect(templatesPage.TemplateProvider.Click()).To(Succeed())
					Expect(templatesPage.TemplateProvider.SendKeys("\uE003")).To(Succeed()) // sending back space key
					Expect(templatesPage.TemplateProvider.SendKeys("azure")).To(Succeed())
					Expect(templatesPage.TemplateProviderPopup.At(0).Click()).To(Succeed())

					tileCount, _ = templatesPage.TemplateTiles.Count()
					Eventually(tileCount).Should(Equal(azureTemplateCount), "The number of azure provider template tiles rendered should be equal to number of azure templates created")
				})
			})
		})

		Context("[UI] When Capi Templates are available in the cluster", func() {
			It("Verify I should be able to select a template of my choice", func() {

				// test selection with 50 capiTemplates
				templateFiles = gitopsTestRunner.CreateApplyCapitemplates(50, "capi-server-v1-capitemplate.yaml")

				pages.NavigateToPage(webDriver, "Templates")
				By("And wait for Templates page to be fully rendered", func() {
					templatesPage := pages.GetTemplatesPage(webDriver)
					templatesPage.WaitForPageToLoad(webDriver)
				})

				By("And I should choose a template - grid view", func() {
					templateTile := pages.GetTemplateTile(webDriver, "cluster-template-9")

					Eventually(templateTile.Description).Should(MatchText("This is test template 9"))
					Expect(templateTile.CreateTemplate).Should(BeFound())
					Expect(templateTile.CreateTemplate.Click()).To(Succeed())
				})

				By("And wait for Create cluster page to be fully rendered - grid view", func() {
					createPage := pages.GetCreateClusterPage(webDriver)
					Eventually(createPage.CreateHeader).Should(MatchText(".*Create new cluster.*"))
				})

				By("And I should change the templates view to 'table'", func() {
					pages.NavigateToPage(webDriver, "Templates")
					templatesPage := pages.GetTemplatesPage(webDriver)
					templatesPage.WaitForPageToLoad(webDriver)

					Expect(templatesPage.SelectView("table").Click()).To(Succeed())
				})

				By("And I should choose a template - table view", func() {

					templateRow := pages.GetTemplateRow(webDriver, "cluster-template-10")
					Eventually(templateRow.Provider).Should(MatchText(""))
					Eventually(templateRow.Description).Should(MatchText("This is test template 10"))
					Expect(templateRow.CreateTemplate).Should(BeFound())
					Expect(templateRow.CreateTemplate.Click()).To(Succeed())
				})

				By("And wait for Create cluster page to be fully rendered - table view", func() {
					createPage := pages.GetCreateClusterPage(webDriver)
					Eventually(createPage.CreateHeader).Should(MatchText(".*Create new cluster.*"))
				})
			})
		})

		Context("[UI] When only invalid Capi Template(s) are available in the cluster", func() {
			It("Verify UI shows message related to an invalid template(s)", func() {

				By("Apply/Install invalid CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-invalid-capitemplate.yaml")
				})

				pages.NavigateToPage(webDriver, "Templates")
				By("And wait for Templates page to be fully rendered", func() {
					templatesPage := pages.GetTemplatesPage(webDriver)
					templatesPage.WaitForPageToLoad(webDriver)
				})

				By("And User should see message informing user of the invalid template in the cluster - grid view", func() {
					templateTile := pages.GetTemplateTile(webDriver, "cluster-invalid-template-0")
					Eventually(templateTile.ErrorHeader).Should(BeFound())
					Expect(templateTile.ErrorDescription).Should(BeFound())
					Expect(templateTile.CreateTemplate).ShouldNot(BeEnabled())
				})

				By("And I should change the templates view to 'table'", func() {
					templatesPage := pages.GetTemplatesPage(webDriver)
					Expect(templatesPage.SelectView("table").Click()).To(Succeed())
				})

				By("And User should see message informing user of the invalid template in the cluster - table view", func() {
					templateRow := pages.GetTemplateRow(webDriver, "cluster-invalid-template-0")
					Eventually(templateRow.Provider).Should(MatchText(""))
					Eventually(templateRow.Description).Should(MatchText("Couldn't load template body"))
					Expect(templateRow.CreateTemplate).ShouldNot(BeEnabled())
				})
			})
		})

		Context("[UI] When both valid and invalid Capi Templates are available in the cluster", func() {
			It("Verify UI shows message related to an invalid template(s) and renders the available valid template(s)", func() {

				noOfValidTemplates := 3
				By("Apply/Install valid CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(noOfValidTemplates, "capi-server-v1-template-eks-fargate.yaml")
				})

				noOfInvalidTemplates := 1
				By("Apply/Install invalid CAPITemplate", func() {
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(noOfInvalidTemplates, "capi-server-v1-invalid-capitemplate.yaml")...)
				})

				pages.NavigateToPage(webDriver, "Templates")
				By("And wait for Templates page to be fully rendered", func() {
					templatesPage := pages.GetTemplatesPage(webDriver)
					templatesPage.WaitForPageToLoad(webDriver)
					Eventually(templatesPage.TemplateHeader).Should(BeVisible())

					count, _ := templatesPage.TemplateCount.Text()
					templateCount, _ := strconv.Atoi(count)
					tileCount, _ := templatesPage.TemplateTiles.Count()

					Eventually(templateCount).Should(Equal(noOfValidTemplates+noOfInvalidTemplates), "The template header count should be equal to templates created")
					Eventually(tileCount).Should(Equal(noOfValidTemplates+noOfInvalidTemplates), "The number of template tiles rendered should be equal to number of templates created")
				})

				By("And User should see message informing user of the invalid template in the cluster", func() {
					templateTile := pages.GetTemplateTile(webDriver, "cluster-invalid-template-0")
					Eventually(templateTile.ErrorHeader).Should(BeFound())
					Expect(templateTile.ErrorDescription).Should(BeFound())
					Expect(templateTile.CreateTemplate).ShouldNot(BeEnabled())
				})
			})
		})

		Context("[UI] When Capi Template is available in the cluster", func() {
			It("Verify template parameters should be rendered dynamically and can be set for the selected template", func() {

				By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-eks-fargate.yaml")
				})

				pages.NavigateToPage(webDriver, "Templates")
				By("And wait for Templates page to be fully rendered", func() {
					templatesPage := pages.GetTemplatesPage(webDriver)
					templatesPage.WaitForPageToLoad(webDriver)
				})

				By("And I should change the templates view to 'table'", func() {
					templatesPage := pages.GetTemplatesPage(webDriver)
					Expect(templatesPage.SelectView("table").Click()).To(Succeed())
				})

				By("And I should choose a template - table view", func() {
					templateRow := pages.GetTemplateRow(webDriver, "eks-fargate-template-0")
					Expect(templateRow.CreateTemplate.Click()).To(Succeed())
				})

				createPage := pages.GetCreateClusterPage(webDriver)
				By("And wait for Create cluster page to be fully rendered", func() {
					createPage.WaitForPageToLoad(webDriver)
					Eventually(createPage.CreateHeader).Should(MatchText(".*Create new cluster.*"))
				})

				clusterName := "my-eks-cluster"
				region := "east"
				sshKey := "abcdef1234567890"
				k8Version := "1.19.7"
				paramSection := make(map[string][]TemplateField)
				paramSection["1.Cluster"] = []TemplateField{
					{
						Name:   "CLUSTER_NAME",
						Value:  clusterName,
						Option: "",
					},
				}
				paramSection["3.AWSManagedControlPlane"] = []TemplateField{
					{
						Name:   "AWS_REGION",
						Value:  region,
						Option: "",
					},
					{
						Name:   "AWS_SSH_KEY_NAME",
						Value:  sshKey,
						Option: "",
					},
					{
						Name:   "KUBERNETES_VERSION",
						Value:  k8Version,
						Option: "",
					},
				}

				setParameterValues(createPage, paramSection)

				By("Then I should preview the PR", func() {
					Expect(createPage.PreviewPR.Click()).To(Succeed())
					preview := pages.GetPreview(webDriver)
					pages.WaitForDynamicSecToAppear(webDriver)

					Eventually(preview.PreviewLabel).Should(BeFound())
					pages.ScrollWindow(webDriver, 0, 500)

					Eventually(preview.PreviewText).Should(MatchText(fmt.Sprintf(`kind: Cluster[\s\w\d./:-]*name: %[1]v\s+spec:[\s\w\d./:-]*controlPlaneRef:[\s\w\d./:-]*name: %[1]v-control-plane\s+infrastructureRef:[\s\w\d./:-]*kind: AWSManagedCluster\s+name: %[1]v`, clusterName)))
					Eventually(preview.PreviewText).Should((MatchText(fmt.Sprintf(`kind: AWSManagedCluster\s+metadata:\s+name: %[1]v`, clusterName))))
					Eventually(preview.PreviewText).Should((MatchText(fmt.Sprintf(`kind: AWSManagedControlPlane\s+metadata:\s+name: %[1]v-control-plane\s+spec:\s+region: %[2]v\s+sshKeyName: %[3]v\s+version: %[4]v`, clusterName, region, sshKey, k8Version))))
					Eventually(preview.PreviewText).Should((MatchText(fmt.Sprintf(`kind: AWSFargateProfile\s+metadata:\s+name: %[1]v-fargate-0`, clusterName))))
				})
			})
		})

		Context("[UI] When Capi Template is available in the cluster", func() {
			It("@integration Verify pull request can be created for capi template to the management cluster", func() {

				defer gitopsTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
				defer func() {
					_ = deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})
				}()

				By("And template repo does not already exist", func() {
					gitopsTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
					_ = deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})
				})

				var repoAbsolutePath string
				By("When I create a private repository for cluster configs", func() {
					repoAbsolutePath = gitopsTestRunner.InitAndCreateEmptyRepo(CLUSTER_REPOSITORY, true)
					testFile := createTestFile("README.md", "# gitops-capi-template")

					gitopsTestRunner.GitAddCommitPush(repoAbsolutePath, testFile)
				})

				By("And repo created has private visibility", func() {
					Expect(gitopsTestRunner.GetRepoVisibility(GITHUB_ORG, CLUSTER_REPOSITORY)).Should(ContainSubstring("true"))
				})

				By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
				})

				pages.NavigateToPage(webDriver, "Templates")
				By("And wait for Templates page to be fully rendered", func() {
					templatesPage := pages.GetTemplatesPage(webDriver)
					templatesPage.WaitForPageToLoad(webDriver)
				})

				By("And User should choose a template", func() {
					templateTile := pages.GetTemplateTile(webDriver, "cluster-template-development-0")
					Expect(templateTile.CreateTemplate.Click()).To(Succeed())
				})

				createPage := pages.GetCreateClusterPage(webDriver)
				By("And wait for Create cluster page to be fully rendered", func() {
					createPage.WaitForPageToLoad(webDriver)
					Eventually(createPage.CreateHeader).Should(MatchText(".*Create new cluster.*"))
				})

				// Parameter values
				clusterName := "quick-capd-cluster"
				namespace := "quick-capi"
				k8Version := "1.19.7"

				paramSection := make(map[string][]TemplateField)
				paramSection["7.MachineDeployment"] = []TemplateField{
					{
						Name:   "CLUSTER_NAME",
						Value:  clusterName,
						Option: "",
					},
					{
						Name:   "KUBERNETES_VERSION",
						Value:  k8Version,
						Option: "1.19.8",
					},
					{
						Name:   "NAMESPACE",
						Value:  namespace,
						Option: "",
					},
				}

				setParameterValues(createPage, paramSection)

				By("And press the Preview PR button", func() {
					Expect(createPage.PreviewPR.Click()).To(Succeed())
				})

				//Pull request values
				prBranch := "feature-capd"
				prTitle := "My first pull request"
				prCommit := "First capd capi template"

				By("And set GitOps values for pull request", func() {
					gitops := pages.GetGitOps(webDriver)
					pages.WaitForDynamicSecToAppear(webDriver)
					Eventually(gitops.GitOpsLabel).Should(BeFound())

					pages.ScrollWindow(webDriver, 0, 4000)

					Expect(gitops.GitOpsFields[0].Label).Should(BeFound())
					Expect(gitops.GitOpsFields[0].Field.SendKeys(prBranch)).To(Succeed())
					Expect(gitops.GitOpsFields[1].Label).Should(BeFound())
					Expect(gitops.GitOpsFields[1].Field.SendKeys(prTitle)).To(Succeed())
					Expect(gitops.GitOpsFields[2].Label).Should(BeFound())
					Expect(gitops.GitOpsFields[2].Field.SendKeys(prCommit)).To(Succeed())

					Expect(gitops.CreatePR.Click()).To(Succeed())
				})

				var prUrl string
				clustersPage := pages.GetClustersPage(webDriver)
				By("Then I should see cluster appears in the cluster dashboard with the expected status", func() {
					clusterInfo := pages.FindClusterInList(clustersPage, clusterName)
					Eventually(clusterInfo.Status).Should(HaveText("Creation PR"))
					anchor := clusterInfo.Status.Find("a")
					Eventually(anchor).Should(BeFound())
					prUrl, _ = anchor.Attribute("href")
				})

				By("And I should veriyfy the pull request in the cluster config repository", func() {
					pullRequest := gitopsTestRunner.ListPullRequest(repoAbsolutePath)
					Expect(pullRequest[0]).Should(Equal(prTitle))
					Expect(pullRequest[1]).Should(Equal(prBranch))
					Expect(strings.TrimSuffix(pullRequest[2], "\n")).Should(Equal(prUrl))
				})

				By("And the manifests are present in the cluster config repository", func() {
					gitopsTestRunner.PullBranch(repoAbsolutePath, prBranch)
					_, err := os.Stat(fmt.Sprintf("%s/management/%s.yaml", repoAbsolutePath, clusterName))
					Expect(err).ShouldNot(HaveOccurred(), "Cluster config can not be found.")
				})
			})
		})

		Context("[UI] When Capi Template is available in the cluster", func() {
			It("@integration Verify pull request can not be created by using exiting repository branch", func() {

				defer gitopsTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
				defer func() {
					_ = deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})
				}()

				By("And template repo does not already exist", func() {
					gitopsTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
					_ = deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})
				})

				var repoAbsolutePath string
				By("When I create a private repository for cluster configs", func() {
					repoAbsolutePath = gitopsTestRunner.InitAndCreateEmptyRepo(CLUSTER_REPOSITORY, true)
					testFile := createTestFile("README.md", "# gitops-capi-template")

					gitopsTestRunner.GitAddCommitPush(repoAbsolutePath, testFile)
				})

				branchName := "test-branch"
				By("And create new git repository branch", func() {
					gitopsTestRunner.CreateGitRepoBranch(repoAbsolutePath, branchName)
				})

				By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
				})

				pages.NavigateToPage(webDriver, "Templates")
				By("And wait for Templates page to be fully rendered", func() {
					templatesPage := pages.GetTemplatesPage(webDriver)
					templatesPage.WaitForPageToLoad(webDriver)
				})

				By("And User should choose a template", func() {
					templateTile := pages.GetTemplateTile(webDriver, "cluster-template-development-0")
					Expect(templateTile.CreateTemplate.Click()).To(Succeed())
				})

				createPage := pages.GetCreateClusterPage(webDriver)
				By("And wait for Create cluster page to be fully rendered", func() {
					createPage.WaitForPageToLoad(webDriver)
					Eventually(createPage.CreateHeader).Should(MatchText(".*Create new cluster.*"))
				})

				// Parameter values
				clusterName := "quick-capd-cluster2"
				namespace := "quick-capi"
				k8Version := "1.19.7"

				paramSection := make(map[string][]TemplateField)
				paramSection["7.MachineDeployment"] = []TemplateField{
					{
						Name:   "CLUSTER_NAME",
						Value:  clusterName,
						Option: "",
					},
					{
						Name:   "KUBERNETES_VERSION",
						Value:  k8Version,
						Option: "1.19.8",
					},
					{
						Name:   "NAMESPACE",
						Value:  namespace,
						Option: "",
					},
				}

				setParameterValues(createPage, paramSection)

				By("And press the Preview PR button", func() {
					Expect(createPage.PreviewPR.Click()).To(Succeed())
				})

				//Pull request values
				prTitle := "My first pull request"
				prCommit := "First capd capi template"

				gitops := pages.GetGitOps(webDriver)
				By("And set GitOps values for pull request", func() {
					pages.WaitForDynamicSecToAppear(webDriver)
					Eventually(gitops.GitOpsLabel).Should(BeFound())

					pages.ScrollWindow(webDriver, 0, 4000)

					Expect(gitops.GitOpsFields[0].Label).Should(BeFound())
					Expect(gitops.GitOpsFields[0].Field.SendKeys(branchName)).To(Succeed())
					Expect(gitops.GitOpsFields[1].Label).Should(BeFound())
					Expect(gitops.GitOpsFields[1].Field.SendKeys(prTitle)).To(Succeed())
					Expect(gitops.GitOpsFields[2].Label).Should(BeFound())
					Expect(gitops.GitOpsFields[2].Field.SendKeys(prCommit)).To(Succeed())

					Expect(gitops.CreatePR.Click()).To(Succeed())
				})

				By("Then I should not see pull request to be created", func() {
					Eventually(gitops.ErrorBar).Should(MatchText(fmt.Sprintf(`unable to create pull request.+unable to create new branch "%s"`, branchName)))
				})
			})
		})

		Context("[UI] When no infrastructure provider credentials are available in the management cluster", func() {
			It("@integration Verify no credentials exists in management cluster", func() {
				By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
				})

				pages.NavigateToPage(webDriver, "Templates")
				By("And wait for Templates page to be fully rendered", func() {
					templatesPage := pages.GetTemplatesPage(webDriver)
					templatesPage.WaitForPageToLoad(webDriver)
				})

				By("And User should choose a template", func() {
					templateTile := pages.GetTemplateTile(webDriver, "cluster-template-development-0")
					Expect(templateTile.CreateTemplate.Click()).To(Succeed())
				})

				createPage := pages.GetCreateClusterPage(webDriver)
				By("And wait for Create cluster page to be fully rendered", func() {
					createPage.WaitForPageToLoad(webDriver)
					Eventually(createPage.CreateHeader).Should(MatchText(".*Create new cluster.*"))
				})

				By("Then no infrastructure provider identity can be selected", func() {

					Expect(createPage.Credentials.Click()).To(Succeed())
					Expect(pages.GetCredentials(webDriver).Count()).Should(Equal(1), "Credentials count in the cluster should be '0'' excluding 'None'")

					Expect(pages.GetCredential(webDriver, "None").Click()).To(Succeed())

				})
			})
		})

		Context("[UI] When infrastructure provider credentials are available in the management cluster", func() {
			It("@integration Verify matching selected credential can be used for cluster creation", func() {
				defer gitopsTestRunner.DeleteIPCredentials("AWS")
				defer gitopsTestRunner.DeleteIPCredentials("AZURE")

				By("Apply/Install CAPITemplates", func() {
					eksTemplateFile := gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-aws.yaml")
					azureTemplateFiles := gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-azure.yaml")
					templateFiles = append(azureTemplateFiles, eksTemplateFile...)
				})

				By("And create infrastructure provider credentials)", func() {
					gitopsTestRunner.CreateIPCredentials("AWS")
					gitopsTestRunner.CreateIPCredentials("AZURE")
				})

				pages.NavigateToPage(webDriver, "Templates")
				By("And wait for Templates page to be fully rendered", func() {
					templatesPage := pages.GetTemplatesPage(webDriver)
					templatesPage.WaitForPageToLoad(webDriver)
				})

				By("And User should choose a template", func() {
					templateTile := pages.GetTemplateTile(webDriver, "aws-cluster-template-0")
					Expect(templateTile.CreateTemplate.Click()).To(Succeed())
				})

				createPage := pages.GetCreateClusterPage(webDriver)
				By("And wait for Create cluster page to be fully rendered", func() {
					createPage.WaitForPageToLoad(webDriver)
					Eventually(createPage.CreateHeader).Should(MatchText(".*Create new cluster.*"))
				})

				By("Then AWS test-role-identity credential can be selected", func() {

					Expect(createPage.Credentials.Click()).To(Succeed())
					// FIXME - credentials may or may no be filtered
					// Expect(pages.GetCredentials(webDriver).Count()).Should(Equal(4), "Credentials count in the cluster should be '3' excluding 'None")
					Expect(pages.GetCredential(webDriver, "test-role-identity").Click()).To(Succeed())
				})

				// AWS template parameter values
				awsClusterName := "my-aws-cluster"
				awsRegion := "eu-west-3"
				awsK8version := "1.19.8"
				awsSshKeyName := "my-aws-ssh-key"
				awsNamespace := "default"
				awsControlMAchineType := "t4g.large"
				awsNodeMAchineType := "t3.micro"

				paramSection := make(map[string][]TemplateField)
				paramSection["2.AWSCluster"] = []TemplateField{
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
						Name:   "CLUSTER_NAME",
						Value:  awsClusterName,
						Option: "",
					},
					{
						Name:   "NAMESPACE",
						Value:  awsNamespace,
						Option: "",
					},
				}

				paramSection["3.KubeadmControlPlane"] = []TemplateField{
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
				}

				paramSection["4.AWSMachineTemplate"] = []TemplateField{
					{
						Name:   "AWS_CONTROL_PLANE_MACHINE_TYPE",
						Value:  awsControlMAchineType,
						Option: "",
					},
				}

				paramSection["5.MachineDeployment"] = []TemplateField{
					{
						Name:   "WORKER_MACHINE_COUNT",
						Value:  "3",
						Option: "",
					},
				}

				paramSection["6.AWSMachineTemplate"] = []TemplateField{
					{
						Name:   "AWS_NODE_MACHINE_TYPE",
						Value:  awsNodeMAchineType,
						Option: "",
					},
				}

				setParameterValues(createPage, paramSection)

				By("Then I should see PR preview containing identity reference added in the template", func() {
					Expect(createPage.PreviewPR.Click()).To(Succeed())
					preview := pages.GetPreview(webDriver)
					pages.WaitForDynamicSecToAppear(webDriver)

					Eventually(preview.PreviewLabel).Should(BeFound())
					pages.ScrollWindow(webDriver, 0, 500)

					Eventually(preview.PreviewText).Should(MatchText(fmt.Sprintf(`kind: AWSCluster\s+metadata:\s+name: %s[\s\w\d-.:/]+identityRef:[\s\w\d-.:/]+kind: AWSClusterRoleIdentity\s+name: test-role-identity`, awsClusterName)))
				})

			})
		})

		Context("[UI] When infrastructure provider credentials are available in the management cluster", func() {
			It("@integration Verify user can not use wrong credentials for infrastructure provider", func() {
				defer gitopsTestRunner.DeleteIPCredentials("AWS")

				By("Apply/Install CAPITemplates", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-azure.yaml")
				})

				By("And create infrastructure provider credentials)", func() {
					gitopsTestRunner.CreateIPCredentials("AWS")
				})

				pages.NavigateToPage(webDriver, "Templates")
				By("And wait for Templates page to be fully rendered", func() {
					templatesPage := pages.GetTemplatesPage(webDriver)
					templatesPage.WaitForPageToLoad(webDriver)
				})

				By("And User should choose a template", func() {
					templateTile := pages.GetTemplateTile(webDriver, "azure-capi-quickstart-template-0")
					Expect(templateTile.CreateTemplate.Click()).To(Succeed())
				})

				createPage := pages.GetCreateClusterPage(webDriver)
				By("And wait for Create cluster page to be fully rendered", func() {
					createPage.WaitForPageToLoad(webDriver)
					Eventually(createPage.CreateHeader).Should(MatchText(".*Create new cluster.*"))
				})

				By("Then AWS aws-test-identity credential can be selected", func() {

					Expect(createPage.Credentials.Click()).To(Succeed())
					// FIXME - credentials may or may no be filtered
					Expect(pages.GetCredential(webDriver, "test-role-identity").Click()).To(Succeed())
				})

				// Azure template parameter values
				azureClusterName := "my-azure-cluster"
				azureK8version := "1.19.7"
				azureNamespace := "default"
				azureControlMAchineType := "HBv2"
				azureNodeMAchineType := "Dasv4"

				paramSection := make(map[string][]TemplateField)
				paramSection["2.AzureCluster"] = []TemplateField{
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
				}

				paramSection["3.KubeadmControlPlane"] = []TemplateField{
					{
						Name:   "CONTROL_PLANE_MACHINE_COUNT",
						Value:  "2",
						Option: "",
					},
					{
						Name:   "KUBERNETES_VERSION",
						Value:  azureK8version,
						Option: "",
					},
				}

				paramSection["4.AzureMachineTemplate"] = []TemplateField{
					{
						Name:   "AZURE_CONTROL_PLANE_MACHINE_TYPE",
						Value:  azureControlMAchineType,
						Option: "",
					},
				}

				paramSection["5.MachineDeployment"] = []TemplateField{
					{
						Name:   "WORKER_MACHINE_COUNT",
						Value:  "3",
						Option: "",
					},
				}

				paramSection["6.AzureMachineTemplate"] = []TemplateField{
					{
						Name:   "AZURE_NODE_MACHINE_TYPE",
						Value:  azureNodeMAchineType,
						Option: "",
					},
				}

				setParameterValues(createPage, paramSection)

				By("Then I should see PR preview without identity reference added to the template", func() {
					Expect(createPage.PreviewPR.Click()).To(Succeed())
					preview := pages.GetPreview(webDriver)
					pages.WaitForDynamicSecToAppear(webDriver)

					Eventually(preview.PreviewLabel).Should(BeFound())
					pages.ScrollWindow(webDriver, 0, 500)

					Eventually(preview.PreviewText).ShouldNot(MatchText(`kind: AWSCluster[\s\w\d-.:/]+identityRef:`), "Identity reference should not be found in preview pull request AzureCluster object")
				})

			})
		})

		Context("[UI] When leaf cluster pull request is available in the management cluster", func() {
			kubeconfigPath := path.Join(os.Getenv("HOME"), "Downloads", "kubeconfig")
			appName := "management"
			capdClusterName := "ui-end-to-end-capd-cluster"

			JustBeforeEach(func() {
				_ = deleteFile([]string{kubeconfigPath})

				log.Println("Connecting cluster to itself")
				leaf := LeafSpec{
					Status:          "Ready",
					IsWKP:           false,
					AlertManagerURL: "",
					KubeconfigPath:  "",
				}
				connectACluster(webDriver, gitopsTestRunner, leaf)
			})

			JustAfterEach(func() {
				_ = deleteFile([]string{kubeconfigPath})
				removeGitopsCapiClusters(appName, []string{capdClusterName}, GITOPS_DEFAULT_NAMESPACE)

				gitopsTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
				_ = deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})

				log.Println("Deleting all the wkp agents")
				_ = gitopsTestRunner.KubectlDeleteAllAgents([]string{})
				_ = gitopsTestRunner.ResetDatabase()
				gitopsTestRunner.VerifyWegoPodsRunning()
			})

			It("@smoke @integration @capd Verify leaf CAPD cluster can be provisioned and kubeconfig is available for cluster operations", func() {

				By("And template repo does not already exist", func() {
					gitopsTestRunner.DeleteRepo(CLUSTER_REPOSITORY)
					_ = deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})
				})

				var repoAbsolutePath string
				By("When I create a private repository for cluster configs", func() {
					repoAbsolutePath = gitopsTestRunner.InitAndCreateEmptyRepo(CLUSTER_REPOSITORY, true)
					testFile := createTestFile("README.md", "# gitops-capi-template")

					gitopsTestRunner.GitAddCommitPush(repoAbsolutePath, testFile)
				})

				By("And I install gitops to my active cluster", func() {
					Expect(FileExists(GITOPS_BIN_PATH)).To(BeTrue(), fmt.Sprintf("%s can not be found.", GITOPS_BIN_PATH))
					installAndVerifyGitops(GITOPS_DEFAULT_NAMESPACE)
				})

				addCommand := fmt.Sprintf("app add . --path=./management  --name=%s  --auto-merge=true", appName)
				By(fmt.Sprintf("And I run gitops app add command '%s in namespace %s from dir %s'", addCommand, GITOPS_DEFAULT_NAMESPACE, repoAbsolutePath), func() {
					command := exec.Command("sh", "-c", fmt.Sprintf("cd %s && %s %s", repoAbsolutePath, GITOPS_BIN_PATH, addCommand))
					session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
					Eventually(session).Should(gexec.Exit())
				})

				By("And I install Docker provider infrastructure", func() {
					installInfrastructureProvider("docker")
				})

				By("Then I Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
				})

				pages.NavigateToPage(webDriver, "Templates")
				By("And wait for Templates page to be fully rendered", func() {
					templatesPage := pages.GetTemplatesPage(webDriver)
					templatesPage.WaitForPageToLoad(webDriver)
				})

				By("And User should choose a template", func() {
					templateTile := pages.GetTemplateTile(webDriver, "cluster-template-development-0")
					Expect(templateTile.CreateTemplate.Click()).To(Succeed())
				})

				createPage := pages.GetCreateClusterPage(webDriver)
				By("And wait for Create cluster page to be fully rendered", func() {
					createPage.WaitForPageToLoad(webDriver)
					Eventually(createPage.CreateHeader).Should(MatchText(".*Create new cluster.*"))
				})

				// // Parameter values
				clusterName := capdClusterName
				namespace := "default"
				k8Version := "1.19.7"

				paramSection := make(map[string][]TemplateField)
				paramSection["7.MachineDeployment"] = []TemplateField{
					{
						Name:   "CLUSTER_NAME",
						Value:  clusterName,
						Option: "",
					},
					{
						Name:   "KUBERNETES_VERSION",
						Value:  k8Version,
						Option: "1.19.8",
					},
					{
						Name:   "NAMESPACE",
						Value:  namespace,
						Option: "",
					},
				}

				setParameterValues(createPage, paramSection)

				By("And press the Preview PR button", func() {
					Expect(createPage.PreviewPR.Click()).To(Succeed())
				})

				//Pull request values
				prBranch := "ui-end-end-branch"
				prTitle := "CAPD pull request"
				prCommit := "CAPD capi template"

				By("And set GitOps values for pull request", func() {
					gitops := pages.GetGitOps(webDriver)
					pages.WaitForDynamicSecToAppear(webDriver)
					Eventually(gitops.GitOpsLabel).Should(BeFound())

					pages.ScrollWindow(webDriver, 0, 4000)

					Expect(gitops.GitOpsFields[0].Label).Should(BeFound())
					Expect(gitops.GitOpsFields[0].Field.SendKeys(prBranch)).To(Succeed())
					Expect(gitops.GitOpsFields[1].Label).Should(BeFound())
					Expect(gitops.GitOpsFields[1].Field.SendKeys(prTitle)).To(Succeed())
					Expect(gitops.GitOpsFields[2].Label).Should(BeFound())
					Expect(gitops.GitOpsFields[2].Field.SendKeys(prCommit)).To(Succeed())

					Expect(gitops.CreatePR.Click()).To(Succeed())
				})

				clustersPage := pages.GetClustersPage(webDriver)
				By("Then I should see cluster appears in the cluster dashboard with 'Creation PR' status", func() {
					clusterInfo := pages.FindClusterInList(clustersPage, clusterName)
					Eventually(clusterInfo.Status).Should(HaveText("Creation PR"))
				})

				By("Then I should merge the pull request to start cluster provisioning", func() {
					gitopsTestRunner.MergePullRequest(repoAbsolutePath, prBranch)
				})

				By("Then I should see cluster status changes to 'Cluster found'", func() {
					Eventually(pages.FindClusterInList(clustersPage, clusterName).Status, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_15SECONDS).Should(HaveText("Cluster found"))
				})

				By("And I should download the kubeconfig for the CAPD capi cluster", func() {
					clusterInfo := pages.FindClusterInList(clustersPage, clusterName)
					Expect(clusterInfo.Status.Click()).To(Succeed())
					clusterStatus := pages.GetClusterStatus(webDriver)
					Eventually(clusterStatus.Phase, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_15SECONDS).Should(HaveText(`"Provisioned"`))

					fileErr := func() error {
						Expect(clusterStatus.KubeConfigButton.Click()).To(Succeed())
						_, err := os.Stat(kubeconfigPath)
						return err

					}
					Eventually(fileErr, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_15SECONDS).ShouldNot(HaveOccurred())
				})

				By("And verify the kubeconfig is correct", func() {
					contents, err := ioutil.ReadFile(kubeconfigPath)
					Expect(err).ShouldNot(HaveOccurred())
					Eventually(contents).Should(MatchRegexp(fmt.Sprintf(`context:\s+cluster: %s`, clusterName)))
				})

				By("Then I should select the cluster to create the delete pull request", func() {
					clusterInfo := pages.FindClusterInList(clustersPage, clusterName)
					Expect(clusterInfo.Checkbox.Click()).To(Succeed())

					Eventually(webDriver.FindByXPath(`//button[@id="delete-cluster"][@disabled]`)).ShouldNot(BeFound())
					Expect(clustersPage.PRDeleteClusterButton.Click()).To(Succeed())

					deletePR := pages.GetDeletePRPopup(webDriver)
					Expect(deletePR.PRDescription.SendKeys("Delete CAPD capi cluster, it is not required any more")).To(Succeed())
					Expect(deletePR.DeleteClusterButton.Click()).To(Succeed())
				})

				var deletePRbranch string
				var deletePRUrl string
				By("And I should veriyfy the delete pull request in the cluster config repository", func() {
					clusterInfo := pages.FindClusterInList(clustersPage, clusterName)

					var pullRequest []string
					pr := func() []string {
						pullRequest = gitopsTestRunner.ListPullRequest(repoAbsolutePath)
						return pullRequest
					}
					Eventually(pr).Should(HaveLen(3))

					deletePRbranch = pullRequest[1]
					deletePRUrl = strings.TrimSuffix(pullRequest[2], "\n")
					Eventually(clusterInfo.Status.Find(`a`)).Should(BeFound())
					Expect(clusterInfo.Status.Find(`a`).Attribute("href")).Should(MatchRegexp(deletePRUrl))
				})

				By("And the delete pull request manifests are not present in the cluster config repository", func() {
					gitopsTestRunner.PullBranch(repoAbsolutePath, deletePRbranch)
					_, err := os.Stat(fmt.Sprintf("%s/management/%s.yaml", repoAbsolutePath, clusterName))
					Expect(err).Should(MatchError(os.ErrNotExist), "Cluster config is found when expected to be deleted.")
				})

				By(fmt.Sprintf("Then I should see the '%s' cluster status changes to Deletion PR", clusterName), func() {
					clusterInfo := pages.FindClusterInList(clustersPage, clusterName)
					Eventually(clusterInfo.Status).Should(HaveText("Deletion PR"))
				})

				// By("Then I should merge the delete pull request to delete cluster", func() {
				// 	gitopsTestRunner.MergePullRequest(repoAbsolutePath, deletePRbranch)
				// })

			})
		})

		Context("[UI] When entitlement is available in the cluster", func() {
			DEPLOYMENT_APP := "my-mccp-cluster-service"

			checkEntitlement := func(typeEntitelment string, beFound bool) {
				checkOutput := func() bool {
					found, _ := pages.GetEntitelment(webDriver, typeEntitelment).Visible()
					Expect(webDriver.Refresh()).ShouldNot(HaveOccurred())
					return found

				}

				Expect(webDriver.Refresh()).ShouldNot(HaveOccurred())
				if beFound {
					Eventually(checkOutput, ASSERTION_DEFAULT_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(BeTrue())
				} else {
					Eventually(checkOutput, ASSERTION_DEFAULT_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(BeFalse())
				}

			}

			JustAfterEach(func() {
				By("When I apply the valid entitlement", func() {
					Expect(gitopsTestRunner.KubectlApply([]string{}, "../../utils/scripts/entitlement-secret.yaml"), "Failed to create/configure entitlement")
				})

				By("Then I restart the cluster service pod for valid entitlemnt to take effect", func() {
					Expect(gitopsTestRunner.RestartDeploymentPods([]string{}, DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
				})

				By("And I should not see the error or warning message for valid entitlement", func() {
					checkEntitlement("expired", false)
					checkEntitlement("missing", false)
				})
			})

			It("@integration Verify cluster service acknowledges the entitlement presences", func() {

				By("When I delete the entitlement", func() {
					Expect(gitopsTestRunner.KubectlDelete([]string{}, "../../utils/scripts/entitlement-secret.yaml"), "Failed to delete entitlement secret")
				})

				By("Then I restart the cluster service pod for missing entitlemnt to take effect", func() {
					Expect(gitopsTestRunner.RestartDeploymentPods([]string{}, DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE)).ShouldNot(HaveOccurred(), "Failed restart deployment successfully")
				})

				By("And I should see the error message for missing entitlement", func() {
					checkEntitlement("missing", true)
				})

				By("When I apply the expired entitlement", func() {
					Expect(gitopsTestRunner.KubectlApply([]string{}, "../../utils/data/entitlement-secret-expired.yaml"), "Failed to create/configure entitlement")
				})

				By("Then I restart the cluster service pod for expired entitlemnt to take effect", func() {
					Expect(gitopsTestRunner.RestartDeploymentPods([]string{}, DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
				})

				By("And I should see the warning message for expired entitlement", func() {
					checkEntitlement("expired", true)
				})

				By("When I apply the invalid entitlement", func() {
					Expect(gitopsTestRunner.KubectlApply([]string{}, "../../utils/data/entitlement-secret-invalid.yaml"), "Failed to create/configure entitlement")
				})

				By("Then I restart the cluster service pod for invalid entitlemnt to take effect", func() {
					Expect(gitopsTestRunner.RestartDeploymentPods([]string{}, DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
				})

				By("And I should see the error message for invalid entitlement", func() {
					checkEntitlement("invalid", true)
				})
			})
		})
	})
}
