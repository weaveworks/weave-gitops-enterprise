package disconnect

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/connector"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/core/logger"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type disconnectOptionsFlags struct {
	RemoteClusterContext   string
	ServiceAccountName     string
	ClusterRoleBindingName string
	Namespace              string
	Debug                  string
}

var disconnectOptionsCmdFlags disconnectOptionsFlags

func DisconnectCommand(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cluster",
		Aliases: []string{"clusters"},
		Short:   "Disconnect cluster from remote cluster by deleting resources",
		Example: `
# Disconnect cluster
gitops disconnect cluster [PARAMS] <CLUSTER_NAME>
`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MinimumNArgs(1),
		RunE:          disconnectClusterCmdRunE(opts),
	}

	cmd.Flags().StringVar(&disconnectOptionsCmdFlags.RemoteClusterContext, "connect-context", "", "Context name of the remote cluster")
	cmd.Flags().StringVar(&disconnectOptionsCmdFlags.ServiceAccountName, "service-account", "weave-gitops-enterprise", "Service account name to be created/used")
	cmd.Flags().StringVar(&disconnectOptionsCmdFlags.ClusterRoleBindingName, "cluster-role-binding", "weave-gitops-enterprise", "Cluster role binding name to be created/used")
	cmd.Flags().StringVarP(&disconnectOptionsCmdFlags.Namespace, "namespace", "n", "default", "Namespace of remote cluster")
	cmd.Flags().StringVarP(&disconnectOptionsCmdFlags.Debug, "debug", "d", "INFO", "Verbose level of logs")

	return cmd
}

func disconnectClusterCmdRunE(opts *config.Options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clusterName := args[0]

		options := connector.ClusterConnectionOptions{
			GitopsClusterName:      types.NamespacedName{Name: clusterName, Namespace: disconnectOptionsCmdFlags.Namespace},
			ServiceAccountName:     disconnectOptionsCmdFlags.ServiceAccountName,
			ClusterRoleBindingName: disconnectOptionsCmdFlags.ClusterRoleBindingName,
			RemoteClusterContext:   disconnectOptionsCmdFlags.RemoteClusterContext,
			ConfigPath:             opts.Kubeconfig,
		}

		newLogger, _ := logger.New(disconnectOptionsCmdFlags.Debug, false)
		ctx := log.IntoContext(cmd.Context(), newLogger)

		return connector.DisconnectCluster(ctx, &options)

	}
}
