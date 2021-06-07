package templates

import (
	"context"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/capi-templates/flavours"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// TemplateParams is a map of parameter to value for rendering templates.
type TemplateParams map[string]string

// TemplateGetter implementations get templates by name.
type TemplateGetter interface {
	Get(ctx context.Context, name string) ([]byte, error)
}

// TemplateLister implementations list templates from a Library.
type TemplateLister interface {
	List(ctx context.Context) (map[string][]byte, error)
}

// Library represents a library of Templates indexed by name.
type Library interface {
	TemplateGetter
	TemplateLister
}

func LoadTemplatesFromConfigmap(ctx context.Context, clientset kubernetes.Clientset, namespace string, name string) (map[string]*flavours.CAPITemplate, error) {
	log.Debugf("querying kubernetes for configmap: %s/%s\n", namespace, name)
	templateConfigMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return nil, fmt.Errorf("configmap %s not found in %s namespace\n", name, namespace)
	} else if err != nil {
		return nil, fmt.Errorf("error getting configmap: %s\n", err)
	}
	log.Debugf("got template configmap: %v\n", templateConfigMap)

	tm, err := flavours.ParseConfigMap(*templateConfigMap)
	if errors.IsNotFound(err) {
		return nil, fmt.Errorf("error parsing CAPI templates from configmap: %s\n", err)
	}
	return tm, nil
}

func GetTemplate(ctx context.Context, clientset kubernetes.Clientset, namespace, cmName, tmName string) (*flavours.CAPITemplate, error) {
	tm := &flavours.CAPITemplate{}
	tl, err := LoadTemplatesFromConfigmap(ctx, clientset, namespace, cmName)
	if err != nil {
		return nil, err
	}

	for _, template := range tl {
		if tmName == template.ObjectMeta.Name {
			tm = template
		}
	}
	return tm, nil
}

func GetTemplateParams(ctx context.Context, clientset kubernetes.Clientset, namespace, cmName, tmName string) ([]flavours.Param, error) {
	tm, err := GetTemplate(ctx, clientset, namespace, cmName, tmName)
	if err != nil {
		return nil, err
	}
	return flavours.ParamsFromSpec(tm.Spec)
}

// TODO: move this into a grpc handler
func RenderTemplate(ctx context.Context, clientset kubernetes.Clientset, namespace, cmName, tmName string, vars map[string]string) ([][]byte, error) {
	tm, err := GetTemplate(ctx, clientset, os.Getenv("POD_NAMESPACE"), os.Getenv("TEMPLATE_CONFIGMAP_NAME"), tmName)
	if err != nil {
		return nil, err
	}

	render, err := flavours.Render(tm.Spec, vars)
	if err != nil {
		return nil, err
	}
	return render, nil
}
