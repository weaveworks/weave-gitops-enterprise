package controller

import (
	"context"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	helmwatcher "github.com/weaveworks/weave-gitops-enterprise/pkg/helm/watcher/controller"
)

// HelmWatcherReconciler runs the `reconcile` loop for the watcher.
type HelmWatcherReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	UseProxy      bool
	ClusterRef    types.NamespacedName
	ClientConfig  *rest.Config
	Cache         helm.ChartsCacherWriter
	ValuesFetcher helm.ValuesFetcher
}

// Reconcile is either called when there is a new HelmRepository or, when there is an update to a HelmRepository.
// Because the watcher watches all helmrepositories, it will update data for all of them.
func (r *HelmWatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues(
		"repository", req.NamespacedName,
		"useProxy", r.UseProxy,
		"cluster", r.ClusterRef,
	)

	// get source object
	var repository sourcev1.HelmRepository
	if err := r.Get(ctx, req.NamespacedName, &repository); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Examine if the object is under deletion
	if !repository.ObjectMeta.GetDeletionTimestamp().IsZero() {
		return r.reconcileDelete(ctx, repository)
	}

	if repository.Status.Artifact == nil {
		return ctrl.Result{}, nil
	}

	log.Info("found the repository")

	// Reconcile is called for two reasons. One, the repository was just created, two there is a new revision.
	// Because of that, we don't care what's in the cache. We will always fetch and set it.

	indexFile, err := r.ValuesFetcher.GetIndexFile(ctx, r.ClientConfig, types.NamespacedName{
		Name:      repository.Name,
		Namespace: repository.Namespace,
	}, r.UseProxy)

	if err != nil {
		log.Error(err, "failed to get index file")
		return ctrl.Result{}, err
	}

	for name, versions := range indexFile.Entries {
		for _, version := range versions {
			isProfile := helm.Profiles(&repository, version)
			chartKind := "chart"
			if isProfile {
				chartKind = "profile"
			}
			err := r.Cache.AddChart(
				ctx, name, version.Version, chartKind,
				version.Annotations[helm.LayerAnnotation],
				r.ClusterRef,
				helm.ObjectReference{Name: repository.Name, Namespace: repository.Namespace},
			)
			if err != nil {
				log.Error(err, "failed to add chart to cache", "name", name, "version", version.Version)
			}
		}
	}

	log.Info("cached data from repository", "url", repository.Status.URL, "number of profiles", len(indexFile.Entries))

	return ctrl.Result{}, nil
}

func (r *HelmWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sourcev1.HelmRepository{}).
		WithEventFilter(predicate.Or(helmwatcher.ArtifactUpdatePredicate{}, helmwatcher.DeletePredicate{})).
		Complete(r)
}

func (r *HelmWatcherReconciler) reconcileDelete(ctx context.Context, repository sourcev1.HelmRepository) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues(
		"repository", repository,
		"cluster", r.ClusterRef,
	)

	log.Info("deleting repository cache")
	if err := r.Cache.Delete(ctx, helm.ObjectReference{Name: repository.Name, Namespace: repository.Namespace}, r.ClusterRef); err != nil {
		log.Error(err, "failed to remove cache for repository")
		return ctrl.Result{}, err
	}

	log.Info("deleted repository cache")

	// Stop reconciliation as the object is being deleted
	return ctrl.Result{}, nil
}
