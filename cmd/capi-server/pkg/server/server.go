package server

import (
	"bytes"
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/capi/flavours"
	capiv1 "github.com/weaveworks/wks/cmd/capi-server/pkg/protos"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/templates"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/utils"
)

type server struct {
	library templates.Library
	capiv1.UnimplementedClustersServiceServer
}

func NewClusterServer(library templates.Library) capiv1.ClustersServiceServer {
	return &server{library: library}
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
	log.Infof("message with params: %v", msg.Values)
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
