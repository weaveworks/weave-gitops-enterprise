package server

import (
	"context"
	"fmt"

	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
)

func (s *server) CanaryHooks(ctx context.Context, msg *capiv1_proto.CanaryHooksRequest) (*capiv1_proto.CanaryHooksResponse, error) {
	fmt.Println("foooooooo---------------------------------o")

	return &capiv1_proto.CanaryHooksResponse{}, nil
}

func (s *server) CanaryGates(ctx context.Context, msg *capiv1_proto.CanaryGatesRequest) (*capiv1_proto.CanaryGatesResponse, error) {
	return &capiv1_proto.CanaryGatesResponse{}, nil
}
