package test

import (
	"bufio"
	"bytes"
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

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/cmdutil"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/utilities/config"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/utilities/git"
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

type ComponentType string

const (
	Deployment  ComponentType = "deployment"
	StatefulSet ComponentType = "statefulset"
)

type Component struct {
	Namespace string
	Name      string
	Type      ComponentType
}

// Sets the timeout for 'kubectl wait' for a deployment
var defaultTimeout = "2000s"

// Sets the retries for the pods of a deployment/statefulset to be scheduled
var defaultRetries = 50

// Sets the wait interval for checking if a deployment/statefulset has been scheduled (in seconds)
var defaultRetryInterval = 8

// All the components of a WKP cluster, ordered by deployment time
var allComponents = func(track string) []Component {
	var components = []Component{
		{"wkp-flux", "flux", Deployment},
		{"wkp-flux", "flux-helm-operator", Deployment},
		{"wkp-flux", "memcached", Deployment},
		{"wkp-tiller", "tiller-deploy", Deployment},
		{"wkp-gitops-repo-broker", "gitops-repo-broker", Deployment},
		{"wkp-scope", "weave-scope-cluster-agent-weave-scope", Deployment},
		{"wkp-scope", "weave-scope-frontend-weave-scope", Deployment},
		{"wkp-grafana", "grafana", Deployment},
		{"wkp-prometheus", "prometheus-operator-kube-state-metrics", Deployment},
		{"wkp-prometheus", "prometheus-operator-kube-p-operator", Deployment},
		{"wkp-external-dns", "external-dns", Deployment},
		{"wkp-ui", "wkp-ui-server", Deployment},
		{"wkp-ui", "wkp-ui-nginx-ingress-controller", Deployment},
		{"wkp-ui", "wkp-ui-nginx-ingress-controller-default-backend", Deployment},
		{"wkp-gitops-repo-broker", "nats", StatefulSet},
	}

	if track == "wks-ssh" || track == "wks-footloose" {
		components = append(components, Component{"weavek8sops", "wks-controller", Deployment})
	}

	return components
}

var tmpDirPath = "/tmp/cluster_dir"

func (c *context) checkComponentRunning(component Component) bool {
	found := false
	for retry := 0; retry < defaultRetries; retry++ {
		if c.checkResourceExists(component) {
			found = true
			break
		}
		time.Sleep(time.Duration(defaultRetryInterval) * time.Second)
	}
	if !found || !c.checkResourceRunning(component) {
		log.Errorf("Component is not running: %s Was found: %v\n", component.Name, found)
		return false
	}
	return true
}

func (c *context) checkResourceExists(component Component) bool {
	var cmdItems []string
	if component.Type == Deployment {
		name := "deployment/" + component.Name
		cmdItems = []string{"kubectl", "wait", "--for=condition=available", "--timeout", defaultTimeout,
			"--namespace", component.Namespace, name}
	} else if component.Type == StatefulSet {
		// kubectl wait does not work for StatefulSet so we wait for the first pod instead
		// https://github.com/kubernetes/kubernetes/issues/79606
		selector := fmt.Sprintf("statefulset.kubernetes.io/pod-name=%s-0", component.Name)
		cmdItems = []string{"kubectl", "wait", "--for=condition=ready", "--timeout", defaultTimeout,
			"--namespace", component.Namespace, "--selector", selector, "pod"}
	}
	cmd := exec.Command(cmdItems[0], cmdItems[1:]...)
	cmd.Env = c.env
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Infof("Resource: %s timed out\nOutput: %s", component.Name, string(output))

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

func (c *context) checkResourceRunning(component Component) bool {
	cmdItems := []string{"kubectl", "get", string(component.Type), "-n", component.Namespace, component.Name}
	cmd := exec.Command(cmdItems[0], cmdItems[1:]...)
	cmd.Env = c.env
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Infof("Failed to get resource: %s", string(output))
		return false
	}
	return true
}

// getContext returns a "context" object containing all the information needed to perform most
// test tasks. Methods on the context object can be used to implement integration tests and manage
// temporary directories, git repositories, and clusters.
func getContext(t *testing.T) *context {
	err := os.Mkdir(tmpDirPath, 0755)
	require.NoError(t, err)
	return getContextFrom(t, tmpDirPath)
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
	c.runAndCheckError(c.wkBin, "setup", "install")
}

// copyConfigFileIntoPlace copies an external config file into the setup directory under the context's temporary directory
func (c *context) copyConfigFileIntoPlace(filename ...string) {
	path := c.configFile
	if len(filename) > 0 {
		path = filename[0]
	}
	c.copyFile(path, c.runtimeConfigFilePath())
	c.commitChanges()
}

// commitChanges commits any outstanding changes in the git repo under the context's temporary directory
func (c *context) commitChanges() {
	c.runAndCheckError("git", "add", "-u")
	c.runAndCheckError("git", "commit", "-m", "checkpoint")
}

// runtimeConfigFilePath returns the path to the config.yaml file within the context's temporary directory
func (c *context) runtimeConfigFilePath() string {
	return filepath.Join(c.tmpDir, "setup", "config.yaml")
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

// setupCluster invokes "wk setup run" within the context's temporary directory
func (c *context) setupCluster() {
	deleteRepo(c)
	c.repoExists = true
	assert.Empty(c.t, retryProcess(func() string {
		err := c.runCommandPassThrough(c.wkBin, "setup", "run")
		if err != nil {
			return "retry"
		}
		return ""
	}, 5, 60*time.Second))
}

// cleanupCluster invokes the cleanup.sh script within the context's temporary directory to delete the cluster
func (c *context) cleanupCluster(flags ...testFlag) {
	env := append(c.env, "SKIP_PROMPT=1", "DELETE_REPO_ON_CLEANUP=yes")
	cmd := exec.Command(filepath.Join(c.tmpDir, "setup", "cleanup.sh"), c.wkBin)
	cmd.Env = env
	err := cmdutil.Run(cmd)
	if flagSet(flags, noError) {
		return
	}
	assert.NoError(c.t, err)
}

// checkClusterAtExpectedNumberOfNodes waits for the cluster to reach the requested number of nodes
func (c *context) checkClusterAtExpectedNumberOfNodes(expectedNumberOfNodes int) {
	retryInfo := &connectRetryInfo{0, 6}
	for retry := 0; retry < 90; retry++ {
		c.checkProgress(retryInfo)
		nodes := getAllReadyNodes(c.t, c.env)
		log.Infof("Nodes => Retry: %d, Expected: %d, Count: %d", retry, expectedNumberOfNodes, len(nodes))
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
	for _, component := range allComponents(c.conf.Track) {
		if !c.checkComponentRunning(component) {
			_ = c.showItems("pods")
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
	c.runAndCheckError("bash", "-c", cmdCreateSecret)

	// wait until the controller deployment is ready
	c.runAndCheckError("kubectl", "wait", "--for", "condition=available", "--timeout=180s", "deployment/sealed-secrets-controller", "--namespace", "kube-system")

	// create a SealedSecret
	c.runAndCheckError("kubectl", "create", "-f", "mysealedsecret.json")

	defer os.Remove(filepath.Join(c.tmpDir, "mysealedsecret.json"))

	// check that the Secret created from the SealedSecret has the right field
	cmdGetPasswordField := "kubectl get secret mysecret -o=jsonpath='{.data.password}' | base64 --decode"

	var secretData []byte
	var err error
	for retry := 0; retry < 5; retry++ {
		time.Sleep(5 * time.Second)
		cmd := exec.Command("bash", "-c", cmdGetPasswordField)
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
	return c.runCommandPassThrough("kubectl", "get", itemType, "--all-namespaces", "-o", "wide")
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
		// wkp-simon-test-1-4:NetworkUnavailable=False;MemoryPressure=False;DiskPressure=False;PIDPressure=False;Ready=True;
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
	_ = c.runCommandPassThrough("sh", "-c", cmdstr) // there's nothing we can do with the error
}

func getVersions(envVar string) []string {
	return strings.Split(os.Getenv(envVar), ",")
}

func getTestDir(t *testing.T) string {
	testDir, err := os.Getwd()
	require.NoError(t, err)
	return testDir
}

func printWKBinaryInfo(t *testing.T, wkBin string) {
	cmd := exec.Command(wkBin, "version")
	version, err := cmd.CombinedOutput()
	assert.NoError(t, err)
	log.Printf("WK Version : %s", version)
	log.Printf("WK Path : %s", wkBin)
}

func getWkBinary(t *testing.T) string {
	var path string
	var err error
	path = os.Getenv("WK_BINARY_PATH")
	if path == "" {
		path, err = filepath.Abs(fmt.Sprintf("%s/../../../cmd/wk/wk", getTestDir(t)))
	}
	require.NoError(t, err, "error getting wk binary path")
	printWKBinaryInfo(t, path)
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

	if err := fileScanner.Err(); err != nil {
		log.Info("Error on pushCount", err)
	}

	expectedPushes := 1
	// ssh/footloose we have to delete the old flux which is another push
	if os.Getenv("SKIP_COMPONENTS") != "true" && (c.conf.Track == "wks-ssh" || c.conf.Track == "wks-footloose") {
		expectedPushes = 2
	}

	assert.Equal(c.t, expectedPushes, lineCount, "Number of pushes matches expected.")
	log.Info("Testing number of pushes passed.")
}

// Run a command in c.tmpDir, collect stdout and stderr for checking.
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

// Run a command in c.tmpDir and fail the test on error
func (c *context) runAndCheckError(name string, arg ...string) {
	c.t.Helper()
	cmd := exec.Command(name, arg...)
	cmd.Dir = c.tmpDir
	cmd.Env = c.env
	err := cmdutil.Run(cmd)
	require.NoError(c.t, err)
}

// Run the given command from current directory.
func runCommandFromCurrentDir(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Run a command in c.tmpDir, passing through stdout/stderr to parent's
func (c *context) runCommandPassThrough(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = c.tmpDir
	cmd.Env = c.env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func retryUntilCreated(c *context, apiResource string, namespace string, name string) {
	args := []string{"get", apiResource, name}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	for i := 0; i < 20; i++ {
		time.Sleep(5 * time.Second)
		cmd := exec.Command("kubectl", args...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err == nil {
			c.t.Logf("%s %s created", apiResource, name)
			return
		}
	}
	c.t.Fatalf("%s %s was not created", apiResource, name)
}

type callback func() string

func retryProcess(callback callback, retry int, retryWait time.Duration) string {
	response := ""
	for i := 0; i < retry; i++ {
		response = callback()
		if response == "" {
			return ""
		}
		fmt.Printf("response: %s\n", response)
		fmt.Printf("waiting %s\n", retryWait.String())
		time.Sleep(retryWait)
		fmt.Printf("retrying (%d out of %d) \n", i, retry)
	}
	if response != "" {
		return fmt.Sprintf("response: %s", response)
	}
	return ""
}
