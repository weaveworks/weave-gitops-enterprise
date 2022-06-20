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

func NewClusterServer(log logr.Logger, clustersLibrary clusters.Library, templatesLibrary templates.Library, clientsFactory clustersmngr.ClientsFactory, provider git.Provider, clientGetter kube.ClientGetter, discoveryClient discovery.DiscoveryInterface, ns string, profileHelmRepositoryName string, helmRepositoryCacheDir string) capiv1_proto.ClustersServiceServer {
	return &server{
		log:                       log,
		clustersLibrary:           clustersLibrary,
		templatesLibrary:          templatesLibrary,
		clientsFactory:            clientsFactory,
		provider:                  provider,
		clientGetter:              clientGetter,
		discoveryClient:           discoveryClient,
		ns:                        ns,
		profileHelmRepositoryName: profileHelmRepositoryName,
		helmRepositoryCacheDir:    helmRepositoryCacheDir,
	}
}
