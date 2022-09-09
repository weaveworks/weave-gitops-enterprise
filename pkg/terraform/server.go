package terraform

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	tfctrl "github.com/weaveworks/tf-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/terraform"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/terraform/internal/convert"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServerOpts struct {
	logr.Logger
	ClientsFactory clustersmngr.ClientsFactory
}

type server struct {
	pb.UnimplementedTerraformServer

	log     logr.Logger
	clients clustersmngr.ClientsFactory
}

func Hydrate(ctx context.Context, mux *runtime.ServeMux, opts ServerOpts) error {
	s := NewTerraformServer(opts)

	return pb.RegisterTerraformHandlerServer(ctx, mux, s)
}

func NewTerraformServer(opts ServerOpts) pb.TerraformServer {
	return &server{
		log:     opts.Logger,
		clients: opts.ClientsFactory,
	}
}

func (s *server) ListTerraformObjects(ctx context.Context, msg *pb.ListTerraformObjectsRequest) (*pb.ListTerraformObjectsResponse, error) {
	c, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))

	if err != nil {
		return nil, fmt.Errorf("getting impersonated client: %w", err)
	}

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &tfctrl.TerraformList{}
	})

	opts := []client.ListOption{}

	if msg.Pagination != nil {
		opts = append(opts, client.Limit(msg.Pagination.PageSize))
		if msg.Pagination.PageToken != "" {
			opts = append(opts, client.Continue(msg.Pagination.PageToken))
		}
	}

	if msg.Namespace != "" {
		opts = append(opts, client.InNamespace(msg.Namespace))
	}

	listErrors := []*pb.TerraformListError{}

	if err := c.ClusteredList(ctx, clist, false, opts...); err != nil {
		var errs clustersmngr.ClusteredListError

		if !errors.As(err, &errs) {
			return nil, fmt.Errorf("terraform clustered list: %w", errs)
		}

		for _, e := range errs.Errors {

			listErrors = append(listErrors, &pb.TerraformListError{
				ClusterName: e.Cluster,
				Message:     e.Err.Error(),
			})

		}

	}

	results := []*pb.TerraformObject{}

	for clusterName, lists := range clist.Lists() {
		for _, l := range lists {
			list, ok := l.(*tfctrl.TerraformList)
			if !ok {
				continue
			}

			for _, t := range list.Items {
				o := convert.ToPBTerraformObject(clusterName, t)
				results = append(results, &o)
			}
		}
	}

	return &pb.ListTerraformObjectsResponse{
		Objects: results,
		Errors:  listErrors,
	}, nil
}
