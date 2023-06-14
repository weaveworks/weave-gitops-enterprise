package acceptance

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
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

func controllerStatus(controllerName, namespace string) error {
	return runCommandPassThroughWithoutOutput("sh", "-c", fmt.Sprintf("kubectl rollout status deployment %s -n %s", controllerName, namespace))
}

func checkClusterService(endpointURL string) {
	adminPassword := GetEnv("CLUSTER_ADMIN_PASSWORD", "")
	gomega.Eventually(func(g gomega.Gomega) {
		logger.Info("Trying to login to cluster service")
		// login to obtain cookie
		stdOut, _ := runCommandAndReturnStringOutput(
			fmt.Sprintf(
				// insecure for self-signed tls
				`curl --insecure  -d '{"username":"%s","password":"%s"}' -H "Content-Type: application/json" -X POST %s/oauth2/sign_in -c -`,
				AdminUserName, adminPassword, endpointURL,
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
				cookie, endpointURL,
			),
			ASSERTION_1MINUTE_TIME_OUT,
		)
		g.Expect(stdOut).To(gomega.MatchRegexp("200"), "Cluster Service is not healthy: %v", stdErr)

	}, ASSERTION_2MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).Should(gomega.Succeed())
}

type Request struct {
	Path string
	Body []byte
}

// Wait until we get a good looking response from /v1/<resource>
// Ignore all errors (connection refused, 500s etc)
func waitForGitopsResources(ctx context.Context, request Request, timeout time.Duration, timeoutCtx ...time.Duration) error {
	contextTimeout := ASSERTION_5MINUTE_TIME_OUT
	if len(timeoutCtx) > 0 {
		contextTimeout = timeoutCtx[0]
	}
	adminPassword := GetEnv("CLUSTER_ADMIN_PASSWORD", "")
	waitCtx, cancel := context.WithTimeout(ctx, contextTimeout)
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
		resp, err := client.Post(testUiUrl+"/oauth2/sign_in", "application/json", bytes.NewReader([]byte(fmt.Sprintf(`{"username":"%s", "password":"%s"}`, AdminUserName, adminPassword))))
		if err != nil {
			logger.Tracef("error logging in (waiting for a success, retrying): %v", err)
			return false, nil
		}
		if resp.StatusCode != http.StatusOK {
			logger.Tracef("wrong status from login (waiting for a ok, retrying): %v", resp.StatusCode)
			return false, nil
		}
		// fetch gitops resource
		if request.Body != nil {
			resp, err = client.Post(testUiUrl+"/v1/"+request.Path, "application/json", bytes.NewReader(request.Body))
		} else {
			resp, err = client.Get(testUiUrl + "/v1/" + request.Path)
		}
		if err != nil {
			logger.Tracef("error getting %s in (waiting for a success, retrying): %v", request.Path, err)
			return false, nil
		}
		if resp.StatusCode != http.StatusOK {
			logger.Tracef("wrong status from %s (waiting for a ok, retrying): %v", request.Path, resp.StatusCode)
			return false, nil
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, nil
		}

		parseUrl, err := url.Parse(request.Path)
		if err != nil {
			logger.Errorf("failed to parse URL: %v", request.Path)
			return false, nil
		}

		return regexp.MatchString(strings.ToLower(fmt.Sprintf(`%s[\\"]+`, strings.Split(parseUrl.Path, "/")[0])), strings.ToLower(string(bodyBytes)))
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
			gomega.Expect(fmt.Errorf("error: Provided authentication type '%s' is not supported for CLI", userCredentials.UserType))
		}
	}

	cmd = fmt.Sprintf(`%s --endpoint %s %s %s %s`, gitopsBinPath, wgeEndpointUrl, insecureFlag, authFlag, cmd)
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

func reconcile(action, resource, resourceType, resourceName, namespace, kubeconfig string) {
	if kubeconfig != "" {
		kubeconfig = "--kubeconfig=" + kubeconfig
	}

	cmdSuspend := fmt.Sprintf("flux %s %s %s %s --namespace %s %s", action, resource, resourceType, resourceName, namespace, kubeconfig)
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
			configFile = "--config " + path.Join(testDataPath, configFile)
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
	contents, err := os.ReadFile(kubeconfigPath)
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
		switch app.Name {
		case "observability": // layer-0
			gomega.Expect(waitForResource("deploy", "observability-grafana", app.TargetNamespace, kubeconfigPath, ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())
			gomega.Expect(waitForResource("deploy", "observability-kube-state-metrics", app.TargetNamespace, kubeconfigPath, ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())
			waitForResourceState("Ready", "true", "pods", app.TargetNamespace, "release="+"observability", kubeconfigPath, ASSERTION_3MINUTE_TIME_OUT)
		case "postgres": // ks
			gomega.Expect(waitForResource("deploy", "postgres ", app.TargetNamespace, kubeconfigPath, ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())
			waitForResourceState("Ready", "true", "pods", app.TargetNamespace, "app=postgres", kubeconfigPath, ASSERTION_3MINUTE_TIME_OUT)
		case "podinfo": // ks
			gomega.Expect(waitForResource("deploy", "podinfo ", app.TargetNamespace, kubeconfigPath, ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())
			waitForResourceState("Ready", "true", "pods", app.TargetNamespace, "app=podinfo", kubeconfigPath, ASSERTION_3MINUTE_TIME_OUT)
		case "metallb": // layer-0
			gomega.Expect(waitForResource("deploy", app.TargetNamespace+"-metallb-controller ", app.TargetNamespace, kubeconfigPath, ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())
			waitForResourceState("Ready", "true", "pods", app.TargetNamespace, "app.kubernetes.io/name="+"metallb", kubeconfigPath, ASSERTION_3MINUTE_TIME_OUT)
		case "cert-manager": //l ayer-0
			gomega.Expect(waitForResource("deploy", app.TargetNamespace+"-cert-manager", app.TargetNamespace, kubeconfigPath, ASSERTION_2MINUTE_TIME_OUT)).To(gomega.Succeed())
			waitForResourceState("Ready", "true", "pods", app.TargetNamespace, "", kubeconfigPath, ASSERTION_3MINUTE_TIME_OUT)
		case "weave-policy-agent": // layer-1
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

// Copy the flux-system git repo and its secret from the flux-system namespace to a given namespace
func copyFluxSystemGitRepo(namespace string) {
	ginkgo.By("Copy flux-system git repo and its secret from the flux-system namespace a tenant namespace", func() {
		// Copy the flux-system git repo and its secret from the flux-system namespace to a given namespace
		// This is required for the cluster to be able to sync with the git repo
		err := runCommandPassThrough("sh", "-c", "kubectl get gitrepositories -n flux-system flux-system -o yaml | sed 's/  namespace: flux-system/  namespace: "+namespace+"/' | kubectl apply -f -")
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to get gitrepositories from flux-system namespace")

		err = runCommandPassThrough("sh", "-c", "kubectl get secret -n flux-system flux-system -o yaml | sed 's/  namespace: flux-system/  namespace: "+namespace+"/' | kubectl apply -f -")
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to get secret from flux-system namespace")

		err = runCommandPassThrough("sh", "-c", "kubectl annotate -n "+namespace+" gitrepo flux-system weave.works/repo-role=default")
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "Failed to annotate gitrepo with weave.works/repo-role=default")
	})
}

func createClusterResourceSet(clusterName string, nameSpace string) (resourceSet string) {
	ginkgo.By(fmt.Sprintf("Add ClusterResourceSet resource for %s cluster to management cluster", clusterName), func() {
		contents, err := os.ReadFile(path.Join(testDataPath, "bootstrap/calico-crs.yaml"))
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
		contents, err := os.ReadFile(path.Join(testDataPath, "bootstrap/calico-crs-configmap.yaml"))
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
	tmplConfig := path.Join(testDataPath, "bootstrap/gitops-cluster-bootstrap-config.yaml")
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

func connectGitopsCluster(clusterName string, nameSpace string, bootstrapLabel string, kubeconfigSecret string) (gitopsCluster string) {
	ginkgo.By(fmt.Sprintf("Add GitopsCluster resource for %s cluster to management cluster", clusterName), func() {
		contents, err := os.ReadFile(path.Join(testDataPath, "kustomization/gitops-cluster.yaml"))
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

func addSource(sourceType, sourceName, namespace, url, branchName, kubeconfig string) {
	ginkgo.By(fmt.Sprintf("Adding %s %s Source", sourceType, sourceName), func() {
		if kubeconfig != "" {
			kubeconfig = "--kubeconfig=" + kubeconfig
		}

		var err error
		switch sourceType {
		case "git":
			err = runCommandPassThrough("sh", "-c", fmt.Sprintf("flux create source git %s --url=%s --branch=%s --interval=30s --namespace %s %s", sourceName, url, branchName, namespace, kubeconfig))
		case "helm":
			err = runCommandPassThrough("sh", "-c", fmt.Sprintf("flux create source helm %s --url=%s --interval=30s --namespace %s %s", sourceName, url, namespace, kubeconfig))
		}

		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Failed to create %srepository source: %s", sourceType, sourceName))
	})
}

func addKustomizationBases(clusterType, clusterName, clusterNamespace string) {
	ginkgo.By("And add kustomization bases for common resources for leaf cluster", func() {
		repoAbsolutePath := path.Join(configRepoAbsolutePath(gitProviderEnv))
		leafClusterPath := path.Join(repoAbsolutePath, "clusters", clusterNamespace, clusterName)
		clusterBasesPath := path.Join(repoAbsolutePath, "clusters", "bases")

		pathErr := func() error {
			pullGitRepo(repoAbsolutePath)
			_, err := os.Stat(path.Join(leafClusterPath, "flux-system", "kustomization.yaml"))
			return err

		}
		gomega.Eventually(pathErr, ASSERTION_1MINUTE_TIME_OUT, POLL_INTERVAL_5SECONDS).ShouldNot(gomega.HaveOccurred(), fmt.Sprintf("Leaf cluster %s repository path doesn't exists", clusterName))

		if clusterType != "capi" {
			gomega.Expect(copyFile(path.Join(testDataPath, "kustomization/clusters-bases-kustomization.yaml"), leafClusterPath)).Should(gomega.Succeed(), fmt.Sprintf("Failed to copy clusters-bases-kustomization.yaml to %s", leafClusterPath))
		}

		gomega.Expect(createDirectory(clusterBasesPath)).Should(gomega.Succeed(), fmt.Sprintf("Failed to create %s directory", clusterBasesPath))
		gomega.Expect(copyFile(path.Join(testDataPath, "rbac/user-roles.yaml"), clusterBasesPath)).Should(gomega.Succeed(), fmt.Sprintf("Failed to copy user-roles.yaml to %s", clusterBasesPath))
		gomega.Expect(copyFile(path.Join(testDataPath, "rbac/admin-role-bindings.yaml"), clusterBasesPath)).Should(gomega.Succeed(), fmt.Sprintf("Failed to copy admin-role-bindings.yaml to %s", clusterBasesPath))
		gomega.Expect(copyFile(path.Join(testDataPath, "rbac/user-role-bindings.yaml"), clusterBasesPath)).Should(gomega.Succeed(), fmt.Sprintf("Failed to copy user-role-bindings.yaml to %s", clusterBasesPath))

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
