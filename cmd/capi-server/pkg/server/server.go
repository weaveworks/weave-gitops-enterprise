package server

import (
	"bytes"
	"context"
	"fmt"
	"os"

	capiv1 "github.com/weaveworks/wks/cmd/capi-server/pkg/protos"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/templates"
	"k8s.io/client-go/kubernetes"
)

type server struct {
	clientset kubernetes.Clientset
	capiv1.UnimplementedClustersServiceServer
}

func NewClusterServer(clientset kubernetes.Clientset) capiv1.ClustersServiceServer {
	return &server{clientset: clientset}
}

func (s *server) ListTemplates(ctx context.Context, msg *capiv1.ListTemplatesRequest) (*capiv1.ListTemplatesResponse, error) {
	tl, err := templates.LoadTemplatesFromConfigmap(ctx, s.clientset, os.Getenv("POD_NAMESPACE"), os.Getenv("TEMPLATE_CONFIGMAP_NAME"))
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
		templates = append(templates, &capiv1.Template{
			Name:        t.GetName(),
			Description: t.Spec.Description,
			Parameters:  params,
		})
	}

	return &capiv1.ListTemplatesResponse{Templates: templates, Total: int32(len(tl))}, err
}

func (s *server) ListTemplateParams(ctx context.Context, msg *capiv1.ListTemplateParamsRequest) (*capiv1.ListTemplateParamsResponse, error) {
	templateParams, err := templates.GetTemplateParams(ctx, s.clientset, os.Getenv("POD_NAMESPACE"), os.Getenv("TEMPLATE_CONFIGMAP_NAME"), msg.TemplateName)
	if err != nil {
		return nil, fmt.Errorf("error looking up template params for %v", msg.TemplateName)
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
	tm, err := templates.RenderTemplate(ctx, s.clientset, os.Getenv("POD_NAMESPACE"), os.Getenv("TEMPLATE_CONFIGMAP_NAME"), msg.TemplateName, msg.Values.Values)
	if err != nil {
		return nil, fmt.Errorf("error rendering template %v", msg.TemplateName)
	}

	result := bytes.Join(tm, []byte("\n---\n"))
	resultStr := string(result[:])

	return &capiv1.RenderTemplateResponse{RenderedTemplate: resultStr}, err
}
