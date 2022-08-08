package tenancy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/hashicorp/go-multierror"
	pacv2beta1 "github.com/weaveworks/policy-agent/api/v2beta1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	errs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const tenantLabel = "toolkit.fluxcd.io/tenant"

var (
	namespaceTypeMeta      = typeMeta("Namespace", "v1")
	serviceAccountTypeMeta = typeMeta("ServiceAccount", "v1")
	roleBindingTypeMeta    = typeMeta("RoleBinding", "rbac.authorization.k8s.io/v1")
	policyTypeMeta         = typeMeta(pacv2beta1.PolicyKind, pacv2beta1.GroupVersion.String())
)

// AllowedRepository defines the allowed urls for each source type
type AllowedRepository struct {
	URL  string `yaml:"url"`
	Kind string `yaml:"kind"`
}

// Tenant represents a tenant that we generate resources for in the tenancy
// system.
type Tenant struct {
	Name                string              `yaml:"name"`
	Namespaces          []string            `yaml:"namespaces"`
	ClusterRole         string              `yaml:"clusterRole"`
	Labels              map[string]string   `yaml:"labels"`
	AllowedRepositories []AllowedRepository `yaml:"allowedRepositories"`
}

// Validate returns an error if any of the fields isn't valid
func (t Tenant) Validate() error {
	var result error

	if err := validation.IsQualifiedName(t.Name); len(err) > 0 {
		result = multierror.Append(result, fmt.Errorf("invalid tenant name: %s", err))
	}

	if len(t.Namespaces) == 0 {
		result = multierror.Append(result, errors.New("must provide at least one namespace"))
	}

	for _, allowedRepository := range t.AllowedRepositories {
		if err := validatePolicyRepoKind(allowedRepository.Kind); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}

// CreateTenants creates resources for tenants given a file for definition.
func CreateTenants(ctx context.Context, tenants []Tenant, c client.Client) error {
	resources, err := GenerateTenantResources(tenants...)
	if err != nil {
		return fmt.Errorf("failed to generate tenant output: %w", err)
	}

	for _, resource := range resources {
		err = upsert(ctx, c, resource)
		if err != nil {
			return fmt.Errorf("failed to create resource %s: %w", resource.GetName(), err)
		}
	}

	return nil
}

// upsert applies runtime objects to the cluster, if they already exist,
// patching them with type specific elements.
func upsert(ctx context.Context, kubeClient client.Client, obj client.Object) error {
	existing := runtimeObjectFromObject(obj)
	err := kubeClient.Get(ctx, client.ObjectKeyFromObject(obj), existing)
	if err != nil {
		if errs.IsNotFound(err) {
			if err := kubeClient.Create(ctx, obj); err != nil {
				return err
			}
			return nil
		}
		return err
	}

	switch to := obj.(type) {
	case *rbacv1.RoleBinding:
		existingRB := existing.(*rbacv1.RoleBinding)
		if !equality.Semantic.DeepDerivative(to.Subjects, existingRB.Subjects) ||
			!equality.Semantic.DeepDerivative(to.RoleRef, existingRB.RoleRef) ||
			!equality.Semantic.DeepDerivative(to.GetLabels(), existingRB.GetLabels()) {
			if err := kubeClient.Delete(ctx, existing); err != nil {
				return err
			}
			if err := kubeClient.Create(ctx, to); err != nil {
				return err
			}
		}
	default:
		if !equality.Semantic.DeepDerivative(obj.GetLabels(), existing.GetLabels()) {
			existing.SetLabels(obj.GetLabels())
			if err := kubeClient.Update(ctx, existing); err != nil {
				return err
			}
		}
	}
	return nil
}

func runtimeObjectFromObject(o client.Object) client.Object {
	return reflect.New(reflect.TypeOf(o).Elem()).Interface().(client.Object)
}

// ExportTenants exports all the tenants to a file.
func ExportTenants(tenants []Tenant, out io.Writer) error {
	resources, err := GenerateTenantResources(tenants...)
	if err != nil {
		return fmt.Errorf("failed to generate tenant output: %w", err)
	}

	return outputResources(out, resources)
}

func marshalOutput(out io.Writer, output runtime.Object) error {
	data, err := yaml.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	_, err = fmt.Fprintf(out, "%s", data)
	if err != nil {
		return fmt.Errorf("failed to write data: %v", err)
	}

	return nil
}

func outputResources(out io.Writer, resources []client.Object) error {
	for _, v := range resources {
		if err := marshalOutput(out, v); err != nil {
			return fmt.Errorf("failed outputting tenant: %w", err)
		}

		if _, err := out.Write([]byte("---\n")); err != nil {
			return err
		}
	}

	return nil
}

// GenerateTenantResources creates all the resources for tenants.
func GenerateTenantResources(tenants ...Tenant) ([]client.Object, error) {
	generated := []client.Object{}

	for _, tenant := range tenants {
		if err := tenant.Validate(); err != nil {
			return nil, err
		}
		// TODO: validate tenant name for creation of namespace.
		tenantLabels := tenant.Labels
		if tenantLabels == nil {
			tenantLabels = map[string]string{}
		}

		tenantLabels[tenantLabel] = tenant.Name

		for _, namespace := range tenant.Namespaces {
			generated = append(generated, newNamespace(namespace, tenantLabels))
			generated = append(generated, newServiceAccount(tenant.Name, namespace, tenantLabels))
			generated = append(generated, newRoleBinding(tenant.Name, namespace, tenant.ClusterRole, tenantLabels))
		}
		if len(tenant.AllowedRepositories) != 0 {
			policy, err := newPolicy(tenant.Name, tenant.Namespaces, tenant.AllowedRepositories, tenant.Labels)
			if err != nil {
				return nil, err
			}
			generated = append(generated, policy)
		}
	}

	return generated, nil
}

func newNamespace(name string, labels map[string]string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: namespaceTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}

func newServiceAccount(name, namespace string, labels map[string]string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: serviceAccountTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
}

func newRoleBinding(name, namespace, clusterRole string, labels map[string]string) *rbacv1.RoleBinding {
	if clusterRole == "" {
		clusterRole = "cluster-admin"
	}

	return &rbacv1.RoleBinding{
		TypeMeta: roleBindingTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clusterRole,
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "User",
				Name:     "gotk:" + namespace + ":reconciler",
			},
			{
				Kind:      "ServiceAccount",
				Name:      name,
				Namespace: namespace,
			},
		},
	}
}

func newPolicy(tenantName string, namespaces []string, allowedRepositories []AllowedRepository, labels map[string]string) (*pacv2beta1.Policy, error) {
	policyName := fmt.Sprintf("weave.policies.tenancy.%s-allowed-repositories", tenantName)
	policy := &pacv2beta1.Policy{
		TypeMeta: policyTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:   policyName,
			Labels: labels,
		},
		Spec: pacv2beta1.PolicySpec{
			ID:          policyName,
			Name:        fmt.Sprintf("%s allowed repositories", tenantName),
			Category:    "weave.categories.tenancy",
			Severity:    "high",
			Description: "Controls the allowed repositories to be used as sources",
			Targets: pacv2beta1.PolicyTargets{
				Kinds:      policyRepoKinds,
				Namespaces: namespaces,
			},
			Code: policyCode,
			Tags: []string{"tenancy"},
		},
	}
	var gitURLs, bucketEndpoints, helmURLs []string
	for _, allowedRepository := range allowedRepositories {
		switch allowedRepository.Kind {
		case policyRepoGitKind:
			gitURLs = append(gitURLs, allowedRepository.URL)
		case policyRepoBucketKind:
			bucketEndpoints = append(bucketEndpoints, allowedRepository.URL)
		case policyRepoHelmKind:
			helmURLs = append(helmURLs, allowedRepository.URL)
		}
	}

	policyParams, err := generatePolicyRepoParams(gitURLs, bucketEndpoints, helmURLs)
	if err != nil {
		return nil, err
	}
	policy.Spec.Parameters = policyParams
	return policy, nil
}

func typeMeta(kind, apiVersion string) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       kind,
		APIVersion: apiVersion,
	}
}

// Parse a raw tenant declaration, and parses it from the YAML and returns the
// extracted Tenants.
func Parse(filename string) ([]Tenant, error) {
	tenantsYAML, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read tenants file for export: %w", err)
	}

	var tenancy struct {
		Tenants []Tenant `yaml:"tenants"`
	}

	err = yaml.Unmarshal(tenantsYAML, &tenancy)
	if err != nil {
		return nil, err
	}

	return tenancy.Tenants, nil
}
