package tenants

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	pacv2beta1 "github.com/weaveworks/policy-agent/api/v2beta1"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/tenancy"
	"github.com/weaveworks/weave-gitops/pkg/kube"

	apiextentionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	policyCRDName = "policies.pac.weave.works"
)

type tenantCommandFlags struct {
	name                string
	namespaces          []string
	fromFile            string
	export              bool
	skipPreFlightChecks bool
}

var flags tenantCommandFlags

var TenantsCommand = &cobra.Command{
	Use:   "tenants",
	Short: "Create or update tenant resources",
	Example: `
	  # Create a tenant using name and namespace flags
	  gitops create tenants --name test-tenant1 --namespace test-ns1 --namespace test-ns2

	  # Create tenants from a file
	  gitops create tenants --from-file tenants.yaml

	  # Export tenant resources to a file
	  gitops create tenants --from-file tenants.yaml --export > tenants.yaml

	  # Export tenant resources to stdout
	  gitops create tenants --from-file tenants.yaml --export
	`,
	RunE: createTenantsCmdRunE(),
}

func init() {
	TenantsCommand.Flags().StringVar(&flags.name, "name", "", "the name of the tenant to be created")
	TenantsCommand.Flags().StringSliceVar(&flags.namespaces, "namespace", []string{}, "a list of namespaces for the tenant")
	TenantsCommand.Flags().StringVar(&flags.fromFile, "from-file", "", "the file containing the tenant declarations")
	TenantsCommand.Flags().BoolVar(&flags.export, "export", false, "export in YAML format to stdout")
	TenantsCommand.Flags().BoolVar(&flags.skipPreFlightChecks, "skip-preflight-checks", false, "skip preflight checks before creating resources in cluster")

	cobra.CheckErr(TenantsCommand.MarkFlagRequired("from-file"))
}

func createTenantsCmdRunE() func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		tenants := []tenancy.Tenant{}

		if flags.fromFile != "" {
			parsedTenants, err := tenancy.Parse(flags.fromFile)
			if err != nil {
				return fmt.Errorf("failed to parse tenants file %s for export: %w", flags.fromFile, err)
			}

			tenants = append(tenants, parsedTenants...)
		}

		if flags.name != "" {
			tenants = append(tenants, tenancy.Tenant{
				Name:       flags.name,
				Namespaces: flags.namespaces,
			})
		}

		if flags.export {
			err := tenancy.ExportTenants(tenants, os.Stdout)
			if err != nil {
				return err
			}

			return nil
		}

		ctx := context.Background()

		config, contextName, err := kube.RestConfig()
		if err != nil {
			return fmt.Errorf("could not create default config: %w", err)
		}

		kubeClient, err := kube.NewKubeHTTPClientWithConfig(config, contextName, pacv2beta1.AddToScheme)
		if err != nil {
			return fmt.Errorf("failed to create kube client: %w", err)
		}

		if !flags.skipPreFlightChecks {
			err := preFlightCheck(ctx, tenants, kubeClient)
			if err != nil {
				return fmt.Errorf("preflight check failed with error: %w", err)
			}
		}

		err = tenancy.CreateTenants(ctx, tenants, kubeClient)
		if err != nil {
			return err
		}

		return nil
	}
}

func preFlightCheck(ctx context.Context, tenants []tenancy.Tenant, kubeClient *kube.KubeHTTP) error {
	var hasPolicy bool

	for _, tenant := range tenants {
		if len(tenant.AllowedRepositories) > 0 {
			hasPolicy = true
			break
		}
	}

	if !hasPolicy {
		return nil
	}

	crd := &apiextentionsv1.CustomResourceDefinition{}

	err := kubeClient.Get(ctx, client.ObjectKey{Name: policyCRDName}, crd)
	if err != nil {
		return err
	}

	return nil
}
