package server

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hashicorp/go-multierror"

	ctrl "github.com/weaveworks/gitopssets-controller/api/v1alpha1"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/mgmtfetcher"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/gitopssets"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	coretypes "github.com/weaveworks/weave-gitops/core/server/types"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)


var (
	KustomizeNameKey      = fmt.Sprintf("%s/name", kustomizev1.GroupVersion.Group)
	KustomizeNamespaceKey = fmt.Sprintf("%s/namespace", kustomizev1.GroupVersion.Group)
	HelmNameKey           = fmt.Sprintf("%s/name", helmv2.GroupVersion.Group)
	HelmNamespaceKey      = fmt.Sprintf("%s/namespace", helmv2.GroupVersion.Group)
)

type ServerOpts struct {
	logr.Logger
	ClientsFactory    clustersmngr.ClustersManager
	ManagementFetcher *mgmtfetcher.ManagementCrossNamespacesFetcher
	Scheme            *k8sruntime.Scheme
	Cluster           string
}

type server struct {
	pb.UnimplementedGitOpsSetsServer

	log               logr.Logger
	clients           clustersmngr.ClustersManager
	managementFetcher *mgmtfetcher.ManagementCrossNamespacesFetcher
	scheme            *k8sruntime.Scheme
	cluster           string
}

func Hydrate(ctx context.Context, mux *runtime.ServeMux, opts ServerOpts) error {
	s := NewGitOpsSetsServer(opts)

	return pb.RegisterGitOpsSetsHandlerServer(ctx, mux, s)
}

func NewGitOpsSetsServer(opts ServerOpts) pb.GitOpsSetsServer {
	return &server{
		log:               opts.Logger,
		clients:           opts.ClientsFactory,
		managementFetcher: opts.ManagementFetcher,
		scheme:            opts.Scheme,
	}
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

func (s *server) GetReconciledObjects(ctx context.Context, msg *pb.GetReconciledObjectsRequest) (*pb.GetReconciledObjectsResponse, error) {
	clustersClient, err := s.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	var opts client.MatchingLabels

	switch msg.AutomationKind {
	case kustomizev1.KustomizationKind:
		opts = client.MatchingLabels{
			KustomizeNameKey:      msg.AutomationName,
			KustomizeNamespaceKey: msg.Namespace,
		}
	case helmv2.HelmReleaseKind:
		opts = client.MatchingLabels{
			HelmNameKey:      msg.AutomationName,
			HelmNamespaceKey: msg.Namespace,
		}
	default:
		return nil, fmt.Errorf("unsupported application kind: %s", msg.AutomationKind)
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

	for _, namespaces := range clusterUserNamespaces {
		for _, ns := range namespaces {
			nsOpts := client.InNamespace(ns.Name)

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
							// s.logger.V(logger.LogLevelDebug).Info(
							// 	"forbidden list request",
							// 	"cluster", msg.ClusterName,
							// 	"automation", msg.AutomationName,
							// 	"namespace", msg.Namespace,
							// 	"gvk", gvk.String(),
							// )
							// Our service account (or impersonated user) may not have the ability to see the resource in question,
							// in the given namespace. We pretend it doesn't exist and keep looping.
							// We need logging to make this error more visible.
							return
						}

						if k8serrors.IsTimeout(err) {
							// s.logger.Error(err, "List timedout", "gvk", gvk.String())

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
		}
	}

	wg.Wait()

	objects := []*pb.Object{}
	respErrors := multierror.Error{}

	for _, unstructuredObj := range result {
		tenant := GetTenant(unstructuredObj.GetNamespace(), msg.ClusterName, clusterUserNamespaces)

		var o *pb.Object

		var obj client.Object = &unstructuredObj

		if unstructuredObj.GetKind() == "Secret" {
			obj, err = sanitizeSecret(&unstructuredObj)
			if err != nil {
				respErrors = *multierror.Append(fmt.Errorf("error sanitizing secrets: %w", err), respErrors.Errors...)
				continue
			}
		}

		o, err = coretypes.K8sObjectToProto(obj, msg.ClusterName, tenant, nil)
		if err != nil {
			respErrors = *multierror.Append(fmt.Errorf("error converting objects: %w", err), respErrors.Errors...)
			continue
		}

		objects = append(objects, o)
	}

	return &pb.GetReconciledObjectsResponse{Objects: objects}, respErrors.ErrorOrNil()
}


func (cs *server) GetChildObjects(ctx context.Context, msg *pb.GetChildObjectsRequest) (*pb.GetChildObjectsResponse, error) {
	clustersClient, err := cs.clients.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	opts := client.InNamespace(msg.Namespace)

	listResult := unstructured.UnstructuredList{}

	listResult.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   msg.GroupVersionKind.Group,
		Version: msg.GroupVersionKind.Version,
		Kind:    msg.GroupVersionKind.Kind,
	})

	if err := clustersClient.List(ctx, msg.ClusterName, &listResult, opts); err != nil {
		return nil, fmt.Errorf("could not get unstructured object: %s", err)
	}

	respErrors := multierror.Error{}
	clusterUserNamespaces := cs.clients.GetUserNamespaces(auth.Principal(ctx))
	objects := []*pb.Object{}

ItemsLoop:
	for _, obj := range listResult.Items {
		refs := obj.GetOwnerReferences()
		if len(refs) == 0 {
			// Ignore items without OwnerReference.
			// for example: dev-weave-gitops-test-connection
			continue ItemsLoop
		}

		for _, ref := range refs {
			if ref.UID != types.UID(msg.ParentUid) {
				// Assuming all owner references have the same parent UID,
				// this is not the child we are looking for.
				// Skip the rest of the operations in Items loops.
				continue ItemsLoop
			}
		}

		tenant := GetTenant(obj.GetNamespace(), msg.ClusterName, clusterUserNamespaces)

		obj, err := coretypes.K8sObjectToProto(&obj, msg.ClusterName, tenant, nil)

		if err != nil {
			respErrors = *multierror.Append(fmt.Errorf("error converting objects: %w", err), respErrors.Errors...)
			continue
		}
		objects = append(objects, obj)
	}

	return &pb.GetChildObjectsResponse{Objects: objects}, nil
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
