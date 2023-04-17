// Code generated by counterfeiter. DO NOT EDIT.
package collectorfakes

import (
	"context"
	"sync"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/collector"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
)

type FakeCollector struct {
	StartStub        func(context.Context) error
	startMutex       sync.RWMutex
	startArgsForCall []struct {
		arg1 context.Context
	}
	startReturns struct {
		result1 error
	}
	startReturnsOnCall map[int]struct {
		result1 error
	}
	StatusStub        func(string) (string, error)
	statusMutex       sync.RWMutex
	statusArgsForCall []struct {
		arg1 string
	}
	statusReturns struct {
		result1 string
		result2 error
	}
	statusReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	StopStub        func(context.Context) error
	stopMutex       sync.RWMutex
	stopArgsForCall []struct {
		arg1 context.Context
	}
	stopReturns struct {
		result1 error
	}
	stopReturnsOnCall map[int]struct {
		result1 error
	}
	UnwatchStub        func(string) error
	unwatchMutex       sync.RWMutex
	unwatchArgsForCall []struct {
		arg1 string
	}
	unwatchReturns struct {
		result1 error
	}
	unwatchReturnsOnCall map[int]struct {
		result1 error
	}
	WatchStub        func(context.Context, cluster.Cluster) error
	watchMutex       sync.RWMutex
	watchArgsForCall []struct {
		arg1 context.Context
		arg2 cluster.Cluster
	}
	watchReturns struct {
		result1 error
	}
	watchReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeCollector) Start(arg1 context.Context) error {
	fake.startMutex.Lock()
	ret, specificReturn := fake.startReturnsOnCall[len(fake.startArgsForCall)]
	fake.startArgsForCall = append(fake.startArgsForCall, struct {
		arg1 context.Context
	}{arg1})
	stub := fake.StartStub
	fakeReturns := fake.startReturns
	fake.recordInvocation("Start", []interface{}{arg1})
	fake.startMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeCollector) StartCallCount() int {
	fake.startMutex.RLock()
	defer fake.startMutex.RUnlock()
	return len(fake.startArgsForCall)
}

func (fake *FakeCollector) StartCalls(stub func(context.Context) error) {
	fake.startMutex.Lock()
	defer fake.startMutex.Unlock()
	fake.StartStub = stub
}

func (fake *FakeCollector) StartArgsForCall(i int) context.Context {
	fake.startMutex.RLock()
	defer fake.startMutex.RUnlock()
	argsForCall := fake.startArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeCollector) StartReturns(result1 error) {
	fake.startMutex.Lock()
	defer fake.startMutex.Unlock()
	fake.StartStub = nil
	fake.startReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeCollector) StartReturnsOnCall(i int, result1 error) {
	fake.startMutex.Lock()
	defer fake.startMutex.Unlock()
	fake.StartStub = nil
	if fake.startReturnsOnCall == nil {
		fake.startReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.startReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeCollector) Status(arg1 string) (string, error) {
	fake.statusMutex.Lock()
	ret, specificReturn := fake.statusReturnsOnCall[len(fake.statusArgsForCall)]
	fake.statusArgsForCall = append(fake.statusArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.StatusStub
	fakeReturns := fake.statusReturns
	fake.recordInvocation("Status", []interface{}{arg1})
	fake.statusMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeCollector) StatusCallCount() int {
	fake.statusMutex.RLock()
	defer fake.statusMutex.RUnlock()
	return len(fake.statusArgsForCall)
}

func (fake *FakeCollector) StatusCalls(stub func(string) (string, error)) {
	fake.statusMutex.Lock()
	defer fake.statusMutex.Unlock()
	fake.StatusStub = stub
}

func (fake *FakeCollector) StatusArgsForCall(i int) string {
	fake.statusMutex.RLock()
	defer fake.statusMutex.RUnlock()
	argsForCall := fake.statusArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeCollector) StatusReturns(result1 string, result2 error) {
	fake.statusMutex.Lock()
	defer fake.statusMutex.Unlock()
	fake.StatusStub = nil
	fake.statusReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeCollector) StatusReturnsOnCall(i int, result1 string, result2 error) {
	fake.statusMutex.Lock()
	defer fake.statusMutex.Unlock()
	fake.StatusStub = nil
	if fake.statusReturnsOnCall == nil {
		fake.statusReturnsOnCall = make(map[int]struct {
			result1 string
			result2 error
		})
	}
	fake.statusReturnsOnCall[i] = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeCollector) Stop(arg1 context.Context) error {
	fake.stopMutex.Lock()
	ret, specificReturn := fake.stopReturnsOnCall[len(fake.stopArgsForCall)]
	fake.stopArgsForCall = append(fake.stopArgsForCall, struct {
		arg1 context.Context
	}{arg1})
	stub := fake.StopStub
	fakeReturns := fake.stopReturns
	fake.recordInvocation("Stop", []interface{}{arg1})
	fake.stopMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeCollector) StopCallCount() int {
	fake.stopMutex.RLock()
	defer fake.stopMutex.RUnlock()
	return len(fake.stopArgsForCall)
}

func (fake *FakeCollector) StopCalls(stub func(context.Context) error) {
	fake.stopMutex.Lock()
	defer fake.stopMutex.Unlock()
	fake.StopStub = stub
}

func (fake *FakeCollector) StopArgsForCall(i int) context.Context {
	fake.stopMutex.RLock()
	defer fake.stopMutex.RUnlock()
	argsForCall := fake.stopArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeCollector) StopReturns(result1 error) {
	fake.stopMutex.Lock()
	defer fake.stopMutex.Unlock()
	fake.StopStub = nil
	fake.stopReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeCollector) StopReturnsOnCall(i int, result1 error) {
	fake.stopMutex.Lock()
	defer fake.stopMutex.Unlock()
	fake.StopStub = nil
	if fake.stopReturnsOnCall == nil {
		fake.stopReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.stopReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeCollector) Unwatch(arg1 string) error {
	fake.unwatchMutex.Lock()
	ret, specificReturn := fake.unwatchReturnsOnCall[len(fake.unwatchArgsForCall)]
	fake.unwatchArgsForCall = append(fake.unwatchArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.UnwatchStub
	fakeReturns := fake.unwatchReturns
	fake.recordInvocation("Unwatch", []interface{}{arg1})
	fake.unwatchMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeCollector) UnwatchCallCount() int {
	fake.unwatchMutex.RLock()
	defer fake.unwatchMutex.RUnlock()
	return len(fake.unwatchArgsForCall)
}

func (fake *FakeCollector) UnwatchCalls(stub func(string) error) {
	fake.unwatchMutex.Lock()
	defer fake.unwatchMutex.Unlock()
	fake.UnwatchStub = stub
}

func (fake *FakeCollector) UnwatchArgsForCall(i int) string {
	fake.unwatchMutex.RLock()
	defer fake.unwatchMutex.RUnlock()
	argsForCall := fake.unwatchArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeCollector) UnwatchReturns(result1 error) {
	fake.unwatchMutex.Lock()
	defer fake.unwatchMutex.Unlock()
	fake.UnwatchStub = nil
	fake.unwatchReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeCollector) UnwatchReturnsOnCall(i int, result1 error) {
	fake.unwatchMutex.Lock()
	defer fake.unwatchMutex.Unlock()
	fake.UnwatchStub = nil
	if fake.unwatchReturnsOnCall == nil {
		fake.unwatchReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.unwatchReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeCollector) Watch(arg1 context.Context, arg2 cluster.Cluster) error {
	fake.watchMutex.Lock()
	ret, specificReturn := fake.watchReturnsOnCall[len(fake.watchArgsForCall)]
	fake.watchArgsForCall = append(fake.watchArgsForCall, struct {
		arg1 context.Context
		arg2 cluster.Cluster
	}{arg1, arg2})
	stub := fake.WatchStub
	fakeReturns := fake.watchReturns
	fake.recordInvocation("Watch", []interface{}{arg1, arg2})
	fake.watchMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeCollector) WatchCallCount() int {
	fake.watchMutex.RLock()
	defer fake.watchMutex.RUnlock()
	return len(fake.watchArgsForCall)
}

func (fake *FakeCollector) WatchCalls(stub func(context.Context, cluster.Cluster) error) {
	fake.watchMutex.Lock()
	defer fake.watchMutex.Unlock()
	fake.WatchStub = stub
}

func (fake *FakeCollector) WatchArgsForCall(i int) (context.Context, cluster.Cluster) {
	fake.watchMutex.RLock()
	defer fake.watchMutex.RUnlock()
	argsForCall := fake.watchArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeCollector) WatchReturns(result1 error) {
	fake.watchMutex.Lock()
	defer fake.watchMutex.Unlock()
	fake.WatchStub = nil
	fake.watchReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeCollector) WatchReturnsOnCall(i int, result1 error) {
	fake.watchMutex.Lock()
	defer fake.watchMutex.Unlock()
	fake.WatchStub = nil
	if fake.watchReturnsOnCall == nil {
		fake.watchReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.watchReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeCollector) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.startMutex.RLock()
	defer fake.startMutex.RUnlock()
	fake.statusMutex.RLock()
	defer fake.statusMutex.RUnlock()
	fake.stopMutex.RLock()
	defer fake.stopMutex.RUnlock()
	fake.unwatchMutex.RLock()
	defer fake.unwatchMutex.RUnlock()
	fake.watchMutex.RLock()
	defer fake.watchMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeCollector) recordInvocation(key string, args []interface{}) {
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

var _ collector.Collector = new(FakeCollector)
