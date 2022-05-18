package server

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/mkmik/multierror"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"

	gapiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/gitopstemplate/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/credentials"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
)

func (s *server) CreateTfControllerPullRequest(ctx context.Context, msg *proto.CreateTfControllerPullRequestRequest) (*proto.CreateTfControllerPullRequestResponse, error) {
	gp, err := getGitProvider(ctx)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.Unauthenticated, "error creating pull request: %s", err.Error())
	}

	if err := validateCreateTfControllerPR(msg); err != nil {
		s.log.Error(err, "Failed to create pull request, message payload was invalid")
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "validation error: %s", err.Error())
	}

	tmpl, err := s.templatesLibrary.Get(ctx, msg.TemplateName, gapiv1.Kind)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.Internal, "unable to get template %q: %s", msg.TemplateName, err)
	}

	tmplWithValues, err := renderTemplateWithValues(tmpl, msg.TemplateName, viper.GetString("capi-templates-namespace"), msg.ParameterValues)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.Internal, "failed to render template with parameter values: %s", err)
	}

	err = templates.ValidateRenderedTemplates(tmplWithValues)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.Internal, "validation error rendering template %v, %s", msg.TemplateName, err)
	}

	client, err := s.clientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	tmplWithValuesAndCredentials, err := credentials.CheckAndInjectCredentials(s.log, client, tmplWithValues, nil, msg.TemplateName)
	if err != nil {
		return nil, err
	}

	// FIXME: parse and read from Cluster in yaml template
	templateName, ok := msg.ParameterValues["RESOURCE_NAME"]
	if !ok {
		return nil, grpcStatus.Errorf(codes.Internal, "unable to find 'RESOURCE_NAME' parameter in supplied values")
	}

	path := getTfControllerManifestPath(templateName)
	content := string(tmplWithValuesAndCredentials[:])
	files := []gitprovider.CommitFile{
		{
			Path:    &path,
			Content: &content,
		},
	}

	repositoryURL := viper.GetString("tfcontroller-templates-repository-url")
	if msg.RepositoryUrl != "" {
		repositoryURL = msg.RepositoryUrl
	}
	baseBranch := viper.GetString("tfcontroller-templates-repository-base-branch")
	if msg.BaseBranch != "" {
		baseBranch = msg.BaseBranch
	}
	if msg.HeadBranch == "" {
		msg.HeadBranch = getHash(msg.RepositoryUrl, msg.ParameterValues["RESOURCE_NAME"], msg.BaseBranch)
	}
	if msg.Title == "" {
		msg.Title = fmt.Sprintf("Gitops add cluster %s", msg.ParameterValues["RESOURCE_NAME"])
	}
	if msg.Description == "" {
		msg.Description = fmt.Sprintf("Pull request to create cluster %s", msg.ParameterValues["RESOURCE_NAME"])
	}
	if msg.CommitMessage == "" {
		msg.CommitMessage = "Add Template Manifests"
	}
	_, err = s.provider.GetRepository(ctx, *gp, repositoryURL)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.Internal, "failed to access repo %s: %s", repositoryURL, err)
	}

	// FIXME: maybe this should reconcile rather than just try to create in case of other errors, e.g. database row creation
	res, err := s.provider.WriteFilesToBranchAndCreatePullRequest(ctx, git.WriteFilesToBranchAndCreatePullRequestRequest{
		GitProvider:       *gp,
		RepositoryURL:     repositoryURL,
		ReposistoryAPIURL: msg.RepositoryApiUrl,
		HeadBranch:        msg.HeadBranch,
		BaseBranch:        baseBranch,
		Title:             msg.Title,
		Description:       msg.Description,
		CommitMessage:     msg.CommitMessage,
		Files:             files,
	})
	if err != nil {
		s.log.Error(err, "Failed to create pull request")
		return nil, grpcStatus.Errorf(codes.Internal, "failed to create pull request %s", err)
	}

	return &proto.CreateTfControllerPullRequestResponse{
		WebUrl: res.WebURL,
	}, nil
}

func validateCreateTfControllerPR(msg *proto.CreateTfControllerPullRequestRequest) error {
	var err error

	if msg.TemplateName == "" {
		err = multierror.Append(err, fmt.Errorf("template name must be specified"))
	}

	if msg.ParameterValues == nil {
		err = multierror.Append(err, fmt.Errorf("parameter values must be specified"))
	}

	return err
}

func getTfControllerManifestPath(templateName string) string {
	return filepath.Join(
		viper.GetString("tfcontroller-repository-path"),
		fmt.Sprintf("%s.yaml", templateName),
	)
}
