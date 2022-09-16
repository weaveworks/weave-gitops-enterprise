package server

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/clusters"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
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

// TODO: Remove this when syncing with the database!
// ObjectReference points to a resource.
type ObjectReference struct {
	Kind      string
	Name      string
	Namespace string
}

type Chart struct {
	Name    string
	Version string
}

type chartsCache interface {
	ListChartsByRepositoryAndCluster(ctx context.Context, repoRef ObjectReference, clusterRef types.NamespacedName) ([]Chart, error)
	GetChartValues(ctx context.Context, repoRef ObjectReference, clusterRef types.NamespacedName, chart Chart) ([]byte, error)
}

type server struct {
	log              logr.Logger
	templatesLibrary templates.Library
	clustersLibrary  clusters.Library
	clustersManager  clustersmngr.ClustersManager
	provider         git.Provider
	clientGetter     kube.ClientGetter
	discoveryClient  discovery.DiscoveryInterface
	capiv1_proto.UnimplementedClustersServiceServer
	ns                        string // The namespace where cluster objects reside
	profileHelmRepositoryName string
	helmRepositoryCacheDir    string
	capiEnabled               bool
	chartsCache               chartsCache
}

type ServerOpts struct {
	Logger                    logr.Logger
	TemplatesLibrary          templates.Library
	ClustersLibrary           clusters.Library
	ClustersManager           clustersmngr.ClustersManager
	GitProvider               git.Provider
	ClientGetter              kube.ClientGetter
	DiscoveryClient           discovery.DiscoveryInterface
	ClustersNamespace         string
	ProfileHelmRepositoryName string
	HelmRepositoryCacheDir    string
	CAPIEnabled               bool
	ChartsCache               chartsCache
}

func NewClusterServer(opts ServerOpts) capiv1_proto.ClustersServiceServer {
	return &server{
		log:                       opts.Logger,
		clustersLibrary:           opts.ClustersLibrary,
		templatesLibrary:          opts.TemplatesLibrary,
		clustersManager:           opts.ClustersManager,
		provider:                  opts.GitProvider,
		clientGetter:              opts.ClientGetter,
		discoveryClient:           opts.DiscoveryClient,
		ns:                        opts.ClustersNamespace,
		profileHelmRepositoryName: opts.ProfileHelmRepositoryName,
		helmRepositoryCacheDir:    opts.HelmRepositoryCacheDir,
		capiEnabled:               opts.CAPIEnabled,
		chartsCache:               opts.ChartsCache,
	}
}
