package server

import (
	"context"
	"errors"
	"fmt"

	ctrl "github.com/weaveworks/gitopssets-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/gitopssets"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/gitopssets/internal/convert"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *server) ListGitOpsSets(ctx context.Context, msg *pb.ListGitOpsSetsRequest) (*pb.ListGitOpsSetsResponse, error) {
	c, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))

	if err != nil {
		return nil, fmt.Errorf("getting impersonated client: %w", err)
	}

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &ctrl.GitOpsSetList{}
	})

	opts := []client.ListOption{}

	if msg.Namespace != "" {
		opts = append(opts, client.InNamespace(msg.Namespace))
	}

	listErrors := []*pb.GitOpsSetListError{}

	if err := c.ClusteredList(ctx, clist, false, opts...); err != nil {
		var errs clustersmngr.ClusteredListError

		if !errors.As(err, &errs) {
			return nil, fmt.Errorf("converting to ClusteredListError: %w", errs)
		}

		for _, e := range errs.Errors {
			if apimeta.IsNoMatchError(e.Err) {
				// Skip reporting an error if a leaf cluster does not have the tf-controller CRD installed.
				// It is valid for leaf clusters to not have tf installed.
				s.log.Info("tf-controller crd not present on cluster, skipping error", "cluster", e.Cluster)
				continue
			}

			listErrors = append(listErrors, &pb.GitOpsSetListError{
				ClusterName: e.Cluster,
				Message:     e.Err.Error(),
			})

		}

	}

	gitopssets := []*pb.GitOpsSet{}

	for clusterName, lists := range clist.Lists() {
		for _, l := range lists {
			list, ok := l.(*ctrl.GitOpsSetList)
			if !ok {
				continue
			}

			for _, gs := range list.Items {
				gitopssets = append(gitopssets, convert.GitOpsToProto(clusterName,gs))
			}
		}
	}

	return &pb.ListGitOpsSetsResponse{
		Gitopssets: gitopssets,
		Errors:  listErrors,
	}, nil
}