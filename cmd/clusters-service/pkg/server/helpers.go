package server

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	apitemplates "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/templates"
)

func renderTemplateWithValues(t *apitemplates.Template, name, namespace string, values map[string]string) ([][]byte, error) {
	opts := []templates.RenderOptFunc{
		templates.InNamespace(namespace),
	}
	if viper.GetString("inject-prune-annotation") != "disabled" {
		opts = append(opts, templates.InjectPruneAnnotation)
	}

	processor, err := templates.NewProcessorForTemplate(*t)
	if err != nil {
		return nil, err
	}

	params, err := processor.Params()
	if err != nil {
		return nil, err
	}

	for _, param := range params {
		_, ok := values[param.Name]
		if !ok && values != nil {
			if param.Required {
				return nil, fmt.Errorf("missing required parameter: %s", param.Name)
			}

			values[param.Name] = ""
		}
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

func getProvider(t *apitemplates.Template, annotation string) string {
	meta, err := templates.ParseTemplateMeta(*t, annotation)

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
