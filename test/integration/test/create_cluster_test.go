package test

// This test consists of two parts:
// - A test function that will install clusters of various Kubernetes versions
// - Default data for the test function that will install and upgrade a footloose/docker cluster
//
// The test may be run against different platforms by editing the local config.yaml file before running
//
// To select specific specific versions, set the environment variable CLUSTER_VERSIONS to a comma-separated
// list of versions, e.g. "1.14.1,1.15.7". A cluster will be created running each version in succession.
//
// To use your own configuration, set the CONFIG_FILE variable to the path to your config file.
//
// For a faster test, you can set the SKIP_COMPONENTS environment variable to "true". This will skip deploying the wks
//   components but run the upgrade with just the basic pods deployed.
//
// This test performs the following actions:
// - In a new temporary directory:
//   - Runs 'wk setup install' to populate the directory with all necessary scripts, manifests, etc.
//   - Copies local 'config.yaml' into 'setup/config.yaml' in the temporary directory
//   - Runs 'wk setup run' to create a git repo and a cluster
// - Check that the components described in components.js are running
// - Tear down cluster via cleanup script
// - Remove remote and local git repos
//
// It can be run via "go test" but requires a long timeout -- try "go test -run TestClusterCreation --timeout=99999s"

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/wks/pkg/utilities/git"
	"github.com/weaveworks/wksctl/pkg/plan/runners/ssh"
)

var (
	nameExp        = regexp.MustCompile(`(?m)\s*['"]?name['"]?:\s*['"]?wk-cluster['"]?`)
	regionsExp     = regexp.MustCompile(`("regions":)(\s*"eu-north-1")`)
	clusterNameExp = regexp.MustCompile(`("name":)(\s*"wk-cluster")`)
)

func TestClusterCreation(t *testing.T) {
	versions := getVersions("CLUSTER_VERSIONS")
	if len(versions) == 1 && versions[0] == "" {
		versions = []string{"1.16.3", "1.17.5"}
		fmt.Printf("Using default versions: '%s'\n", strings.Join(versions, ","))
	}

	region := os.Getenv("CLUSTER_REGION")
	if region == "" {
		region = "eu-north-1"
	}

	runClusterCreationTest(t, versions, region)
}

func runClusterCreationTest(t *testing.T, versions []string, region string) {
	// Clean up context in case of error
	var c *context
	defer func() {
		if c != nil {
			c.cleanupCluster(noError)
			c.cleanup()
		}
		if retry := recover(); retry != nil {
			runClusterCreationTest(t, versions, region)
		}
	}()

	for _, version := range versions {
		c = getContext(t)

		// Create cluster
		log.Info("Creating and initializing cluster...")
		c.installWKPFiles()
		c.copyConfigFileIntoPlace()
		c.updateConfigFileWithVersion(version)

		if c.conf.Track == "eks" {
			updateClusterNameAndRegionForTest(c, region)
		}

		c.setupCluster()

		// Check that all components are functioning
		if !c.shouldSkipComponents() {
			// Check for all expected pods
			log.Info("Checking that expected pods are running...")
			c.checkClusterRunning()
		}

		// Wait for all nodes to be up
		c.checkClusterAtExpectedNumberOfNodes(3)

		if c.conf.Track != "eks" {
			checkApiServerAndKubeletArguments(c)
		}

		// Check that sealed secrets work
		c.assertSealedSecretsCanBeCreated()

		// Clean up this cluster instance
		c.cleanupCluster()

		// Clean up this temporary directory
		c.cleanup()

		// Already cleaned up
		c = nil
	}
}

func checkApiServerAndKubeletArguments(c *context) {
	machinesPath := filepath.Join(c.tmpDir, "setup", "machines.yaml")
	machinesInfo, err := git.GetFileObjectInfo(machinesPath)
	assert.NoError(c.t, err)

	apiServerArgs := c.conf.WKSConfig.APIServerArguments
	kubeletArgs := c.conf.WKSConfig.KubeletArguments

	roleNodes := git.FindNestedFields(machinesInfo.ObjectNode, "items", "*", "metadata", "labels", "set")
	portNodes := git.FindNestedFields(machinesInfo.ObjectNode, "items", "*", "spec", "providerSpec", "value", "public", "port")

	for idx, portNode := range portNodes {
		sshClient := getTestSSHClient(c, portNode.Value)

		for _, kubeletArg := range kubeletArgs {
			argString := fmt.Sprintf("%s=%s", kubeletArg.Name, kubeletArg.Value)
			_, err := sshClient.RunCommand(fmt.Sprintf("ps -ef | grep -v 'ps -ef' | grep /usr/bin/kubelet | grep %s", argString), nil)
			assert.NoError(c.t, err)
		}

		if roleNodes[idx].Value == "master" {
			for _, apiServerArg := range apiServerArgs {
				argString := fmt.Sprintf("%s=%s", apiServerArg.Name, apiServerArg.Value)
				_, err := sshClient.RunCommand(fmt.Sprintf("ps -ef | grep -v 'ps -ef' | grep kube-apiserver | grep %s", argString), nil)
				assert.NoError(c.t, err)
			}
		}
	}
}

func getTestSSHClient(c *context, portStr string) *ssh.Client {
	portNum, err := strconv.Atoi(portStr)
	assert.NoError(c.t, err)

	sshClient, err := ssh.NewClient(ssh.ClientParams{
		User:           "root",
		Host:           "127.0.0.1",
		Port:           uint16(portNum),
		PrivateKeyPath: filepath.Join(c.tmpDir, "setup", "cluster-key"),
	})
	assert.NoError(c.t, err)
	return sshClient
}

// Update wk-cluster.yaml and components.js with cluster name and region
func updateClusterNameAndRegionForTest(c *context, region string) {
	clusterName := generateClusterName()
	log.Infof("Cluster: %s", clusterName)
	configPath := filepath.Join(c.tmpDir, "cluster", "platform", "clusters", "default", "wk-cluster.yaml")
	data, err := ioutil.ReadFile(configPath)
	assert.NoError(c.t, err)
	newdata := updateWithClusterName(data, clusterName)
	err = ioutil.WriteFile(configPath, newdata, 0644)
	assert.NoError(c.t, err)
	componentPath := filepath.Join(c.tmpDir, "cluster", "platform", "components.js")
	data, err = ioutil.ReadFile(componentPath)
	assert.NoError(c.t, err)
	newdata = updateComponentsWithClusterNameAndRegion(data, clusterName, region)
	err = ioutil.WriteFile(componentPath, newdata, 0644)
	assert.NoError(c.t, err)
}

func updateWithClusterName(data []byte, clusterName string) []byte {
	return nameExp.ReplaceAll(data, []byte("\n  name: \""+clusterName+"\""))
}

func updateComponentsWithClusterNameAndRegion(data []byte, clusterName, region string) []byte {
	return clusterNameExp.ReplaceAll(
		regionsExp.ReplaceAll(data, []byte(fmt.Sprintf(`$1"%s"`, region))),
		[]byte(fmt.Sprintf(`$1"%s"`, clusterName)))
}

func generateClusterName() string {
	return fmt.Sprintf("wk-%s-cluster", generateIDString())
}

func generateIDString() string {
	return fmt.Sprintf("%s-%d-%d", strings.ToLower(os.Getenv("USER")), os.Getppid(), os.Getpid())
}
