package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/wks/pkg/utilities/config"
	"github.com/weaveworks/wks/pkg/utilities/git"
)

type context struct {
	t          *testing.T
	configFile string
	conf       *config.WKPConfig
	wkBin      string
	testDir    string
	tmpDir     string
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

// getContext returns a "context" object containing all the information needed to perform most
// test tasks. Methods on the context object can be used to implement integration tests and manage
// temporary directories, git repositories, and clusters.
func getContext(testval *testing.T) *context {
	tmpDir, err := ioutil.TempDir("", "tmp_dir")
	require.NoError(testval, err)
	log.Infof("Using temporary directory: %s\n", tmpDir)
	file, conf := getConfigInfo(testval)

	return &context{
		t:          testval,
		configFile: file,
		conf:       conf,
		wkBin:      getWkBinary(testval),
		testDir:    getTestDir(testval),
		tmpDir:     tmpDir,
		env:        getEnvironment(testval, tmpDir),
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
	cmd := c.createSetupCommand()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
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

// runtimeConfigFilePath returns the path to the config.yaml file within the context's temporary directory
func (c *context) runtimeConfigFilePath() string {
	return filepath.Join(c.tmpDir, "setup", "config.yaml")
}

// updateConfigFileWithVersion updates the Kubernetes version for each wks machine in the config.yaml file
func (c *context) updateConfigFileWithVersion(version string) {
	path := c.runtimeConfigFilePath()
	info := getFileInfo(c.t, path)
	log.Infof("Configuring Kubernetes version: %s", version)
	err := git.UpdateYAMLFile(info, []string{"wksConfig", "kubernetesVersion"}, version)
	require.NoError(c.t, err)
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
	for retry := 0; retry < 45; retry++ {
		c.checkProgress(retryInfo)
		nodes := getAllReadyNodes(c.t, c.env)
		log.Infof("Retry: %d, Count: %d", retry, len(nodes))
		if len(nodes) >= expectedNumberOfNodes {
			log.Infof("Reached expected node count")
			return
		}
		time.Sleep(60 * time.Second)
	}
	assert.FailNowf(c.t, "Never reached expected node count", "Expected: %d, got: %d", expectedNumberOfNodes,
		len(getAllNodeVersions(c.t, c.env)))
}

// checkClusterRunning checks if all expected pods are running
func (c *context) checkClusterRunning() {
	checkForExpectedPods(c.t, c.tmpDir, c.getAllPods())
}

// PodData stores a component id (namespace + name)
type PodData struct {
	Namespace, Name string
}

var lineNameExp = regexp.MustCompile(`\s*['"]?name['"]?:\s*['"]?([^\s'"]+)['"]?`)

// getAllPods returns PodData for each pod running the cluster
func (c *context) getAllPods() []PodData {
	cmdItems := []string{"kubectl", "get", "pods", "--all-namespaces", "-o", "json"}
	cmd := exec.Command(cmdItems[0], cmdItems[1:]...)
	cmd.Env = c.env
	podJson, err := cmd.CombinedOutput()
	assert.NoError(c.t, err)
	pods := []PodData{}
	podList := map[string]interface{}{}
	if err := json.Unmarshal(podJson, &podList); err != nil {
		assert.FailNowf(c.t, "Invalid pod data", "returned pod data: %s", podJson)
	}
	podEntries := podList["items"].([]interface{})
	for _, entry := range podEntries {
		meta := entry.(map[string]interface{})["metadata"].(map[string]interface{})
		pods = append(pods, PodData{Namespace: meta["namespace"].(string), Name: meta["name"].(string)})
	}
	return pods
}

// shouldSkipComponents returns true if the SKIP_COMPONENTS environment variable is set
func (c *context) shouldSkipComponents() bool {
	skipComponents, found := os.LookupEnv("SKIP_COMPONENTS")
	return found && skipComponents == "true"
}

// checkForRequiredEnvVars ensures that any argument environment variable are set
func (c *context) checkForRequiredEnvVars(requiredVars ...string) {
	unset := []string{}
	for _, evar := range requiredVars {
		if os.Getenv(evar) == "" {
			unset = append(unset, evar)
		}
	}
	assert.Lenf(c.t, unset, 0, "Missing environment variable(s): %+v", unset)
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
	if err := c.showItems("nodes"); err != nil {
		return err
	}
	if err := c.showItems("pods"); err != nil {
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

func checkForExpectedPods(t *testing.T, dir string, pods []PodData) {
	componentsFilePath := filepath.Join(dir, "cluster", "platform", "components.js")
	expectedPodIDs := extractComponentIDs(t, componentsFilePath)
	for retries := 0; retries < 10; retries++ {
		if checkPods(expectedPodIDs, pods) {
			return
		}
		time.Sleep(60 * time.Second)
	}
	assert.FailNowf(t, "Expected pods not found", "found: %v, expected: %v", pods, expectedPodIDs)
}

func checkPods(expected, pods []PodData) bool {
	for _, podData := range expected {
		found := false
		for _, pod := range pods {
			if pod.Namespace == podData.Namespace && pod.Name == podData.Name {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Extract namespace and name from components; assumes we will continue to follow the convention that
// the namespace name is: wkp-<component name>
func extractComponentIDs(t *testing.T, componentsPath string) []PodData {
	filedata, err := ioutil.ReadFile(componentsPath)
	assert.NoError(t, err)
	lines := strings.Split(string(filedata), "\n")
	result := []PodData{}
	for _, line := range lines {
		if namespaceMatch := lineNameExp.FindStringSubmatch(line); namespaceMatch != nil {
			namespace := namespaceMatch[1] // "0" is whole match
			name := extractComponentName(namespace)
			if !strings.HasPrefix(namespace, "wkp-") || name == "eks-controller" || name == "flux-bootstrap" {
				// only for management cluster and startup
				continue
			}
			result = append(result, PodData{Namespace: namespace, Name: name})
		}
	}
	return result
}

func extractComponentName(namespace string) string {
	return strings.TrimPrefix(namespace, "wkp-")
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

func getEnvironment(t *testing.T, dir string) []string {
	_, conf := getConfigInfo(t)
	kubeconfigPath := filepath.Join(dir, "setup", "weavek8sops", conf.ClusterName, "kubeconfig")
	return append(getEnvironmentWithoutKubeconfig(t, dir), fmt.Sprintf("KUBECONFIG=%s", kubeconfigPath))
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
