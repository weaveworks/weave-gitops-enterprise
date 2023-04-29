package helm

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster/clusterfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	testingclient "k8s.io/client-go/testing"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
)

func (w fakeResponseWrapper) DoRaw(ctx context.Context) ([]byte, error) {
	return w.raw, nil
}

func (w fakeResponseWrapper) Stream(ctx context.Context) (io.ReadCloser, error) {
	return nil, nil
}

func newFakeResponseWrapper(raw []byte) fakeResponseWrapper {
	return fakeResponseWrapper{raw: raw}
}

type fakeResponseWrapper struct {
	raw []byte
}

// MakeTar creates a tar archive containing a single file chart-name/values.yaml
// with the given content.
func MakeTar(t *testing.T, chartName, content string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	hdr := &tar.Header{
		Name: fmt.Sprintf("%s/values.yaml", chartName),
		Mode: 0600,
		Size: int64(len(content)),
	}
	err := tw.WriteHeader(hdr)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tw.Write([]byte(content))
	if err != nil {
		t.Fatal(err)
	}
	err = tw.Close()
	if err != nil {
		t.Fatal(err)
	}
	err = gw.Close()
	if err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func watchForHelmChartAndUpdateStatus(fakeClient client.Client) error {
	return wait.PollImmediate(time.Second, 10*time.Second, func() (bool, error) {
		// List all helm charts in the namespace flux-system
		helmCharts := &sourcev1.HelmChartList{}
		err := fakeClient.List(context.Background(), helmCharts, client.InNamespace("flux-system"))
		if err != nil {
			return false, err
		}
		if len(helmCharts.Items) == 0 {
			return false, nil
		}
		hc := helmCharts.Items[0]
		hc.Status.URL = "http://source-controller.flux-system.svc.cluster.local./demo-index.yaml/index.yaml"
		hc.Status.Conditions = []metav1.Condition{
			{
				Type:   meta.ReadyCondition,
				Status: metav1.ConditionTrue,
			},
		}
		err = fakeClient.Status().Update(context.Background(), &hc)
		if err != nil {
			return false, nil
		}
		return true, nil
	})
}

func TestGetIndexFile(t *testing.T) {
	fakeClient := createFakeClient(t, &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "weaveworks-charts",
			Namespace: "flux-system",
		},
		Status: sourcev1.HelmRepositoryStatus{
			URL: "http://source-controller.flux-system.svc.cluster.local./demo-index.yaml/index.yaml",
		},
	})

	fakeKubeClient := kubefake.NewSimpleClientset()
	fakeKubeClient.AddProxyReactor("services", func(action testingclient.Action) (handled bool, ret rest.ResponseWrapper, err error) {
		i := &repo.IndexFile{
			APIVersion: "v1",
			Entries: map[string]repo.ChartVersions{
				"demo": {
					{
						Metadata: &chart.Metadata{
							Name:    "demo",
							Version: "0.1.0",
						},
						URLs: []string{"http://source-controller.flux-system.svc.cluster.local./demo-0.1.0.tgz"},
					},
				},
			},
		}
		data, err := yaml.Marshal(i)
		if err != nil {
			t.Fatal(err)
		}
		return true, newFakeResponseWrapper(data), nil
	})

	fakeCluster := new(clusterfakes.FakeCluster)
	fakeCluster.GetServerClientReturns(fakeClient, nil)
	fakeCluster.GetServerClientsetReturns(fakeKubeClient, nil)
	v := valuesFetcher{}

	indexFile, err := v.GetIndexFile(context.Background(), fakeCluster, types.NamespacedName{
		Namespace: "flux-system",
		Name:      "weaveworks-charts",
	}, true)

	if err != nil {
		t.Fatal(err)
	}

	if len(indexFile.Entries) != 1 {
		t.Fatal("expected one entry in index file")
	}

	if len(indexFile.Entries["demo"]) != 1 {
		t.Fatal("expected one version of demo chart")
	}

	if indexFile.Entries["demo"][0].Version != "0.1.0" {
		t.Fatal("expected version 0.1.0")
	}

}

func TestGetValues(t *testing.T) {
	fakeClient := createFakeClient(t)
	fakeKubeClient := kubefake.NewSimpleClientset()
	fakeKubeClient.AddProxyReactor("services", func(action testingclient.Action) (handled bool, ret rest.ResponseWrapper, err error) {
		data := MakeTar(t, "cert-manager", "cert-manager:\n  installCRDs: true\n")
		return true, newFakeResponseWrapper(data), nil
	})

	fakeCluster := new(clusterfakes.FakeCluster)
	fakeCluster.GetServerClientReturns(fakeClient, nil)
	fakeCluster.GetServerClientsetReturns(fakeKubeClient, nil)
	v := valuesFetcher{}

	go func() {
		err := watchForHelmChartAndUpdateStatus(fakeClient)
		if err != nil {
			t.Errorf("error watching for helm chart: %s", err)
		}
	}()

	values, err := v.GetValuesFile(
		context.Background(),
		fakeCluster,
		types.NamespacedName{Namespace: "flux-system", Name: "weaveworks-charts"},
		Chart{Name: "cert-manager", Version: "0.0.8"},
		true,
	)

	if err != nil {
		t.Fatal(err)
	}

	if string(values) != "cert-manager:\n  installCRDs: true\n" {
		t.Fatalf("expected %q, got %q", "cert-manager:\n  installCRDs: true\n", string(values))
	}
}

func createFakeClient(t *testing.T, clusterState ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		sourcev1.AddToScheme,
	}
	err := schemeBuilder.AddToScheme(scheme)
	if err != nil {
		t.Fatal(err)
	}

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(clusterState...).
		Build()

	return c
}

func TestValuesFetching(t *testing.T) {
	t.Skip("integration test")

	// Integration test innit
	config := ctrl.GetConfigOrDie()
	scheme, err := kube.CreateScheme()
	if err != nil {
		t.Fatal(err)
	}
	cl, err := cluster.NewSingleCluster("test", config, scheme, kube.UserPrefixes{})
	if err != nil {
		t.Fatal(err)
	}

	f := NewValuesFetcher()
	index, err := f.GetIndexFile(context.Background(), cl, types.NamespacedName{Name: "weaveworks-charts", Namespace: "flux-system"}, true)
	if err != nil {
		t.Fatal(err)
	}
	certManager := index.Entries["cert-manager"][0]
	fmt.Printf("cert-manager: %v\n", certManager)
	if certManager.Version != "0.0.8" {
		t.Fatalf("expected cert-manager version 0.0.8 got %s", certManager.Version)
	}

	data, err := f.GetValuesFile(
		context.TODO(),
		cl,
		types.NamespacedName{Namespace: "flux-system", Name: "weaveworks-charts"},
		Chart{
			Name:    "cert-manager",
			Version: "0.0.8",
		},
		true,
	)

	if err != nil {
		t.Fatal(err)
	}

	if len(data) == 0 {
		t.Fatal("no data")
	}

	expected := "cert-manager:\n  installCRDs: true\n"
	if string(data) != expected {
		t.Fatalf("expected %q, got %q", expected, string(data))
	}
}
