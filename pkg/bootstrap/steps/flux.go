package steps

import (
	"context"
	"fmt"
	"strings"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/bootstrap/utils"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	k8s_client "sigs.k8s.io/controller-runtime/pkg/client"
)

// user messages
const (
	fluxBoostrapCheckMsg     = "checking flux"
	fluxExistingInstallMsg   = "flux is installed"
	fluxExistingBootstrapMsg = "flux is bootstrapped"
	fluxRecoverMsg           = "flux is not bootstrapped in 'flux-system' namespace: more info https://fluxcd.io/flux/installation"
	fluxFatalErrorMsg        = "flux is not bootstrapped, please bootstrap Flux in 'flux-system' namespace: more info https://fluxcd.io/flux/installation"
)

// FluxConfig holds configuration about the existing Flux in the cluser
type FluxConfig struct {
	// Url flux-system git repository url
	Url string
	// Scheme flux-system git repository scheme
	Scheme string
	// Branch flux-system git repository branch
	Branch string
	// Path flux-system kustomization path to reconcile
	Path string
	// IsInstalled indicates whether flux is already installed
	IsInstalled bool
}

// NewFluxConfig creates a Flux configuration out of the existing cluster
func NewFluxConfig(logger logger.Logger, client k8s_client.Client) (FluxConfig, error) {
	var runner runner.CLIRunner

	logger.Actionf("verifying flux installation")
	out, err := runner.Run("flux", "check")
	if err != nil {
		if strings.Contains(string(out), "customresourcedefinitions.apiextensions.k8s.io \"gitrepositories.source.toolkit.fluxcd.io\" not found") {
			return FluxConfig{
				IsInstalled: false,
			}, nil
		}
		return FluxConfig{}, fmt.Errorf("flux installed error: %v. %s", string(out), fluxRecoverMsg)
	}
	logger.Successf(fluxExistingInstallMsg)

	logger.Actionf("verifying flux reconciliation")
	out, err = runner.Run("flux", "reconcile", "kustomization", "flux-system")
	if err != nil {
		return FluxConfig{}, fmt.Errorf("flux bootstrapped error: %v. %s", string(out), fluxFatalErrorMsg)
	}
	logger.Successf(fluxExistingBootstrapMsg)

	repo, err := utils.GetGitRepositoryObject(client, WGEDefaultRepoName, WGEDefaultNamespace)
	if err != nil {
		return FluxConfig{}, fmt.Errorf("failed to get flux repository: %v", err)
	}

	repoUrl, scheme, err := normaliseUrl(repo.Spec.URL)
	if err != nil {
		return FluxConfig{}, fmt.Errorf("failed to parse flux repository: %v", err)
	}

	repoPath, err := getRepoPath(client, WGEDefaultRepoName, WGEDefaultNamespace)
	if err != nil {
		return FluxConfig{}, fmt.Errorf("failed to get flux repo path: %v", err)
	}

	logger.Successf("detected git scheme: %s", scheme)

	return FluxConfig{
		Url:    repoUrl,
		Scheme: scheme,
		Branch: repo.Spec.Reference.Branch,
		Path:   repoPath,

		IsInstalled: true,
	}, nil
}

// getRepoPath get the path for flux installation (flux-system) Kustomization.
func getRepoPath(client k8s_client.Client, repoName string, namespace string) (string, error) {
	kustomization := &kustomizev1.Kustomization{}

	if err := client.Get(context.Background(), k8s_client.ObjectKey{
		Namespace: namespace,
		Name:      repoName,
	}, kustomization); err != nil {
		return "", err
	}

	return kustomization.Spec.Path, nil
}

// VerifyFluxInstallation checks that Flux is present in the cluster. It fails in case not and returns next steps to install it.
var VerifyFluxInstallation = BootstrapStep{
	Name: fluxBoostrapCheckMsg,
	Step: doNothingStep,
}
