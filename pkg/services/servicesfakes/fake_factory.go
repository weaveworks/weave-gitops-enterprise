// Code generated by counterfeiter. DO NOT EDIT.
package servicesfakes

import (
	"context"
	"sync"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/git"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/services"
)

type FakeFactory struct {
	GetGitClientsStub        func(context.Context, *kube.KubeHTTP, gitproviders.Client, services.GitConfigParams) (git.Git, gitproviders.GitProvider, error)
	getGitClientsMutex       sync.RWMutex
	getGitClientsArgsForCall []struct {
		arg1 context.Context
		arg2 *kube.KubeHTTP
		arg3 gitproviders.Client
		arg4 services.GitConfigParams
	}
	getGitClientsReturns struct {
		result1 git.Git
		result2 gitproviders.GitProvider
		result3 error
	}
	getGitClientsReturnsOnCall map[int]struct {
		result1 git.Git
		result2 gitproviders.GitProvider
		result3 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeFactory) GetGitClients(arg1 context.Context, arg2 *kube.KubeHTTP, arg3 gitproviders.Client, arg4 services.GitConfigParams) (git.Git, gitproviders.GitProvider, error) {
	fake.getGitClientsMutex.Lock()
	ret, specificReturn := fake.getGitClientsReturnsOnCall[len(fake.getGitClientsArgsForCall)]
	fake.getGitClientsArgsForCall = append(fake.getGitClientsArgsForCall, struct {
		arg1 context.Context
		arg2 *kube.KubeHTTP
		arg3 gitproviders.Client
		arg4 services.GitConfigParams
	}{arg1, arg2, arg3, arg4})
	stub := fake.GetGitClientsStub
	fakeReturns := fake.getGitClientsReturns
	fake.recordInvocation("GetGitClients", []interface{}{arg1, arg2, arg3, arg4})
	fake.getGitClientsMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3
}

func (fake *FakeFactory) GetGitClientsCallCount() int {
	fake.getGitClientsMutex.RLock()
	defer fake.getGitClientsMutex.RUnlock()
	return len(fake.getGitClientsArgsForCall)
}

func (fake *FakeFactory) GetGitClientsCalls(stub func(context.Context, *kube.KubeHTTP, gitproviders.Client, services.GitConfigParams) (git.Git, gitproviders.GitProvider, error)) {
	fake.getGitClientsMutex.Lock()
	defer fake.getGitClientsMutex.Unlock()
	fake.GetGitClientsStub = stub
}

func (fake *FakeFactory) GetGitClientsArgsForCall(i int) (context.Context, *kube.KubeHTTP, gitproviders.Client, services.GitConfigParams) {
	fake.getGitClientsMutex.RLock()
	defer fake.getGitClientsMutex.RUnlock()
	argsForCall := fake.getGitClientsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakeFactory) GetGitClientsReturns(result1 git.Git, result2 gitproviders.GitProvider, result3 error) {
	fake.getGitClientsMutex.Lock()
	defer fake.getGitClientsMutex.Unlock()
	fake.GetGitClientsStub = nil
	fake.getGitClientsReturns = struct {
		result1 git.Git
		result2 gitproviders.GitProvider
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeFactory) GetGitClientsReturnsOnCall(i int, result1 git.Git, result2 gitproviders.GitProvider, result3 error) {
	fake.getGitClientsMutex.Lock()
	defer fake.getGitClientsMutex.Unlock()
	fake.GetGitClientsStub = nil
	if fake.getGitClientsReturnsOnCall == nil {
		fake.getGitClientsReturnsOnCall = make(map[int]struct {
			result1 git.Git
			result2 gitproviders.GitProvider
			result3 error
		})
	}
	fake.getGitClientsReturnsOnCall[i] = struct {
		result1 git.Git
		result2 gitproviders.GitProvider
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeFactory) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getGitClientsMutex.RLock()
	defer fake.getGitClientsMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeFactory) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ services.Factory = new(FakeFactory)
