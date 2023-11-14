package configuration

import (
	"fmt"
	"time"

	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	corev1 "k8s.io/api/core/v1"

	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	clusterreflectorv1alpha1 "github.com/weaveworks/cluster-reflector-controller/api/v1alpha1"
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
	CategoryAutomation       ObjectCategory = "automation"
	CategorySource           ObjectCategory = "source"
	CategoryEvent            ObjectCategory = "event"
	CategoryGitopsSet        ObjectCategory = "gitopsset"
	CategoryTemplate         ObjectCategory = "template"
	CategoryRBAC             ObjectCategory = "rbac"
	CategoryClusterDiscovery ObjectCategory = "clusterdiscovery"
)

// ObjectKind is the main structur for a object that explorer is able to manage. It includes all the configuration and
// behaviour that is required for both collection and querying.
type ObjectKind struct {
	// Gvk is the GroupVersionKind of the objectKind
	Gvk schema.GroupVersionKind `json:"groupVersionKind"`
	// NewClientObjectFunc is a function that returns a new kuberentes object for the objectKind.
	NewClientObjectFunc func() client.Object `json:"-"`
	// AddToSchemeFunc is a function that adds the objectKind to the kubernetes scheme.
	AddToSchemeFunc func(*runtime.Scheme) error `json:"-"`
	// RetentionPolicy is a function to define retention for objects of this objectkind. For example for event to be retained for 24 hours.
	RetentionPolicy RetentionPolicy `json:"-"`
	// FilterFunc is a function to filter objects of this objectkind. For example to only retain events from a particular source.
	FilterFunc FilterFunc
	// StatusFunc is a function to get the status of an object of this objectkind. It allows to customise status resolution by objectkind.
	StatusFunc func(obj client.Object) ObjectStatus
	// MessageFunc is a function to get the message of an object of this objectkind. It allows to customise message resolution by objectkind.
	MessageFunc func(obj client.Object) string
	// Labels defines a list of labels that you are interested to collect and query for the object kind. For example, templates, defines templateType as label.
	Labels []string
	// Category defines the category of the objectkind. It allows to group objectkinds in the UI.
	Category ObjectCategory
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
		Category:        CategoryAutomation,
	}
	KustomizationObjectKind = ObjectKind{
		Gvk: kustomizev1.GroupVersion.WithKind(kustomizev1.KustomizationKind),
		NewClientObjectFunc: func() client.Object {
			return &kustomizev1.Kustomization{}
		},
		AddToSchemeFunc: kustomizev1.AddToScheme,
		StatusFunc:      defaultFluxObjectStatusFunc,
		MessageFunc:     defaultFluxObjectMessageFunc,
		Category:        CategoryAutomation,
	}
	HelmRepositoryObjectKind = ObjectKind{
		Gvk: sourcev1beta2.GroupVersion.WithKind(sourcev1beta2.HelmRepositoryKind),
		NewClientObjectFunc: func() client.Object {
			return &sourcev1beta2.HelmRepository{}
		},
		AddToSchemeFunc: sourcev1.AddToScheme,
		StatusFunc:      defaultFluxObjectStatusFunc,
		MessageFunc:     defaultFluxObjectMessageFunc,
		Category:        CategorySource,
	}
	HelmChartObjectKind = ObjectKind{
		Gvk: sourcev1beta2.GroupVersion.WithKind(sourcev1beta2.HelmChartKind),
		NewClientObjectFunc: func() client.Object {
			return &sourcev1beta2.HelmChart{}
		},
		AddToSchemeFunc: sourcev1.AddToScheme,
		StatusFunc:      defaultFluxObjectStatusFunc,
		MessageFunc:     defaultFluxObjectMessageFunc,
		Category:        CategorySource,
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
		Category:        CategoryGitopsSet,
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
			"weave.works/template-type",
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

	AutomatedClusterDiscoveryKind = ObjectKind{
		Gvk: clusterreflectorv1alpha1.GroupVersion.WithKind("AutomatedClusterDiscovery"),
		NewClientObjectFunc: func() client.Object {
			return &clusterreflectorv1alpha1.AutomatedClusterDiscovery{}
		},
		AddToSchemeFunc: clusterreflectorv1alpha1.AddToScheme,
		StatusFunc:      defaultFluxObjectStatusFunc,
		MessageFunc:     defaultFluxObjectMessageFunc,
		Category:        CategoryClusterDiscovery,
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
	AutomatedClusterDiscoveryKind,
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
		// Generally, the Ready message has the most useful error message
		if c.Type == "Ready" || c.Type == "Available" {
			return c.Message
		}
	}

	return ""
}

type AutomatedClusterDiscoveryAdaptor struct {
	client.Object
}

func (a *AutomatedClusterDiscoveryAdaptor) GetConditions() []metav1.Condition {
	acd := a.Object.(*clusterreflectorv1alpha1.AutomatedClusterDiscovery)
	return acd.Status.Conditions
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
	case *clusterreflectorv1alpha1.AutomatedClusterDiscovery:
		return &AutomatedClusterDiscoveryAdaptor{Object: t}, nil
	}

	return nil, fmt.Errorf("unknown object type: %T", obj)
}
