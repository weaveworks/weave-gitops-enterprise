package server_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-logr/logr/testr"
	"github.com/google/go-cmp/cmp"
	protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/server"
)

type assertPreviewYAMLResponseFunc func(response *protos.PreviewYAMLResponse, err error) error

func assert(fns ...assertPreviewYAMLResponseFunc) assertPreviewYAMLResponseFunc {
	return func(response *protos.PreviewYAMLResponse, err error) error {
		for _, fn := range fns {
			if assertErr := fn(response, err); assertErr != nil {
				return assertErr
			}
		}
		return nil
	}
}

func assertSuccess() assertPreviewYAMLResponseFunc {
	return func(response *protos.PreviewYAMLResponse, err error) error {
		if err != nil {
			return fmt.Errorf("Expected success but was error: %v", err)
		}
		return nil
	}
}

func assertFailure(expected error) assertPreviewYAMLResponseFunc {
	return func(response *protos.PreviewYAMLResponse, err error) error {
		if err == nil {
			return fmt.Errorf("Expected failure but got success")
		}
		diff := cmp.Diff(expected.Error(), err.Error())
		if diff != "" {
			return fmt.Errorf("Mismatch from expected failure (-want +got):\n%s", diff)
		}
		return nil
	}
}

func assertGoldenValue(expected string) assertPreviewYAMLResponseFunc {
	return assert(
		assertSuccess(),
		func(response *protos.PreviewYAMLResponse, err error) error {
			diff := cmp.Diff(expected, response.Preview.Content)
			if diff != "" {
				return fmt.Errorf("Mismatch from expected value (-want +got):\n%s", diff)
			}
			return nil
		})
}

func assertGoldenFile(goldenFile string) assertPreviewYAMLResponseFunc {
	goldenFileContents, fileErr := os.ReadFile(goldenFile)
	return assert(
		assertSuccess(),
		func(response *protos.PreviewYAMLResponse, err error) error {
			if fileErr != nil {
				return fmt.Errorf("Error reading golden file '%s': %s", goldenFile, fileErr)
			}
			expectedOutput := string(goldenFileContents)
			if assertErr := assertGoldenValue(expectedOutput)(response, err); assertErr != nil {
				return fmt.Errorf("Mismatch from golden file '%s': %v", goldenFile, assertErr)
			}
			return nil
		},
	)
}

func TestPreviewYAML(t *testing.T) {
	cases := []struct {
		name    string
		request *protos.PreviewYAMLRequest
		assert  assertPreviewYAMLResponseFunc
	}{
		{
			"GitRepository missing name",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1.GitRepositoryKind,
				Values: map[string]string{
					"name": "",
				},
			},
			assertFailure(fmt.Errorf("cannot generate preview for %q: %v", sourcev1.GitRepositoryKind, "name is required")),
		},
		{
			"GitRepository missing namespace",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1.GitRepositoryKind,
				Values: map[string]string{
					"name":      "podinfo",
					"namespace": "",
				},
			},
			assertFailure(fmt.Errorf("cannot generate preview for %q: %v", sourcev1.GitRepositoryKind, "namespace is required")),
		},
		{
			"GitRepository missing interval",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1.GitRepositoryKind,
				Values: map[string]string{
					"name":      "podinfo",
					"namespace": "flux-system",
					"interval":  "",
				},
			},
			assertFailure(fmt.Errorf("cannot generate preview for %q: %v", sourcev1.GitRepositoryKind, "interval is required")),
		},
		{
			"GitRepository invalid interval",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1.GitRepositoryKind,
				Values: map[string]string{
					"name":      "podinfo",
					"namespace": "flux-system",
					"interval":  "foo",
				},
			},
			assertFailure(fmt.Errorf("cannot generate preview for %q: %v", sourcev1.GitRepositoryKind, "invalid interval value: time: invalid duration \"foo\"")),
		},
		{
			"GitRepository missing url",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1.GitRepositoryKind,
				Values: map[string]string{
					"name":      "podinfo",
					"namespace": "flux-system",
					"interval":  "1m0s",
					"url":       "",
				},
			},
			assertFailure(fmt.Errorf("cannot generate preview for %q: %v", sourcev1.GitRepositoryKind, "url is required")),
		},
		{
			"GitRepository unsupported url scheme",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1.GitRepositoryKind,
				Values: map[string]string{
					"name":      "podinfo",
					"namespace": "flux-system",
					"interval":  "1m0s",
					"url":       "scp://domain.local",
				},
			},
			assertFailure(fmt.Errorf("cannot generate preview for %q: %v", sourcev1.GitRepositoryKind, fmt.Errorf("url scheme %q is not supported", "scp"))),
		},
		{
			"GitRepository with commit",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1.GitRepositoryKind,
				Values: map[string]string{
					"name":      "podinfo",
					"namespace": "flux-system",
					"url":       "https://github.com/stefanprodan/podinfo",
					"commit":    "c88a2f41",
					"interval":  "1m0s",
				},
			},
			assertGoldenFile("testdata/preview/source-git-commit.yaml"),
		},
		{
			"GitRepository with commit in branch",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1.GitRepositoryKind,
				Values: map[string]string{
					"name":      "podinfo",
					"namespace": "flux-system",
					"url":       "https://github.com/stefanprodan/podinfo",
					"branch":    "test",
					"commit":    "c88a2f41",
					"interval":  "1m0s",
				},
			},
			assertGoldenFile("testdata/preview/source-git-branch-commit.yaml"),
		},
		{
			"GitRepository with ref name",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1.GitRepositoryKind,
				Values: map[string]string{
					"name":      "podinfo",
					"namespace": "flux-system",
					"url":       "https://github.com/stefanprodan/podinfo",
					"ref-name":  "refs/heads/main",
					"interval":  "1m0s",
				},
			},
			assertGoldenFile("testdata/preview/source-git-refname.yaml"),
		},
		{
			"GitRepository with semver",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1.GitRepositoryKind,
				Values: map[string]string{
					"name":      "podinfo",
					"namespace": "flux-system",
					"url":       "https://github.com/stefanprodan/podinfo",
					"semver":    "v1.0.1",
					"interval":  "1m0s",
				},
			},
			assertGoldenFile("testdata/preview/source-git-semver.yaml"),
		},
		{
			"GitRepository with tag",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1.GitRepositoryKind,
				Values: map[string]string{
					"name":      "podinfo",
					"namespace": "flux-system",
					"url":       "https://github.com/stefanprodan/podinfo",
					"tag":       "test",
					"interval":  "1m0s",
				},
			},
			assertGoldenFile("testdata/preview/source-git-tag.yaml"),
		},
		{
			"GitRepository with branch",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1.GitRepositoryKind,
				Values: map[string]string{
					"name":      "podinfo",
					"namespace": "flux-system",
					"url":       "https://github.com/stefanprodan/podinfo",
					"branch":    "test",
					"interval":  "1m0s",
				},
			},
			assertGoldenFile("testdata/preview/source-git-branch.yaml"),
		},
		{
			"GitRepository with secretRef",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1.GitRepositoryKind,
				Values: map[string]string{
					"name":            "podinfo",
					"namespace":       "flux-system",
					"url":             "https://github.com/stefanprodan/podinfo",
					"secret-ref-name": "basic-access-auth",
					"branch":          "test",
					"interval":        "1m0s",
				},
			},
			assertGoldenFile("testdata/preview/source-git-secretref.yaml"),
		},
		{
			"HelmRepository missing name",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1beta2.HelmRepositoryKind,
				Values: map[string]string{
					"name": "",
				},
			},
			assertFailure(fmt.Errorf("cannot generate preview for %q: %v", sourcev1beta2.HelmRepositoryKind, "name is required")),
		},
		{
			"HelmRepository missing namespace",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1beta2.HelmRepositoryKind,
				Values: map[string]string{
					"name":      "podinfo",
					"namespace": "",
				},
			},
			assertFailure(fmt.Errorf("cannot generate preview for %q: %v", sourcev1beta2.HelmRepositoryKind, "namespace is required")),
		},
		{
			"HelmRepository missing interval",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1beta2.HelmRepositoryKind,
				Values: map[string]string{
					"name":      "podinfo",
					"namespace": "flux-system",
					"interval":  "",
				},
			},
			assertFailure(fmt.Errorf("cannot generate preview for %q: %v", sourcev1beta2.HelmRepositoryKind, "interval is required")),
		},
		{
			"HelmRepository invalid interval",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1beta2.HelmRepositoryKind,
				Values: map[string]string{
					"name":      "podinfo",
					"namespace": "flux-system",
					"interval":  "foo",
				},
			},
			assertFailure(fmt.Errorf("cannot generate preview for %q: %v", sourcev1beta2.HelmRepositoryKind, "invalid interval value: time: invalid duration \"foo\"")),
		},
		{
			"HelmRepository missing url",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1beta2.HelmRepositoryKind,
				Values: map[string]string{
					"name":      "podinfo",
					"namespace": "flux-system",
					"interval":  "1m0s",
					"url":       "",
				},
			},
			assertFailure(fmt.Errorf("cannot generate preview for %q: %v", sourcev1beta2.HelmRepositoryKind, "url is required")),
		},
		{
			"HelmRepository invalid type",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1beta2.HelmRepositoryKind,
				Values: map[string]string{
					"name":      "podinfo",
					"namespace": "flux-system",
					"interval":  "1m0s",
					"type":      "foo",
				},
			},
			assertFailure(fmt.Errorf("cannot generate preview for %q: %v", sourcev1beta2.HelmRepositoryKind, "invalid type")),
		},
		{
			"HelmRepository invalid provider",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1beta2.HelmRepositoryKind,
				Values: map[string]string{
					"name":      "podinfo",
					"namespace": "flux-system",
					"interval":  "1m0s",
					"provider":  "foo",
				},
			},
			assertFailure(fmt.Errorf("cannot generate preview for %q: %v", sourcev1beta2.HelmRepositoryKind, "invalid provider")),
		},
		{
			"HelmRepository invalid pass credentials",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1beta2.HelmRepositoryKind,
				Values: map[string]string{
					"name":             "podinfo",
					"namespace":        "flux-system",
					"url":              "https://stefanprodan.github.io/charts/podinfo",
					"secret-ref-name":  "basic-access-auth",
					"pass-credentials": "foo",
					"interval":         "1m0s",
				},
			},
			assertFailure(fmt.Errorf("cannot generate preview for %q: %v", sourcev1beta2.HelmRepositoryKind, "invalid pass-credentials value")),
		},
		{
			"HelmRepository HTTPS",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1beta2.HelmRepositoryKind,
				Values: map[string]string{
					"name":      "podinfo",
					"namespace": "flux-system",
					"interval":  "1m0s",
					"url":       "https://stefanprodan.github.io/charts/podinfo",
				},
			},
			assertGoldenFile("testdata/preview/source-helm-https.yaml"),
		},
		{
			"HelmRepository OCI",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1beta2.HelmRepositoryKind,
				Values: map[string]string{
					"name":      "podinfo",
					"namespace": "flux-system",
					"interval":  "1m0s",
					"type":      "oci",
					"url":       "oci://ghcr.io/stefanprodan/charts/podinfo",
				},
			},
			assertGoldenFile("testdata/preview/source-helm-oci.yaml"),
		},
		{
			"HelmRepository with secretRef",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1beta2.HelmRepositoryKind,
				Values: map[string]string{
					"name":            "podinfo",
					"namespace":       "flux-system",
					"url":             "https://stefanprodan.github.io/charts/podinfo",
					"secret-ref-name": "basic-access-auth",
					"interval":        "1m0s",
				},
			},
			assertGoldenFile("testdata/preview/source-helm-secretref.yaml"),
		},
		{
			"HelmRepository with secretRef and passCredentials",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1beta2.HelmRepositoryKind,
				Values: map[string]string{
					"name":             "podinfo",
					"namespace":        "flux-system",
					"url":              "https://stefanprodan.github.io/charts/podinfo",
					"secret-ref-name":  "basic-access-auth",
					"pass-credentials": "true",
					"interval":         "1m0s",
				},
			},
			assertGoldenFile("testdata/preview/source-helm-secretref-pass-credentials.yaml"),
		},
		{
			"Bucket missing name",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1beta2.BucketKind,
				Values: map[string]string{
					"name": "",
				},
			},
			assertFailure(fmt.Errorf("cannot generate preview for %q: %v", sourcev1beta2.BucketKind, "name is required")),
		},
		{
			"Bucket missing namespace",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1beta2.BucketKind,
				Values: map[string]string{
					"name":      "podinfo",
					"namespace": "",
				},
			},
			assertFailure(fmt.Errorf("cannot generate preview for %q: %v", sourcev1beta2.BucketKind, "namespace is required")),
		},
		{
			"OCIRepository missing name",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1beta2.OCIRepositoryKind,
				Values: map[string]string{
					"name": "",
				},
			},
			assertFailure(fmt.Errorf("cannot generate preview for %q: %v", sourcev1beta2.OCIRepositoryKind, "name is required")),
		},
		{
			"OCIRepository missing namespace",
			&protos.PreviewYAMLRequest{
				Kind: sourcev1beta2.OCIRepositoryKind,
				Values: map[string]string{
					"name":      "podinfo",
					"namespace": "",
				},
			},
			assertFailure(fmt.Errorf("cannot generate preview for %q: %v", sourcev1beta2.OCIRepositoryKind, "namespace is required")),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := server.NewClusterServer(server.ServerOpts{
				Logger: testr.New(t),
			})
			response, err := s.PreviewYAML(context.Background(), tc.request)

			if err := tc.assert(response, err); err != nil {
				t.Error(err)
			}
		})
	}
}
