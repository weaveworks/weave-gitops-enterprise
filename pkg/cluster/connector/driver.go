package connector

// ConnectionOptions describes the options for connecting a cluster.
// type ConnectionOptions struct {
// 	HubContext         string
// 	SpokeContext       string
// 	ServiceAccountName string
// }

// func ConnectSpokeToHubCluster(ctx context.Context, pathOpts *clientcmd.PathOptions, options ConnectionOptions) error {
// 	spokeConfig, _ := ConfigForContext(pathOpts, options.SpokeContext)
// 	hubConfig, _ := ConfigForContext(pathOpts, options.HubContext)

// 	spokeClient := kubernetes.NewForConfig(spokeConfig)

// 	if err := ReconcileServiceAccount(ctx, spokeClient.CoreV1(), options.ServiceAccountName); err != nil {
// 		return err
// 	}

// 	if err := ReconcileClusterRoleBindingForServiceAccount(ctx, spokeClient.CoreV1(), options.ServiceAccountName); err != nil {
// 		return err
// 	}

// 	if err := ReconcileServiceAccountTokenSecret(ctx, spokeClient.CoreV1(), options.ServiceAccountName); err != nil {
// 		return err
// 	}

// 	token, err := WaitForServiceAccountSecretToken(ctx, spokeClient.CoreV1(), options.ServiceAccountName)
// 	if err != nil {
// 		return err
// 	}

// 	if err := ReconcileClusterSecret(ctx, hubClient.CoreV1(), options.SpokeContext, token); err != nil {
// 		return err
// 	}
// }
