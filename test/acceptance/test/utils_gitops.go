package acceptance

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"runtime"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func waitForResource(resourceType string, resourceName string, namespace string, kubeconfig string, timeout time.Duration) error {
	pollInterval := 5
	if timeout < 5*time.Second {
		timeout = 5 * time.Second
	}

	if kubeconfig != "" {
		kubeconfig = "--kubeconfig=" + kubeconfig
	}

	timeoutInSeconds := int(timeout.Seconds())
	cmd := fmt.Sprintf("kubectl %s get %s %s -n %s", kubeconfig, resourceType, resourceName, namespace)
	logger.Trace(cmd)
	for i := pollInterval; i < timeoutInSeconds; i += pollInterval {
		logger.Tracef("Waiting for %s in namespace: %s... : %d second(s) passed of %d seconds timeout", resourceType+"/"+resourceName, namespace, i, timeoutInSeconds)
		err := runCommandPassThroughWithoutOutput("sh", "-c", cmd)
		if err == nil {
			stdOut, _ := runCommandAndReturnStringOutput(cmd)

			noResourcesFoundMessage := ""
			if namespace == "default" {
				noResourcesFoundMessage = "No resources found"
			} else {
				noResourcesFoundMessage = fmt.Sprintf("No resources found in %s namespace", namespace)
			}
			if len(stdOut) == 0 || strings.Contains(stdOut, noResourcesFoundMessage) {
				logger.Infof("Got message => {" + noResourcesFoundMessage + "} Continue looking for resource(s)")
			} else {
				return nil
			}
		}
		time.Sleep(time.Duration(pollInterval) * time.Second)
	}
	return fmt.Errorf("error: Failed to find the resource %s of type %s, timeout reached", resourceName, resourceType)
}

func waitForResourceState(state string, statusCondition string, resourceName string, nameSpace string, selector string, kubeconfig string, timeout time.Duration) {
	if kubeconfig != "" {
		kubeconfig = "--kubeconfig=" + kubeconfig
	}

	if selector != "" {
		selector = fmt.Sprintf("--selector='%s'", selector)
	}

	logger.Tracef("Waiting for %s '%s' state in namespace: %s", resourceName, state, nameSpace)

	cmd := fmt.Sprintf(" kubectl wait --for=condition=%s=%s --timeout=%s %s -n %s --all %s %s",
		state, statusCondition, fmt.Sprintf("%.0fs", timeout.Seconds()), resourceName, nameSpace, selector, kubeconfig)
	logger.Trace(cmd)
	_, stdErr := runCommandAndReturnStringOutput(cmd, ASSERTION_6MINUTE_TIME_OUT)
	Expect(stdErr).Should(BeEmpty(), fmt.Sprintf("%s resource has failed to become %s.", resourceName, state))
}

func verifyFluxControllers(namespace string) {
	Expect(waitForResource("deploy", "helm-controller", namespace, "", ASSERTION_2MINUTE_TIME_OUT))
	Expect(waitForResource("deploy", "kustomize-controller", namespace, "", ASSERTION_2MINUTE_TIME_OUT))
	Expect(waitForResource("deploy", "notification-controller", namespace, "", ASSERTION_2MINUTE_TIME_OUT))
	Expect(waitForResource("deploy", "source-controller", namespace, "", ASSERTION_2MINUTE_TIME_OUT))
	Expect(waitForResource("pods", "", namespace, "", ASSERTION_2MINUTE_TIME_OUT))
}

func verifyCoreControllers(namespace string) {
	Expect(waitForResource("pods", "", namespace, "", ASSERTION_2MINUTE_TIME_OUT))

	By("And I wait for the gitops core controllers to be ready", func() {
		waitForResourceState("Ready", "true", "pod", namespace, "app.kubernetes.io/name=weave-gitops", "", ASSERTION_3MINUTE_TIME_OUT)
	})
}

func verifyEnterpriseControllers(releaseName string, mccpPrefix, namespace string) {
	// SOMETIMES (?) (with helm install ./local-path), the mccpPrefix is skipped
	Expect(waitForResource("deploy", releaseName+"-"+mccpPrefix+"event-writer", namespace, "", ASSERTION_2MINUTE_TIME_OUT))
	Expect(waitForResource("deploy", releaseName+"-"+mccpPrefix+"cluster-service", namespace, "", ASSERTION_2MINUTE_TIME_OUT))
	Expect(waitForResource("pods", "", namespace, "", ASSERTION_2MINUTE_TIME_OUT))

	By("And I wait for the gitops enterprise controllers to be ready", func() {
		waitForResourceState("Ready", "true", "pod", namespace, "", "", ASSERTION_3MINUTE_TIME_OUT)
	})
}

func controllerStatus(controllerName, namespace string) error {
	return runCommandPassThroughWithoutOutput("sh", "-c", fmt.Sprintf("kubectl rollout status deployment %s -n %s", controllerName, namespace))
}

func CheckClusterService(capiEndpointURL string) {
	adminPassword := GetEnv("CLUSTER_ADMIN_PASSWORD", "")
	Eventually(func(g Gomega) {
		// login to obtain cookie
		stdOut, _ := runCommandAndReturnStringOutput(
			fmt.Sprintf(
				// insecure for self-signed tls
				`curl --insecure  -d '{"username":"%s","password":"%s"}' -H "Content-Type: application/json" -X POST %s/oauth2/sign_in -c -`,
				AdminUserName, adminPassword, capiEndpointURL,
			),
			ASSERTION_1MINUTE_TIME_OUT,
		)
		g.Expect(stdOut).To(MatchRegexp(`id_token\s*(.*)`), "Failed to fetch cookie/Cluster Service is not healthy")

		re := regexp.MustCompile(`id_token\s*(.*)`)
		match := re.FindAllStringSubmatch(stdOut, -1)
		cookie := match[0][1]
		stdOut, stdErr := runCommandAndReturnStringOutput(
			fmt.Sprintf(
				`curl --insecure --silent --cookie "id_token=%s" -v --output /dev/null --write-out %%{http_code} %s/v1/templates`,
				cookie, capiEndpointURL,
			),
			ASSERTION_1MINUTE_TIME_OUT,
		)
		g.Expect(stdOut).To(MatchRegexp("200"), "Cluster Service is not healthy: %v", stdErr)

	}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(Succeed())
}

func runWegoAddCommand(repoAbsolutePath string, addCommand string, namespace string) {
	logger.Infof("Add command to run: %s in namespace %s from dir %s", addCommand, namespace, repoAbsolutePath)
	_, errOutput := runCommandAndReturnStringOutput(fmt.Sprintf("cd %s && %s %s", repoAbsolutePath, gitops_bin_path, addCommand))
	Expect(errOutput).Should(BeEmpty())
}

func waitForGitRepoReady(appName string, namespace string) {
	waitForResourceState("Ready", "true", "GitRepositories", namespace, "", "", ASSERTION_3MINUTE_TIME_OUT)
	Expect(waitForResource("GitRepositories", appName, namespace, "", ASSERTION_5MINUTE_TIME_OUT)).To(Succeed())
}

func bootstrapAndVerifyFlux(gp GitProviderEnv, gitopsNamespace string, manifestRepoURL string) {
	cmdInstall := fmt.Sprintf(`flux bootstrap %s --owner=%s --repository=%s --branch=main --hostname=%s --path=./clusters/my-cluster`, gp.Type, gp.Org, gp.Repo, gp.Hostname)
	logger.Info(cmdInstall)

	verifyGitRepositories := false
	for i := 1; i < 5; i++ {
		deleteGitopsDeploySecret(gitopsNamespace)
		deleteGitopsGitRepository(gitopsNamespace)
		_, _ = runCommandAndReturnStringOutput(cmdInstall, ASSERTION_5MINUTE_TIME_OUT)
		verifyFluxControllers(gitopsNamespace)

		// Check if GitRepository resource is Ready
		logger.Tracef("Waiting for GitRepositories 'Ready' state in namespace: %s", gitopsNamespace)
		cmdGitRepository := fmt.Sprintf(" kubectl wait --for=condition=Ready --timeout=90s -n %s GitRepositories --all", gitopsNamespace)
		_, stdErr := runCommandAndReturnStringOutput(cmdGitRepository, ASSERTION_2MINUTE_TIME_OUT)
		if stdErr == "" {
			verifyGitRepositories = true
			break
		}
	}
	Expect(verifyGitRepositories).Should(BeTrue(), "GitRepositories resource has failed to become READY.")
}

func removeGitopsCapiClusters(clusternames []string) {
	deleteClusters("capi", clusternames)
}

func listGitopsApplication(appName string, nameSpace string) string {
	var stdOut string
	cmd := fmt.Sprintf("%s get app %s", gitops_bin_path, appName)
	By(fmt.Sprintf("And I run '%s'", cmd), func() {
		stdOut, _ = runCommandAndReturnStringOutput(cmd)
	})
	return stdOut
}

func deleteGitopsGitRepository(nameSpace string) {
	cmd := fmt.Sprintf(`kubectl delete GitRepositories -n %v flux-system`, nameSpace)
	By("And I delete GitRepository resource", func() {
		logger.Trace(cmd)
		_, _ = runCommandAndReturnStringOutput(cmd)
	})
}

func deleteGitopsDeploySecret(nameSpace string) {
	cmd := fmt.Sprintf(`kubectl delete secrets -n %v flux-system`, nameSpace)
	By("And I delete deploy key secret", func() {
		_, _ = runCommandAndReturnStringOutput(cmd)
	})
}

func clusterWorkloadNonePublicIP(clusterKind string) string {
	var expernal_ip string
	if clusterKind == "EKS" || clusterKind == "GKE" {
		node_name, _ := runCommandAndReturnStringOutput(`kubectl get node --selector='!node-role.kubernetes.io/master' -o name | head -n 1`)
		worker_name := strings.Split(node_name, "/")[1]
		expernal_ip, _ = runCommandAndReturnStringOutput(fmt.Sprintf(`kubectl get nodes -o jsonpath="{.items[?(@.metadata.name=='%s')].status.addresses[?(@.type=='ExternalIP')].address}"`, worker_name))
	} else {
		switch runtime.GOOS {
		case "darwin":
			expernal_ip, _ = runCommandAndReturnStringOutput(`ifconfig en0 | grep -i MASK | awk '{print $2}' | cut -f2 -d:`)
		case "linux":
			expernal_ip, _ = runCommandAndReturnStringOutput(`ifconfig eth0 | grep -i MASK | awk '{print $2}' | cut -f2 -d:`)
		}
	}
	return expernal_ip
}

func createCluster(clusterType string, clusterName string, configFile string) {
	if clusterType == "kind" {
		err := runCommandPassThrough("kind", "create", "cluster", "--name", clusterName, "--image=kindest/node:v1.20.7", "--config", "../../utils/data/"+configFile)
		Expect(err).ShouldNot(HaveOccurred())
	} else {
		Fail(fmt.Sprintf("%s cluster type is not supported for test WGE upgrade", clusterType))
	}
}

func deleteClusters(clusterType string, clusters []string) {
	for _, cluster := range clusters {
		if clusterType == "kind" {
			logger.Infof("Deleting cluster: %s", cluster)
			err := runCommandPassThrough("kind", "delete", "cluster", "--name", cluster)
			Expect(err).ShouldNot(HaveOccurred())
		} else {
			err := runCommandPassThrough("kubectl", "get", "cluster", cluster)
			if err == nil {
				logger.Infof("Deleting cluster: %s", cluster)
				err := runCommandPassThrough("kubectl", "delete", "cluster", cluster)
				Expect(err).ShouldNot(HaveOccurred())
				err = runCommandPassThrough("kubectl", "get", "cluster", cluster)
				Expect(err).Should(HaveOccurred(), fmt.Sprintf("Failed to delete cluster %s", cluster))
			}
		}
	}
}

func verifyCapiClusterKubeconfig(kubeconfigPath string, capiCluster string) {
	contents, err := ioutil.ReadFile(kubeconfigPath)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(contents).Should(MatchRegexp(fmt.Sprintf(`context:\s+cluster: %s`, capiCluster)))

	if runtime.GOOS == "darwin" {
		// Point the kubeconfig to the exposed port of the load balancer, rather than the inaccessible container IP.
		_, stdErr := runCommandAndReturnStringOutput(fmt.Sprintf(`sed -i -e "s/server:.*/server: https:\/\/$(docker port %s-lb 6443/tcp | sed "s/0.0.0.0/127.0.0.1/")/g" %s`, capiCluster, kubeconfigPath))
		Expect(stdErr).Should(BeEmpty(), "Failed to delete ClusterBootstrapConfig secret")
	}
}

func verifyCapiClusterHealth(kubeconfigPath string, capiCluster string, profiles []string, namespace string) {

	Expect(waitForResource("nodes", "", "default", kubeconfigPath, ASSERTION_2MINUTE_TIME_OUT))
	waitForResourceState("Ready", "true", "nodes", "default", "", kubeconfigPath, ASSERTION_5MINUTE_TIME_OUT)

	Expect(waitForResource("pods", "", namespace, kubeconfigPath, ASSERTION_2MINUTE_TIME_OUT))
	waitForResourceState("Ready", "true", "pods", namespace, "", kubeconfigPath, ASSERTION_3MINUTE_TIME_OUT)

	for _, profile := range profiles {
		// Check all profiles are installed in layering order
		switch profile {
		case "observability":
			Expect(waitForResource("deploy", capiCluster+"-observability-grafana", namespace, kubeconfigPath, ASSERTION_2MINUTE_TIME_OUT))
			Expect(waitForResource("deploy", capiCluster+"-observability-kube-state-metrics", namespace, kubeconfigPath, ASSERTION_2MINUTE_TIME_OUT))
			waitForResourceState("Ready", "true", "pods", namespace, "release="+capiCluster+"-observability", kubeconfigPath, ASSERTION_3MINUTE_TIME_OUT)
		case "podinfo":
			Expect(waitForResource("deploy", capiCluster+"-podinfo ", namespace, kubeconfigPath, ASSERTION_2MINUTE_TIME_OUT))
			waitForResourceState("Ready", "true", "pods", namespace, "app.kubernetes.io/name="+capiCluster+"-podinfo", kubeconfigPath, ASSERTION_3MINUTE_TIME_OUT)
		}
	}
}
