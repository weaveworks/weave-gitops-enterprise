package acceptance

import (
	"fmt"
	"path"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

func DescribeClusters(gitopsTestRunner GitopsTestRunner) {
	var _ = Describe("Multi-Cluster Control Plane Templates", func() {

		Context("[UI] When no leaf cluster is connected", func() {
			It("Verify coneected cluster dashboard shows only management cluster", Label("integration"), func() {
				pages.NavigateToPage(webDriver, "Clusters")
				clustersPage := pages.GetClustersPage(webDriver)
				clustersPage.WaitForPageToLoad(webDriver)

				By("And wait for Clusters page to be rendered", func() {
					Eventually(clustersPage.ClusterHeader).Should(BeVisible())
					Eventually(clustersPage.ClusterCount).Should(MatchText(`1`))
					Expect(pages.CountClusters(clustersPage)).To(Equal(1), "There should not be any cluster in cluster table")
				})

				clusterInfo := pages.FindClusterInList(clustersPage, "management")
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

		Context("[UI] Cluster(s) can be connected connected", func() {
			var mgmtClusterContext string
			leafCluster := "wge-leaf-kind"
			leafClusterNamespace := "default"
			serviceAccountName := "multi-cluster-service"
			leafClusterkubeconfig := leafCluster + "-kubeconfig"

			JustBeforeEach(func() {
				mgmtClusterContext, _ = runCommandAndReturnStringOutput("kubectl config current-context")
				// Create vanilla kind leaf cluster
				createCluster("kind", leafCluster, "")
			})

			JustAfterEach(func() {
				err := runCommandPassThrough("kubectl", "config", "use-context", mgmtClusterContext)
				Expect(err).ShouldNot(HaveOccurred(), "Failed to switch to management cluster context: "+mgmtClusterContext)

				deleteGitopsCluster([]string{leafCluster}, leafClusterNamespace)
				deleteKubeconfigSecret([]string{leafClusterkubeconfig}, leafClusterNamespace)

				deleteClusters("kind", []string{leafCluster})
			})

			It("Verify a cluster can be connected and dashboard is updated accordingly", Label("kind-gitops-cluster", "integration", "browser-logs"), func() {
				pages.NavigateToPage(webDriver, "Clusters")
				clustersPage := pages.GetClustersPage(webDriver)
				clustersPage.WaitForPageToLoad(webDriver)
				existingClustersCount := pages.CountClusters(clustersPage)

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

					containerID, stdErr := runCommandAndReturnStringOutput(fmt.Sprintf(`docker ps | grep %s | tr -s ' '|cut -f1 -d ' '`, leafCluster))
					Expect(stdErr).Should(BeEmpty(), "Failed to get container ID of kind cluster")

					caCertificate := "/tmp/ca.crt"
					err := runCommandPassThrough("sh", "-c", fmt.Sprintf(`docker cp %s:/etc/kubernetes/pki/ca.crt %s`, containerID, caCertificate))
					Expect(err).ShouldNot(HaveOccurred(), "Failed to get CA certificate of kind cluster")
					caAuthority, stdErr := runCommandAndReturnStringOutput(fmt.Sprintf(`cat %s | base64`, caCertificate))
					Expect(stdErr).Should(BeEmpty(), "Failed to get CA Authority of kind cluster")

					controlPlane, stdErr := runCommandAndReturnStringOutput(`kubectl get nodes | grep control-plane | tr -s ' '|cut -f1 -d ' '`)
					Expect(stdErr).Should(BeEmpty(), "Failed to get control plane of kind cluster")

					env := []string{"CLUSTER_NAME=" + leafCluster, "CA_AUTHORITY=" + caAuthority, fmt.Sprintf("ENDPOINT=https://%s:6443", controlPlane), "TOKEN=" + token}
					err = runCommandPassThroughWithEnv(env, "sh", "-c", fmt.Sprintf("%s > /tmp/%s", path.Join(getCheckoutRepoPath(), "test/utils/scripts/static-kubeconfig.sh"), leafClusterkubeconfig))
					Expect(err).ShouldNot(HaveOccurred(), "Failed to create kubeconfig for service account")
				})

				By(fmt.Sprintf("Add GitopsCluster resource for %s cluster to management cluster", leafCluster), func() {
					err := runCommandPassThrough("kubectl", "config", "use-context", mgmtClusterContext)
					Expect(err).ShouldNot(HaveOccurred(), "Failed to switch to management cluster context: "+mgmtClusterContext)

					gitopsCluster, err := generateGitopsClutermanifest(leafCluster, leafClusterNamespace, "bootstrap", leafClusterkubeconfig)
					Expect(err).To(BeNil(), "Failed to generate GitopsCluster manifest yaml")

					Expect(gitopsTestRunner.KubectlApply([]string{}, gitopsCluster), fmt.Sprintf("Failed to create GitopsCluster resource for  cluster: %s", leafCluster))
				})

				By("And wait for GitopsCluster to be visibe on the dashboard", func() {
					Eventually(clustersPage.ClusterHeader).Should(BeVisible())

					totalClusterCount := existingClustersCount + 1
					Eventually(clustersPage.ClusterCount).Should(MatchText(strconv.Itoa(totalClusterCount)), fmt.Sprintf("Dashboard failed to update with expected gitopscluster count: %d", totalClusterCount))
					Eventually(func(g Gomega) int {
						return pages.CountClusters(clustersPage)
					}, ASSERTION_30SECONDS_TIME_OUT).Should(Equal(totalClusterCount), fmt.Sprintf("There should be %d cluster enteries in cluster table", totalClusterCount))
				})

				clusterInfo := pages.FindClusterInList(clustersPage, leafCluster)
				By("And verify GitopsCluster Name", func() {
					Eventually(clusterInfo.Name).Should(MatchText(leafCluster), fmt.Sprintf("Failed to list GitopsCluster in the cluster table: %s", leafCluster))
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

				By("Create secret in management cluster for the generated leaf cluster kubeconfig", func() {
					err := runCommandPassThrough("sh", "-c", fmt.Sprintf(`kubectl create secret generic %[1]v --from-file=value=/tmp/%[1]v -n %s`, leafClusterkubeconfig, leafClusterNamespace))
					Expect(err).ShouldNot(HaveOccurred(), "Failed to create secret for leaf cluster kubeconfig")
				})

				By("Verify GitopsCluster status after creating kubeconfig secret", func() {
					Eventually(clusterInfo.Status).Should(MatchText("Ready"))
				})

				By("Delete Gitops cluster from the management cluster", func() {
					deleteGitopsCluster([]string{leafCluster}, leafClusterNamespace)
				})

				By("And wait for GitopsCluster to disappear from Clusters page", func() {
					Eventually(clustersPage.ClusterCount).Should(MatchText(strconv.Itoa(existingClustersCount)), fmt.Sprintf("Dashboard failed to update with expected gitopscluster count: %d", existingClustersCount))
					Expect(pages.CountClusters(clustersPage)).To(Equal(existingClustersCount), fmt.Sprintf("There should be %d cluster enteries in cluster table", existingClustersCount))
				})
			})
		})
	})
}
