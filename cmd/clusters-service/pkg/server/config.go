package server

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"sigs.k8s.io/yaml"
)

func (s *server) GetConfig(ctx context.Context, msg *capiv1_proto.GetConfigRequest) (*capiv1_proto.GetConfigResponse, error) {

	repositoryURL := viper.GetString("capi-templates-repository-url")
	mngtClusterName := viper.GetString("cluster-name")

	uiConfig, err := yaml.Marshal(s.uiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal UI config: %w", err)
	}

	return &capiv1_proto.GetConfigResponse{
		RepositoryURL:         repositoryURL,
		UiConfig:              string(uiConfig),
		ManagementClusterName: mngtClusterName,
	}, nil
}
