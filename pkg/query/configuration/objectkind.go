package configuration

import (
	"fmt"
	"time"

	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	corev1 "k8s.io/api/core/v1"

	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	gitopssets "github.com/weaveworks/gitopssets-controller/api/v1alpha1"
	capiv1 "github.com/weaveworks/templates-controller/apis/capi/v1alpha2"
	gapiv1 "github.com/weaveworks/templates-controller/apis/gitops/v1alpha2"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FilterFunc can be used to only retain relevant objects.
// For example, we may want to keep Events, but only Events from a particular source.
type FilterFunc func(obj client.Object) bool

type ObjectCategory string

const (
	CategoryAutomation ObjectCategory = "automation"
	CategorySource     ObjectCategory = "source"
	CategoryEvent      ObjectCategory = "event"
	CategoryGitopsSet  ObjectCategory = "gitopsset"
	CategoryTemplate   ObjectCategory = "template"
	CategoryRBAC       ObjectCategory = "rbac"
)

type ObjectKind struct {
	Gvk                 schema.GroupVersionKind     `json:"groupVersionKind"`
	NewClientObjectFunc func() client.Object        `json:"-"`
	AddToSchemeFunc     func(*runtime.Scheme) error `json:"-"`
	RetentionPolicy     RetentionPolicy             `json:"-"`
	FilterFunc          FilterFunc
	StatusFunc          func(obj client.Object) ObjectStatus
	MessageFunc         func(obj client.Object) string
	Labels              []string
	Category            ObjectCategory
}

type ObjectStatus string

const (
	Success  ObjectStatus = "Success"
	Failed   ObjectStatus = "Failed"
	NoStatus ObjectStatus = "-"
)

func (ok ObjectKind) String() string {
	return ok.Gvk.String()
}

func (o ObjectKind) Validate() error {
	if o.Gvk.Kind == "" {
		return fmt.Errorf("missing gvk")
	}
	if o.NewClientObjectFunc == nil {
		return fmt.Errorf("missing client func")
	}
	if o.AddToSchemeFunc == nil {
		return fmt.Errorf("missing add to scheme func")
	}

	if o.StatusFunc == nil {
		return fmt.Errorf("missing status func")
	}

	if o.MessageFunc == nil {
		return fmt.Errorf("missing message func")
	}

	if o.Category == "" {
		return fmt.Errorf("missing category")
	}

	return nil
}

type FluxObject interface {
	client.Object
	GetConditions() []metav1.Condition
}

var (
	HelmReleaseObjectKind = ObjectKind{
		Gvk: helmv2beta1.GroupVersion.WithKind(helmv2beta1.HelmReleaseKind),
		NewClientObjectFunc: func() client.Object {
			return &helmv2beta1.HelmRelease{}
		},
		AddToSchemeFunc: helmv2beta1.AddToScheme,
		StatusFunc:      defaultFluxObjectStatusFunc,
		MessageFunc:     defaultFluxObjectMessageFunc,

		Category: CategoryAutomation,
	}
	KustomizationObjectKind = ObjectKind{
		Gvk: kustomizev1.GroupVersion.WithKind(kustomizev1.KustomizationKind),
		NewClientObjectFunc: func() client.Object {
			return &kustomizev1.Kustomization{}
		},
		AddToSchemeFunc: kustomizev1.AddToScheme,
		StatusFunc:      defaultFluxObjectStatusFunc,
		MessageFunc:     defaultFluxObjectMessageFunc,

		Category: CategoryAutomation,
	}
	HelmRepositoryObjectKind = ObjectKind{
		Gvk: sourcev1beta2.GroupVersion.WithKind(sourcev1beta2.HelmRepositoryKind),
		NewClientObjectFunc: func() client.Object {
			return &sourcev1beta2.HelmRepository{}
		},
		AddToSchemeFunc: sourcev1.AddToScheme,
		StatusFunc:      defaultFluxObjectStatusFunc,
		MessageFunc:     defaultFluxObjectMessageFunc,

		Category: CategorySource,
	}
	HelmChartObjectKind = ObjectKind{
		Gvk: sourcev1beta2.GroupVersion.WithKind(sourcev1beta2.HelmChartKind),
		NewClientObjectFunc: func() client.Object {
			return &sourcev1beta2.HelmChart{}
		},
		AddToSchemeFunc: sourcev1.AddToScheme,
		StatusFunc:      defaultFluxObjectStatusFunc,
		MessageFunc:     defaultFluxObjectMessageFunc,

		Category: CategorySource,
	}
	GitRepositoryObjectKind = ObjectKind{
		Gvk: sourcev1.GroupVersion.WithKind(sourcev1.GitRepositoryKind),
		NewClientObjectFunc: func() client.Object {
			return &sourcev1.GitRepository{}
		},
		AddToSchemeFunc: sourcev1.AddToScheme,
		StatusFunc:      defaultFluxObjectStatusFunc,
		MessageFunc:     defaultFluxObjectMessageFunc,
		Category:        CategorySource,
	}
	OCIRepositoryObjectKind = ObjectKind{
		Gvk: sourcev1beta2.GroupVersion.WithKind(sourcev1beta2.OCIRepositoryKind),
		NewClientObjectFunc: func() client.Object {
			return &sourcev1beta2.OCIRepository{}
		},
		AddToSchemeFunc: sourcev1beta2.AddToScheme,
		StatusFunc:      defaultFluxObjectStatusFunc,
		MessageFunc:     defaultFluxObjectMessageFunc,
		Category:        CategorySource,
	}
	BucketObjectKind = ObjectKind{
		Gvk: sourcev1beta2.GroupVersion.WithKind(sourcev1beta2.BucketKind),
		NewClientObjectFunc: func() client.Object {
			return &sourcev1beta2.Bucket{}
		},
		AddToSchemeFunc: sourcev1beta2.AddToScheme,
		StatusFunc:      defaultFluxObjectStatusFunc,
		MessageFunc:     defaultFluxObjectMessageFunc,
		Category:        CategorySource,
	}
	RoleObjectKind = ObjectKind{
		Gvk: rbacv1.SchemeGroupVersion.WithKind("Role"),
		NewClientObjectFunc: func() client.Object {
			return &rbacv1.Role{}
		},
		AddToSchemeFunc: rbacv1.AddToScheme,
		Category:        CategoryRBAC,
	}
	ClusterRoleObjectKind = ObjectKind{
		Gvk: rbacv1.SchemeGroupVersion.WithKind("ClusterRole"),
		NewClientObjectFunc: func() client.Object {
			return &rbacv1.ClusterRole{}
		},
		AddToSchemeFunc: rbacv1.AddToScheme,
		Category:        CategoryRBAC,
	}
	RoleBindingObjectKind = ObjectKind{
		Gvk: rbacv1.SchemeGroupVersion.WithKind("RoleBinding"),
		NewClientObjectFunc: func() client.Object {
			return &rbacv1.RoleBinding{}
		},
		AddToSchemeFunc: rbacv1.AddToScheme,
		Category:        CategoryRBAC,
	}
	ClusterRoleBindingObjectKind = ObjectKind{
		Gvk: rbacv1.SchemeGroupVersion.WithKind("ClusterRoleBinding"),
		NewClientObjectFunc: func() client.Object {
			return &rbacv1.ClusterRoleBinding{}
		},
		AddToSchemeFunc: rbacv1.AddToScheme,
		Category:        CategoryRBAC,
	}

	PolicyAgentAuditEventObjectKind = ObjectKind{
		Gvk: corev1.SchemeGroupVersion.WithKind("Event"),
		NewClientObjectFunc: func() client.Object {
			return &corev1.Event{}
		},
		AddToSchemeFunc: corev1.AddToScheme,
		FilterFunc: func(obj client.Object) bool {
			e, ok := obj.(*corev1.Event)
			if !ok {
				return false
			}

			return e.Labels["pac.weave.works/type"] == "Audit" && e.Source.Component == "policy-agent"
		},
		RetentionPolicy: RetentionPolicy(24 * time.Hour),
		StatusFunc: func(obj client.Object) ObjectStatus {
			e, ok := obj.(*corev1.Event)
			if !ok {
				return NoStatus
			}

			if e.Type == "Normal" {
				return Success
			}

			return Failed
		},
		MessageFunc: func(obj client.Object) string {
			e, ok := obj.(*corev1.Event)
			if !ok {
				return ""
			}

			return e.Message
		},
		Category: CategoryEvent,
	}

	GitOpsSetsObjectKind = ObjectKind{
		Gvk: gitopssets.GroupVersion.WithKind("GitOpsSet"),
		NewClientObjectFunc: func() client.Object {
			return &gitopssets.GitOpsSet{}
		},
		AddToSchemeFunc: gitopssets.AddToScheme,
		StatusFunc:      defaultFluxObjectStatusFunc,
		MessageFunc:     defaultFluxObjectMessageFunc,

		Category: CategoryGitopsSet,
	}

	GitopsTemplateObjectKind = ObjectKind{
		Gvk: gapiv1.GroupVersion.WithKind(gapiv1.Kind),
		NewClientObjectFunc: func() client.Object {
			return &gapiv1.GitOpsTemplate{}
		},
		AddToSchemeFunc: gapiv1.AddToScheme,
		StatusFunc: func(obj client.Object) ObjectStatus {
			return NoStatus
		},
		MessageFunc: func(obj client.Object) string {
			e, ok := obj.(*gapiv1.GitOpsTemplate)
			if !ok {
				return ""
			}

			return e.Spec.Description
		},
		Labels: []string{
			"templateType",
		},
		Category: CategoryTemplate,
	}
	CapiTemplateObjectKind = ObjectKind{
		Gvk: capiv1.GroupVersion.WithKind(capiv1.Kind),
		NewClientObjectFunc: func() client.Object {
			return &capiv1.CAPITemplate{}
		},
		AddToSchemeFunc: capiv1.AddToScheme,
		StatusFunc: func(obj client.Object) ObjectStatus {
			return NoStatus
		},
		MessageFunc: func(obj client.Object) string {
			e, ok := obj.(*capiv1.CAPITemplate)
			if !ok {
				return ""
			}

			return e.Spec.Description
		},
		Category: CategoryTemplate,
	}
)

// SupportedObjectKinds list with the default supported Object resources to query.
var SupportedObjectKinds = []ObjectKind{
	HelmReleaseObjectKind,
	KustomizationObjectKind,
	HelmRepositoryObjectKind,
	HelmChartObjectKind,
	GitRepositoryObjectKind,
	OCIRepositoryObjectKind,
	BucketObjectKind,
	PolicyAgentAuditEventObjectKind,
	GitOpsSetsObjectKind,
	GitopsTemplateObjectKind,
	CapiTemplateObjectKind,
}

// SupportedRbacKinds list with the default supported RBAC resources.
var SupportedRbacKinds = []ObjectKind{
	RoleObjectKind,
	ClusterRoleObjectKind,
	RoleBindingObjectKind,
	ClusterRoleBindingObjectKind,
}

// defaultFluxObjectStatusFunc is the default status function for Flux objects.
// Flux objects all report status via the Conditions field, so we can standardize on that.
func defaultFluxObjectStatusFunc(obj client.Object) ObjectStatus {
	fo, err := ToFluxObject(obj)
	if err != nil {
		return Failed
	}

	for _, c := range fo.GetConditions() {
		if ObjectStatus(c.Type) == NoStatus {
			return NoStatus
		}
		if c.Type == "Ready" || c.Type == "Available" {
			if c.Status == "True" {
				return Success
			}

			return Failed
		}
	}

	return Failed
}

func defaultFluxObjectMessageFunc(obj client.Object) string {
	fo, err := ToFluxObject(obj)
	if err != nil {
		return ""
	}

	for _, c := range fo.GetConditions() {
		if c.Message != "" {
			return c.Message
		}
	}

	return ""
}

func ToFluxObject(obj client.Object) (FluxObject, error) {
	switch t := obj.(type) {
	case *helmv2beta1.HelmRelease:
		return t, nil
	case *kustomizev1.Kustomization:
		return t, nil
	case *sourcev1beta2.HelmRepository:
		return t, nil
	case *sourcev1beta2.HelmChart:
		return t, nil
	case *sourcev1beta2.Bucket:
		return t, nil
	case *sourcev1.GitRepository:
		return t, nil
	case *sourcev1beta2.OCIRepository:
		return t, nil
	case *gitopssets.GitOpsSet:
		return t, nil
	}

	return nil, fmt.Errorf("unknown object type: %T", obj)
}
