package objectcollector

import (
	"fmt"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	automation1 "github.com/fluxcd/image-automation-controller/api/v1beta1"
	reflectorv1 "github.com/fluxcd/image-reflector-controller/api/v1beta1"
	kustomizev2 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type objectCollector struct {
	col collector.Collector
	log logr.Logger
	w   store.StoreWriter
}

var DefaultObjectCollectorKinds = []schema.GroupVersionKind{
	helmv2.GroupVersion.WithKind(helmv2.HelmReleaseKind),
	automation1.GroupVersion.WithKind(automation1.ImageUpdateAutomationKind),
	reflectorv1.GroupVersion.WithKind(reflectorv1.ImagePolicyKind),
	kustomizev2.GroupVersion.WithKind(kustomizev2.KustomizationKind),
	sourcev1.GroupVersion.WithKind(sourcev1.GitRepositoryKind),
	sourcev1.GroupVersion.WithKind(sourcev1.HelmRepositoryKind),
	sourcev1.GroupVersion.WithKind(sourcev1.BucketKind),
}

func NewObjectCollector(log logr.Logger, mgr clustersmngr.ClustersManager, w store.StoreWriter, kinds []schema.GroupVersionKind) *objectCollector {
	if kinds == nil {
		kinds = DefaultObjectCollectorKinds
	}

	col := collector.NewCollector(collector.CollectorOpts{
		Log:            log,
		ClusterManager: mgr,
		ObjectKinds:    kinds,
		PollInterval:   10 * time.Second,
	})

	return &objectCollector{
		col: col,
		log: log,
		w:   w,
	}
}

func (o *objectCollector) Start() {
	go func() {
		ch, error := o.col.Start()
		if error != nil {
			o.log.Error(error, "failed to start object collector")
			return
		}

		for {
			select {
			case objects := <-ch:
				converted := convertToObject(objects)
				if err := o.w.StoreObjects(converted); err != nil {
					o.log.Error(err, "failed to store objects")
					continue
				}
			}
		}
	}()
}

func convertToObject(objs []collector.ObjectRecord) []models.Object {
	objects := make([]models.Object, 0, len(objs))
	for _, obj := range objs {
		o := obj.Object()

		v, k := o.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()

		fullKind := fmt.Sprintf("%s/%s", v, k)

		objects = append(objects, models.Object{
			Kind:      fullKind,
			Name:      o.GetName(),
			Namespace: o.GetNamespace(),
			Cluster:   obj.ClusterName(),
		})
	}

	return objects
}
