package server

import (
	"context"

	gapiv1 "github.com/weaveworks/templates-controller/apis/gitops/v1alpha2"
	proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
)

func (s *server) CreateTfControllerPullRequest(ctx context.Context, msg *proto.CreateTfControllerPullRequestRequest) (*proto.CreateTfControllerPullRequestResponse, error) {
	res, err := s.CreatePullRequest(ctx, &proto.CreatePullRequestRequest{
		RepositoryUrl:    msg.RepositoryUrl,
		HeadBranch:       msg.HeadBranch,
		BaseBranch:       msg.BaseBranch,
		Title:            msg.Title,
		Description:      msg.Description,
		Name:             msg.Name,
		Namespace:        msg.Namespace,
		ParameterValues:  msg.ParameterValues,
		CommitMessage:    msg.CommitMessage,
		RepositoryApiUrl: msg.RepositoryApiUrl,
		TemplateKind:     gapiv1.Kind,
	})

	var tfRes *proto.CreateTfControllerPullRequestResponse
	if res != nil {
		tfRes = &proto.CreateTfControllerPullRequestResponse{
			WebUrl: res.WebUrl,
		}
	}

	return tfRes, err
}
