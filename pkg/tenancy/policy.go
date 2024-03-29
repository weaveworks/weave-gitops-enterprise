package tenancy

import (
	_ "embed"
	"encoding/json"
	"fmt"

	pacv2beta2 "github.com/weaveworks/policy-agent/api/v2beta2"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

const (
	policyRepoGitKind       = "GitRepository"
	policyRepoBucketKind    = "Bucket"
	policyRepoHelmKind      = "HelmRepository"
	policyRepoOCIKind       = "OCIRepository"
	policyClustersKind      = "GitopsCluster"
	policyKustomizationKind = "Kustomization"
	policyHelmReleaseKind   = "HelmRelease"
)

var (
	policyRepoKinds = []string{
		policyRepoGitKind,
		policyRepoBucketKind,
		policyRepoHelmKind,
		policyRepoOCIKind,
	}
	//go:embed policies/allowed_repositories.rego
	repoPolicyCode string
	//go:embed policies/allowed_clusters.rego
	clusterPolicyCode string
	//go:embed policies/allowed_application_deploy.rego
	applicationPolicyCode string
)

func validatePolicyRepoKind(kind string) error {
	for _, repoKind := range policyRepoKinds {
		if repoKind == kind {
			return nil
		}
	}

	return fmt.Errorf("invalid repository kind: %s", kind)
}

func generatePolicyRepoParams(gitURLs, bucketEndpoints, helmURLs, ociURLs []string) ([]pacv2beta2.PolicyParameters, error) {
	gitBytes, err := json.Marshal(gitURLs)
	if err != nil {
		return nil, fmt.Errorf("error while setting policy parameters values: %w", err)
	}

	bucketBytes, err := json.Marshal(bucketEndpoints)
	if err != nil {
		return nil, fmt.Errorf("error while setting policy parameters values: %w", err)
	}

	helmBytes, err := json.Marshal(helmURLs)
	if err != nil {
		return nil, fmt.Errorf("error while setting policy parameters values: %w", err)
	}

	ociBytes, err := json.Marshal(ociURLs)
	if err != nil {
		return nil, fmt.Errorf("error while setting policy parameters values: %w", err)
	}

	return []pacv2beta2.PolicyParameters{
		{
			Name: "git_urls",
			Type: "array",
			Value: &apiextensionsv1.JSON{
				Raw: gitBytes,
			},
		},
		{
			Name: "bucket_endpoints",
			Type: "array",
			Value: &apiextensionsv1.JSON{
				Raw: bucketBytes,
			},
		},
		{
			Name: "helm_urls",
			Type: "array",
			Value: &apiextensionsv1.JSON{
				Raw: helmBytes,
			},
		},
		{
			Name: "oci_urls",
			Type: "array",
			Value: &apiextensionsv1.JSON{
				Raw: ociBytes,
			},
		},
	}, nil
}
