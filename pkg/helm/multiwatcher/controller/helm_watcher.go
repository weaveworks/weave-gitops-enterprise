package controller

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/fluxcd/pkg/version"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
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

// HelmVersionFilterAnnotation applied to a HelmRepository configures the versions
// of charts that are pulled from it.
const HelmVersionFilterAnnotation = "weave.works/helm-version-filter"

// Profiles is a predicate for scanning charts with the ProfileAnnotation.
var Profiles = func(hr *sourcev1.HelmRepository, v *repo.ChartVersion) bool {
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
// Because the watcher watches all Helmrepositories, it will update data for all of them.
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

	indexFile, err := r.ValuesFetcher.GetIndexFile(ctx, r.Cluster, types.NamespacedName{
		Name:      repository.Name,
		Namespace: repository.Namespace,
	}, r.UseProxy)

	if err != nil {
		log.Error(err, "failed to get index file")
		return ctrl.Result{}, err
	}

	LoadIndex(ctx, indexFile, r.Cache, r.ClusterRef, &repository, log)

	log.Info("cached data from repository", "url", repository.Status.URL, "number of profiles", len(indexFile.Entries))

	return ctrl.Result{}, nil
}

func (r *HelmWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sourcev1.HelmRepository{}).
		WithEventFilter(predicate.Or(ArtifactUpdatePredicate{}, DeletePredicate{})).
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

// LoadIndex loads the index file for a HelmRepository into the charts cache.
//
// The charts are filtered if the HelmRepository has appropriate annotations.
func LoadIndex(ctx context.Context, index *repo.IndexFile, cache helm.ChartsCacheWriter, clusterRef types.NamespacedName, helmRepo *sourcev1.HelmRepository, log logr.Logger) {
	constraint, err := parseChartFilter(helmRepo)
	if err != nil {
		log.Error(err, "loading chart cache")
	}

	for name, versions := range index.Entries {
		for _, chartVersion := range versions {
			isProfile := Profiles(helmRepo, chartVersion)
			chartKind := "chart"
			if isProfile {
				chartKind = "profile"
			}

			ref := helm.ObjectReference{Name: helmRepo.Name, Namespace: helmRepo.Namespace}

			if constraint != nil {
				v, err := version.ParseVersion(chartVersion.Version)
				if err != nil {
					log.Error(err, "failed to parse version for chart", "name", name, "version", chartVersion.Version)
					continue
				}
				if !constraint.Check(v) {
					if err := cache.RemoveChart(ctx, name, chartVersion.Version, clusterRef, ref); err != nil {
						log.Error(err, "failed to delete chart from cache", "name", name, "version", chartVersion.Version)
					}
					continue
				}

			}

			err := cache.AddChart(
				ctx,
				name,
				chartVersion.Version,
				chartKind,
				chartVersion.Annotations[helm.LayerAnnotation],
				clusterRef,
				ref,
			)
			if err != nil {
				log.Error(err, "failed to add chart to cache", "name", name, "version", chartVersion.Version)
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

func parseChartFilter(hr *sourcev1.HelmRepository) (*semver.Constraints, error) {
	annotation := hr.GetAnnotations()[HelmVersionFilterAnnotation]
	if annotation == "" {
		return nil, nil
	}

	constraint, err := semver.NewConstraint(annotation)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chart version contraint %q on HelmRepository %s: %w", annotation, client.ObjectKeyFromObject(hr), err)
	}

	return constraint, nil
}
