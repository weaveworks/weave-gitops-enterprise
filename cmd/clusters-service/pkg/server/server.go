package server

import (
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"gorm.io/gorm"
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
	provider         git.Provider
	clientGetter     kube.ClientGetter
	discoveryClient  discovery.DiscoveryInterface
	capiv1_proto.UnimplementedClustersServiceServer
	db                        *gorm.DB
	ns                        string // The namespace where cluster objects reside
	profileHelmRepositoryName string
	helmRepositoryCacheDir    string
}

func NewClusterServer(log logr.Logger, clustersLibrary clusters.Library, templatesLibrary templates.Library, provider git.Provider, clientGetter kube.ClientGetter, discoveryClient discovery.DiscoveryInterface, db *gorm.DB, ns string, profileHelmRepositoryName string, helmRepositoryCacheDir string) capiv1_proto.ClustersServiceServer {
	return &server{
		log:                       log,
		clustersLibrary:           clustersLibrary,
		templatesLibrary:          templatesLibrary,
		provider:                  provider,
		clientGetter:              clientGetter,
		discoveryClient:           discoveryClient,
		db:                        db,
		ns:                        ns,
		profileHelmRepositoryName: profileHelmRepositoryName,
		helmRepositoryCacheDir:    helmRepositoryCacheDir,
	}
}
