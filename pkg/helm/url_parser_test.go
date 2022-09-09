package helm

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Service represents the elements that we need to use the client Proxy to fetch
// a URL.
type Service struct {
	Scheme    string
	Namespace string
	Name      string
	Path      string
	Port      string
}

// ParseArtifactURL takes HelmRepository Artifact URL for a remote cluster and
// returns the components of the URL.
func ParseArtifactURL(serviceURL string) (*Service, error) {
	return nil, nil
}

func TestParseService(t *testing.T) {
	artifactURL := "http://source-controller.flux-system.svc.cluster.local./demo-index.yaml"

	parsed, err := ParseArtifactURL(artifactURL)
	if err != nil {
		t.Fatal(err)
	}

	want := &Service{
		Scheme:    "http",
		Namespace: "flux-system",
		Name:      "source-controller",
		Path:      "/demo-index.yaml",
		Port:      "80",
	}
	if diff := cmp.Diff(want, parsed); diff != "" {
		t.Fatalf("failed to parse URL:\n%s", diff)
	}
}
