package tenancy

import (
	_ "embed"
	"encoding/json"
	"fmt"

	pacv2beta1 "github.com/weaveworks/policy-agent/api/v2beta1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

const (
	policyRepoGitKind    = "GitRepository"
	policyRepoBucketKind = "Bucket"
	policyRepoHelmKind   = "HelmRepository"
	policyClusterskind   = "GitopsCluster"
)

var (
	policyRepoKinds = []string{
		policyRepoGitKind,
		policyRepoBucketKind,
		policyRepoHelmKind,
	}
	//go:embed policies/allowed_repositories.rego
	repoPolicyCode string
	//go:embed policies/allowed_clusters.rego
	clusterPolicyCode string
)

func validatePolicyRepoKind(kind string) error {
	for _, repoKind := range policyRepoKinds {
		if repoKind == kind {
			return nil
		}
	}
	return fmt.Errorf("invalid repository kind: %s", kind)
}

func generatePolicyRepoParams(gitURLs, bucketEndpoints, helmURLs []string) ([]pacv2beta1.PolicyParameters, error) {
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
	return []pacv2beta1.PolicyParameters{
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
	}, nil
}
