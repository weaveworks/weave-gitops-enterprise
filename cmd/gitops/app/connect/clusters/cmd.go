package connect

import (
	"context"

	"github.com/go-logr/stdr"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/connector"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type connectOptionsFlags struct {
	RemoteClusterContext   string
	ConfigPath             string
	ServiceAccountName     string
	ClusterRoleName        string
	ClusterRoleBindingName string
	Namespace              string
}

var connectOptionsCmdFlags connectOptionsFlags

func ConnectCommand(opts *config.Options, client *adapters.HTTPClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cluster",
		Aliases: []string{"clusters"},
		Short:   "Connect cluster with remote cluster",
		Example: `
# Connect cluster
gitops connect cluster [PARAMS] <CLUSTER_NAME>
`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MinimumNArgs(1),
		RunE:          connectClusterCmdRunE(opts, client),
	}

	cmd.Flags().StringVar(&connectOptionsCmdFlags.RemoteClusterContext, "connect-context", "", "Context name of the remote cluster")
	cmd.Flags().StringVar(&connectOptionsCmdFlags.ConfigPath, "config-path", "", "kubeconfig path of hub cluster")
	cmd.Flags().StringVar(&connectOptionsCmdFlags.ServiceAccountName, "service-account", "weave-gitops-enterprise", "Service account name to be created/used")
	cmd.Flags().StringVar(&connectOptionsCmdFlags.ClusterRoleName, "cluster-role", "weave-gitops-enterprise", "Cluster role name to be created/used")
	cmd.Flags().StringVar(&connectOptionsCmdFlags.ClusterRoleBindingName, "cluster-role-binding", "weave-gitops-enterprise", "Cluster role binding name to be created/used")
	cmd.Flags().StringVar(&connectOptionsCmdFlags.Namespace, "namespace", "default", "namespace of remote cluster")

	return cmd
}

func connectClusterCmdRunE(opts *config.Options, client *adapters.HTTPClient) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clusterName := args[0]

		options := connector.ClusterConnectionOptions{
			ServiceAccountName:     connectOptionsCmdFlags.ServiceAccountName,
			ClusterRoleName:        connectOptionsCmdFlags.ClusterRoleName,
			ClusterRoleBindingName: connectOptionsCmdFlags.ClusterRoleBindingName,
			GitopsClusterName:      types.NamespacedName{Name: clusterName, Namespace: connectOptionsCmdFlags.Namespace},
			RemoteClusterContext:   connectOptionsCmdFlags.RemoteClusterContext,
			ConfigPath:             connectOptionsCmdFlags.ConfigPath,
		}

		logger := stdr.New(nil)
		ctx := log.IntoContext(context.Background(), logger)

		return connector.ConnectCluster(ctx, &options)

	}
}
