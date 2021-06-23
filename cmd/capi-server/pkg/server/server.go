package server

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/fluxcd/go-git-providers/gitprovider"
	log "github.com/sirupsen/logrus"

	"github.com/mkmik/multierror"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/capi/flavours"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/git"
	capiv1 "github.com/weaveworks/wks/cmd/capi-server/pkg/protos"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/templates"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/utils"
)

type server struct {
	library  templates.Library
	provider git.Provider
	capiv1.UnimplementedClustersServiceServer
}

func NewClusterServer(library templates.Library, provider git.Provider) capiv1.ClustersServiceServer {
	return &server{library: library, provider: provider}
}

func (s *server) ListTemplates(ctx context.Context, msg *capiv1.ListTemplatesRequest) (*capiv1.ListTemplatesResponse, error) {
	tl, err := s.library.List(ctx)
	templates := []*capiv1.Template{}

	// FIXME: probably a clever way to do this conversion / align types
	for _, t := range tl {
		params := []*capiv1.Parameter{}
		for _, p := range t.Spec.Params {
			params = append(params, &capiv1.Parameter{
				Name:        p.Name,
				Description: p.Description,
			})
		}

		var responseBody []byte
		for _, rt := range t.Spec.ResourceTemplates {
			encodedResourceTemplate, err := utils.B64ResourceTemplate(rt)
			if err != nil {
				return nil, err
			}
			responseBody = append(responseBody, encodedResourceTemplate...)
		}

		templates = append(templates, &capiv1.Template{
			Name:        t.GetName(),
			Description: t.Spec.Description,
			Parameters:  params,
			Body:        string(responseBody),
		})
	}

	sort.Slice(templates, func(i, j int) bool { return templates[i].Name < templates[j].Name })
	return &capiv1.ListTemplatesResponse{Templates: templates, Total: int32(len(tl))}, err
}

func (s *server) ListTemplateParams(ctx context.Context, msg *capiv1.ListTemplateParamsRequest) (*capiv1.ListTemplateParamsResponse, error) {
	tm, err := s.library.Get(ctx, msg.TemplateName)
	if err != nil {
		return nil, fmt.Errorf("error looking up template %v: %v", msg.TemplateName, err)
	}
	templateParams, err := flavours.ParamsFromSpec(tm.Spec)
	if err != nil {
		return nil, fmt.Errorf("error looking up template params for %v, %v", msg.TemplateName, err)
	}

	params := []*capiv1.Parameter{}
	for _, p := range templateParams {
		params = append(params, &capiv1.Parameter{
			Name:        p.Name,
			Description: p.Description,
			// TODO: add other param properties to the protobuf and here.
		})
	}

	return &capiv1.ListTemplateParamsResponse{Parameters: params}, err
}

func (s *server) RenderTemplate(ctx context.Context, msg *capiv1.RenderTemplateRequest) (*capiv1.RenderTemplateResponse, error) {
	log.WithFields(log.Fields{
		"request_values": msg.Values,
	}).Info("Received message")
	tm, err := s.library.Get(ctx, msg.TemplateName)
	if err != nil {
		return nil, fmt.Errorf("error looking up template %v: %v", msg.TemplateName, err)
	}
	templateBits, err := flavours.Render(tm.Spec, msg.Values.Values)
	if err != nil {
		return nil, fmt.Errorf("error rendering template %v, %v", msg.TemplateName, err)
	}

	result := bytes.Join(templateBits, []byte("\n---\n"))
	resultStr := string(result[:])

	return &capiv1.RenderTemplateResponse{RenderedTemplate: resultStr}, err
}

func (s *server) CreatePullRequest(ctx context.Context, msg *capiv1.CreatePullRequestRequest) (*capiv1.CreatePullRequestResponse, error) {
	if err := validate(msg); err != nil {
		log.WithError(err).Errorf("Failed to create pull request, message payload was invalid")
		return nil, err
	}

	tmpl, err := s.library.Get(ctx, msg.TemplateName)
	if err != nil {
		return nil, fmt.Errorf("unable to get template %q: %w", msg.TemplateName, err)
	}
	tmplWithValues, err := flavours.Render(tmpl.Spec, msg.ParameterValues)
	if err != nil {
		return nil, fmt.Errorf("unable to render template %q: %w", msg.TemplateName, err)
	}
	result := bytes.Join(tmplWithValues, []byte("\n---\n"))

	clusterName, ok := msg.ParameterValues["CLUSTER_NAME"]
	if !ok {
		return nil, fmt.Errorf("unable to find 'CLUSTER_NAME' parameter in supplied values")
	}
	path := fmt.Sprintf("management/%s.yaml", clusterName)
	content := string(result[:])

	res, err := s.provider.WriteFilesToBranchAndCreatePullRequest(ctx, git.WriteFilesToBranchAndCreatePullRequestRequest{
		GitProvider: git.GitProvider{
			Type:     os.Getenv("GIT_PROVIDER_TYPE"),
			Token:    os.Getenv("GIT_PROVIDER_TOKEN"),
			Hostname: os.Getenv("GIT_PROVIDER_HOSTNAME"),
		},
		RepositoryURL: msg.RepositoryUrl,
		HeadBranch:    msg.HeadBranch,
		BaseBranch:    msg.BaseBranch,
		Title:         msg.Title,
		Description:   msg.Description,
		CommitMessage: msg.CommitMessage,
		Files: []gitprovider.CommitFile{
			gitprovider.CommitFile{
				Path:    &path,
				Content: &content,
			},
		},
	})
	if err != nil {
		log.WithError(err).Errorf("Failed to create pull request")
		return nil, err
	}
	return &capiv1.CreatePullRequestResponse{
		WebUrl: res.WebURL,
	}, nil
}

func validate(msg *capiv1.CreatePullRequestRequest) error {
	var err error

	if msg.TemplateName == "" {
		err = multierror.Append(err, fmt.Errorf("template name must be specified"))
	}

	if msg.ParameterValues == nil {
		err = multierror.Append(err, fmt.Errorf("parameter values must be specified"))
	}

	if msg.RepositoryUrl == "" {
		err = multierror.Append(err, fmt.Errorf("repository url must be specified"))
	}

	if msg.HeadBranch == "" {
		err = multierror.Append(err, fmt.Errorf("head branch must be specified"))
	}

	if msg.BaseBranch == "" {
		err = multierror.Append(err, fmt.Errorf("base branch must be specified"))
	}

	if msg.Title == "" {
		err = multierror.Append(err, fmt.Errorf("title must be specified"))
	}

	if msg.Description == "" {
		err = multierror.Append(err, fmt.Errorf("description must be specified"))
	}

	if msg.CommitMessage == "" {
		err = multierror.Append(err, fmt.Errorf("commit message must be specified"))
	}

	return err
}
