package templates

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/capi/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
	tapiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/tfcontroller/v1alpha1"
)

// TemplateGetter implementations get templates by name.
type TemplateGetter interface {
	Get(ctx context.Context, name, templateKind string) (*templates.Template, error)
}

// TemplateLister implementations list templates from a Library.
type TemplateLister interface {
	List(ctx context.Context, templateKind string) (map[string]*templates.Template, error)
}

// Library represents a library of Templates indexed by name.
type Library interface {
	TemplateGetter
	TemplateLister
}

type ConfigMapLibrary struct {
	Log           logr.Logger
	Client        client.Client
	ConfigMapName string
	Namespace     string
}

func (lib *ConfigMapLibrary) List(ctx context.Context, templateKind string) (map[string]*templates.Template, error) {
	lib.Log.Info("Querying Kubernetes for configmap", "namespace", lib.Namespace, "configmapName", lib.ConfigMapName, "kind", templateKind)

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

	tm, err := ParseConfigMap(*templateConfigMap)
	if errors.IsNotFound(err) {
		return nil, fmt.Errorf("error parsing CAPI templates from configmap: %w", err)
	}
	m := make(map[string]*templates.Template)
	for k, v := range tm {
		if v.Kind == templateKind {
			m[k] = v
		}
	}
	return m, nil
}

func (lib *ConfigMapLibrary) Get(ctx context.Context, name, templateKind string) (*templates.Template, error) {
	allTemplates, err := lib.List(ctx, templateKind)
	if err != nil {
		return nil, err
	}
	var t *templates.Template
	for _, tm := range allTemplates {
		if tm.Name == name && tm.Kind == templateKind {
			t = tm
		}
	}
	if t == nil {
		return nil, fmt.Errorf("terraform template %s not found in configmap %s/%s", name, lib.Namespace, lib.ConfigMapName)
	}
	return t, nil
}

type CRDLibrary struct {
	Log          logr.Logger
	ClientGetter kube.ClientGetter
	Namespace    string
}

func (lib *CRDLibrary) Get(ctx context.Context, name, templateKind string) (*templates.Template, error) {
	lib.Log.Info("Getting client from context")
	cl, err := lib.ClientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}
	var result *templates.Template
	switch templateKind {

	case capiv1.Kind:
		var t capiv1.CAPITemplate
		lib.Log.Info("Getting capitemplate", "template", name)
		err = cl.Get(ctx, client.ObjectKey{
			Namespace: lib.Namespace,
			Name:      name,
		}, &t)
		if err != nil {
			lib.Log.Error(err, "Failed to get capitemplate", "template", name)
			return nil, fmt.Errorf("error getting capitemplate %s/%s: %w", lib.Namespace, name, err)
		}
		lib.Log.Info("Got capitemplate", "template", name)
		result = &t.Template
	case tapiv1.Kind:
		var t tapiv1.TFTemplate
		lib.Log.Info("Getting tftemplate", "template", name)
		err = cl.Get(ctx, client.ObjectKey{
			Namespace: lib.Namespace,
			Name:      name,
		}, &t)
		if err != nil {
			lib.Log.Error(err, "Failed to get tftemplate", "template", name)
			return nil, fmt.Errorf("error getting tftemplate %s/%s: %w", lib.Namespace, name, err)
		}
		lib.Log.Info("Got tftemplate", "template", name)
		result = &t.Template
	}

	return result, nil
}

func (lib *CRDLibrary) List(ctx context.Context, templateKind string) (map[string]*templates.Template, error) {
	lib.Log.Info("Getting client from context")
	cl, err := lib.ClientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*templates.Template)
	switch templateKind {
	case capiv1.Kind:
		lib.Log.Info("Querying namespace for CAPITemplate resources", "namespace", lib.Namespace)
		capiTemplateList := capiv1.CAPITemplateList{}
		err = cl.List(ctx, &capiTemplateList, client.InNamespace(lib.Namespace))
		if err != nil {
			return nil, fmt.Errorf("error getting capitemplates: %w", err)
		}
		lib.Log.Info("Got capitemplates", "numberOfTemplates", len(capiTemplateList.Items))
		for i, ct := range capiTemplateList.Items {
			result[ct.ObjectMeta.Name] = &capiTemplateList.Items[i].Template
		}
	case tapiv1.Kind:
		lib.Log.Info("Querying namespace for TFTemplate resources", "namespace", lib.Namespace)
		tfTemplateList := tapiv1.TFTemplateList{}
		err = cl.List(ctx, &tfTemplateList, client.InNamespace(lib.Namespace))
		if err != nil {
			return nil, fmt.Errorf("error getting tftemplates: %w", err)
		}
		lib.Log.Info("Got tftemplates", "numberOfTemplates", len(tfTemplateList.Items))
		for i, ct := range tfTemplateList.Items {
			result[ct.ObjectMeta.Name] = &tfTemplateList.Items[i].Template
		}
	}
	return result, nil
}
