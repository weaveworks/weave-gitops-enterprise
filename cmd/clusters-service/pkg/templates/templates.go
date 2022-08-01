package templates

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/core/logger"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"sigs.k8s.io/controller-runtime/pkg/client"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/capi/v1alpha1"
	gapiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/gitopstemplate/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/templates"
)

// TemplateGetter implementations get templates by name.
type TemplateGetter interface {
	Get(ctx context.Context, name, templateKind string) (templates.Template, error)
}

// TemplateLister implementations list templates from a Library.
type TemplateLister interface {
	List(ctx context.Context, templateKind string) (map[string]templates.Template, error)
}

// Library represents a library of Templates indexed by name.
type Library interface {
	TemplateGetter
	TemplateLister
}

type CRDLibrary struct {
	Log           logr.Logger
	ClientGetter  kube.ClientGetter
	CAPINamespace string
}

func (lib *CRDLibrary) Get(ctx context.Context, name, templateKind string) (templates.Template, error) {
	cl, err := lib.ClientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}
	switch templateKind {
	case capiv1.Kind:
		var t capiv1.CAPITemplate
		lib.Log.V(logger.LogLevelDebug).Info("Getting capitemplate", "template", name)
		err = cl.Get(ctx, client.ObjectKey{
			Namespace: lib.CAPINamespace,
			Name:      name,
		}, &t)
		if err != nil {
			lib.Log.Error(err, "Failed to get capitemplate", "template", name)
			return nil, fmt.Errorf("error getting capitemplate %s/%s: %w", lib.CAPINamespace, name, err)
		}
		lib.Log.V(logger.LogLevelDebug).Info("Got capitemplate", "template", name)
		return &t, nil

	case gapiv1.Kind:
		var t gapiv1.GitOpsTemplate
		lib.Log.V(logger.LogLevelDebug).Info("Getting gitops template", "template", name)
		err = cl.Get(ctx, client.ObjectKey{
			Namespace: lib.CAPINamespace,
			Name:      name,
		}, &t)
		if err != nil {
			lib.Log.Error(err, "Failed to get gitops template", "template", name)
			return nil, fmt.Errorf("error getting gitops template %s/%s: %w", lib.CAPINamespace, name, err)
		}
		lib.Log.V(logger.LogLevelDebug).Info("Got gitops template", "template", name)
		return &t, nil
	}

	return nil, nil
}

func (lib *CRDLibrary) List(ctx context.Context, templateKind string) (map[string]templates.Template, error) {
	cl, err := lib.ClientGetter.Client(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]templates.Template)
	switch templateKind {
	case capiv1.Kind:
		lib.Log.V(logger.LogLevelDebug).Info("Querying namespace for CAPITemplate resources", "namespace", lib.CAPINamespace)
		capiTemplateList := capiv1.CAPITemplateList{}
		err = cl.List(ctx, &capiTemplateList, client.InNamespace(lib.CAPINamespace))
		if err != nil {
			return nil, fmt.Errorf("error getting capitemplates: %w", err)
		}
		lib.Log.V(logger.LogLevelDebug).Info("Got capitemplates", "numberOfTemplates", len(capiTemplateList.Items))
		for i, ct := range capiTemplateList.Items {
			result[ct.ObjectMeta.Name] = &capiTemplateList.Items[i]
		}
	case gapiv1.Kind:
		lib.Log.V(logger.LogLevelDebug).Info("Querying namespace for GitOpsTemplate resources", "namespace", lib.CAPINamespace)
		list := gapiv1.GitOpsTemplateList{}
		err = cl.List(ctx, &list, client.InNamespace(lib.CAPINamespace))
		if err != nil {
			return nil, fmt.Errorf("error getting gitops templates: %w", err)
		}
		lib.Log.V(logger.LogLevelDebug).Info("Got gitops templates", "numberOfTemplates", len(list.Items))
		for i, ct := range list.Items {
			result[ct.ObjectMeta.Name] = &list.Items[i]
		}
	}
	return result, nil
}
