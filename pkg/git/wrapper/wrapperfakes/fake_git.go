// Code generated by counterfeiter. DO NOT EDIT.
package wrapperfakes

import (
	"context"
	"sync"

	git "github.com/go-git/go-git/v5"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/git/wrapper"
)

type FakeGit struct {
	PlainCloneContextStub        func(context.Context, string, bool, *git.CloneOptions) (*git.Repository, error)
	plainCloneContextMutex       sync.RWMutex
	plainCloneContextArgsForCall []struct {
		arg1 context.Context
		arg2 string
		arg3 bool
		arg4 *git.CloneOptions
	}
	plainCloneContextReturns struct {
		result1 *git.Repository
		result2 error
	}
	plainCloneContextReturnsOnCall map[int]struct {
		result1 *git.Repository
		result2 error
	}
	PlainInitStub        func(string, bool) (*git.Repository, error)
	plainInitMutex       sync.RWMutex
	plainInitArgsForCall []struct {
		arg1 string
		arg2 bool
	}
	plainInitReturns struct {
		result1 *git.Repository
		result2 error
	}
	plainInitReturnsOnCall map[int]struct {
		result1 *git.Repository
		result2 error
	}
	PlainOpenStub        func(string) (*git.Repository, error)
	plainOpenMutex       sync.RWMutex
	plainOpenArgsForCall []struct {
		arg1 string
	}
	plainOpenReturns struct {
		result1 *git.Repository
		result2 error
	}
	plainOpenReturnsOnCall map[int]struct {
		result1 *git.Repository
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeGit) PlainCloneContext(arg1 context.Context, arg2 string, arg3 bool, arg4 *git.CloneOptions) (*git.Repository, error) {
	fake.plainCloneContextMutex.Lock()
	ret, specificReturn := fake.plainCloneContextReturnsOnCall[len(fake.plainCloneContextArgsForCall)]
	fake.plainCloneContextArgsForCall = append(fake.plainCloneContextArgsForCall, struct {
		arg1 context.Context
		arg2 string
		arg3 bool
		arg4 *git.CloneOptions
	}{arg1, arg2, arg3, arg4})
	stub := fake.PlainCloneContextStub
	fakeReturns := fake.plainCloneContextReturns
	fake.recordInvocation("PlainCloneContext", []interface{}{arg1, arg2, arg3, arg4})
	fake.plainCloneContextMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeGit) PlainCloneContextCallCount() int {
	fake.plainCloneContextMutex.RLock()
	defer fake.plainCloneContextMutex.RUnlock()
	return len(fake.plainCloneContextArgsForCall)
}

func (fake *FakeGit) PlainCloneContextCalls(stub func(context.Context, string, bool, *git.CloneOptions) (*git.Repository, error)) {
	fake.plainCloneContextMutex.Lock()
	defer fake.plainCloneContextMutex.Unlock()
	fake.PlainCloneContextStub = stub
}

func (fake *FakeGit) PlainCloneContextArgsForCall(i int) (context.Context, string, bool, *git.CloneOptions) {
	fake.plainCloneContextMutex.RLock()
	defer fake.plainCloneContextMutex.RUnlock()
	argsForCall := fake.plainCloneContextArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakeGit) PlainCloneContextReturns(result1 *git.Repository, result2 error) {
	fake.plainCloneContextMutex.Lock()
	defer fake.plainCloneContextMutex.Unlock()
	fake.PlainCloneContextStub = nil
	fake.plainCloneContextReturns = struct {
		result1 *git.Repository
		result2 error
	}{result1, result2}
}

func (fake *FakeGit) PlainCloneContextReturnsOnCall(i int, result1 *git.Repository, result2 error) {
	fake.plainCloneContextMutex.Lock()
	defer fake.plainCloneContextMutex.Unlock()
	fake.PlainCloneContextStub = nil
	if fake.plainCloneContextReturnsOnCall == nil {
		fake.plainCloneContextReturnsOnCall = make(map[int]struct {
			result1 *git.Repository
			result2 error
		})
	}
	fake.plainCloneContextReturnsOnCall[i] = struct {
		result1 *git.Repository
		result2 error
	}{result1, result2}
}

func (fake *FakeGit) PlainInit(arg1 string, arg2 bool) (*git.Repository, error) {
	fake.plainInitMutex.Lock()
	ret, specificReturn := fake.plainInitReturnsOnCall[len(fake.plainInitArgsForCall)]
	fake.plainInitArgsForCall = append(fake.plainInitArgsForCall, struct {
		arg1 string
		arg2 bool
	}{arg1, arg2})
	stub := fake.PlainInitStub
	fakeReturns := fake.plainInitReturns
	fake.recordInvocation("PlainInit", []interface{}{arg1, arg2})
	fake.plainInitMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeGit) PlainInitCallCount() int {
	fake.plainInitMutex.RLock()
	defer fake.plainInitMutex.RUnlock()
	return len(fake.plainInitArgsForCall)
}

func (fake *FakeGit) PlainInitCalls(stub func(string, bool) (*git.Repository, error)) {
	fake.plainInitMutex.Lock()
	defer fake.plainInitMutex.Unlock()
	fake.PlainInitStub = stub
}

func (fake *FakeGit) PlainInitArgsForCall(i int) (string, bool) {
	fake.plainInitMutex.RLock()
	defer fake.plainInitMutex.RUnlock()
	argsForCall := fake.plainInitArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeGit) PlainInitReturns(result1 *git.Repository, result2 error) {
	fake.plainInitMutex.Lock()
	defer fake.plainInitMutex.Unlock()
	fake.PlainInitStub = nil
	fake.plainInitReturns = struct {
		result1 *git.Repository
		result2 error
	}{result1, result2}
}

func (fake *FakeGit) PlainInitReturnsOnCall(i int, result1 *git.Repository, result2 error) {
	fake.plainInitMutex.Lock()
	defer fake.plainInitMutex.Unlock()
	fake.PlainInitStub = nil
	if fake.plainInitReturnsOnCall == nil {
		fake.plainInitReturnsOnCall = make(map[int]struct {
			result1 *git.Repository
			result2 error
		})
	}
	fake.plainInitReturnsOnCall[i] = struct {
		result1 *git.Repository
		result2 error
	}{result1, result2}
}

func (fake *FakeGit) PlainOpen(arg1 string) (*git.Repository, error) {
	fake.plainOpenMutex.Lock()
	ret, specificReturn := fake.plainOpenReturnsOnCall[len(fake.plainOpenArgsForCall)]
	fake.plainOpenArgsForCall = append(fake.plainOpenArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.PlainOpenStub
	fakeReturns := fake.plainOpenReturns
	fake.recordInvocation("PlainOpen", []interface{}{arg1})
	fake.plainOpenMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeGit) PlainOpenCallCount() int {
	fake.plainOpenMutex.RLock()
	defer fake.plainOpenMutex.RUnlock()
	return len(fake.plainOpenArgsForCall)
}

func (fake *FakeGit) PlainOpenCalls(stub func(string) (*git.Repository, error)) {
	fake.plainOpenMutex.Lock()
	defer fake.plainOpenMutex.Unlock()
	fake.PlainOpenStub = stub
}

func (fake *FakeGit) PlainOpenArgsForCall(i int) string {
	fake.plainOpenMutex.RLock()
	defer fake.plainOpenMutex.RUnlock()
	argsForCall := fake.plainOpenArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeGit) PlainOpenReturns(result1 *git.Repository, result2 error) {
	fake.plainOpenMutex.Lock()
	defer fake.plainOpenMutex.Unlock()
	fake.PlainOpenStub = nil
	fake.plainOpenReturns = struct {
		result1 *git.Repository
		result2 error
	}{result1, result2}
}

func (fake *FakeGit) PlainOpenReturnsOnCall(i int, result1 *git.Repository, result2 error) {
	fake.plainOpenMutex.Lock()
	defer fake.plainOpenMutex.Unlock()
	fake.PlainOpenStub = nil
	if fake.plainOpenReturnsOnCall == nil {
		fake.plainOpenReturnsOnCall = make(map[int]struct {
			result1 *git.Repository
			result2 error
		})
	}
	fake.plainOpenReturnsOnCall[i] = struct {
		result1 *git.Repository
		result2 error
	}{result1, result2}
}

func (fake *FakeGit) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.plainCloneContextMutex.RLock()
	defer fake.plainCloneContextMutex.RUnlock()
	fake.plainInitMutex.RLock()
	defer fake.plainInitMutex.RUnlock()
	fake.plainOpenMutex.RLock()
	defer fake.plainOpenMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeGit) recordInvocation(key string, args []interface{}) {
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

var _ wrapper.Git = new(FakeGit)
