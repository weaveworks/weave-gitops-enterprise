package server

import (
	"context"
	"os"

	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
)

func (s *server) GetConfig(ctx context.Context, msg *capiv1_proto.GetConfigRequest) (*capiv1_proto.GetConfigResponse, error) {

	repositoryURL := os.Getenv("CAPI_TEMPLATES_REPOSITORY_URL")

	return &capiv1_proto.GetConfigResponse{RepositoryURL: repositoryURL}, nil
}
