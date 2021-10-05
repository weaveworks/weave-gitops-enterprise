package server

import (
	"fmt"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/capi-server/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/capi-server/pkg/capi"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/capi-server/pkg/protos"
)

func ToTemplateResponse(t *capiv1.CAPITemplate) *capiv1_proto.Template {
	res := &capiv1_proto.Template{
		Name:        t.GetName(),
		Description: t.Spec.Description,
		Provider:    getProvider(t),
	}

	meta, err := capi.ParseTemplateMeta(t)
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
		})
	}
	for _, o := range meta.Objects {
		res.Objects = append(res.Objects, &capiv1_proto.TemplateObject{
			Kind:       o.Kind,
			ApiVersion: o.APIVersion,
			Parameters: o.Params,
			Name:       o.Name,
			DisplayName: o.DisplayName,
		})
	}

	return res
}
