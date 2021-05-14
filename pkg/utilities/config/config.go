package config

import (
	"bytes"
	gcontext "context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	existinginfrav1 "github.com/weaveworks/cluster-api-provider-existinginfra/apis/cluster.weave.works/v1alpha3"
	wksos "github.com/weaveworks/cluster-api-provider-existinginfra/pkg/apis/wksprovider/machine/os"
	"github.com/weaveworks/cluster-api-provider-existinginfra/pkg/cluster/machine"
	"github.com/weaveworks/cluster-api-provider-existinginfra/pkg/plan"
	"github.com/weaveworks/cluster-api-provider-existinginfra/pkg/plan/recipe"
	"github.com/weaveworks/cluster-api-provider-existinginfra/pkg/plan/resource"
	"github.com/weaveworks/cluster-api-provider-existinginfra/pkg/plan/runners/ssh"
	"github.com/weaveworks/cluster-api-provider-existinginfra/pkg/plan/runners/sudo"
	"github.com/weaveworks/cluster-api-provider-existinginfra/pkg/utilities/envcfg"
	"github.com/weaveworks/cluster-api-provider-existinginfra/pkg/utilities/object"
	"github.com/weaveworks/wks/pkg/utilities/versions"
	"github.com/weaveworks/wksctl/pkg/utilities"
	yaml "gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	k8sValidation "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	apierrors "sigs.k8s.io/cluster-api/errors"
)

type GitProvider string

const (
	GitHubProvider GitProvider = "github"
	GitLabProvider GitProvider = "gitlab"
	Namespace      string      = "weavek8sops"
)

// Top-level config parameters
type WKPConfig struct {
	Track                string               `yaml:"track"`
	ClusterName          string               `yaml:"clusterName"`
	GitProvider          GitProvider          `yaml:"gitProvider"`
	GitProviderOrg       string               `yaml:"gitProviderOrg"`
	GitURL               string               `yaml:"gitUrl"`
	DockerIOUser         string               `yaml:"dockerIOUser"`
	DockerIOPasswordFile string               `yaml:"dockerIOPasswordFile"`
	SealedSecretsCert    string               `yaml:"sealedSecretsCertificate"`
	SealedSecretsKey     string               `yaml:"sealedSecretsPrivateKey"`
	EnabledFeatures      EnabledFeatures      `yaml:"enabledFeatures"`
	ExperimentalFeatures ExperimentalFeatures `yaml:"experimentalFeatures,omitempty"`
	FleetManagementDB    FleetManagementDB    `yaml:"fleetManagementDB,omitempty"`
	EKSConfig            EKSConfig            `yaml:"eksConfig"`
	WKSConfig            WKSConfig            `yaml:"wksConfig"`
	ImageRepository      string               `yaml:"imageRepository"`
}

// Map of the database configuration used for the MCCP
type FleetManagementDB struct {
	DatabaseType   string         `yaml:"databaseType"`
	DatabaseURI    string         `yaml:"databaseURI"`
	SqliteConfig   SqliteConfig   `yaml:"sqliteConfig,omitempty"`
	PostgresConfig PostgresConfig `yaml:"postgresConfig,omitempty"`
}

// Map of the sqlite configuration
type SqliteConfig struct {
	HostPathVolume        bool   `yaml:"hostPathVolume"`
	Path                  string `yaml:"path"`
	PersistentVolumeClaim bool   `yaml:"persistentVolumeClaim"`
}

// Map of the postgres configuration
type PostgresConfig struct {
	DatabaseName string `yaml:"databaseName"`
}

// Map of WKP features that can be toggled on/off
type EnabledFeatures struct {
	TeamWorkspaces  bool `yaml:"teamWorkspaces"`
	FleetManagement bool `yaml:"fleetManagement"`
}

// Map of Experimental WKP features that can be toggled on/off
type ExperimentalFeatures struct {
	EKS_D bool `yaml:"eks-d,omitempty"`
}

// Parameters specific to eks
type EKSConfig struct {
	ClusterRegion        string            `yaml:"clusterRegion"`
	KubernetesVersion    string            `yaml:"kubernetesVersion"`
	NodeGroups           []NodeGroupConfig `yaml:"nodeGroups"`
	ManagedNodeGroupFile string            `yaml:"managedNodeGroupFile"`
	UIALBIngress         bool              `yaml:"uiALBIngress"`
	ConfigFilePath       string            `yaml:"configFilePath"`
}

type NodeGroupConfig struct {
	Name            string                       `yaml:"name,omitempty"`
	InstanceType    string                       `yaml:"instanceType,omitempty"`
	DesiredCapacity int64                        `yaml:"desiredCapacity,omitempty"`
	AMIFamily       string                       `yaml:"amiFamily,omitempty"`
	AMI             string                       `yaml:"ami,omitempty"`
	Labels          map[string]string            `yaml:"labels,omitempty"`
	Bottlerocket    *NodeGroupBottlerocketConfig `yaml:"bottlerocket,omitempty"`
}

type NodeGroupBottlerocketConfig struct {
	EnableAdminContainer bool                                 `yaml:"enableAdminContainer,omitempty"`
	Settings             *NodeGroupBottlerocketSettingsConfig `yaml:"settings,omitempty"`
}

type NodeGroupBottlerocketSettingsConfig struct {
	Motd string `yaml:"motd,omitempty"`
}

// Parameters shared by 'footloose' and 'ssh'
type WKSConfig struct {
	KubernetesVersion     string           `yaml:"kubernetesVersion"`
	ServiceCIDRBlocks     []string         `yaml:"serviceCIDRBlocks"`
	PodCIDRBlocks         []string         `yaml:"podCIDRBlocks"`
	MinDiskSpace          uint64           `yaml:"minDiskSpace"`
	SSHConfig             SSHConfig        `yaml:"sshConfig"`
	FootlooseConfig       FootlooseConfig  `yaml:"footlooseConfig"`
	ControlPlaneLbAddress string           `yaml:"controlPlaneLbAddress"`
	CNI                   string           `yaml:"cni"`
	APIServerArguments    []ServerArgument `yaml:"apiServerArguments"`
	KubeletArguments      []ServerArgument `yaml:"kubeletArguments"`
	Flavor                Flavor           `yaml:"flavor"`
}

// Key/value pairs representing generic arguments to the Kubernetes api server
type ServerArgument struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type Flavor struct {
	Name        string `yaml:"name"`
	ManifestURL string `yaml:"manifestURL"`
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
	Image             string `yaml:"image"`
}

// Templates for generating specific configs
// TODO replace with the template from the provider directly
const clusterFileTemplate = `apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: {{ .ClusterName }}
  namespace: {{ .Namespace }}
spec:
  clusterNetwork:
    services:
      cidrBlocks: {{ .ServiceCIDRBlocks }}
    pods:
      cidrBlocks: {{ .PodCIDRBlocks }}
    serviceDomain: cluster.local
  infrastructureRef:
    apiVersion: "cluster.weave.works/v1alpha3"
    kind: ExistingInfraCluster
    name: {{ .ClusterName }}
---
apiVersion: cluster.weave.works/v1alpha3
kind: "ExistingInfraCluster"
metadata:
  name: {{ .ClusterName }}
  namespace: {{ .Namespace }}
spec:
      user: {{ .SSHUser }}
      kubernetesVersion: {{ .KubernetesVersion }}
      cni: "{{ .CNI }}"
      flavor: {{ .Flavor }}
      {{- if or (.ControlPlaneLbAddress) (.APIServerArguments) }}
      {{- if .ControlPlaneLbAddress }}
      controlPlaneEndpoint: {{ .ControlPlaneLbAddress }}
      {{- end }}
      apiServer:
        extraArguments: {{ .APIServerArguments }}
      {{- end }}
      {{- if .KubeletArguments }}
      kubeletArguments: {{ .KubeletArguments }}
      {{- end }}
      controlPlaneMachineCount: "{{ .CPMachineCount }}"
      workerMachineCount: "{{ .WorkerMachineCount }}"
      os:
        files:
        - source:
            configmap: repo
            key: google-cloud.repo
            contents: |
                [google-compute-engine]
                name=Google Compute Engine
                baseurl=https://packages.cloud.google.com/yum/repos/google-compute-engine-el7-x86_64-stable
                enabled=1
                gpgcheck=1
                repo_gpgcheck=0
                gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
                [google-cloud-sdk]
                name=Google Cloud SDK
                baseurl=https://packages.cloud.google.com/yum/repos/cloud-sdk-el7-x86_64
                enabled=1
                gpgcheck=1
                repo_gpgcheck=0
                gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
          destination: /etc/yum.repos.d/google-cloud.repo
        - source:
            configmap: repo
            key: kubernetes.repo
            contents: |
                [kubernetes]
                name=Kubernetes
                baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
                enabled=1
                gpgcheck=1
                repo_gpgcheck=0
                gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
                exclude=kube*
          destination: /etc/yum.repos.d/kubernetes.repo
        - source:
            configmap: repo
            key: docker-ce.repo
            contents: |
                [docker-ce-stable]
                name=Docker CE Stable - $basearch
                baseurl=https://download.docker.com/linux/centos/7/$basearch/stable
                enabled=1
                gpgcheck=1
                gpgkey=https://download.docker.com/linux/centos/gpg

                [docker-ce-stable-debuginfo]
                name=Docker CE Stable - Debuginfo $basearch
                baseurl=https://download.docker.com/linux/centos/7/debug-$basearch/stable
                enabled=0
                gpgcheck=1
                gpgkey=https://download.docker.com/linux/centos/gpg

                [docker-ce-stable-source]
                name=Docker CE Stable - Sources
                baseurl=https://download.docker.com/linux/centos/7/source/stable
                enabled=0
                gpgcheck=1
                gpgkey=https://download.docker.com/linux/centos/gpg

                [docker-ce-edge]
                name=Docker CE Edge - $basearch
                baseurl=https://download.docker.com/linux/centos/7/$basearch/edge
                enabled=0
                gpgcheck=1
                gpgkey=https://download.docker.com/linux/centos/gpg

                [docker-ce-edge-debuginfo]
                name=Docker CE Edge - Debuginfo $basearch
                baseurl=https://download.docker.com/linux/centos/7/debug-$basearch/edge
                enabled=0
                gpgcheck=1
                gpgkey=https://download.docker.com/linux/centos/gpg

                [docker-ce-edge-source]
                name=Docker CE Edge - Sources
                baseurl=https://download.docker.com/linux/centos/7/source/edge
                enabled=0
                gpgcheck=1
                gpgkey=https://download.docker.com/linux/centos/gpg

                [docker-ce-test]
                name=Docker CE Test - $basearch
                baseurl=https://download.docker.com/linux/centos/7/$basearch/test
                enabled=0
                gpgcheck=1
                gpgkey=https://download.docker.com/linux/centos/gpg

                [docker-ce-test-debuginfo]
                name=Docker CE Test - Debuginfo $basearch
                baseurl=https://download.docker.com/linux/centos/7/debug-$basearch/test
                enabled=0
                gpgcheck=1
                gpgkey=https://download.docker.com/linux/centos/gpg

                [docker-ce-test-source]
                name=Docker CE Test - Sources
                baseurl=https://download.docker.com/linux/centos/7/source/test
                enabled=0
                gpgcheck=1
                gpgkey=https://download.docker.com/linux/centos/gpg

                [docker-ce-nightly]
                name=Docker CE Nightly - $basearch
                baseurl=https://download.docker.com/linux/centos/7/$basearch/nightly
                enabled=0
                gpgcheck=1
                gpgkey=https://download.docker.com/linux/centos/gpg

                [docker-ce-nightly-debuginfo]
                name=Docker CE Nightly - Debuginfo $basearch
                baseurl=https://download.docker.com/linux/centos/7/debug-$basearch/nightly
                enabled=0
                gpgcheck=1
                gpgkey=https://download.docker.com/linux/centos/gpg

                [docker-ce-nightly-source]
                name=Docker CE Nightly - Sources
                baseurl=https://download.docker.com/linux/centos/7/source/nightly
                enabled=0
                gpgcheck=1
                gpgkey=https://download.docker.com/linux/centos/gpg
          destination: /etc/yum.repos.d/docker-ce.repo
        - source:
            configmap: repo
            key: cloud-google-com.gpg.b64
            contents: |
                xsBNBF/Jfl4BCADTPUXdkNu057X+P3STVxCzJpU2Mn+tUamKdSdVambGeYFINcp/EGwNGhdb0a1B
                bHs1SWYZbzwh4d6+p3k4ABzVMO+RpMu/aBx9E5aOn5c8GzHjZ/VEaheqLLhSUcSCzChSZcN5jz0h
                TGhmAGaviMt6RMzSfbIhZPj1kDzBiGd0Qwd/rOPnJr4taPruR3ecBjhHti1/BMGd/lj0F7zQnCjp
                7PrqgpEPBT8jo9wX2wvOyXswSI/GsfbFiaOJfDnYengaEg8sF+u3WOs0Z20cSr6kS76KHpTfa3Jj
                YsfHt8NDw8w4e3H8PwQzNiRP9tXeMASKQz3emMj/ek6HxjihY9qFABEBAAHNumdMaW51eCBSYXB0
                dXJlIEF1dG9tYXRpYyBTaWduaW5nIEtleSAoLy9kZXBvdC9nb29nbGUzL3Byb2R1Y3Rpb24vYm9y
                Zy9jbG91ZC1yYXB0dXJlL2tleXMvY2xvdWQtcmFwdHVyZS1wdWJrZXlzL2Nsb3VkLXJhcHR1cmUt
                c2lnbmluZy1rZXktMjAyMC0xMi0wMy0xNl8wOF8wNS5wdWIpIDxnbGludXgtdGVhbUBnb29nbGUu
                Y29tPsLAaAQTAQgAHAUCX8l+XgkQi1fFwoNvS+sCGwMFCQPDCrACGQEAAEF6CACaekro6aUJJd3m
                VtrtLOOewV8et1jep5ewmpOrew/pajRVBeIbV1awVn0/8EcenFejmP6WFcdCWouDVIS/QmRFQV9N
                6YXN8PiwalrRV3bTKFBHkwa1cEH4AafCGo0cDvJb8N3JnM/Rmb1KSGKr7ZXpmkLtYVqr6Hgzl+sn
                rlH0Xwsl5r3SyvqBgvRYTQKZpKqmBEd1udieVoLSF988kKeNDjFa+Q1SjZPGW+XukgE8kBUbSDx8
                Y8q6Cszh3VVY+5JUeqimRgJ2ADY2/3lEtAZOtmwcBlhY0cPWVqga14E7kTGSWKC6W96Nfy9K7L4Y
                pp8nTMErus181aqwwNfMqnpnzsBNBF/Jfl4BCADDSh+KdBeNjIclVVnRKt0QT5593yF4WVZt/TgN
                uaEZ5vKknooVVIq+cJIfY/3lUqq8Te4dEjodtFyKe5Xuego6qjzs8TYFdCAHXpXRoUolT14m+qkJ
                8rhSrpN0TxIjWJbJdm3NlrgTam5RKJw3ShypNUxyolnHelXxqyKDCkxBSDmR6xcdft3wdQl5IkIA
                wxe6nywmSUtpndGLRJdJraJiaWF2IBjFNg3vTEYj4eoehZd4XrvEyLVrMbKZ5m6f1o6QURuzSrUH
                9JT/ivZqCmhPposClXXX0bbi9K0Z/+uVyk6v76ms3O50rIq0L0YehM8G++qmGO421+0qCLkdD5/j
                ABEBAAHCwF8EGAEIABMFAl/Jfl4JEItXxcKDb0vrAhsMAAAbGggAw7lhSWElZpGV1SI2b2K26PB9
                3fVI1tQYV37WIElCJsajF+/ZDfJJ2d6ncuQSleH5WRccc4hZfKwysA/epqrCnwc7yKsToZ4sw8xs
                JF1UtQ5ENtkdArViBJHS4Y2VZ5DEUmr5EghGtZFh9a6aLoeMVM/nrZCLstDVoPKEpLokHu/gebCw
                fT/n9U1dolFIovg6eKACl5xOx+rzcAVp7R4P527jffudz3dKMdLhPrstG0w5YbyfPPwWMOPp+kUF
                45eYdR7kKKk09VrJNkEGJ0KQQ6imqR1Tn0kyu4cvkfqnCUF0rrn7CdBqLSCv1QRhgr6TChQf7ynW
                sPz5gGdVjh3tI8bATQRgPRBZAQgAtYpc0k9MJ7PrsGchAOSFbWHsgLl02kFBAHe9EqiJUKQ3eBMl
                Ysd0gmp0CLvHRvWat/sdvFgW9jrlz/aHNOsmzlnbtpuzeT2NAVE+AjgN+iVf2K8ZjbPufzPmJwx6
                ab+t44ESDpM181zaOksE7JdsRvXygd00tCDLwZFncOTxqwTORoIUXHnIKEgAMEW1iVzkRxilcJVe
                rTsUGf8agNPITyZ3jH7DBTzl7IrYBkR6F45VFi1Xie9JpiGLAv6QYJSMAs5nQ/BHt/TK5Ul27l1U
                Is9/Ih35712KSxJoDVysyNAx/bSoPN9t5AC86miZSxTiyZv7lSV0VBHykty4VWUDMwARAQABzVFS
                YXB0dXJlIEF1dG9tYXRpYyBTaWduaW5nIEtleSAoY2xvdWQtcmFwdHVyZS1zaWduaW5nLWtleS0y
                MDIxLTAzLTAxLTA4XzAxXzA5LnB1YinCwGgEEwEIABwFAmA9EFkJEP7qkWkwfqBxAhsDBQkDwwqw
                AhkBAAB7ZQgAsUljKd8kXC5rB4cRg7efZ4UjV4aLlojXj0jHubxE0AP5YYqfWcfzT0QmuKuY6SAw
                ZRGDoOu2Gp87XI0lhkiN+V25auNx+Li0sYeD7Ss2TKPlI/J9lTRzmVwXRnLDg3FN8pxeuK+3k0Hr
                1HtmlNCjdqOuejtx6xOIrTlSmMJ55JjbJBuOW/W+wyZ7EOlj7M1HPJTYbGtoASOr3y5evL44+z5V
                sNN9ATP0aDBD6aDgKaIR6LH5zYcSZhNQMcAZDBM8qNpGYT2RofOSw5w2wL40hSqmEj0XipkRYy5a
                Nwz1R2f3XkJ+p6B24FAoS6NtRXn4ZWTurcrK29vNzFjCMmP2Es7ATQRgPRBZAQgA3HTvwMNarnWT
                kWQjS89704kEhXFBWMknHySZ8FLIPH8tJIIPaJRWNBiuYnE+p/7IXNUZSKbqqzkGAWYLSt3UmXzg
                FxNjdtB1Lwvp6yirl11/o3DP19ZB8cF+bRunwdX8jR9Kf0KrMxH2ERybtGOD6J02CLJSE5xM5TeI
                VDev5sdfplj5eD+Ee/4evqe0No7WgpRLXXRdHnjn9ejGuUvH33/NLmQiyaFbt5Tlwk9tqAn+6ph9
                l3XZqhorFEnKsJm5rr99LXUHnZ/vJ4yqNqX6VRdTmuuwlkV3Sk5J7mcm8SPSKXIr8vAiEi9g6NLs
                4o+0ke5HlX+xtUNyt4idMJ+pgwARAQABwsBfBBgBCAATBQJgPRBZCRD+6pFpMH6gcQIbDAAAP9wI
                ABSdoRKdteOH84LTVhzlRb9u4bKzu8GBWcKInPZR0peIhMPJiXP95BF3YPVX/Ztc2xv5GerJZs6X
                7+8wwHTd4dx09Adcq298V80V9M4TmAG0ElJ3Og3poQ2aA1rf8FXHin873mwfVUw80QVFc8Qnbr2O
                oo9KdgD2aZ06857wj6Ah5H8wTAt2cpNRbnoj0z6D9fTNAT66DMvKg1UpBa9Ll9zzOeIUDephkUIO
                R1VQcVDWjJ59sjkHMW0P0/3SpaI3aUZr6RsmI3678hMRPKMGJ/C+5ctje+hnGOpIjdQpk5woHa21
                NEj2nJu128U2JUB8CQhGvR3+P57ogWscFyrnP8s=
          destination: /tmp/cloud-google-com.gpg.b64
        - source:
            configmap: docker
            key: daemon.json
            contents: |
              {
                "log-driver": "json-file",
                "log-opts": {
                  "max-size": "100m"
                },
                "exec-opts": [
                  "native.cgroupdriver=cgroupfs"
                ]
              }
          destination: /etc/docker/daemon.json
      cri:
        kind: docker
        package: docker-ce
        version: 19.03.8
      {{- if .ImageRepository }}
      imageRepository: {{ .ImageRepository }}
      {{- end }}
`

const machineTemplate = `---
apiVersion: cluster.x-k8s.io/v1alpha3
kind: Machine
metadata:
  labels:
    set: {{ .Role }}
  name: {{ .Name }}
  namespace: {{ .Namespace }}
spec:
  clusterName: {{ .ClusterName }}
  version: {{ .KubernetesVersion }}
  infrastructureRef:
    apiVersion: "cluster.weave.works/v1alpha3"
    kind: ExistingInfraMachine
    name: {{ .Name }}
  bootstrap: {}
---
apiVersion: "cluster.weave.works/v1alpha3"
kind: "ExistingInfraMachine"
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
  labels:
    cluster.x-k8s.io/cluster-name: {{ .ClusterName }}
spec:
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
    name: {{ .Name }}
    {{- if .AMIFamily }}
    amiFamily: {{ .AMIFamily }}
    {{- end }}
    {{- if .AMI }}
    ami: {{ .AMI }}
    {{- end }}
    {{- if .Labels }}
    labels:
    {{- range $key, $value := .Labels }}
      "{{ $key }}": "{{ $value }}"
    {{- end }}
    {{- end }}
    {{- if .Bottlerocket }}
    bottlerocket:
      {{- if .Bottlerocket.EnableAdminContainer }}
      enableAdminContainer: {{ .Bottlerocket.EnableAdminContainer }}
      {{- end }}
      {{- if .Bottlerocket.Settings }}
      settings:
      {{- if .Bottlerocket.Settings.Motd }}
        motd: "{{ .Bottlerocket.Settings.Motd }}"
      {{- end }}
      {{- end }}
    {{- end }}`

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
  version: '{{ .KubernetesVersion }}'
`

const footlooseTemplate = `# This file contains high level configuration parameters. The setup.sh script
# takes this file as input and creates lower level manifests.
# backend defines how the machines underpinning Kubernetes nodes are created.
#  - docker: use containers as "VMs" using footloose:
#            https://github.com/weaveworks/footloose
#  - ignite: use footloose with ignite and firecracker to create real VMs using:
#            the ignite backend only works on linux as it requires KVM.
#            https://github.com/weaveworks/ignite.
clusterName: {{ .ClusterName }}
backend: {{ .Backend }}
# Number of nodes allocated for the Kubernetes control plane and workers.
controlPlane:
  nodes: {{ .ControlPlaneNodes }}
workers:
  nodes: {{ .WorkerNodes }}
image: {{ .Image }}
kubernetesVersion: {{ .KubernetesVersion }}
`

const haproxyTemplate = `#---------------------------------------------------------------------
# HAProxy configuration file for the Kubernetes API service.
#
# See the full configuration options online at:
#
#   http://haproxy.1wt.eu/download/1.4/doc/configuration.txt
#
#---------------------------------------------------------------------

#---------------------------------------------------------------------
# Global settings
#---------------------------------------------------------------------
global
    log         127.0.0.1 local2

    pidfile     /var/run/haproxy.pid
    maxconn     4000
    daemon

    # turn on stats unix socket
    stats socket /var/lib/haproxy/stats

#---------------------------------------------------------------------
# common defaults that all the 'listen' and 'backend' sections will
# use if not designated in their block
#---------------------------------------------------------------------
defaults
    mode                    http
    log                     global
    option                  httplog
    option                  dontlognull
    option http-server-close
    option forwardfor       except 127.0.0.0/8
    option                  redispatch
    retries                 3
    timeout http-request    10s
    timeout queue           1m
    timeout connect         10s
    timeout client          1m
    timeout server          1m
    timeout http-keep-alive 10s
    timeout check           10s
    maxconn                 3000

#---------------------------------------------------------------------
# OPTIONAL - stats UI that allows you to see which masters have joined
#            the LB roundrobin
#---------------------------------------------------------------------
frontend stats
    bind *:8404
    stats enable
    stats uri /stats
    stats refresh 10s
    stats admin if LOCALHOST

#---------------------------------------------------------------------
# KubeAPI frontend which proxys to the master nodes
#---------------------------------------------------------------------
frontend kubernetes
    bind *:6443
    default_backend             kubernetes
    mode tcp
    option tcplog

backend kubernetes
    balance     roundrobin
    mode tcp
    option tcp-check
    default-server inter 10s downinter 5s rise 2 fall 2 slowstart 60s maxconn 250 maxqueue 256 weight 100
`

var (
	cidrRegexp                  = regexp.MustCompile(`^([0-9]{1,3}\.){3}[0-9]{1,3}(\/([0-9]|[1-2][0-9]|3[0-2]))?$`)
	controlPlaneLbAddressRegexp = regexp.MustCompile(`^((([0-9]{1,3}\.){3}[0-9]{1,3})|(([a-zA-Z0-9]+(-[a-zA-Z0-9]+)*\.)+[a-zA-Z]{2,}))$`)
)

func unmarshalConfig(configBytes []byte) (*WKPConfig, error) {
	var config WKPConfig
	err := yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// ReadConfig loads a config from the file system into the structs from above
func ReadConfig(path string) (*WKPConfig, error) {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return unmarshalConfig(fileBytes)
}

// SetNodeValue finds the node at the given path, and sets its value
func SetNodeValue(config *yaml.Node, nodePath []string, value string) error {
	currentNode := config
	var errCode int
	for _, node := range nodePath {
		currentNode, errCode = findMapNode(currentNode, node)
		if errCode == 0 {
			return errors.New(fmt.Sprintf("did not find node %v in config.yaml", node))
		}
	}
	currentNode.Value = value
	return nil
}

func findMapNode(n *yaml.Node, key string) (*yaml.Node, int) {
	switch n.Kind {
	case yaml.DocumentNode:
		for _, c := range n.Content {
			if r, p := findMapNode(c, key); r != nil {
				return r, p
			}
		}
	case yaml.MappingNode:
		for i := 0; i < len(n.Content)/2; i++ {
			if n.Content[i*2].Value == key {
				p := i*2 + 1
				return n.Content[p], p
			}
		}
	}
	return nil, 0
}

// WriteConfig writes a modified config.yaml back to the file system
func WriteConfig(path string, config interface{}) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0)
	if err != nil {
		return errors.Wrapf(err, fmt.Sprintf("failed to open file at path %v", path))
	}
	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	err = encoder.Encode(config)
	if err != nil {
		return errors.Wrapf(err, "failed to encode parsed config")
	}
	return nil
}

func createClusterName(env map[string]string) string {
	name := env["USER"]
	if name == "" {
		name = "cluster" // use "wk-cluster" if no user env var found
	}
	return "wk-" + strings.ToLower(name)
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
	if config.DockerIOUser == "" {
		return fmt.Errorf("dockerIOUser must be specified")
	}

	if config.DockerIOPasswordFile == "" {
		return fmt.Errorf("dockerIOPasswordFile must be specified")
	}

	if err := checkValidPath("dockerIOPasswordFile", config.DockerIOPasswordFile); err != nil {
		return err
	}

	if config.ClusterName != "" {
		errs := k8sValidation.IsDNS1123Subdomain(config.ClusterName)
		if len(errs) > 0 {
			return fmt.Errorf("Invalid clusterName: \"%v\", %s", config.ClusterName, strings.Join(errs, ". "))
		}
	}

	switch config.Track {
	case "":
		return fmt.Errorf("track must be specified")
	case "eks", "wks-ssh", "wks-footloose", "wks-components":
		return nil
	default:
		return fmt.Errorf("track must be one of: 'eks', 'wks-ssh', 'wks-footloose' or 'wks-components'")
	}
}

func setDefaultGlobalValues(config *WKPConfig, env map[string]string) {
	if config.ClusterName == "" {
		config.ClusterName = createClusterName(env)
	}
}

func checkRequiredGitValues(config *WKPConfig) error {
	if config.GitProvider != GitHubProvider && config.GitProvider != GitLabProvider {
		return fmt.Errorf("gitProvider must be one of: 'github' or 'gitlab'")
	}

	// All good.
	if config.GitURL != "" {
		if !strings.HasPrefix(config.GitURL, "git@") && !strings.HasPrefix(config.GitURL, "ssh://git@") {
			return fmt.Errorf("gitUrl, if provided, must be a git ssh url that starts with 'git@' or 'ssh://git@'")
		}

		return nil
	}

	// We don't actually support creating gitlab repos right now.
	if config.GitProvider == GitLabProvider {
		return fmt.Errorf("Please provide the url to your gitlab git repository in: gitUrl")
	}

	// Want us to create a github repo tell us the org
	if config.GitProvider == GitHubProvider && config.GitProviderOrg == "" {
		return fmt.Errorf("Please provide the gitProviderOrg where the repository will be created")
	}

	return nil
}

func validateSealedSecretsValues(config *WKPConfig) error {
	// Check that if both certificate and private key they match, or both are left blank
	if config.SealedSecretsCert != "" && config.SealedSecretsKey == "" ||
		config.SealedSecretsCert == "" && config.SealedSecretsKey != "" {
		return fmt.Errorf("please provide both the private key and certificate for the sealed secrets controller")
	} else if config.SealedSecretsCert != "" && config.SealedSecretsKey != "" {
		// Check if cert file exists
		if _, err := os.Stat(config.SealedSecretsCert); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf(`could not find certificate at path: %s
If you specified a relative path, note that it will be evaluated from the directory of your config.yaml`, config.SealedSecretsCert)
			}
		}

		// Check if key file exists
		if _, err := os.Stat(config.SealedSecretsKey); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf(`could not find key at path: %s
If you specified a relative path, note that it will be evaluated from the directory of your config.yaml`, config.SealedSecretsKey)
			}
		}

		_, err := tls.LoadX509KeyPair(config.SealedSecretsCert, config.SealedSecretsKey)
		if err != nil {
			return fmt.Errorf("provided private key and certificate do not match")
		}
	}
	return nil
}

func validateExperimentalFeatures(config *WKPConfig) error {
	if !config.ExperimentalFeatures.EKS_D && (config.WKSConfig.Flavor.Name != "" || config.WKSConfig.CNI != "") {
		return fmt.Errorf("Flavors and CNI overrides are not enabled; enable the experimental 'eks-d' feature to use them")
	}
	return nil
}

// eks values
func checkRequiredEKSValues(eksConfig *EKSConfig) error {
	if eksConfig.ConfigFilePath != "" {
		if _, err := os.Stat(eksConfig.ConfigFilePath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("could not find eksctl config file at path: %s", eksConfig.ConfigFilePath)
			}
		} else {
			return nil
		}
	}

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
	// TODO can we retrieve this value dynamically?
	switch eksConfig.KubernetesVersion {
	case "":
		return fmt.Errorf("A Kubernetes version must be specified")
	case "1.16", "1.17", "1.18", "1.19":
		return nil
	default:
		return fmt.Errorf(`Kubernetes version must be one of: "1.16", "1.17", "1.18" or "1.19"`)
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

	if wksConfig.ControlPlaneLbAddress != "" {
		if !controlPlaneLbAddressRegexp.MatchString(wksConfig.ControlPlaneLbAddress) {
			return fmt.Errorf("%s is not a valid control plane load balancer address; must be a valid IP address or a domain name", wksConfig.ControlPlaneLbAddress)
		}
	}

	err := versions.CheckValidVersion(wksConfig.KubernetesVersion)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	if len(wksConfig.ServiceCIDRBlocks) == 0 {
		return fmt.Errorf("A service CIDR block must be specified")
	}

	for _, cidr := range wksConfig.ServiceCIDRBlocks {
		if !cidrRegexp.MatchString(cidr) {
			return fmt.Errorf("%s is not a valid CIDR specification", cidr)
		}
	}

	if len(wksConfig.PodCIDRBlocks) == 0 {
		return fmt.Errorf("A pod CIDR block must be specified")
	}

	for _, cidr := range wksConfig.PodCIDRBlocks {
		if !cidrRegexp.MatchString(cidr) {
			return fmt.Errorf("%s is not a valid CIDR specification", cidr)
		}
	}

	return nil
}

// ssh values
func checkRequiredSSHValues(config *WKPConfig) error {
	if config.WKSConfig.SSHConfig.SSHKeyFile == "" {
		homedir := os.Getenv("HOME")
		if homedir == "" {
			return fmt.Errorf("no ssh key file specified and no home directory information available")
		}
	} else if err := checkValidPath("sshKeyFile", config.WKSConfig.SSHConfig.SSHKeyFile); err != nil {
		return err
	}

	if len(config.WKSConfig.SSHConfig.Machines) == 0 {
		return fmt.Errorf("No machine information provided")
	}

	masters, workers, err := getMachineCounts(config)
	if err != nil {
		return err
	}
	if masters == 0 || workers == 0 {
		return fmt.Errorf("Invalid machine set. At least one master and one worker must be specified.")
	}

	return nil
}

func getMachineCounts(config *WKPConfig) (int64, int64, error) {
	if config.Track == "wks-footloose" {
		return config.WKSConfig.FootlooseConfig.ControlPlaneNodes, config.WKSConfig.FootlooseConfig.WorkerNodes, nil
	}
	var masters, workers int64
	for idx := range config.WKSConfig.SSHConfig.Machines {
		machine := &config.WKSConfig.SSHConfig.Machines[idx]

		if machine.PublicAddress == "" {
			return 0, 0, fmt.Errorf("a public address must be specified for each machine")
		}

		switch machine.Role {
		case "":
			return 0, 0, fmt.Errorf("a role ('master' or 'worker') must be specified for each machine")
		case "master":
			masters++
		case "worker":
			workers++
		default:
			return 0, 0, fmt.Errorf("Invalid machine role: '%s'. Only 'master' and 'worker' are valid.",
				machine.Role)
		}
	}
	return masters, workers, nil
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
		break
	default:
		return fmt.Errorf("A footloose backend must be either 'docker' or 'ignite'")
	}

	if footlooseConfig.Image != "" {
		if _, err := reference.ParseNamed(footlooseConfig.Image); err != nil {
			return errors.Wrapf(err, "Invalid footloose image reference: '%s'", footlooseConfig.Image)
		}
	}

	return nil
}

// If fleetManagement is enabled, assert that the values set in fleetManagementDB are valid
func checkRequiredFleetManagementDBValues(config *FleetManagementDB) error {
	emptyFleetManagementDBConfig := FleetManagementDB{}
	emptySqliteConfig := SqliteConfig{}
	emptyPostgresConfig := PostgresConfig{}

	if *config == emptyFleetManagementDBConfig {
		return errors.New("fleet management feature is enabled, but its database configuration is missing from config.yaml")
	}

	switch config.DatabaseType {
	case "sqlite":
		if config.SqliteConfig == emptySqliteConfig {
			return errors.New("fleetManagementDB: missing sqliteConfig section from config.yaml")
		}

		if config.SqliteConfig.HostPathVolume && config.SqliteConfig.PersistentVolumeClaim {
			return errors.New("sqliteConfig: please set either HostPathVolume or PersistentVolumeClaim")
		}
	case "postgres":
		if config.PostgresConfig == emptyPostgresConfig {
			return errors.New("fleetManagementDB: missing postgresConfig section from config.yaml")
		}
	default:
		return errors.New("fleetManagementDB: unsupported database type, supported values: sqlite, postgres")
	}

	return nil
}

func checkRequiredValues(config *WKPConfig) error {
	if err := checkRequiredGlobalValues(config); err != nil {
		return err
	}

	if err := checkRequiredGitValues(config); err != nil {
		return err
	}

	if config.Track == "wks-components" {
		return nil
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
		if err := checkRequiredSSHValues(config); err != nil {
			return err
		}
	case "wks-footloose":
		if err := checkRequiredFootlooseValues(&config.WKSConfig.FootlooseConfig); err != nil {
			return err
		}
	}

	if config.EnabledFeatures.FleetManagement {
		if err := checkRequiredFleetManagementDBValues(&config.FleetManagementDB); err != nil {
			return err
		}
	}

	return nil
}

func getEnvironMap() map[string]string {
	env := map[string]string{}
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		env[pair[0]] = pair[1]
	}
	return env
}

func addDefaultValues(config *WKPConfig) {
	setDefaultGlobalValues(config, getEnvironMap())

	if config.Track == "eks" {
		setDefaultEKSValues(&config.EKSConfig)
	} else if config.Track == "wks-ssh" {
		setDefaultSSHValues(&config.WKSConfig.SSHConfig)
	}
}

func processConfig(config *WKPConfig) error {
	addDefaultValues(config)

	if err := checkRequiredValues(config); err != nil {
		return err
	}

	if err := validateSealedSecretsValues(config); err != nil {
		return err
	}

	if err := validateExperimentalFeatures(config); err != nil {
		return err
	}

	return nil
}

// Public functions to generate specific configuration information for the different cluster types
// from the single config file. These are called from 'wk config' subcommands: 'env', 'cluster', 'machines',
// and 'eks'

// GenerateConfig reads a wkp config file and returns a corresponding nested structure after
// checking for required values and setting defaults as necessary.
func GenerateConfig(path string) (*WKPConfig, error) {
	config, err := ReadConfig(path)
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
	str.WriteString(fmt.Sprintf("export GIT_PROVIDER=%s\n", config.GitProvider))
	str.WriteString(fmt.Sprintf("export GIT_PROVIDER_ORG=%s\n", config.GitProviderOrg))
	str.WriteString(fmt.Sprintf("export GIT_URL=%s\n", config.GitURL))
	str.WriteString(fmt.Sprintf("export DOCKER_IO_USER=%s\n", config.DockerIOUser))
	str.WriteString(fmt.Sprintf("export DOCKER_IO_PASSWORD_FILE=%s\n", config.DockerIOPasswordFile))
	str.WriteString(fmt.Sprintf("export SEALED_SECRETS_CERT=%s\n", config.SealedSecretsCert))
	str.WriteString(fmt.Sprintf("export SEALED_SECRETS_KEY=%s\n", config.SealedSecretsKey))
	if config.Track == "eks" {
		if config.EKSConfig.ConfigFilePath != "" {
			str.WriteString(fmt.Sprintf("export EKSCTL_CONFIG_FILE=%s\n", config.EKSConfig.ConfigFilePath))
		} else {
			str.WriteString(fmt.Sprintf("export REGION=%s\n", config.EKSConfig.ClusterRegion))
		}
	} else {
		str.WriteString(fmt.Sprintf("export KUBERNETES_VERSION=%s\n", config.WKSConfig.KubernetesVersion))
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

	for _, machine := range config.WKSConfig.SSHConfig.Machines {
		var populated bytes.Buffer
		err = t.Execute(&populated, struct {
			ClusterName       string
			Name              string
			Namespace         string
			Role              string
			KubernetesVersion string
			PublicAddress     string
			PublicPort        int64
			PrivateAddress    string
			PrivatePort       int64
		}{config.ClusterName, machine.Name, Namespace, machine.Role, config.WKSConfig.KubernetesVersion, machine.PublicAddress, machine.PublicPort,
			machine.PrivateAddress, machine.PrivatePort})
		if err != nil {
			return "", err
		}
		str.WriteString(populated.String())
	}
	return str.String(), nil
}

func buildCIDRBlocks(cidrs []string) string {
	var str strings.Builder
	str.WriteString("[")

	firstTime := true
	for _, cidr := range cidrs {
		if !firstTime {
			str.WriteString(",")
		} else {
			firstTime = false
		}
		str.WriteString(cidr)
	}

	str.WriteString("]")
	return str.String()
}

func buildServerArguments(args []ServerArgument) string {
	var str strings.Builder
	str.WriteString("[")

	firstTime := true
	for _, arg := range args {
		if !firstTime {
			str.WriteString(",")
		} else {
			firstTime = false
		}
		str.WriteString(`{"name":"`)
		str.WriteString(arg.Name)
		str.WriteString(`","value":"`)
		str.WriteString(arg.Value)
		str.WriteString(`"}`)
	}

	str.WriteString("]")
	return str.String()
}

func buildFlavor(flavor Flavor) string {
	return fmt.Sprintf(`{"name": "%s", "manifestURL": "%s"}`, flavor.Name, flavor.ManifestURL)
}

// GenerateClusterFileContentsFromConfig produces the contents of a cluster.yaml file
// usable by quickstarts based on a nested configuration structure (typically created by GenerateConfig)
func GenerateClusterFileContentsFromConfig(config *WKPConfig, configDir string) (string, error) {
	t, err := template.New("cluster-file").Parse(clusterFileTemplate)
	if err != nil {
		return "", err
	}

	controlPlanes, workers, err := getMachineCounts(config)
	if err != nil {
		return "", err
	}

	var populated bytes.Buffer
	err = t.Execute(&populated, struct {
		ClusterName           string
		Namespace             string
		SSHUser               string
		KubernetesVersion     string
		CNI                   string
		Flavor                string
		ServiceCIDRBlocks     string
		PodCIDRBlocks         string
		APIServerArguments    string
		KubeletArguments      string
		ControlPlaneLbAddress string
		ImageRepository       string
		CPMachineCount        string
		WorkerMachineCount    string
	}{
		config.ClusterName,
		Namespace,
		config.WKSConfig.SSHConfig.SSHUser,
		config.WKSConfig.KubernetesVersion,
		config.WKSConfig.CNI,
		buildFlavor(config.WKSConfig.Flavor),
		buildCIDRBlocks(config.WKSConfig.ServiceCIDRBlocks),
		buildCIDRBlocks(config.WKSConfig.PodCIDRBlocks),
		buildServerArguments(config.WKSConfig.APIServerArguments),
		buildServerArguments(config.WKSConfig.KubeletArguments),
		getLoadBalancerAddress(config, configDir),
		config.ImageRepository,
		strconv.FormatInt(controlPlanes, 10),
		strconv.FormatInt(workers, 10),
	})

	if err != nil {
		return "", err
	}
	return populated.String(), nil
}

func getLoadBalancerPublicAddress(conf *WKPConfig) string {
	if conf.Track == "wks-footloose" && conf.WKSConfig.FootlooseConfig.ControlPlaneNodes > 1 {
		return "127.0.0.1"
	}
	return conf.WKSConfig.ControlPlaneLbAddress
}

func getLoadBalancerAddress(conf *WKPConfig, configDir string) string {
	if conf.Track == "wks-footloose" && conf.WKSConfig.FootlooseConfig.ControlPlaneNodes > 1 {
		ips, err := getPrivateIPsFromMachines(configDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not retrieve IPs\n")
			os.Exit(1)
		}
		return incIP(ips[len(ips)-1])
	}
	return conf.WKSConfig.ControlPlaneLbAddress
}

func incIP(ip string) string {
	octets := strings.Split(ip, ".")
	num, err := strconv.Atoi(octets[len(octets)-1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid IP: %s\n", ip)
		os.Exit(1)
	}
	octets = append(octets[0:3], fmt.Sprintf("%d", num+1))
	return strings.Join(octets, ".")
}

func generateNodeGroups(nodeGroups []NodeGroupConfig) (string, error) {
	t, err := template.New("eks-nodegroup").Parse(nodeGroupTemplate)
	if err != nil {
		return "", err
	}

	var str strings.Builder
	firstTime := true
	for _, nodeGroup := range nodeGroups {
		var populated bytes.Buffer
		err = t.Execute(&populated, nodeGroup)
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
		ClusterName       string
		Backend           string
		ControlPlaneNodes int64
		WorkerNodes       int64
		Image             string
		KubernetesVersion string
	}{
		config.ClusterName,
		config.WKSConfig.FootlooseConfig.Backend,
		config.WKSConfig.FootlooseConfig.ControlPlaneNodes,
		config.WKSConfig.FootlooseConfig.WorkerNodes,
		config.WKSConfig.FootlooseConfig.Image,
		config.WKSConfig.KubernetesVersion,
	})

	if err != nil {
		return "", err
	}
	return populated.String(), nil
}

func getPrivateIPsFromMachines(configDir string) ([]string, error) {
	machinesManifestPath := filepath.Join(configDir, "machines.yaml")

	errorsHandler := func(machines []*clusterv1.Machine, eim []*existinginfrav1.ExistingInfraMachine, errors field.ErrorList) ([]*clusterv1.Machine, []*existinginfrav1.ExistingInfraMachine, error) {
		if len(errors) > 0 {
			utilities.PrintErrors(errors)
			return nil, nil, apierrors.InvalidMachineConfiguration(
				"%s failed validation, use --skip-validation to force the operation",
				machinesManifestPath)
		}
		return machines, eim, nil
	}

	_, bl, err := machine.ParseAndDefaultAndValidate(machinesManifestPath, errorsHandler)
	if err != nil {
		return nil, err
	}

	results := []string{}
	for _, m := range bl {
		results = append(results, m.Spec.Private.Address)
	}
	return results, nil
}

func generateHAConfiguration(clusterIPs []string) string {
	var str strings.Builder
	str.WriteString(haproxyTemplate)

	for idx, IP := range clusterIPs {
		str.WriteString(fmt.Sprintf("    server master-%d %s:6443 check\n", idx, IP))
	}

	return str.String()
}

func buildDockerConfigResource(configDir string) (plan.Resource, error) {
	b := plan.NewBuilder()
	filespecs := []struct{ sourcePath, key, destPath string }{
		{"repo-config.yaml", "docker-ce.repo", "/etc/yum.repos.d/docker-ce.repo"},
		{"docker-config.yaml", "daemon.json", "/etc/docker/daemon.json"},
	}

	for idx, spec := range filespecs {
		configPath := filepath.Join(configDir, spec.sourcePath)
		contents, err := ioutil.ReadFile(configPath)
		if err != nil {
			return nil, err
		}
		configMap := &v1.ConfigMap{}
		if err := yaml.Unmarshal(contents, configMap); err != nil {
			return nil, errors.Wrapf(err, "failed to parse config:\n%s", contents)
		}
		fileResource := &resource.File{Destination: spec.destPath}
		fileContents, ok := configMap.Data[spec.key]
		if !ok {
			return nil, fmt.Errorf("No config data for in %q", configPath)
		}
		fileResource.Content = fileContents
		b.AddResource(fmt.Sprintf("install-config-file-%d", idx), fileResource)
	}
	p, err := b.Plan()
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// ConfigureHAProxy takes a WKPConfig which specifies a load balancer and configures the load balancer machine with ha proxy.
func ConfigureHAProxy(conf *WKPConfig, configDir string, loadBalancerSSHPort int) error {
	keyFile := conf.WKSConfig.SSHConfig.SSHKeyFile
	lbAddress := getLoadBalancerPublicAddress(conf)

	if keyFile == "" && conf.Track == "wks-footloose" {
		keyFile = filepath.Join(configDir, "cluster-key")
	}

	sshClient, err := ssh.NewClient(ssh.ClientParams{
		User:           conf.WKSConfig.SSHConfig.SSHUser,
		Host:           lbAddress,
		Port:           uint16(loadBalancerSSHPort),
		PrivateKeyPath: keyFile,
	})
	if err != nil {
		return err
	}
	defer sshClient.Close()
	ctx := gcontext.Background()
	installer, err := wksos.Identify(ctx, sshClient)
	if err != nil {
		return errors.Wrapf(err, "failed to identify operating system for haproxy node (%s)",
			lbAddress)
	}

	runner := &sudo.Runner{Runner: sshClient}

	cfg, err := envcfg.GetEnvSpecificConfig(ctx, installer.PkgType, "default", "", runner)
	if err != nil {
		return err
	}
	// resources
	baseResource := recipe.BuildBasePlan(installer.PkgType)

	dockerConfigResource, err := buildDockerConfigResource(configDir)
	if err != nil {
		return err
	}

	criResource := recipe.BuildCRIPlan(
		ctx,
		&existinginfrav1.ContainerRuntime{
			Kind:    "docker",
			Package: "docker-ce",
			Version: "19.03.8",
		},
		cfg,
		installer.PkgType)

	var ips []string
	if conf.Track == "wks-footloose" {
		ips, err = getPrivateIPsFromMachines(configDir)
		// Only the masters for the load balancer
		ips = ips[0:conf.WKSConfig.FootlooseConfig.ControlPlaneNodes]
		if err != nil {
			return err
		}
	} else if conf.Track == "wks-ssh" {
		ips = []string{}
		for _, m := range conf.WKSConfig.SSHConfig.Machines {
			if m.Role == "master" {
				ips = append(ips, m.PrivateAddress)
			}
		}
	} else {
		return errors.Errorf("Unsupported track for configuring LB: %s", conf.Track)
	}

	haConfigResource := &resource.File{
		Content:     generateHAConfiguration(ips),
		Destination: "/tmp/haproxy.cfg",
	}

	haproxyResource := &resource.Run{
		Script:     object.String("mkdir /tmp/haproxy && docker run --detach --name haproxy -v /tmp/haproxy.cfg:/usr/local/etc/haproxy/haproxy.cfg -v /tmp/haproxy:/var/lib/haproxy -p 6443:6443 haproxy"),
		UndoScript: object.String("rm -rv /tmp/haproxy && docker stop haproxy || true && docker rm haproxy || true"),
	}
	lbPlanBuilder := plan.NewBuilder()
	lbPlanBuilder.AddResource("install:base", baseResource)
	lbPlanBuilder.AddResource("install:docker-repo-config", dockerConfigResource,
		plan.DependOn("install:base"))
	lbPlanBuilder.AddResource("install:cri", criResource, plan.DependOn("install:docker-repo-config"))
	lbPlanBuilder.AddResource("install:ha-config", haConfigResource, plan.DependOn("install:cri"))
	lbPlanBuilder.AddResource("install:haproxy", haproxyResource, plan.DependOn("install:ha-config"))

	lbPlan, err := lbPlanBuilder.Plan()
	if err != nil {
		return err
	}

	err = lbPlan.Undo(ctx, runner, plan.EmptyState)
	if err != nil {
		log.Infof("Pre-plan cleanup failed:\n%s\n", err)
		return err
	}
	_, err = lbPlan.Apply(ctx, runner, plan.EmptyDiff())
	if err != nil {
		log.Errorf("Apply of Plan failed:\n%s\n", err)
		return err
	}
	return nil
}

func GetDockerIOPasswordFromConfig(conf *WKPConfig) (string, error) {
	return GetDockerIOPassword(conf.DockerIOPasswordFile)
}

func GetDockerIOPassword(dockerIOPasswordFile string) (string, error) {
	dockerPasswordBytes, err := ioutil.ReadFile(dockerIOPasswordFile)
	if err != nil {
		return "", errors.Wrapf(err, "read docker password file %q", dockerIOPasswordFile)
	}
	dockerPassword := strings.TrimSuffix(string(dockerPasswordBytes), "\n")
	return dockerPassword, nil
}
