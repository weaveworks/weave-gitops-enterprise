package acceptance

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

type TemplateField struct {
	Name   string
	Value  string
	Option string
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

func installGitOpsTemplate(templateFiles map[string]string) {
	ginkgo.By("Installing GitOpsTemplate...", func() {
		for _, templateFile := range templateFiles {
			err := runCommandPassThrough("kubectl", "apply", "-f", templateFile)
			gomega.Expect(err).To(gomega.BeNil(), fmt.Sprintf("Failed to apply GitOpsTemplate template %s", templateFile))
		}
	})
}

func waitForTemplatesToAppear(templateCpunt int) {
	ginkgo.By("And wait for Templates to be rendered", func() {
		templatesPage := pages.GetTemplatesPage(webDriver)
		gomega.Eventually(func(g gomega.Gomega) {
			g.Expect(webDriver.Refresh()).ShouldNot(gomega.HaveOccurred())
			pages.WaitForPageToLoad(webDriver)
			g.Eventually(templatesPage.TemplateHeader).Should(matchers.BeVisible())
			g.Eventually(templatesPage.CountTemplateRows).Should(gomega.Equal(templateCpunt))
		}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).ShouldNot(gomega.HaveOccurred(), "The number of template rows should be equal to number of templates created")
	})
}

var _ = ginkgo.Describe("Multi-Cluster Control Plane GitOpsTemplates", ginkgo.Label("ui", "template"), func() {
	var templateNamespaces []string

	ginkgo.BeforeEach(func() {
		gomega.Expect(webDriver.Navigate(testUiUrl)).To(gomega.Succeed())

		if !pages.ElementExist(pages.Navbar(webDriver).Title, 3) {
			loginUser()
		}
	})

	ginkgo.AfterEach(func() {
		_ = runCommandPassThrough("kubectl", "delete", "CapiTemplate", "--all")
		_ = runCommandPassThrough("kubectl", "delete", "GitOpsTemplate", "--all")
		deleteNamespace(templateNamespaces)
	})

	ginkgo.Context("[UI] When no GitOps Templates are available in the cluster", func() {

		ginkgo.It("Verify template page renders no GitOpsTemplate", func() {
			ginkgo.By("And wait for  good looking response from /v1/templates", func() {
				gomega.Expect(waitForGitopsResources(context.Background(), Request{Path: "templates"}, POLL_INTERVAL_15SECONDS)).To(gomega.Succeed(), "Failed to get a successful response from /v1/templates")
			})

			pages.NavigateToPage(webDriver, "Templates")
			pages.WaitForPageToLoad(webDriver)
			templatesPage := pages.GetTemplatesPage(webDriver)

			ginkgo.By("And wait for Templates page to be rendered", func() {
				gomega.Eventually(templatesPage.TemplateHeader).Should(matchers.BeVisible())
				gomega.Expect(templatesPage.CountTemplateRows()).To(gomega.Equal(0), "There should not be any template visible/available in template's table")

			})
		})
	})

	ginkgo.Context("[UI] When GitOps Templates are available in the cluster", func() {

		ginkgo.It("Verify template(s) are rendered from the template library.", func() {
			// Namespace for some GitOpsTemplates
			templateNamespaces = []string{"dev-system", "test-system"}
			createNamespace(templateNamespaces)

			templateFiles := map[string]string{
				"capa-cluster-template":             path.Join(testDataPath, "templates/cluster/aws/cluster-template-ec2.yaml"),
				"capa-cluster-template-eks-fargate": path.Join(testDataPath, "templates/cluster/aws/cluster-template-eks-fargate.yaml"),
				"capa-cluster-template-eks":         path.Join(testDataPath, "templates/cluster/aws/cluster-template-eks.yaml"),
				"capa-cluster-template-machinepool": path.Join(testDataPath, "templates/cluster/aws/cluster-template-machinepool.yaml"),
				"capz-cluster-template":             path.Join(testDataPath, "templates/cluster/azure/cluster-template-e2e.yaml"),
				"capd-cluster-template":             path.Join(testDataPath, "templates/cluster/docker/cluster-template.yaml"),
				"capg-cluster-template":             path.Join(testDataPath, "templates/cluster/gcp/cluster-template-gke.yaml"),
				"connect-a-cluster":                 path.Join(testDataPath, "templates/cluster/gitops/cluster-template.yaml"),
				"git-repository-template":           path.Join(testDataPath, "templates/source/git-repository-template.yaml"),
				"helm-repository-template":          path.Join(testDataPath, "templates/source/helm-repository-template.yaml"),
				"git-kustomization-template":        path.Join(testDataPath, "templates/application/git-kustomization-template.yaml"),
				"helmrelease-template":              path.Join(testDataPath, "templates/application/helmrelease-template.yaml"),
			}

			var keys []string
			for k := range templateFiles {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			sourceTemplateCount := 2
			namespaceSourceTemplateCount := 2
			clusterTemplateCount := 8
			awsTemplateCount := 4

			installGitOpsTemplate(templateFiles)
			pages.NavigateToPage(webDriver, "Templates")
			waitForTemplatesToAppear(len(templateFiles))

			templatesPage := pages.GetTemplatesPage(webDriver)
			ginkgo.By("And templates are ordered - table view", func() {
				actual_list := templatesPage.GetTemplateTableList()
				for i, key := range keys {
					gomega.Expect(actual_list[i]).Should(gomega.ContainSubstring(key))
				}
			})

			ginkgo.By("And templates can be filtered by type - table view", func() {
				searchPage := pages.GetSearchPage(webDriver)

				// Select the 'templateType' filter
				filterID := "templateType: source"
				searchPage.SelectFilter("templateType", filterID)
				gomega.Eventually(templatesPage.CountTemplateRows()).Should(gomega.Equal(sourceTemplateCount), "The number of filtered templates should be equal to number of source templates created")
				searchPage.SelectFilter("templateType", filterID, false) // Reset the 'source' templateType filter

				// Select the 'namespace' filter
				filterID = "namespace: dev-system"
				searchPage.SelectFilter("namespace", filterID)
				gomega.Eventually(templatesPage.CountTemplateRows()).Should(gomega.Equal(namespaceSourceTemplateCount), "The number of filtered templates should be equal to number of source templates created in a namespace")
				searchPage.SelectFilter("namespace", filterID, false) // Reset the 'namespace' filter

				// Select the 'cluster' templateType filter
				filterID = "templateType: cluster"
				searchPage.SelectFilter("templateType", filterID)
				gomega.Eventually(templatesPage.CountTemplateRows()).Should(gomega.Equal(clusterTemplateCount), "The number of filtered templates should be equal to number of cluster templates created")

				// Select the 'aws' provider filter
				filterID = "provider: aws"
				searchPage.SelectFilter("provider", filterID)
				gomega.Eventually(templatesPage.CountTemplateRows()).Should(gomega.Equal(awsTemplateCount), "The number of filtered templates should be equal to number of aws templates created")
			})
		})

		ginkgo.It("Verify I should be able to select a template of my choice", func() {
			ginkgo.By("Installing GitOpsTemplate...", func() {
				templatedTemplate := path.Join(testDataPath, "templates/miscellaneous/templated-cluster-template.yaml")
				templateFiles, err := generateTestTemplates(50, templatedTemplate)
				gomega.Expect(err).To(gomega.BeNil(), fmt.Sprintf("Failed to generate template test files from %s", templatedTemplate))
				for _, fileName := range templateFiles {
					err = runCommandPassThrough("kubectl", "apply", "-f", fileName)
					gomega.Expect(err).To(gomega.BeNil(), fmt.Sprintf("Failed to apply GitOpsTemplate template %s", fileName))
				}
			})

			pages.NavigateToPage(webDriver, "Templates")
			pages.WaitForPageToLoad(webDriver)
			templatesPage := pages.GetTemplatesPage(webDriver)

			ginkgo.By("And I should choose a template from the default table view", func() {
				templateRow := templatesPage.GetTemplateInformation(webDriver, "capg-cluster-template-10")
				gomega.Eventually(templateRow.Type).Should(matchers.MatchText(""))
				gomega.Eventually(templateRow.Namespace).Should(matchers.MatchText("default"))
				gomega.Eventually(templateRow.Provider).Should(matchers.MatchText(""))
				gomega.Eventually(templateRow.Description).Should(matchers.MatchText("This is the std. CAPG template 10"))
				gomega.Expect(templateRow.CreateTemplate).Should(matchers.BeFound())
				gomega.Expect(templateRow.CreateTemplate.Click()).To(gomega.Succeed())
			})

			ginkgo.By("And wait for Create resource page to be fully rendered - table view", func() {
				createPage := pages.GetCreateClusterPage(webDriver)
				gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(".*Create new resource.*"))
			})
		})

		ginkgo.It("Verify UI shows message related to an invalid template(s) when valid templates are not available", func() {

			templateFiles := map[string]string{
				"invalid-cluster-template": path.Join(testDataPath, "templates/miscellaneous/invalid-cluster-template.yaml"),
			}

			installGitOpsTemplate(templateFiles)
			pages.NavigateToPage(webDriver, "Templates")
			pages.WaitForPageToLoad(webDriver)
			templatesPage := pages.GetTemplatesPage(webDriver)

			ginkgo.By("And User should see message informing user of the invalid template in the cluster - table view", func() {
				templateRow := templatesPage.GetTemplateInformation(webDriver, "invalid-cluster-template")
				gomega.Eventually(templateRow.Type).Should(matchers.MatchText(""))
				gomega.Eventually(templateRow.Namespace).Should(matchers.MatchText("default"), "Failed to match the namespace for invalid template")
				gomega.Eventually(templateRow.Provider).Should(matchers.MatchText(""), "The should be no provider for invalid template")
				gomega.Expect(templateRow.Description).Should(matchers.MatchText("Couldn't load template body"), "Failed to find invalid template error message")
				gomega.Expect(templateRow.CreateTemplate).ShouldNot(matchers.BeEnabled(), "The button 'USE THIS TEMPLATE' should be disabled for invalid emplate")
			})
		})

		ginkgo.It("Verify UI shows message related to an invalid template(s) and renders the available valid template(s)", func() {

			templateNamespaces = []string{"dev-system", "test-system"}
			createNamespace(templateNamespaces)

			templateFiles := map[string]string{
				"capa-cluster-template-eks-fargate": path.Join(testDataPath, "templates/cluster/aws/cluster-template-eks-fargate.yaml"),
				"helm-repository-template":          path.Join(testDataPath, "templates/source/helm-repository-template.yaml"),
				"git-kustomization-template":        path.Join(testDataPath, "templates/application/git-kustomization-template.yaml"),
				"invalid-cluster-template":          path.Join(testDataPath, "templates/miscellaneous/invalid-cluster-template.yaml"),
			}
			installGitOpsTemplate(templateFiles)

			pages.NavigateToPage(webDriver, "Templates")
			waitForTemplatesToAppear(len(templateFiles))
			templatesPage := pages.GetTemplatesPage(webDriver)

			ginkgo.By("And User should see message informing user of the invalid template in the cluster", func() {
				templateRow := templatesPage.GetTemplateInformation(webDriver, "invalid-cluster-template")
				gomega.Eventually(templateRow.Type).Should(matchers.MatchText(""))
				gomega.Eventually(templateRow.Namespace).Should(matchers.MatchText("default"), "Failed to match the namespace for invalid template")
				gomega.Eventually(templateRow.Provider).Should(matchers.MatchText(""), "The should be no provider for invalid template")
				gomega.Expect(templateRow.Description).Should(matchers.MatchText("Couldn't load template body"), "Failed to find invalid template error message")
				gomega.Expect(templateRow.CreateTemplate).ShouldNot(matchers.BeEnabled(), "The button 'USE THIS TEMPLATE' should be disabled for invalid emplate")
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

		ginkgo.It("Verify pull request for cluster can be created to the management cluster", func() {
			repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)

			templateFiles := map[string]string{
				"capa-cluster-template-eks-fargate": path.Join(testDataPath, "templates/cluster/docker/cluster-template.yaml"),
			}

			installGitOpsTemplate(templateFiles)
			pages.NavigateToPage(webDriver, "Templates")
			pages.WaitForPageToLoad(webDriver)
			templatesPage := pages.GetTemplatesPage(webDriver)

			templateName := "capd-cluster-template"
			ginkgo.By("And I should choose a template", func() {
				templateRow := templatesPage.GetTemplateInformation(webDriver, templateName)
				gomega.Expect(templateRow.CreateTemplate.Click()).To(gomega.Succeed())
			})

			createPage := pages.GetCreateClusterPage(webDriver)
			ginkgo.By("And wait for Create resource page to be fully rendered", func() {
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
				gomega.Eventually(preview.Close.Click).Should(gomega.Succeed(), "Failed to close the preview dialog")
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

		ginkgo.It("Verify pull request can not be created by using exiting repository branch", func() {
			repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
			// Checkout repo main branch in case of test failure
			defer checkoutRepoBranch(repoAbsolutePath, "main")

			branchName := "ui-test-branch"
			ginkgo.By("And create new git repository branch", func() {
				_ = createGitRepoBranch(repoAbsolutePath, branchName)
			})

			templateFiles := map[string]string{
				"capa-cluster-template-eks-fargate": path.Join(testDataPath, "templates/cluster/docker/cluster-template.yaml"),
			}

			installGitOpsTemplate(templateFiles)
			pages.NavigateToPage(webDriver, "Templates")
			pages.WaitForPageToLoad(webDriver)
			templatesPage := pages.GetTemplatesPage(webDriver)

			ginkgo.By("And I should choose a template", func() {
				templateRow := templatesPage.GetTemplateInformation(webDriver, "capd-cluster-template")
				gomega.Expect(templateRow.CreateTemplate.Click()).To(gomega.Succeed())
			})

			createPage := pages.GetCreateClusterPage(webDriver)
			ginkgo.By("And wait for Create resource page to be fully rendered", func() {
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
				pages.ClearFieldValue(gitops.PullRequestTitle)
				gomega.Expect(gitops.PullRequestTitle.SendKeys(prTitle)).To(gomega.Succeed())
				pages.ClearFieldValue(gitops.CommitMessage)
				gomega.Expect(gitops.CommitMessage.SendKeys(prCommit)).To(gomega.Succeed())

				authenticateWithGitProvider(webDriver, gitProviderEnv.Type, gitProviderEnv.Hostname)
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

		ginkgo.It("Verify render type 'envsubst' supported functions", func() {
			templateFiles := map[string]string{
				"capz-cluster-template": path.Join(testDataPath, "templates/cluster/azure/cluster-template-e2e.yaml"),
			}

			installGitOpsTemplate(templateFiles)
			pages.NavigateToPage(webDriver, "Templates")
			pages.WaitForPageToLoad(webDriver)
			templatesPage := pages.GetTemplatesPage(webDriver)

			ginkgo.By("And I should choose a template", func() {
				templateRow := templatesPage.GetTemplateInformation(webDriver, "capz-cluster-template")
				gomega.Expect(templateRow.CreateTemplate.Click()).To(gomega.Succeed())
			})

			createPage := pages.GetCreateClusterPage(webDriver)
			ginkgo.By("And wait for Create resource page to be fully rendered", func() {
				pages.WaitForPageToLoad(webDriver)
				gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(".*Create new resource.*"))
			})

			// Parameter values
			clusterName := "quick-CAPZ-cluster"
			namespace := "CAPZ-SYSTEM"
			controlPlaneMachineCount := "2"

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
					Value:  controlPlaneMachineCount,
					Option: "",
				},
			}

			setParameterValues(createPage, parameters)

			preview := pages.GetPreview(webDriver)
			ginkgo.By("Then I should preview the PR", func() {
				gomega.Eventually(func(g gomega.Gomega) {
					g.Expect(createPage.PreviewPR.Click()).Should(gomega.Succeed())
					g.Expect(preview.Title.Text()).Should(gomega.MatchRegexp("PR Preview"))

				}, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Succeed(), "Failed to get PR preview")
			})

			ginkgo.By("Then verify PR preview contents for envsubst functions", func() {
				// Verify resource definition preview
				gomega.Eventually(preview.GetPreviewTab("Resource Definition").Click).Should(gomega.Succeed(), "Failed to switch to 'RESOURCE DEFINITION' preview tab")
				// Verify CLUSTER_NAME and NAMESPACE parameter values should be converted to lowecase - template is using envsubst function ${var,,}
				gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`kind: Cluster[\s\w\d./:-]*metadata:[\s\w\d./:-]*name: %s[\s\w\d./:-]*namespace: %s`, strings.ToLower(clusterName), strings.ToLower(namespace))))
				gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`machineTemplate[\s\w\d./:-]*infrastructureRef:[\s\w\d./:-]*name: %s[\s\w\d./:-]*replicas: %s`, strings.ToLower(clusterName), controlPlaneMachineCount)))
				// Verify WORKER_MACHINE_COUNT should be same as CONTROL_PLANE_MACHINE_COUNT - template is using envsubst function ${var:-${default}}
				gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`kind: MachineDeployment[\s\w\d./:-]*spec:[\s\w\d./:-]*clusterName: %s[\s\w\d./:-]*replicas: %s`, strings.ToLower(clusterName), controlPlaneMachineCount)))

				// Verify profile tab view is disabled due to no profile is part of pull request
				gomega.Expect(preview.GetPreviewTab("Profiles").Attribute("class")).Should(gomega.MatchRegexp("Mui-disabled"), "'PROFILES' preview tab should be disabled")
				// Verify kustomizations preview
				gomega.Eventually(preview.GetPreviewTab("Kustomizations").Click).Should(gomega.Succeed(), "Failed to switch to 'KUSTOMIZATION' preview tab")
				gomega.Eventually(preview.Text).Should(matchers.MatchText(`kind: Kustomization[\s\w\d./:-]*name: clusters-bases-kustomization[\s\w\d./:-]*namespace: flux-system`))
			})
		})

		ginkgo.It("Verify render type 'templating' supported functions", func() {
			// Namespace for some GitOpsTemplates
			templateNamespaces = []string{"test-system"}
			createNamespace(templateNamespaces)

			templateName := "git-kustomization-template"
			templateNamespace := "default"
			templateFiles := map[string]string{
				templateName: path.Join(testDataPath, "templates/application/git-kustomization-template.yaml"),
			}

			installGitOpsTemplate(templateFiles)
			pages.NavigateToPage(webDriver, "Templates")
			waitForTemplatesToAppear(len(templateFiles))
			templatesPage := pages.GetTemplatesPage(webDriver)

			ginkgo.By("And I should choose a template", func() {
				templateRow := templatesPage.GetTemplateInformation(webDriver, templateName)
				gomega.Eventually(templateRow.CreateTemplate.Click, ASSERTION_2MINUTE_TIME_OUT).Should(gomega.Succeed())
			})

			createPage := pages.GetCreateClusterPage(webDriver)
			ginkgo.By("And wait for Create resource page to be fully rendered", func() {
				pages.WaitForPageToLoad(webDriver)
				gomega.Eventually(createPage.CreateHeader).Should(matchers.MatchText(".*Create new resource.*"))
			})

			app := Application{
				Name:            "my-podinfo",
				Namespace:       "dev-system",
				Source:          "podinfo",
				Path:            "./kustomize",
				TargetNamespace: "dev-system",
				Description:     `Podinfo is a tiny web application made with Go that showcases best practices of running microservices in Kubernetes. Podinfo is used by CNCF projects like Flux and Flagger for end-to-end testing and workshops.`,
			}

			var parameters = []TemplateField{
				{
					Name:   "RESOURCE_NAME",
					Value:  app.Name,
					Option: "",
				},
				{
					Name:   "NAMESPACE",
					Value:  app.Namespace,
					Option: "",
				},
				{
					Name:   "PATH",
					Value:  app.Path,
					Option: "",
				},
				{
					Name:   "SOURCE_NAME",
					Value:  app.Source,
					Option: "",
				},
				{
					Name:   "TARGET_NAMESPACE",
					Value:  app.TargetNamespace,
					Option: "",
				},
				{
					Name:   "DESCRIPTION",
					Value:  app.Description,
					Option: "",
				},
			}

			setParameterValues(createPage, parameters)

			preview := pages.GetPreview(webDriver)
			ginkgo.By("Then I should preview the PR", func() {
				gomega.Eventually(func(g gomega.Gomega) {
					g.Expect(createPage.PreviewPR.Click()).Should(gomega.Succeed())
					g.Expect(preview.Title.Text()).Should(gomega.MatchRegexp("PR Preview"))

				}, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Succeed(), "Failed to get PR preview")
			})

			ginkgo.By("Then verify PR preview contents for templating functions", func() {
				// Verify resource definition preview
				gomega.Eventually(preview.GetPreviewTab("Resource Definition").Click).Should(gomega.Succeed(), "Failed to switch to 'RESOURCE DEFINITION' preview tab")
				// Verify resource is labelled with template name and namespace
				gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`labels:[\s]*templates.weave.works/template-name: %s[\s]*templates.weave.works/template-namespace: %s`, templateName, templateNamespace)))
				gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`kind: Kustomization[\s]*metadata:[|=\s\w\d./:-]*name: %s[\s]*namespace: %s`, app.Name, app.Namespace)))
				// Verify PATH should be assigned the same value set as parameter - template is using templating functions '.params.PATH | empty |  ternary "./" .params.PATH'
				gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`path: %s`, app.Path)))
				// Verify applicartion description is base64 encoder - template is using templating function '.params.DESCRIPTION | b64enc'
				desEnc := base64.StdEncoding.EncodeToString([]byte(app.Description))
				gomega.Eventually(preview.Text).Should(matchers.MatchText(fmt.Sprintf(`metadata.weave.works/description: \|[\s]*%s\s`, desEnc)))

				// Verify profiles and kustomization tab views are disabled because no profiles and kustomizations are part of pull request
				gomega.Expect(preview.GetPreviewTab("Profiles").Attribute("class")).Should(gomega.MatchRegexp("Mui-disabled"), "'PROFILES' preview tab should be disabled")
				gomega.Expect(preview.GetPreviewTab("Kustomizations").Attribute("class")).Should(gomega.MatchRegexp("Mui-disabled"), "'KUSTOMIZATIONS' preview tab should be disabled")
			})
		})
	})
})
