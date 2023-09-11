package connect

import (
	"github.com/go-logr/stdr"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/connector"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type connectOptionsFlags struct {
	RemoteClusterContext   string
	ServiceAccountName     string
	ClusterRoleBindingName string
	Namespace              string
	Verbose                int16
}

var connectOptionsCmdFlags connectOptionsFlags

func ConnectCommand(opts *config.Options) *cobra.Command {
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
		RunE:          connectClusterCmdRunE(opts),
	}

	cmd.Flags().StringVar(&connectOptionsCmdFlags.RemoteClusterContext, "connect-context", "", "Context name of the remote cluster")
	cmd.Flags().StringVar(&connectOptionsCmdFlags.ServiceAccountName, "service-account", "weave-gitops-enterprise", "Service account name to be created/used")
	cmd.Flags().StringVar(&connectOptionsCmdFlags.ClusterRoleBindingName, "cluster-role-binding", "weave-gitops-enterprise", "Cluster role binding name to be created/used")
	cmd.Flags().StringVarP(&connectOptionsCmdFlags.Namespace, "namespace", "n", "default", "Namespace of remote cluster")
	cmd.Flags().Int16VarP(&connectOptionsCmdFlags.Verbose, "verbose", "v", 10, "Verbose level of logs")

	return cmd
}

func connectClusterCmdRunE(opts *config.Options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clusterName := args[0]

		options := connector.ClusterConnectionOptions{
			ServiceAccountName:     connectOptionsCmdFlags.ServiceAccountName,
			ClusterRoleBindingName: connectOptionsCmdFlags.ClusterRoleBindingName,
			GitopsClusterName:      types.NamespacedName{Name: clusterName, Namespace: connectOptionsCmdFlags.Namespace},
			RemoteClusterContext:   connectOptionsCmdFlags.RemoteClusterContext,
			ConfigPath:             opts.Kubeconfig,
		}

		logger := stdr.New(nil)
		ctx := log.IntoContext(cmd.Context(), logger.V(int(connectOptionsCmdFlags.Verbose)))

		return connector.ConnectCluster(ctx, &options)

	}
}
