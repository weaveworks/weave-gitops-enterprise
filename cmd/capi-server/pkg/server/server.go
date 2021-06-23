package server

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/mkmik/multierror"
	log "github.com/sirupsen/logrus"
	capiv1 "github.com/weaveworks/wks/cmd/capi-server/api/v1alpha1"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/capi"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/git"
	capiv1_proto "github.com/weaveworks/wks/cmd/capi-server/pkg/protos"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/templates"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/utils"
)

type server struct {
	library  templates.Library
	provider git.Provider
	capiv1_proto.UnimplementedClustersServiceServer
}

func NewClusterServer(library templates.Library, provider git.Provider) capiv1_proto.ClustersServiceServer {
	return &server{library: library, provider: provider}
}

func (s *server) ListTemplates(ctx context.Context, msg *capiv1_proto.ListTemplatesRequest) (*capiv1_proto.ListTemplatesResponse, error) {
	tl, err := s.library.List(ctx)
	templates := []*capiv1_proto.Template{}

	for _, t := range tl {
		templateWithMeta, err := ToTemplateResponse(t)
		if err != nil {
			return nil, err
		}
		templates = append(templates, templateWithMeta)
	}

	sort.Slice(templates, func(i, j int) bool { return templates[i].Name < templates[j].Name })
	return &capiv1_proto.ListTemplatesResponse{Templates: templates, Total: int32(len(tl))}, err
}

func (s *server) ListTemplateParams(ctx context.Context, msg *capiv1_proto.ListTemplateParamsRequest) (*capiv1_proto.ListTemplateParamsResponse, error) {
	tm, err := s.library.Get(ctx, msg.TemplateName)
	if err != nil {
		return nil, fmt.Errorf("error looking up template %v: %v", msg.TemplateName, err)
	}
	t, err := ToTemplateResponse(tm)
	if err != nil {
		return nil, fmt.Errorf("error looking up template params for %v, %v", msg.TemplateName, err)
	}

	return &capiv1_proto.ListTemplateParamsResponse{Parameters: t.Parameters, Objects: t.Objects}, err
}

func (s *server) RenderTemplate(ctx context.Context, msg *capiv1_proto.RenderTemplateRequest) (*capiv1_proto.RenderTemplateResponse, error) {
	log.WithFields(log.Fields{
		"request_values": msg.Values,
	}).Info("Received message")
	tm, err := s.library.Get(ctx, msg.TemplateName)
	if err != nil {
		return nil, fmt.Errorf("error looking up template %v: %v", msg.TemplateName, err)
	}
	templateBits, err := capi.Render(tm.Spec, msg.Values.Values)
	if err != nil {
		return nil, fmt.Errorf("error rendering template %v, %v", msg.TemplateName, err)
	}

	result := bytes.Join(templateBits, []byte("\n---\n"))
	resultStr := string(result[:])

	return &capiv1_proto.RenderTemplateResponse{RenderedTemplate: resultStr}, err
}

func ToTemplateResponse(t *capiv1.CAPITemplate) (*capiv1_proto.Template, error) {
	// FIXME: probably a clever way to do this conversion / align types

	meta, err := capi.ParseTemplateMeta(t)
	if err != nil {
		return nil, err
	}

	params := []*capiv1_proto.Parameter{}
	for _, p := range meta.Params {
		params = append(params, &capiv1_proto.Parameter{
			Name:        p.Name,
			Description: p.Description,
			Options:     p.Options,
			Required:    p.Required,
		})
	}

	objects := []*capiv1_proto.TemplateObject{}
	for _, o := range meta.Objects {
		objects = append(objects, &capiv1_proto.TemplateObject{
			Kind:       o.Kind,
			ApiVersion: o.APIVersion,
			Parameters: o.Params,
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

	return &capiv1_proto.Template{
		Name:        t.GetName(),
		Description: t.Spec.Description,
		Parameters:  params,
		Body:        string(responseBody),
		Objects:     objects,
	}, nil
}

func (s *server) CreatePullRequest(ctx context.Context, msg *capiv1_proto.CreatePullRequestRequest) (*capiv1_proto.CreatePullRequestResponse, error) {
	if err := validate(msg); err != nil {
		log.WithError(err).Errorf("Failed to create pull request, message payload was invalid")
		return nil, err
	}

	tmpl, err := s.library.Get(ctx, msg.TemplateName)
	if err != nil {
		return nil, fmt.Errorf("unable to get template %q: %w", msg.TemplateName, err)
	}
	tmplWithValues, err := capi.Render(tmpl.Spec, msg.ParameterValues)
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

	repositoryURL := os.Getenv("CAPI_TEMPLATES_REPOSITORY_URL")
	if msg.RepositoryUrl != "" {
		repositoryURL = msg.RepositoryUrl
	}
	baseBranch := os.Getenv("CAPI_TEMPLATES_REPOSITORY_BASE_BRANCH")
	if msg.BaseBranch != "" {
		baseBranch = msg.BaseBranch
	}

	res, err := s.provider.WriteFilesToBranchAndCreatePullRequest(ctx, git.WriteFilesToBranchAndCreatePullRequestRequest{
		GitProvider: git.GitProvider{
			Type:     os.Getenv("GIT_PROVIDER_TYPE"),
			Token:    os.Getenv("GIT_PROVIDER_TOKEN"),
			Hostname: os.Getenv("GIT_PROVIDER_HOSTNAME"),
		},
		RepositoryURL: repositoryURL,
		HeadBranch:    msg.HeadBranch,
		BaseBranch:    baseBranch,
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
	return &capiv1_proto.CreatePullRequestResponse{
		WebUrl: res.WebURL,
	}, nil
}

func validate(msg *capiv1_proto.CreatePullRequestRequest) error {
	var err error

	if msg.TemplateName == "" {
		err = multierror.Append(err, fmt.Errorf("template name must be specified"))
	}

	if msg.ParameterValues == nil {
		err = multierror.Append(err, fmt.Errorf("parameter values must be specified"))
	}

	if msg.HeadBranch == "" {
		err = multierror.Append(err, fmt.Errorf("head branch must be specified"))
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
