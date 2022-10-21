package fake

import (
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	authv1 "k8s.io/api/authorization/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	typedauth "k8s.io/client-go/kubernetes/typed/authorization/v1"
	k8stesting "k8s.io/client-go/testing"
)

type FakeNamespaceCache struct {
	Namespaces []*v1.Namespace
	Err        error
}

func (f *FakeNamespaceCache) List() ([]*v1.Namespace, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	return f.Namespaces, nil
}

type FakeAuthClientGetter struct{}

func (f *FakeAuthClientGetter) Get(user *auth.UserPrincipal) (typedauth.AuthorizationV1Interface, error) {
	cli := k8sfake.NewSimpleClientset()
	cli.PrependReactor("create", "selfsubjectrulesreviews", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &authv1.SelfSubjectRulesReview{
			Status: authv1.SubjectRulesReviewStatus{
				ResourceRules: []authv1.ResourceRule{
					{
						Verbs:     []string{"*"},
						APIGroups: []string{"*"},
						Resources: []string{"*"},
					},
				},
			},
		}, nil
	})
	return cli.AuthorizationV1(), nil
}
