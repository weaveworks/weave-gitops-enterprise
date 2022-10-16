package server

import (
	"fmt"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/capi/v1alpha2"
	gapiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/gitopstemplate/v1alpha2"
	apitemplates "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
)

func ToTemplateResponse(t apitemplates.Template) *capiv1_proto.Template {
	var annotation string
	templateKind := t.GetObjectKind().GroupVersionKind().Kind
	switch templateKind {
	case capiv1.Kind:
		annotation = templates.CAPIDisplayNameAnnotation
	case gapiv1.Kind:
		annotation = templates.GitOpsTemplateNameAnnotation
	}
	res := &capiv1_proto.Template{
		Name:         t.GetName(),
		Description:  t.GetSpec().Description,
		Provider:     getProvider(t, annotation),
		Annotations:  t.GetAnnotations(),
		Labels:       t.GetLabels(),
		TemplateKind: templateKind,
	}

	meta, err := templates.ParseTemplateMeta(t, annotation)
	if err != nil {
		res.Error = fmt.Sprintf("Couldn't load template body: %s", err.Error())
		return res
	}

	for _, p := range meta.Params {
		res.Parameters = append(res.Parameters, &capiv1_proto.Parameter{
			Name:        p.Name,
			Description: p.Description,
			Options:     p.Options,
			Required:    p.Required,
			Default:     p.Default,
		})
	}
	for _, o := range meta.Objects {
		res.Objects = append(res.Objects, &capiv1_proto.TemplateObject{
			Kind:        o.Kind,
			ApiVersion:  o.APIVersion,
			Parameters:  o.Params,
			Name:        o.Name,
			DisplayName: o.DisplayName,
		})
	}

	return res
}
