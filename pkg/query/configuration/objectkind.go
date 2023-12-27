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

// ObjectKind is the main structure for a object that explorer is able to manage. It includes all the configuration and
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
	// GetConditionsFunc is a function to get the conditions
	GetConditionsFunc func(obj client.Object) ([]metav1.Condition, error)
	// GetSuspendedFunc is a function to calculate whether the given object is suspended.
	GetSuspendedFunc func(obj client.Object) (bool, error)
	// StatusFunc is a function to calculate the status out of the object status and its configuration given by objectkind.
	StatusFunc func(obj client.Object, objectKind ObjectKind) (ObjectStatus, error)
	// MessageFunc is a function to get the message of an object of this objectkind. It allows to customise message resolution by objectkind.
	MessageFunc func(obj client.Object, objectKind ObjectKind) (string, error)
	// Labels defines a list of labels that you are interested to collect and query for the object kind. For example, templates, defines templateType as label.
	Labels []string
	// Category defines the category of the objectkind. It allows to group objectkinds in the UI.
	Category ObjectCategory
	// HumanReadableLabelKeys is a map of label keys to human readable names. It allows to customise the label names in the UI.
	// Values should be dash case: template-type, some-value, etc.
	HumanReadableLabelKeys map[string]string
}

type ObjectStatus string

const (
	Success       ObjectStatus = "Success"
	Failed        ObjectStatus = "Failed"
	Reconciling   ObjectStatus = "Reconciling"
	Suspended     ObjectStatus = "Suspended"
	PendingAction ObjectStatus = "PendingAction"
	NoStatus      ObjectStatus = "-"
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

var (
	HelmReleaseObjectKind = ObjectKind{
		Gvk: helmv2beta1.GroupVersion.WithKind(helmv2beta1.HelmReleaseKind),
		NewClientObjectFunc: func() client.Object {
			return &helmv2beta1.HelmRelease{}
		},
		AddToSchemeFunc: helmv2beta1.AddToScheme,
		GetConditionsFunc: func(obj client.Object) ([]metav1.Condition, error) {
			hr, ok := obj.(*helmv2beta1.HelmRelease)
			if !ok {
				return nil, fmt.Errorf("object is not a HelmRelease")
			}
			return hr.GetConditions(), nil
		},
		GetSuspendedFunc: func(obj client.Object) (bool, error) {
			hr, ok := obj.(*helmv2beta1.HelmRelease)
			if !ok {
				return false, fmt.Errorf("object is not a HelmRelease")
			}
			return hr.Spec.Suspend, nil
		},
		StatusFunc:  defaultStatusFunc,
		MessageFunc: defaultMessageFunc,
		Category:    CategoryAutomation,
	}

	KustomizationObjectKind = ObjectKind{
		Gvk: kustomizev1.GroupVersion.WithKind(kustomizev1.KustomizationKind),
		NewClientObjectFunc: func() client.Object {
			return &kustomizev1.Kustomization{}
		},
		AddToSchemeFunc: kustomizev1.AddToScheme,
		GetConditionsFunc: func(obj client.Object) ([]metav1.Condition, error) {
			ks, ok := obj.(*kustomizev1.Kustomization)
			if !ok {
				return nil, fmt.Errorf("object is not a Kustomization")
			}
			return ks.GetConditions(), nil
		},
		GetSuspendedFunc: func(obj client.Object) (bool, error) {
			ks, ok := obj.(*kustomizev1.Kustomization)
			if !ok {
				return false, fmt.Errorf("object is not a Kustomization")
			}
			return ks.Spec.Suspend, nil
		},
		StatusFunc:  defaultStatusFunc,
		MessageFunc: defaultMessageFunc,
		Category:    CategoryAutomation,
	}

	HelmRepositoryObjectKind = ObjectKind{
		Gvk: sourcev1beta2.GroupVersion.WithKind(sourcev1beta2.HelmRepositoryKind),
		NewClientObjectFunc: func() client.Object {
			return &sourcev1beta2.HelmRepository{}
		},
		AddToSchemeFunc: sourcev1.AddToScheme,
		GetConditionsFunc: func(obj client.Object) ([]metav1.Condition, error) {
			hr, ok := obj.(*sourcev1beta2.HelmRepository)
			if !ok {
				return nil, fmt.Errorf("object is not a HelmRepository")
			}
			return hr.GetConditions(), nil
		},
		GetSuspendedFunc: func(obj client.Object) (bool, error) {
			hr, ok := obj.(*sourcev1beta2.HelmRepository)
			if !ok {
				return false, fmt.Errorf("object is not a HelmRepository")
			}
			return hr.Spec.Suspend, nil
		},
		StatusFunc:  defaultStatusFunc,
		MessageFunc: defaultMessageFunc,
		Category:    CategorySource,
	}

	HelmChartObjectKind = ObjectKind{
		Gvk: sourcev1beta2.GroupVersion.WithKind(sourcev1beta2.HelmChartKind),
		NewClientObjectFunc: func() client.Object {
			return &sourcev1beta2.HelmChart{}
		},
		AddToSchemeFunc: sourcev1.AddToScheme,
		GetConditionsFunc: func(obj client.Object) ([]metav1.Condition, error) {
			chart, ok := obj.(*sourcev1beta2.HelmChart)
			if !ok {
				return nil, fmt.Errorf("object is not a HelmChart")
			}
			return chart.GetConditions(), nil
		},
		GetSuspendedFunc: func(obj client.Object) (bool, error) {
			chart, ok := obj.(*sourcev1beta2.HelmChart)
			if !ok {
				return false, fmt.Errorf("object is not a HelmChart")
			}
			return chart.Spec.Suspend, nil
		},
		StatusFunc:  defaultStatusFunc,
		MessageFunc: defaultMessageFunc,
		Category:    CategorySource,
	}

	GitRepositoryObjectKind = ObjectKind{
		Gvk: sourcev1.GroupVersion.WithKind(sourcev1.GitRepositoryKind),
		NewClientObjectFunc: func() client.Object {
			return &sourcev1.GitRepository{}
		},
		AddToSchemeFunc: sourcev1.AddToScheme,
		GetConditionsFunc: func(obj client.Object) ([]metav1.Condition, error) {
			repo, ok := obj.(*sourcev1.GitRepository)
			if !ok {
				return nil, fmt.Errorf("object is not a GitRepository")
			}
			return repo.GetConditions(), nil
		},
		GetSuspendedFunc: func(obj client.Object) (bool, error) {
			repo, ok := obj.(*sourcev1.GitRepository)
			if !ok {
				return false, fmt.Errorf("object is not a GitRepository")
			}
			return repo.Spec.Suspend, nil
		},
		StatusFunc:  defaultStatusFunc,
		MessageFunc: defaultMessageFunc,
		Category:    CategorySource,
	}

	OCIRepositoryObjectKind = ObjectKind{
		Gvk: sourcev1beta2.GroupVersion.WithKind(sourcev1beta2.OCIRepositoryKind),
		NewClientObjectFunc: func() client.Object {
			return &sourcev1beta2.OCIRepository{}
		},
		AddToSchemeFunc: sourcev1beta2.AddToScheme,
		GetConditionsFunc: func(obj client.Object) ([]metav1.Condition, error) {
			repo, ok := obj.(*sourcev1beta2.OCIRepository)
			if !ok {
				return nil, fmt.Errorf("object is not an OCIRepository")
			}
			return repo.GetConditions(), nil
		},
		GetSuspendedFunc: func(obj client.Object) (bool, error) {
			repo, ok := obj.(*sourcev1beta2.OCIRepository)
			if !ok {
				return false, fmt.Errorf("object is not an OCIRepository")
			}
			return repo.Spec.Suspend, nil
		},
		StatusFunc:  defaultStatusFunc,
		MessageFunc: defaultMessageFunc,
		Category:    CategorySource,
	}

	BucketObjectKind = ObjectKind{
		Gvk: sourcev1beta2.GroupVersion.WithKind(sourcev1beta2.BucketKind),
		NewClientObjectFunc: func() client.Object {
			return &sourcev1beta2.Bucket{}
		},
		AddToSchemeFunc: sourcev1beta2.AddToScheme,
		GetConditionsFunc: func(obj client.Object) ([]metav1.Condition, error) {
			b, ok := obj.(*sourcev1beta2.Bucket)
			if !ok {
				return nil, fmt.Errorf("object is not a Bucket")
			}
			return b.GetConditions(), nil
		},
		GetSuspendedFunc: func(obj client.Object) (bool, error) {
			b, ok := obj.(*sourcev1beta2.Bucket)
			if !ok {
				return false, fmt.Errorf("object is not a Bucket")
			}
			return b.Spec.Suspend, nil
		},
		StatusFunc:  defaultStatusFunc,
		MessageFunc: defaultMessageFunc,
		Category:    CategorySource,
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
		StatusFunc: func(obj client.Object, _ ObjectKind) (ObjectStatus, error) {
			e, ok := obj.(*corev1.Event)
			if !ok {
				return "", fmt.Errorf("object is not an Event")
			}

			if e.Type == "Normal" {
				return Success, nil
			}

			return Failed, nil
		},
		MessageFunc: func(obj client.Object, _ ObjectKind) (string, error) {
			e, ok := obj.(*corev1.Event)
			if !ok {
				return "", fmt.Errorf("object is not an Event")
			}

			return e.Message, nil
		},
		Category: CategoryEvent,
	}

	GitOpsSetsObjectKind = ObjectKind{
		Gvk: gitopssets.GroupVersion.WithKind("GitOpsSet"),
		NewClientObjectFunc: func() client.Object {
			return &gitopssets.GitOpsSet{}
		},
		AddToSchemeFunc: gitopssets.AddToScheme,
		GetConditionsFunc: func(obj client.Object) ([]metav1.Condition, error) {
			gs, ok := obj.(*gitopssets.GitOpsSet)
			if !ok {
				return nil, fmt.Errorf("object is not a GitOpsSet")
			}
			return gs.GetConditions(), nil
		},
		GetSuspendedFunc: func(obj client.Object) (bool, error) {
			gs, ok := obj.(*gitopssets.GitOpsSet)
			if !ok {
				return false, fmt.Errorf("object is not a GitOpsSet")
			}
			return gs.Spec.Suspend, nil
		},
		StatusFunc:  defaultStatusFunc,
		MessageFunc: defaultMessageFunc,
		Category:    CategoryGitopsSet,
	}

	GitopsTemplateObjectKind = ObjectKind{
		Gvk: gapiv1.GroupVersion.WithKind(gapiv1.Kind),
		NewClientObjectFunc: func() client.Object {
			return &gapiv1.GitOpsTemplate{}
		},
		AddToSchemeFunc: gapiv1.AddToScheme,
		StatusFunc:      noStatusFunc,
		MessageFunc: func(obj client.Object, _ ObjectKind) (string, error) {
			e, ok := obj.(*gapiv1.GitOpsTemplate)
			if !ok {
				return "", fmt.Errorf("object is not a GitOpsTemplate")
			}

			return e.Spec.Description, nil
		},
		Labels: []string{
			"weave.works/template-type",
		},
		Category: CategoryTemplate,
		HumanReadableLabelKeys: map[string]string{
			"weave.works/template-type": "template-type",
		},
	}

	CapiTemplateObjectKind = ObjectKind{
		Gvk: capiv1.GroupVersion.WithKind(capiv1.Kind),
		NewClientObjectFunc: func() client.Object {
			return &capiv1.CAPITemplate{}
		},
		AddToSchemeFunc: capiv1.AddToScheme,
		StatusFunc:      noStatusFunc,
		MessageFunc: func(obj client.Object, _ ObjectKind) (string, error) {
			e, ok := obj.(*capiv1.CAPITemplate)
			if !ok {
				return "", fmt.Errorf("object is not a CAPITemplate")
			}

			return e.Spec.Description, nil
		},
		Category: CategoryTemplate,
	}

	AutomatedClusterDiscoveryObjectKind = ObjectKind{
		Gvk: clusterreflectorv1alpha1.GroupVersion.WithKind("AutomatedClusterDiscovery"),
		NewClientObjectFunc: func() client.Object {
			return &clusterreflectorv1alpha1.AutomatedClusterDiscovery{}
		},
		AddToSchemeFunc: clusterreflectorv1alpha1.AddToScheme,
		GetConditionsFunc: func(obj client.Object) ([]metav1.Condition, error) {
			acd, ok := obj.(*clusterreflectorv1alpha1.AutomatedClusterDiscovery)
			if !ok {
				return nil, fmt.Errorf("object is not an AutomatedClusterDiscovery")
			}
			return acd.Status.Conditions, nil
		},
		GetSuspendedFunc: func(obj client.Object) (bool, error) {
			acd, ok := obj.(*clusterreflectorv1alpha1.AutomatedClusterDiscovery)
			if !ok {
				return false, fmt.Errorf("object is not an AutomatedClusterDiscovery")
			}
			return acd.Spec.Suspend, nil
		},
		StatusFunc:  defaultStatusFunc,
		MessageFunc: defaultMessageFunc,
		Category:    CategoryClusterDiscovery,
	}
)

func noStatusFunc(_ client.Object, _ ObjectKind) (ObjectStatus, error) {
	return NoStatus, nil
}

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
	AutomatedClusterDiscoveryObjectKind,
}

// SupportedRbacKinds list with the default supported RBAC resources.
var SupportedRbacKinds = []ObjectKind{
	RoleObjectKind,
	ClusterRoleObjectKind,
	RoleBindingObjectKind,
	ClusterRoleBindingObjectKind,
}

// defaultStatusFunc is the default status function for Flux objects.
// The status of Flux objects is computed based on the values of the Conditions and Spec.Suspend fields,
// so we can standardize on that, based on the current object statuses, computed in non-Explorer UI:
// https://github.com/weaveworks/weave-gitops/blob/cc3c17632334ffa56838c4765e68ce388bde6b2f/ui/components/KubeStatusIndicator.tsx#L18-L25
// TODO: Remove the reference to non-Explorer UI logic once we move computing all object statuses to the backend.
func defaultStatusFunc(obj client.Object, objectKind ObjectKind) (ObjectStatus, error) {
	if objectKind.GetConditionsFunc == nil {
		return "", fmt.Errorf("missing get conditions func")
	}

	if objectKind.StatusFunc == nil {
		return "", fmt.Errorf("missing status func")
	}

	conditions, err := objectKind.GetConditionsFunc(obj)
	if err != nil {
		return "", fmt.Errorf("getting object conditions: %w", err)
	}

	suspended, err := objectKind.GetSuspendedFunc(obj)
	if err != nil {
		return "", fmt.Errorf("getting suspended object status: %w", err)
	}

	if suspended {
		return Suspended, nil
	}

	for _, c := range conditions {
		if ObjectStatus(c.Type) == NoStatus {
			return NoStatus, nil
		}

		if c.Type == "Ready" || c.Type == "Available" {
			if c.Status == "True" {
				return Success, nil
			}

			if c.Status == "Unknown" {
				if c.Reason == "Progressing" {
					return Reconciling, nil
				}
				if c.Reason == "TerraformPlannedWithChanges" {
					return PendingAction, nil
				}
			}

			return Failed, nil
		}
	}

	return Failed, nil
}

func defaultMessageFunc(obj client.Object, objectKind ObjectKind) (string, error) {
	if objectKind.GetConditionsFunc == nil {
		return "", fmt.Errorf("missing get conditions func")
	}

	conditions, err := objectKind.GetConditionsFunc(obj)
	if err != nil {
		return "", fmt.Errorf("getting object conditions: %w", err)
	}

	for _, c := range conditions {
		// Generally, the Ready message has the most useful error message
		if c.Type == "Ready" || c.Type == "Available" {
			return c.Message, nil
		}
	}

	return "", nil
}
