package collector_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"testing"
	"time"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
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
		kind       string
		errPattern string
	}{
		{
			name: "can watch helm releases",
			kind: v2beta1.HelmReleaseKind,
		},
		{
			name: "can watch kustomizations",
			kind: v1beta2.KustomizationKind,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx, err := aCollector(ctx)
			if err != nil {
				t.Fatalf(err.Error())
			}
			ctx, err = aKubernetesClusterToWatch(ctx)
			if err != nil {
				t.Fatalf(err.Error())
			}
			ctx, err = aKindToWatch(ctx, tt.kind)
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

func aCollector(ctx context.Context) (context.Context, error) {
	//store
	dbDir, err := os.MkdirTemp("", "db")
	store, err := store.NewStore(dbDir, log)
	g.Expect(err).To(BeNil())
	log.Info("created inmemory store")
	ctx = context.WithValue(ctx, storeKey{}, store)

	opts := collector.CollectorOpts{
		Log: log,
	}

	collector, err := collector.NewCollector(opts, store, nil)
	g.Expect(err).To(BeNil())
	g.Expect(collector).ToNot(BeNil())
	ctx = context.WithValue(ctx, collectorKey{}, collector)
	log.Info("collector created")

	start, err := collector.Start(ctx)
	g.Expect(err).To(BeNil())
	g.Expect(start).ToNot(BeNil())

	log.Info("collector created")
	return ctx, nil
}

type clientKey struct{}

func aKubernetesClusterToWatch(ctx context.Context) (context.Context, error) {
	//create config
	cfg, err := kubeEnvironment()
	if err != nil {
		return ctx, fmt.Errorf("could not start kube environment: %w", err)
	}
	ctx = context.WithValue(ctx, configRefKey{}, cfg)
	log.Info(fmt.Sprintf("kube environment created: %s", cfg.Host))
	clusterRef := types.NamespacedName{
		Name:      cfg.Host,
		Namespace: "default",
	}
	ctx = context.WithValue(ctx, clusterRefKey{}, clusterRef)

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
	log.Info("kube client created")
	ctx = context.WithValue(ctx, clientKey{}, runtimeClient)
	return ctx, nil
}

type watchKindKey struct{}

func aKindToWatch(ctx context.Context, kind string) (context.Context, error) {
	log.Info(fmt.Sprintf("kind to watch: %s", kind))
	return context.WithValue(ctx, watchKindKey{}, kind), nil
}

type clusterRefKey struct{}

type configRefKey struct{}

type numItemsKey struct{}

func aKubernetesClusterWithResourcesOfThatKind(ctx context.Context) (context.Context, error) {
	runtimeClient := ctx.Value(clientKey{}).(client.Client)
	watchKind := ctx.Value(watchKindKey{}).(string)
	numItems, err := getNumItemsByKind(ctx, runtimeClient, watchKind)
	if err != nil {
		return ctx, err
	}
	log.Info(fmt.Sprintf("number of resources found: %d", numItems))
	if numItems < 1 {
		return ctx, fmt.Errorf("not found elements in the cluster")
	}
	ctx = context.WithValue(ctx, numItemsKey{}, int64(numItems))
	return ctx, nil
}

func getNumItemsByKind(ctx context.Context, client client.Client, kind string) (int, error) {
	switch kind {
	case v2beta1.HelmReleaseKind:
		list := v2beta1.HelmReleaseList{}
		err := client.List(ctx, &list)
		if err != nil {
			return 0, err
		}
		return len(list.Items), nil
	case v1beta2.KustomizationKind:
		list := v1beta2.KustomizationList{}
		err := client.List(ctx, &list)
		if err != nil {
			return 0, err
		}
		return len(list.Items), nil
	default:
		return 0, fmt.Errorf("not supported: %s", kind)
	}
}

func watchedTheKindInTheCluster(ctx context.Context) (context.Context, error) {
	c := ctx.Value(collectorKey{}).(collector.ClusterWatcher)
	clusterRef := ctx.Value(clusterRefKey{}).(types.NamespacedName)
	cfg := ctx.Value(configRefKey{}).(*rest.Config)
	err := c.Watch(clusterRef, cfg, ctx, log)
	if err != nil {
		log.Info("could not add cluster:", err)
		return ctx, err
	}
	log.Info("watcher added to cluster:", clusterRef)

	isTrue := g.Eventually(func() bool {
		status, err := c.Status(clusterRef)
		if err != nil {
			log.Info("cannot get cluster watcher status:", err)
			return false
		}
		log.Info("waiting for started status:", status)
		//TODO move me to cluster status instead of watcher
		return status == string(collector.ClusterWatchingStarted)
	}).Should(BeTrue())

	if !isTrue {
		return ctx, errors.New("watcher not started")
	}

	log.Info("watcher has started")
	return ctx, nil
}

type storeKey struct{}

func iGotAllTheResults(ctx context.Context) (context.Context, error) {
	store := ctx.Value(storeKey{}).(store.Store)
	kind := ctx.Value(watchKindKey{}).(string)
	numDocsExpected := ctx.Value(numItemsKey{}).(int64)
	log.Info(fmt.Sprintf("expected num docs: '%d'", numDocsExpected))

	isTrue := g.Eventually(func() bool {
		numDocuments, err := store.CountObjects(ctx, kind)
		if err != nil {
			log.Error(err, "error counting")
			return false
		}
		log.Info(fmt.Sprintf("found num docs: '%d' ", numDocuments))

		return numDocuments == numDocsExpected
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
	log.Info("environment started")

	return cfg, err
}
