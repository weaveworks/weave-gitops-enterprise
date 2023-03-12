package namespaces

import (
	"context"
	"fmt"

	clustersv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	gitopssetsv1alpha1 "github.com/weaveworks/gitopssets-controller/api/v1alpha1"
	pipelinesv1alpha1 "github.com/weaveworks/pipeline-controller/api/v1alpha1"
	capiv1 "github.com/weaveworks/templates-controller/apis/capi/v1alpha2"
	gapiv1 "github.com/weaveworks/templates-controller/apis/gitops/v1alpha2"
	authv1 "k8s.io/api/authorization/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	typedauth "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

type RequiredResource struct {
	schema.GroupVersionResource
	Kind string
}

var requiredResources = []RequiredResource{
	{
		GroupVersionResource: pipelinesv1alpha1.GroupVersion.WithResource("pipelines"),
		Kind:                 pipelinesv1alpha1.PipelineKind,
	},
	{
		GroupVersionResource: capiv1.GroupVersion.WithResource("capitemplates"),
		Kind:                 capiv1.Kind,
	},
	{
		GroupVersionResource: gapiv1.GroupVersion.WithResource("gitopstemplates"),
		Kind:                 gapiv1.Kind,
	},
	{
		GroupVersionResource: gitopssetsv1alpha1.GroupVersion.WithResource("gitopssets"),
		Kind:                 "GitOpsSet",
	},
	{
		GroupVersionResource: clustersv1alpha1.GroupVersion.WithResource("gitopsclusters"),
		Kind:                 "GitopsCluster",
	},
}

const (
	allAllowed = "*"
	listVerb   = "list"
	getVerb    = "get"
)

type ruleCache struct {
	apiGroups map[string]struct{}
	resources map[string]struct{}
}

func buildCache(ctx context.Context, client typedauth.AuthorizationV1Interface, namespaces []*v1.Namespace) (map[string][]string, error) {
	resourceNamespaces := make(map[string][]string)
	for _, requiredResource := range requiredResources {
		// make sure each resource is set in map
		resourceNamespaces[requiredResource.Kind] = []string{}
	}
	for i := range namespaces {
		namespace := namespaces[i]
		reviewReqeust := &authv1.SelfSubjectRulesReview{
			Spec: authv1.SelfSubjectRulesReviewSpec{
				Namespace: namespace.Name,
			},
		}

		review, err := client.SelfSubjectRulesReviews().Create(ctx, reviewReqeust, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get user rules for namespace %s: %w", namespace.Name, err)
		}

		rulesCache := buildRules(review.Status.ResourceRules)

		for _, requiredResource := range requiredResources {
			if findRequiredResource(requiredResource, rulesCache) {
				resourceNamespaces[requiredResource.Kind] = append(resourceNamespaces[requiredResource.Kind], namespace.Name)
			}
		}
	}

	return resourceNamespaces, nil
}

func contains(key string, cache map[string]struct{}) bool {
	_, checkAll := cache[allAllowed]
	_, checkExplicit := cache[key]
	return checkExplicit || checkAll
}

func buildRules(rules []authv1.ResourceRule) []ruleCache {
	var rulesCache []ruleCache
	for _, rule := range rules {
		groups := make(map[string]struct{})
		resources := make(map[string]struct{})
		verbs := make(map[string]struct{})
		for _, verb := range rule.Verbs {
			verbs[verb] = struct{}{}
		}
		if contains(listVerb, verbs) && contains(getVerb, verbs) {

			for _, apiGroup := range rule.APIGroups {
				groups[apiGroup] = struct{}{}
			}
			for _, resource := range rule.Resources {
				resources[resource] = struct{}{}
			}

			rulesCache = append(rulesCache, ruleCache{
				apiGroups: groups,
				resources: resources,
			})

		}
	}
	return rulesCache
}

func findRequiredResource(resource RequiredResource, rulesCache []ruleCache) bool {
	for _, ruleCache := range rulesCache {
		if contains(resource.Group, ruleCache.apiGroups) && contains(resource.Resource, ruleCache.resources) {
			return true
		}
	}
	return false
}
