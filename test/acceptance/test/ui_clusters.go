package acceptance

import (
	"encoding/base64"
	"fmt"
	"os"
	"path"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	"github.com/sclevine/agouti/matchers"
	"github.com/weaveworks/weave-gitops-enterprise/test/acceptance/test/pages"
)

type ClusterConfig struct {
	Type      string
	Name      string
	Namespace string
}

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

		contents, err := os.ReadFile(caCertificate)
		gomega.Expect(err).Should(gomega.BeNil(), fmt.Sprintf("Failed to read CA Certificate for %s cluster", leafClusterName))
		caAuthority := base64.StdEncoding.EncodeToString([]byte(contents))

		controlPlane, stdErr := runCommandAndReturnStringOutput(`kubectl get nodes | grep control-plane | tr -s ' '|cut -f1 -d ' '`)
		gomega.Expect(stdErr).Should(gomega.BeEmpty(), "Failed to get control plane of kind cluster")

		env := append(os.Environ(), "CLUSTER_NAME="+leafClusterName, "CA_AUTHORITY="+caAuthority, fmt.Sprintf("ENDPOINT=https://%s:6443", controlPlane), "TOKEN="+token)
		err = runCommandPassThroughWithEnv(env, "sh", "-c", fmt.Sprintf("%s > /tmp/%s", path.Join(testScriptsPath, "static-kubeconfig.sh"), leafClusterkubeconfig))
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

func waitForLeafClusterAvailability(leafCluster string, status string) {
	ginkgo.By("Verify GitopsCluster status after kubeconfig secret creation", func() {
		pages.NavigateToPage(webDriver, "Clusters")
		clustersPage := pages.GetClustersPage(webDriver)
		pages.WaitForPageToLoad(webDriver)
		clusterInfo := clustersPage.FindClusterInList(leafCluster)

		gomega.Eventually(clusterInfo.Status, ASSERTION_3MINUTE_TIME_OUT).Should(matchers.MatchText(status), "Failed to have expected leaf Cluster status: Ready")
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
