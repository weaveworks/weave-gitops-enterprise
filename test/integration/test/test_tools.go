package test

import (
	"bufio"
	"bytes"
	contxt "context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fluxcd/go-git-providers/gitprovider"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/wks/pkg/cmdutil"
	"github.com/weaveworks/wks/pkg/utilities/config"
	"github.com/weaveworks/wks/pkg/utilities/git"

	"github.com/weaveworks/wks/pkg/github/ggp"
)

type context struct {
	t          *testing.T
	configFile string
	conf       *config.WKPConfig
	wkBin      string
	testDir    string
	tmpDir     string
	randGen    *rand.Rand
	env        []string
	repoExists bool
}

type testFlag int

const (
	noError testFlag = iota
)

type connectRetryInfo struct {
	current, max int
}

func flagSet(flags []testFlag, flag testFlag) bool {
	for _, f := range flags {
		if f == flag {
			return true
		}
	}
	return false
}

type Component struct {
	Namespace string
	Name      string
}

// Sets the timeout for 'kubectl wait' for a deployment
var defaultDeploymentTimeout = "300s"

// Sets the retries for the pods of a deployment to be scheduled
var defaultDeploymentRetries = 50

// Sets the wait interval for checking if a deployment has been scheduled (in seconds)
var defaultDeploymentRetryInterval = 8

// All the components of a WKP cluster, ordered by deployment time
var allComponents = []Component{
	{"weavek8sops", "wks-controller"},
	{"wkp-flux", "flux"},
	{"wkp-flux", "flux-helm-operator"},
	{"wkp-flux", "memcached"},
	{"wkp-tiller", "tiller-deploy"},
	{"wkp-gitops-repo-broker", "gitops-repo-broker"},
	// {"wkp-github-service", "github-service"}, deployed only when team-workspaces are enabled
	{"wkp-scope", "weave-scope-cluster-agent-weave-scope"},
	{"wkp-scope", "weave-scope-frontend-weave-scope"},
	{"wkp-grafana", "grafana"},
	{"wkp-grafana", "grafana"},
	{"wkp-prometheus", "prometheus-operator-kube-state-metrics"},
	{"wkp-prometheus", "prometheus-operator-operator"},
	// {"wkp-workspaces", "repository-controller"}, deployed only when team-workspaces are enabled
	{"wkp-external-dns", "external-dns"},
	{"wkp-ui", "wkp-ui-server"},
	{"wkp-ui", "wkp-ui-nginx-ingress-controller"},
	{"wkp-ui", "wkp-ui-nginx-ingress-controller-default-backend"},
}

func (c *context) checkComponentRunning(component Component) bool {
	found := false
	for retry := 0; retry < defaultDeploymentRetries; retry++ {
		if c.checkDeploymentExists(component) {
			found = true
			break
		}
		time.Sleep(time.Duration(defaultDeploymentRetryInterval) * time.Second)
	}
	if !found || !c.checkDeploymentRunning(component) {
		log.Errorf("Component is not running: %s Was found: %v\n", component.Name, found)
		return false
	}
	return true
}

func (c *context) checkDeploymentExists(component Component) bool {
	deploymentName := "deployment/" + component.Name
	cmdItems := []string{"kubectl", "wait", "--for=condition=available", "--timeout", defaultDeploymentTimeout,
		"--namespace", component.Namespace, deploymentName}
	cmd := exec.Command(cmdItems[0], cmdItems[1:]...)
	cmd.Env = c.env
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Infof("Deployment: %s timed out\nOutput: %s", component.Name, string(output))

		// Printing all pods for debugging
		cmdItems := []string{"kubectl", "get", "pods", "--all-namespaces"}
		cmd := exec.Command(cmdItems[0], cmdItems[1:]...)
		cmd.Env = c.env
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Infof("Failed to get pods: %s", string(output))
			return false
		}
		log.Infof("Current state:\n%s", string(output))
		return false
	}
	return true
}

func (c *context) checkDeploymentRunning(component Component) bool {
	cmdItems := []string{"kubectl", "get", "deployment", "-n", component.Namespace, component.Name}
	cmd := exec.Command(cmdItems[0], cmdItems[1:]...)
	cmd.Env = c.env
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Infof("Failed to get deployment: %s", string(output))
		return false
	}
	return true
}

// getContext returns a "context" object containing all the information needed to perform most
// test tasks. Methods on the context object can be used to implement integration tests and manage
// temporary directories, git repositories, and clusters.
func getContext(t *testing.T) *context {
	tmpDir, err := ioutil.TempDir("", "tmp_dir")
	require.NoError(t, err)
	return getContextFrom(t, tmpDir)
}

func getContextFrom(t *testing.T, tmpDir string) *context {
	log.Infof("Using temporary directory: %s\n", tmpDir)
	file, conf := getConfigInfo(t)
	s := rand.NewSource(time.Now().Unix())

	return &context{
		t:          t,
		configFile: file,
		conf:       conf,
		wkBin:      getWkBinary(t),
		testDir:    getTestDir(t),
		tmpDir:     tmpDir,
		randGen:    rand.New(s),
		env:        getEnvironmentWithoutKubeconfig(t, tmpDir),
		repoExists: false,
	}
}

// cleanup removes any temporary directories and git repos created for this context
func (c *context) cleanup() {
	log.Infof("Deleting temporary dir: %s", c.tmpDir)
	os.RemoveAll(c.tmpDir)
	if c.repoExists {
		log.Infof("Deleting git repo: %s/%s", c.conf.GitProviderOrg, c.conf.ClusterName)
		deleteRepo(c)
	}
}

// installWKPFiles runs "wk setup install" within the temporary directory associated with the context
func (c *context) installWKPFiles() {
	err := cmdutil.Run(c.createSetupCommand())
	require.NoError(c.t, err)
}

// copyConfigFileIntoPlace copies an external config file into the setup directory under the context's temporary directory
func (c *context) copyConfigFileIntoPlace() {
	c.copyFile(c.configFile, c.runtimeConfigFilePath())
	c.commitChanges()
}

// commitChanges commits any outstanding changes in the git repo under the context's temporary directory
func (c *context) commitChanges() {
	cmd := exec.Command("git", "add", "-u")
	cmd.Dir = c.tmpDir
	err := cmd.Run()
	assert.NoError(c.t, err)
	cmd = exec.Command("git", "commit", "-m", "checkpoint")
	cmd.Dir = c.tmpDir
	err = cmd.Run()
	assert.NoError(c.t, err)
}

func (c *context) commitChangesAndPush(message string) {
	cmd := exec.Command("git", "add", "-u")
	cmd.Dir = c.tmpDir
	err := cmd.Run()
	assert.NoError(c.t, err)

	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = c.tmpDir
	err = cmd.Run()
	assert.NoError(c.t, err)

	cmd = exec.Command("git", "push")
	cmd.Dir = c.tmpDir
	err = cmd.Run()
	assert.NoError(c.t, err)
}

// runtimeConfigFilePath returns the path to the config.yaml file within the context's temporary directory
func (c *context) runtimeConfigFilePath() string {
	return filepath.Join(c.tmpDir, "setup", "config.yaml")
}

// updateConfigFileWithVersion updates the Kubernetes version for each wks machine in the config.yaml file
func (c *context) updateConfigMachines(updateFn func([]config.MachineSpec) []config.MachineSpec) {
	c.conf.WKSConfig.SSHConfig.Machines = updateFn(c.conf.WKSConfig.SSHConfig.Machines)
	path := c.runtimeConfigFilePath()
	err := config.WriteConfig(path, c.conf)
	require.NoError(c.t, err)
}

// updateConfigFileWithVersion updates the Kubernetes version for each wks machine in the config.yaml file
func (c *context) updateConfigFileWithVersion(version string) {
	path := c.runtimeConfigFilePath()
	info := getFileInfo(c.t, path)
	log.Infof("Configuring Kubernetes version: %s", version)
	c.conf.WKSConfig.KubernetesVersion = version
	err := git.UpdateYAMLFile(info, []string{"wksConfig", "kubernetesVersion"}, version)
	require.NoError(c.t, err)
}

// chooseTestOS randomly selects either a CentOS or Ubuntu image for the cluster nodes used to run an upgrade test
func (c *context) chooseTestOS() {
	osImage := randomOSImageChoice(c)
	path := c.runtimeConfigFilePath()
	info := getFileInfo(c.t, path)
	log.Infof("Configuring test to use: %s", osImage)
	if err := git.UpdateYAMLFile(info, []string{"wksConfig", "footlooseConfig", "image"}, osImage); err != nil {
		err = git.UpdateYAMLFile(info, []string{"wksConfig", "footlooseConfig"}, []interface{}{"image", osImage})
		require.NoError(c.t, err)
	}
}

// setupCluster invokes "wk setup run" within the context's temporary directory
func (c *context) setupCluster() {
	deleteRepo(c)
	c.repoExists = true
	cmd := exec.Command(c.wkBin, "setup", "run")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = getEnvironmentWithoutKubeconfig(c.t, c.tmpDir)
	cmd.Dir = c.tmpDir
	err := cmd.Run()
	require.NoError(c.t, err)
}

// cleanupCluster invokes the cleanup.sh script within the context's temporary directory to delete the cluster
func (c *context) cleanupCluster(flags ...testFlag) {
	env := append(c.env, "SKIP_PROMPT=1", "DELETE_REPO_ON_CLEANUP=yes")
	cmd := exec.Command(filepath.Join(c.tmpDir, "setup", "cleanup.sh"), c.wkBin)
	cmd.Env = env
	err := cmd.Run()
	if flagSet(flags, noError) {
		return
	}
	assert.NoError(c.t, err)
}

// checkClusterAtCorrectVersion retrieves the version from each node in the cluster and ensures that they
// all match the argument version
func (c *context) checkClusterAtCorrectVersion(correctVersion string) {
	retryInfo := &connectRetryInfo{0, 8}
	for retry := 0; retry < 45; retry++ {
		c.checkProgress(retryInfo)
		okay := true
		versions := getAllNodeVersions(c.t, c.env)
		log.Infof("Retry: %d, Versions: %s", retry, strings.Join(versions, ", "))
		for _, version := range versions {
			if version != correctVersion {
				okay = false
				break
			}
		}
		if okay {
			return
		}
		time.Sleep(60 * time.Second)
	}
	assert.FailNowf(c.t, "Found incorrect version", "Expected: %s, got: %s", correctVersion,
		strings.Join(getAllNodeVersions(c.t, c.env), ", "))
}

// checkClusterAtExpectedNumberOfNodes waits for the cluster to reach the requested number of nodes
func (c *context) checkClusterAtExpectedNumberOfNodes(expectedNumberOfNodes int) {
	retryInfo := &connectRetryInfo{0, 6}
	for retry := 0; retry < 90; retry++ {
		c.checkProgress(retryInfo)
		nodes := getAllReadyNodes(c.t, c.env)
		log.Infof("Retry: %d, Expected: %d, Count: %d", retry, expectedNumberOfNodes, len(nodes))
		if len(nodes) == expectedNumberOfNodes {
			log.Infof("Reached expected node count")
			return
		}
		time.Sleep(30 * time.Second)
	}
	assert.FailNowf(c.t, "Never reached expected node count", "Expected: %d, got: %d", expectedNumberOfNodes,
		len(getAllNodeVersions(c.t, c.env)))
}

// isClusterRunning checks if all expected components are running
func (c *context) isClusterRunning() bool {
	for _, component := range allComponents {
		if !c.checkComponentRunning(component) {
			return false
		}
	}
	return true
}

// PodData stores a component id (namespace + name)
type PodData struct {
	Namespace, Name string
}

// shouldSkipComponents returns true if the SKIP_COMPONENTS environment variable is set
func (c *context) shouldSkipComponents() bool {
	skipComponents, found := os.LookupEnv("SKIP_COMPONENTS")
	return found && skipComponents == "true"
}

// assertSealedSecretsCanBeCreated ensures that sealed secret creation works correctly and the created secrets
// are valid
func (c *context) assertSealedSecretsCanBeCreated() {
	// create a sealed secret with a field password=supersekrekt
	// kubeseal uses the certificate created when the cluster was spun up
	kubesealPath := filepath.Join(c.tmpDir, "bin", "kubeseal")
	cmdCreateSecret := fmt.Sprintf("kubectl create secret generic --dry-run --output json mysecret --from-literal=password=supersekret | %s --cert=%s/setup/sealed-secrets-cert.crt > mysealedsecret.json", kubesealPath, c.tmpDir)
	cmd := exec.Command("bash", "-c", cmdCreateSecret)
	cmd.Env = c.env
	err := cmd.Run()
	assert.NoError(c.t, err)

	// wait until the controller deployment is ready
	cmdItems := []string{"kubectl", "wait", "--for", "condition=available", "--timeout=120s", "deployment/sealed-secrets-controller", "--namespace", "kube-system"}
	cmd = exec.Command(cmdItems[0], cmdItems[1:]...)
	cmd.Env = c.env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	assert.NoError(c.t, err)

	// create a SealedSecret
	cmdItems = []string{"kubectl", "create", "-f", "mysealedsecret.json"}
	cmd = exec.Command(cmdItems[0], cmdItems[1:]...)
	cmd.Env = c.env
	err = cmd.Run()
	assert.NoError(c.t, err)

	defer os.Remove("mysealedsecret.json")

	// check that the Secret created from the SealedSecret has the right field
	cmdGetPasswordField := "kubectl get secret mysecret -o=jsonpath='{.data.password}' | base64 --decode"

	var secretData []byte
	for retry := 0; retry < 5; retry++ {
		time.Sleep(5 * time.Second)
		cmd = exec.Command("bash", "-c", cmdGetPasswordField)
		cmd.Env = c.env
		secretData, err = cmd.CombinedOutput()
		if string(secretData) == "supersekret" {
			break
		}
	}

	assert.NoError(c.t, err)
	assert.Equal(c.t, "supersekret", string(secretData), "The decrypted password should match the original value")
	log.Info("Sealed secrets controller installed ok.")
}

// createSetupCommand creates a new invocation of "wk setup install"
func (c *context) createSetupCommand() *exec.Cmd {
	cmd := exec.Command(c.wkBin, "setup", "install")
	cmd.Dir = c.tmpDir
	cmd.Env = c.env
	return cmd
}

// copyFile copies the contents of a named file to destination file
func (c *context) copyFile(filename, dstpath string) {
	_, err := os.Stat(filename)
	assert.NoError(c.t, err)
	data, err := ioutil.ReadFile(filename)
	assert.NoError(c.t, err)
	err = ioutil.WriteFile(dstpath, data, 0666)
	assert.NoError(c.t, err)
}

// checkProgress attempts to contact the cluster and display statistics for nodes and pods
// If it fails to talk to the cluster for a specified number of times in a row, it will
// generate a panic to restart the test.
func (c *context) checkProgress(retryInfo *connectRetryInfo) {
	if c.showNodesAndPods() != nil {
		retryInfo.current++
	} else {
		retryInfo.current = 0
	}
	if retryInfo.current >= retryInfo.max {
		panic("Can't talk to cluster")
	}
}

// showNodesAndPods displays the current set of nodes and pods in tabular format
func (c *context) showNodesAndPods() error {
	if err := c.showItems("pods"); err != nil {
		return err
	}
	if err := c.showItems("nodes"); err != nil {
		return err
	}
	return nil
}

// showItems displays the current set of a specified object type in tabular format
func (c *context) showItems(itemType string) error {
	cmd := exec.Command("kubectl", "get", itemType, "--all-namespaces", "-o", "wide")
	cmd.Env = c.env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func randomOSImageChoice(c *context) string {
	return []string{"quay.io/footloose/centos7", "quay.io/footloose/ubuntu18.04"}[c.randGen.Intn(2)]
}

func getAllNodeVersions(t *testing.T, env []string) []string {
	cmdItems := []string{"kubectl", "get", "nodes", "--all-namespaces", "-o", "jsonpath={.items[*].status.nodeInfo.kubeletVersion}"}
	cmd := exec.Command(cmdItems[0], cmdItems[1:]...)
	cmd.Env = env
	versions, err := cmd.CombinedOutput()
	if err != nil {
		log.Infof("Failed to retrieve versions: %s", versions)
		return []string{"<unknown>"}
	}
	vstrings := strings.Split(string(versions), " ")
	result := []string{}
	for _, vstr := range vstrings {
		result = append(result, vstr[1:])
	}
	return result
}

func getAllReadyNodes(t *testing.T, env []string) []string {
	cmdItems := []string{"kubectl", "get", "nodes", "--all-namespaces", "-o",
		`jsonpath={range .items[*]}{"\n"}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}`}
	cmd := exec.Command(cmdItems[0], cmdItems[1:]...)
	cmd.Env = env
	cmdResults, err := cmd.CombinedOutput()
	if err != nil {
		log.Infof("Failed to retrieve ready nodes: %s", cmdResults)
		return nil
	}
	nodeStrings := strings.Split(string(cmdResults), "\n")
	result := []string{}
	for _, nstr := range nodeStrings {
		if strings.Contains(nstr, "Ready=True") {
			result = append(result, nstr)
		}
	}
	return result
}

func getFileInfo(t *testing.T, path string) *git.ObjectInfo {
	info, err := git.GetFileObjectInfo(path)
	assert.NoError(t, err)
	return info
}

func deleteRepo(c *context) {
	org := c.conf.GitProviderOrg
	name := c.conf.ClusterName
	cmdstr := fmt.Sprintf("hub delete -y %s/%s", org, name)
	cmd := exec.Command("sh", "-c", cmdstr)
	cmd.Env = c.env
	cmd.Run() // there's nothing we can do with the error
}

func getVersions(envVar string) []string {
	return strings.Split(os.Getenv(envVar), ",")
}

func getTestDir(t *testing.T) string {
	testDir, err := os.Getwd()
	require.NoError(t, err)
	return testDir
}

func getWkBinary(t *testing.T) string {
	path, err := filepath.Abs(fmt.Sprintf("%s/../../../cmd/wk/wk", getTestDir(t)))
	require.NoError(t, err, "error getting wk binary path")
	return path
}

func getConfigInfo(t *testing.T) (string, *config.WKPConfig) {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "./config.yaml"
	}
	wkpConf, err := config.GenerateConfig(configFile)
	require.NoError(t, err)
	return configFile, wkpConf
}

func getEnvironmentWithoutKubeconfig(t *testing.T, dir string) []string {
	env := os.Environ()
	_, conf := getConfigInfo(t)
	kubeconfigPath := filepath.Join(dir, "setup", "weavek8sops", conf.ClusterName, "kubeconfig")
	return append(env, fmt.Sprintf("KUBECONFIG_PATH=%s", kubeconfigPath), getEntitlementEnvironmentEntry(getTestDir(t)))
}

func getEntitlementEnvironmentEntry(testDir string) string {
	entitlementEntry := "WKP_ENTITLEMENTS=%s/../../../entitlements/2018-08-31-weaveworks.entitlements"
	return fmt.Sprintf(entitlementEntry, testDir)
}

func (c *context) testCIDRBlocks(podCIDRBlock, serviceCIDRBlock string) {
	cmdItems := []string{"get", "pods", "-l", "name=wks-controller", "--namespace=weavek8sops",
		"-o", "jsonpath={.items[].status.podIP}"}
	podIP := c.getIP(cmdItems, "wks-controller")
	c.assertIPisWithinRange(string(podIP), podCIDRBlock, "Pod")

	cmdItems = []string{"get", "service", "kubernetes", "--namespace=default",
		"-o", "jsonpath={.spec.clusterIP}"}
	serviceIP := c.getIP(cmdItems, "Kubernetes service")
	c.assertIPisWithinRange(string(serviceIP), serviceCIDRBlock, "Service")
}

func (c *context) getIP(cmdItems []string, msg string) string {
	cmd := exec.Command("kubectl", cmdItems...)
	ip, err := cmd.CombinedOutput()
	assert.NoError(c.t, err)
	log.Printf("%s has IP: %s\n", msg, string(ip))
	return string(ip)
}

func (c *context) assertIPisWithinRange(ip, ipRange, msg string) {
	_, subnet, err := net.ParseCIDR(ipRange)
	assert.NoError(c.t, err)
	parsedIP := net.ParseIP(ip)
	isValid := subnet.Contains(parsedIP)
	log.Printf("%s IP %s is inside %s range? %v\n", msg, parsedIP, ipRange, isValid)
	assert.True(c.t, isValid)

}

// setupPrePushHook installs a pre-push hook that counts push's number happening after wk setup run
func (c *context) setupPrePushHook() {
	prePushPath := filepath.Join(c.tmpDir, "/.git/hooks/pre-push")
	pushCountPath := filepath.Join(c.tmpDir, "/.git/hooks/push-count.txt")
	prePush, _ := os.Create(prePushPath)
	pushCount, _ := os.Create(pushCountPath)

	_, err := prePush.WriteString("#!/bin/sh\n\necho \"pushed\" >> " + pushCountPath)
	assert.NoError(c.t, err)

	prePush.Close()
	pushCount.Close()

	err = os.Chmod(prePushPath, 0755)
	require.NoError(c.t, err)
	log.Info("pre-push hook installed.")
}

// checkPushCount counts the number of push's added to push-count.txt
func (c *context) checkPushCount() {
	pushCountPath := filepath.Join(c.tmpDir, "/.git/hooks/push-count.txt")
	pushCount, _ := os.Open(pushCountPath)
	fileScanner := bufio.NewScanner(pushCount)
	lineCount := 0
	for fileScanner.Scan() {
		lineCount++
	}

	expectedPushes := 1
	if os.Getenv("SKIP_COMPONENTS") != "true" {
		expectedPushes = 2
	}
	assert.Equal(c.t, expectedPushes, lineCount, "Number of pushes matches expected.")
	log.Info("Testing number of pushes passed.")
}

// checkLogsContainString gets the logs of a pod in a namespace and checks if a message is contained in them
func (c *context) checkLogsContainString(namespace, podName, message string) bool {
	cmd := exec.Command("kubectl", "logs", "-n", namespace, fmt.Sprintf("-l=name=%s", podName))
	logs, err := cmd.CombinedOutput()
	assert.NoError(c.t, err)
	log.Printf("logs of pod %s: \n%s\n", podName, string(logs))
	return strings.Contains(string(logs), message)
}

// checkLogsContainString gets the logs of a pod in a namespace and checks if a message is contained in them
func (c *context) checkDeploymentLogsContainString(namespace, deploymentName, message string) bool {
	cmd := exec.Command("kubectl", "logs", "-n", namespace, fmt.Sprintf("deployment/%s", deploymentName))
	logs, err := cmd.CombinedOutput()
	assert.NoError(c.t, err)
	return strings.Contains(string(logs), message)
}

func runCommand(c *context, firstItem string, cmdItems ...string) ([]byte, []byte, error) {
	cmd := exec.Command(firstItem, cmdItems...)
	var stdout, stderr bytes.Buffer
	cmd.Dir = c.tmpDir
	cmd.Env = c.env
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

// findTeamAccess returns team access permission to a specific repo
func (c *context) findTeamAccess(orgName, repoName, teamName string) (gitprovider.TeamAccess, error) {
	ggpClient, err := ggp.CreateGithubClient()
	assert.NoError(c.t, err)
	ctx := contxt.Background()

	orgRepoRef := ggp.NewOrgRepoRef(orgName, repoName)

	repo, err := ggpClient.OrgRepositories().Get(ctx, orgRepoRef)
	assert.NoError(c.t, err)

	teamAccess, err := repo.TeamAccess().Get(ctx, teamName)

	return teamAccess, err
}

func (c *context) getWorkspaceYaml(repoName, orgName, teams string) string {
	workspaceYamlTemplate := `
apiVersion: wkp.weave.works/v1alpha1
kind: Workspace
metadata:
  name: demo
  namespace: wkp-workspaces
spec:
  interval: 1m
  suspend: false
  gitProvider:
    type: github
    hostname: github.com
    tokenRef:
      name: github-token
  gitRepository:
    name: %s
    owner: %s
    teams: %s
  clusterScope:
    role: namespace-admin
    namespaces:
    - name: demo-app
    - name: demo-db
    - name: demo-resource-quota
      resourceQuota:
        hard:
          limits.cpu: "2"
          limits.memory: 2Gi
          requests.cpu: "1"
          requests.memory: 1Gi
    - name: demo-limit-range
      limitRange:
        limits:
          - max:
              memory: 1Gi
            min:
              memory: 500Mi
            type: Container
    networkPolicy: workspace-isolation
`

	return fmt.Sprintf(workspaceYamlTemplate, repoName, orgName, teams)
}

func (c *context) pushFileToGit(commitMessage, path string) {
	commitMessage = fmt.Sprintf("-m%s", commitMessage)
	cmd := exec.Command("git", "add", path)
	cmd.Dir = c.tmpDir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	assert.NoError(c.t, err)

	cmd = exec.Command("git", "commit", commitMessage)
	cmd.Dir = c.tmpDir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	assert.NoError(c.t, err)

	cmd = exec.Command("git", "push", "origin", "master")
	cmd.Dir = c.tmpDir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	assert.NoError(c.t, err)
}
