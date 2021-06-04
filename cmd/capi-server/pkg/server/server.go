package server

import (
	"context"
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
		params := []*capiv1.Param{}
		for _, p := range t.Spec.Params {
			params = append(params, &capiv1.Param{
				Name:        p.Name,
				Description: p.Description,
			})
		}
		templates = append(templates, &capiv1.Template{
			Name:        t.GetName(),
			Description: t.Spec.Description,
			Params:      params,
		})
	}

	return &capiv1.ListTemplatesResponse{Templates: templates, Total: int32(len(tl))}, err
}
