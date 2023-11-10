package preview_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-logr/logr/testr"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/mock"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/preview"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/preview"
	"google.golang.org/protobuf/types/known/durationpb"
	"k8s.io/utils/ptr"
)

// assertFunc is a function type for assertion functions. Functions that
// implement this are meant to inspect the result of the gRPC call and
// return nil for assertion success or an error for assertion failure.
type assertFunc func(result string, err error) error

// assert is used to combine and execute multiple assertion functions
func assert(fns ...assertFunc) assertFunc {
	return func(result string, err error) error {
		for _, fn := range fns {
			if assertErr := fn(result, err); assertErr != nil {
				return assertErr
			}
		}
		return nil
	}
}

func assertSuccess() assertFunc {
	return func(result string, err error) error {
		if err != nil {
			return fmt.Errorf("Expected success but got an error: %v", err)
		}
		return nil
	}
}

func assertFailure(expected error) assertFunc {
	return func(result string, err error) error {
		if err == nil {
			return errors.New("Expected an error but got success")
		}
		diff := cmp.Diff(expected.Error(), err.Error())
		if diff != "" {
			return fmt.Errorf("Mismatch from expected failure (-want +got):\n%s", diff)
		}
		return nil
	}
}

func assertGoldenValue(expected string) assertFunc {
	return assert(
		assertSuccess(),
		func(result string, err error) error {
			diff := cmp.Diff(expected, result)
			if diff != "" {
				return fmt.Errorf("Mismatch from expected value (-want +got):\n%s", diff)
			}
			return nil
		})
}

func assertGoldenFile(goldenFile string) assertFunc {
	goldenFileContents, fileErr := os.ReadFile(goldenFile)
	return assert(
		assertSuccess(),
		func(result string, err error) error {
			if fileErr != nil {
				return fmt.Errorf("Error reading golden file '%s': %s", goldenFile, fileErr)
			}
			expectedOutput := string(goldenFileContents)
			if assertErr := assertGoldenValue(expectedOutput)(result, err); assertErr != nil {
				return fmt.Errorf("Mismatch from golden file '%s': %v", goldenFile, assertErr)
			}
			return nil
		},
	)
}

func TestGetYAML_GitRepository(t *testing.T) {
	cases := []struct {
		name   string
		obj    *pb.GitRepository
		assert assertFunc
	}{
		{
			"missing name",
			&pb.GitRepository{},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1.GitRepositoryKind, "name is required")),
		},
		{
			"missing namespace",
			&pb.GitRepository{
				Name: "podinfo",
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1.GitRepositoryKind, "namespace is required")),
		},
		{
			"missing url",
			&pb.GitRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1.GitRepositoryKind, "url is required")),
		},
		{
			"missing interval",
			&pb.GitRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Url:       "https://github.com/stefanprodan/podinfo",
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1.GitRepositoryKind, "invalid interval value")),
		},
		{
			"invalid interval",
			&pb.GitRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Url:       "https://github.com/stefanprodan/podinfo",
				Interval:  &durationpb.Duration{Seconds: 0},
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1.GitRepositoryKind, "invalid interval value")),
		},
		{
			"unsupported url scheme",
			&pb.GitRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Url:       "scp://domain.local",
				Interval:  &durationpb.Duration{Seconds: 60},
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1.GitRepositoryKind, fmt.Errorf("url scheme %q is not supported", "scp"))),
		},
		{
			"missing git ref",
			&pb.GitRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Url:       "https://github.com/stefanprodan/podinfo",
				Interval:  &durationpb.Duration{Seconds: 60},
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1.GitRepositoryKind, "a Git ref is required")),
		},
		{
			"commit",
			&pb.GitRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Url:       "https://github.com/stefanprodan/podinfo",
				Commit:    ptr.To("c88a2f41"),
				Interval:  &durationpb.Duration{Seconds: 60},
			},
			assertGoldenFile("testdata/source-git-commit.yaml"),
		},
		{
			"commit in branch",
			&pb.GitRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Url:       "https://github.com/stefanprodan/podinfo",
				Branch:    ptr.To("test"),
				Commit:    ptr.To("c88a2f41"),
				Interval:  &durationpb.Duration{Seconds: 60},
			},
			assertGoldenFile("testdata/source-git-branch-commit.yaml"),
		},
		{
			"ref name",
			&pb.GitRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Url:       "https://github.com/stefanprodan/podinfo",
				RefName:   ptr.To("refs/heads/main"),
				Interval:  &durationpb.Duration{Seconds: 60},
			},
			assertGoldenFile("testdata/source-git-refname.yaml"),
		},
		{
			"semver",
			&pb.GitRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Url:       "https://github.com/stefanprodan/podinfo",
				Interval:  &durationpb.Duration{Seconds: 60},
				Semver:    ptr.To("v1.0.1"),
			},
			assertGoldenFile("testdata/source-git-semver.yaml"),
		},
		{
			"tag",
			&pb.GitRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Url:       "https://github.com/stefanprodan/podinfo",
				Interval:  &durationpb.Duration{Seconds: 60},
				Tag:       ptr.To("test"),
			},
			assertGoldenFile("testdata/source-git-tag.yaml"),
		},
		{
			"branch",
			&pb.GitRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Url:       "https://github.com/stefanprodan/podinfo",
				Branch:    ptr.To("test"),
				Interval:  &durationpb.Duration{Seconds: 60},
			},
			assertGoldenFile("testdata/source-git-branch.yaml"),
		},
		{
			"secretRef",
			&pb.GitRepository{
				Name:          "podinfo",
				Namespace:     "flux-system",
				Url:           "https://github.com/stefanprodan/podinfo",
				SecretRefName: ptr.To("basic-access-auth"),
				Branch:        ptr.To("test"),
				Interval:      &durationpb.Duration{Seconds: 60},
			},
			assertGoldenFile("testdata/source-git-secretref.yaml"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := preview.NewPreviewServiceServer(preview.ServerOpts{
				Logger: testr.New(t),
			})

			b, err := json.Marshal(tc.obj)
			if err != nil {
				t.Errorf("failed to encode object as JSON: %v", err)
			}

			request := &pb.GetYAMLRequest{
				Resource: &pb.TypedObject{
					Type:   sourcev1.GitRepositoryKind,
					Object: string(b),
				},
			}
			response, err := s.GetYAML(context.Background(), request)

			if err := tc.assert(response.GetFile().GetContent(), err); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestGetYAML_HelmRepository(t *testing.T) {
	cases := []struct {
		name   string
		obj    *pb.HelmRepository
		assert assertFunc
	}{
		{
			"missing name",
			&pb.HelmRepository{},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.HelmRepositoryKind, "name is required")),
		},
		{
			"missing namespace",
			&pb.HelmRepository{
				Name: "podinfo",
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.HelmRepositoryKind, "namespace is required")),
		},
		{
			"missing interval",
			&pb.HelmRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.HelmRepositoryKind, "invalid interval value")),
		},
		{
			"invalid interval",
			&pb.HelmRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Interval:  &durationpb.Duration{Seconds: 0},
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.HelmRepositoryKind, "invalid interval value")),
		},
		{
			"missing url",
			&pb.HelmRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Interval:  &durationpb.Duration{Seconds: 60},
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.HelmRepositoryKind, "url is required")),
		},
		{
			"invalid type",
			&pb.HelmRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Url:       "https://stefanprodan.github.io/charts/podinfo",
				Interval:  &durationpb.Duration{Seconds: 60},
				Type:      ptr.To("foo"),
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.HelmRepositoryKind, "invalid type")),
		},
		{
			"invalid provider",
			&pb.HelmRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Interval:  &durationpb.Duration{Seconds: 60},
				Url:       "https://stefanprodan.github.io/charts/podinfo",
				Provider:  ptr.To("foo"),
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.HelmRepositoryKind, "invalid provider")),
		},
		{
			"https",
			&pb.HelmRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Url:       "https://stefanprodan.github.io/charts/podinfo",
				Interval:  &durationpb.Duration{Seconds: 60},
			},
			assertGoldenFile("testdata/source-helm-https.yaml"),
		},
		{
			"oci",
			&pb.HelmRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Url:       "oci://ghcr.io/stefanprodan/charts/podinfo",
				Interval:  &durationpb.Duration{Seconds: 60},
				Type:      ptr.To("oci"),
			},
			assertGoldenFile("testdata/source-helm-oci.yaml"),
		},
		{
			"secretRef",
			&pb.HelmRepository{
				Name:          "podinfo",
				Namespace:     "flux-system",
				Url:           "https://stefanprodan.github.io/charts/podinfo",
				Interval:      &durationpb.Duration{Seconds: 60},
				SecretRefName: ptr.To("basic-access-auth"),
			},
			assertGoldenFile("testdata/source-helm-secretref.yaml"),
		},
		{
			"secretRef and passCredentials",
			&pb.HelmRepository{
				Name:            "podinfo",
				Namespace:       "flux-system",
				Url:             "https://stefanprodan.github.io/charts/podinfo",
				Interval:        &durationpb.Duration{Seconds: 60},
				SecretRefName:   ptr.To("basic-access-auth"),
				PassCredentials: ptr.To(true),
			},
			assertGoldenFile("testdata/source-helm-secretref-pass-credentials.yaml"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := preview.NewPreviewServiceServer(preview.ServerOpts{
				Logger: testr.New(t),
			})

			b, err := json.Marshal(tc.obj)
			if err != nil {
				t.Errorf("failed to encode object as JSON: %v", err)
			}
			request := &pb.GetYAMLRequest{
				Resource: &pb.TypedObject{
					Type:   sourcev1beta2.HelmRepositoryKind,
					Object: string(b),
				},
			}
			response, err := s.GetYAML(context.Background(), request)

			if err := tc.assert(response.GetFile().GetContent(), err); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestGetYAML_Bucket(t *testing.T) {
	cases := []struct {
		name   string
		obj    *pb.Bucket
		assert assertFunc
	}{
		{
			"missing name",
			&pb.Bucket{},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.BucketKind, "name is required")),
		},
		{
			"missing namespace",
			&pb.Bucket{
				Name: "podinfo",
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.BucketKind, "namespace is required")),
		},
		{
			"missing interval",
			&pb.Bucket{
				Name:      "podinfo",
				Namespace: "flux-system",
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.BucketKind, "invalid interval value")),
		},
		{
			"invalid interval",
			&pb.Bucket{
				Name:      "podinfo",
				Namespace: "flux-system",
				Interval:  &durationpb.Duration{Seconds: 0},
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.BucketKind, "invalid interval value")),
		},
		{
			"missing bucket name",
			&pb.Bucket{
				Name:      "podinfo",
				Namespace: "flux-system",
				Interval:  &durationpb.Duration{Seconds: 60},
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.BucketKind, "bucket name is required")),
		},
		{
			"missing endpoint",
			&pb.Bucket{
				Name:       "podinfo",
				Namespace:  "flux-system",
				Interval:   &durationpb.Duration{Seconds: 60},
				BucketName: "test",
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.BucketKind, "endpoint is required")),
		},
		{
			"invalid provider",
			&pb.Bucket{
				Name:       "podinfo",
				Namespace:  "flux-system",
				Interval:   &durationpb.Duration{Seconds: 60},
				BucketName: "test",
				Endpoint:   "minio.example.com",
				Provider:   ptr.To("foo"),
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.BucketKind, "invalid provider")),
		},
		{
			"generic provider requires a secretRef",
			&pb.Bucket{
				Name:       "podinfo",
				Namespace:  "flux-system",
				Interval:   &durationpb.Duration{Seconds: 60},
				BucketName: "test",
				Endpoint:   "minio.example.com",
				Provider:   ptr.To(sourcev1beta2.GenericBucketProvider),
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.BucketKind, "generic provider requires a secretRef")),
		},
		{
			"generic",
			&pb.Bucket{
				Name:          "podinfo",
				Namespace:     "flux-system",
				BucketName:    "test",
				Endpoint:      "minio.example.com",
				Interval:      &durationpb.Duration{Seconds: 60},
				Provider:      ptr.To(sourcev1beta2.GenericBucketProvider),
				SecretRefName: ptr.To("minio-bucket-secret"),
			},
			assertGoldenFile("testdata/source-bucket-generic.yaml"),
		},
		{
			"region",
			&pb.Bucket{
				Name:       "podinfo",
				Namespace:  "flux-system",
				BucketName: "test",
				Endpoint:   "minio.example.com",
				Interval:   &durationpb.Duration{Seconds: 60},
				Region:     ptr.To("us-east-1"),
			},
			assertGoldenFile("testdata/source-bucket-region.yaml"),
		},
		{
			"insecure",
			&pb.Bucket{
				Name:       "podinfo",
				Namespace:  "flux-system",
				BucketName: "test",
				Endpoint:   "minio.example.com",
				Interval:   &durationpb.Duration{Seconds: 60},
				Insecure:   ptr.To(true),
			},
			assertGoldenFile("testdata/source-bucket-insecure.yaml"),
		},
		{
			"secretRef",
			&pb.Bucket{
				Name:          "podinfo",
				Namespace:     "flux-system",
				BucketName:    "test",
				Endpoint:      "minio.example.com",
				Interval:      &durationpb.Duration{Seconds: 60},
				SecretRefName: ptr.To("minio-bucket-secret"),
			},
			assertGoldenFile("testdata/source-bucket-secretref.yaml"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := preview.NewPreviewServiceServer(preview.ServerOpts{
				Logger: testr.New(t),
			})

			b, err := json.Marshal(tc.obj)
			if err != nil {
				t.Errorf("failed to encode object as JSON: %v", err)
			}
			request := &pb.GetYAMLRequest{
				Resource: &pb.TypedObject{
					Type:   sourcev1beta2.BucketKind,
					Object: string(b),
				},
			}
			response, err := s.GetYAML(context.Background(), request)

			if err := tc.assert(response.GetFile().GetContent(), err); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestGetYAML_OCIRepository(t *testing.T) {
	cases := []struct {
		name   string
		obj    *pb.OCIRepository
		assert assertFunc
	}{
		{
			"missing name",
			&pb.OCIRepository{},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.OCIRepositoryKind, "name is required")),
		},
		{
			"missing namespace",
			&pb.OCIRepository{
				Name: "podinfo",
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.OCIRepositoryKind, "namespace is required")),
		},
		{
			"missing interval",
			&pb.OCIRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Url:       "oci://ghcr.io/stefanprodan/manifests/podinfo",
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.OCIRepositoryKind, "invalid interval value")),
		},
		{
			"invalid interval",
			&pb.OCIRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Url:       "oci://ghcr.io/stefanprodan/manifests/podinfo",
				Interval:  &durationpb.Duration{Seconds: 0},
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.OCIRepositoryKind, "invalid interval value")),
		},
		{
			"missing url",
			&pb.OCIRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Interval:  &durationpb.Duration{Seconds: 60},
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.OCIRepositoryKind, "url is required")),
		},
		{
			"invalid url scheme",
			&pb.OCIRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Url:       "https://ghcr.io/stefanprodan/manifests/podinfo",
				Interval:  &durationpb.Duration{Seconds: 60},
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.OCIRepositoryKind, "url scheme must be set to \"oci\"")),
		},
		{
			"missing ref",
			&pb.OCIRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Url:       "oci://ghcr.io/stefanprodan/manifests/podinfo",
				Interval:  &durationpb.Duration{Seconds: 60},
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.OCIRepositoryKind, "ref is required")),
		},
		{
			"invalid provider",
			&pb.OCIRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Url:       "oci://ghcr.io/stefanprodan/manifests/podinfo",
				Interval:  &durationpb.Duration{Seconds: 60},
				Provider:  ptr.To("foo"),
			},
			assertFailure(fmt.Errorf("failed to generate YAML for %q: %v", sourcev1beta2.OCIRepositoryKind, "invalid provider")),
		},
		{
			"generic",
			&pb.OCIRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Url:       "oci://ghcr.io/stefanprodan/manifests/podinfo",
				Interval:  &durationpb.Duration{Seconds: 60},
				Tag:       ptr.To("test"),
				Provider:  ptr.To(sourcev1beta2.GenericOCIProvider),
				Insecure:  ptr.To(true),
			},
			assertGoldenFile("testdata/source-oci-generic.yaml"),
		},
		{
			"tag",
			&pb.OCIRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Url:       "oci://ghcr.io/stefanprodan/manifests/podinfo",
				Interval:  &durationpb.Duration{Seconds: 60},
				Tag:       ptr.To("test"),
			},
			assertGoldenFile("testdata/source-oci-tag.yaml"),
		},
		{
			"semver",
			&pb.OCIRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Url:       "oci://ghcr.io/stefanprodan/manifests/podinfo",
				Interval:  &durationpb.Duration{Seconds: 60},
				Semver:    ptr.To("v1.0.1"),
			},
			assertGoldenFile("testdata/source-oci-semver.yaml"),
		},
		{
			"digest",
			&pb.OCIRepository{
				Name:      "podinfo",
				Namespace: "flux-system",
				Url:       "oci://ghcr.io/stefanprodan/manifests/podinfo",
				Interval:  &durationpb.Duration{Seconds: 60},
				Digest:    ptr.To("sha256:a9561eb1b190625c9adb5a9513e72c4dedafc1cb2d4c5236c9a6957ec7dfd5a9"),
			},
			assertGoldenFile("testdata/source-oci-digest.yaml"),
		},
		{
			"secretRef",
			&pb.OCIRepository{
				Name:          "podinfo",
				Namespace:     "flux-system",
				Url:           "oci://ghcr.io/stefanprodan/manifests/podinfo",
				Interval:      &durationpb.Duration{Seconds: 60},
				Tag:           ptr.To("test"),
				SecretRefName: ptr.To("oci-registry"),
			},
			assertGoldenFile("testdata/source-oci-secretref.yaml"),
		},
		{
			"serviceAccount",
			&pb.OCIRepository{
				Name:               "podinfo",
				Namespace:          "flux-system",
				Url:                "oci://ghcr.io/stefanprodan/manifests/podinfo",
				Interval:           &durationpb.Duration{Seconds: 60},
				Tag:                ptr.To("test"),
				ServiceAccountName: ptr.To("oci-registry-sa"),
			},
			assertGoldenFile("testdata/source-oci-serviceaccount.yaml"),
		},
		{
			"certSecretRef",
			&pb.OCIRepository{
				Name:              "podinfo",
				Namespace:         "flux-system",
				Url:               "oci://ghcr.io/stefanprodan/manifests/podinfo",
				Interval:          &durationpb.Duration{Seconds: 60},
				Tag:               ptr.To("test"),
				CertSecretRefName: ptr.To("oci-registry"),
			},
			assertGoldenFile("testdata/source-oci-certsecretref.yaml"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := preview.NewPreviewServiceServer(preview.ServerOpts{
				Logger: testr.New(t),
			})

			b, err := json.Marshal(tc.obj)
			if err != nil {
				t.Errorf("failed to encode object as JSON: %v", err)
			}
			request := &pb.GetYAMLRequest{
				Resource: &pb.TypedObject{
					Type:   sourcev1beta2.OCIRepositoryKind,
					Object: string(b),
				},
			}
			response, err := s.GetYAML(context.Background(), request)

			if err := tc.assert(response.GetFile().GetContent(), err); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestCreatePullRequest_ValidationErrors(t *testing.T) {
	cases := []struct {
		name    string
		request *pb.CreatePullRequestRequest
		assert  assertFunc
	}{
		{
			"missing repositoryUrl",
			&pb.CreatePullRequestRequest{},
			assertFailure(fmt.Errorf("failed to create pull request: %v", "repository URL is required")),
		},
		{
			"missing headBranch",
			&pb.CreatePullRequestRequest{
				RepositoryUrl: "https://github.com/weaveworks/weave-gitops",
			},
			assertFailure(fmt.Errorf("failed to create pull request: %v", "head branch is required")),
		},
		{
			"missing baseBranch",
			&pb.CreatePullRequestRequest{
				RepositoryUrl: "https://github.com/weaveworks/weave-gitops",
				HeadBranch:    "add-git-repo",
			},
			assertFailure(fmt.Errorf("failed to create pull request: %v", "base branch is required")),
		},
		{
			"missing title",
			&pb.CreatePullRequestRequest{
				RepositoryUrl: "https://github.com/weaveworks/weave-gitops",
				HeadBranch:    "add-git-repo",
				BaseBranch:    "main",
			},
			assertFailure(fmt.Errorf("failed to create pull request: %v", "title is required")),
		},
		{
			"missing description",
			&pb.CreatePullRequestRequest{
				RepositoryUrl: "https://github.com/weaveworks/weave-gitops",
				HeadBranch:    "add-git-repo",
				BaseBranch:    "main",
				Title:         "Add GitRepository",
			},
			assertFailure(fmt.Errorf("failed to create pull request: %v", "description is required")),
		},
		{
			"missing commit message",
			&pb.CreatePullRequestRequest{
				RepositoryUrl: "https://github.com/weaveworks/weave-gitops",
				HeadBranch:    "add-git-repo",
				BaseBranch:    "main",
				Title:         "Add GitRepository",
				Description:   "Adding GitRepository for app",
			},
			assertFailure(fmt.Errorf("failed to create pull request: %v", "commit message is required")),
		},
		{
			"missing resource",
			&pb.CreatePullRequestRequest{
				RepositoryUrl: "https://github.com/weaveworks/weave-gitops",
				HeadBranch:    "add-git-repo",
				BaseBranch:    "main",
				Title:         "Add GitRepository",
				Description:   "Adding GitRepository for app",
				CommitMessage: "feat: Add GitRepository",
			},
			assertFailure(fmt.Errorf("failed to create pull request: %v", "resource is required")),
		},
		{
			"unsupported resource type",
			&pb.CreatePullRequestRequest{
				RepositoryUrl: "https://github.com/weaveworks/weave-gitops",
				HeadBranch:    "add-git-repo",
				BaseBranch:    "main",
				Title:         "Add GitRepository",
				Description:   "Adding GitRepository for app",
				CommitMessage: "feat: Add GitRepository",
				Resource: &pb.TypedObject{
					Type: "foo",
				},
			},
			assertFailure(fmt.Errorf("failed to create pull request: %v", "unsupported type: foo")),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := preview.NewPreviewServiceServer(preview.ServerOpts{
				Logger: testr.New(t),
			})
			response, err := s.CreatePullRequest(context.Background(), tc.request)

			if err := tc.assert(response.GetWebUrl(), err); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestCreatePullRequest(t *testing.T) {
	gitRepo := &pb.GitRepository{
		Name:      "podinfo",
		Namespace: "flux-system",
		Url:       "https://github.com/stefanprodan/podinfo",
		Branch:    ptr.To("main"),
		Interval:  &durationpb.Duration{Seconds: 60},
	}
	gitRepoJSON, _ := json.Marshal(gitRepo)
	req := &pb.CreatePullRequestRequest{
		RepositoryUrl: "https://github.com/weaveworks/weave-gitops",
		HeadBranch:    "add-git-repo",
		BaseBranch:    "main",
		Title:         "Add GitRepository",
		Description:   "Adding GitRepository for app",
		CommitMessage: "feat: Add GitRepository",
		Resource: &pb.TypedObject{
			Type:   sourcev1.GitRepositoryKind,
			Object: string(gitRepoJSON),
		},
	}

	fakeProvider := &TestProvider{}
	s := preview.NewPreviewServiceServer(preview.ServerOpts{
		Logger:          testr.New(t),
		ProviderCreator: NewTestFactory(fakeProvider),
	})
	yamlRes, _ := s.GetYAML(context.Background(), &pb.GetYAMLRequest{Resource: req.Resource})

	// Setup mock
	fakeProvider.On("CreatePullRequest",
		context.Background(),
		git.PullRequestInput{
			RepositoryURL: req.RepositoryUrl,
			Title:         req.Title,
			Body:          req.Description,
			Head:          req.HeadBranch,
			Base:          req.BaseBranch,
			Commits: []git.Commit{
				{
					CommitMessage: req.CommitMessage,
					Files: []git.CommitFile{
						{
							Path:    yamlRes.GetFile().Path,
							Content: &yamlRes.GetFile().Content,
						},
					},
				},
			},
		}).Return(
		&git.PullRequest{
			Link: "https://github.com/weaveworks/weave-gitops/pull/42",
		}, nil)

	res, err := s.CreatePullRequest(context.Background(), req)

	fakeProvider.AssertExpectations(t)

	if err != nil {
		t.Errorf("expected no error to occur but got: %v", err)
	}

	if res == nil || res.WebUrl != "https://github.com/weaveworks/weave-gitops/pull/42" {
		t.Error("expected response to include PR link")
	}
}

func NewTestFactory(provider *TestProvider) git.ProviderCreator {
	return &TestProviderFactory{
		provider: provider,
	}
}

type TestProviderFactory struct {
	provider *TestProvider
}

func (f *TestProviderFactory) Create(providerName string, opts ...git.ProviderWithFn) (git.Provider, error) {
	return f.provider, nil
}

type TestProvider struct {
	mock.Mock
}

func (p *TestProvider) CreatePullRequest(ctx context.Context, input git.PullRequestInput) (*git.PullRequest, error) {
	args := p.Called(ctx, input)
	return args.Get(0).(*git.PullRequest), args.Error(1)
}

func (p *TestProvider) Setup(git.ProviderOption) error {
	return nil
}

func (p *TestProvider) GetRepository(ctx context.Context, repoURL string) (*git.Repository, error) {
	return nil, nil
}

func (p *TestProvider) GetTreeList(ctx context.Context, repoUrl, sha, path string) ([]*git.TreeEntry, error) {
	return nil, nil
}

func (p *TestProvider) ListPullRequests(ctx context.Context, repoUrl string) ([]*git.PullRequest, error) {
	return nil, nil
}
