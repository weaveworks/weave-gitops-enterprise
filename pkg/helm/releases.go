package helm

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

const DefaultBufferSize = 2048

// MakeHelmRelease returns a HelmRelease object given a name, version, cluster, namespace, and HelmRepository's name and namespace.
func MakeHelmRelease(name, version, cluster, namespace string, helmRepository types.NamespacedName) *helmv2.HelmRelease {
	return &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster + "-" + name,
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: helmv2.GroupVersion.Identifier(),
			Kind:       helmv2.HelmReleaseKind,
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:   name,
					Version: version,
					SourceRef: helmv2.CrossNamespaceObjectReference{
						APIVersion: sourcev1.GroupVersion.Identifier(),
						Kind:       sourcev1.HelmRepositoryKind,
						Name:       helmRepository.Name,
						Namespace:  helmRepository.Namespace,
					},
				},
			},
			Interval: metav1.Duration{Duration: time.Minute},
		},
	}
}

// AppendHelmReleaseToString appends "---" and a HelmRelease to string that may or may not be empty.
// This creates the content of a manifest that contains HelmReleases separated by "---".
func AppendHelmReleaseToString(content string, newRelease *helmv2.HelmRelease) (string, error) {
	var sb strings.Builder
	if content != "" {
		sb.WriteString(content + "\n")
	}

	helmReleaseManifest, err := yaml.Marshal(newRelease)
	if err != nil {
		return "", fmt.Errorf("failed to marshal HelmRelease: %w", err)
	}

	sb.WriteString("---\n" + string(helmReleaseManifest))

	return sb.String(), nil
}

// SplitHelmReleaseYAML splits a manifest file that contains one or more Helm Releases that may be separated by '---'.
func SplitHelmReleaseYAML(resources []byte) ([]*helmv2.HelmRelease, error) {
	var helmReleaseList []*helmv2.HelmRelease

	decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(resources), DefaultBufferSize)

	for {
		var value helmv2.HelmRelease
		if err := decoder.Decode(&value); err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		helmReleaseList = append(helmReleaseList, &value)
	}

	return helmReleaseList, nil
}

// MarshalHelmReleases marshals a list of HelmReleases.
func MarshalHelmReleases(existingReleases []*helmv2.HelmRelease) (string, error) {
	var sb strings.Builder

	for _, r := range existingReleases {
		b, err := yaml.Marshal(r)
		if err != nil {
			return "", fmt.Errorf("failed to marshal: %w", err)
		}

		sb.WriteString("---\n" + string(b))
	}

	return sb.String(), nil
}
