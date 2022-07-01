package server

import (
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"k8s.io/client-go/discovery"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/clusters"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
)

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
	log              logr.Logger
	templatesLibrary templates.Library
	clustersLibrary  clusters.Library
	clientsFactory   clustersmngr.ClientsFactory
	provider         git.Provider
	clientGetter     kube.ClientGetter
	discoveryClient  discovery.DiscoveryInterface
	capiv1_proto.UnimplementedClustersServiceServer
	ns                        string // The namespace where cluster objects reside
	profileHelmRepositoryName string
	helmRepositoryCacheDir    string
}

type ServerOpts struct {
	Logger                    logr.Logger
	TemplatesLibrary          templates.Library
	ClustersLibrary           clusters.Library
	ClientsFactory            clustersmngr.ClientsFactory
	GitProvider               git.Provider
	ClientGetter              kube.ClientGetter
	DiscoveryClient           discovery.DiscoveryInterface
	ClustersNamespace         string
	ProfileHelmRepositoryName string
	HelmRepositoryCacheDir    string
}

func NewClusterServer(opts ServerOpts) capiv1_proto.ClustersServiceServer {
	return &server{
		log:                       opts.Logger,
		clustersLibrary:           opts.ClustersLibrary,
		templatesLibrary:          opts.TemplatesLibrary,
		clientsFactory:            opts.ClientsFactory,
		provider:                  opts.GitProvider,
		clientGetter:              opts.ClientGetter,
		discoveryClient:           opts.DiscoveryClient,
		ns:                        opts.ClustersNamespace,
		profileHelmRepositoryName: opts.ProfileHelmRepositoryName,
		helmRepositoryCacheDir:    opts.HelmRepositoryCacheDir,
	}
}
