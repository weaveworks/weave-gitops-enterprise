package server

import (
	"context"

	"github.com/spf13/viper"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
)

func (s *server) GetConfig(ctx context.Context, msg *capiv1_proto.GetConfigRequest) (*capiv1_proto.GetConfigResponse, error) {

	repositoryURL := viper.GetString("capi-templates-repository-url")

	return &capiv1_proto.GetConfigResponse{RepositoryURL: repositoryURL}, nil
}
