package terraform

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hashicorp/go-multierror"
	tfctrl "github.com/weaveworks/tf-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/terraform"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/terraform/internal/adapter"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/terraform/internal/convert"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/fluxsync"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServerOpts struct {
	logr.Logger
	ClientsFactory clustersmngr.ClustersManager
	Scheme         *k8sruntime.Scheme
}

type server struct {
	pb.UnimplementedTerraformServer

	log     logr.Logger
	clients clustersmngr.ClustersManager
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

	if err := c.ClusteredList(ctx, clist, true, opts...); err != nil {
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
		Type:   result.GetObjectKind().GroupVersionKind().Kind,
	}, nil
}

func (s *server) GetTerraformObjectPlan(ctx context.Context, msg *pb.GetTerraformObjectPlanRequest) (*pb.GetTerraformObjectPlanResponse, error) {
	c, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("getting impersonated client: %w", err)
	}

	n := types.NamespacedName{Name: msg.Name, Namespace: msg.Namespace}

	obj := &tfctrl.Terraform{}
	if err := c.Get(ctx, msg.ClusterName, n, obj); err != nil {
		return nil, fmt.Errorf("getting object with name %s in namespace %s: %w", msg.Name, msg.Namespace, err)
	}

	result := &pb.GetTerraformObjectPlanResponse{
		EnablePlanViewing: obj.Spec.StoreReadablePlan == "human",
	}

	if obj.Spec.StoreReadablePlan != "human" {
		result.Error = "no human-readable plan found"
		return result, nil
	}

	planKey := types.NamespacedName{
		Name:      fmt.Sprintf("tfplan-%s-%s", obj.WorkspaceName(), msg.Name),
		Namespace: msg.Namespace,
	}

	var tfplanCM corev1.ConfigMap
	if err := c.Get(ctx, msg.ClusterName, planKey, &tfplanCM); err != nil {
		result.Error = fmt.Sprintf("getting terraform plan: %s", err.Error())
		return result, nil
	}

	result.Plan = tfplanCM.Data["tfplan"]

	return result, nil
}

func (s *server) SyncTerraformObjects(ctx context.Context, msg *pb.SyncTerraformObjectsRequest) (*pb.SyncTerraformObjectsResponse, error) {
	clustersClient, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("getting impersonated client: %w", err)
	}

	respErrors := multierror.Error{}

	for _, sync := range msg.Objects {
		c, err := clustersClient.Scoped(sync.ClusterName)
		if err != nil {
			respErrors = *multierror.Append(fmt.Errorf("getting scoped client: %w", err), respErrors.Errors...)
			continue
		}

		key := client.ObjectKey{
			Name:      sync.Name,
			Namespace: sync.Namespace,
		}

		obj := adapter.TerraformObjectAdapter{Terraform: &tfctrl.Terraform{}}

		if err := c.Get(ctx, key, obj.AsClientObject()); err != nil {
			respErrors = *multierror.Append(fmt.Errorf("getting object %s in namespace %s: %w", sync.Name, sync.Namespace, err), respErrors.Errors...)
			continue
		}

		if err := fluxsync.RequestReconciliation(ctx, c, key, obj.GroupVersionKind()); err != nil {
			respErrors = *multierror.Append(fmt.Errorf("requesting reconciliation: %w", err), respErrors.Errors...)
			continue
		}

		if err := fluxsync.WaitForSync(ctx, c, key, obj); err != nil {
			respErrors = *multierror.Append(fmt.Errorf("waiting for sync: %w", err), respErrors.Errors...)
			continue
		}
	}

	return &pb.SyncTerraformObjectsResponse{Success: true}, respErrors.ErrorOrNil()
}

func (s *server) ToggleSuspendTerraformObjects(ctx context.Context, msg *pb.ToggleSuspendTerraformObjectsRequest) (*pb.ToggleSuspendTerraformObjectsResponse, error) {
	clustersClient, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("getting impersonated client: %w", err)
	}

	respErrors := multierror.Error{}

	for _, obj := range msg.Objects {
		c, err := clustersClient.Scoped(obj.ClusterName)
		if err != nil {
			respErrors = *multierror.Append(fmt.Errorf("getting scoped client: %w", err), respErrors.Errors...)
			continue
		}

		key := client.ObjectKey{
			Name:      obj.Name,
			Namespace: obj.Namespace,
		}

		obj := &tfctrl.Terraform{}

		if err := c.Get(ctx, key, obj); err != nil {
			respErrors = *multierror.Append(fmt.Errorf("getting object %s in namespace %s: %w", obj.Name, obj.Namespace, err), respErrors.Errors...)
			continue
		}

		patch := client.MergeFrom(obj.DeepCopy())

		obj.Spec.Suspend = msg.Suspend

		if err := c.Patch(ctx, obj, patch); err != nil {
			respErrors = *multierror.Append(fmt.Errorf("patching object: %w", err), respErrors.Errors...)
			continue
		}
	}

	return &pb.ToggleSuspendTerraformObjectsResponse{}, respErrors.ErrorOrNil()
}

func (s *server) ReplanTerraformObject(ctx context.Context, msg *pb.ReplanTerraformObjectRequest) (*pb.ReplanTerraformObjectResponse, error) {
	clustersClient, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("getting impersonated client: %w", err)
	}

	if msg.Name == "" || msg.Namespace == "" {
		return nil, fmt.Errorf("object name or namespace is empty")
	}

	c, err := clustersClient.Scoped(msg.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("getting scoped client for cluster %s: %w", msg.ClusterName, err)
	}

	key := client.ObjectKey{
		Name:      msg.Name,
		Namespace: msg.Namespace,
	}

	if err := replan(ctx, c, key); err != nil {
		return nil, err
	}

	obj := adapter.TerraformObjectAdapter{Terraform: &tfctrl.Terraform{}}

	if err := c.Get(ctx, key, obj.AsClientObject()); err != nil {
		return nil, fmt.Errorf("getting object %s in namespace %s: %w", msg.Name, msg.Namespace, err)
	}

	if err := fluxsync.RequestReconciliation(ctx, c, key, obj.GroupVersionKind()); err != nil {
		return nil, fmt.Errorf("requesting reconciliation: %w", err)
	}

	return &pb.ReplanTerraformObjectResponse{ReplanRequested: true}, nil
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

func replan(ctx context.Context, kubeClient client.Client, namespacedName types.NamespacedName) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		terraform := &tfctrl.Terraform{}
		if err := kubeClient.Get(ctx, namespacedName, terraform); err != nil {
			return err
		}

		patch := client.MergeFrom(terraform.DeepCopy())
		// clear the pending plan
		apimeta.SetStatusCondition(&terraform.Status.Conditions, metav1.Condition{
			Type:    meta.ReadyCondition,
			Status:  metav1.ConditionFalse,
			Reason:  "ReplanRequested",
			Message: "Replan requested",
		})
		terraform.Status.Plan.Pending = ""
		terraform.Status.LastPlannedRevision = ""
		terraform.Status.LastAttemptedRevision = ""
		return kubeClient.Status().Patch(ctx, terraform, patch, client.FieldOwner("tf-controller"))
	})
}
