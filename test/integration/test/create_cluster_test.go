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
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	nameExp        = regexp.MustCompile(`(?m)\s*['"]?name['"]?:\s*['"]?wk-cluster['"]?`)
	regionsExp     = regexp.MustCompile(`("regions":)(\s*"eu-north-1")`)
	clusterNameExp = regexp.MustCompile(`("name":)(\s*"wk-cluster")`)
)

func TestClusterCreation(t *testing.T) {
	versions := getVersions("CLUSTER_VERSIONS")
	if len(versions) == 1 && versions[0] == "" {
		versions = []string{"1.20.2"}
		fmt.Printf("Using default versions: '%s'\n", strings.Join(versions, ","))
	}

	region := os.Getenv("CLUSTER_REGION")
	if region == "" {
		region = "eu-north-1"
	}

	runClusterCreationTests(t, versions, region)
}

func runClusterCreationTests(t *testing.T, versions []string, region string) {
	// Clean up context in case of error
	var c *context

	version := versions[0]
	retryCount := 0

	defer func() {
		if os.Getenv("CLEANUP_REPO") != "false" {
			c.cleanupCluster(noError)
			log.Info("Deleting remote repo in defer...")
			c.cleanup()
			log.Info("done.")
		}
		if retry := recover(); retry != nil {
			log.Infof("failure occurred for %v, retry count %v", version, retryCount)
			retryCount++

			if retryCount <= 2 {
				runClusterCreationTest(c, t, version, region)
			} else {
				log.Infof("gave up on retry %v when trying for version %v", retryCount, version)
			}
		}
	}()

	for i, runVersion := range versions {
		log.Infof("All versions: %v\n", versions)
		c = getContext(t)
		runClusterCreationTest(c, t, runVersion, region)

		// clean up between runs but leave the final to the defer
		hasAnotherRun := i < len(versions)-1
		if hasAnotherRun {
			log.Infof("Another run to do (%v / %v), cleaning up", i+1, len(versions))
			c.cleanupCluster()
			c.cleanup()
		} else {
			log.Infof("Final run completed, leaving cleanup to defer()")
		}
	}
}

func runClusterCreationTest(c *context, t *testing.T, version string, region string) {
	log.Infof("\nTesting version: %v\n\n", version)

	// Create cluster
	log.Info("Creating and initializing cluster...")
	c.installWKPFiles()
	c.copyConfigFileIntoPlace()
	c.updateConfigFileWithVersion(version)

	if c.conf.Track == "eks" {
		updateClusterNameAndRegionForTest(c, region)
	}

	c.setupPrePushHook()

	c.setupCluster()

	c.checkPushCount()

	// Wait for all nodes to be up
	expectedNodes := 3
	if c.conf.Track == "wks-ssh" {
		expectedNodes = len(c.conf.WKSConfig.SSHConfig.Machines)
	}

	// We don't care about the cluster if you're running on this track
	if c.conf.Track != "wks-components" {
		c.checkClusterAtExpectedNumberOfNodes(expectedNodes)
	}

	// Label the worker node for mccp db.
	_ = runCommandFromCurrentDir("../../utils/scripts/mccp-setup-helpers.sh", "label")

	// Check that all components are functioning
	if !c.shouldSkipComponents() {
		// Check for all expected pods
		log.Info("Checking that expected pods are running...")
		if assert.True(t, c.isClusterRunning()) {
			log.Info("All pods are running.")
		}
	}

	//Add workspace provider token
	if os.Getenv("ADD_TEAM_WORKSPACE_TOKEN") == "true" {
		log.Info("Creating the secret with github token for the workspace")
		_, stderr, err := runCommand(c, c.wkBin, "workspaces", "add-provider",
			"--type", "github",
			"--token", os.Getenv("WORKSPACES_ORG_ADMIN_TOKEN"),
			"--secret-name", "github-token",
			"--git-commit-push")
		assert.NoError(c.t, err, string(stderr))

		retryUntilCreated(c, "secret", "wkp-workspaces", "github-token")

		assert.True(t, c.checkComponentRunning(
			Component{Name: "wkp-workspaces-controller", Namespace: "wkp-workspaces", Type: Deployment}))

		assert.True(t, c.checkComponentRunning(
			Component{Name: "kustomize-controller", Namespace: "wkp-workspaces", Type: Deployment}))

		assert.True(t, c.checkComponentRunning(
			Component{Name: "notification-controller", Namespace: "wkp-workspaces", Type: Deployment}))

		assert.True(t, c.checkComponentRunning(
			Component{Name: "source-controller", Namespace: "wkp-workspaces", Type: Deployment}))

	}

	if c.conf.Track == "wks-footloose" {
		checkKubeconfigWorksWithDefaultArgs(c)
	}

	// Check that sealed secrets work
	c.assertSealedSecretsCanBeCreated()

	if c.conf.Track == "wks-footloose" || c.conf.Track == "wks-ssh" {
		// Check that the pod and service CIDR blocks have been set
		c.testCIDRBlocks(c.conf.WKSConfig.PodCIDRBlocks[0], c.conf.WKSConfig.ServiceCIDRBlocks[0])
	}
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

func checkKubeconfigWorksWithDefaultArgs(c *context) {
	err := c.runCommandPassThrough(c.wkBin, "kubeconfig")
	assert.NoError(c.t, err)
}
