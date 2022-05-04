package templates

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/capi"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/tfcontroller"
)

// TemplateGetter implementations get templates by name.
type TemplateGetter interface {
	Get(ctx context.Context, name string) (*apiv1.CAPITemplate, error)
}

// TemplateLister implementations list templates from a Library.
type TemplateLister interface {
	List(ctx context.Context) (map[string]*apiv1.CAPITemplate, error)
}

// TODO: Unify this with Templates which are the same thing.
// TemplateGetter implementations get templates by name.
type TFControllerTemplateGetter interface {
	GetTFControllerTemplate(ctx context.Context, name string) (*apiv1.TFTemplate, error)
}

// TemplateLister implementations list templates from a Library.
type TFControllerTemplateLister interface {
	ListTFControllerTemplate(ctx context.Context) (map[string]*apiv1.TFTemplate, error)
}

// Library represents a library of Templates indexed by name.
type Library interface {
	TemplateGetter
	TemplateLister

	TFControllerTemplateGetter
	TFControllerTemplateLister
}

type ConfigMapLibrary struct {
	Log           logr.Logger
	Client        client.Client
	ConfigMapName string
	Namespace     string
}

func (lib *ConfigMapLibrary) List(ctx context.Context) (map[string]*apiv1.CAPITemplate, error) {
	lib.Log.Info("Querying Kubernetes for configmap", "namespace", lib.Namespace, "configmapName", lib.ConfigMapName)

	templateConfigMap := &v1.ConfigMap{}
	err := lib.Client.Get(ctx, client.ObjectKey{
		Namespace: lib.Namespace,
		Name:      lib.ConfigMapName,
	}, templateConfigMap)
	if errors.IsNotFound(err) {
		return nil, fmt.Errorf("configmap %s not found in %s namespace: %w", lib.ConfigMapName, lib.Namespace, err)
	} else if err != nil {
		return nil, fmt.Errorf("error getting configmap: %w", err)
	}
	lib.Log.Info("Got template configmap", "configmap", templateConfigMap)

	tm, err := capi.ParseConfigMap(*templateConfigMap)
	if errors.IsNotFound(err) {
		return nil, fmt.Errorf("error parsing CAPI templates from configmap: %w", err)
	}
	return tm, nil
}

func (lib *ConfigMapLibrary) Get(ctx context.Context, name string) (*apiv1.CAPITemplate, error) {
	allTemplates, err := lib.List(ctx)
	if err != nil {
		return nil, err
	}
	var foundTemplate *apiv1.CAPITemplate
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

func (lib *ConfigMapLibrary) ListTFControllerTemplate(ctx context.Context) (map[string]*apiv1.TFTemplate, error) {
	lib.Log.Info("Querying Kubernetes for configmap", "namespace", lib.Namespace, "configmapName", lib.ConfigMapName)

	templateConfigMap := &v1.ConfigMap{}
	err := lib.Client.Get(ctx, client.ObjectKey{
		Namespace: lib.Namespace,
		Name:      lib.ConfigMapName,
	}, templateConfigMap)
	if errors.IsNotFound(err) {
		return nil, fmt.Errorf("configmap %s not found in %s namespace: %w", lib.ConfigMapName, lib.Namespace, err)
	} else if err != nil {
		return nil, fmt.Errorf("error getting configmap: %w", err)
	}
	lib.Log.Info("Got template configmap", "configmap", templateConfigMap)

	tm, err := tfcontroller.ParseConfigMap(*templateConfigMap)
	if errors.IsNotFound(err) {
		return nil, fmt.Errorf("error parsing CAPI templates from configmap: %w", err)
	}
	return tm, nil
}

func (lib *ConfigMapLibrary) GetTFControllerTemplate(ctx context.Context, name string) (*apiv1.TFTemplate, error) {
	allTemplates, err := lib.ListTFControllerTemplate(ctx)
	if err != nil {
		return nil, err
	}
	var foundTemplate *apiv1.TFTemplate
	for _, tm := range allTemplates {
		if tm.Name == name {
			foundTemplate = tm
		}
	}
	if foundTemplate == nil {
		return nil, fmt.Errorf("tftemplate %s not found in configmap %s/%s", name, lib.Namespace, lib.ConfigMapName)
	}

	return foundTemplate, nil
}

type CRDLibrary struct {
	Log          logr.Logger
	ClientGetter kube.ClientGetter
	Namespace    string
}

func (lib *CRDLibrary) Get(ctx context.Context, name string) (*apiv1.CAPITemplate, error) {
	lib.Log.Info("Getting client from context")
	cl, err := lib.ClientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	capiTemplate := apiv1.CAPITemplate{}
	lib.Log.Info("Getting capitemplate", "template", name)
	err = cl.Get(ctx, client.ObjectKey{
		Namespace: lib.Namespace,
		Name:      name,
	}, &capiTemplate)
	if err != nil {
		lib.Log.Error(err, "Failed to get capitemplate", "template", name)
		return nil, fmt.Errorf("error getting capitemplate %s/%s: %w", lib.Namespace, name, err)
	}
	lib.Log.Info("Got capitemplate", "template", name)

	return &capiTemplate, nil
}

func (lib *CRDLibrary) List(ctx context.Context) (map[string]*apiv1.CAPITemplate, error) {
	lib.Log.Info("Getting client from context")
	cl, err := lib.ClientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	lib.Log.Info("Querying namespace for CAPITemplate resources", "namespace", lib.Namespace)
	capiTemplateList := apiv1.CAPITemplateList{}
	err = cl.List(ctx, &capiTemplateList, client.InNamespace(lib.Namespace))
	if err != nil {
		return nil, fmt.Errorf("error getting capitemplates: %w", err)
	}
	lib.Log.Info("Got capitemplates", "numberOfTemplates", len(capiTemplateList.Items))

	result := map[string]*apiv1.CAPITemplate{}
	for i, ct := range capiTemplateList.Items {
		result[ct.ObjectMeta.Name] = &capiTemplateList.Items[i]
	}
	return result, nil
}

func (lib *CRDLibrary) GetTFControllerTemplate(ctx context.Context, name string) (*apiv1.TFTemplate, error) {
	lib.Log.Info("Getting client from context")
	cl, err := lib.ClientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	template := apiv1.TFTemplate{}
	lib.Log.Info("Getting template", "template", name)
	err = cl.Get(ctx, client.ObjectKey{
		Namespace: lib.Namespace,
		Name:      name,
	}, &template)
	if err != nil {
		lib.Log.Error(err, "Failed to get template", "template", name)
		return nil, fmt.Errorf("error getting template %s/%s: %w", lib.Namespace, name, err)
	}
	lib.Log.Info("Got template", "template", name)

	return &template, nil
}

func (lib *CRDLibrary) ListTFControllerTemplate(ctx context.Context) (map[string]*apiv1.TFTemplate, error) {
	lib.Log.Info("Getting client from context")
	cl, err := lib.ClientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	lib.Log.Info("Querying namespace for Template resources", "namespace", lib.Namespace)
	templateList := apiv1.TFTemplateList{}
	err = cl.List(ctx, &templateList, client.InNamespace(lib.Namespace))
	if err != nil {
		return nil, fmt.Errorf("error getting templates: %w", err)
	}
	lib.Log.Info("Got templates", "numberOfTemplates", len(templateList.Items))

	result := map[string]*apiv1.TFTemplate{}
	for i, ct := range templateList.Items {
		result[ct.ObjectMeta.Name] = &templateList.Items[i]
	}
	return result, nil
}
