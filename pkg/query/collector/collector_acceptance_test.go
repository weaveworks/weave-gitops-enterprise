//go:build acceptance

package collector_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/accesscollector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster/clusterfakes"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const (
	defaultTimeout  = time.Second * 10
	defaultInterval = time.Second
)

type collectorKey struct{}

var log logr.Logger
var g *WithT

func TestCollectorAcceptance(t *testing.T) {
	g = NewWithT(t)
	g.SetDefaultEventuallyTimeout(defaultTimeout)
	g.SetDefaultEventuallyPollingInterval(defaultInterval)

	log = testr.New(t)
	tests := []struct {
		name       string
		gvk        schema.GroupVersionKind
		errPattern string
	}{
		{
			name: "can watch helm releases",
			gvk:  v2beta1.GroupVersion.WithKind("HelmRelease"),
		},
		{
			name: "can watch kustomizations",
			gvk:  v1beta2.GroupVersion.WithKind("Kustomization"),
		},
		{
			name: "can watch cluster roles",
			gvk:  schema.GroupVersion{Group: "rbac.authorization.k8s.io", Version: "v1"}.WithKind("ClusterRole"),
		},
		{
			name: "can watch roles",
			gvk:  schema.GroupVersion{Group: "rbac.authorization.k8s.io", Version: "v1"}.WithKind("Role"),
		},
		{
			name: "can watch cluster role bindings",
			gvk:  schema.GroupVersion{Group: "rbac.authorization.k8s.io", Version: "v1"}.WithKind("ClusterRoleBinding"),
		},
		{
			name: "can watch role bindings",
			gvk:  schema.GroupVersion{Group: "rbac.authorization.k8s.io", Version: "v1"}.WithKind("RoleBinding"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx, err := aKubernetesClusterToWatch(ctx)
			if err != nil {
				t.Fatalf(err.Error())
			}
			ctx, err = aGvkToWatch(ctx, tt.gvk)
			if err != nil {
				t.Fatalf(err.Error())
			}
			ctx, err = aCollector(ctx)
			if err != nil {
				t.Fatalf(err.Error())
			}
			ctx, err = aKubernetesClusterWithResourcesOfThatKind(ctx)
			if err != nil {
				t.Fatalf(err.Error())
			}

			ctx, err = watchedTheKindInTheCluster(ctx)
			if err != nil {
				t.Fatalf(err.Error())
			}
			ctx, err = iGotAllTheResults(ctx)
			if err != nil {
				t.Fatalf(err.Error())
			}
		})
	}
}

func TestAccessCollectorAcceptance(t *testing.T) {
	g = NewWithT(t)
	g.SetDefaultEventuallyTimeout(defaultTimeout)
	g.SetDefaultEventuallyPollingInterval(defaultInterval)

	log = testr.New(t)
	tests := []struct {
		name       string
		gvk        schema.GroupVersionKind
		errPattern string
	}{
		{
			name: "can watch rbac via access collector",
			gvk:  schema.GroupVersion{Group: "rbac.authorization.k8s.io", Version: "v1"}.WithKind("RoleBinding"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx, err := aKubernetesClusterToWatch(ctx)
			if err != nil {
				t.Fatalf(err.Error())
			}
			ctx, err = aGvkToWatch(ctx, tt.gvk)
			if err != nil {
				t.Fatalf(err.Error())
			}
			ctx, err = anAccessCollector(ctx)
			if err != nil {
				t.Fatalf(err.Error())
			}
			ctx, err = aKubernetesClusterWithResourcesOfThatKind(ctx)
			if err != nil {
				t.Fatalf(err.Error())
			}

			ctx, err = watchedTheKindInTheCluster(ctx)
			if err != nil {
				t.Fatalf(err.Error())
			}
			ctx, err = iGotAllTheResults(ctx)
			if err != nil {
				t.Fatalf(err.Error())
			}
		})
	}

}

func aKubernetesClusterToWatch(ctx context.Context) (context.Context, error) {
	//create config
	cfg, err := kubeEnvironment()
	if err != nil {
		return ctx, fmt.Errorf("could not start kube environment: %w", err)
	}

	//create runtime client with application schemes
	v2beta1.AddToScheme(scheme.Scheme)
	v1beta2.AddToScheme(scheme.Scheme)
	if err != nil {
		return ctx, err
	}
	runtimeClient, err := client.New(cfg, client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		return ctx, err
	}

	clusterRef := types.NamespacedName{
		Name:      cfg.Host,
		Namespace: "default",
	}
	newCluster := makeCluster(clusterRef.Name, cfg, runtimeClient, log)
	ctx = context.WithValue(ctx, clusterKey{}, newCluster)
	return ctx, nil
}

func makeCluster(name string, config *rest.Config, client client.Client, log logr.Logger) cluster.Cluster {
	cluster := clusterfakes.FakeCluster{}
	cluster.GetNameReturns(name)
	cluster.GetServerConfigReturns(config, nil)
	cluster.GetServerClientReturns(client, nil)

	return &cluster
}

func aCollector(ctx context.Context) (context.Context, error) {
	watchGvk := ctx.Value(watchGvkKey{}).(schema.GroupVersionKind)
	g.Expect(watchGvk).NotTo(BeNil())
	//retrieve clusterName to watch
	c := ctx.Value(clusterKey{}).(cluster.Cluster)
	g.Expect(c).NotTo(BeNil())

	//create store
	dbDir, err := os.MkdirTemp("", "db")
	s, err := store.NewStore(dbDir, log)
	g.Expect(err).To(BeNil())
	ctx = context.WithValue(ctx, storeKey{}, s)

	opts := collector.CollectorOpts{
		Log: log,
		Clusters: []cluster.Cluster{
			c,
		},
		ObjectKinds: []schema.GroupVersionKind{
			watchGvk,
		},
		NewWatcherFunc:     nil,
		ProcessRecordsFunc: nil,
	}

	collector, err := collector.NewCollector(opts, s)
	g.Expect(err).To(BeNil())
	g.Expect(collector).ToNot(BeNil())
	ctx = context.WithValue(ctx, collectorKey{}, collector)

	err = collector.Start()
	g.Expect(err).To(BeNil())

	return ctx, nil
}

func anAccessCollector(ctx context.Context) (context.Context, error) {

	//create store
	dbDir, err := os.MkdirTemp("", "db")
	store, err := store.NewStore(dbDir, log)
	g.Expect(err).To(BeNil())
	ctx = context.WithValue(ctx, storeKey{}, store)

	//retrieve clusterName to watch
	c := ctx.Value(clusterKey{}).(cluster.Cluster)
	g.Expect(c).NotTo(BeNil())

	//create access collector

	opts := collector.CollectorOpts{
		Log: log,
		Clusters: []cluster.Cluster{
			c,
		},
	}

	collector, err := accesscollector.NewAccessRulesCollector(store, opts)
	g.Expect(err).To(BeNil())
	g.Expect(collector).ToNot(BeNil())
	ctx = context.WithValue(ctx, collectorKey{}, collector)

	err = collector.Start(ctx)
	g.Expect(err).To(BeNil())

	return ctx, nil
}

type watchGvkKey struct{}

func aGvkToWatch(ctx context.Context, gvk schema.GroupVersionKind) (context.Context, error) {

	return context.WithValue(ctx, watchGvkKey{}, gvk), nil
}

type clusterKey struct{}

type numItemsKey struct{}

func aKubernetesClusterWithResourcesOfThatKind(ctx context.Context) (context.Context, error) {
	cluster := ctx.Value(clusterKey{}).(cluster.Cluster)
	runtimeClient, err := cluster.GetServerClient()
	if err != nil {
		return ctx, fmt.Errorf("could not retrieve clusterName client: %w", err)
	}
	watchGvk := ctx.Value(watchGvkKey{}).(schema.GroupVersionKind)
	numItems, err := getNumItemsByGvk(ctx, runtimeClient, watchGvk)
	if err != nil {
		return ctx, err
	}
	if numItems < 1 {
		return ctx, fmt.Errorf("not found elements in the clusterName")
	}
	ctx = context.WithValue(ctx, numItemsKey{}, int64(numItems))
	return ctx, nil
}

func getNumItemsByGvk(ctx context.Context, client client.Client, gvk schema.GroupVersionKind) (int, error) {

	list := unstructured.UnstructuredList{}

	list.SetGroupVersionKind(gvk)

	err := client.List(ctx, &list)
	if err != nil {
		return 0, err
	}
	return len(list.Items), nil
}

func watchedTheKindInTheCluster(ctx context.Context) (context.Context, error) {
	c := ctx.Value(collectorKey{}).(collector.ClusterWatcher)
	cluster := ctx.Value(clusterKey{}).(cluster.Cluster)

	isTrue := g.Eventually(func() bool {
		status, err := c.Status(cluster)
		if err != nil {
			log.Error(err, "cannot get clusterName watcher status")
			return false
		}

		//TODO move me to clusterName status instead of watcher
		return status == string(collector.ClusterWatchingStarted)
	}).Should(BeTrue())

	if !isTrue {
		return ctx, errors.New("watcher not started")
	}

	return ctx, nil
}

type storeKey struct{}

func iGotAllTheResults(ctx context.Context) (context.Context, error) {
	store := ctx.Value(storeKey{}).(store.Store)
	numDocsExpected := ctx.Value(numItemsKey{}).(int)

	isTrue := g.Eventually(func() bool {
		objs, err := store.GetObjects(ctx)
		if err != nil {
			return false
		}

		return len(objs) == numDocsExpected
	}).Should(BeTrue())

	if !isTrue {
		return ctx, fmt.Errorf("not found same number of documents")
	}
	return ctx, nil
}

func kubeEnvironment() (*rest.Config, error) {
	var err error

	useExistingCluster := true
	testEnv := &envtest.Environment{
		UseExistingCluster: &useExistingCluster,
	}

	cfg, err := testEnv.Start()
	if err != nil {
		return nil, fmt.Errorf("error on starting environment:%e", err)
	}

	return cfg, err
}
