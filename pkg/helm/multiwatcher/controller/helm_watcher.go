package controller

import (
	"context"

	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/helm"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
)

// ProfileAnnotation is the annotation that Helm charts must have to indicate
// that they provide a Profile.
const ProfileAnnotation = "weave.works/profile"

// RepositoryProfilesAnnotation is the annotation that Helm Repositories must
// have to indicate that all charts are to be considered as Profiles.
const RepositoryProfilesAnnotation = "weave.works/profiles"

// Profiles is a predicate for scanning charts with the ProfileAnnotation.
var Profiles = func(hr *sourcev1beta2.HelmRepository, v *repo.ChartVersion) bool {
	return hasAnnotation(v.Metadata.Annotations, ProfileAnnotation) ||
		hasAnnotation(hr.ObjectMeta.Annotations, RepositoryProfilesAnnotation)
}

// HelmWatcherReconciler runs the `reconcile` loop for the watcher.
type HelmWatcherReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	UseProxy      bool
	ClusterRef    types.NamespacedName
	Cluster       cluster.Cluster
	Cache         helm.ChartsCacheWriter
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
	var repository sourcev1beta2.HelmRepository
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

	indexFile, err := r.ValuesFetcher.GetIndexFile(ctx, r.Cluster, types.NamespacedName{
		Name:      repository.Name,
		Namespace: repository.Namespace,
	}, r.UseProxy)

	if err != nil {
		log.Error(err, "failed to get index file")
		return ctrl.Result{}, err
	}

	LoadIndex(indexFile, r.Cache, r.ClusterRef, &repository, log)

	log.Info("cached data from repository", "url", repository.Status.URL, "number of profiles", len(indexFile.Entries))

	return ctrl.Result{}, nil
}

func (r *HelmWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sourcev1beta2.HelmRepository{}).
		WithEventFilter(predicate.Or(ArtifactUpdatePredicate{}, DeletePredicate{})).
		Complete(r)
}

func (r *HelmWatcherReconciler) reconcileDelete(ctx context.Context, repository sourcev1beta2.HelmRepository) (ctrl.Result, error) {
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

// LoadIndex loads the index file for a HelmRepository into the charts cache
func LoadIndex(index *repo.IndexFile, cache helm.ChartsCacheWriter, clusterRef types.NamespacedName, helmRepo *sourcev1beta2.HelmRepository, log logr.Logger) {
	for name, versions := range index.Entries {
		for _, version := range versions {
			isProfile := Profiles(helmRepo, version)
			chartKind := "chart"
			if isProfile {
				chartKind = "profile"
			}
			err := cache.AddChart(
				context.Background(),
				name,
				version.Version,
				chartKind,
				version.Annotations[helm.LayerAnnotation],
				clusterRef,
				helm.ObjectReference{Name: helmRepo.Name, Namespace: helmRepo.Namespace},
			)
			if err != nil {
				log.Error(err, "failed to add chart to cache", "name", name, "version", version.Version)
			}
		}
	}
}

func hasAnnotation(cm map[string]string, name string) bool {
	for k := range cm {
		if k == name {
			return true
		}
	}

	return false
}
