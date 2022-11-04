package server

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/capi/v1alpha1"
	apitemplates "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
)

func renderTemplateWithValues(t apitemplates.Template, name, namespace string, values map[string]string) ([][]byte, error) {
	opts := []templates.RenderOptFunc{
		templates.InNamespace(namespace),
		templates.InjectLabels(map[string]string{
			"templates.weave.works/template-name":      name,
			"templates.weave.works/template-namespace": viper.GetString("capi-templates-namespace"),
		}),
	}

	if shouldInjectPruneAnnotation(t) {
		opts = append(opts, templates.InjectPruneAnnotation)
	}

	processor, err := templates.NewProcessorForTemplate(t)
	if err != nil {
		return nil, err
	}

	templateBits, err := processor.RenderTemplates(values, opts...)
	if err != nil {
		if missing, ok := isMissingVariableError(err); ok {
			return nil, fmt.Errorf("error rendering template %v due to missing variables: %s", name, missing)
		}
		return nil, fmt.Errorf("error rendering template %v, %v", name, err)
	}

	return templateBits, nil
}

func shouldInjectPruneAnnotation(t apitemplates.Template) bool {
	anno := t.GetAnnotations()[templates.InjectPruneAnnotationAnnotation]
	if anno != "" {
		return anno == "true"
	}

	return viper.GetString("inject-prune-annotation") != "disabled" && isCAPITemplate(t)
}

func getProvider(t apitemplates.Template, annotation string) string {
	meta, err := templates.ParseTemplateMeta(t, annotation)

	if err != nil {
		return ""
	}

	for _, obj := range meta.Objects {
		if p, ok := providers[obj.Kind]; ok {
			return p
		}
	}

	return ""
}

func isMissingVariableError(err error) (string, bool) {
	errStr := err.Error()
	prefix := "processing template: value for variables"
	suffix := "is not set. Please set the value using os environment variables or the clusterctl config file"
	if strings.HasPrefix(errStr, prefix) && strings.HasSuffix(errStr, suffix) {
		missing := strings.TrimSpace(errStr[len(prefix):strings.Index(errStr, suffix)])
		return missing, true
	}
	return "", false
}

func getClusterNamespace(clusterNamespace string) string {
	namespace := "default"
	if clusterNamespace == "" {
		ns := viper.GetString("capi-clusters-namespace")
		if ns != "" {
			namespace = ns
		}

	} else {
		namespace = clusterNamespace
	}
	return namespace
}

func isCAPITemplate(t apitemplates.Template) bool {
	return t.GetObjectKind().GroupVersionKind().Kind == capiv1.Kind
}
