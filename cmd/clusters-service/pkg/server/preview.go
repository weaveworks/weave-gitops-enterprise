package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	protos "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type generateYAMLFunc func(values map[string]string) (string, error)

func (s *server) PreviewYAML(ctx context.Context, msg *protos.PreviewYAMLRequest) (*protos.PreviewYAMLResponse, error) {
	var generate generateYAMLFunc

	switch msg.Kind {
	case sourcev1.GitRepositoryKind:
		generate = generateGitRepositoryYAML
	case sourcev1beta2.HelmRepositoryKind:
		generate = generateHelmRepositoryYAML
	case sourcev1beta2.BucketKind:
		generate = generateBucketYAML
	case sourcev1beta2.OCIRepositoryKind:
		generate = generateOCIRepositoryYAML
	default:
		return &protos.PreviewYAMLResponse{}, fmt.Errorf("kind %q is not supported", msg.Kind)
	}

	content, err := generate(msg.Values)
	if err != nil {
		return &protos.PreviewYAMLResponse{}, fmt.Errorf("cannot generate preview for %q: %w", msg.Kind, err)
	}

	return &protos.PreviewYAMLResponse{
		Preview: &protos.CommitFile{
			Path:    "",
			Content: content,
		},
	}, nil
}

func generateGitRepositoryYAML(values map[string]string) (string, error) {
	name := values["name"]
	if name == "" {
		return "", errors.New("name is required")
	}

	namespace := values["namespace"]
	if namespace == "" {
		return "", errors.New("namespace is required")
	}

	inputInterval := values["interval"]
	if inputInterval == "" {
		return "", errors.New("interval is required")
	}
	interval, err := time.ParseDuration(inputInterval)
	if err != nil {
		return "", fmt.Errorf("invalid interval value: %w", err)
	}

	inputUrl := values["url"]
	if inputUrl == "" {
		return "", errors.New("url is required")
	}
	u, err := url.Parse(inputUrl)
	if err != nil {
		return "", fmt.Errorf("invalid url value: %w", err)
	}
	if u.Scheme != "ssh" && u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("url scheme %q is not supported", u.Scheme)
	}

	gvk := sourcev1.GroupVersion.WithKind(sourcev1.GitRepositoryKind)
	gitRepository := sourcev1.GitRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       gvk.Kind,
			APIVersion: gvk.GroupVersion().String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: sourcev1.GitRepositorySpec{
			URL: inputUrl,
			Interval: metav1.Duration{
				Duration: interval,
			},
			Reference: &sourcev1.GitRepositoryRef{},
		},
	}

	branch := values["branch"]
	tag := values["tag"]
	semver := values["semver"]
	commit := values["commit"]
	refName := values["ref-name"]
	if branch == "" && tag == "" && semver == "" && commit == "" && refName == "" {
		return "", fmt.Errorf("a Git ref is required, use one of the following: branch, tag, semver, commit or ref-name")
	}

	if commit != "" {
		gitRepository.Spec.Reference.Commit = commit
		gitRepository.Spec.Reference.Branch = branch
	} else if refName != "" {
		gitRepository.Spec.Reference.Name = refName
	} else if semver != "" {
		gitRepository.Spec.Reference.SemVer = semver
	} else if tag != "" {
		gitRepository.Spec.Reference.Tag = tag
	} else {
		gitRepository.Spec.Reference.Branch = branch
	}

	if secretRefName, ok := values["secret-ref-name"]; ok && secretRefName != "" {
		gitRepository.Spec.SecretRef = &meta.LocalObjectReference{
			Name: secretRefName,
		}
	}

	return printExport(&gitRepository)
}

func generateHelmRepositoryYAML(values map[string]string) (string, error) {
	name := values["name"]
	if name == "" {
		return "", errors.New("name is required")
	}

	namespace := values["namespace"]
	if namespace == "" {
		return "", errors.New("namespace is required")
	}

	inputInterval := values["interval"]
	if inputInterval == "" {
		return "", errors.New("interval is required")
	}
	interval, err := time.ParseDuration(inputInterval)
	if err != nil {
		return "", fmt.Errorf("invalid interval value: %w", err)
	}

	gvk := sourcev1beta2.GroupVersion.WithKind(sourcev1beta2.HelmRepositoryKind)
	helmRepository := sourcev1beta2.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       gvk.Kind,
			APIVersion: gvk.GroupVersion().String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: sourcev1beta2.HelmRepositorySpec{
			Interval: metav1.Duration{
				Duration: interval,
			},
		},
	}

	if repoType, ok := values["type"]; ok && (repoType == "default" || repoType == "oci") {
		helmRepository.Spec.Type = repoType
	} else if ok {
		return "", errors.New("invalid type")
	}

	var validOCIProviders = []string{
		sourcev1beta2.GenericOCIProvider,
		sourcev1beta2.AmazonOCIProvider,
		sourcev1beta2.AzureOCIProvider,
		sourcev1beta2.GoogleOCIProvider,
	}

	provider, ok := values["provider"]
	if ok && !slices.Contains(validOCIProviders, provider) {
		return "", errors.New("invalid provider")
	}

	inputUrl := values["url"]
	if inputUrl == "" {
		return "", errors.New("url is required")
	}
	u, err := url.Parse(inputUrl)
	if err != nil {
		return "", fmt.Errorf("invalid url value: %w", err)
	}

	helmRepository.Spec.URL = inputUrl

	if u.Scheme == sourcev1beta2.HelmRepositoryTypeOCI {
		helmRepository.Spec.Type = sourcev1beta2.HelmRepositoryTypeOCI
		helmRepository.Spec.Provider = provider
	}

	if secretRefName, ok := values["secret-ref-name"]; ok && secretRefName != "" {
		helmRepository.Spec.SecretRef = &meta.LocalObjectReference{
			Name: secretRefName,
		}

		if passCredentials, ok := values["pass-credentials"]; ok {
			if passCredentialsValue, err := strconv.ParseBool(passCredentials); err != nil {
				return "", errors.New("invalid pass-credentials value")
			} else {
				helmRepository.Spec.PassCredentials = passCredentialsValue
			}
		} else {
			helmRepository.Spec.PassCredentials = false
		}
	}

	return printExport(&helmRepository)
}

func generateBucketYAML(values map[string]string) (string, error) {
	name := values["name"]
	if name == "" {
		return "", errors.New("name is required")
	}

	namespace := values["namespace"]
	if namespace == "" {
		return "", errors.New("namespace is required")
	}

	return "", nil
}

func generateOCIRepositoryYAML(values map[string]string) (string, error) {
	name := values["name"]
	if name == "" {
		return "", errors.New("name is required")
	}

	namespace := values["namespace"]
	if namespace == "" {
		return "", errors.New("namespace is required")
	}

	return "", nil
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
