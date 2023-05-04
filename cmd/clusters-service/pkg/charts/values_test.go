package charts

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	fluxmeta "github.com/fluxcd/pkg/apis/meta"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/google/go-cmp/cmp"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"

	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/test"
)

const (
	testNamespace  = "testing"
	testSecretName = "https-credentials"
)

var _ ChartClient = (*HelmChartClient)(nil)

func TestValuesForChart(t *testing.T) {
	ts := httptest.NewServer(makeServeMux(t))
	defer ts.Close()
	hr := makeTestHelmRepository(ts.URL)
	c := &ChartReference{Chart: "demo-profile", Version: "0.0.1", SourceRef: referenceForRepository(hr)}
	cc := makeChartClient(t, makeTestClient(t), hr)

	values, err := cc.ValuesForChart(context.TODO(), c)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]interface{}{
		"favoriteDrink": "coffee",
	}
	if diff := cmp.Diff(want, values); diff != "" {
		t.Fatalf("failed to get values:\n%s", diff)
	}
}

func TestValuesForChart_basic_auth_via_Secret(t *testing.T) {
	fc := makeTestClient(t, makeTestSecret("test", "password"))
	ts := httptest.NewServer(basicAuthHandler(t, makeServeMux(t), "test", "password"))
	defer ts.Close()

	hr := makeTestHelmRepository(ts.URL, func(hr *sourcev1beta2.HelmRepository) {
		hr.Spec.SecretRef = &fluxmeta.LocalObjectReference{
			Name: testSecretName,
		}
	})
	c := &ChartReference{Chart: "demo-profile", Version: "0.0.1", SourceRef: referenceForRepository(hr)}
	cc := makeChartClient(t, fc, hr)

	values, err := cc.ValuesForChart(context.TODO(), c)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]interface{}{
		"favoriteDrink": "coffee",
	}
	if diff := cmp.Diff(want, values); diff != "" {
		t.Fatalf("failed to get values:\n%s", diff)
	}
}

func TestUpdateCache_with_bad_url(t *testing.T) {
	hr := makeTestHelmRepository("http://[::1]:namedport/index.yaml")
	cc := NewHelmChartClient(makeTestClient(t), testNamespace, hr)

	err := cc.UpdateCache(context.TODO())
	test.AssertErrorMatch(t, "invalid chart URL format", err)
}

func TestUpdateCache_with_missing_missing_secret_for_auth(t *testing.T) {
	fc := makeTestClient(t)
	ts := httptest.NewServer(basicAuthHandler(t, makeServeMux(t), "test", "password"))
	defer ts.Close()
	hr := makeTestHelmRepository(ts.URL, func(hr *sourcev1beta2.HelmRepository) {
		hr.Spec.SecretRef = &fluxmeta.LocalObjectReference{
			Name: testSecretName,
		}
	})
	tempDir, err := os.MkdirTemp("", "prefix")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatal(err)
		}
	})
	cc := NewHelmChartClient(fc, testNamespace, hr, WithCacheDir(tempDir))

	err = cc.UpdateCache(context.TODO())
	test.AssertErrorMatch(t, `repository authentication: secrets "https-credentials" not found`, err)
}

func TestValuesForChart_missing_version(t *testing.T) {
	ts := httptest.NewServer(makeServeMux(t))
	defer ts.Close()
	hr := makeTestHelmRepository(ts.URL)
	c := &ChartReference{Chart: "demo-profile", Version: "0.0.2", SourceRef: referenceForRepository(hr)}
	cc := makeChartClient(t, makeTestClient(t), hr)

	_, err := cc.ValuesForChart(context.TODO(), c)
	test.AssertErrorMatch(t, `chart "demo-profile" version "0.0.2" not found`, err)
}

func TestValuesForChart_missing_chart(t *testing.T) {
	ts := httptest.NewServer(makeServeMux(t, func(ri *repo.IndexFile) {
		ri.Entries["demo-profile"][0].Metadata.Version = "0.0.2"
		ri.Entries["demo-profile"][0].URLs = nil
	}))
	defer ts.Close()
	hr := makeTestHelmRepository(ts.URL)
	c := &ChartReference{Chart: "demo-profile", Version: "0.0.2", SourceRef: referenceForRepository(hr)}
	cc := makeChartClient(t, makeTestClient(t), hr)

	_, err := cc.ValuesForChart(context.TODO(), c)
	test.AssertErrorMatch(t, `chart "demo-profile" version "0.0.2" has no downloadable URLs`, err)
}

func TestFileFromChart(t *testing.T) {
	ts := httptest.NewServer(makeServeMux(t))
	defer ts.Close()
	hr := makeTestHelmRepository(ts.URL)
	c := &ChartReference{Chart: "demo-profile", Version: "0.0.1", SourceRef: referenceForRepository(hr)}
	cc := makeChartClient(t, makeTestClient(t), hr)

	values, err := cc.FileFromChart(context.TODO(), c, "values.yaml")
	if err != nil {
		t.Fatal(err)
	}

	want := []byte("favoriteDrink: coffee\n")
	if diff := cmp.Diff(want, values); diff != "" {
		t.Fatalf("failed to get values:\n%s", diff)
	}
}

func TestFileFromChart_with_unknown_name(t *testing.T) {
	ts := httptest.NewServer(makeServeMux(t))
	defer ts.Close()
	hr := makeTestHelmRepository(ts.URL)
	c := &ChartReference{Chart: "demo-profile", Version: "0.0.1", SourceRef: referenceForRepository(hr)}
	cc := makeChartClient(t, makeTestClient(t), hr)

	_, err := cc.FileFromChart(context.TODO(), c, "unknown.yaml")
	test.AssertErrorMatch(t, `failed to find file: unknown.yaml`, err)
}

func TestFileFromChart_missing_version(t *testing.T) {
	ts := httptest.NewServer(makeServeMux(t))
	defer ts.Close()
	hr := makeTestHelmRepository(ts.URL)
	c := &ChartReference{Chart: "demo-profile", Version: "0.0.2", SourceRef: referenceForRepository(hr)}
	cc := makeChartClient(t, makeTestClient(t), hr)

	_, err := cc.FileFromChart(context.TODO(), c, "values.yaml")
	test.AssertErrorMatch(t, `chart "demo-profile" version "0.0.2" not found`, err)
}

func TestFileFromChart_missing_chart(t *testing.T) {
	ts := httptest.NewServer(makeServeMux(t, func(ri *repo.IndexFile) {
		ri.Entries["demo-profile"][0].Metadata.Version = "0.0.2"
		ri.Entries["demo-profile"][0].URLs = nil
	}))
	defer ts.Close()
	hr := makeTestHelmRepository(ts.URL)
	c := &ChartReference{Chart: "demo-profile", Version: "0.0.2", SourceRef: referenceForRepository(hr)}
	cc := makeChartClient(t, makeTestClient(t), hr)

	_, err := cc.FileFromChart(context.TODO(), c, "values.yaml")
	test.AssertErrorMatch(t, `chart "demo-profile" version "0.0.2" has no downloadable URLs`, err)
}

func TestMakeHelmReleasesInLayers(t *testing.T) {
	emptyValues := map[string]interface{}{}
	testValues := map[string]interface{}{
		"testing": "value",
		"allowed": false,
	}
	dependsOn := func(name string) func(hr *helmv2.HelmRelease) {
		return func(hr *helmv2.HelmRelease) {
			hr.Spec.DependsOn = append(hr.Spec.DependsOn,
				fluxmeta.NamespacedObjectReference{Name: name})
		}
	}
	layerLabel := func(name string) func(hr *helmv2.HelmRelease) {
		return func(hr *helmv2.HelmRelease) {
			hr.ObjectMeta.Labels = map[string]string{
				LayerLabel: name,
			}
		}
	}

	hr := makeTestHelmRepository("https://example.com/charts", func(h *sourcev1beta2.HelmRepository) {
		h.ObjectMeta.Namespace = "helm-repo-ns"
	})
	layeredTests := []struct {
		name     string
		installs []ChartInstall
		want     []*helmv2.HelmRelease
	}{
		{
			name:     "install with no layers",
			installs: []ChartInstall{{Layer: "", Values: emptyValues, Ref: makeTestChartReference("test-chart", "0.0.1", hr)}},
			want:     []*helmv2.HelmRelease{makeTestHelmRelease("test-chart", hr.GetName(), hr.GetNamespace(), "test-chart", "0.0.1")},
		},
		{
			name:     "install with values",
			installs: []ChartInstall{{Layer: "", Values: testValues, Ref: makeTestChartReference("test-chart", "0.0.1", hr)}},
			want: []*helmv2.HelmRelease{makeTestHelmRelease("test-chart", "testing", hr.GetNamespace(), "test-chart", "0.0.1", func(hr *helmv2.HelmRelease) {
				hr.Spec.Values = &apiextensionsv1.JSON{Raw: []byte(`{"allowed":false,"testing":"value"}`)}
			})},
		},
		{
			name:     "install with one layer",
			installs: []ChartInstall{{Layer: "layer-0", Values: emptyValues, Ref: makeTestChartReference("test-chart", "0.0.1", hr)}},
			want:     []*helmv2.HelmRelease{makeTestHelmRelease("test-chart", "testing", hr.GetNamespace(), "test-chart", "0.0.1", layerLabel("layer-0"))},
		},
		{
			name: "install with two layers",
			installs: []ChartInstall{
				{Layer: "layer-0", Values: emptyValues, Ref: makeTestChartReference("test-chart", "0.0.1", hr)},
				{Layer: "layer-1", Values: emptyValues, Ref: makeTestChartReference("other-chart", "0.0.1", hr)}},
			want: []*helmv2.HelmRelease{
				makeTestHelmRelease("other-chart", "testing", hr.GetNamespace(), "other-chart", "0.0.1", dependsOn("test-chart"), layerLabel("layer-1")),
				makeTestHelmRelease("test-chart", "testing", hr.GetNamespace(), "test-chart", "0.0.1", layerLabel("layer-0"))},
		},
		{
			name: "install with two charts in layer",
			installs: []ChartInstall{
				{Layer: "layer-0", Values: emptyValues, Ref: makeTestChartReference("other-chart", "0.0.1", hr)},
				{Layer: "layer-0", Values: emptyValues, Ref: makeTestChartReference("new-chart", "0.0.2", hr)},
				{Layer: "layer-1", Values: emptyValues, Ref: makeTestChartReference("test-chart", "0.0.1", hr)}},
			want: []*helmv2.HelmRelease{
				makeTestHelmRelease("new-chart", "testing", hr.GetNamespace(), "new-chart", "0.0.2", layerLabel("layer-0")),
				makeTestHelmRelease("other-chart", "testing", hr.GetNamespace(), "other-chart", "0.0.1", layerLabel("layer-0")),
				makeTestHelmRelease("test-chart", "testing", hr.GetNamespace(), "test-chart", "0.0.1",
					dependsOn("other-chart"),
					dependsOn("new-chart"), layerLabel("layer-1")),
			},
		},
		{
			name: "install with empty layer and a layer",
			installs: []ChartInstall{
				{Layer: "", Values: emptyValues, Ref: makeTestChartReference("test-chart", "0.0.1", hr)},
				{Layer: "layer-1", Values: emptyValues, Ref: makeTestChartReference("other-chart", "0.0.1", hr)}},
			want: []*helmv2.HelmRelease{
				makeTestHelmRelease("other-chart", "testing", hr.GetNamespace(), "other-chart", "0.0.1", layerLabel("layer-1")),
				makeTestHelmRelease("test-chart", "testing", hr.GetNamespace(), "test-chart", "0.0.1", dependsOn("other-chart")),
			},
		},
		{
			name: "install with namespace",
			installs: []ChartInstall{
				{Layer: "", Values: emptyValues, Ref: makeTestChartReference("test-chart", "0.0.1", hr), Namespace: "test-system"}},
			want: []*helmv2.HelmRelease{
				makeTestHelmRelease("test-chart", "testing", hr.GetNamespace(), "test-chart", "0.0.1", func(hr *helmv2.HelmRelease) {
					hr.Spec.TargetNamespace = "test-system"
					hr.Spec.Install.CreateNamespace = true
				}),
			},
		},
		{
			name: "install with spec",
			installs: []ChartInstall{
				{ProfileTemplate: makeTestHelmReleaseSpec(t, func(hr map[string]interface{}) {
					hr["Spec"] = map[string]interface{}{
						"Interval": metav1.Duration{Duration: 42 * time.Hour},
					}
				}), Layer: "", Values: emptyValues, Ref: makeTestChartReference("test-chart", "0.0.1", hr), Namespace: "test-system"}},
			want: []*helmv2.HelmRelease{
				makeTestHelmRelease("test-chart", "testing", hr.GetNamespace(), "test-chart", "0.0.1", func(hr *helmv2.HelmRelease) {
					hr.Spec.TargetNamespace = "test-system"
					hr.Spec.Install.CreateNamespace = true
					hr.Spec.Interval = metav1.Duration{Duration: 42 * time.Hour}
				}),
			},
		},
	}

	for _, tt := range layeredTests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := MakeHelmReleasesInLayers(hr.GetNamespace(), tt.installs)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, r); diff != "" {
				t.Fatalf("failed to create HelmReleases:\n%s", diff)
			}
		})
	}
}

func makeTestChartReference(name, version string, hr *sourcev1beta2.HelmRepository) ChartReference {
	return ChartReference{
		Chart:     name,
		Version:   version,
		SourceRef: referenceForRepository(hr),
	}
}

func makeTestHelmReleaseSpec(t *testing.T, opts ...func(map[string]interface{})) string {
	t.Helper()
	spec := map[string]interface{}{}

	for _, opt := range opts {
		opt(spec)
	}

	b, err := yaml.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}

	return string(b)
}

func makeServeMux(t *testing.T, opts ...func(*repo.IndexFile)) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/charts/index.yaml", func(w http.ResponseWriter, req *http.Request) {
		b, err := yaml.Marshal(makeTestChartIndex(opts...))
		if err != nil {
			t.Fatal(err)
		}
		_, err = w.Write(b)
		if err != nil {
			t.Fatal(err)
		}

	})
	mux.Handle("/", http.FileServer(http.Dir("testdata")))
	return mux
}

func referenceForRepository(s *sourcev1beta2.HelmRepository) helmv2.CrossNamespaceObjectReference {
	return helmv2.CrossNamespaceObjectReference{
		APIVersion: s.TypeMeta.APIVersion,
		Kind:       s.TypeMeta.Kind,
		Name:       s.ObjectMeta.Name,
		Namespace:  s.ObjectMeta.Namespace,
	}
}

func makeTestChartIndex(opts ...func(*repo.IndexFile)) *repo.IndexFile {
	ri := &repo.IndexFile{
		APIVersion: "v1",
		Entries: map[string]repo.ChartVersions{
			"demo-profile": {
				{
					Metadata: &chart.Metadata{
						Annotations: map[string]string{
							ProfileAnnotation: "demo-profile",
						},
						Description: "Simple demo profile",
						Home:        "https://example.com/testing",
						Name:        "demo-profile",
						Sources: []string{
							"https://example.com/testing",
						},
						Version: "0.0.1",
					},
					Created: time.Now(),
					Digest:  "aaff4545f79d8b2913a10cb400ebb6fa9c77fe813287afbacf1a0b897cdffffff",
					URLs: []string{
						"/charts/demo-profile-0.1.0.tgz",
					},
				},
			},
		},
	}
	for _, o := range opts {
		o(ri)
	}
	return ri
}

func basicAuthHandler(t *testing.T, next http.Handler, user, pass string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if ok && (u == user && p == pass) {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="test"`)
		w.WriteHeader(401)
		_, err := w.Write([]byte("401 Unauthorized\n"))
		if err != nil {
			t.Fatal(err)
		}
	})
}

func makeTestClient(t *testing.T, objs ...runtime.Object) client.Client {
	t.Helper()
	s := runtime.NewScheme()
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	return fake.NewClientBuilder().WithScheme(s).WithRuntimeObjects(objs...).Build()
}

// Based on https://fluxcd.io/docs/components/source/helmrepositories/
func makeTestSecret(user, pass string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		Type: corev1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:      testSecretName,
			Namespace: testNamespace,
		},
		Data: map[string][]byte{
			"username": []byte(user),
			"password": []byte(pass),
		},
	}
}

func makeChartClient(t *testing.T, cl client.Client, hr *sourcev1beta2.HelmRepository) *HelmChartClient {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "prefix")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatal(err)
		}
	})
	cc := NewHelmChartClient(cl, testNamespace, hr, WithCacheDir(tempDir))
	if err := cc.UpdateCache(context.TODO()); err != nil {
		t.Fatal(err)
	}
	return cc
}

func makeTestHelmRelease(name, repoName, repoNS, chart, version string, opts ...func(*helmv2.HelmRelease)) *helmv2.HelmRelease {
	hr := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: repoNS,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: helmv2.GroupVersion.Identifier(),
			Kind:       helmv2.HelmReleaseKind,
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:   chart,
					Version: version,
					SourceRef: helmv2.CrossNamespaceObjectReference{
						APIVersion: sourcev1beta2.GroupVersion.Identifier(),
						Kind:       sourcev1beta2.HelmRepositoryKind,
						Name:       repoName,
						Namespace:  repoNS,
					},
				},
			},
			Install: &helmv2.Install{
				CRDs: helmv2.CreateReplace,
			},
			Upgrade: &helmv2.Upgrade{
				CRDs: helmv2.CreateReplace,
			},
			Interval: metav1.Duration{Duration: time.Minute},
			Values:   &apiextensionsv1.JSON{Raw: []byte("{}")},
		},
	}
	for _, o := range opts {
		o(hr)
	}
	return hr
}
