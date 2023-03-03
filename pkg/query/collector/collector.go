package collector

import (
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Collector
type Collector interface {
	Start() (<-chan []ObjectRecord, error)
	Stop() error
}

type CollectorOpts struct {
	Log            logr.Logger
	ClusterManager clustersmngr.ClustersManager
	ObjectKinds    []schema.GroupVersionKind
	PollInterval   time.Duration
}

//counterfeiter:generate . ObjectRecord
type ObjectRecord interface {
	ClusterName() string
	Object() client.Object
}

func NewCollector(opts CollectorOpts) Collector {
	return &pollingCollector{
		mgr:    opts.ClusterManager,
		log:    opts.Log,
		kinds:  opts.ObjectKinds,
		ticker: time.NewTicker(opts.PollInterval),
		quit:   make(chan bool, 1),
		msg:    make(chan []ObjectRecord, 1),
	}
}

func convertK8sToModelObject(obj *unstructured.Unstructured) models.Object {
	return models.Object{
		Kind:      obj.GetKind(),
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}
}
