package utils

import (
	"encoding/json"
	"fmt"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops-enterprise/cmd/gitops/app/bootstrap/domain"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8syaml "sigs.k8s.io/yaml"
)

func CreateHelmReleaseYamlString(hr helmv2.HelmRelease) (string, error) {
	helmRelease := helmv2.HelmRelease{
		TypeMeta: v1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.Identifier(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:              hr.Name,
			Namespace:         hr.Namespace,
			CreationTimestamp: v1.Now(),
		}, Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:             hr.Spec.Chart.Spec.Chart,
					ReconcileStrategy: sourcev1.ReconcileStrategyChartVersion,
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.HelmRepositoryKind,
						Name:      hr.Spec.Chart.Spec.SourceRef.Name,
						Namespace: hr.Spec.Chart.Spec.SourceRef.Namespace,
					},
					Version: hr.Spec.Chart.Spec.Version,
				},
			},
			Install: &helmv2.Install{
				CRDs:            hr.Spec.Install.CRDs,
				CreateNamespace: hr.Spec.Install.CreateNamespace,
			},
			Upgrade: &helmv2.Upgrade{
				CRDs: hr.Spec.Upgrade.CRDs,
			},
			Interval: v1.Duration{
				Duration: hr.Spec.Interval.Duration,
			},
			Values: hr.Spec.Values,
		},
	}

	helmReleaseBytes, err := k8syaml.Marshal(helmRelease)
	if err != nil {
		return "", err
	}
	return string(helmReleaseBytes), nil
}

func CreateHelmRepositoryYamlString(helmRepo sourcev1.HelmRepository) (string, error) {
	repo := sourcev1.HelmRepository{
		TypeMeta: v1.TypeMeta{
			APIVersion: sourcev1.GroupVersion.Identifier(),
			Kind:       sourcev1.HelmRepositoryKind,
		},
		ObjectMeta: v1.ObjectMeta{
			Name:              helmRepo.Name,
			Namespace:         helmRepo.Namespace,
			CreationTimestamp: v1.Now(),
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL: helmRepo.Spec.URL,
			Interval: v1.Duration{
				Duration: helmRepo.Spec.Interval.Duration,
			},
			SecretRef: &meta.LocalObjectReference{
				Name: helmRepo.Spec.SecretRef.Name,
			},
		},
	}
	repoBytes, err := k8syaml.Marshal(repo)
	if err != nil {
		return "", err
	}
	return string(repoBytes), nil
}

// ReconcileFlux reconcile flux defaults
func ReconcileFlux(helmReleaseName ...string) error {
	var runner runner.CLIRunner
	out, err := runner.Run("flux", "reconcile", "source", "git", "flux-system")
	if err != nil {
		return fmt.Errorf("%s%s", err.Error(), string(out))
	}

	out, err = runner.Run("flux", "reconcile", "kustomization", "flux-system")
	if err != nil {
		return fmt.Errorf("%s%s", err.Error(), string(out))
	}

	if len(helmReleaseName) > 0 {
		out, err = runner.Run("flux", "reconcile", "helmrelease", helmReleaseName[0])
		if err != nil {
			return fmt.Errorf("%s%s", err.Error(), string(out))
		}
	}

	return nil
}

func GetCurrentValuesForHelmRelease(name string, namespace string) (domain.ValuesFile, error) {
	var runner runner.CLIRunner
	out, err := runner.Run("kubectl", "get", "helmrelease", name, "-n", namespace, "-o", "jsonpath=\"{.spec.values}\"")
	if err != nil {
		return domain.ValuesFile{}, fmt.Errorf("%s%s", err.Error(), string(out))
	}

	values := domain.ValuesFile{}
	err = json.Unmarshal(out[1:len(out)-1], &values)
	if err != nil {
		return domain.ValuesFile{}, fmt.Errorf("%s%s", err.Error(), string(out))
	}

	return values, nil
}

func GetCurrentVersionForHelmRelease(name string, namespace string) (string, error) {
	var runner runner.CLIRunner
	out, err := runner.Run("kubectl", "get", "helmrelease", name, "-n", namespace, "-o", "jsonpath=\"{.spec.chart.spec.version}\"")
	if err != nil {
		return "", fmt.Errorf("%s%s", err.Error(), string(out))
	}

	return string(out[1 : len(out)-1]), nil
}
