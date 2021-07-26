package acceptance

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/wks/test/acceptance/test/pages"
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

func DescribeMCCPTemplates(mccpTestRunner MCCPTestRunner) {

	var _ = Describe("Multi-Cluster Control Plane Templates", func() {

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
				err = webDriver.Size(1800, 2500)
				Expect(err).NotTo(HaveOccurred())
			}

			By("When I navigate to MCCP UI Page", func() {

				Expect(webDriver.Navigate(GetWkpUrl())).To(Succeed())

			})
		})

		AfterEach(func() {
			mccpTestRunner.DeleteApplyCapiTemplates(templateFiles)
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

				noOfTemplates := 5
				templateFiles = mccpTestRunner.CreateApplyCapitemplates(noOfTemplates, "capi-server-v1-capitemplate.yaml")

				pages.NavigateToPage(webDriver, "Templates")
				templatesPage := pages.GetTemplatesPage(webDriver)

				By("And wait for Templates page to be fully rendered", func() {
					templatesPage.WaitForPageToLoad(webDriver)
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

		Context("[UI] When Capi Templates are available in the cluster", func() {
			It("Verify I should be able to select a template of my choice", func() {

				// test selection with 50 capiTemplates
				templateFiles = mccpTestRunner.CreateApplyCapitemplates(50, "capi-server-v1-capitemplate.yaml")

				pages.NavigateToPage(webDriver, "Templates")
				By("And wait for Templates page to be fully rendered", func() {
					templatesPage := pages.GetTemplatesPage(webDriver)
					templatesPage.WaitForPageToLoad(webDriver)
				})

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

		Context("[UI] When only invalid Capi Template(s) are available in the cluster", func() {
			It("Verify UI shows message related to an invalid template(s)", func() {

				By("Apply/Install invalid CAPITemplate", func() {
					templateFiles = mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-invalid-capitemplate.yaml")
				})

				pages.NavigateToPage(webDriver, "Templates")
				By("And wait for Templates page to be fully rendered", func() {
					templatesPage := pages.GetTemplatesPage(webDriver)
					templatesPage.WaitForPageToLoad(webDriver)
				})

				By("And User should see message informing user of the invalid template in the cluster", func() {
					templateTile := pages.GetTemplateTile(webDriver, "cluster-invalid-template-0")
					Eventually(templateTile.ErrorHeader).Should(BeFound())
					Expect(templateTile.ErrorDescription).Should(BeFound())
					Expect(templateTile.CreateTemplate).ShouldNot(BeEnabled())
				})
			})
		})

		Context("[UI] When both valid and invalid Capi Templates are available in the cluster", func() {
			It("Verify UI shows message related to an invalid template(s) and renders the available valid template(s)", func() {

				noOfValidTemplates := 3
				By("Apply/Insall valid CAPITemplate", func() {
					templateFiles = mccpTestRunner.CreateApplyCapitemplates(noOfValidTemplates, "capi-server-v1-template-eks-fargate.yaml")
				})

				noOfInvalidTemplates := 1
				By("Apply/Insall invalid CAPITemplate", func() {
					templateFiles = append(templateFiles, mccpTestRunner.CreateApplyCapitemplates(noOfInvalidTemplates, "capi-server-v1-invalid-capitemplate.yaml")...)
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

				By("Apply/Insall CAPITemplate", func() {
					templateFiles = mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-eks-fargate.yaml")
				})

				pages.NavigateToPage(webDriver, "Templates")
				By("And wait for Templates page to be fully rendered", func() {
					templatesPage := pages.GetTemplatesPage(webDriver)
					templatesPage.WaitForPageToLoad(webDriver)
				})

				By("And User should choose a template", func() {
					templateTile := pages.GetTemplateTile(webDriver, "eks-fargate-template-0")
					Expect(templateTile.CreateTemplate.Click()).To(Succeed())
				})

				createPage := pages.GetCreateClusterPage(webDriver)
				By("And wait for Create cluster page to be fully rendered", func() {
					createPage.WaitForPageToLoad(webDriver)
					Eventually(createPage.CreateHeader).Should(MatchText(".*Create new cluster.*"))
					// Eventually(createPage.TemplateName).Should(MatchText(".*eks-fargate-template-0.*"))
				})

				clusterName := "my-eks-cluster"
				region := "east"
				sshKey := "abcdef1234567890"
				k8Version := "1.19.7"
				paramSection := make(map[string][]TemplateField)
				paramSection["1. Cluster"] = []TemplateField{
					{
						Name:   "CLUSTER_NAME",
						Value:  clusterName,
						Option: "",
					},
				}
				paramSection["3. AWSManagedControlPlane"] = []TemplateField{
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
			It("@Integration Verify pull request can be created for capi template to the management cluster", func() {

				defer mccpTestRunner.deleteRepo(CLUSTER_REPOSITORY)
				defer deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})

				By("And template repo does not already exist", func() {
					mccpTestRunner.deleteRepo(CLUSTER_REPOSITORY)
					deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})
				})

				var repoAbsolutePath string
				By("When I create a private repository for cluster configs", func() {
					repoAbsolutePath = mccpTestRunner.initAndCreateEmptyRepo(CLUSTER_REPOSITORY, true)
					testFile := createTestFile("README.md", "# mccp-capi-template")

					mccpTestRunner.gitAddCommitPush(repoAbsolutePath, testFile)
				})

				By("And repo created has private visibility", func() {
					Expect(mccpTestRunner.getRepoVisibility(GITHUB_ORG, CLUSTER_REPOSITORY)).Should(ContainSubstring("true"))
				})

				By("Apply/Insall CAPITemplate", func() {
					templateFiles = mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
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
				paramSection["7. MachineDeployment"] = []TemplateField{
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
					Eventually(clusterInfo.Status).Should(HaveText("PR Created"))
					prUrl, _ = clusterInfo.Status.Find("a").Attribute("href")
				})

				By("And I should veriyfy the pull request in the cluster config repository", func() {
					pullRequest := mccpTestRunner.listPullRequest(repoAbsolutePath)
					Expect(pullRequest[0]).Should(Equal(prTitle))
					Expect(pullRequest[1]).Should(Equal(prBranch))
					Expect(strings.TrimSuffix(pullRequest[2], "\n")).Should(Equal(prUrl))
				})

				By("And the manifests are present in the cluster config repository", func() {
					mccpTestRunner.pullBranch(repoAbsolutePath, prBranch)
					_, err := os.Stat(fmt.Sprintf("%s/management/%s.yaml", repoAbsolutePath, clusterName))
					Expect(err).ShouldNot(HaveOccurred(), "Cluster config can not be found.")
				})
			})
		})

		Context("[UI] When Capi Template is available in the cluster", func() {
			It("@Integration Verify pull request can not be created by using exiting repository branch", func() {

				defer mccpTestRunner.deleteRepo(CLUSTER_REPOSITORY)
				defer deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})

				By("And template repo does not already exist", func() {
					mccpTestRunner.deleteRepo(CLUSTER_REPOSITORY)
					deleteDirectory([]string{path.Join("/tmp", CLUSTER_REPOSITORY)})
				})

				var repoAbsolutePath string
				By("When I create a private repository for cluster configs", func() {
					repoAbsolutePath = mccpTestRunner.initAndCreateEmptyRepo(CLUSTER_REPOSITORY, true)
					testFile := createTestFile("README.md", "# mccp-capi-template")

					mccpTestRunner.gitAddCommitPush(repoAbsolutePath, testFile)
				})

				branchName := "test-branch"
				By("And create new git repository branch", func() {
					mccpTestRunner.createGitRepoBranch(repoAbsolutePath, branchName)
				})

				By("Apply/Insall CAPITemplate", func() {
					templateFiles = mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
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
				paramSection["7. MachineDeployment"] = []TemplateField{
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
			It("@Integration Verify no credentials exists in mccp", func() {
				By("Apply/Insall CAPITemplate", func() {
					templateFiles = mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-capd.yaml")
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
			It("@Integration Verify matching selected credential can be used for cluster creation", func() {
				defer mccpTestRunner.deleteIPCredentials("AWS")
				defer mccpTestRunner.deleteIPCredentials("AZURE")

				By("Apply/Insall CAPITemplates", func() {
					eksTemplateFile := mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-aws.yaml")
					azureTemplateFiles := mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-azure.yaml")
					templateFiles = append(azureTemplateFiles, eksTemplateFile...)
				})

				By("And create infrastructure provider credentials)", func() {
					mccpTestRunner.createIPCredentials("AWS")
					mccpTestRunner.createIPCredentials("AZURE")
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
				paramSection["2. AWSCluster"] = []TemplateField{
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

				paramSection["3. KubeadmControlPlane"] = []TemplateField{
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

				paramSection["4. AWSMachineTemplate"] = []TemplateField{
					{
						Name:   "AWS_CONTROL_PLANE_MACHINE_TYPE",
						Value:  awsControlMAchineType,
						Option: "",
					},
				}

				paramSection["5. MachineDeployment"] = []TemplateField{
					{
						Name:   "WORKER_MACHINE_COUNT",
						Value:  "3",
						Option: "",
					},
				}

				paramSection["6. AWSMachineTemplate"] = []TemplateField{
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
			It("@Integration Verify user can not use wrong credentials for infrastructure provider", func() {
				defer mccpTestRunner.deleteIPCredentials("AWS")

				By("Apply/Insall CAPITemplates", func() {
					templateFiles = mccpTestRunner.CreateApplyCapitemplates(1, "capi-server-v1-template-azure.yaml")
				})

				By("And create infrastructure provider credentials)", func() {
					mccpTestRunner.createIPCredentials("AWS")
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
				paramSection["2. AzureCluster"] = []TemplateField{
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

				paramSection["3. KubeadmControlPlane"] = []TemplateField{
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

				paramSection["4. AzureMachineTemplate"] = []TemplateField{
					{
						Name:   "AZURE_CONTROL_PLANE_MACHINE_TYPE",
						Value:  azureControlMAchineType,
						Option: "",
					},
				}

				paramSection["5. MachineDeployment"] = []TemplateField{
					{
						Name:   "WORKER_MACHINE_COUNT",
						Value:  "3",
						Option: "",
					},
				}

				paramSection["6. AzureMachineTemplate"] = []TemplateField{
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
	})
}
