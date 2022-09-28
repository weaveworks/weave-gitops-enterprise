package acceptance

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/wait"
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
	gomega.Expect(stdErr).Should(gomega.BeEmpty(), fmt.Sprintf("%s resource has failed to become %s.", resourceName, state))
}

func verifyFluxControllers(namespace string) {
	gomega.Expect(waitForResource("deploy", "helm-controller", namespace, "", ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())
	gomega.Expect(waitForResource("deploy", "kustomize-controller", namespace, "", ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())
	gomega.Expect(waitForResource("deploy", "notification-controller", namespace, "", ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())
	gomega.Expect(waitForResource("deploy", "source-controller", namespace, "", ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())
	gomega.Expect(waitForResource("pods", "", namespace, "", ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())
}

func verifyCoreControllers(namespace string) {
	gomega.Expect(waitForResource("pods", "", namespace, "", ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())

	ginkgo.By("And I wait for the gitops core controllers to be ready", func() {
		waitForResourceState("Ready", "true", "pod", namespace, "app.kubernetes.io/name=weave-gitops", "", ASSERTION_3MINUTE_TIME_OUT)
	})
}

func verifyEnterpriseControllers(releaseName string, mccpPrefix, namespace string) {
	// SOMETIMES (?) (with helm install ./local-path), the mccpPrefix is skipped
	gomega.Expect(waitForResource("deploy", releaseName+"-"+mccpPrefix+"cluster-service", namespace, "", ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())
	gomega.Expect(waitForResource("pods", "", namespace, "", ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())

	ginkgo.By("And I wait for the gitops enterprise controllers to be ready", func() {
		waitForResourceState("Ready", "true", "pod", namespace, "", "", ASSERTION_3MINUTE_TIME_OUT)
	})
}

func controllerStatus(controllerName, namespace string) error {
	return runCommandPassThroughWithoutOutput("sh", "-c", fmt.Sprintf("kubectl rollout status deployment %s -n %s", controllerName, namespace))
}

func CheckClusterService(capiEndpointURL string) {
	adminPassword := GetEnv("CLUSTER_ADMIN_PASSWORD", "")
	gomega.Eventually(func(g gomega.Gomega) {
		logger.Info("Trying to login to cluster service")
		// login to obtain cookie
		stdOut, _ := runCommandAndReturnStringOutput(
			fmt.Sprintf(
				// insecure for self-signed tls
				`curl --insecure  -d '{"username":"%s","password":"%s"}' -H "Content-Type: application/json" -X POST %s/oauth2/sign_in -c -`,
				AdminUserName, adminPassword, capiEndpointURL,
			),
			ASSERTION_1MINUTE_TIME_OUT,
		)
		g.Expect(stdOut).To(gomega.MatchRegexp(`id_token\s*(.*)`), "Failed to fetch cookie/Cluster Service is not healthy")

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
		g.Expect(stdOut).To(gomega.MatchRegexp("200"), "Cluster Service is not healthy: %v", stdErr)

	}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Succeed())
}

// Wait until we get a good looking response from /v1/<resource>
// Ignore all errors (connection refused, 500s etc)
func waitForGitopsResources(ctx context.Context, resourcePath string, timeout time.Duration) error {
	adminPassword := GetEnv("CLUSTER_ADMIN_PASSWORD", "")
	waitCtx, cancel := context.WithTimeout(ctx, ASSERTION_5MINUTE_TIME_OUT)
	defer cancel()

	return wait.PollUntil(time.Second*1, func() (bool, error) {
		jar, _ := cookiejar.New(&cookiejar.Options{})
		client := http.Client{
			Timeout: timeout,
			Jar:     jar,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
		// login to fetch cookie
		resp, err := client.Post(test_ui_url+"/oauth2/sign_in", "application/json", bytes.NewReader([]byte(fmt.Sprintf(`{"username":"%s", "password":"%s"}`, AdminUserName, adminPassword))))
		if err != nil {
			logger.Tracef("error logging in (waiting for a success, retrying): %v", err)
			return false, nil
		}
		if resp.StatusCode != http.StatusOK {
			logger.Tracef("wrong status from login (waiting for a ok, retrying): %v", resp.StatusCode)
			return false, nil
		}
		// fetch gitops resource
		resp, err = client.Get(test_ui_url + "/v1/" + resourcePath)
		if err != nil {
			logger.Tracef("error getting %s in (waiting for a success, retrying): %v", resourcePath, err)
			return false, nil
		}
		if resp.StatusCode != http.StatusOK {
			logger.Tracef("wrong status from %s (waiting for a ok, retrying): %v", resourcePath, resp.StatusCode)
			return false, nil
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, nil
		}
		bodyString := string(bodyBytes)

		// FIXME: could this be a 200 instead?
		parseUrl, err := url.Parse(resourcePath)
		if err != nil {
			logger.Errorf("failed to parse URL: %v", resourcePath)
			return false, nil
		}
		resourceName := parseUrl.Path
		return strings.Contains(strings.ToLower(bodyString), strings.ToLower(fmt.Sprintf(`%s":`, resourceName))), nil
	}, waitCtx.Done())
}

func runGitopsCommand(cmd string, timeout ...time.Duration) (stdOut, stdErr string) {
	// Using self signed certs, all `gitops get clusters` etc commands should use insecure tls connections
	insecureFlag := "--insecure-skip-tls-verify"
	var authFlag string

	// // Login via cluster user account (basic authentication)
	authFlag = fmt.Sprintf("--username %s --password %s", userCredentials.ClusterUserName, userCredentials.ClusterUserPassword)
	if mgmtClusterKind != KindMgmtCluster {
		switch userCredentials.UserType {
		case ClusterUserLogin:
			if mgmtClusterKind == GKEMgmtCluster {
				authFlag = "" // Login via native cluster admin/token
			}
		case OidcUserLogin:
			authFlag = fmt.Sprintf("--kubeconfig=%s", userCredentials.UserKubeconfig)
		default:
			gomega.Expect(fmt.Errorf("error: Provided authento=ication type '%s' is not supported for CLI", userCredentials.UserType))
		}
	}

	cmd = fmt.Sprintf(`%s --endpoint %s %s %s %s`, gitops_bin_path, capi_endpoint_url, insecureFlag, authFlag, cmd)
	ginkgo.By(fmt.Sprintf(`And I run '%s'`, cmd), func() {
		assert_timeout := ASSERTION_DEFAULT_TIME_OUT
		if len(timeout) > 0 {
			assert_timeout = timeout[0]
		}
		stdOut, stdErr = runCommandAndReturnStringOutput(cmd, assert_timeout)
	})

	return stdOut, stdErr
}

func waitForGitRepoReady(appName string, namespace string) {
	gomega.Expect(waitForResource("GitRepositories", appName, namespace, "", ASSERTION_5MINUTE_TIME_OUT)).To(gomega.Succeed())
	waitForResourceState("Ready", "true", "GitRepositories", namespace, "", "", ASSERTION_3MINUTE_TIME_OUT)
}

func bootstrapAndVerifyFlux(gp GitProviderEnv, gitopsNamespace string, manifestRepoURL string) {
	cmdInstall := fmt.Sprintf(`flux bootstrap %s --owner=%s --repository=%s --branch=main --hostname=%s --path=./clusters/management`, gp.Type, gp.Org, gp.Repo, gp.Hostname)
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
	gomega.Expect(verifyGitRepositories).Should(gomega.BeTrue(), "GitRepositories resource has failed to become READY.")
}

func suspendReconciliation(sourceType string, sourceName string, namespace string) {
	cmdSuspend := fmt.Sprintf("flux suspend source %s %s --namespace %s", sourceType, sourceName, namespace)
	_, _ = runCommandAndReturnStringOutput(cmdSuspend, ASSERTION_30SECONDS_TIME_OUT)

}

func resumeReconciliation(sourceType string, sourceName string, namespace string) {
	cmdSuspend := fmt.Sprintf("flux resume source %s %s --namespace %s", sourceType, sourceName, namespace)
	_, _ = runCommandAndReturnStringOutput(cmdSuspend, ASSERTION_30SECONDS_TIME_OUT)
}

func removeGitopsCapiClusters(capiClusters []ClusterConfig) {
	for _, cluster := range capiClusters {
		deleteCluster(cluster.Type, cluster.Name, cluster.Namespace)
	}
}

func deleteGitopsGitRepository(nameSpace string) {
	cmd := fmt.Sprintf(`kubectl delete GitRepositories -n %v flux-system`, nameSpace)
	ginkgo.By("And I delete GitRepository resource", func() {
		logger.Trace(cmd)
		_, _ = runCommandAndReturnStringOutput(cmd)
	})
}

func deleteGitopsDeploySecret(nameSpace string) {
	cmd := fmt.Sprintf(`kubectl delete secrets -n %v flux-system`, nameSpace)
	ginkgo.By("And I delete deploy key secret", func() {
		_, _ = runCommandAndReturnStringOutput(cmd)
	})
}

func deleteSecret(kubeconfigSecrets []string, nameSpace string) {
	for _, secret := range kubeconfigSecrets {
		err := runCommandPassThrough("sh", "-c", fmt.Sprintf(`kubectl get secret %s -n %s`, secret, nameSpace))
		if err == nil {
			_, _ = runCommandAndReturnStringOutput(fmt.Sprintf(`kubectl delete secret %s -n %s`, secret, nameSpace))
		}
	}
}

func createCluster(clusterType string, clusterName string, configFile string) {
	if clusterType == "kind" {
		if configFile != "" {
			configFile = "--config " + path.Join(getCheckoutRepoPath(), "test/utils/data", configFile)
		}
		err := runCommandPassThrough("sh", "-c", fmt.Sprintf("kind create cluster --name %s --image=kindest/node:v1.23.4 %s", clusterName, configFile))
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		gomega.Expect(waitForResource("pods", "", "kube-system", "", ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())
		waitForResourceState("Ready", "true", "pods", "kube-system", "", "", ASSERTION_2MINUTE_TIME_OUT)
	} else {
		ginkgo.Fail(fmt.Sprintf("%s cluster type is not supported", clusterType))
	}
}

func deleteCluster(clusterType string, cluster string, nameSpace string) {
	if clusterType == "kind" {
		logger.Infof("Deleting cluster: %s", cluster)
		err := runCommandPassThrough("kind", "delete", "cluster", "--name", cluster)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	} else {
		err := runCommandPassThrough("kubectl", "get", "cluster", cluster, "-n", nameSpace)
		if err == nil {
			logger.Infof("Deleting cluster %s in namespace %s", cluster, nameSpace)
			err := runCommandPassThrough("kubectl", "delete", "cluster", cluster, "-n", nameSpace)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			err = runCommandPassThrough("kubectl", "get", "cluster", cluster, "-n", nameSpace)
			gomega.Expect(err).Should(gomega.HaveOccurred(), fmt.Sprintf("Failed to delete cluster %s", cluster))
		}
	}
}

func verifyCapiClusterKubeconfig(kubeconfigPath string, capiCluster string) {
	contents, err := ioutil.ReadFile(kubeconfigPath)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Eventually(contents).Should(gomega.MatchRegexp(fmt.Sprintf(`context:\s+cluster: %s`, capiCluster)))

	if runtime.GOOS == "darwin" {
		// Point the kubeconfig to the exposed port of the load balancer, rather than the inaccessible container IP.
		_, stdErr := runCommandAndReturnStringOutput(fmt.Sprintf(`sed -i -e "s/server:.*/server: https:\/\/$(docker port %s-lb 6443/tcp | sed "s/0.0.0.0/127.0.0.1/")/g" %s`, capiCluster, kubeconfigPath))
		gomega.Expect(stdErr).Should(gomega.BeEmpty(), "Failed to delete ClusterBootstrapConfig secret")
	}
}

func verifyCapiClusterHealth(kubeconfigPath string, applications []Application) {

	gomega.Expect(waitForResource("nodes", "", "default", kubeconfigPath, ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())
	waitForResourceState("Ready", "true", "nodes", "default", "", kubeconfigPath, ASSERTION_5MINUTE_TIME_OUT)

	gomega.Expect(waitForResource("pods", "", GITOPS_DEFAULT_NAMESPACE, kubeconfigPath, ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())
	waitForResourceState("Ready", "true", "pods", GITOPS_DEFAULT_NAMESPACE, "", kubeconfigPath, ASSERTION_3MINUTE_TIME_OUT)

	for _, app := range applications {
		// Check all profiles are installed in layering order
		switch app.Name {
		case "observability":
			gomega.Expect(waitForResource("deploy", "observability-grafana", app.TargetNamespace, kubeconfigPath, ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())
			gomega.Expect(waitForResource("deploy", "observability-kube-state-metrics", app.TargetNamespace, kubeconfigPath, ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())
			waitForResourceState("Ready", "true", "pods", app.TargetNamespace, "release="+"observability", kubeconfigPath, ASSERTION_3MINUTE_TIME_OUT)
		case "podinfo":
			gomega.Expect(waitForResource("deploy", "podinfo ", app.TargetNamespace, kubeconfigPath, ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())
			waitForResourceState("Ready", "true", "pods", app.TargetNamespace, "app.kubernetes.io/name="+"podinfo", kubeconfigPath, ASSERTION_3MINUTE_TIME_OUT)
		case "metallb":
			gomega.Expect(waitForResource("deploy", "metallb-controller ", app.TargetNamespace, kubeconfigPath, ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())
			waitForResourceState("Ready", "true", "pods", app.TargetNamespace, "app.kubernetes.io/name="+"metallb", kubeconfigPath, ASSERTION_3MINUTE_TIME_OUT)
		case "cert-manager":
			gomega.Expect(waitForResource("deploy", "cert-manager-cert-manager", app.TargetNamespace, kubeconfigPath, ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())
			waitForResourceState("Ready", "true", "pods", app.TargetNamespace, "", kubeconfigPath, ASSERTION_3MINUTE_TIME_OUT)
		case "weave-policy-agent":
			gomega.Expect(waitForResource("deploy", "policy-agent", app.TargetNamespace, kubeconfigPath, ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())
			waitForResourceState("Ready", "true", "pods", app.TargetNamespace, "", kubeconfigPath, ASSERTION_3MINUTE_TIME_OUT)
		}
	}
}

func createPATSecret(clusterNamespace string, patSecret string) {
	ginkgo.By("Create personal access token secret in management cluster for ClusterBootstrapConfig", func() {
		tokenType := "GITHUB_TOKEN"
		if gitProviderEnv.Type != GitProviderGitHub {
			tokenType = "GITLAB_TOKEN"
		}

		err := runCommandPassThrough("sh", "-c", fmt.Sprintf(`kubectl create secret generic %s --from-literal %s=%s -n %s`, patSecret, tokenType, gitProviderEnv.Token, clusterNamespace))
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to create personal access token secret for ClusterBootstrapConfig")
	})
}

func createClusterResourceSet(clusterName string, nameSpace string) (resourceSet string) {
	ginkgo.By(fmt.Sprintf("Add ClusterResourceSet resource for %s cluster to management cluster", clusterName), func() {
		contents, err := ioutil.ReadFile(path.Join(getCheckoutRepoPath(), "test/utils/data/calico-crs.yaml"))
		gomega.Expect(err).To(gomega.BeNil(), "Failed to read calico-crs template yaml")

		t := template.Must(template.New("cluster-resource-set").Parse(string(contents)))

		// Prepare  data to insert into the template.
		type TemplateInput struct {
			Name      string
			NameSpace string
		}
		input := TemplateInput{clusterName, nameSpace}

		resourceSet = path.Join("/tmp", clusterName+"-calico-crs.yaml")

		f, err := os.Create(resourceSet)
		gomega.Expect(err).To(gomega.BeNil(), "Failed to create ClusterResourceSet manifest yaml")

		err = t.Execute(f, input)
		f.Close()
		gomega.Expect(err).To(gomega.BeNil(), "Failed to generate ClusterResourceSet manifest yaml")

		err = runCommandPassThrough("kubectl", "apply", "-f", resourceSet)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to create ClusterResourceSet resource for  cluster: %s", clusterName))
	})
	return resourceSet
}

func createCRSConfigmap(clusterName string, nameSpace string) (configmap string) {
	ginkgo.By(fmt.Sprintf("Add ClusterResourceSet configmap resource for %s cluster to management cluster", clusterName), func() {
		contents, err := ioutil.ReadFile(path.Join(getCheckoutRepoPath(), "test/utils/data/calico-crs-configmap.yaml"))
		gomega.Expect(err).To(gomega.BeNil(), "Failed to read calico-crs-configmap template yaml")

		t := template.Must(template.New("crs-configmap").Parse(string(contents)))

		// Prepare  data to insert into the template.
		type TemplateInput struct {
			Name      string
			NameSpace string
		}
		input := TemplateInput{clusterName, nameSpace}

		configmap = path.Join("/tmp", clusterName+"-calico-crs-configmap.yaml")

		f, err := os.Create(configmap)
		gomega.Expect(err).To(gomega.BeNil(), "Failed to create calico-crs-configmap manifest yaml")

		err = t.Execute(f, input)
		f.Close()
		gomega.Expect(err).To(gomega.BeNil(), "Failed to generate calico-crs-configmap manifest yaml")

		err = runCommandPassThrough("kubectl", "apply", "-f", configmap)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to create ClusterResourceSet Configmap resource for  cluster: %s", clusterName))
	})
	return configmap
}

func createClusterBootstrapConfig(clusterName string, nameSpace string, bootstrapLabel string, patSecret string) (bootstrapConfig string) {
	tmplConfig := path.Join(getCheckoutRepoPath(), "test", "utils", "data", "gitops-cluster-bootstrap-config.yaml")
	bootstrapConfig = path.Join("/tmp", nameSpace+"-gitops-cluster-bootstrap-config.yaml")

	ginkgo.By(fmt.Sprintf("Add ClusterBootstrapConfig resource for %s cluster to management cluster", clusterName), func() {
		cmd := fmt.Sprintf(`cat %s | \
			sed s,{{NAME}},%s,g | \
			sed s,{{NAMESPACE}},%s,g | \
			sed s,{{BOOTSTRAP}},%s,g | \
			sed s,{{PAT_SECRET}},%s,g | \
			sed s,{{GIT_PROVIDER}},%s,g | \
			sed s,{{GITOPS_REPO_NAME}},%s,g | \
			sed s,{{GITOPS_REPO_OWNER}},%s,g | \
			sed s,{{GIT_PROVIDER_HOSTNAME}},%s,g`, tmplConfig, clusterName, nameSpace, bootstrapLabel, patSecret, gitProviderEnv.Type, gitProviderEnv.Repo, gitProviderEnv.Org, gitProviderEnv.Hostname)
		err := runCommandPassThrough("sh", "-c", fmt.Sprintf("%s > %s", cmd, bootstrapConfig))
		gomega.Expect(err).To(gomega.BeNil(), "Failed to generate ClusterBootstrapConfig manifest yaml")

		err = runCommandPassThrough("kubectl", "apply", "-f", bootstrapConfig)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to create ClusterBootstrapConfig resource for  cluster: %s", clusterName))
	})

	return bootstrapConfig
}

func connectGitopsCuster(clusterName string, nameSpace string, bootstrapLabel string, kubeconfigSecret string) (gitopsCluster string) {
	ginkgo.By(fmt.Sprintf("Add GitopsCluster resource for %s cluster to management cluster", clusterName), func() {
		contents, err := ioutil.ReadFile(path.Join(getCheckoutRepoPath(), "test/utils/data/gitops-cluster.yaml"))
		gomega.Expect(err).To(gomega.BeNil(), "Failed to read GitopsCluster template yaml")

		t := template.Must(template.New("gitops-cluster").Parse(string(contents)))

		// Prepare  data to insert into the template.
		type TemplateInput struct {
			ClusterName      string
			NameSpace        string
			Bootstrap        string
			KubeconfigSecret string
		}
		input := TemplateInput{clusterName, nameSpace, bootstrapLabel, kubeconfigSecret}

		gitopsCluster = path.Join("/tmp", clusterName+"-gitops-cluster.yaml")

		f, err := os.Create(gitopsCluster)
		gomega.Expect(err).To(gomega.BeNil(), "Failed to create GitopsCluster manifest yaml")

		err = t.Execute(f, input)
		f.Close()
		gomega.Expect(err).To(gomega.BeNil(), "Failed to generate GitopsCluster manifest yaml")

		err = runCommandPassThrough("kubectl", "apply", "-f", gitopsCluster)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to create GitopsCluster resource for  cluster: %s", clusterName))
	})
	return gitopsCluster
}

func addSource(sourceType, sourceName, namespace, url, kubeconfig string) {
	ginkgo.By(fmt.Sprintf("Adding %s %s Source", sourceType, sourceName), func() {
		if kubeconfig != "" {
			kubeconfig = "--kubeconfig=" + kubeconfig
		}

		var err error
		switch sourceType {
		case "git":
			err = runCommandPassThrough("sh", "-c", fmt.Sprintf("flux create source git %s --url=%s --branch=master --interval=30s --namespace %s %s", sourceName, url, namespace, kubeconfig))
		case "helm":
			err = runCommandPassThrough("sh", "-c", fmt.Sprintf("flux create source helm %s --url=%s --interval=30s --namespace %s %s", sourceName, url, namespace, kubeconfig))
		}

		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to create %srepository source: %s", sourceType, sourceName))
	})
}

func deleteSource(sourceType, sourceName, namespace, kubeconfig string) {
	ginkgo.By(fmt.Sprintf("Delete %s %s Source", sourceType, sourceName), func() {
		_ = runCommandPassThrough("sh", "-c", fmt.Sprintf("flux delete source %s %s --silent --namespace %s %s", sourceType, sourceName, namespace, kubeconfig))
	})
}

func addKustomizationBases(clusterType, clusterName, clusterNamespace string) {
	ginkgo.By("And add kustomization bases for common resources for leaf cluster)", func() {
		repoAbsolutePath := path.Join(configRepoAbsolutePath(gitProviderEnv))
		checkoutTestDataPath := path.Join(getCheckoutRepoPath(), "test", "utils", "data")
		leafClusterPath := path.Join(repoAbsolutePath, "clusters", clusterNamespace, clusterName)
		clusterBasesPath := path.Join(repoAbsolutePath, "clusters", "bases")

		pathErr := func() error {
			pullGitRepo(repoAbsolutePath)
			_, err := os.Stat(path.Join(leafClusterPath, "flux-system", "kustomization.yaml"))
			return err

		}
		gomega.Eventually(pathErr, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Leaf cluster %s repository path doesn't exists", clusterName))

		if clusterType != "capi" {
			gomega.Expect(copyFile(path.Join(checkoutTestDataPath, "clusters-bases-kustomization.yaml"), leafClusterPath)).Should(gomega.Succeed(), fmt.Sprintf("Failed to copy clusters-bases-kustomization.yaml to %s", leafClusterPath))
		}

		gomega.Expect(createDirectory(clusterBasesPath)).Should(gomega.Succeed(), fmt.Sprintf("Failed to create %s directory", clusterBasesPath))
		gomega.Expect(copyFile(path.Join(checkoutTestDataPath, "user-roles.yaml"), clusterBasesPath)).Should(gomega.Succeed(), fmt.Sprintf("Failed to copy user-roles.yaml to %s", clusterBasesPath))
		gomega.Expect(copyFile(path.Join(checkoutTestDataPath, "admin-role-bindings.yaml"), clusterBasesPath)).Should(gomega.Succeed(), fmt.Sprintf("Failed to copy admin-role-bindings.yaml to %s", clusterBasesPath))
		gomega.Expect(copyFile(path.Join(checkoutTestDataPath, "user-role-bindings.yaml"), clusterBasesPath)).Should(gomega.Succeed(), fmt.Sprintf("Failed to copy user-role-bindings.yaml to %s", clusterBasesPath))

		gitUpdateCommitPush(repoAbsolutePath, "Adding kustomization bases files")
	})
}

func createNamespace(namespaces []string) {
	for _, namespace := range namespaces {
		err := runCommandPassThrough("sh", "-c", fmt.Sprintf(`kubectl create namespace %s`, namespace))
		if err != nil {
			// 2nd attempt to create namespace
			_ = runCommandPassThrough("sh", "-c", fmt.Sprintf(`kubectl create namespace %s`, namespace))
		}
	}
}

func deleteNamespace(namespaces []string) {
	for _, namespace := range namespaces {
		err := runCommandPassThrough("sh", "-c", fmt.Sprintf(`kubectl delete namespace %s`, namespace))
		if err != nil {
			_ = runCommandPassThrough("sh", "-c", fmt.Sprintf(`kubectl delete namespace %s`, namespace))
		}
	}
}

func getApplicationCount() int {
	stdOut, _ := runCommandAndReturnStringOutput("kubectl get Kustomization -A --output name | wc -l")
	kCount, _ := strconv.Atoi(strings.TrimSpace(stdOut))

	stdOut, _ = runCommandAndReturnStringOutput("kubectl get HelmRelease -A --output name | wc -l")
	hCount, _ := strconv.Atoi(strings.TrimSpace(stdOut))

	return kCount + hCount
}

func getClustersCount() int {
	stdOut, _ := runCommandAndReturnStringOutput("kubectl get GitopsCluster --output name | wc -l")
	cCount, _ := strconv.Atoi(strings.TrimSpace(stdOut))
	return cCount + 1 // management cluster is a pseudo cluster
}

func getPoliciesCount() int {
	stdOut, _ := runCommandAndReturnStringOutput("kubectl get policies --output name | wc -l")
	pCount, _ := strconv.Atoi(strings.TrimSpace(stdOut))
	return pCount
}

func getViolationsCount() int {
	stdOut, _ := runCommandAndReturnStringOutput("kubectl  get events --field-selector reason=PolicyViolation  --output name | wc -l")
	vCount, _ := strconv.Atoi(strings.TrimSpace(stdOut))
	return vCount
}
