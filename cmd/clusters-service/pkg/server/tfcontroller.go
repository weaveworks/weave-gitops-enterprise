package server

import (
	"context"

	gapiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/gitopstemplate/v1alpha1"
	proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
)

func (s *server) CreateTfControllerPullRequest(ctx context.Context, msg *proto.CreateTfControllerPullRequestRequest) (*proto.CreateTfControllerPullRequestResponse, error) {
	res, err := s.CreatePullRequest(ctx, &proto.CreatePullRequestRequest{
		RepositoryUrl:    msg.RepositoryUrl,
		HeadBranch:       msg.HeadBranch,
		BaseBranch:       msg.BaseBranch,
		Title:            msg.Title,
		Description:      msg.Description,
		TemplateName:     msg.TemplateName,
		ParameterValues:  msg.ParameterValues,
		CommitMessage:    msg.CommitMessage,
		RepositoryApiUrl: msg.RepositoryApiUrl,

		TemplateKind: gapiv1.Kind,
	})

	return &proto.CreateTfControllerPullRequestResponse{
		WebUrl: res.WebUrl,
	}, err
}
