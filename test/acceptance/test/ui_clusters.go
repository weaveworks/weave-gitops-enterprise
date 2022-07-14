package acceptance

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

func useClusterContext(clusterContext string) {
	Expect(runCommandPassThrough("kubectl", "config", "use-context", clusterContext)).ShouldNot(HaveOccurred(), "Failed to switch to cluster context: "+clusterContext)
}

func createLeafClusterKubeconfig(leafClusterContext string, leafClusterName string, leafClusterNamespace string) string {
	serviceAccountName := "multi-cluster-service"
	leafClusterkubeconfig := leafClusterName + "-kubeconfig"

	currentContext, _ := runCommandAndReturnStringOutput("kubectl config current-context")
	useClusterContext(leafClusterContext)

	By(fmt.Sprintf("Create a service account used for cluster connect: %s", serviceAccountName), func() {
		err := runCommandPassThrough("sh", "-c", fmt.Sprintf(`kubectl create serviceaccount %s`, serviceAccountName))
		Expect(err).ShouldNot(HaveOccurred(), "Failed to create service account")
	})

	By(fmt.Sprintf("Add RBAC permissions for the service account: %s", serviceAccountName), func() {
		err := runCommandPassThrough("sh", "-c", fmt.Sprintf(`kubectl create clusterrole %s-reader --verb="*" --resource="*.*"`, serviceAccountName))
		Expect(err).ShouldNot(HaveOccurred(), "Failed to create clusterrole for service account")

		err = runCommandPassThrough("sh", "-c", fmt.Sprintf(`kubectl create clusterrolebinding read-%[1]v --clusterrole=%[1]v-reader --serviceaccount=default:%[1]v --user=kind-%[1]v`, serviceAccountName))
		Expect(err).ShouldNot(HaveOccurred(), "Failed to create clusterrolebinding for service account")
	})

	By(fmt.Sprintf("And create kubeconfig for the service account: %s", serviceAccountName), func() {
		secret, stdErr := runCommandAndReturnStringOutput(fmt.Sprintf(`kubectl get secrets  --field-selector type=kubernetes.io/service-account-token | grep %s|tr -s ' '|cut -f1 -d ' '`, serviceAccountName))
		Expect(stdErr).Should(BeEmpty(), "Failed to get service account secret")

		token, stdErr := runCommandAndReturnStringOutput(fmt.Sprintf(`kubectl get secret %s  -o jsonpath={.data.token} | base64 -d`, secret))
		Expect(stdErr).Should(BeEmpty(), "Failed to get service account token")

		containerID, stdErr := runCommandAndReturnStringOutput(fmt.Sprintf(`docker ps | grep %s | tr -s ' '|cut -f1 -d ' '`, leafClusterName))
		Expect(stdErr).Should(BeEmpty(), "Failed to get container ID of kind cluster")

		caCertificate := "/tmp/ca.crt"
		err := runCommandPassThrough("sh", "-c", fmt.Sprintf(`docker cp %s:/etc/kubernetes/pki/ca.crt %s`, containerID, caCertificate))
		Expect(err).ShouldNot(HaveOccurred(), "Failed to get CA certificate of kind cluster")

		contents, err := ioutil.ReadFile(caCertificate)
		Expect(err).Should(BeNil(), fmt.Sprintf("Failed to read CA Certificate for %s cluster", leafClusterName))
		caAuthority := base64.StdEncoding.EncodeToString([]byte(contents))

		controlPlane, stdErr := runCommandAndReturnStringOutput(`kubectl get nodes | grep control-plane | tr -s ' '|cut -f1 -d ' '`)
		Expect(stdErr).Should(BeEmpty(), "Failed to get control plane of kind cluster")

		env := []string{"CLUSTER_NAME=" + leafClusterName, "CA_AUTHORITY=" + caAuthority, fmt.Sprintf("ENDPOINT=https://%s:6443", controlPlane), "TOKEN=" + token}
		err = runCommandPassThroughWithEnv(env, "sh", "-c", fmt.Sprintf("%s > /tmp/%s", path.Join(getCheckoutRepoPath(), "test/utils/scripts/static-kubeconfig.sh"), leafClusterkubeconfig))
		Expect(err).ShouldNot(HaveOccurred(), "Failed to create kubeconfig for service account")
	})
	useClusterContext(currentContext)
	return leafClusterkubeconfig
}

func createLeafClusterSecret(leafClusterNamespace string, leafClusterkubeconfig string) {
	By("Create secret in management cluster for the generated leaf cluster kubeconfig", func() {
		err := runCommandPassThrough("sh", "-c", fmt.Sprintf(`kubectl create secret generic %[1]v --from-file=value=/tmp/%[1]v -n %s`, leafClusterkubeconfig, leafClusterNamespace))
		Expect(err).ShouldNot(HaveOccurred(), "Failed to create secret for leaf cluster kubeconfig")
	})
}

func verifyDashboard(dashboard *agouti.Selection, clusterName string, dashboardName string) {
	By(fmt.Sprintf("And verify %s GitopsCluster dashboard: %s)", clusterName, dashboardName), func() {
		Eventually(dashboard).Should(BeFound(), fmt.Sprintf("Failed to have expected '%s' dashboard for GitopsCluster", dashboardName))
		Expect(dashboard.Click()).To(Succeed(), fmt.Sprintf("Failed to navigate to '%s' dashboard", dashboardName)) // opens dashboard in a new tab/window
		Expect(webDriver.NextWindow()).ShouldNot(HaveOccurred(), fmt.Sprintf("Failed to switch to '%s' window", dashboardName))
		Eventually(webDriver.Title).Should(MatchRegexp(dashboardName), fmt.Sprintf("Failed to verify '%s' dashboard title", dashboardName))
		Eventually(webDriver.CloseWindow).Should(Succeed(), fmt.Sprintf("Failed to close '%s' dashboard window", dashboardName))
		Expect(webDriver.SwitchToWindow(WGE_WINDOW_NAME)).ShouldNot(HaveOccurred(), "Failed to switch to weave gitops enterprise dashboard")
	})
}

func DescribeClusters(gitopsTestRunner GitopsTestRunner) {
	var _ = Describe("Multi-Cluster Control Plane Clusters", func() {

		Context("[UI] When no leaf cluster is connected", func() {
			It("Verify connected cluster dashboard shows only management cluster", Label("integration"), func() {
				pages.NavigateToPage(webDriver, "Clusters")
				clustersPage := pages.GetClustersPage(webDriver)
				pages.WaitForPageToLoad(webDriver)

				By("And wait for Clusters page to be rendered", func() {
					Eventually(clustersPage.ClusterHeader).Should(BeVisible())
					Eventually(clustersPage.ClusterCount).Should(MatchText(`1`))
					Expect(clustersPage.CountClusters()).To(Equal(1), "There should be a single cluster in cluster table")
				})

				clusterInfo := clustersPage.FindClusterInList("management")
				By("And verify GitopsCluster Name", func() {
					Eventually(clusterInfo.Name).Should(MatchText("management"), "Failed to list management cluster in the cluster table")
				})

				By("And verify GitopsCluster Type", func() {
					Eventually(clusterInfo.Type).Should(MatchText("other"), "Failed to have expected management cluster type: other")
				})

				// By("And verify GitopsCluster Namespace", func() {
				// 	Eventually(clusterInfo.Namespace).Should(MatchText(GITOPS_DEFAULT_NAMESPACE), fmt.Sprintf("Failed to have expected management cluster namespace: %s", GITOPS_DEFAULT_NAMESPACE))
				// })

				By("And verify GitopsCluster status", func() {
					Eventually(clusterInfo.Status).Should(MatchText("Ready"), "Failed to have expected management cluster status: Ready")
				})
			})
		})

		Context("[UI] Cluster(s) can be connected", func() {
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

			JustBeforeEach(func() {
				mgmtClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
				// Create vanilla kind leaf cluster
				createCluster("kind", leafClusterName, "")
				leafClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
			})

			JustAfterEach(func() {
				useClusterContext(mgmtClusterContext)

				deleteSecret([]string{leafClusterkubeconfig, patSecret}, leafClusterNamespace)
				_ = gitopsTestRunner.KubectlDelete([]string{}, clusterBootstrapCopnfig)
				_ = gitopsTestRunner.KubectlDelete([]string{}, gitopsCluster)

				deleteClusters("kind", []string{leafClusterName}, "")
			})

			It("Verify a cluster can be connected and dashboard is updated accordingly", Label("kind-gitops-cluster", "integration", "browser-logs"), func() {
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

				By("And wait for GitopsCluster to be visibe on the dashboard", func() {
					Eventually(clustersPage.ClusterHeader).Should(BeVisible())

					totalClusterCount := existingClustersCount + 1
					Eventually(clustersPage.ClusterCount, ASSERTION_1MINUTE_TIME_OUT).Should(MatchText(strconv.Itoa(totalClusterCount)), fmt.Sprintf("Dashboard failed to update with expected gitopscluster count: %d", totalClusterCount))
					Eventually(func(g Gomega) int {
						return clustersPage.CountClusters()
					}, ASSERTION_30SECONDS_TIME_OUT).Should(Equal(totalClusterCount), fmt.Sprintf("There should be %d cluster enteries in cluster table", totalClusterCount))
				})

				clusterInfo := clustersPage.FindClusterInList(leafClusterName)
				By("And verify GitopsCluster Name", func() {
					Eventually(clusterInfo.Name).Should(MatchText(leafClusterName), fmt.Sprintf("Failed to list GitopsCluster in the cluster table: %s", leafClusterName))
				})

				By("And verify GitopsCluster Type", func() {
					Eventually(clusterInfo.Type).Should(MatchText("other"), "Failed to have expected GitopsCluster type: other")
				})

				By("And verify GitopsCluster Namespace", func() {
					Eventually(clusterInfo.Namespace).Should(MatchText(leafClusterNamespace), fmt.Sprintf("Failed to have expected GitopsCluster namespace: %s", leafClusterNamespace))
				})

				By("And verify GitopsCluster status", func() {
					Eventually(clusterInfo.Status).Should(MatchText("Not Ready"), "Failed to have expected GitopsCluster status: Not Ready")
				})

				createLeafClusterSecret(leafClusterNamespace, leafClusterkubeconfig)

				By("Verify GitopsCluster status after creating kubeconfig secret", func() {
					Eventually(clusterInfo.Status, ASSERTION_30SECONDS_TIME_OUT).Should(MatchText("Ready"))
				})

				By("And add kustomization bases for common resources for leaf cluster)", func() {
					addKustomizationBases(leafClusterName, leafClusterNamespace)
				})

				By(fmt.Sprintf("And I verify %s GitopsCluster is bootstraped)", leafClusterName), func() {
					useClusterContext(leafClusterContext)
					verifyFluxControllers(GITOPS_DEFAULT_NAMESPACE)
					waitForGitRepoReady("flux-system", GITOPS_DEFAULT_NAMESPACE)
					useClusterContext(mgmtClusterContext)
				})

				verifyDashboard(clusterInfo.GetDashboard("grafana"), leafClusterName, "Grafana")

				By(fmt.Sprintf("And navigate to '%s' GitopsCluster page", leafClusterName), func() {
					logger.Info(clusterInfo.Name.Text())
					Eventually(clusterInfo.Name.Click).Should(Succeed(), fmt.Sprintf("Failed to navigate to %s GitopsCluster detail page", leafClusterName))
				})

				clusterDetailPage := pages.GetClusterDetailPage(webDriver)
				By(fmt.Sprintf("And verify '%s' cluster page", leafClusterName), func() {
					Eventually(clusterDetailPage.Header.Text).Should(MatchRegexp(leafClusterName), "Failed to verify leaf cluster name")

					Eventually(clusterDetailPage.Kubeconfig.Text).Should(MatchRegexp("Download the kubeconfig here"), "Failed to verify download kubeconfig link.button on cluster page")
					Eventually(func(g Gomega) error {
						g.Expect(clusterDetailPage.Kubeconfig.Click()).To(Succeed())
						_, err := os.Stat(downloadedKubeconfigPath)
						return err
					}, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).ShouldNot(HaveOccurred(), fmt.Sprintf("Failed to download %s cluster kubeconfig file", leafClusterName))

					Eventually(clusterDetailPage.Namespace.Text).Should(MatchRegexp(leafClusterNamespace), "Failed to verify leaf cluster namespace on cluster page")
					verifyDashboard(clusterDetailPage.GetDashboard("prometheus"), leafClusterName, "Prometheus")
					Expect(clusterDetailPage.GetLabels()).Should(ConsistOf(ClusterLables), "Failed to verify cluster labels on cluster page")
				})

				By(fmt.Sprintf("And verify '%s' cluster applications", leafClusterName), func() {
					Eventually(clusterDetailPage.Applications.Click).Should(Succeed(), fmt.Sprintf("Failed to navigate to %s cluster's applications page", leafClusterName))

					applicationsPage := pages.GetApplicationsPage(webDriver)
					Eventually(applicationsPage.ApplicationHeader).Should(BeVisible())

					expAppCount := 2
					Eventually(func(g Gomega) int {
						return applicationsPage.CountApplications()
					}, ASSERTION_3MINUTE_TIME_OUT).Should(Equal(expAppCount), fmt.Sprintf("There should be %d application enteries in application table", expAppCount))

					appName := "clusters-bases-kustomization"
					appSource := "flux-system"
					applicationInfo := applicationsPage.FindApplicationInList(appName)

					Eventually(applicationInfo.Name).Should(MatchText(appName), fmt.Sprintf("Failed to list %s application in  application table", appName))
					Eventually(applicationInfo.Type).Should(MatchText("Kustomization"), fmt.Sprintf("Failed to have expected %s application type: Kustomization", appName))
					Eventually(applicationInfo.Namespace).Should(MatchText(GITOPS_DEFAULT_NAMESPACE), fmt.Sprintf("Failed to have expected %s application namespace: %s", appName, GITOPS_DEFAULT_NAMESPACE))
					Eventually(applicationInfo.Cluster).Should(MatchText(leafClusterNamespace+`/`+leafClusterName), fmt.Sprintf("Failed to have expected violation cluster name: %s", leafClusterNamespace+`/`+leafClusterName))
					Eventually(applicationInfo.Source).Should(MatchText("flux-system"), fmt.Sprintf("Failed to have expected %s application source: %s", appName, appSource))
					Eventually(applicationInfo.Status).Should(MatchText("Ready"), fmt.Sprintf("Failed to have expected %s application status: Ready", appName))
					Eventually(applicationInfo.Name.Click).Should(Succeed(), fmt.Sprintf("Failed to navigate to %s application detail page", appName))

					// Navigate back to clusters page
					pages.NavigateToPage(webDriver, "Clusters")
					pages.WaitForPageToLoad(webDriver)
				})

				By("Verify deleting GitopsCluster resource from the management cluster", func() {
					// Clean up kubeconfig secret, gitopscluster finalizer will wait for it now
					deleteSecret([]string{leafClusterkubeconfig}, leafClusterNamespace)
					Eventually(clusterInfo.Status).Should(MatchText("Not Ready"), "Failed to have expected GitopsCluster status: Not Ready")

					_ = gitopsTestRunner.KubectlDelete([]string{}, gitopsCluster)
				})

				By("And wait for GitopsCluster to disappear from Clusters page", func() {
					Eventually(clustersPage.ClusterCount, ASSERTION_1MINUTE_TIME_OUT).Should(MatchText(strconv.Itoa(existingClustersCount)), fmt.Sprintf("Dashboard failed to update with expected gitopscluster count: %d", existingClustersCount))
					Expect(clustersPage.CountClusters()).To(Equal(existingClustersCount), fmt.Sprintf("There should be %d cluster enteries in cluster table", existingClustersCount))
				})
			})
		})
	})
}
