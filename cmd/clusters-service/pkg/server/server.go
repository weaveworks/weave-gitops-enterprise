package server

import (
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/mgmtfetcher"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
)

const defaultAutomationNamespace = "flux-system"

var providers = map[string]string{
	"AWSCluster":             "aws",
	"AWSManagedCluster":      "aws",
	"AWSManagedControlPlane": "aws",
	"AzureCluster":           "azure",
	"AzureManagedCluster":    "azure",
	"DOCluster":              "digitalocean",
	"DockerCluster":          "docker",
	"GCPCluster":             "gcp",
	"OpenStackCluster":       "openstack",
	"PacketCluster":          "packet",
	"VSphereCluster":         "vsphere",
}

type server struct {
	log             logr.Logger
	clustersManager clustersmngr.ClustersManager
	provider        git.Provider
	clientGetter    kube.ClientGetter
	discoveryClient discovery.DiscoveryInterface
	capiv1_proto.UnimplementedClustersServiceServer
	ns                        string // The namespace where cluster objects reside
	profileHelmRepositoryName string
	helmRepositoryCacheDir    string
	capiEnabled               bool
	cluster                   types.NamespacedName

	restConfig        *rest.Config
	chartJobs         *helm.Jobs
	valuesFetcher     helm.ValuesFetcher
	chartsCache       helm.ChartsCacheReader
	managementFetcher *mgmtfetcher.ManagementCrossNamespacesFetcher
}

type ServerOpts struct {
	Logger                    logr.Logger
	ClustersManager           clustersmngr.ClustersManager
	GitProvider               git.Provider
	ClientGetter              kube.ClientGetter
	DiscoveryClient           discovery.DiscoveryInterface
	ClustersNamespace         string
	ProfileHelmRepositoryName string
	HelmRepositoryCacheDir    string
	CAPIEnabled               bool
	Cluster                   types.NamespacedName

	RestConfig        *rest.Config
	ChartJobs         *helm.Jobs
	ChartsCache       helm.ChartsCacheReader
	ValuesFetcher     helm.ValuesFetcher
	ManagementFetcher *mgmtfetcher.ManagementCrossNamespacesFetcher
}

func NewClusterServer(opts ServerOpts) capiv1_proto.ClustersServiceServer {
	return &server{
		log:                       opts.Logger,
		clustersManager:           opts.ClustersManager,
		provider:                  opts.GitProvider,
		clientGetter:              opts.ClientGetter,
		discoveryClient:           opts.DiscoveryClient,
		ns:                        opts.ClustersNamespace,
		profileHelmRepositoryName: opts.ProfileHelmRepositoryName,
		helmRepositoryCacheDir:    opts.HelmRepositoryCacheDir,
		capiEnabled:               opts.CAPIEnabled,
		restConfig:                opts.RestConfig,
		chartJobs:                 helm.NewJobs(),
		chartsCache:               opts.ChartsCache,
		valuesFetcher:             opts.ValuesFetcher,
		managementFetcher:         opts.ManagementFetcher,
		cluster:                   opts.Cluster,
	}
}
