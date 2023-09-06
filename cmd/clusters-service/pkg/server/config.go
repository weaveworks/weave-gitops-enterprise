package server

import (
	"context"

	"github.com/spf13/viper"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/gitauth/server/gitproviders"
	"k8s.io/apimachinery/pkg/types"
)

func toHelmRepositoryRef(helmRepo types.NamespacedName) *capiv1_proto.HelmRepositoryRef {
	return &capiv1_proto.HelmRepositoryRef{
		Name:      helmRepo.Name,
		Namespace: helmRepo.Namespace,
	}
}

func (s *server) GetConfig(ctx context.Context, msg *capiv1_proto.GetConfigRequest) (*capiv1_proto.GetConfigResponse, error) {

	repositoryURL := viper.GetString("capi-templates-repository-url")
	mngtClusterName := viper.GetString("cluster-name")
	gitHostTypes := gitproviders.GitHostTypes(gitproviders.ViperGetStringMapString("git-host-types"))
	defaultHelmRepository := s.profileHelmRepository

	return &capiv1_proto.GetConfigResponse{
		RepositoryUrl:         repositoryURL,
		UiConfig:              s.uiConfig,
		ManagementClusterName: mngtClusterName,
		GitHostTypes:          gitHostTypes,
		DefaultHelmRepository: toHelmRepositoryRef(defaultHelmRepository),
	}, nil
}
