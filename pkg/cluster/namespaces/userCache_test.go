package namespaces

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	authv1 "k8s.io/api/authorization/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestUserCacheBuild(t *testing.T) {
	cli := fake.NewSimpleClientset()
	cli.PrependReactor("create", "selfsubjectrulesreviews", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {

		return true, &authv1.SelfSubjectRulesReview{
			Status: authv1.SubjectRulesReviewStatus{
				ResourceRules: []authv1.ResourceRule{
					{
						Verbs:     []string{"list", "get"},
						APIGroups: []string{"gitops.weave.works"},
						Resources: []string{"gitopsclusters"},
					},
				},
			},
		}, nil
	})

	namespaces := []string{"test-ns"}
	var nsList []*v1.Namespace
	for _, ns := range namespaces {
		nsList = append(nsList, &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: ns,
			},
		})

	}

	userID := "id"

	u := NewUsersResourcesNamespaces()
	err := u.Build(context.Background(), userID, cli.AuthorizationV1(), nsList)
	assert.NoError(t, err)

	gotnsList, found := u.Get(userID, "GitopsCluster")
	assert.True(t, found)
	assert.Equal(t, gotnsList, namespaces)

}
