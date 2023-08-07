package configuration

import (
	"fmt"
	"time"

	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	corev1 "k8s.io/api/core/v1"

	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
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
)

type ObjectKind struct {
	Gvk                 schema.GroupVersionKind     `json:"groupVersionKind"`
	NewClientObjectFunc func() client.Object        `json:"-"`
	AddToSchemeFunc     func(*runtime.Scheme) error `json:"-"`
	RetentionPolicy     RetentionPolicy             `json:"-"`
	FilterFunc          FilterFunc
	StatusFunc          func(obj client.Object) ObjectStatus
	MessageFunc         func(obj client.Object) string
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
	}
	ClusterRoleObjectKind = ObjectKind{
		Gvk: rbacv1.SchemeGroupVersion.WithKind("ClusterRole"),
		NewClientObjectFunc: func() client.Object {
			return &rbacv1.ClusterRole{}
		},
		AddToSchemeFunc: rbacv1.AddToScheme,
	}
	RoleBindingObjectKind = ObjectKind{
		Gvk: rbacv1.SchemeGroupVersion.WithKind("RoleBinding"),
		NewClientObjectFunc: func() client.Object {
			return &rbacv1.RoleBinding{}
		},
		AddToSchemeFunc: rbacv1.AddToScheme,
	}
	ClusterRoleBindingObjectKind = ObjectKind{
		Gvk: rbacv1.SchemeGroupVersion.WithKind("ClusterRoleBinding"),
		NewClientObjectFunc: func() client.Object {
			return &rbacv1.ClusterRoleBinding{}
		},
		AddToSchemeFunc: rbacv1.AddToScheme,
	}

	PolicyAgentEventObjectKind = ObjectKind{
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

			return e.Source.Component == "policy-agent"
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
	PolicyAgentEventObjectKind,
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
	case *corev1.Event:
		e, ok := obj.(*corev1.Event)
		if !ok {
			return nil, fmt.Errorf("failed to cast object to event")
		}
		return &eventAdapter{e}, nil
	}

	return nil, fmt.Errorf("unknown object type: %T", obj)
}

type EventLike interface {
	client.Object
	GetConditions() []metav1.Condition
}

type eventAdapter struct {
	*corev1.Event
}

func (ea *eventAdapter) GetConditions() []metav1.Condition {
	cond := metav1.Condition{
		Type:    string(NoStatus),
		Message: ea.Message,
		Status:  "True",
	}
	return []metav1.Condition{cond}
}
