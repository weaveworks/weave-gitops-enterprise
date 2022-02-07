package server

import (
	"context"
	"path/filepath"

	"github.com/go-logr/logr"
	wegogit "github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"gorm.io/gorm"
	"k8s.io/client-go/discovery"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/credentials"
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
	log             logr.Logger
	library         templates.Library
	provider        git.Provider
	clientGetter    kube.ClientGetter
	discoveryClient discovery.DiscoveryInterface
	capiv1_proto.UnimplementedClustersServiceServer
	db                        *gorm.DB
	ns                        string // The namespace where cluster objects reside
	profileHelmRepositoryName string
	helmRepositoryCacheDir    string
}

var DefaultRepositoryPath string = filepath.Join(wegogit.WegoRoot, wegogit.WegoAppDir, "capi")

func NewClusterServer(log logr.Logger, library templates.Library, provider git.Provider, clientGetter kube.ClientGetter, discoveryClient discovery.DiscoveryInterface, db *gorm.DB, ns string, profileHelmRepositoryName string, helmRepositoryCacheDir string) capiv1_proto.ClustersServiceServer {
	return &server{
		log:                       log,
		library:                   library,
		provider:                  provider,
		clientGetter:              clientGetter,
		discoveryClient:           discoveryClient,
		db:                        db,
		ns:                        ns,
		profileHelmRepositoryName: profileHelmRepositoryName,
		helmRepositoryCacheDir:    helmRepositoryCacheDir,
	}
}

// ListCredentials searches the management cluster and lists any objects that match specific given types
func (s *server) ListCredentials(ctx context.Context, msg *capiv1_proto.ListCredentialsRequest) (*capiv1_proto.ListCredentialsResponse, error) {
	client, err := s.clientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	creds := []*capiv1_proto.Credential{}
	foundCredentials, err := credentials.FindCredentials(ctx, client, s.discoveryClient)
	if err != nil {
		return nil, err
	}

	for _, identity := range foundCredentials {
		creds = append(creds, &capiv1_proto.Credential{
			Group:     identity.GroupVersionKind().Group,
			Version:   identity.GroupVersionKind().Version,
			Kind:      identity.GetKind(),
			Name:      identity.GetName(),
			Namespace: identity.GetNamespace(),
		})
	}

	return &capiv1_proto.ListCredentialsResponse{Credentials: creds, Total: int32(len(creds))}, nil
}
