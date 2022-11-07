package helm

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseArtifactURL(t *testing.T) {
	testCases := []struct {
		name        string
		artifactURL string
		want        *Service
		err         error
	}{
		{
			"parses correctly",
			"http://source-controller.flux-system.svc.cluster.local./demo-index.yaml",
			&Service{
				Scheme:    "http",
				Namespace: "flux-system",
				Name:      "source-controller",
				Path:      "/demo-index.yaml",
				Port:      "80",
			},
			nil,
		},
		{
			"url includes Helm index location after artifact url",
			"http://source-controller.flux-system.svc.cluster.local./demo-index.yaml/index.yaml",
			&Service{
				Scheme:    "http",
				Namespace: "flux-system",
				Name:      "source-controller",
				Path:      "/demo-index.yaml",
				Port:      "80",
			},
			nil,
		},

		{
			"wrong url",
			"http://github.com/example.repo",
			nil,
			errors.New("invalid artifact URL http://github.com/example.repo"),
		},
		{
			"empty path",
			"http://source-controller.flux-system.svc.cluster.local/",
			nil,
			errors.New("invalid artifact URL http://source-controller.flux-system.svc.cluster.local/"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseArtifactURL(tt.artifactURL)
			if tt.err != nil {
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("got wrong error:\n%s", diff)
				}
			}

			if diff := cmp.Diff(tt.want, parsed); diff != "" {
				t.Fatalf("failed to parse URL:\n%s", diff)
			}
		})
	}
}
