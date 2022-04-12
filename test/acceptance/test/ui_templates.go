package acceptance

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
	"k8s.io/apimachinery/pkg/util/wait"
)

type TemplateField struct {
	Name   string
	Value  string
	Option string
}

// Wait until we get a good looking response from /v1/profiles
// Ignore all errors (connection refused, 500s etc)
func waitForProfiles(ctx context.Context, timeout time.Duration) error {
	adminPassword := GetEnv("CLUSTER_ADMIN_PASSWORD", "")
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return wait.PollUntil(time.Second*1, func() (bool, error) {
		jar, _ := cookiejar.New(&cookiejar.Options{})
		client := http.Client{
			Timeout: 5 * time.Second,
			Jar:     jar,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
		// login to fetch cookie
		resp, err := client.Post(test_ui_url+"/oauth2/sign_in", "application/json", bytes.NewReader([]byte(fmt.Sprintf(`{"username":"admin", "password":"%s"}`, adminPassword))))
		if err != nil {
			logger.Tracef("error logging in (waiting for a success, retrying): %v", err)
			return false, nil
		}
		if resp.StatusCode != http.StatusOK {
			logger.Tracef("wrong status from login (waiting for a ok, retrying): %v", resp.StatusCode)
			return false, nil
		}
		// fetch profiles
		resp, err = client.Get(test_ui_url + "/v1/profiles")
		if err != nil {
			logger.Tracef("error getting profiles in (waiting for a success, retrying): %v", err)
			return false, nil
		}
		if resp.StatusCode != http.StatusOK {
			logger.Tracef("wrong status from profiles (waiting for a ok, retrying): %v", resp.StatusCode)
			return false, nil
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, nil
		}
		bodyString := string(bodyBytes)

		return strings.Contains(bodyString, `"profiles":`), nil
	}, waitCtx.Done())
}

func setParameterValues(createPage *pages.CreateCluster, paramSection map[string][]TemplateField) {
	for section, parameters := range paramSection {
		By(fmt.Sprintf("And set template section %s parameter values", section), func() {
			templateSection := createPage.GetTemplateSection(webDriver, section)
			Expect(templateSection.Name).Should(HaveText(section))

			if len(parameters) == 0 {
				Expect(len(templateSection.Fields)).To(BeNumerically("==", 0), fmt.Sprintf("No form fields should be visible for section %s", section))
			}

			for i := 0; i < len(parameters); i++ {
				paramSet := false
				for j := 0; j < len(templateSection.Fields); j++ {
					val, _ := templateSection.Fields[j].Label.Text()
					if strings.Contains(val, parameters[i].Name) {
						By("And set template parameter values", func() {
							if parameters[i].Option != "" {
								if pages.ElementExist(templateSection.Fields[j].ListBox) {
									selectOption := func() bool {
										_ = templateSection.Fields[j].ListBox.Click()
										time.Sleep(POLL_INTERVAL_100MILLISECONDS)
										visible, _ := pages.GetOption(webDriver, parameters[i].Option).Visible()
										return visible
									}
									Eventually(selectOption, ASSERTION_DEFAULT_TIME_OUT).Should(BeTrue(), fmt.Sprintf("Failed to select parameter option '%s' in section '%s' ", parameters[i].Name, section))
									Expect(pages.GetOption(webDriver, parameters[i].Option).Click()).To(Succeed())
									paramSet = true
								}
							} else {
								Expect(templateSection.Fields[j].Field.SendKeys(parameters[i].Value)).To(Succeed())
								paramSet = true
							}
						})
					}
				}
				Expect(paramSet).Should(BeTrue(), fmt.Sprintf("Parameter '%s' isn't found in section '%s' ", parameters[i].Name, section))
			}
		})
	}
}

func selectCredentials(createPage *pages.CreateCluster, credentialName string, credentialCount int) {
	selectCredential := func() bool {
		Eventually(createPage.Credentials.Click).Should(Succeed())
		// Credentials are not filtered for selected template
		if cnt, _ := pages.GetCredentials(webDriver).Count(); cnt > 0 {
			Eventually(pages.GetCredentials(webDriver).Count).Should(Equal(credentialCount), fmt.Sprintf(`Credentials count in the cluster should be '%d' including 'None`, credentialCount))
			Expect(pages.GetCredential(webDriver, credentialName).Click()).To(Succeed())
		}

		credentialText, _ := createPage.Credentials.Text()
		return strings.Contains(credentialText, credentialName)
	}
	Eventually(selectCredential, ASSERTION_30SECONDS_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(BeTrue())
}

func DescribeTemplates(gitopsTestRunner GitopsTestRunner) {
	var _ = Describe("Multi-Cluster Control Plane Templates", func() {
		templateFiles := []string{}

		BeforeEach(func() {
			Expect(webDriver.Navigate(test_ui_url)).To(Succeed())
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
					templateFiles = append(templateFiles, gitopsTestRunner.CreateApplyCapitemplates(5, "capi-template-capd.yaml")...)
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
				paramSection["2.AWSManagedCluster"] = nil
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
				paramSection["4.AWSFargateProfile"] = nil

				setParameterValues(createPage, paramSection)

				By("Then I should preview the PR", func() {
					Expect(createPage.PreviewPR.Click()).To(Succeed())
					preview := pages.GetPreview(webDriver)

					Eventually(preview.Title).Should(MatchText("PR Preview"))

					Eventually(preview.Text).Should(MatchText(fmt.Sprintf(`kind: Cluster[\s\w\d./:-]*name: %[1]v\s+namespace: default\s+spec:[\s\w\d./:-]*controlPlaneRef:[\s\w\d./:-]*name: %[1]v-control-plane\s+infrastructureRef:[\s\w\d./:-]*kind: AWSManagedCluster\s+name: %[1]v`, clusterName)))
					Eventually(preview.Text).Should((MatchText(fmt.Sprintf(`kind: AWSManagedCluster\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v`, clusterName))))
					Eventually(preview.Text).Should((MatchText(fmt.Sprintf(`kind: AWSManagedControlPlane\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v-control-plane\s+namespace: default\s+spec:\s+region: %[2]v\s+sshKeyName: %[3]v\s+version: %[4]v`, clusterName, region, sshKey, k8Version))))
					Eventually(preview.Text).Should((MatchText(fmt.Sprintf(`kind: AWSFargateProfile\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %[1]v-fargate-0`, clusterName))))

					Eventually(preview.Close.Click).Should(Succeed())
				})
			})
		})

		Context("[UI] When Capi Template is available in the cluster", func() {

			JustAfterEach(func() {
				// Force clean the repository directory for subsequent tests
				cleanGitRepository("management")
			})

			It("@integration @git Verify pull request can be created for capi template to the management cluster", func() {
				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)

				By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-template-capd.yaml")
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
				k8Version := "1.22.0"

				paramSection := make(map[string][]TemplateField)
				paramSection["1.Cluster"] = []TemplateField{
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
				}
				paramSection["4.KubeadmControlPlane"] = []TemplateField{
					{
						Name:   "KUBERNETES_VERSION",
						Value:  "",
						Option: k8Version,
					},
				}

				setParameterValues(createPage, paramSection)

				By("Then I should preview the PR", func() {
					Expect(createPage.PreviewPR.Click()).To(Succeed())
					preview := pages.GetPreview(webDriver)
					Eventually(preview.Close.Click).Should(Succeed())
				})

				//Pull request values
				prBranch := "ui-feature-capd"
				prTitle := "My first pull request"
				prCommit := "First capd capi template"

				By("And set GitOps values for pull request", func() {
					gitops := pages.GetGitOps(webDriver)
					Eventually(gitops.GitOpsLabel).Should(BeFound())

					pages.ScrollWindow(webDriver, 0, 4000)

					Expect(gitops.GitOpsFields[0].Label).Should(BeFound())
					Expect(gitops.GitOpsFields[0].Field.SendKeys(prBranch)).To(Succeed())
					Expect(gitops.GitOpsFields[1].Label).Should(BeFound())
					Expect(gitops.GitOpsFields[1].Field.SendKeys(prTitle)).To(Succeed())
					Expect(gitops.GitOpsFields[2].Label).Should(BeFound())
					Expect(gitops.GitOpsFields[2].Field.SendKeys(prCommit)).To(Succeed())

					AuthenticateWithGitProvider(webDriver, gitProviderEnv.Type, gitProviderEnv.Hostname)
					Eventually(gitops.GitCredentials).Should(BeVisible())
					if pages.ElementExist(gitops.ErrorBar) {
						Expect(gitops.ErrorBar.Click()).To(Succeed())
					}

					Eventually(gitops.CreatePR.Click).Should(Succeed())
				})

				var prUrl string
				clustersPage := pages.GetClustersPage(webDriver)
				By("Then I should see cluster appears in the cluster dashboard with 'Creation PR' status", func() {
					clusterInfo := pages.FindClusterInList(clustersPage, clusterName)
					Eventually(clusterInfo.Status, ASSERTION_1MINUTE_TIME_OUT).Should(HaveText("Creation PR"))
					anchor := clusterInfo.Status.Find("a")
					Eventually(anchor).Should(BeFound())
					prUrl, _ = anchor.Attribute("href")
				})

				var createPRUrl string
				By("And I should veriyfy the pull request in the cluster config repository", func() {
					createPRUrl = verifyPRCreated(gitProviderEnv, repoAbsolutePath)
					Expect(createPRUrl).Should(Equal(prUrl))
				})

				By("And the manifests are present in the cluster config repository", func() {
					mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
					pullGitRepo(repoAbsolutePath)
					_, err := os.Stat(fmt.Sprintf("%s/management/%s.yaml", repoAbsolutePath, clusterName))
					Expect(err).ShouldNot(HaveOccurred(), "Cluster config can not be found.")
				})
			})

			It("@integration @git Verify pull request can not be created by using exiting repository branch", func() {
				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)

				branchName := "ui-test-branch"
				By("And create new git repository branch", func() {
					_ = createGitRepoBranch(repoAbsolutePath, branchName)
				})

				By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-template-capd.yaml")
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
				k8Version := "1.22.0"

				paramSection := make(map[string][]TemplateField)
				paramSection["1.Cluster"] = []TemplateField{
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
				}
				paramSection["4.KubeadmControlPlane"] = []TemplateField{
					{
						Name:   "KUBERNETES_VERSION",
						Value:  "",
						Option: k8Version,
					},
				}

				setParameterValues(createPage, paramSection)

				//Pull request values
				prTitle := "My first pull request"
				prCommit := "First capd capi template"

				gitops := pages.GetGitOps(webDriver)
				By("And set GitOps values for pull request", func() {
					Eventually(gitops.GitOpsLabel).Should(BeFound())

					pages.ScrollWindow(webDriver, 0, 4000)

					Expect(gitops.GitOpsFields[0].Label).Should(BeFound())
					Expect(gitops.GitOpsFields[0].Field.SendKeys(branchName)).To(Succeed())
					Expect(gitops.GitOpsFields[1].Label).Should(BeFound())
					Expect(gitops.GitOpsFields[1].Field.SendKeys(prTitle)).To(Succeed())
					Expect(gitops.GitOpsFields[2].Label).Should(BeFound())
					Expect(gitops.GitOpsFields[2].Field.SendKeys(prCommit)).To(Succeed())

					AuthenticateWithGitProvider(webDriver, gitProviderEnv.Type, gitProviderEnv.Hostname)
					Eventually(gitops.GitCredentials).Should(BeVisible())

					if pages.ElementExist(gitops.ErrorBar) {
						Expect(gitops.ErrorBar.Click()).To(Succeed())
					}

					Expect(gitops.CreatePR.Click()).To(Succeed())
				})

				By("Then I should not see pull request to be created", func() {
					Eventually(gitops.ErrorBar).Should(MatchText(fmt.Sprintf(`unable to create pull request.+unable to create new branch "%s"`, branchName)))
				})
			})
		})

		Context("[UI] When no infrastructure provider credentials are available in the management cluster", func() {
			It("@integration @git Verify no credentials exists in management cluster", func() {
				By("Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-template-capd.yaml")
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
					selectCredentials(createPage, "None", 1)
				})
			})
		})

		Context("[UI] When infrastructure provider credentials are available in the management cluster", func() {

			JustAfterEach(func() {
				gitopsTestRunner.DeleteIPCredentials("AWS")
				gitopsTestRunner.DeleteIPCredentials("AZURE")
			})

			It("@integration @git Verify matching selected credential can be used for cluster creation", func() {
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

				paramSection := make(map[string][]TemplateField)
				paramSection["1.Cluster"] = []TemplateField{
					{
						Name:   "CLUSTER_NAME",
						Value:  awsClusterName,
						Option: "",
					},
				}
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

					Eventually(preview.Title).Should(MatchText("PR Preview"))

					Eventually(preview.Text).Should(MatchText(fmt.Sprintf(`kind: AWSCluster\s+metadata:\s+annotations:[\s\w\d/:.-]+name: %s[\s\w\d-.:/]+identityRef:[\s\w\d-.:/]+kind: AWSClusterRoleIdentity\s+name: test-role-identity`, awsClusterName)))
					Eventually(preview.Close.Click).Should(Succeed())
				})

			})
		})

		Context("[UI] When infrastructure provider credentials are available in the management cluster", func() {

			JustAfterEach(func() {
				gitopsTestRunner.DeleteIPCredentials("AWS")
			})

			It("@integration @git Verify user can not use wrong credentials for infrastructure provider", func() {
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
					selectCredentials(createPage, "aws-test-identity", 3)
				})

				// Azure template parameter values
				azureClusterName := "my-azure-cluster"
				azureK8version := "1.21.2"
				azureNamespace := "default"
				azureControlMAchineType := "Standard_D2s_v3"
				azureNodeMAchineType := "Standard_D4_v4"

				paramSection := make(map[string][]TemplateField)
				paramSection["1.Cluster"] = []TemplateField{
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
						Value:  "",
						Option: azureK8version,
					},
				}

				paramSection["4.AzureMachineTemplate"] = []TemplateField{
					{
						Name:   "AZURE_CONTROL_PLANE_MACHINE_TYPE",
						Value:  "",
						Option: azureControlMAchineType,
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
						Value:  "",
						Option: azureNodeMAchineType,
					},
				}

				setParameterValues(createPage, paramSection)

				By("Then I should see PR preview without identity reference added to the template", func() {
					Expect(createPage.PreviewPR.Click()).To(Succeed())
					preview := pages.GetPreview(webDriver)

					Eventually(preview.Title).Should(MatchText("PR Preview"))

					Eventually(preview.Text).ShouldNot(MatchText(`kind: AWSCluster[\s\w\d-.:/]+identityRef:`), "Identity reference should not be found in preview pull request AzureCluster object")
					Eventually(preview.Close.Click).Should(Succeed())
				})

			})
		})

		Context("[UI] When leaf cluster pull request is available in the management cluster", func() {
			clusterNamespace := map[string]string{
				GitProviderGitLab: "default",
				// GitProviderGitHub: "github-system", (FIXME: there is an existing bug #563)
				GitProviderGitHub: "default",
			}

			clusterPath := "./clusters"
			capdClusterName := "ui-end-to-end-capd-cluster"
			downloadedKubeconfigPath := getDownloadedKubeconfigPath(capdClusterName)
			kustomizationFile := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "test_kustomization.yaml")

			JustBeforeEach(func() {
				_ = deleteFile([]string{downloadedKubeconfigPath})

				logger.Info("Connecting cluster to itself")
				leaf := LeafSpec{
					Status:          "Ready",
					IsWKP:           false,
					AlertManagerURL: "",
					KubeconfigPath:  "",
				}
				connectACluster(webDriver, gitopsTestRunner, leaf)
			})

			JustAfterEach(func() {
				_ = deleteFile([]string{downloadedKubeconfigPath})

				// Force clean the repository directory for subsequent tests
				cleanGitRepository(clusterPath)
				// Force delete capicluster incase delete PR fails to delete to free resources
				removeGitopsCapiClusters([]string{capdClusterName})
			})

			It("@smoke @integration @capd @git Verify leaf CAPD cluster can be provisioned and kubeconfig is available for cluster operations", func() {
				repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)

				By("Wait for cluster-service to cache profiles", func() {
					Expect(waitForProfiles(context.Background(), ASSERTION_30SECONDS_TIME_OUT)).To(Succeed())
				})

				By("Then I Apply/Install CAPITemplate", func() {
					templateFiles = gitopsTestRunner.CreateApplyCapitemplates(1, "capi-template-capd-observability.yaml")
				})

				pages.NavigateToPage(webDriver, "Templates")
				By("And wait for Templates page to be fully rendered", func() {
					templatesPage := pages.GetTemplatesPage(webDriver)
					templatesPage.WaitForPageToLoad(webDriver)
				})

				By("And User should choose a template", func() {
					templateTile := pages.GetTemplateTile(webDriver, "cluster-template-development-observability-0")
					Expect(templateTile.CreateTemplate.Click()).To(Succeed())
				})

				createPage := pages.GetCreateClusterPage(webDriver)
				By("And wait for Create cluster page to be fully rendered", func() {
					createPage.WaitForPageToLoad(webDriver)
					Eventually(createPage.CreateHeader).Should(MatchText(".*Create new cluster.*"))
				})

				// Parameter values
				clusterName := capdClusterName
				namespace := clusterNamespace[gitProviderEnv.Type]
				k8Version := "1.23.3"
				controlPlaneMachineCount := "1"
				workerMachineCount := "1"

				paramSection := make(map[string][]TemplateField)
				paramSection["1.Cluster"] = []TemplateField{
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
				}
				paramSection["4.KubeadmControlPlane"] = []TemplateField{
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
				}
				paramSection["7.MachineDeployment"] = []TemplateField{
					{
						Name:   "WORKER_MACHINE_COUNT",
						Value:  workerMachineCount,
						Option: "",
					},
				}

				setParameterValues(createPage, paramSection)
				pages.ScrollWindow(webDriver, 0, 4000)

				By("And select the podinfo profile to install", func() {
					Eventually(createPage.ProfileSelect.Click).Should(Succeed())
					Eventually(createPage.SelectProfile("podinfo").Click).Should(Succeed())
					pages.DissmissProfilePopup(webDriver)
				})

				By("And verify selected podinfo profile values.yaml", func() {
					profile := pages.GetProfile(webDriver, "podinfo")

					Eventually(profile.Version.Click).Should(Succeed())
					Eventually(pages.GetOption(webDriver, "6.0.1").Click).Should(Succeed())

					Eventually(profile.Layer.Text).Should(MatchRegexp("layer-1"))

					Eventually(profile.Values.Click).Should(Succeed())
					valuesYaml := pages.GetValuesYaml(webDriver)

					Eventually(valuesYaml.Title.Text).Should(MatchRegexp("podinfo"))
					Eventually(valuesYaml.TextArea.Text).Should(MatchRegexp("tag: 6.0.1"))
					Eventually(valuesYaml.Cancel.Click).Should(Succeed())
				})

				By("And verify default observability profile values.yaml", func() {
					profile := pages.GetProfile(webDriver, "observability")
					Eventually(profile.Layer.Text).Should(MatchRegexp("layer-0"))

					Eventually(profile.Values.Click).Should(Succeed())
					valuesYaml := pages.GetValuesYaml(webDriver)

					Eventually(valuesYaml.Title.Text).Should(MatchRegexp("observability"))
					Eventually(valuesYaml.TextArea.Text).Should(MatchRegexp("kube-prometheus-stack:"))
					Eventually(valuesYaml.Cancel.Click).Should(Succeed())
				})

				By("Then I should preview the PR", func() {
					Expect(createPage.PreviewPR.Click()).To(Succeed())
					preview := pages.GetPreview(webDriver)

					Eventually(preview.Title).Should(MatchText("PR Preview"))

					Eventually(preview.Text).Should(MatchText(`kind: Cluster[\s\w\d./:-]*metadata:[\s\w\d./:-]*labels:[\s\w\d./:-]*cni: calico[\s\w\d./:-]*weave.works/capi: bootstrap`))

					Eventually(preview.Close.Click).Should(Succeed())
				})

				// Pull request values
				prBranch := fmt.Sprintf("br-%s", clusterName)
				prTitle := "CAPD pull request"
				prCommit := "CAPD capi template"

				By("And set GitOps values for pull request", func() {
					gitops := pages.GetGitOps(webDriver)
					Eventually(gitops.GitOpsLabel).Should(BeFound())

					Expect(gitops.GitOpsFields[0].Label).Should(BeFound())
					Expect(gitops.GitOpsFields[0].Field.SendKeys(prBranch)).To(Succeed())
					Expect(gitops.GitOpsFields[1].Label).Should(BeFound())
					Expect(gitops.GitOpsFields[1].Field.SendKeys(prTitle)).To(Succeed())
					Expect(gitops.GitOpsFields[2].Label).Should(BeFound())
					Expect(gitops.GitOpsFields[2].Field.SendKeys(prCommit)).To(Succeed())

					AuthenticateWithGitProvider(webDriver, gitProviderEnv.Type, gitProviderEnv.Hostname)
					Eventually(gitops.GitCredentials).Should(BeVisible())

					// Wait for template to be reloaded before submitting
					Eventually(webDriver.Find("#root_CLUSTER_NAME-label")).Should(BeFound())

					Expect(gitops.CreatePR.Click()).To(Succeed())
				})

				clustersPage := pages.GetClustersPage(webDriver)
				By("Then I should see cluster appears in the cluster dashboard with 'Creation PR' status", func() {
					clusterInfo := pages.FindClusterInList(clustersPage, clusterName)
					Eventually(clusterInfo.Status, ASSERTION_1MINUTE_TIME_OUT).Should(HaveText("Creation PR"))
				})

				By("Then I should merge the pull request to start cluster provisioning", func() {
					createPRUrl := verifyPRCreated(gitProviderEnv, repoAbsolutePath)
					mergePullRequest(gitProviderEnv, repoAbsolutePath, createPRUrl)
				})

				By("And I add a test kustomization file to the management appliction (because flux doesn't reconcile empty folders on deletion)", func() {
					pullGitRepo(repoAbsolutePath)
					_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("cp -f %s %s", kustomizationFile, path.Join(repoAbsolutePath, clusterPath)))
					gitUpdateCommitPush(repoAbsolutePath, "")
				})

				By("Then I should see cluster status changes to 'Cluster found'", func() {
					waitForGitRepoReady("flux-system", GITOPS_DEFAULT_NAMESPACE)
					Eventually(pages.FindClusterInList(clustersPage, clusterName).Status, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_15SECONDS).Should(HaveText("Cluster found"))
				})

				By("And I should download the kubeconfig for the CAPD capi cluster", func() {
					clusterInfo := pages.FindClusterInList(clustersPage, clusterName)
					Expect(clusterInfo.ShowStatusDetail.Click()).To(Succeed())
					clusterStatus := pages.GetClusterStatus(webDriver)
					Eventually(clusterStatus.Phase, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_15SECONDS).Should(HaveText(`"Provisioned"`))

					fileErr := func() error {
						Expect(clusterStatus.KubeConfigButton.Click()).To(Succeed())
						_, err := os.Stat(downloadedKubeconfigPath)
						return err

					}
					Eventually(fileErr, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_15SECONDS).ShouldNot(HaveOccurred())
				})

				By(fmt.Sprintf("And verify that %s capd cluster kubeconfig is correct", clusterName), func() {
					verifyCapiClusterKubeconfig(downloadedKubeconfigPath, clusterName)
				})

				By(fmt.Sprintf("And I verify %s capd cluster is healthy and profiles are installed)", clusterName), func() {
					// List of Profiles in order of layering
					profiles := []string{"observability", "podinfo"}
					verifyCapiClusterHealth(downloadedKubeconfigPath, clusterName, profiles, GITOPS_DEFAULT_NAMESPACE)
				})

				By("Then I should select the cluster to create the delete pull request", func() {
					clusterInfo := pages.FindClusterInList(clustersPage, clusterName)
					Expect(clusterInfo.Checkbox.Click()).To(Succeed())

					Eventually(webDriver.FindByXPath(`//button[@id="delete-cluster"][@disabled]`)).ShouldNot(BeFound())
					Expect(clustersPage.PRDeleteClusterButton.Click()).To(Succeed())

					deletePR := pages.GetDeletePRPopup(webDriver)
					Expect(deletePR.PRDescription.SendKeys("Delete CAPD capi cluster, it is not required any more")).To(Succeed())

					AuthenticateWithGitProvider(webDriver, gitProviderEnv.Type, gitProviderEnv.Hostname)
					Eventually(deletePR.GitCredentials).Should(BeVisible())

					Expect(deletePR.DeleteClusterButton.Click()).To(Succeed())
				})

				By(fmt.Sprintf("Then I should see the '%s' cluster status changes to Deletion PR", clusterName), func() {
					clusterInfo := pages.FindClusterInList(clustersPage, clusterName)
					Eventually(clusterInfo.Status, ASSERTION_30SECONDS_TIME_OUT).Should(HaveText("Deletion PR"))
				})

				var deletePRUrl string
				By("And I should veriyfy the delete pull request in the cluster config repository", func() {
					deletePRUrl = verifyPRCreated(gitProviderEnv, repoAbsolutePath)
				})

				By("Then I should merge the delete pull request to delete cluster", func() {
					mergePullRequest(gitProviderEnv, repoAbsolutePath, deletePRUrl)
				})

				By("And the delete pull request manifests are not present in the cluster config repository", func() {
					pullGitRepo(repoAbsolutePath)
					_, err := os.Stat(fmt.Sprintf("%s/management/%s.yaml", repoAbsolutePath, clusterName))
					Expect(err).Should(MatchError(os.ErrNotExist), "Cluster config is found when expected to be deleted.")
				})

				By(fmt.Sprintf("Then I should see the '%s' cluster deleted", clusterName), func() {
					clusterFound := func() error {
						return runCommandPassThrough("kubectl", "get", "cluster", clusterName)
					}
					Eventually(clusterFound, ASSERTION_5MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(HaveOccurred())
				})
			})
		})

		Context("[UI] When entitlement is available in the cluster", func() {
			DEPLOYMENT_APP := "my-mccp-cluster-service"

			checkEntitlement := func(typeEntitelment string, beFound bool) {
				checkOutput := func() bool {
					if !pages.ElementExist(pages.GetClustersPage(webDriver).Version) {
						Expect(webDriver.Refresh()).ShouldNot(HaveOccurred())
					}
					found, _ := pages.GetEntitelment(webDriver, typeEntitelment).Visible()
					return found

				}

				Expect(webDriver.Refresh()).ShouldNot(HaveOccurred())
				loginUser()

				if beFound {
					Eventually(checkOutput, ASSERTION_2MINUTE_TIME_OUT).Should(BeTrue())
				} else {
					Eventually(checkOutput, ASSERTION_2MINUTE_TIME_OUT).Should(BeFalse())
				}

			}

			JustAfterEach(func() {
				By("When I apply the valid entitlement", func() {
					Expect(gitopsTestRunner.KubectlApply([]string{}, path.Join(getCheckoutRepoPath(), "test", "utils", "scripts", "entitlement-secret.yaml")), "Failed to create/configure entitlement")
				})

				By("Then I restart the cluster service pod for valid entitlemnt to take effect", func() {
					Expect(gitopsTestRunner.RestartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
				})

				By("And the Cluster service is healthy", func() {
					CheckClusterService(capi_endpoint_url)
				})

				By("And I should not see the error or warning message for valid entitlement", func() {
					checkEntitlement("expired", false)
					checkEntitlement("missing", false)
				})
			})

			It("@integration Verify cluster service acknowledges the entitlement presences", func() {

				By("When I delete the entitlement", func() {
					Expect(gitopsTestRunner.KubectlDelete([]string{}, path.Join(getCheckoutRepoPath(), "test", "utils", "scripts", "entitlement-secret.yaml")), "Failed to delete entitlement secret")
				})

				By("Then I restart the cluster service pod for missing entitlemnt to take effect", func() {
					Expect(gitopsTestRunner.RestartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE)).ShouldNot(HaveOccurred(), "Failed restart deployment successfully")
				})

				By("And I should see the error message for missing entitlement", func() {
					checkEntitlement("missing", true)

				})

				By("When I apply the expired entitlement", func() {
					Expect(gitopsTestRunner.KubectlApply([]string{}, path.Join(getCheckoutRepoPath(), "test", "utils", "data", "entitlement-secret-expired.yaml")), "Failed to create/configure entitlement")
				})

				By("Then I restart the cluster service pod for expired entitlemnt to take effect", func() {
					Expect(gitopsTestRunner.RestartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
				})

				By("And I should see the warning message for expired entitlement", func() {
					checkEntitlement("expired", true)
				})

				By("When I apply the invalid entitlement", func() {
					Expect(gitopsTestRunner.KubectlApply([]string{}, path.Join(getCheckoutRepoPath(), "test", "utils", "data", "entitlement-secret-invalid.yaml")), "Failed to create/configure entitlement")
				})

				By("Then I restart the cluster service pod for invalid entitlemnt to take effect", func() {
					Expect(gitopsTestRunner.RestartDeploymentPods(DEPLOYMENT_APP, GITOPS_DEFAULT_NAMESPACE), "Failed restart deployment successfully")
				})

				By("And I should see the error message for invalid entitlement", func() {
					checkEntitlement("invalid", true)
				})
			})
		})
	})
}
