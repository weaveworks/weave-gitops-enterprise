package server

import (
	"fmt"
	"os"
	"strings"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/capi"
)

func renderTemplateWithValues(t *capiv1.CAPITemplate, name string, values map[string]string) ([][]byte, error) {
	opts := []capi.RenderOptFunc{
		capi.InNamespace(os.Getenv("CAPI_CLUSTERS_NAMESPACE")),
	}
	if os.Getenv("INJECT_PRUNE_ANNOTATION") != "disabled" {
		opts = append(opts, capi.InjectPruneAnnotation)
	}

	templateBits, err := capi.Render(t.Spec, values, opts...)
	if err != nil {
		if missing, ok := isMissingVariableError(err); ok {
			return nil, fmt.Errorf("error rendering template %v due to missing variables: %s", name, missing)
		}
		return nil, fmt.Errorf("error rendering template %v, %v", name, err)
	}

	return templateBits, nil
}

func getProvider(t *capiv1.CAPITemplate) string {
	meta, err := capi.ParseTemplateMeta(t)

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
