package templates

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	capiv1 "github.com/weaveworks/wks/cmd/capi-server/api/v1alpha1"
	"github.com/weaveworks/wks/cmd/capi-server/pkg/capi"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TemplateGetter implementations get templates by name.
type TemplateGetter interface {
	Get(ctx context.Context, name string) (*capiv1.CAPITemplate, error)
}

// TemplateLister implementations list templates from a Library.
type TemplateLister interface {
	List(ctx context.Context) (map[string]*capiv1.CAPITemplate, error)
}

// Library represents a library of Templates indexed by name.
type Library interface {
	TemplateGetter
	TemplateLister
}

type ConfigMapLibrary struct {
	Client        client.Client
	ConfigMapName string
	Namespace     string
}

func (lib *ConfigMapLibrary) List(ctx context.Context) (map[string]*capiv1.CAPITemplate, error) {
	fmt.Printf("querying kubernetes for configmap: %s/%s\n", lib.Namespace, lib.ConfigMapName)

	templateConfigMap := &v1.ConfigMap{}
	err := lib.Client.Get(ctx, client.ObjectKey{
		Namespace: lib.Namespace,
		Name:      lib.ConfigMapName,
	}, templateConfigMap)
	if errors.IsNotFound(err) {
		return nil, fmt.Errorf("configmap %s not found in %s namespace", lib.ConfigMapName, lib.Namespace)
	} else if err != nil {
		return nil, fmt.Errorf("error getting configmap: %s", err)
	}
	log.Debugf("got template configmap: %v\n", templateConfigMap)

	tm, err := capi.ParseConfigMap(*templateConfigMap)
	if errors.IsNotFound(err) {
		return nil, fmt.Errorf("error parsing CAPI templates from configmap: %s", err)
	}
	return tm, nil
}

func (lib *ConfigMapLibrary) Get(ctx context.Context, name string) (*capiv1.CAPITemplate, error) {
	allTemplates, err := lib.List(ctx)
	if err != nil {
		return nil, err
	}
	var foundTemplate *capiv1.CAPITemplate
	for _, tm := range allTemplates {
		if tm.Name == name {
			foundTemplate = tm
		}
	}
	if foundTemplate == nil {
		return nil, fmt.Errorf("capitemplate %s not found in configmap %s/%s", name, lib.Namespace, lib.ConfigMapName)
	}

	return foundTemplate, nil
}

type CRDLibrary struct {
	Client    client.Client
	Namespace string
}

func (lib *CRDLibrary) Get(ctx context.Context, name string) (*capiv1.CAPITemplate, error) {
	capiTemplate := capiv1.CAPITemplate{}
	log.Infof("getting capitemplate: %s\n", name)
	err := lib.Client.Get(ctx, client.ObjectKey{
		Namespace: lib.Namespace,
		Name:      name,
	}, &capiTemplate)
	log.Infof("err: %s\n", err)
	if err != nil {
		return nil, fmt.Errorf("error getting capitemplate %s/%s: %s", lib.Namespace, name, err)
	}
	log.Infof("got capitemplate: %v\n", name)

	return &capiTemplate, nil
}

func (lib *CRDLibrary) List(ctx context.Context) (map[string]*capiv1.CAPITemplate, error) {
	log.Infof("querying namespace %s for CAPITemplate resources\n", lib.Namespace)
	capiTemplateList := capiv1.CAPITemplateList{}
	err := lib.Client.List(ctx, &capiTemplateList, client.InNamespace(lib.Namespace))
	if err != nil {
		return nil, fmt.Errorf("error getting capitemplates: %s", err)
	}
	log.Infof("got capitemplates len: %v\n", len(capiTemplateList.Items))

	result := map[string]*capiv1.CAPITemplate{}
	for i, ct := range capiTemplateList.Items {
		result[ct.ObjectMeta.Name] = &capiTemplateList.Items[i]
	}
	return result, nil
}
