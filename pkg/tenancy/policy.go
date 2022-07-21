package tenancy

import (
	"encoding/json"
	"fmt"

	pacv2beta1 "github.com/weaveworks/policy-agent/api/v2beta1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var (
	policyRepoGitKind    = "GitRepository"
	policyRepoBucketKind = "Bucket"
	policyRepoHelmKind   = "HelmRepository"
)

var policyRepoKinds = []string{
	policyRepoGitKind,
	policyRepoBucketKind,
	policyRepoHelmKind,
}

func validatePolicyRepoKind(kind string) error {
	for _, repoKind := range policyRepoKinds {
		if repoKind == kind {
			return nil
		}
	}
	return fmt.Errorf("invalid repository kind: %s", kind)
}

func generatePolicyRepoParams(gitUrls, bucketEndpoints, hemlUrls []string) ([]pacv2beta1.PolicyParameters, error) {
	gitBytes, err := json.Marshal(gitUrls)
	if err != nil {
		return nil, fmt.Errorf("error while setting policy parameters values: %w", err)
	}

	bucketBytes, err := json.Marshal(bucketEndpoints)
	if err != nil {
		return nil, fmt.Errorf("error while setting policy parameters values: %w", err)
	}

	helmBytes, err := json.Marshal(hemlUrls)
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

const policyCode = `
package weave.tenancy.allowed_repositories

controller_input := input.review.object
namespace := controller_input.metadata.namespace
violation[result] {
	controller_input.kind == "GitRepository"
	urls := input.parameters.git_urls
	url := controller_input.spec.url
	not contains_array(url, urls)
	result = {	
	"issue detected": true,
	"msg": sprintf("Git repository url %v is not allowed for namespace %v", [url, namespace]),
	}
}
violation[result] {
	controller_input.kind == "Bucket"
	urls := input.parameters.bucket_endpoints
	url := controller_input.spec.endpoint
	not contains_array(url, urls)
	result = {	
	"issue detected": true,
	"msg": sprintf("Bucket endpoint %v is not allowed for namespace %v", [url, namespace]),
	}
}
violation[result] {
	controller_input.kind == "HelmRepository"
	urls := input.parameters.helm_urls
	url := controller_input.spec.url
	not contains_array(url, urls)
	result = {	
	"issue detected": true,
	"msg": sprintf("Helm erpository url %v is not allowed for namespace %v", [url, namespace]),
	}
}
contains_array(item, items) {
	items[_] = item
}
`
