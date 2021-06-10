package templates

import (
	"context"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	appv1 "github.com/weaveworks/wks/cmd/capi-server/api/v1"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/capi/flavours"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

func LoadTemplatesFromConfigmap(ctx context.Context, clientset kubernetes.Clientset, namespace string, name string) (map[string]*appv1.CAPITemplate, error) {
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

func GetTemplate(ctx context.Context, crdRestClient *rest.RESTClient, namespace, tmName string) (*appv1.CAPITemplate, error) {
	capiTemplate := appv1.CAPITemplate{}
	log.Infof("getting capitemplate: %s\n", tmName)
	err := crdRestClient.Get().Namespace(namespace).Resource("capitemplates").Name(tmName).Do(ctx).Into(&capiTemplate)
	log.Infof("err: %s\n", err)
	if err != nil {
		return nil, fmt.Errorf("error getting capitemplate %s/%s: %s\n", namespace, tmName, err)
	}
	log.Infof("got capitemplate: %v\n", capiTemplate)

	return &capiTemplate, nil
}

func GetTemplateParams(ctx context.Context, crdRestClient *rest.RESTClient, namespace, tmName string) ([]flavours.Param, error) {
	tm, err := GetTemplate(ctx, crdRestClient, namespace, tmName)
	if err != nil {
		return nil, err
	}
	return flavours.ParamsFromSpec(tm.Spec)
}

// TODO: move this into a grpc handler
func RenderTemplate(ctx context.Context, crdRestClient *rest.RESTClient, namespace, tmName string, vars map[string]string) ([][]byte, error) {
	log.Infof("rendering template with vars: %v", vars)
	tm, err := GetTemplate(ctx, crdRestClient, os.Getenv("POD_NAMESPACE"), tmName)
	if err != nil {
		return nil, err
	}
	log.Infof("got template: %v", tm.ObjectMeta.Name)

	render, err := flavours.Render(tm.Spec, vars)
	if err != nil {
		return nil, err
	}
	return render, nil
}

func LoadTemplatesFromCustomResources(ctx context.Context, crdRestClient *rest.RESTClient, namespace string) (map[string]*appv1.CAPITemplate, error) {
	log.Debugf("querying namespace %s for CAPITemplate resources\n", namespace)
	capiTemplateList := appv1.CAPITemplateList{}
	err := crdRestClient.Get().Resource("capitemplates").Do(ctx).Into(&capiTemplateList)
	if err != nil {
		return nil, fmt.Errorf("error getting capitemplates: %s\n", err)
	}
	log.Debugf("got capitemplates: %v\n", capiTemplateList.Items)

	result := map[string]*appv1.CAPITemplate{}
	for _, ct := range capiTemplateList.Items {
		result[ct.ObjectMeta.Name] = &ct
	}
	return result, nil
}
