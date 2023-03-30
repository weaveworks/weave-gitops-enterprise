package namespaces

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	authv1 "k8s.io/api/authorization/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestBuildCache(t *testing.T) {
	namespaces := []*v1.Namespace{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-ns",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
		},
	}
	testCases := []struct {
		name   string
		review *authv1.SelfSubjectRulesReview
		want   map[string][]string
		err    error
	}{
		{
			name: "all exist explicit",
			review: &authv1.SelfSubjectRulesReview{
				Status: authv1.SubjectRulesReviewStatus{
					ResourceRules: []authv1.ResourceRule{
						{
							Verbs:     []string{"list", "get"},
							APIGroups: []string{"gitops.weave.works", "pipelines.weave.works"},
							Resources: []string{"gitopsclusters", "pipelines"},
						},
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"capi.weave.works", "templates.weave.works"},
							Resources: []string{"capitemplates", "gitopstemplates", "gitopssets"},
						},
					},
				},
			},
			want: map[string][]string{
				"Pipeline":       {"test-ns", "default"},
				"CAPITemplate":   {"test-ns", "default"},
				"GitOpsSet":      {"test-ns", "default"},
				"GitOpsTemplate": {"test-ns", "default"},
				"GitopsCluster":  {"test-ns", "default"},
			},
		},
		{
			name: "all exist *",
			review: &authv1.SelfSubjectRulesReview{
				Status: authv1.SubjectRulesReviewStatus{
					ResourceRules: []authv1.ResourceRule{
						{
							Verbs:     []string{"*"},
							APIGroups: []string{"*"},
							Resources: []string{"*"},
						},
					},
				},
			},
			want: map[string][]string{
				"Pipeline":       {"test-ns", "default"},
				"CAPITemplate":   {"test-ns", "default"},
				"GitOpsSet":      {"test-ns", "default"},
				"GitOpsTemplate": {"test-ns", "default"},
				"GitopsCluster":  {"test-ns", "default"},
			},
		},
		{
			name: "partial exist",
			review: &authv1.SelfSubjectRulesReview{
				Status: authv1.SubjectRulesReviewStatus{
					ResourceRules: []authv1.ResourceRule{
						{
							Verbs:     []string{"list", "get"},
							APIGroups: []string{"gitops.weave.works", "pipelines.weave.works"},
							Resources: []string{"gitopsclusters", "pipelines"},
						},
					},
				},
			},
			want: map[string][]string{
				"Pipeline":       {"test-ns", "default"},
				"CAPITemplate":   {},
				"GitOpsSet":      {},
				"GitOpsTemplate": {},
				"GitopsCluster":  {"test-ns", "default"},
			},
		},
		{
			name: "error getting subject review",
			err:  errors.New("failed to get user rules for namespace test-ns: server error"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			cli := fake.NewSimpleClientset()
			cli.PrependReactor("create", "selfsubjectrulesreviews", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
				if tt.err != nil {
					return true, nil, errors.New("server error")
				}
				return true, tt.review, nil
			})
			got, gotErr := buildCache(context.Background(), cli.AuthorizationV1(), namespaces)
			if tt.err != nil {
				assert.Equal(t, tt.err.Error(), gotErr.Error(), "unexpected error result")
			}
			assert.Equal(t, tt.want, got)
		})
	}

}
