package clusterswatcher

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/cluster/fetcher"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type Options struct {
	ClusterFetcher clustersmngr.ClusterFetcher
}

type watcher struct {
	clustersFetcher clustersmngr.ClusterFetcher
}

type Watcher interface {
	Watch(ctx context.Context) error
}

func NewWatcher(opts Options) (Watcher, error) {
	w := &watcher{
		clustersFetcher: opts.ClusterFetcher,
	}

	return w, nil
}

func (w *watcher) Watch(ctx context.Context) error {
	// TODO: this need to be reactive and update the cluster list as it changes
	clusters, err := w.clustersFetcher.Fetch(ctx)
	if err != nil {
		return fmt.Errorf("failed fetching clusters list: %w", err)
	}

	clients := map[string]cache.SharedIndexInformer{}

	for _, c := range clusters {
		client, err := dynamic.NewForConfig(restConfigFromCluster(c))
		if err != nil {
			return fmt.Errorf("failed to create cluster client: %w", err)

		}

		clients[c.Name] = createHelmReleaseInformer(client)
	}

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	clients[fetcher.ManagementClusterName].Run(ctx.Done())

	return nil
}

func (w *watcher) createClustersInformers(ctx context.Context) {

}

func createHelmReleaseInformer(client dynamic.Interface) cache.SharedIndexInformer {
	resource := helmv2.GroupVersion.WithResource("helmreleases")
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, time.Minute, corev1.NamespaceAll, nil)
	informer := factory.ForResource(resource).Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: helmReleaseUpdateHandler,
	})

	return informer
}

func helmReleaseUpdateHandler(oldObj, newObj interface{}) {
	hr := objToHelmRelease(newObj)

	if apimeta.IsStatusConditionPresentAndEqual(hr.GetConditions(), helmv2.ReleasedCondition, v1.ConditionTrue) {
		fmt.Println("Promote release")
	}
}

func objToHelmRelease(obj interface{}) helmv2.HelmRelease {
	un := obj.(*unstructured.Unstructured)

	var hr helmv2.HelmRelease
	runtime.DefaultUnstructuredConverter.FromUnstructured(un.UnstructuredContent(), &hr)

	return hr
}

func restConfigFromCluster(cluster clustersmngr.Cluster) *rest.Config {
	return &rest.Config{
		Host:            cluster.Server,
		TLSClientConfig: cluster.TLSConfig,
		BearerToken:     cluster.BearerToken,
	}
}
