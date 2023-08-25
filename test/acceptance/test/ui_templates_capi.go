package acceptance

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

func addKustomizationManifests(manifestYamls []string) string {
	manifestPath := "./apps"
	ginkgo.By("Add Application/Kustomization manifests to management cluster's repository main branch)", func() {
		repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
		pullGitRepo(repoAbsolutePath)

		for _, yaml := range manifestYamls {
			manifest := path.Join(testDataPath, yaml)
			_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("mkdir -p %[2]v && cp -f %[1]v %[2]v", manifest, path.Join(repoAbsolutePath, manifestPath, strings.TrimSuffix(filepath.Base(yaml), filepath.Ext(yaml)))))
		}

		gitUpdateCommitPush(repoAbsolutePath, "Adding application kustomization manifests")
	})
	return manifestPath
}

var _ = ginkgo.Describe("Multi-Cluster Control Plane GitOpsTemplates for CAPI cluster", ginkgo.Label("ui", "template"), func() {
	clusterPath := "./clusters/management/clusters"

	ginkgo.BeforeEach(func() {
		gomega.Expect(webDriver.Navigate(testUiUrl)).To(gomega.Succeed())

		if !pages.ElementExist(pages.Navbar(webDriver).Title, 3) {
			loginUser()
		}
	})

	ginkgo.AfterEach(func() {
		_ = runCommandPassThrough("kubectl", "delete", "CapiTemplate", "--all")
		_ = runCommandPassThrough("kubectl", "delete", "GitOpsTemplate", "--all")
	})

	ginkgo.Context("When leaf cluster pull request is available in the management cluster", ginkgo.Label("capd"), func() {
		var clusterBootstrapCopnfig string
		var clusterResourceSet string
		var crsConfigmap string
		var downloadedKubeconfigPath string
		var capdCluster ClusterConfig
		var appSourcePath string

		clusterNamespace := map[string]string{
			GitProviderGitLab: "capi-test-system",
			GitProviderGitHub: "default",
		}

		bootstrapLabel := "bootstrap"
		patSecret := "capi-pat"

		ginkgo.JustBeforeEach(func() {
			capdCluster = ClusterConfig{"capd", "ui-end-to-end-capd-cluster-" + strings.ToLower(randString(6)), clusterNamespace[gitProviderEnv.Type]}
			downloadedKubeconfigPath = path.Join(os.Getenv("HOME"), "Downloads", fmt.Sprintf("%s.kubeconfig", capdCluster.Name))
			_ = deleteFile([]string{downloadedKubeconfigPath})

			createNamespace([]string{capdCluster.Namespace})
			createPATSecret(capdCluster.Namespace, patSecret)
			clusterBootstrapCopnfig = createClusterBootstrapConfig(capdCluster.Name, capdCluster.Namespace, bootstrapLabel, patSecret)
			clusterResourceSet = createClusterResourceSet(capdCluster.Name, capdCluster.Namespace)
			crsConfigmap = createCRSConfigmap(capdCluster.Name, capdCluster.Namespace)
		})

		ginkgo.JustAfterEach(func() {
			pages.CloseOtherWindows(webDriver, enterpriseWindow)
			_ = deleteFile([]string{downloadedKubeconfigPath})
			deleteSecret([]string{patSecret}, capdCluster.Namespace)
			_ = runCommandPassThrough("kubectl", "delete", "-f", clusterBootstrapCopnfig)
			_ = runCommandPassThrough("kubectl", "delete", "-f", crsConfigmap)
			_ = runCommandPassThrough("kubectl", "delete", "-f", clusterResourceSet)

			reconcile("suspend", "source", "git", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
			// Force clean the repository directory for subsequent tests
			cleanGitRepository(clusterPath)
			cleanGitRepository(path.Join("./clusters", capdCluster.Namespace))
			cleanGitRepository(appSourcePath)
			// Force delete capicluster incase delete PR fails to delete to free resources
			removeGitopsCapiClusters([]ClusterConfig{capdCluster})
			reconcile("resume", "source", "git", "flux-system", GITOPS_DEFAULT_NAMESPACE, "")
		})

		ginkgo.It("Verify leaf CAPD cluster can be provisioned and kubeconfig is available for cluster operations", ginkgo.Label("smoke"), func() {
			repoAbsolutePath := configRepoAbsolutePath(gitProviderEnv)
			appSourcePath = addKustomizationManifests([]string{"deployments/postgres-manifest.yaml", "deployments/podinfo-manifest.yaml"})

			templateFiles := map[string]string{
				"capd-cluster-template": path.Join(testDataPath, "templates/cluster/docker/cluster-template.yaml"),
			}
			installGitOpsTemplate(templateFiles)

			ginkgo.By("And wait for cluster-service to cache profiles", func() {
				gomega.Expect(waitForGitopsResources(context.Background(), Request{Path: `charts/list?repository.name=weaveworks-charts&repository.namespace=flux-system&repository.cluster.name=management`}, POLL_INTERVAL_5SECONDS, ASSERTION_15MINUTE_TIME_OUT)).To(gomega.Succeed(), "Failed to get a successful response from /v1/charts")
			})

			pages.NavigateToPage(webDriver, "Templates")
			waitForTemplatesToAppear(len(templateFiles))

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
				Path:            "apps/postgres-manifest",
				CreateNamespace: false,
			}
			gomega.Expect(createPage.AddApplication.Click()).Should(gomega.Succeed(), "Failed to click 'Add application' button")
			application := pages.GetAddApplication(webDriver, 1)
			AddKustomizationApp(application, postgres)

			podinfo := Application{
				Type:            "kustomization",
				Name:            "podinfo",
				Namespace:       GITOPS_DEFAULT_NAMESPACE,
				TargetNamespace: GITOPS_DEFAULT_NAMESPACE,
				Path:            "apps/podinfo-manifest",
				CreateNamespace: false,
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

				gomega.Eventually(preview.GetPreviewTab("Resource Definition").Click).Should(gomega.Succeed(), "Failed to switch to 'RESOURCE DEFINITION' preview tab")
				gomega.Eventually(preview.Path.At(0)).Should(matchers.MatchText(path.Join(clusterPath, leafCluster.Namespace, leafCluster.Name+".yaml")))
				gomega.Eventually(preview.Text.At(0)).Should(matchers.MatchText("cni: calico"))
				gomega.Eventually(preview.Text.At(0)).Should(matchers.MatchText("weave.works/flux: bootstrap"))
				gomega.Eventually(preview.Text.At(0)).Should(matchers.MatchText(fmt.Sprintf(`name: %s\s+namespace: %s`, leafCluster.Name, leafCluster.Namespace)))

				gomega.Eventually(preview.GetPreviewTab("Profiles").Click).Should(gomega.Succeed(), "Failed to switch to 'PROFILES' preview tab")
				gomega.Eventually(preview.GetPreviewTab("Kustomizations").Click).Should(gomega.Succeed(), "Failed to switch to 'KUSTOMIZATION' preview tab")

				gomega.Eventually(preview.Close.Click).Should(gomega.Succeed(), "Failed to close the preview dialog")
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

				authenticateWithGitProvider(webDriver, gitProviderEnv.Type, gitProviderEnv.Hostname)
				gomega.Eventually(deletePR.GitCredentials).Should(matchers.BeVisible())

				gomega.Expect(deletePR.DeleteClusterButton.Click()).To(gomega.Succeed())
			})

			ginkgo.By("Then I should see a toast with a link to the deletion PR", func() {
				messages := pages.GetMessages(webDriver)
				gomega.Eventually(messages.Success, ASSERTION_1MINUTE_TIME_OUT).Should(matchers.MatchText("PR created successfully"), "Failed to create pull request to delete capi cluster")
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
	})
})
