package acceptance

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	"github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

func useClusterContext(clusterContext string) {
	gomega.Expect(runCommandPassThrough("kubectl", "config", "use-context", clusterContext)).ShouldNot(gomega.HaveOccurred(), "Failed to switch to cluster context: "+clusterContext)
}

func createLeafClusterKubeconfig(leafClusterContext string, leafClusterName string, leafClusterNamespace string) string {
	serviceAccountName := "multi-cluster-service"
	leafClusterkubeconfig := leafClusterName + "-kubeconfig"

	currentContext, _ := runCommandAndReturnStringOutput("kubectl config current-context")
	useClusterContext(leafClusterContext)

	ginkgo.By(fmt.Sprintf("Create a service account used for cluster connect: %s", serviceAccountName), func() {
		err := runCommandPassThrough("sh", "-c", fmt.Sprintf(`kubectl create serviceaccount %s`, serviceAccountName))
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to create service account")
	})

	ginkgo.By(fmt.Sprintf("Add RBAC permissions for the service account: %s", serviceAccountName), func() {
		err := runCommandPassThrough("sh", "-c", fmt.Sprintf(`kubectl create clusterrole %s-reader --verb="*" --resource="*.*"`, serviceAccountName))
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to create clusterrole for service account")

		err = runCommandPassThrough("sh", "-c", fmt.Sprintf(`kubectl create clusterrolebinding read-%[1]v --clusterrole=%[1]v-reader --serviceaccount=default:%[1]v --user=kind-%[1]v`, serviceAccountName))
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to create clusterrolebinding for service account")
	})

	ginkgo.By(fmt.Sprintf("And create kubeconfig for the service account: %s", serviceAccountName), func() {
		secret, stdErr := runCommandAndReturnStringOutput(fmt.Sprintf(`kubectl get secrets  --field-selector type=kubernetes.io/service-account-token | grep %s|tr -s ' '|cut -f1 -d ' '`, serviceAccountName))
		gomega.Expect(stdErr).Should(gomega.BeEmpty(), "Failed to get service account secret")

		token, stdErr := runCommandAndReturnStringOutput(fmt.Sprintf(`kubectl get secret %s  -o jsonpath={.data.token} | base64 -d`, secret))
		gomega.Expect(stdErr).Should(gomega.BeEmpty(), "Failed to get service account token")

		containerID, stdErr := runCommandAndReturnStringOutput(fmt.Sprintf(`docker ps | grep %s | tr -s ' '|cut -f1 -d ' '`, leafClusterName))
		gomega.Expect(stdErr).Should(gomega.BeEmpty(), "Failed to get container ID of kind cluster")

		caCertificate := "/tmp/ca.crt"
		err := runCommandPassThrough("sh", "-c", fmt.Sprintf(`docker cp %s:/etc/kubernetes/pki/ca.crt %s`, containerID, caCertificate))
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to get CA certificate of kind cluster")

		contents, err := ioutil.ReadFile(caCertificate)
		gomega.Expect(err).Should(gomega.BeNil(), fmt.Sprintf("Failed to read CA Certificate for %s cluster", leafClusterName))
		caAuthority := base64.StdEncoding.EncodeToString([]byte(contents))

		controlPlane, stdErr := runCommandAndReturnStringOutput(`kubectl get nodes | grep control-plane | tr -s ' '|cut -f1 -d ' '`)
		gomega.Expect(stdErr).Should(gomega.BeEmpty(), "Failed to get control plane of kind cluster")

		env := []string{"CLUSTER_NAME=" + leafClusterName, "CA_AUTHORITY=" + caAuthority, fmt.Sprintf("ENDPOINT=https://%s:6443", controlPlane), "TOKEN=" + token}
		err = runCommandPassThroughWithEnv(env, "sh", "-c", fmt.Sprintf("%s > /tmp/%s", path.Join(getCheckoutRepoPath(), "test/utils/scripts/static-kubeconfig.sh"), leafClusterkubeconfig))
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to create kubeconfig for service account")
	})
	useClusterContext(currentContext)
	return leafClusterkubeconfig
}

func createLeafClusterSecret(leafClusterNamespace string, leafClusterkubeconfig string) {
	ginkgo.By("Create secret in management cluster for the generated leaf cluster kubeconfig", func() {
		err := runCommandPassThrough("sh", "-c", fmt.Sprintf(`kubectl create secret generic %[1]v --from-file=value=/tmp/%[1]v -n %s`, leafClusterkubeconfig, leafClusterNamespace))
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to create secret for leaf cluster kubeconfig")
	})
}

func verifyDashboard(dashboard *agouti.Selection, clusterName string, dashboardName string) {
	ginkgo.By(fmt.Sprintf("And verify %s Cluster dashboard/metada link: %s)", clusterName, dashboardName), func() {
		currentWindow, err := webDriver.Session().GetWindow()
		gomega.Expect(err).To(gomega.BeNil(), "Failed to get weave gitops enterprise dashboard window")

		gomega.Eventually(dashboard).Should(matchers.BeFound(), fmt.Sprintf("Failed to have expected '%s' dashboard for GitopsCluster", dashboardName))
		gomega.Expect(dashboard.Click()).To(gomega.Succeed(), fmt.Sprintf("Failed to navigate to '%s' dashboard", dashboardName)) // opens dashboard in a new tab/window
		gomega.Expect(webDriver.NextWindow()).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to switch to '%s' window", dashboardName))
		gomega.Eventually(webDriver.Title).Should(gomega.MatchRegexp(dashboardName), fmt.Sprintf("Failed to verify '%s' dashboard title", dashboardName))
		gomega.Eventually(webDriver.CloseWindow).Should(gomega.Succeed(), fmt.Sprintf("Failed to close '%s' dashboard window", dashboardName))
		gomega.Expect(webDriver.Session().SetWindow(currentWindow)).ShouldNot(gomega.HaveOccurred(), "Failed to switch to weave gitops enterprise dashboard")
	})
}

func DescribeClusters(gitopsTestRunner GitopsTestRunner) {
	var _ = ginkgo.Describe("Multi-Cluster Control Plane Clusters", func() {

		ginkgo.BeforeEach(func() {
			gomega.Expect(webDriver.Navigate(test_ui_url)).To(gomega.Succeed())

			if !pages.ElementExist(pages.Navbar(webDriver).Title, 3) {
				loginUser()
			}
		})

		ginkgo.Context("[UI] When no leaf cluster is connected", func() {
			ginkgo.It("Verify connected cluster dashboard shows only management cluster", ginkgo.Label("integration"), func() {
				pages.NavigateToPage(webDriver, "Clusters")

				pages.NavigateToPage(webDriver, "Clusters")

				ginkgo.By("And wait for  good looking response from /v1/clusters", func() {
					gomega.Expect(waitForGitopsResources(context.Background(), "clusters", POLL_INTERVAL_15SECONDS)).To(gomega.Succeed(), "Failed to get a successful response from /v1/clusters")
				})

				clustersPage := pages.GetClustersPage(webDriver)
				pages.WaitForPageToLoad(webDriver)

				ginkgo.By("And wait for Clusters page to be rendered", func() {
					gomega.Eventually(clustersPage.ClusterHeader).Should(matchers.BeVisible())
					// gomega.Eventually(clustersPage.ClusterCount).Should(matchers.MatchText(`1`))
					gomega.Expect(clustersPage.CountClusters()).To(gomega.Equal(1), "There should be a single cluster in cluster table")
				})

				clusterInfo := clustersPage.FindClusterInList("management")
				ginkgo.By("And verify GitopsCluster Name", func() {
					gomega.Eventually(clusterInfo.Name).Should(matchers.MatchText("management"), "Failed to list management cluster in the cluster table")
				})

				ginkgo.By("And verify GitopsCluster Type", func() {
					gomega.Eventually(clusterInfo.Type).Should(matchers.BeVisible(), "Failed to have expected management cluster type image/icon")
				})

				// ginkgo.By("And verify GitopsCluster Namespace", func() {
				// 	gomega.Eventually(clusterInfo.Namespace).Should(matchers.MatchText(GITOPS_DEFAULT_NAMESPACE), fmt.Sprintf("Failed to have expected management cluster namespace: %s", GITOPS_DEFAULT_NAMESPACE))
				// })

				ginkgo.By("And verify GitopsCluster status", func() {
					gomega.Eventually(clusterInfo.Status).Should(matchers.MatchText("Ready"), "Failed to have expected management cluster status: Ready")
				})
			})
		})

		ginkgo.Context("[UI] Cluster(s) can be connected", func() {
			var mgmtClusterContext string
			var leafClusterContext string
			var leafClusterkubeconfig string
			var clusterBootstrapCopnfig string
			var gitopsCluster string

			bootstrapLabel := "bootstrap"
			patSecret := "leaf-pat"
			leafClusterName := "wge-leaf-kind"
			leafClusterNamespace := "test-system"
			ClusterLables := []string{"weave.works/flux: bootstrap", "weave.works/apps: backup"}
			downloadedKubeconfigPath := getDownloadedKubeconfigPath(leafClusterName)

			ginkgo.JustBeforeEach(func() {
				mgmtClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
				// Create vanilla kind leaf cluster
				createCluster("kind", leafClusterName, "")
				leafClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
			})

			ginkgo.JustAfterEach(func() {
				useClusterContext(mgmtClusterContext)

				deleteSecret([]string{leafClusterkubeconfig, patSecret}, leafClusterNamespace)
				_ = gitopsTestRunner.KubectlDelete([]string{}, clusterBootstrapCopnfig)
				_ = gitopsTestRunner.KubectlDelete([]string{}, gitopsCluster)

				deleteCluster("kind", leafClusterName, "")
			})

			ginkgo.It("Verify a cluster can be connected and dashboard is updated accordingly", ginkgo.Label("kind-gitops-cluster", "integration", "browser-logs"), func() {
				existingClustersCount := getClustersCount()

				pages.NavigateToPage(webDriver, "Clusters")
				clustersPage := pages.GetClustersPage(webDriver)
				pages.WaitForPageToLoad(webDriver)

				leafClusterkubeconfig = createLeafClusterKubeconfig(leafClusterContext, leafClusterName, leafClusterNamespace)
				useClusterContext(mgmtClusterContext)
				createNamespace([]string{leafClusterNamespace})
				createPATSecret(leafClusterNamespace, patSecret)
				clusterBootstrapCopnfig = createClusterBootstrapConfig(leafClusterName, leafClusterNamespace, bootstrapLabel, patSecret)
				gitopsCluster = connectGitopsCuster(leafClusterName, leafClusterNamespace, bootstrapLabel, leafClusterkubeconfig)

				ginkgo.By("And wait for GitopsCluster to be visibe on the dashboard", func() {
					gomega.Eventually(clustersPage.ClusterHeader).Should(matchers.BeVisible())

					totalClusterCount := existingClustersCount + 1
					// gomega.Eventually(clustersPage.ClusterCount, ASSERTION_1MINUTE_TIME_OUT).Should(matchers.MatchText(strconv.Itoa(totalClusterCount)), fmt.Sprintf("Dashboard failed to update with expected gitopscluster count: %d", totalClusterCount))
					gomega.Eventually(func(g gomega.Gomega) int {
						return clustersPage.CountClusters()
					}, ASSERTION_30SECONDS_TIME_OUT).Should(gomega.Equal(totalClusterCount), fmt.Sprintf("There should be %d cluster enteries in cluster table", totalClusterCount))
				})

				clusterInfo := clustersPage.FindClusterInList(leafClusterName)
				ginkgo.By("And verify GitopsCluster Name", func() {
					gomega.Eventually(clusterInfo.Name).Should(matchers.MatchText(leafClusterName), fmt.Sprintf("Failed to list GitopsCluster in the cluster table: %s", leafClusterName))
				})

				ginkgo.By("And verify GitopsCluster Type", func() {
					gomega.Eventually(clusterInfo.Type).Should(matchers.BeVisible(), "Failed to have expected GitopsCluster type image/icon")
				})

				ginkgo.By("And verify GitopsCluster Namespace", func() {
					gomega.Eventually(clusterInfo.Namespace).Should(matchers.MatchText(leafClusterNamespace), fmt.Sprintf("Failed to have expected GitopsCluster namespace: %s", leafClusterNamespace))
				})

				ginkgo.By("And verify GitopsCluster status", func() {
					gomega.Eventually(clusterInfo.Status).Should(matchers.MatchText("Not Ready"), "Failed to have expected GitopsCluster status: Not Ready")
				})

				createLeafClusterSecret(leafClusterNamespace, leafClusterkubeconfig)

				ginkgo.By("Verify GitopsCluster status after creating kubeconfig secret", func() {
					gomega.Eventually(clusterInfo.Status, ASSERTION_30SECONDS_TIME_OUT).Should(matchers.MatchText("Ready"))
				})

				addKustomizationBases("leaf", leafClusterName, leafClusterNamespace)

				ginkgo.By(fmt.Sprintf("And I verify %s GitopsCluster is bootstraped)", leafClusterName), func() {
					useClusterContext(leafClusterContext)
					verifyFluxControllers(GITOPS_DEFAULT_NAMESPACE)
					waitForGitRepoReady("flux-system", GITOPS_DEFAULT_NAMESPACE)
					useClusterContext(mgmtClusterContext)
				})

				verifyDashboard(clusterInfo.GetDashboard("grafana"), leafClusterName, "Grafana")

				ginkgo.By(fmt.Sprintf("And navigate to '%s' GitopsCluster page", leafClusterName), func() {
					logger.Info(clusterInfo.Name.Text())
					gomega.Eventually(clusterInfo.Name.Find("a").Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to %s GitopsCluster detail page", leafClusterName))
				})

				clusterDetailPage := pages.GetClusterDetailPage(webDriver)
				ginkgo.By(fmt.Sprintf("And verify '%s' cluster page", leafClusterName), func() {
					gomega.Eventually(clusterDetailPage.Header.Text).Should(gomega.MatchRegexp(leafClusterName), "Failed to verify leaf cluster name")

					gomega.Eventually(func(g gomega.Gomega) error {
						g.Expect(clusterDetailPage.Kubeconfig.Click()).To(gomega.Succeed())
						_, err := os.Stat(downloadedKubeconfigPath)
						return err
					}, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to download %s cluster kubeconfig file", leafClusterName))

					gomega.Eventually(clusterDetailPage.Namespace.Text).Should(gomega.MatchRegexp(leafClusterNamespace), "Failed to verify leaf cluster namespace on cluster page")
					TakeScreenShot("prior-dashboard-leaf-cluster")
					verifyDashboard(clusterDetailPage.GetDashboard("prometheus"), leafClusterName, "Prometheus")

					gomega.Expect(clusterDetailPage.GetDashboard("javascript")).ShouldNot(matchers.BeFound(), "XXSVulnerable link shound not be found")
					gomega.Expect(clusterDetailPage.Dashboards.FindByXPath(fmt.Sprintf(`//li[.="%s"]`, "javascript"))).Should(matchers.BeFound(), "Failed to find static Vulnerable label")

					gomega.Expect(clusterDetailPage.GetLabels()).Should(gomega.ConsistOf(ClusterLables), "Failed to verify cluster labels on cluster page")
				})

				ginkgo.By(fmt.Sprintf("And verify '%s' cluster applications", leafClusterName), func() {
					gomega.Eventually(clusterDetailPage.Applications.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to %s cluster's applications page", leafClusterName))

					applicationsPage := pages.GetApplicationsPage(webDriver)
					gomega.Eventually(applicationsPage.ApplicationHeader).Should(matchers.BeVisible())

					expAppCount := 2
					gomega.Eventually(func(g gomega.Gomega) int {
						return applicationsPage.CountApplications()
					}, ASSERTION_3MINUTE_TIME_OUT).Should(gomega.Equal(expAppCount), fmt.Sprintf("There should be %d application enteries in application table", expAppCount))

					appName := "clusters-bases-kustomization"
					appSource := "flux-system"
					applicationInfo := applicationsPage.FindApplicationInList(appName)

					gomega.Eventually(applicationInfo.Name).Should(matchers.MatchText(appName), fmt.Sprintf("Failed to list %s application in  application table", appName))
					gomega.Eventually(applicationInfo.Type).Should(matchers.MatchText("Kustomization"), fmt.Sprintf("Failed to have expected %s application type: Kustomization", appName))
					gomega.Eventually(applicationInfo.Namespace).Should(matchers.MatchText(GITOPS_DEFAULT_NAMESPACE), fmt.Sprintf("Failed to have expected %s application namespace: %s", appName, GITOPS_DEFAULT_NAMESPACE))
					gomega.Eventually(applicationInfo.Cluster).Should(matchers.MatchText(leafClusterNamespace+`/`+leafClusterName), fmt.Sprintf("Failed to have expected violation cluster name: %s", leafClusterNamespace+`/`+leafClusterName))
					gomega.Eventually(applicationInfo.Source).Should(matchers.MatchText("flux-system"), fmt.Sprintf("Failed to have expected %s application source: %s", appName, appSource))
					gomega.Eventually(applicationInfo.Status).Should(matchers.MatchText("Ready"), fmt.Sprintf("Failed to have expected %s application status: Ready", appName))
					gomega.Eventually(applicationInfo.Name.Click).Should(gomega.Succeed(), fmt.Sprintf("Failed to navigate to %s application detail page", appName))

					// Navigate back to clusters page
					pages.NavigateToPage(webDriver, "Clusters")
					pages.WaitForPageToLoad(webDriver)
				})

				ginkgo.By("Verify deleting GitopsCluster resource from the management cluster", func() {
					// Clean up kubeconfig secret, gitopscluster finalizer will wait for it now
					deleteSecret([]string{leafClusterkubeconfig}, leafClusterNamespace)
					gomega.Eventually(clusterInfo.Status).Should(matchers.MatchText("Not Ready"), "Failed to have expected GitopsCluster status: Not Ready")

					_ = gitopsTestRunner.KubectlDelete([]string{}, gitopsCluster)
				})

				ginkgo.By("And wait for GitopsCluster to disappear from Clusters page", func() {
					// gomega.Eventually(clustersPage.ClusterCount, ASSERTION_1MINUTE_TIME_OUT).Should(matchers.MatchText(strconv.Itoa(existingClustersCount)), fmt.Sprintf("Dashboard failed to update with expected gitopscluster count: %d", existingClustersCount))
					gomega.Expect(clustersPage.CountClusters()).To(gomega.Equal(existingClustersCount), fmt.Sprintf("There should be %d cluster enteries in cluster table", existingClustersCount))
				})
			})
		})
	})
}
