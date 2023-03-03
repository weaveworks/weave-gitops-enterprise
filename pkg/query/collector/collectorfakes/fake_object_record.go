// Code generated by counterfeiter. DO NOT EDIT.
package collectorfakes

import (
	"sync"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FakeObjectRecord struct {
	ClusterNameStub        func() string
	clusterNameMutex       sync.RWMutex
	clusterNameArgsForCall []struct {
	}
	clusterNameReturns struct {
		result1 string
	}
	clusterNameReturnsOnCall map[int]struct {
		result1 string
	}
	ObjectStub        func() client.Object
	objectMutex       sync.RWMutex
	objectArgsForCall []struct {
	}
	objectReturns struct {
		result1 client.Object
	}
	objectReturnsOnCall map[int]struct {
		result1 client.Object
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeObjectRecord) ClusterName() string {
	fake.clusterNameMutex.Lock()
	ret, specificReturn := fake.clusterNameReturnsOnCall[len(fake.clusterNameArgsForCall)]
	fake.clusterNameArgsForCall = append(fake.clusterNameArgsForCall, struct {
	}{})
	stub := fake.ClusterNameStub
	fakeReturns := fake.clusterNameReturns
	fake.recordInvocation("ClusterName", []interface{}{})
	fake.clusterNameMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeObjectRecord) ClusterNameCallCount() int {
	fake.clusterNameMutex.RLock()
	defer fake.clusterNameMutex.RUnlock()
	return len(fake.clusterNameArgsForCall)
}

func (fake *FakeObjectRecord) ClusterNameCalls(stub func() string) {
	fake.clusterNameMutex.Lock()
	defer fake.clusterNameMutex.Unlock()
	fake.ClusterNameStub = stub
}

func (fake *FakeObjectRecord) ClusterNameReturns(result1 string) {
	fake.clusterNameMutex.Lock()
	defer fake.clusterNameMutex.Unlock()
	fake.ClusterNameStub = nil
	fake.clusterNameReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeObjectRecord) ClusterNameReturnsOnCall(i int, result1 string) {
	fake.clusterNameMutex.Lock()
	defer fake.clusterNameMutex.Unlock()
	fake.ClusterNameStub = nil
	if fake.clusterNameReturnsOnCall == nil {
		fake.clusterNameReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.clusterNameReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *FakeObjectRecord) Object() client.Object {
	fake.objectMutex.Lock()
	ret, specificReturn := fake.objectReturnsOnCall[len(fake.objectArgsForCall)]
	fake.objectArgsForCall = append(fake.objectArgsForCall, struct {
	}{})
	stub := fake.ObjectStub
	fakeReturns := fake.objectReturns
	fake.recordInvocation("Object", []interface{}{})
	fake.objectMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeObjectRecord) ObjectCallCount() int {
	fake.objectMutex.RLock()
	defer fake.objectMutex.RUnlock()
	return len(fake.objectArgsForCall)
}

func (fake *FakeObjectRecord) ObjectCalls(stub func() client.Object) {
	fake.objectMutex.Lock()
	defer fake.objectMutex.Unlock()
	fake.ObjectStub = stub
}

func (fake *FakeObjectRecord) ObjectReturns(result1 client.Object) {
	fake.objectMutex.Lock()
	defer fake.objectMutex.Unlock()
	fake.ObjectStub = nil
	fake.objectReturns = struct {
		result1 client.Object
	}{result1}
}

func (fake *FakeObjectRecord) ObjectReturnsOnCall(i int, result1 client.Object) {
	fake.objectMutex.Lock()
	defer fake.objectMutex.Unlock()
	fake.ObjectStub = nil
	if fake.objectReturnsOnCall == nil {
		fake.objectReturnsOnCall = make(map[int]struct {
			result1 client.Object
		})
	}
	fake.objectReturnsOnCall[i] = struct {
		result1 client.Object
	}{result1}
}

func (fake *FakeObjectRecord) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.clusterNameMutex.RLock()
	defer fake.clusterNameMutex.RUnlock()
	fake.objectMutex.RLock()
	defer fake.objectMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeObjectRecord) recordInvocation(key string, args []interface{}) {
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

var _ collector.ObjectRecord = new(FakeObjectRecord)
