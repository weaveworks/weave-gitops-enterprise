package collector

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	"go.uber.org/zap/zapcore"
	"k8s.io/client-go/rest"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector/clusters/clustersfakes"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store/storefakes"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster/clusterfakes"
	l "github.com/weaveworks/weave-gitops/core/logger"
)

func TestStart(t *testing.T) {
	g := NewGomegaWithT(t)
	log, loggerPath := newLoggerWithLevel(t, "INFO")

	fakeStore := &storefakes.FakeStore{}
	cm := &clustersfakes.FakeSubscriber{}
	cmw := &clustersfakes.FakeSubscription{}
	cm.SubscribeReturns(cmw)
	opts := CollectorOpts{
		Log:            log,
		Clusters:       cm,
		NewWatcherFunc: newFakeWatcher,
		ServiceAccount: ImpersonateServiceAccount{
			Namespace: "flux-system",
			Name:      "collector",
		},
	}
	collector, err := newWatchingCollector(opts)
	g.Expect(err).To(BeNil())
	g.Expect(collector).NotTo(BeNil())

	tests := []struct {
		name                string
		clusters            []cluster.Cluster
		expectedLogError    string
		notExpectedLog      string
		expectedNumClusters int
	}{
		{
			name:                "can start collector for empty collection",
			clusters:            []cluster.Cluster{},
			expectedLogError:    "",
			expectedNumClusters: 0,
		},
		{
			name:                "can start collector with not watchable clusters",
			clusters:            []cluster.Cluster{makeInvalidFakeCluster("test-cluster")},
			expectedLogError:    "cannot watch cluster",
			notExpectedLog:      "watching cluster",
			expectedNumClusters: 0,
		},
		{
			name:                "can start collector with watchable clusters",
			clusters:            []cluster.Cluster{makeValidFakeCluster("test-cluster")},
			expectedLogError:    "",
			expectedNumClusters: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm.GetClustersReturns(tt.clusters)
			ctx, cancel := context.WithCancel(context.TODO())
			defer cancel()
			go func() {
				g.Expect(collector.Start(ctx)).To(Succeed())
			}()

			// assert any error for an individual cluster has been
			// logged, and the corresponding success message has not
			// been logged.
			if tt.expectedLogError != "" || tt.notExpectedLog != "" {
				logs, err := os.ReadFile(loggerPath)
				g.Expect(err).To(BeNil())
				logss := string(logs)
				if tt.expectedLogError != "" {
					g.Expect(logss).To(MatchRegexp(tt.expectedLogError))
				}
				// NB this will only work if there's no cluster that can succeed!
				if tt.notExpectedLog != "" {
					g.Expect(logss).NotTo(MatchRegexp(tt.notExpectedLog))
				}
			}

			g.Expect(err).To(BeNil())
			g.Expect(fakeStore).NotTo(BeNil())
			g.Expect(len(collector.clusterWatchers)).To(Equal(tt.expectedNumClusters))
		})
	}
}

func makeInvalidFakeCluster(name string) cluster.Cluster {
	cluster := new(clusterfakes.FakeCluster)
	cluster.GetNameReturns(name)
	return cluster
}

func makeValidFakeCluster(name string) cluster.Cluster {
	config := &rest.Config{
		Host: "http://idontexist",
	}

	cluster := clusterfakes.FakeCluster{}
	cluster.GetNameReturns(name)
	cluster.GetServerConfigReturns(config, nil)
	return &cluster
}

func TestStop(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)

	routineCountBefore := runtime.NumGoroutine()

	cm := &clustersfakes.FakeSubscriber{}
	cmw := &clustersfakes.FakeSubscription{}
	cm.SubscribeReturns(cmw)

	opts := CollectorOpts{
		Log:            log,
		Clusters:       cm,
		NewWatcherFunc: newFakeWatcher,
		ServiceAccount: ImpersonateServiceAccount{
			Namespace: "flux-system",
			Name:      "collector",
		},
	}
	collector, err := newWatchingCollector(opts)
	g.Expect(err).To(BeNil())
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	go func() {
		g.Expect(collector.Start(ctx)).To(Succeed())
	}()
	g.Expect(err).To(BeNil())

	cancel()
	g.Eventually(runtime.NumGoroutine, "5s", "0.2s").Should(Equal(routineCountBefore), "no leaked goroutines")
}

func TestClusterWatcher_Watch(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	opts := CollectorOpts{
		Log:            log,
		NewWatcherFunc: newFakeWatcher,
		ServiceAccount: ImpersonateServiceAccount{
			Namespace: "flux-system",
			Name:      "collector",
		},
	}
	collector, err := newWatchingCollector(opts)
	g.Expect(err).To(BeNil())
	g.Expect(collector).NotTo(BeNil())

	c := makeValidFakeCluster("testcluster")

	tests := []struct {
		name       string
		cluster    cluster.Cluster
		errPattern string
	}{
		{
			name:       "can watch cluster",
			cluster:    c,
			errPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = collector.watch(tt.cluster)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(err).To(BeNil())
			status, err := collector.Status(tt.cluster.GetName())
			g.Expect(err).To(BeNil())
			g.Expect(ClusterWatchingStarted).To(BeIdenticalTo(status))
		})
	}
}

func TestClusterWatcher_Unwatch(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	clusterName := "testCluster"

	tests := []struct {
		name        string
		clusterName string
		errPattern  string
	}{
		{
			name:        "unwatch empty cluster throws error",
			clusterName: "",
			errPattern:  "cluster name is empty",
		},
		{
			name:        "unwatch non-existing cluster throws error",
			clusterName: "idontexist",
			errPattern:  "cluster watcher not found",
		},
		{
			name:        "unwatch existing cluster unwatches it",
			clusterName: clusterName,
			errPattern:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := CollectorOpts{
				Log:            log,
				NewWatcherFunc: newFakeWatcher,
				ServiceAccount: ImpersonateServiceAccount{
					Namespace: "flux-system",
					Name:      "collector",
				},
			}
			collector, err := newWatchingCollector(opts)
			g.Expect(err).To(BeNil())
			g.Expect(collector).NotTo(BeNil())

			c := makeValidFakeCluster(clusterName)
			g.Expect(collector.watch(c)).To(Succeed())

			if tt.errPattern == "" {
				s, err := collector.Status(tt.clusterName)
				g.Expect(err).To(BeNil())
				g.Expect(s).To(BeIdenticalTo(ClusterWatchingStarted))
			}
			err = collector.unwatch(tt.clusterName)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			} else {
				// fetching the status after it's unwatched should
				// error, since it will have been forgotten.
				_, err := collector.Status(tt.clusterName)
				g.Expect(err).To(HaveOccurred())
			}
		})
	}
}

func TestClusterWatcher_Status(t *testing.T) {
	g := NewGomegaWithT(t)
	log := testr.New(t)
	options := CollectorOpts{
		Log:            log,
		NewWatcherFunc: newFakeWatcher,
		ServiceAccount: ImpersonateServiceAccount{
			Namespace: "flux-system",
			Name:      "collector",
		},
	}
	collector, err := newWatchingCollector(options)
	g.Expect(err).To(BeNil())
	g.Expect(collector).NotTo(BeNil())
	g.Expect(len(collector.clusterWatchers)).To(Equal(0))
	existingClusterName := "test"
	c := makeValidFakeCluster(existingClusterName)
	err = collector.watch(c)
	g.Expect(err).To(BeNil())

	tests := []struct {
		name           string
		clusterName    string
		errPattern     string
		expectedStatus string
	}{
		{
			name:        "cannot get status for non existing cluster",
			clusterName: "",
			errPattern:  "cluster name is empty",
		},
		{
			name:        "could not get status for non existing cluster",
			clusterName: "dontexist",
			errPattern:  "cluster not found",
		},
		{
			name:           "could get status for existing cluster",
			clusterName:    existingClusterName,
			expectedStatus: string(ClusterWatchingStarted),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := collector.Status(tt.clusterName)
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(status).To(Equal(tt.expectedStatus))
		})
	}
}

func newFakeWatcher(clusterName string, config *rest.Config) (Starter, error) {
	log.Info("created fake watcher")
	return &fakeWatcher{log: log}, nil
}

type fakeWatcher struct {
	log logr.Logger
}

func (f *fakeWatcher) Start(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// newLoggerWithLevel creates a logger and a path to the file it
// writes to, so you can check the contents of the log during the
// test.
func newLoggerWithLevel(t *testing.T, logLevel string) (logr.Logger, string) {
	g := NewGomegaWithT(t)

	tmp, err := os.MkdirTemp("", "query-server-test")
	g.Expect(err).ShouldNot(HaveOccurred())
	path := filepath.Join(tmp, "log")

	level, err := zapcore.ParseLevel(logLevel)
	g.Expect(err).ShouldNot(HaveOccurred())
	cfg := l.BuildConfig(
		l.WithLogLevel(level),
		l.WithMode(false),
		l.WithOutAndErrPaths("stdout", "stderr"),
		l.WithOutAndErrPaths(path, path),
	)

	log, err := l.NewFromConfig(cfg)
	g.Expect(err).NotTo(HaveOccurred())

	t.Cleanup(func() {
		if strings.HasPrefix(path, os.TempDir()) {
			err := os.Remove(path)
			if err != nil {
				t.Fatal(err)
			}
		}
	})

	return log, path
}

func Test_makeImpersonateConfig(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		name               string
		config             *rest.Config
		namespace          string
		serviceAccountName string
		errPattern         string
	}{
		{
			name: "cannot create impersonation config if invalid params",
			config: &rest.Config{
				Host: "http://idontexist",
			},
			errPattern: "service acccount cannot be empty",
		},
		{
			name: "cannot create impersonation config if invalid params",
			config: &rest.Config{
				Host: "http://idontexist",
			},
			namespace:          "flux-system",
			serviceAccountName: "collector",
			errPattern:         "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := makeServiceAccountImpersonationConfig(tt.config, tt.namespace, tt.serviceAccountName)
			if err != nil {
				return
			}
			if tt.errPattern != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.errPattern)))
				return
			}
			g.Expect(config.Impersonate.UserName).To(ContainSubstring(tt.serviceAccountName))
		})
	}
}
