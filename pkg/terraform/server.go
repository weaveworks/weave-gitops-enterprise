package terraform

import (
	"bytes"
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
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServerOpts struct {
	logr.Logger
	ClientsFactory clustersmngr.ClientsFactory
	Scheme         *k8sruntime.Scheme
}

type server struct {
	pb.UnimplementedTerraformServer

	log     logr.Logger
	clients clustersmngr.ClientsFactory
	scheme  *k8sruntime.Scheme
}

func Hydrate(ctx context.Context, mux *runtime.ServeMux, opts ServerOpts) error {
	s := NewTerraformServer(opts)

	return pb.RegisterTerraformHandlerServer(ctx, mux, s)
}

func NewTerraformServer(opts ServerOpts) pb.TerraformServer {
	return &server{
		log:     opts.Logger,
		clients: opts.ClientsFactory,
		scheme:  opts.Scheme,
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
				o := convert.ToPBTerraformObject(clusterName, &t)
				results = append(results, &o)
			}
		}
	}

	return &pb.ListTerraformObjectsResponse{
		Objects: results,
		Errors:  listErrors,
	}, nil
}

func (s *server) GetTerraformObject(ctx context.Context, msg *pb.GetTerraformObjectRequest) (*pb.GetTerraformObjectResponse, error) {
	c, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("getting impersonated client: %w", err)
	}

	n := types.NamespacedName{Name: msg.Name, Namespace: msg.Namespace}

	result := &tfctrl.Terraform{}
	if err := c.Get(ctx, msg.ClusterName, n, result); err != nil {
		return nil, fmt.Errorf("getting object with name %s in namespace %s: %w", msg.Name, msg.Namespace, err)
	}

	yaml, err := serializeObj(s.scheme, result)
	if err != nil {
		return nil, fmt.Errorf("serializing yaml: %w", err)
	}

	obj := convert.ToPBTerraformObject(msg.ClusterName, result)

	return &pb.GetTerraformObjectResponse{
		Object: &obj,
		Yaml:   string(yaml),
	}, nil
}

func serializeObj(scheme *k8sruntime.Scheme, obj client.Object) ([]byte, error) {

	obj.GetObjectKind().SetGroupVersionKind(tfctrl.GroupVersion.WithKind(tfctrl.TerraformKind))

	serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme, scheme, json.SerializerOptions{
		Pretty: true,
		Yaml:   true,
		Strict: true,
	})

	buf := bytes.NewBufferString("")

	if err := serializer.Encode(obj, buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
