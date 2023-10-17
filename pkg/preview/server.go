package preview

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/preview"
	"golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type ServerOpts struct {
	logr.Logger
}

type server struct {
	pb.UnimplementedPreviewServiceServer

	log logr.Logger
}

func Hydrate(ctx context.Context, mux *runtime.ServeMux, opts ServerOpts) error {
	s := NewPreviewServiceServer(opts)

	return pb.RegisterPreviewServiceHandlerServer(ctx, mux, s)
}

func NewPreviewServiceServer(opts ServerOpts) pb.PreviewServiceServer {
	return &server{
		log: opts.Logger,
		// clients: opts.ClientsFactory,
		// scheme:  opts.Scheme,
	}
}

func (s *server) GetYAML(ctx context.Context, msg *pb.GetYAMLRequest) (*pb.GetYAMLResponse, error) {
	var (
		yaml string
		err  error
	)
	switch msg.Type {
	case sourcev1.GitRepositoryKind:
		var m pb.GitRepository
		if err := json.Unmarshal([]byte(msg.Resource), &m); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON object: %w", err)
		}
		yaml, err = generateGitRepositoryYAML(&m)
	case sourcev1beta2.HelmRepositoryKind:
		var m pb.HelmRepository
		if err := json.Unmarshal([]byte(msg.Resource), &m); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON object: %w", err)
		}
		yaml, err = generateHelmRepositoryYAML(&m)
	case sourcev1beta2.BucketKind:
		var m pb.Bucket
		if err := json.Unmarshal([]byte(msg.Resource), &m); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON object: %w", err)
		}
		yaml, err = generateBucketYAML(&m)
	case sourcev1beta2.OCIRepositoryKind:
		var m pb.OCIRepository
		if err := json.Unmarshal([]byte(msg.Resource), &m); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON object: %w", err)
		}
		yaml, err = generateOCIRepositoryYAML(&m)
	default:
		return nil, fmt.Errorf("unsupported type: %v", msg.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate YAML for %q: %w", msg.Type, err)
	}

	return &pb.GetYAMLResponse{
		Yaml: yaml,
	}, nil
}

func generateGitRepositoryYAML(resource *pb.GitRepository) (string, error) {
	if resource.GetName() == "" {
		return "", errors.New("name is required")
	}

	if resource.GetNamespace() == "" {
		return "", errors.New("namespace is required")
	}

	if resource.GetInterval().IsValid() && resource.GetInterval().Seconds == 0 {
		return "", errors.New("invalid interval value")
	}

	if resource.GetUrl() == "" {
		return "", errors.New("url is required")
	}
	url, err := url.Parse(resource.GetUrl())
	if err != nil {
		return "", fmt.Errorf("invalid url value: %w", err)
	}
	if url.Scheme != "ssh" && url.Scheme != "http" && url.Scheme != "https" {
		return "", fmt.Errorf("url scheme %q is not supported", url.Scheme)
	}

	gvk := sourcev1.GroupVersion.WithKind(sourcev1.GitRepositoryKind)
	gitRepository := sourcev1.GitRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       gvk.Kind,
			APIVersion: gvk.GroupVersion().String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resource.GetName(),
			Namespace: resource.GetNamespace(),
		},
		Spec: sourcev1.GitRepositorySpec{
			URL: resource.GetUrl(),
			Interval: metav1.Duration{
				Duration: resource.GetInterval().AsDuration(),
			},
			Reference: &sourcev1.GitRepositoryRef{},
		},
	}

	if resource.GetBranch() == "" && resource.GetTag() == "" &&
		resource.GetSemver() == "" && resource.GetCommit() == "" && resource.GetRefName() == "" {
		return "", errors.New("a Git ref is required")
	}

	if resource.GetCommit() != "" {
		gitRepository.Spec.Reference.Commit = resource.GetCommit()
		gitRepository.Spec.Reference.Branch = resource.GetBranch()
	} else if resource.GetRefName() != "" {
		gitRepository.Spec.Reference.Name = resource.GetRefName()
	} else if resource.GetSemver() != "" {
		gitRepository.Spec.Reference.SemVer = resource.GetSemver()
	} else if resource.GetTag() != "" {
		gitRepository.Spec.Reference.Tag = resource.GetTag()
	} else {
		gitRepository.Spec.Reference.Branch = resource.GetBranch()
	}

	if resource.GetSecretRefName() != "" {
		gitRepository.Spec.SecretRef = &meta.LocalObjectReference{
			Name: resource.GetSecretRefName(),
		}
	}

	return printExport(&gitRepository)
}

func generateHelmRepositoryYAML(resource *pb.HelmRepository) (string, error) {
	if resource.GetName() == "" {
		return "", errors.New("name is required")
	}

	if resource.GetNamespace() == "" {
		return "", errors.New("namespace is required")
	}

	if resource.GetInterval().IsValid() && resource.GetInterval().Seconds == 0 {
		return "", errors.New("invalid interval value")
	}

	gvk := sourcev1beta2.GroupVersion.WithKind(sourcev1beta2.HelmRepositoryKind)
	helmRepository := sourcev1beta2.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       gvk.Kind,
			APIVersion: gvk.GroupVersion().String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resource.GetName(),
			Namespace: resource.GetNamespace(),
		},
		Spec: sourcev1beta2.HelmRepositorySpec{
			Interval: metav1.Duration{
				Duration: resource.GetInterval().AsDuration(),
			},
		},
	}

	if resource.Type != nil {
		if resource.GetType() == sourcev1beta2.HelmRepositoryTypeDefault || resource.GetType() == sourcev1beta2.HelmRepositoryTypeOCI {
			helmRepository.Spec.Type = resource.GetType()
		} else {
			return "", errors.New("invalid type")
		}
	}

	var validProviders = []string{
		sourcev1beta2.GenericOCIProvider,
		sourcev1beta2.AmazonOCIProvider,
		sourcev1beta2.AzureOCIProvider,
		sourcev1beta2.GoogleOCIProvider,
	}

	if resource.Provider != nil && !slices.Contains(validProviders, resource.GetProvider()) {
		return "", errors.New("invalid provider")
	}

	if resource.GetUrl() == "" {
		return "", errors.New("url is required")
	}
	url, err := url.Parse(resource.Url)
	if err != nil {
		return "", fmt.Errorf("invalid url value: %w", err)
	}

	helmRepository.Spec.URL = resource.Url

	if url.Scheme == sourcev1beta2.HelmRepositoryTypeOCI {
		helmRepository.Spec.Type = sourcev1beta2.HelmRepositoryTypeOCI
		helmRepository.Spec.Provider = resource.GetProvider()
	}

	if resource.GetSecretRefName() != "" {
		helmRepository.Spec.SecretRef = &meta.LocalObjectReference{
			Name: resource.GetSecretRefName(),
		}

		if resource.PassCredentials != nil {
			helmRepository.Spec.PassCredentials = resource.GetPassCredentials()
		} else {
			helmRepository.Spec.PassCredentials = false
		}
	}

	return printExport(&helmRepository)
}

func generateBucketYAML(resource *pb.Bucket) (string, error) {
	if resource.GetName() == "" {
		return "", errors.New("name is required")
	}

	if resource.GetNamespace() == "" {
		return "", errors.New("namespace is required")
	}

	if resource.GetInterval().IsValid() && resource.GetInterval().Seconds == 0 {
		return "", errors.New("invalid interval value")
	}

	if resource.GetBucketName() == "" {
		return "", errors.New("bucket name is required")
	}

	if resource.GetEndpoint() == "" {
		return "", errors.New("endpoint is required")
	}

	gvk := sourcev1beta2.GroupVersion.WithKind(sourcev1beta2.BucketKind)
	bucket := sourcev1beta2.Bucket{
		TypeMeta: metav1.TypeMeta{
			Kind:       gvk.Kind,
			APIVersion: gvk.GroupVersion().String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resource.GetName(),
			Namespace: resource.GetNamespace(),
		},
		Spec: sourcev1beta2.BucketSpec{
			Interval: metav1.Duration{
				Duration: resource.GetInterval().AsDuration(),
			},
			BucketName: resource.GetBucketName(),
			Endpoint:   resource.GetEndpoint(),
		},
	}

	var validBucketProviders = []string{
		sourcev1beta2.GenericBucketProvider,
		sourcev1beta2.AmazonBucketProvider,
		sourcev1beta2.AzureBucketProvider,
		sourcev1beta2.GoogleBucketProvider,
	}

	if resource.Provider != nil {
		if !slices.Contains(validBucketProviders, resource.GetProvider()) {
			return "", errors.New("invalid provider")
		}

		if resource.GetProvider() == sourcev1beta2.GenericBucketProvider && resource.GetSecretRefName() == "" {
			return "", errors.New("generic provider requires a secretRef")
		}

		bucket.Spec.Provider = resource.GetProvider()
	}

	if resource.Region != nil && resource.GetRegion() != "" {
		bucket.Spec.Region = resource.GetRegion()
	}

	if resource.Insecure != nil {
		bucket.Spec.Insecure = resource.GetInsecure()
	}

	if resource.GetSecretRefName() != "" {
		bucket.Spec.SecretRef = &meta.LocalObjectReference{
			Name: resource.GetSecretRefName(),
		}
	}

	return printExport(&bucket)
}

func generateOCIRepositoryYAML(resource *pb.OCIRepository) (string, error) {
	if resource.GetName() == "" {
		return "", errors.New("name is required")
	}

	if resource.GetNamespace() == "" {
		return "", errors.New("namespace is required")
	}

	if resource.GetInterval().IsValid() && resource.GetInterval().Seconds == 0 {
		return "", errors.New("invalid interval value")
	}

	if resource.GetUrl() == "" {
		return "", errors.New("url is required")
	}
	url, err := url.Parse(resource.GetUrl())
	if err != nil {
		return "", fmt.Errorf("invalid url value: %w", err)
	}
	if url.Scheme != "oci" {
		return "", fmt.Errorf("url scheme must be set to %q", "oci")
	}

	gvk := sourcev1beta2.GroupVersion.WithKind(sourcev1beta2.OCIRepositoryKind)
	ociRepository := sourcev1beta2.OCIRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       gvk.Kind,
			APIVersion: gvk.GroupVersion().String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resource.GetName(),
			Namespace: resource.GetNamespace(),
		},
		Spec: sourcev1beta2.OCIRepositorySpec{
			URL: resource.GetUrl(),
			Interval: metav1.Duration{
				Duration: resource.GetInterval().AsDuration(),
			},
			Reference: &sourcev1beta2.OCIRepositoryRef{},
		},
	}

	var validProviders = []string{
		sourcev1beta2.GenericOCIProvider,
		sourcev1beta2.AmazonOCIProvider,
		sourcev1beta2.AzureOCIProvider,
		sourcev1beta2.GoogleOCIProvider,
	}

	if resource.Provider != nil {
		if !slices.Contains(validProviders, resource.GetProvider()) {
			return "", errors.New("invalid provider")
		} else {
			ociRepository.Spec.Provider = resource.GetProvider()
		}
	}

	if resource.GetTag() == "" && resource.GetSemver() == "" && resource.GetDigest() == "" {
		return "", errors.New("ref is required")
	}

	if resource.GetTag() != "" {
		ociRepository.Spec.Reference.Tag = resource.GetTag()
	} else if resource.GetSemver() != "" {
		ociRepository.Spec.Reference.SemVer = resource.GetSemver()
	} else {
		ociRepository.Spec.Reference.Digest = resource.GetDigest()
	}

	if resource.Insecure != nil {
		ociRepository.Spec.Insecure = *resource.Insecure
	}

	if resource.GetServiceAccountName() != "" {
		ociRepository.Spec.ServiceAccountName = resource.GetServiceAccountName()
	}

	if resource.GetSecretRefName() != "" {
		ociRepository.Spec.SecretRef = &meta.LocalObjectReference{
			Name: resource.GetSecretRefName(),
		}
	}

	if resource.GetCertSecretRefName() != "" {
		ociRepository.Spec.CertSecretRef = &meta.LocalObjectReference{
			Name: resource.GetCertSecretRefName(),
		}
	}

	return printExport(&ociRepository)
}

func printExport(export interface{}) (string, error) {
	data, err := yaml.Marshal(export)
	if err != nil {
		return "", err
	}
	return resourceToString(data), nil
}

func resourceToString(data []byte) string {
	data = bytes.Replace(data, []byte("  creationTimestamp: null\n"), []byte(""), 1)
	data = bytes.Replace(data, []byte("status: {}\n"), []byte(""), 1)
	data = bytes.TrimSpace(data)
	return string(data)
}
