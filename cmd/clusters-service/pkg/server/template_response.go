package server

import (
	"fmt"

	capiv1 "github.com/weaveworks/templates-controller/apis/capi/v1alpha2"
	apitemplates "github.com/weaveworks/templates-controller/apis/core"
	gapiv1 "github.com/weaveworks/templates-controller/apis/gitops/v1alpha2"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
	"k8s.io/apimachinery/pkg/types"
)

const TemplateTypeLabel = "weave.works/template-type"

func ToTemplateResponse(t apitemplates.Template, defaultRepo types.NamespacedName) *capiv1_proto.Template {
	var annotation string
	templateKind := t.GetObjectKind().GroupVersionKind().Kind
	switch templateKind {
	case capiv1.Kind:
		annotation = templates.CAPIDisplayNameAnnotation
	case gapiv1.Kind:
		annotation = templates.GitOpsTemplateNameAnnotation
	}

	templateType := t.GetLabels()[TemplateTypeLabel]

	res := &capiv1_proto.Template{
		Name:         t.GetName(),
		Description:  t.GetSpec().Description,
		Provider:     getProvider(t, annotation),
		Annotations:  t.GetAnnotations(),
		Labels:       t.GetLabels(),
		TemplateKind: templateKind,
		TemplateType: templateType,
		Namespace:    t.GetNamespace(),
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

	res.Profiles, err = templates.GetProfilesFromTemplate(t, defaultRepo)
	if err != nil {
		res.Error = fmt.Sprintf("Couldn't load profiles from template: %s", err.Error())
		return res
	}

	return res
}
