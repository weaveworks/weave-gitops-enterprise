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
	"gorm.io/gorm"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/capi"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/credentials"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/git"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/models"
	common_utils "github.com/weaveworks/weave-gitops-enterprise/common/database/utils"
)

func (s *server) CreateTfControllerPullRequest(ctx context.Context, msg *capiv1_proto.CreateTfControllerPullRequestRequest) (*capiv1_proto.CreateTfControllerPullRequestResponse, error) {
	gp, err := getGitProvider(ctx)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.Unauthenticated, "error creating pull request: %s", err.Error())
	}

	if err := validateCreateTfControllerPR(msg); err != nil {
		s.log.Error(err, "Failed to create pull request, message payload was invalid")
		return nil, err
	}

	tmpl, err := s.templatesLibrary.GetTFControllerTemplate(ctx, msg.TemplateName)
	if err != nil {
		return nil, fmt.Errorf("unable to get template %q: %w", msg.TemplateName, err)
	}

	tmplWithValues, err := renderTFControllerTemplateWithValues(tmpl, msg.TemplateName, msg.ParameterValues)
	if err != nil {
		return nil, fmt.Errorf("failed to render template with parameter values: %w", err)
	}

	err = capi.ValidateRenderedTemplates(tmplWithValues)
	if err != nil {
		return nil, fmt.Errorf("validation error rendering template %v, %v", msg.TemplateName, err)
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
	templateName, ok := msg.ParameterValues["TEMPLATE_NAME"]
	if !ok {
		return nil, fmt.Errorf("unable to find 'TEMPLATE_NAME' parameter in supplied values")
	}
	// FIXME: parse and read from Cluster in yaml template
	templateNamespace, ok := msg.ParameterValues["NAMESPACE"]
	if !ok {
		s.log.Info("Couldn't find NAMESPACE param in request, using 'default'.")
		// TODO: https://weaveworks.atlassian.net/browse/WKP-2205
		templateNamespace = "default"
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
		msg.HeadBranch = getHash(msg.RepositoryUrl, msg.ParameterValues["TEMPLATE_NAME"], msg.BaseBranch)
	}
	if msg.Title == "" {
		msg.Title = fmt.Sprintf("Gitops add cluster %s", msg.ParameterValues["TEMPLATE_NAME"])
	}
	if msg.Description == "" {
		msg.Description = fmt.Sprintf("Pull request to create cluster %s", msg.ParameterValues["TEMPLATE_NAME"])
	}
	if msg.CommitMessage == "" {
		msg.CommitMessage = "Add Template Manifests"
	}
	_, err = s.provider.GetRepository(ctx, *gp, repositoryURL)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.Unauthenticated, "failed to access repo %s: %s", repositoryURL, err)
	}

	var pullRequestURL string
	err = s.db.Transaction(func(tx *gorm.DB) error {
		t, err := common_utils.Generate()
		if err != nil {
			return fmt.Errorf("error generating token for new cluster: %v", err)
		}

		c := &models.TFController{
			Name:                  templateName,
			TFControllerName:      templateName,
			TFControllerNamespace: templateNamespace,
			Token:                 t,
		}
		if err := tx.Create(c).Error; err != nil {
			return err
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
			return err
		}

		// Create the PR, this shouldn't fail, but if it does it will rollback the Cluster but not the delete the PR
		pullRequestURL = res.WebURL
		pr := &models.PullRequest{
			URL:  pullRequestURL,
			Type: "create",
		}
		if err := tx.Create(pr).Error; err != nil {
			return err
		}

		c.PullRequests = append(c.PullRequests, pr)
		if err := tx.Save(c).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("unable to create pull request and cluster rows for %q: %w", msg.TemplateName, err)
	}

	return &capiv1_proto.CreateTfControllerPullRequestResponse{
		WebUrl: pullRequestURL,
	}, nil
}

func validateCreateTfControllerPR(msg *capiv1_proto.CreateTfControllerPullRequestRequest) error {
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
