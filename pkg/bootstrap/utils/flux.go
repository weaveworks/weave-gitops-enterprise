package utils

import (
	"context"
	"encoding/json"
	"fmt"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s_client "sigs.k8s.io/controller-runtime/pkg/client"
	k8syaml "sigs.k8s.io/yaml"
)

type FluxClient interface {
	ReconcileFlux() error
	ReconcileHelmRelease(hrName string) error
}

type CmdFluxClient struct{}

// CreateHelmReleaseYamlString create HelmRelease yaml string to add to file.
func CreateHelmReleaseYamlString(hr helmv2.HelmRelease) (string, error) {
	helmRelease := helmv2.HelmRelease{
		TypeMeta: v1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.Identifier(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      hr.Name,
			Namespace: hr.Namespace,
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

// CreateHelmRepositoryYamlString create HelmRepository yaml string to add to file.
func CreateHelmRepositoryYamlString(helmRepo sourcev1.HelmRepository) (string, error) {
	repo := sourcev1.HelmRepository{
		TypeMeta: v1.TypeMeta{
			APIVersion: sourcev1.GroupVersion.Identifier(),
			Kind:       sourcev1.HelmRepositoryKind,
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      helmRepo.Name,
			Namespace: helmRepo.Namespace,
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

// ReconcileFlux reconcile flux default source and kustomization
// Reconciliation is important to apply the effect of adding resources to the git repository
func (fc CmdFluxClient) ReconcileFlux() error {
	var runner runner.CLIRunner
	out, err := runner.Run("flux", "reconcile", "source", "git", "flux-system")
	if err != nil {
		// adding an error message, err is meaningless
		return fmt.Errorf("failed to reconcile flux source flux-system: %s", string(out))
	}

	out, err = runner.Run("flux", "reconcile", "kustomization", "flux-system")
	if err != nil {
		// adding an error message, err is meaningless
		return fmt.Errorf("failed to reconcile flux kustomization flux-system: %s", string(out))
	}

	return nil
}

// ReconcileHelmRelease reconcile a particular helmrelease
func (fc CmdFluxClient) ReconcileHelmRelease(hrName string) error {
	var runner runner.CLIRunner
	out, err := runner.Run("flux", "reconcile", "helmrelease", hrName)
	if err != nil {
		// adding an error message, err is meaningless
		return fmt.Errorf("failed to reconcile flux helmrelease %s: %s", hrName, string(out))
	}

	return nil
}

// GetHelmReleaseProperty extract a property from a specific helmrelease values file
func GetHelmReleaseProperty(client k8s_client.Client, releaseName string, namespace string, property string) (string, error) {
	helmrelease := &helmv2.HelmRelease{}
	if err := client.Get(context.Background(), k8s_client.ObjectKey{
		Namespace: namespace,
		Name:      releaseName,
	}, helmrelease); err != nil {
		return "", err
	}

	switch property {
	case "version":
		return helmrelease.Spec.Chart.Spec.Version, nil
	case "domain":
		values := map[string]interface{}{}
		if err := json.Unmarshal(helmrelease.Spec.Values.Raw, &values); err != nil {
			return "", err
		}
		return values["ingress"].(map[string]interface{})["hosts"].([]interface{})[0].(map[string]interface{})["host"].(string), nil
	default:
		return "", fmt.Errorf("unsupported property: %s", property)
	}
}

// GetHelmReleaseValues gets the current values from a specific helmrelease.
func GetHelmReleaseValues(client k8s_client.Client, name string, namespace string) ([]byte, error) {
	helmrelease := &helmv2.HelmRelease{}
	if err := client.Get(context.Background(), k8s_client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, helmrelease); err != nil {
		return nil, err
	}

	return helmrelease.Spec.Values.Raw, nil
}
