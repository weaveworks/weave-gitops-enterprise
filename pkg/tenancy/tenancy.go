package tenancy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/fluxcd/pkg/runtime/patch"
	"github.com/hashicorp/go-multierror"
	pacv2beta1 "github.com/weaveworks/policy-agent/api/v2beta1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	errs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	tenantLabel     = "toolkit.fluxcd.io/tenant"
	defaultRoleKind = "ClusterRole"
	defaultRoleName = "cluster-admin"
)

var (
	namespaceTypeMeta      = typeMeta("Namespace", "v1")
	serviceAccountTypeMeta = typeMeta("ServiceAccount", "v1")
	roleBindingTypeMeta    = typeMeta("RoleBinding", "rbac.authorization.k8s.io/v1")
	policyTypeMeta         = typeMeta(pacv2beta1.PolicyKind, pacv2beta1.GroupVersion.String())
	roleTypeMeta           = typeMeta("Role", "rbac.authorization.k8s.io/v1")
)

// ServiceAccountOptions is additional configuration for generating
// ServiceAccounts.
type ServiceAccountOptions struct {
	Name string `yaml:"name"`
}

// AllowedRepository defines the allowed urls for each source type
type AllowedRepository struct {
	URL  string `yaml:"url"`
	Kind string `yaml:"kind"`
}

// AllowedCluster defines the allowed secret names that contains cluster's kubeconfig
type AllowedCluster struct {
	KubeConfig string `yaml:"kubeConfig"`
}

// TenanTeamRBAC defines the permissions of a tenant
type TenantTeamRBAC struct {
	GroupNames []string            `yaml:"groupNames"`
	Rules      []rbacv1.PolicyRule `yaml:"rules"`
}

// TenantDeploymentRBAC defines the permissions of the tenants service account
type TenantDeploymentRBAC struct {
	Rules []rbacv1.PolicyRule `yaml:"rules"`
}

// Config represents the structure of the Tenancy file.
type Config struct {
	ServiceAccount *ServiceAccountOptions `yaml:"serviceAccount,optional"`
	Tenants        []Tenant               `yaml:"tenants"`
}

// Tenant represents a tenant that we generate resources for in the tenancy
// system.
type Tenant struct {
	Name                string                `yaml:"name"`
	Namespaces          []string              `yaml:"namespaces"`
	Labels              map[string]string     `yaml:"labels"`
	AllowedRepositories []AllowedRepository   `yaml:"allowedRepositories"`
	AllowedClusters     []AllowedCluster      `yaml:"allowedClusters"`
	TeamRBAC            *TenantTeamRBAC       `yaml:"teamRBAC,omitempty"`
	DeploymentRBAC      *TenantDeploymentRBAC `yaml:"deploymentRBAC,omitempty"`
	RoleName            string                `yaml:"roleName"`
	RoleKind            string                `yaml:"roleKind"`
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

	if t.TeamRBAC != nil {
		if len(t.TeamRBAC.GroupNames) == 0 || len(t.TeamRBAC.Rules) == 0 {
			result = multierror.Append(result, errors.New("must provide group names and team rules in team RBAC"))
		}
	}

	if t.DeploymentRBAC != nil {
		if len(t.DeploymentRBAC.Rules) == 0 {
			result = multierror.Append(result, errors.New("must provide rules in deployment RBAC"))
		}
	}

	return result
}

// ApplyTenants applies resources for state defined in each tenant given a file for definition.
func ApplyTenants(ctx context.Context, config *Config, c client.Client, prune bool, out io.Writer) error {
	newResources, err := GenerateTenantResources(config)
	if err != nil {
		return fmt.Errorf("failed to generate tenant output: %w", err)
	}
	existingResources, err := getCurrentResources(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to check current tenant resources: %w", err)
	}
	err = cleanResources(ctx, c, newResources, existingResources, prune, out)
	if err != nil {
		return fmt.Errorf("failed to clean up policies resources: %w", err)
	}
	for _, resource := range newResources {
		err := upsert(ctx, c, resource, out)
		if err != nil {
			return fmt.Errorf("failed to create resource %s: %w", resource.GetName(), err)
		}
	}

	return nil
}

// getCurrentResources checks current tenant resources that exists on a cluster
func getCurrentResources(ctx context.Context, kubeClient client.Client) ([]client.Object, error) {
	var resources []client.Object
	existingTypeMetas := []metav1.TypeMeta{
		namespaceTypeMeta,
		roleBindingTypeMeta,
		serviceAccountTypeMeta,
		roleTypeMeta,
		policyTypeMeta,
	}

	opts := []client.ListOption{client.HasLabels{tenantLabel}}

	for _, existingTypeMeta := range existingTypeMetas {
		existing := unstructured.UnstructuredList{}
		existing.SetGroupVersionKind(existingTypeMeta.GroupVersionKind())

		err := kubeClient.List(ctx, &existing, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to list tenant resources: %w", err)
		}
		for i := range existing.Items {
			resources = append(resources, &existing.Items[i])
		}
	}

	return resources, nil
}

// cleanResources cleanup resources that should be pruned when certain configuration is removed
func cleanResources(ctx context.Context, kubeClient client.Client, newResources, existingResources []client.Object, prune bool, out io.Writer) error {
	resourcesToDelete := getResourcesToDelete(newResources, existingResources)
	for i := range resourcesToDelete {
		resourceToDeleteID := getObjectID(resourcesToDelete[i])
		if prune {
			err := kubeClient.Delete(ctx, resourcesToDelete[i])
			if err != nil {
				fmt.Fprintf(out, "failed to clean up tenant resource %s: %s", resourceToDeleteID, err)
				continue
			}
			fmt.Fprintf(out, "%s deleted\n", resourceToDeleteID)
		} else {
			fmt.Fprintf(out, "%s no longer defined as part of a tenant use --prune to remove\n", resourceToDeleteID)
		}
	}
	return nil
}

// upsert applies runtime objects to the cluster, if they already exist,
// patching them with type specific elements.
func upsert(ctx context.Context, kubeClient client.Client, obj client.Object, out io.Writer) error {
	existing := runtimeObjectFromObject(obj)
	objectID := getObjectID(obj)

	err := kubeClient.Get(ctx, client.ObjectKeyFromObject(obj), existing)
	if err != nil {
		if errs.IsNotFound(err) {
			if err := kubeClient.Create(ctx, obj); err != nil {
				return err
			}

			fmt.Fprintf(out, "%s created\n", objectID)

			return nil
		}

		return err
	}

	patchHelper, err := patch.NewHelper(existing, kubeClient)
	if err != nil {
		return err
	}

	switch to := obj.(type) {
	case *rbacv1.RoleBinding:
		existingRB := existing.(*rbacv1.RoleBinding)
		if !equality.Semantic.DeepEqual(to.Subjects, existingRB.Subjects) ||
			!equality.Semantic.DeepDerivative(to.RoleRef, existingRB.RoleRef) ||
			!equality.Semantic.DeepDerivative(to.GetLabels(), existingRB.GetLabels()) {
			if err := kubeClient.Delete(ctx, existing); err != nil {
				return err
			}

			if err := kubeClient.Create(ctx, to); err != nil {
				return err
			}

			fmt.Fprintf(out, "%s recreated\n", objectID)
		}
	case *rbacv1.Role:
		existingRole := existing.(*rbacv1.Role)

		var changed bool

		if !equality.Semantic.DeepDerivative(to.GetLabels(), existingRole.GetLabels()) {
			existingRole.SetLabels(to.GetLabels())

			changed = true
		}

		if !equality.Semantic.DeepDerivative(to.Rules, existingRole.Rules) {
			existingRole.Rules = to.Rules
			changed = true
		}

		if changed {
			if err := kubeClient.Update(ctx, existing); err != nil {
				return fmt.Errorf("failed to update existing role: %w", err)
			}

			fmt.Fprintf(out, "%s updated\n", objectID)
		}
	case *pacv2beta1.Policy:
		existingPolicy := existing.(*pacv2beta1.Policy)

		var changed bool

		if !equality.Semantic.DeepDerivative(to.GetLabels(), existingPolicy.GetLabels()) {
			existingPolicy.SetLabels(to.GetLabels())

			changed = true
		}

		if !equality.Semantic.DeepDerivative(to.Spec, existingPolicy.Spec) {
			existingPolicy.Spec = to.Spec
			changed = true
		}

		if changed {
			if err := patchHelper.Patch(ctx, existing); err != nil {
				return fmt.Errorf("failed to patch existing policy: %w", err)
			}

			fmt.Fprintf(out, "%s updated\n", objectID)
		}
	default:
		if !equality.Semantic.DeepDerivative(obj.GetLabels(), existing.GetLabels()) {
			existing.SetLabels(obj.GetLabels())

			if err := kubeClient.Update(ctx, existing); err != nil {
				return err
			}

			fmt.Fprintf(out, "%s updated\n", objectID)
		}
	}

	return nil
}

// ExportTenants exports all the tenants to a file.
func ExportTenants(config *Config, out io.Writer) error {
	resources, err := GenerateTenantResources(config)
	if err != nil {
		return fmt.Errorf("failed to generate tenant output: %w", err)
	}

	return outputResources(out, resources)
}

// generateTenantResource create resources for a tenant
func generateTenantResource(tenant Tenant, serviceAccount *ServiceAccountOptions) ([]client.Object, error) {
	generated := []client.Object{}
	if err := tenant.Validate(); err != nil {
		return nil, err
	}
	// TODO: validate tenant name for creation of namespace.
	tenantLabels := tenant.Labels
	if tenantLabels == nil {
		tenantLabels = map[string]string{}
	}

	tenantLabels[tenantLabel] = tenant.Name
	serviceAccountName := tenant.Name
	isGlobalServiceAccount := false
	if serviceAccount != nil {
		serviceAccountName = serviceAccount.Name
		isGlobalServiceAccount = true
	}

	for _, namespace := range tenant.Namespaces {
		generated = append(generated, newNamespace(namespace, tenantLabels))
		if !isGlobalServiceAccount {
			generated = append(generated, newServiceAccount(serviceAccountName, namespace, tenantLabels))
		}
		if tenant.DeploymentRBAC != nil {
			generated = append(generated, newServiceAccountRole(tenant.Name, namespace, tenantLabels, tenant.DeploymentRBAC.Rules))
			generated = append(generated, newServiceAccountRoleBinding(tenant.Name, namespace, serviceAccountName, tenantLabels))
		} else {
			generated = append(generated, newDefaultServiceAccountRoleBinding(
				tenant.Name,
				namespace,
				serviceAccountName,
				tenant.RoleKind,
				tenant.RoleName,
				tenantLabels,
			))
		}

		if tenant.TeamRBAC != nil {
			generated = append(generated, newTeamRole(tenant.Name, namespace, tenantLabels, tenant.TeamRBAC.Rules))
			generated = append(generated, newTeamRoleBinding(tenant.Name, namespace, tenant.TeamRBAC.GroupNames, tenantLabels))
		}
	}

	policy, err := newAllowedApplicationDeployPolicy(tenant.Name, serviceAccountName, tenant.Namespaces, tenantLabels)
	if err != nil {
		return nil, err
	}
	generated = append(generated, policy)

	if len(tenant.AllowedRepositories) != 0 {
		policy, err := newAllowedRepositoriesPolicy(tenant.Name, tenant.Namespaces, tenant.AllowedRepositories, tenantLabels)
		if err != nil {
			return nil, err
		}
		generated = append(generated, policy)
	}

	if len(tenant.AllowedClusters) != 0 {
		policy, err := newAllowedClustersPolicy(tenant.Name, tenant.Namespaces, tenant.AllowedClusters, tenantLabels)
		if err != nil {
			return nil, err
		}
		generated = append(generated, policy)
	}

	return generated, nil

}

// GenerateTenantResources creates all the resources for tenants.
func GenerateTenantResources(config *Config) ([]client.Object, error) {
	generated := []client.Object{}

	for _, tenant := range config.Tenants {
		tenantGenerated, err := generateTenantResource(tenant, config.ServiceAccount)
		if err != nil {
			return nil, err
		}

		generated = append(generated, tenantGenerated...)
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

func newServiceAccountRoleBinding(tenantName, namespace, serviceAccountName string, labels map[string]string) *rbacv1.RoleBinding {
	name := fmt.Sprintf("%s-service-account", tenantName)
	return newRoleBinding(
		name,
		namespace,
		"Role",
		name,
		[]rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccountName,
				Namespace: namespace,
			},
		},
		labels,
	)
}

func newDefaultServiceAccountRoleBinding(tenantName, namespace, serviceAccountName, roleKind, roleName string, labels map[string]string) *rbacv1.RoleBinding {
	if roleKind == "" {
		roleKind = defaultRoleKind
	}

	if roleName == "" {
		roleName = defaultRoleName
	}

	return newRoleBinding(
		fmt.Sprintf("%s-service-account", tenantName),
		namespace,
		roleKind,
		roleName,
		[]rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccountName,
				Namespace: namespace,
			},
		},
		labels,
	)
}

func newTeamRoleBinding(tenantName, namespace string, groupNames []string, labels map[string]string) *rbacv1.RoleBinding {
	name := fmt.Sprintf("%s-team", tenantName)
	subjects := []rbacv1.Subject{}
	for _, groupName := range groupNames {
		subjects = append(subjects, rbacv1.Subject{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Group",
			Name:     groupName,
		})
	}
	return newRoleBinding(
		name,
		namespace,
		"Role",
		name,
		subjects,
		labels,
	)
}

func newRoleBinding(name, namespace, roleKind, roleName string, subjects []rbacv1.Subject, labels map[string]string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		TypeMeta: roleBindingTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     roleKind,
			Name:     roleName,
		},
		Subjects: subjects,
	}
}

func newTeamRole(name, namespace string, labels map[string]string, rules []rbacv1.PolicyRule) *rbacv1.Role {
	return newRole(fmt.Sprintf("%s-team", name), namespace, labels, rules)
}

func newServiceAccountRole(name, namespace string, labels map[string]string, rules []rbacv1.PolicyRule) *rbacv1.Role {
	return newRole(fmt.Sprintf("%s-service-account", name), namespace, labels, rules)
}

func newRole(name, namespace string, labels map[string]string, rules []rbacv1.PolicyRule) *rbacv1.Role {
	return &rbacv1.Role{
		TypeMeta: roleTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Rules: rules,
	}
}

func newAllowedRepositoriesPolicy(tenantName string, namespaces []string, allowedRepositories []AllowedRepository, labels map[string]string) (*pacv2beta1.Policy, error) {
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
			Standards:   []pacv2beta1.PolicyStandard{},
			Targets: pacv2beta1.PolicyTargets{
				Labels:     []map[string]string{},
				Kinds:      policyRepoKinds,
				Namespaces: namespaces,
			},
			Code: repoPolicyCode,
			Tags: []string{"tenancy"},
		},
	}

	var gitURLs, bucketEndpoints, helmURLs, ociURLs []string

	for _, allowedRepository := range allowedRepositories {
		switch allowedRepository.Kind {
		case policyRepoGitKind:
			gitURLs = append(gitURLs, allowedRepository.URL)
		case policyRepoBucketKind:
			bucketEndpoints = append(bucketEndpoints, allowedRepository.URL)
		case policyRepoHelmKind:
			helmURLs = append(helmURLs, allowedRepository.URL)
		case policyRepoOCIKind:
			ociURLs = append(ociURLs, allowedRepository.URL)
		}
	}

	policyParams, err := generatePolicyRepoParams(gitURLs, bucketEndpoints, helmURLs, ociURLs)
	if err != nil {
		return nil, err
	}

	policy.Spec.Parameters = policyParams

	return policy, nil
}

func newAllowedClustersPolicy(tenantName string, namespaces []string, allowedClusters []AllowedCluster, labels map[string]string) (*pacv2beta1.Policy, error) {
	policyName := fmt.Sprintf("weave.policies.tenancy.%s-allowed-clusters", tenantName)

	var clusterSecrets []string

	for _, allowedCluster := range allowedClusters {
		clusterSecrets = append(clusterSecrets, allowedCluster.KubeConfig)
	}

	clusterSecretstBytes, err := json.Marshal(clusterSecrets)
	if err != nil {
		return nil, fmt.Errorf("error while setting policy parameters values: %w", err)
	}

	policy := &pacv2beta1.Policy{
		TypeMeta: policyTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:   policyName,
			Labels: labels,
		},
		Spec: pacv2beta1.PolicySpec{
			ID:          policyName,
			Name:        fmt.Sprintf("%s allowed clusters", tenantName),
			Category:    "weave.categories.tenancy",
			Severity:    "high",
			Description: "Controls the allowed clusters to be added",
			Standards:   []pacv2beta1.PolicyStandard{},
			Targets: pacv2beta1.PolicyTargets{
				Labels:     []map[string]string{},
				Kinds:      []string{policyClustersKind, policyKustomizationKind},
				Namespaces: namespaces,
			},
			Code: clusterPolicyCode,
			Tags: []string{"tenancy"},
			Parameters: []pacv2beta1.PolicyParameters{
				{
					Name: "cluster_secrets",
					Type: "array",
					Value: &apiextensionsv1.JSON{
						Raw: clusterSecretstBytes,
					},
				},
			},
		},
	}

	return policy, nil
}

func newAllowedApplicationDeployPolicy(tenantName, serviceAccountName string, namespaces []string, labels map[string]string) (*pacv2beta1.Policy, error) {
	policyName := fmt.Sprintf("weave.policies.tenancy.%s-allowed-application-deploy", tenantName)

	namespacesBytes, err := json.Marshal(namespaces)
	if err != nil {
		return nil, fmt.Errorf("error while setting policy parameters values: %w", err)
	}

	serviceAccountNameBytes, err := json.Marshal(serviceAccountName)
	if err != nil {
		return nil, fmt.Errorf("error while setting policy parameters values: %w", err)
	}

	policy := &pacv2beta1.Policy{
		TypeMeta: policyTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:   policyName,
			Labels: labels,
		},
		Spec: pacv2beta1.PolicySpec{
			ID:          policyName,
			Name:        fmt.Sprintf("%s allowed application deploy", tenantName),
			Category:    "weave.categories.tenancy",
			Severity:    "high",
			Description: "Determines which helm release and kustomization can be used in a tenant",
			Standards:   []pacv2beta1.PolicyStandard{},
			Targets: pacv2beta1.PolicyTargets{
				Labels:     []map[string]string{},
				Kinds:      []string{policyHelmReleaseKind, policyKustomizationKind},
				Namespaces: namespaces,
			},
			Code: applicationPolicyCode,
			Tags: []string{"tenancy"},
			Parameters: []pacv2beta1.PolicyParameters{
				{
					Name: "namespaces",
					Type: "array",
					Value: &apiextensionsv1.JSON{
						Raw: namespacesBytes,
					},
				},
				{
					Name: "service_account_name",
					Type: "string",
					Value: &apiextensionsv1.JSON{
						Raw: serviceAccountNameBytes,
					},
				},
			},
		},
	}

	return policy, nil
}

// Parse a raw tenant declaration, and parses it from the YAML and returns the
// extracted Tenants.
func Parse(filename string) (*Config, error) {
	tenantsYAML, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read tenants file for export: %w", err)
	}

	var tenancy struct {
		ServiceAccount *ServiceAccountOptions `yaml:"serviceAccount,optional"`
		Tenants        []Tenant               `yaml:"tenants"`
	}

	err = yaml.Unmarshal(tenantsYAML, &tenancy)
	if err != nil {
		return nil, err
	}

	return &Config{Tenants: tenancy.Tenants, ServiceAccount: tenancy.ServiceAccount}, nil
}
