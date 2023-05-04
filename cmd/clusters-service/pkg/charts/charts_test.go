package charts

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestScanCharts_with_no_matches(t *testing.T) {
	ts := httptest.NewServer(http.FileServer(http.Dir("testdata/no_profiles")))
	profiles, err := ScanCharts(context.TODO(), makeTestHelmRepository(ts.URL), Profiles)
	if err != nil {
		t.Fatal(err)
	}

	want := []*capiv1_proto.Profile{}
	if diff := cmp.Diff(want, profiles); diff != "" {
		t.Fatalf("expected no profiles:\n%s", diff)
	}
}

func TestScanCharts_with_matching_charts(t *testing.T) {
	ts := httptest.NewServer(http.FileServer(http.Dir("testdata/with_profiles")))
	profiles, err := ScanCharts(context.TODO(), makeTestHelmRepository(ts.URL), Profiles)
	if err != nil {
		t.Fatal(err)
	}

	want := []*capiv1_proto.Profile{
		{
			Name:        "demo-profile",
			Home:        "https://helm.sh/helm",
			Sources:     []string{"https://github.com/helm/charts"},
			Description: "Simple demo profile",
			Keywords:    []string{"gitops", "demo"},
			Maintainers: []*capiv1_proto.Maintainer{
				{
					Name:  "WeaveWorks",
					Email: "maintainers@weave.works",
					Url:   "",
				},
				{
					Name:  "CNCF",
					Email: "",
					Url:   "cncf.io",
				},
			},
			Icon: "https://helm.sh/icon.png",
			AvailableVersions: []string{
				"1.1.2",
				"1.1.2-rc1",
				"1.1.0",
			},
		},
	}
	if diff := cmp.Diff(want, profiles, cmpopts.IgnoreUnexported(capiv1_proto.Profile{}, capiv1_proto.Maintainer{})); diff != "" {
		t.Fatalf("expected no diff:\n%s", diff)
	}
}

func TestScanCharts_errors(t *testing.T) {
	ts := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	profilesTests := []struct {
		name     string
		indexURL string
		wantErr  string
	}{
		{"no index file", ts.URL, "fetching profiles.*404 Not Found"},
		{"invalid chart url", "http://[::1]:namedport/index.yaml", "fetching profiles.*invalid port"},
		{"invalid scheme", "sftp://localhost:4222/index.yaml", "fetching profiles.*no provider for scheme: sftp"},
		{"empty file", ts.URL + "/invalid", "fetching profiles.*no API version specified"},
		{"invalid yaml", ts.URL + "/brokenyaml", "fetching profiles.*yaml: cannot decode"},
		{"invalid semantic versions", ts.URL + "/invalid-versions", "parsing chart demo-profile: 1..2: Invalid Semantic Version"},
	}

	for _, tt := range profilesTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ScanCharts(context.TODO(), makeTestHelmRepository(tt.indexURL), Profiles)
			test.AssertErrorMatch(t, tt.wantErr, err)
		})
	}

}

func makeTestHelmRepository(base string, opts ...func(*sourcev1beta2.HelmRepository)) *sourcev1beta2.HelmRepository {
	hr := &sourcev1beta2.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1beta2.HelmRepositoryKind,
			APIVersion: sourcev1beta2.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testing",
			Namespace: "test-ns",
		},
		Spec: sourcev1beta2.HelmRepositorySpec{
			URL:      base + "/charts",
			Interval: metav1.Duration{Duration: time.Minute * 10},
		},
		Status: sourcev1beta2.HelmRepositoryStatus{
			URL: base + "/index.yaml",
		},
	}
	for _, o := range opts {
		o(hr)
	}
	return hr
}
