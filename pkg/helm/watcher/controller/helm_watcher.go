package controller

import (
	"context"
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
)

const (
	watcherFinalizer = "finalizers.helm.watcher"
)

// HelmWatcherReconciler runs the `reconcile` loop for the watcher.
type HelmWatcherReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	ClusterRef    types.NamespacedName
	ClientConfig  *rest.Config
	Cache         helm.ChartsCacherWriter
	ValuesFetcher helm.ValuesFetcher
}

// +kubebuilder:rbac:groups=helm.watcher,resources=helmrepositories,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=helm.watcher,resources=helmrepositories/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=helm.watcher,resources=helmrepositories/finalizers,verbs=get;create;update;patch;delete

// Reconcile is either called when there is a new HelmRepository or, when there is an update to a HelmRepository.
// Because the watcher watches all helmrepositories, it will update data for all of them.
func (r *HelmWatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues("repository", req.NamespacedName)

	// get source object
	var repository sourcev1.HelmRepository
	if err := r.Get(ctx, req.NamespacedName, &repository); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Add our finalizer if it does not exist
	if !controllerutil.ContainsFinalizer(&repository, watcherFinalizer) {
		patch := client.MergeFrom(repository.DeepCopy())
		controllerutil.AddFinalizer(&repository, watcherFinalizer)

		if err := r.Patch(ctx, &repository, patch); err != nil {
			log.Error(err, "unable to register finalizer")
			return ctrl.Result{}, err
		}
	}

	// Examine if the object is under deletion
	if !repository.ObjectMeta.GetDeletionTimestamp().IsZero() {
		return r.reconcileDelete(ctx, repository)
	}

	if repository.Status.Artifact == nil {
		return ctrl.Result{}, nil
	}

	log.Info("found the repository: ", "name", repository.Name)
	// Reconcile is called for two reasons. One, the repository was just created, two there is a new revision.
	// Because of that, we don't care what's in the cache. We will always fetch and set it.

	indexFile, err := r.ValuesFetcher.GetIndexFile(ctx, r.ClientConfig, types.NamespacedName{
		Name:      repository.Name,
		Namespace: repository.Namespace,
	}, false)

	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get index file: %w", err)
	}
	for name, versions := range indexFile.Entries {
		for _, version := range versions {
			isProfile := helm.Profiles(&repository, version)
			chartKind := "chart"
			if isProfile {
				chartKind = "profile"
			}
			r.Cache.AddChart(
				ctx, name, version.Version, chartKind,
				version.Annotations[helm.LayerAnnotation],
				r.ClusterRef,
				helm.ObjectReference{Name: repository.Name, Namespace: repository.Namespace},
			)
		}
	}

	log.Info("cached data from repository", "url", repository.Status.URL, "name", repository.Name, "number of profiles", len(indexFile.Entries))

	return ctrl.Result{}, nil
}

func (r *HelmWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sourcev1.HelmRepository{}).
		WithEventFilter(predicate.Or(ArtifactUpdatePredicate{}, DeletePredicate{})).
		Complete(r)
}

func (r *HelmWatcherReconciler) reconcileDelete(ctx context.Context, repository sourcev1.HelmRepository) (ctrl.Result, error) {
	log := logr.FromContextOrDiscard(ctx)

	log.Info("deleting repository cache", "namespace", repository.Namespace, "name", repository.Name)

	if err := r.Cache.Delete(ctx, helm.ObjectReference{Name: repository.Name, Namespace: repository.Namespace}, r.ClusterRef); err != nil {
		log.Error(err, "failed to remove cache for repository", "namespace", repository.Namespace, "name", repository.Name)
		return ctrl.Result{}, err
	}

	log.Info("deleted repository cache", "namespace", repository.Namespace, "name", repository.Name)
	// Remove our finalizer from the list and update it
	controllerutil.RemoveFinalizer(&repository, watcherFinalizer)

	if err := r.Update(ctx, &repository); err != nil {
		log.Error(err, "failed to update repository to remove the finalizer", "namespace", repository.Namespace, "name", repository.Name)
		return ctrl.Result{}, err
	}

	log.Info("removed finalizer from repository", "namespace", repository.Namespace, "name", repository.Name)
	// Stop reconciliation as the object is being deleted
	return ctrl.Result{}, nil
}

// ConvertStringListToSemanticVersionList converts a slice of strings into a slice of semantic version.
func ConvertStringListToSemanticVersionList(versions []string) ([]*semver.Version, error) {
	var result []*semver.Version

	for _, v := range versions {
		ver, err := semver.NewVersion(v)
		if err != nil {
			return nil, err
		}

		result = append(result, ver)
	}

	return result, nil
}

// SortVersions sorts semver versions in decreasing order.
func SortVersions(versions []*semver.Version) {
	sort.SliceStable(versions, func(i, j int) bool {
		return versions[i].GreaterThan(versions[j])
	})
}
