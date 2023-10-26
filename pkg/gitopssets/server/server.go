package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hashicorp/go-multierror"

	ctrl "github.com/weaveworks/gitopssets-controller/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/gitopssets"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/gitopssets/adapter"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/gitopssets/internal/convert"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/fluxsync"
	"github.com/weaveworks/weave-gitops/core/logger"
	core "github.com/weaveworks/weave-gitops/core/server"
	"github.com/weaveworks/weave-gitops/pkg/health"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8s_json "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	GitOpsSetNameKey      = fmt.Sprintf("%s/name", ctrl.GroupVersion.Group)
	GitOpsSetNamespaceKey = fmt.Sprintf("%s/namespace", ctrl.GroupVersion.Group)
)

type ServerOpts struct {
	logr.Logger
	ClientsFactory clustersmngr.ClustersManager
	HealthChecker  health.HealthChecker
}

type server struct {
	pb.UnimplementedGitOpsSetsServer

	log           logr.Logger
	clients       clustersmngr.ClustersManager
	healthChecker health.HealthChecker
}

func Hydrate(ctx context.Context, mux *runtime.ServeMux, opts ServerOpts) error {
	s := NewGitOpsSetsServer(opts)

	return pb.RegisterGitOpsSetsHandlerServer(ctx, mux, s)
}

func NewGitOpsSetsServer(opts ServerOpts) pb.GitOpsSetsServer {
	return &server{
		log:           opts.Logger,
		clients:       opts.ClientsFactory,
		healthChecker: opts.HealthChecker,
	}
}

func (s *server) ListGitOpsSets(ctx context.Context, msg *pb.ListGitOpsSetsRequest) (*pb.ListGitOpsSetsResponse, error) {
	clustersClient, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))
	clientConnectionErrors, err := toListErrors(err)
	if err != nil {
		return nil, fmt.Errorf("error extracting errors: %w", err)
	}

	gitopsSets, gitopsSetsListErrors, err := s.listGitopsSets(ctx, clustersClient)
	if err != nil {
		return nil, fmt.Errorf("listing gitops sets: %w", err)
	}

	response := pb.ListGitOpsSetsResponse{
		Errors:     append(clientConnectionErrors, gitopsSetsListErrors...),
		Gitopssets: gitopsSets,
	}

	return &response, nil
}

func (s *server) listGitopsSets(ctx context.Context, cl clustersmngr.Client) ([]*pb.GitOpsSet, []*pb.GitOpsSetListError, error) {
	list := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &ctrl.GitOpsSetList{}
	})

	listErrors := []*pb.GitOpsSetListError{}
	if err := cl.ClusteredList(ctx, list, true); err != nil {
		var errs clustersmngr.ClusteredListError

		if !errors.As(err, &errs) {
			return nil, nil, fmt.Errorf("converting to ClusteredListError: %w", errs)
		}

		for _, e := range errs.Errors {
			if apimeta.IsNoMatchError(e.Err) {
				// Skip reporting an error if a leaf cluster does not have the gitopssets CRD installed.
				s.log.Info("gitopssets crd not present on cluster, skipping error", "cluster", e.Cluster)
				continue
			}

			if k8serrors.IsForbidden(e.Err) {
				// Skip reporting an error if user doesn't have permission to list gitopssets on a cluster.
				s.log.V(logger.LogLevelDebug).Info("user does not have permission to list gitopssets in namespace/cluster", "namespace", e.Namespace, "cluster", e.Cluster)
				continue
			}

			listErrors = append(listErrors, &pb.GitOpsSetListError{
				ClusterName: e.Cluster,
				Message:     e.Err.Error(),
				Namespace:   e.Namespace,
			})
		}
	}

	gitopsSets := []*pb.GitOpsSet{}
	for clusterName, objs := range list.Lists() {
		for i := range objs {
			// TODO: why do we need this?
			obj, ok := objs[i].(*ctrl.GitOpsSetList)
			if !ok {
				continue
			}

			for _, gs := range obj.Items {
				gitOpsSet, err := convert.GitOpsToProto(clusterName, gs)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to convert gitopsset: %w", err)
				}
				gitopsSets = append(gitopsSets, gitOpsSet)
			}
		}
	}

	return gitopsSets, listErrors, nil
}

func (s *server) GetGitOpsSet(ctx context.Context, msg *pb.GetGitOpsSetRequest) (*pb.GetGitOpsSetResponse, error) {
	c, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("getting impersonated client: %w", err)
	}

	n := types.NamespacedName{Name: msg.Name, Namespace: msg.Namespace}

	result := ctrl.GitOpsSet{}
	if err := c.Get(ctx, msg.ClusterName, n, &result); err != nil {
		return nil, fmt.Errorf("getting object with name %s in namespace %s: %w", msg.Name, msg.Namespace, err)
	}

	// client.Get does not always populate TypeMeta field, without this `kind` and
	// `apiVersion` are not returned in YAML representation.
	// https://github.com/kubernetes-sigs/controller-runtime/issues/1517#issuecomment-844703142
	result.GetObjectKind().SetGroupVersionKind(ctrl.GroupVersion.WithKind("GitOpsSet"))

	gitOpsSet, err := convert.GitOpsToProto(msg.ClusterName, result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert gitopsset: %w", err)
	}

	return &pb.GetGitOpsSetResponse{
		GitopsSet: gitOpsSet,
	}, nil
}

func (s *server) ToggleSuspendGitOpsSet(ctx context.Context, msg *pb.ToggleSuspendGitOpsSetRequest) (*pb.ToggleSuspendGitOpsSetResponse, error) {
	clustersClient, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("getting impersonated client: %w", err)
	}

	c, err := clustersClient.Scoped(msg.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("getting scoped client: %w", err)
	}

	key := client.ObjectKey{
		Name:      msg.Name,
		Namespace: msg.Namespace,
	}

	obj := &ctrl.GitOpsSet{}

	if err := c.Get(ctx, key, obj); err != nil {
		return nil, fmt.Errorf("getting object %s in namespace %s: %w", msg.Name, msg.Namespace, err)
	}

	patch := client.MergeFrom(obj.DeepCopy())

	obj.Spec.Suspend = msg.Suspend

	if err := c.Patch(ctx, obj, patch); err != nil {
		return nil, fmt.Errorf("patching object: %w", err)
	}

	return &pb.ToggleSuspendGitOpsSetResponse{}, nil
}

func (s *server) unstructuredToInventoryEntry(ctx context.Context, clusterName string, objWithChildren []*core.ObjectWithChildren, clusterUserNamespaces map[string][]v1.Namespace, healthChecker health.HealthChecker) ([]*pb.InventoryEntry, error) {
	inventoryEntries := []*pb.InventoryEntry{}

	for _, objWithChildrenEntry := range objWithChildren {
		unstructuredObj := objWithChildrenEntry.Object
		children, err := s.unstructuredToInventoryEntry(ctx, clusterName, objWithChildrenEntry.Children, clusterUserNamespaces, healthChecker)
		if err != nil {
			return nil, err
		}

		bytes, err := unstructuredObj.MarshalJSON()
		if err != nil {
			return nil, err
		}

		clusterUserNss := s.clients.GetUserNamespaces(auth.Principal(ctx))
		tenant := core.GetTenant(unstructuredObj.GetNamespace(), clusterName, clusterUserNss)

		health, err := s.healthChecker.Check(*unstructuredObj)
		if err != nil {
			return nil, err
		}

		entry := &pb.InventoryEntry{
			Payload:     string(bytes),
			Tenant:      tenant,
			ClusterName: clusterName,
			Children:    children,
			Health: &pb.HealthStatus{
				Status:  string(health.Status),
				Message: health.Message,
			},
		}

		inventoryEntries = append(inventoryEntries, entry)

	}
	return inventoryEntries, nil

}

func (s *server) GetInventory(ctx context.Context, msg *pb.GetInventoryRequest) (*pb.GetInventoryResponse, error) {
	clustersClient, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	client, err := clustersClient.Scoped(msg.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("error getting scoped client for cluster=%s: %w", msg.ClusterName, err)
	}

	// get the gitopssets
	// read the inventory from it
	var inventoryRefs []*unstructured.Unstructured
	inventoryRefs, err = s.getGitOpsSetInventory(ctx, client, msg.Name, msg.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed getting kustomization inventory: %w", err)
	}

	// go and get all the objects and maybe children of the inventory items
	objsWithChildren, err := core.GetObjectsWithChildren(ctx, inventoryRefs, client, msg.WithChildren, s.log)
	if err != nil {
		return nil, fmt.Errorf("failed getting objects with children: %w", err)
	}

	// post process it into a response object:
	// add health data and tenancy data and redact secrets etc
	clusterUserNamespaces := s.clients.GetUserNamespaces(auth.Principal(ctx))
	entries, err := s.unstructuredToInventoryEntry(ctx, msg.ClusterName, objsWithChildren, clusterUserNamespaces, s.healthChecker)
	if err != nil {
		return nil, fmt.Errorf("failed converting inventory entry: %w", err)
	}

	return &pb.GetInventoryResponse{
		Entries: entries,
	}, nil

}

func (s *server) getGitOpsSetInventory(ctx context.Context, k8sClient client.Client, name, namespace string) ([]*unstructured.Unstructured, error) {
	gs := &ctrl.GitOpsSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(gs), gs); err != nil {
		return nil, fmt.Errorf("failed to get kustomization: %w", err)
	}

	if gs.Status.Inventory == nil {
		return nil, nil
	}

	if gs.Status.Inventory.Entries == nil {
		return nil, nil
	}

	objects := []*unstructured.Unstructured{}
	for _, ref := range gs.Status.Inventory.Entries {
		obj, err := core.ResourceRefToUnstructured(ref.ID, ref.Version)
		if err != nil {
			s.log.Error(err, "failed converting inventory entry", "entry", ref)
			return nil, err
		}
		objects = append(objects, &obj)
	}

	return objects, nil
}

func (s *server) unstructuredToInventoryEntry(ctx context.Context, clusterName string, objWithChildren []*core.ObjectWithChildren, clusterUserNamespaces map[string][]v1.Namespace, healthChecker health.HealthChecker) ([]*pb.InventoryEntry, error) {
	inventoryEntries := []*pb.InventoryEntry{}

	for _, objWithChildrenEntry := range objWithChildren {
		unstructuredObj := objWithChildrenEntry.Object
		children, err := s.unstructuredToInventoryEntry(ctx, clusterName, objWithChildrenEntry.Children, clusterUserNamespaces, healthChecker)
		if err != nil {
			return nil, err
		}

		bytes, err := unstructuredObj.MarshalJSON()
		if err != nil {
			return nil, err
		}

		clusterUserNss := s.clients.GetUserNamespaces(auth.Principal(ctx))
		tenant := core.GetTenant(unstructuredObj.GetNamespace(), clusterName, clusterUserNss)

		health, err := s.healthChecker.Check(*unstructuredObj)
		if err != nil {
			return nil, err
		}

		entry := &pb.InventoryEntry{
			Payload:     string(bytes),
			Tenant:      tenant,
			ClusterName: clusterName,
			Children:    children,
			Health: &pb.HealthStatus{
				Status:  string(health.Status),
				Message: health.Message,
			},
		}

		inventoryEntries = append(inventoryEntries, entry)

	}
	return inventoryEntries, nil

}

func (s *server) GetInventory(ctx context.Context, msg *pb.GetInventoryRequest) (*pb.GetInventoryResponse, error) {
	clustersClient, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	client, err := clustersClient.Scoped(msg.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("error getting scoped client for cluster=%s: %w", msg.ClusterName, err)
	}

	// get the gitopssets
	// read the inventory from it
	var inventoryRefs []*unstructured.Unstructured
	inventoryRefs, err = s.getGitOpsSetInventory(ctx, client, msg.Name, msg.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed getting kustomization inventory: %w", err)
	}

	// go and get all the objects and maybe children of the inventory items
	objsWithChildren, err := core.GetObjectsWithChildren(ctx, inventoryRefs, client, msg.WithChildren, s.log)
	if err != nil {
		return nil, fmt.Errorf("failed getting objects with children: %w", err)
	}

	// post process it into a response object:
	// add health data and tenancy data and redact secrets etc
	clusterUserNamespaces := s.clients.GetUserNamespaces(auth.Principal(ctx))
	entries, err := s.unstructuredToInventoryEntry(ctx, msg.ClusterName, objsWithChildren, clusterUserNamespaces, s.healthChecker)
	if err != nil {
		return nil, fmt.Errorf("failed converting inventory entry: %w", err)
	}

	return &pb.GetInventoryResponse{
		Entries: entries,
	}, nil

}

func (s *server) getGitOpsSetInventory(ctx context.Context, k8sClient client.Client, name, namespace string) ([]*unstructured.Unstructured, error) {
	gs := &ctrl.GitOpsSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(gs), gs); err != nil {
		return nil, fmt.Errorf("failed to get kustomization: %w", err)
	}

	if gs.Status.Inventory == nil {
		return nil, nil
	}

	if gs.Status.Inventory.Entries == nil {
		return nil, nil
	}

	objects := []*unstructured.Unstructured{}
	for _, ref := range gs.Status.Inventory.Entries {
		obj, err := core.ResourceRefToUnstructured(ref.ID, ref.Version)
		if err != nil {
			s.log.Error(err, "failed converting inventory entry", "entry", ref)
			return nil, err
		}
		objects = append(objects, &obj)
	}

	return objects, nil
}

func (s *server) GetReconciledObjects(ctx context.Context, msg *pb.GetReconciledObjectsRequest) (*pb.GetReconciledObjectsResponse, error) {
	clustersClient, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	opts := client.MatchingLabels{
		GitOpsSetNameKey:      msg.Name,
		GitOpsSetNamespaceKey: msg.Namespace,
	}

	var (
		result   = []unstructured.Unstructured{}
		checkDup = map[types.UID]bool{}
		resultMu = sync.Mutex{}

		errs   = &multierror.Error{}
		errsMu = sync.Mutex{}

		wg = sync.WaitGroup{}
	)

	clusterUserNamespaces := s.clients.GetUserNamespaces(auth.Principal(ctx))
	nsOpts := client.InNamespace(msg.Namespace)
	for _, gvk := range msg.Kinds {

		wg.Add(1)

		go func(clusterName string, gvk *pb.GroupVersionKind) {
			defer wg.Done()

			listResult := unstructured.UnstructuredList{}

			listResult.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   gvk.Group,
				Kind:    gvk.Kind,
				Version: gvk.Version,
			})

			if err := clustersClient.List(ctx, msg.ClusterName, &listResult, opts, nsOpts); err != nil {
				if k8serrors.IsForbidden(err) {
					s.log.V(logger.LogLevelDebug).Info(
						"forbidden list request",
						"cluster", msg.ClusterName,
						"automation", msg.Name,
						"namespace", msg.Namespace,
						"gvk", gvk.String(),
					)
					// Our service account (or impersonated user) may not have the ability to see the resource in question,
					// in the given namespace. We pretend it doesn't exist and keep looping.
					// We need logging to make this error more visible.
					return
				}

				if k8serrors.IsTimeout(err) {
					s.log.Error(err, "List timedout", "gvk", gvk.String())
					return
				}

				errsMu.Lock()
				errs = multierror.Append(errs, fmt.Errorf("listing unstructured object: %w", err))
				errsMu.Unlock()
			}

			resultMu.Lock()
			for _, u := range listResult.Items {
				uid := u.GetUID()

				if !checkDup[uid] {
					result = append(result, u)
					checkDup[uid] = true
				}
			}
			resultMu.Unlock()
		}(msg.ClusterName, gvk)
	}

	wg.Wait()

	objects := []*pb.Object{}
	respErrors := multierror.Error{}

	for _, unstructuredObj := range result {
		tenant := core.GetTenant(unstructuredObj.GetNamespace(), msg.ClusterName, clusterUserNamespaces)

		var obj client.Object = &unstructuredObj

		if unstructuredObj.GetKind() == "Secret" {
			obj, err = sanitizeSecret(&unstructuredObj)
			if err != nil {
				respErrors = *multierror.Append(fmt.Errorf("error sanitizing secrets: %w", err), respErrors.Errors...)
				continue
			}
		}

		o, err := K8sObjectToProto(obj, msg.ClusterName, tenant, nil)
		if err != nil {
			respErrors = *multierror.Append(fmt.Errorf("error converting objects: %w", err), respErrors.Errors...)
			continue
		}

		objects = append(objects, o)
	}

	return &pb.GetReconciledObjectsResponse{Objects: objects}, respErrors.ErrorOrNil()
}

func sanitizeSecret(obj *unstructured.Unstructured) (client.Object, error) {
	bytes, err := obj.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshaling secret: %v", err)
	}

	s := &v1.Secret{}

	if err := json.Unmarshal(bytes, s); err != nil {
		return nil, fmt.Errorf("unmarshaling secret: %v", err)
	}

	s.Data = map[string][]byte{"redacted": []byte(nil)}

	return s, nil
}

func K8sObjectToProto(object client.Object, clusterName string, tenant string, inventory []*pb.GroupVersionKind) (*pb.Object, error) {
	var buf bytes.Buffer

	serializer := k8s_json.NewSerializer(k8s_json.DefaultMetaFactory, nil, nil, false)
	if err := serializer.Encode(object, &buf); err != nil {
		return nil, err
	}

	obj := &pb.Object{
		Payload:     buf.String(),
		ClusterName: clusterName,
		Tenant:      tenant,
		Uid:         string(object.GetUID()),
		Inventory:   inventory,
	}

	return obj, nil
}

func (s *server) SyncGitOpsSet(ctx context.Context, msg *pb.SyncGitOpsSetRequest) (*pb.SyncGitOpsSetResponse, error) {
	clustersClient, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("getting impersonated client: %w", err)
	}

	c, err := clustersClient.Scoped(msg.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("getting scoped client: %w", err)
	}

	key := client.ObjectKey{
		Name:      msg.Name,
		Namespace: msg.Namespace,
	}

	obj := adapter.GitOpsSetAdapter{GitOpsSet: &ctrl.GitOpsSet{}}

	if err := c.Get(ctx, key, obj.AsClientObject()); err != nil {
		return nil, fmt.Errorf("getting object %s in namespace %s: %w", msg.Name, msg.Namespace, err)
	}

	if err := fluxsync.RequestReconciliation(ctx, c, key, obj.GroupVersionKind()); err != nil {
		return nil, fmt.Errorf("requesting reconciliation: %w", err)
	}

	if err := fluxsync.WaitForSync(ctx, c, key, obj); err != nil {
		return nil, fmt.Errorf("waiting for sync: %w", err)
	}

	return &pb.SyncGitOpsSetResponse{Success: true}, nil
}

func toListErrors(err error) ([]*pb.GitOpsSetListError, error) {
	respErrors := []*pb.GitOpsSetListError{}

	if err != nil {
		if merr, ok := err.(*multierror.Error); ok {
			for _, err := range merr.Errors {
				if cerr, ok := err.(*clustersmngr.ClientError); ok {
					respErrors = append(respErrors, &pb.GitOpsSetListError{
						ClusterName: cerr.ClusterName,
						Message:     cerr.Error(),
					})
				}
			}
		} else {
			return nil, fmt.Errorf("unexpected error while getting clusters client, error: %w", err)
		}
	}
	return respErrors, nil
}
