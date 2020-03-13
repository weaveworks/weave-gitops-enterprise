package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"text/template"

	yaml "gopkg.in/yaml.v3"
)

// Top-level config parameters
type WKPConfig struct {
	Track                string    `yaml:"track"`
	ClusterName          string    `yaml:"clusterName"`
	GitHubOrg            string    `yaml:"gitHubOrg"`
	DockerIOUser         string    `yaml:"dockerIOUser"`
	DockerIOPasswordFile string    `yaml:"dockerIOPasswordFile"`
	EKSConfig            EKSConfig `yaml:"eksConfig"`
	WKSConfig            WKSConfig `yaml:"wksConfig"`
}

// Parameters specific to eks
type EKSConfig struct {
	ClusterRegion        string            `yaml:"clusterRegion"`
	KubernetesVersion    string            `yaml:"kubernetesVersion"`
	NodeGroups           []NodeGroupConfig `yaml:"nodeGroups"`
	ManagedNodeGroupFile string            `yaml:"managedNodeGroupFile"`
}

type NodeGroupConfig struct {
	Name            string `yaml:"name"`
	InstanceType    string `yaml:"instanceType"`
	DesiredCapacity int64  `yaml:"desiredCapacity"`
}

// Parameters shared by 'footloose' and 'ssh'
type WKSConfig struct {
	KubernetesVersion string          `yaml:"kubernetesVersion"`
	ServiceCIDRBlocks []string        `yaml:"serviceCIDRBlocks"`
	PodCIDRBlocks     []string        `yaml:"podCIDRBlocks"`
	SSHConfig         SSHConfig       `yaml:"sshConfig"`
	FootlooseConfig   FootlooseConfig `yaml:"footlooseConfig"`
}

// Parameters specific to ssh
type SSHConfig struct {
	SSHUser    string        `yaml:"sshUser"`
	SSHKeyFile string        `yaml:"sshKeyFile"`
	Machines   []MachineSpec `yaml:"machines"`
}

type MachineSpec struct {
	Name           string `yaml:"name"`
	Role           string `yaml:"role"`
	PublicAddress  string `yaml:"publicAddress"`
	PublicPort     int64  `yaml:"publicPort"`
	PrivateAddress string `yaml:"privateAddress"`
	PrivatePort    int64  `yaml:"privatePort"`
}

// Parameters specific to footloose
type FootlooseConfig struct {
	Backend           string `yaml:"backend"`
	ControlPlaneNodes int64  `yaml:"controlPlaneNodes"`
	WorkerNodes       int64  `yaml:"workerNodes"`
}

// Templates for generating specific configs
const clusterFileTemplate = `apiVersion: cluster.k8s.io/v1alpha1
kind: Cluster
metadata:
  name: {{ .ClusterName }}
spec:
  clusterNetwork:
    services:
      cidrBlocks: {{ .ServiceCIDRBlocks }}
    pods:
      cidrBlocks: {{ .PodCIDRBlocks }}
    serviceDomain: cluster.local
  providerSpec:
    value:
      apiVersion: baremetalproviderspec/v1alpha1
      kind: BareMetalClusterProviderSpec
      user: {{ .SSHUser }}
      os:
        files:
        - source:
            configmap: repo
            key: kubernetes.repo
          destination: /etc/yum.repos.d/kubernetes.repo
        - source:
            configmap: repo
            key: docker-ce.repo
          destination: /etc/yum.repos.d/docker-ce.repo
        - source:
            configmap: docker
            key: daemon.json
          destination: /etc/docker/daemon.json
      cri:
        kind: docker
        package: docker-ce
        version: 18.09.7
`

const machineTemplate = `- apiVersion: cluster.k8s.io/v1alpha1
  kind: Machine
  metadata:
    labels:
      set: {{ .Role }}
    name: {{ .Name }}
    namespace: weavek8sops
  spec:
    versions:
      kubelet: {{ .KubernetesVersion }}
    providerSpec:
      value:
        apiVersion: baremetalproviderspec/v1alpha1
        kind: BareMetalMachineProviderSpec
        private:
          address: {{ .PrivateAddress }}
          port: {{ .PrivatePort }}
        public:
          address: {{ .PublicAddress }}
          port: {{ .PublicPort }}
`

const nodeGroupTemplate = `  - desiredCapacity: {{ .DesiredCapacity }}
    iam:
      withAddonPolicies:
        albIngress: true
    instanceType: {{ .InstanceType }}
    name: {{ .Name }}`

const eksTemplate = `apiVersion: infrastructure.eksctl.io/v1alpha5
kind: EKSCluster
metadata:
  name: {{ .ClusterName }} # set the AWS cluster name here
spec:
  region: {{ .ClusterRegion }} # set the AWS region here
  cloudWatch:
    clusterLogging:
      enableTypes:
      - audit
      - authenticator
      - controllerManager
  iam:
    serviceAccounts:
    - attachPolicyARNs:
      # FIXME: https://github.com/weaveworks/wk-quickstart-eks/issues/56 "Default cluster-config.js creates a ServiceAccount with AWS AdministratorAccess"
      - arn:aws:iam::aws:policy/AdministratorAccess
      metadata:
        name: ekscontroller
        namespace: wkp-eks-controller
    withOIDC: true
  nodeGroups:
{{ .NodeGroups }}
  managedNodeGroupFile: {{ .ManagedNodeGroupFile }}
  version: {{ .KubernetesVersion }}
`

const footlooseTemplate = `# This file contains high level configuration parameters. The setup.sh script
# takes this file as input and creates lower level manifests.
# backend defines how the machines underpinning Kubernetes nodes are created.
#  - docker: use containers as "VMs" using footloose:
#            https://github.com/weaveworks/footloose
#  - ignite: use footloose with ignite and firecracker to create real VMs using:
#            the ignite backend only works on linux as it requires KVM.
#            https://github.com/weaveworks/ignite.
backend: {{ .Backend }}
# Number of nodes allocated for the Kubernetes control plane and workers.
controlPlane:
  nodes: {{ .ControlPlaneNodes }}
workers:
  nodes: {{ .WorkerNodes }}
`

var (
	cidrRegexp       = regexp.MustCompile(`^([0-9]{1,3}\.){3}[0-9]{1,3}(\/([0-9]|[1-2][0-9]|3[0-2]))?$`)
	k8sVersionRegexp = regexp.MustCompile(`^([1][.](14|15)[.][0-9][0-9]?)$`)
)

func unmarshalConfig(configBytes []byte) (*WKPConfig, error) {
	var config WKPConfig
	err := yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Load a config from the file system into the structs from above
func readConfig(path string) (*WKPConfig, error) {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return unmarshalConfig(fileBytes)
}

func createClusterName() string {
	name := os.Getenv("USER")
	if name == "" {
		name = "cluster" // use "wk-cluster" if no user env var found
	}
	return "wk-" + name
}

// Check if file exists at specified path
func checkValidPath(field, path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("no file found at path: %q for field: %q", path, field)
	}
	return err
}

// The following functions come in pairs to check values and set defaults

// Global values
func checkRequiredGlobalValues(config *WKPConfig) error {
	if config.GitHubOrg == "" {
		return fmt.Errorf("gitHubOrg must be specified")
	}

	if config.DockerIOUser == "" {
		return fmt.Errorf("dockerIOUser must be specified")
	}

	if config.DockerIOPasswordFile == "" {
		return fmt.Errorf("dockerIOPasswordFile must be specified")
	}

	if err := checkValidPath("dockerIOPasswordFile", config.DockerIOPasswordFile); err != nil {
		return err
	}

	switch config.Track {
	case "":
		return fmt.Errorf("track must be specified")
	case "eks", "wks-ssh", "wks-footloose":
		return nil
	default:
		return fmt.Errorf("track must be one of: 'eks', 'wks-ssh', or 'wks-footloose'")
	}
}

func setDefaultGlobalValues(config *WKPConfig) {
	if config.ClusterName == "" {
		config.ClusterName = createClusterName()
	}
}

// eks values
func checkRequiredEKSValues(eksConfig *EKSConfig) error {
	if eksConfig.ClusterRegion == "" {
		return fmt.Errorf("clusterRegion must be specified")
	}

	for _, ng := range eksConfig.NodeGroups {
		if ng.DesiredCapacity < 0 {
			return fmt.Errorf("A node group must have a capacity of at least 1")
		}
	}

	if eksConfig.ManagedNodeGroupFile != "" {
		if err := checkValidPath("managedNodeGroupFile", eksConfig.ManagedNodeGroupFile); err != nil {
			return err
		}
	}

	switch eksConfig.KubernetesVersion {
	case "":
		return fmt.Errorf("A Kubernetes version must be specified")
	case "1.14", "1.15":
		return nil
	default:
		return fmt.Errorf(`Kubernetes version must be one of: "1.14" or "1.15"`)
	}
}

func setDefaultEKSValues(eksConfig *EKSConfig) {
	if len(eksConfig.NodeGroups) == 0 {
		eksConfig.NodeGroups = []NodeGroupConfig{{Name: "ng-0", InstanceType: "m5.large", DesiredCapacity: 3}}
		return
	}

	groups := eksConfig.NodeGroups
	for idx := range groups {
		group := &groups[idx]

		if group.Name == "" {
			group.Name = fmt.Sprintf("ng-%d", idx)
		}

		if group.InstanceType == "" {
			group.InstanceType = "m5.large"
		}

		if group.DesiredCapacity == 0 {
			group.DesiredCapacity = 3
		}
	}
}

// values shared between ssh and footloose
func checkRequiredWKSValues(wksConfig *WKSConfig) error {
	if wksConfig.KubernetesVersion == "" {
		return fmt.Errorf("A Kubernetes version must be specified")
	}

	if !k8sVersionRegexp.MatchString(wksConfig.KubernetesVersion) {
		return fmt.Errorf(
			"%s is not a valid Kubernetes version; must be 1.14.x-1.15.x",
			wksConfig.KubernetesVersion)
	}

	if len(wksConfig.ServiceCIDRBlocks) == 0 {
		return fmt.Errorf("At least one service CIDR block must be specified")
	}

	for _, cidr := range wksConfig.ServiceCIDRBlocks {
		if !cidrRegexp.MatchString(cidr) {
			return fmt.Errorf("%s is not a valid CIDR specification", cidr)
		}
	}

	if len(wksConfig.PodCIDRBlocks) == 0 {
		return fmt.Errorf("At least one pod CIDR block must be specified")
	}

	for _, cidr := range wksConfig.PodCIDRBlocks {
		if !cidrRegexp.MatchString(cidr) {
			return fmt.Errorf("%s is not a valid CIDR specification", cidr)
		}
	}

	return nil
}

// ssh values
func checkRequiredSSHValues(sshConfig *SSHConfig) error {
	if sshConfig.SSHKeyFile == "" {
		homedir := os.Getenv("HOME")
		if homedir == "" {
			return fmt.Errorf("No ssh key file specified and no home directory information available.")
		}
	} else if err := checkValidPath("sshKeyFile", sshConfig.SSHKeyFile); err != nil {
		return err
	}

	if len(sshConfig.Machines) == 0 {
		return fmt.Errorf("No machine information provided")
	}

	masters := 0
	workers := 0
	for idx := range sshConfig.Machines {
		machine := &sshConfig.Machines[idx]

		if machine.PublicAddress == "" {
			return fmt.Errorf("A public address must be specified for each machine")
		}

		switch machine.Role {
		case "":
			return fmt.Errorf("A role ('master' or 'worker') must be specified for each machine")
		case "master":
			masters++
		case "worker":
			workers++
		default:
			return fmt.Errorf("Invalid machine role: '%s'. Only 'master' and 'worker' are valid.",
				machine.Role)
		}
	}

	if masters == 0 || workers == 0 {
		return fmt.Errorf("Invalid machine set. At least one master and one worker must be specified.")
	}

	return nil
}

func setDefaultSSHValues(sshConfig *SSHConfig) {
	if sshConfig.SSHUser == "" {
		sshConfig.SSHUser = "root"
	}

	if sshConfig.SSHKeyFile == "" {
		sshConfig.SSHKeyFile = fmt.Sprintf("%s/.ssh/id_rsa", os.Getenv("HOME"))
	}

	for idx := range sshConfig.Machines {
		machine := &sshConfig.Machines[idx]
		if machine.Name == "" {
			machine.Name = fmt.Sprintf("%s-%s", machine.Role, machine.PublicAddress)
		}

		if machine.PublicPort == 0 {
			machine.PublicPort = 22
		}

		if machine.PrivateAddress == "" {
			machine.PrivateAddress = machine.PublicAddress
		}

		if machine.PrivatePort == 0 {
			machine.PrivatePort = machine.PublicPort
		}
	}
}

// footloose values (no defaults to set so only a check function)
func checkRequiredFootlooseValues(footlooseConfig *FootlooseConfig) error {
	if footlooseConfig.ControlPlaneNodes <= 0 {
		return fmt.Errorf("A footloose specification must have at least one control plane node")
	}

	if footlooseConfig.WorkerNodes <= 0 {
		return fmt.Errorf("A footloose specification must have at least one worker node")
	}

	switch footlooseConfig.Backend {
	case "":
		return fmt.Errorf("A footloose backend must be specified")
	case "docker", "ignite":
		return nil
	default:
		return fmt.Errorf("A footloose backend must be either 'docker' or 'ignite'")
	}
}

func checkRequiredValues(config *WKPConfig) error {
	if err := checkRequiredGlobalValues(config); err != nil {
		return err
	}

	if config.Track == "eks" {
		if err := checkRequiredEKSValues(&config.EKSConfig); err != nil {
			return err
		}
		return nil
	}

	if err := checkRequiredWKSValues(&config.WKSConfig); err != nil {
		return err
	}

	switch config.Track {
	case "wks-ssh":
		if err := checkRequiredSSHValues(&config.WKSConfig.SSHConfig); err != nil {
			return err
		}
	case "wks-footloose":
		if err := checkRequiredFootlooseValues(&config.WKSConfig.FootlooseConfig); err != nil {
			return err
		}
	}

	return nil
}

func addDefaultValues(config *WKPConfig) {
	setDefaultGlobalValues(config)

	if config.Track == "eks" {
		setDefaultEKSValues(&config.EKSConfig)
	} else if config.Track == "wks-ssh" {
		setDefaultSSHValues(&config.WKSConfig.SSHConfig)
	}
}

func processConfig(config *WKPConfig) error {
	if err := checkRequiredValues(config); err != nil {
		return err
	}

	addDefaultValues(config)
	return nil
}

// Public functions to generate specific configuration information for the different cluster types
// from the single config file. These are called from 'wk config' subcommands: 'env', 'cluster', 'machines',
// and 'eks'

// GenerateConfig reads a wkp config file and returns a corresponding nested structure after
// checking for required values and setting defaults as necessary.
func GenerateConfig(path string) (*WKPConfig, error) {
	config, err := readConfig(path)
	if err != nil {
		return nil, err
	}

	if err := processConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// GenerateEnvironmentFromConfig produces a string containing environment variable definitions
// usable by quickstarts based on a nested configuration structure (typically created by GenerateConfig)
func GenerateEnvironmentFromConfig(config *WKPConfig) string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("export TRACK=%s\n", config.Track))
	str.WriteString(fmt.Sprintf("export CLUSTER_NAME=%s\n", config.ClusterName))
	str.WriteString(fmt.Sprintf("export GITHUB_ORG=%s\n", config.GitHubOrg))
	str.WriteString(fmt.Sprintf("export DOCKER_IO_USER=%s\n", config.DockerIOUser))
	str.WriteString(fmt.Sprintf("export DOCKER_IO_PASSWORD_FILE=%s\n", config.DockerIOPasswordFile))
	if config.Track == "eks" {
		str.WriteString(fmt.Sprintf("export REGION=%s\n", config.EKSConfig.ClusterRegion))
	} else {
		str.WriteString(fmt.Sprintf("export SSH_KEY_FILE=%s\n", config.WKSConfig.SSHConfig.SSHKeyFile))
	}

	return str.String()
}

// GenerateMachinesFileContentsFromConfig produces the contents of a machines.yaml file
// usable by quickstarts based on a nested configuration structure (typically created by GenerateConfig)
func GenerateMachinesFileContentsFromConfig(config *WKPConfig) (string, error) {
	t, err := template.New("machine").Parse(machineTemplate)
	if err != nil {
		return "", err
	}

	var str strings.Builder
	str.WriteString("apiVersion: v1\nitems:\n")

	for _, machine := range config.WKSConfig.SSHConfig.Machines {
		var populated bytes.Buffer
		err = t.Execute(&populated, struct {
			Name              string
			Role              string
			KubernetesVersion string
			PublicAddress     string
			PublicPort        int64
			PrivateAddress    string
			PrivatePort       int64
		}{machine.Name, machine.Role, config.WKSConfig.KubernetesVersion, machine.PublicAddress, machine.PublicPort,
			machine.PrivateAddress, machine.PrivatePort})
		if err != nil {
			return "", err
		}
		str.WriteString(populated.String())
	}
	str.WriteString("kind: List\n")
	return str.String(), nil
}

func buildCIDRBlocks(cidrs []string) string {
	var str strings.Builder
	str.WriteString("[")

	firstTime := true
	for _, cidr := range cidrs {
		str.WriteString(cidr)
		if !firstTime {
			str.WriteString(",")
		} else {
			firstTime = false
		}
	}

	str.WriteString("]")
	return str.String()
}

// GenerateClusterFileContentsFromConfig produces the contents of a cluster.yaml file
// usable by quickstarts based on a nested configuration structure (typically created by GenerateConfig)
func GenerateClusterFileContentsFromConfig(config *WKPConfig) (string, error) {
	t, err := template.New("cluster-file").Parse(clusterFileTemplate)
	if err != nil {
		return "", err
	}

	var populated bytes.Buffer
	err = t.Execute(&populated, struct {
		ClusterName       string
		SSHUser           string
		ServiceCIDRBlocks string
		PodCIDRBlocks     string
	}{config.ClusterName,
		config.WKSConfig.SSHConfig.SSHUser,
		buildCIDRBlocks(config.WKSConfig.ServiceCIDRBlocks),
		buildCIDRBlocks(config.WKSConfig.PodCIDRBlocks)})
	if err != nil {
		return "", err
	}
	return populated.String(), nil
}

func generateNodeGroups(nodeGroups []NodeGroupConfig) (string, error) {
	t, err := template.New("eks-nodegroup").Parse(nodeGroupTemplate)
	if err != nil {
		return "", err
	}

	var str strings.Builder
	firstTime := true
	for _, ngroup := range nodeGroups {
		var populated bytes.Buffer
		err = t.Execute(&populated, struct {
			Name            string
			InstanceType    string
			DesiredCapacity int64
		}{ngroup.Name, ngroup.InstanceType, ngroup.DesiredCapacity})
		if err != nil {
			return "", err
		}
		if !firstTime {
			str.WriteString("\n")
		} else {
			firstTime = false
		}
		str.WriteString(populated.String())
	}
	return str.String(), nil
}

// GenerateEKSClusterSpecFromConfig produces an EKSCluster manifest
// usable by quickstarts based on a nested configuration structure (typically created by GenerateConfig)
func GenerateEKSClusterSpecFromConfig(config *WKPConfig) (string, error) {
	t, err := template.New("eks-file").Parse(eksTemplate)
	if err != nil {
		return "", err
	}

	ngroups, err := generateNodeGroups(config.EKSConfig.NodeGroups)
	if err != nil {
		return "", err
	}

	var populated bytes.Buffer
	err = t.Execute(&populated, struct {
		ClusterName          string
		ClusterRegion        string
		KubernetesVersion    string
		NodeGroups           string
		ManagedNodeGroupFile string
	}{config.ClusterName,
		config.EKSConfig.ClusterRegion,
		config.EKSConfig.KubernetesVersion,
		ngroups,
		config.EKSConfig.ManagedNodeGroupFile})

	if err != nil {
		return "", err
	}
	return populated.String(), nil
}

// GenerateFootlooseSpecFromConfig creates a config file that the footloose command can use to generate
// an underlying footloose machine specification.
func GenerateFootlooseSpecFromConfig(config *WKPConfig) (string, error) {
	t, err := template.New("footloose-config").Parse(footlooseTemplate)
	if err != nil {
		return "", err
	}

	var populated bytes.Buffer
	err = t.Execute(&populated, struct {
		Backend           string
		ControlPlaneNodes int64
		WorkerNodes       int64
	}{config.WKSConfig.FootlooseConfig.Backend,
		config.WKSConfig.FootlooseConfig.ControlPlaneNodes,
		config.WKSConfig.FootlooseConfig.WorkerNodes})

	if err != nil {
		return "", err
	}
	return populated.String(), nil
}
