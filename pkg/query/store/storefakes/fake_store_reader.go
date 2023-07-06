// Code generated by counterfeiter. DO NOT EDIT.
package storefakes

import (
	"context"
	"sync"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/internal/models"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/query/store"
)

type FakeStoreReader struct {
	GetAccessRulesStub        func(context.Context) ([]models.AccessRule, error)
	getAccessRulesMutex       sync.RWMutex
	getAccessRulesArgsForCall []struct {
		arg1 context.Context
	}
	getAccessRulesReturns struct {
		result1 []models.AccessRule
		result2 error
	}
	getAccessRulesReturnsOnCall map[int]struct {
		result1 []models.AccessRule
		result2 error
	}
	GetAllObjectsStub        func(context.Context) (store.Iterator, error)
	getAllObjectsMutex       sync.RWMutex
	getAllObjectsArgsForCall []struct {
		arg1 context.Context
	}
	getAllObjectsReturns struct {
		result1 store.Iterator
		result2 error
	}
	getAllObjectsReturnsOnCall map[int]struct {
		result1 store.Iterator
		result2 error
	}
	GetObjectByIDStub        func(context.Context, string) (models.Object, error)
	getObjectByIDMutex       sync.RWMutex
	getObjectByIDArgsForCall []struct {
		arg1 context.Context
		arg2 string
	}
	getObjectByIDReturns struct {
		result1 models.Object
		result2 error
	}
	getObjectByIDReturnsOnCall map[int]struct {
		result1 models.Object
		result2 error
	}
	GetObjectsStub        func(context.Context, []string, store.QueryOption) (store.Iterator, error)
	getObjectsMutex       sync.RWMutex
	getObjectsArgsForCall []struct {
		arg1 context.Context
		arg2 []string
		arg3 store.QueryOption
	}
	getObjectsReturns struct {
		result1 store.Iterator
		result2 error
	}
	getObjectsReturnsOnCall map[int]struct {
		result1 store.Iterator
		result2 error
	}
	GetRoleBindingsStub        func(context.Context) ([]models.RoleBinding, error)
	getRoleBindingsMutex       sync.RWMutex
	getRoleBindingsArgsForCall []struct {
		arg1 context.Context
	}
	getRoleBindingsReturns struct {
		result1 []models.RoleBinding
		result2 error
	}
	getRoleBindingsReturnsOnCall map[int]struct {
		result1 []models.RoleBinding
		result2 error
	}
	GetRolesStub        func(context.Context) ([]models.Role, error)
	getRolesMutex       sync.RWMutex
	getRolesArgsForCall []struct {
		arg1 context.Context
	}
	getRolesReturns struct {
		result1 []models.Role
		result2 error
	}
	getRolesReturnsOnCall map[int]struct {
		result1 []models.Role
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeStoreReader) GetAccessRules(arg1 context.Context) ([]models.AccessRule, error) {
	fake.getAccessRulesMutex.Lock()
	ret, specificReturn := fake.getAccessRulesReturnsOnCall[len(fake.getAccessRulesArgsForCall)]
	fake.getAccessRulesArgsForCall = append(fake.getAccessRulesArgsForCall, struct {
		arg1 context.Context
	}{arg1})
	stub := fake.GetAccessRulesStub
	fakeReturns := fake.getAccessRulesReturns
	fake.recordInvocation("GetAccessRules", []interface{}{arg1})
	fake.getAccessRulesMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeStoreReader) GetAccessRulesCallCount() int {
	fake.getAccessRulesMutex.RLock()
	defer fake.getAccessRulesMutex.RUnlock()
	return len(fake.getAccessRulesArgsForCall)
}

func (fake *FakeStoreReader) GetAccessRulesCalls(stub func(context.Context) ([]models.AccessRule, error)) {
	fake.getAccessRulesMutex.Lock()
	defer fake.getAccessRulesMutex.Unlock()
	fake.GetAccessRulesStub = stub
}

func (fake *FakeStoreReader) GetAccessRulesArgsForCall(i int) context.Context {
	fake.getAccessRulesMutex.RLock()
	defer fake.getAccessRulesMutex.RUnlock()
	argsForCall := fake.getAccessRulesArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeStoreReader) GetAccessRulesReturns(result1 []models.AccessRule, result2 error) {
	fake.getAccessRulesMutex.Lock()
	defer fake.getAccessRulesMutex.Unlock()
	fake.GetAccessRulesStub = nil
	fake.getAccessRulesReturns = struct {
		result1 []models.AccessRule
		result2 error
	}{result1, result2}
}

func (fake *FakeStoreReader) GetAccessRulesReturnsOnCall(i int, result1 []models.AccessRule, result2 error) {
	fake.getAccessRulesMutex.Lock()
	defer fake.getAccessRulesMutex.Unlock()
	fake.GetAccessRulesStub = nil
	if fake.getAccessRulesReturnsOnCall == nil {
		fake.getAccessRulesReturnsOnCall = make(map[int]struct {
			result1 []models.AccessRule
			result2 error
		})
	}
	fake.getAccessRulesReturnsOnCall[i] = struct {
		result1 []models.AccessRule
		result2 error
	}{result1, result2}
}

func (fake *FakeStoreReader) GetAllObjects(arg1 context.Context) (store.Iterator, error) {
	fake.getAllObjectsMutex.Lock()
	ret, specificReturn := fake.getAllObjectsReturnsOnCall[len(fake.getAllObjectsArgsForCall)]
	fake.getAllObjectsArgsForCall = append(fake.getAllObjectsArgsForCall, struct {
		arg1 context.Context
	}{arg1})
	stub := fake.GetAllObjectsStub
	fakeReturns := fake.getAllObjectsReturns
	fake.recordInvocation("GetAllObjects", []interface{}{arg1})
	fake.getAllObjectsMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeStoreReader) GetAllObjectsCallCount() int {
	fake.getAllObjectsMutex.RLock()
	defer fake.getAllObjectsMutex.RUnlock()
	return len(fake.getAllObjectsArgsForCall)
}

func (fake *FakeStoreReader) GetAllObjectsCalls(stub func(context.Context) (store.Iterator, error)) {
	fake.getAllObjectsMutex.Lock()
	defer fake.getAllObjectsMutex.Unlock()
	fake.GetAllObjectsStub = stub
}

func (fake *FakeStoreReader) GetAllObjectsArgsForCall(i int) context.Context {
	fake.getAllObjectsMutex.RLock()
	defer fake.getAllObjectsMutex.RUnlock()
	argsForCall := fake.getAllObjectsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeStoreReader) GetAllObjectsReturns(result1 store.Iterator, result2 error) {
	fake.getAllObjectsMutex.Lock()
	defer fake.getAllObjectsMutex.Unlock()
	fake.GetAllObjectsStub = nil
	fake.getAllObjectsReturns = struct {
		result1 store.Iterator
		result2 error
	}{result1, result2}
}

func (fake *FakeStoreReader) GetAllObjectsReturnsOnCall(i int, result1 store.Iterator, result2 error) {
	fake.getAllObjectsMutex.Lock()
	defer fake.getAllObjectsMutex.Unlock()
	fake.GetAllObjectsStub = nil
	if fake.getAllObjectsReturnsOnCall == nil {
		fake.getAllObjectsReturnsOnCall = make(map[int]struct {
			result1 store.Iterator
			result2 error
		})
	}
	fake.getAllObjectsReturnsOnCall[i] = struct {
		result1 store.Iterator
		result2 error
	}{result1, result2}
}

func (fake *FakeStoreReader) GetObjectByID(arg1 context.Context, arg2 string) (models.Object, error) {
	fake.getObjectByIDMutex.Lock()
	ret, specificReturn := fake.getObjectByIDReturnsOnCall[len(fake.getObjectByIDArgsForCall)]
	fake.getObjectByIDArgsForCall = append(fake.getObjectByIDArgsForCall, struct {
		arg1 context.Context
		arg2 string
	}{arg1, arg2})
	stub := fake.GetObjectByIDStub
	fakeReturns := fake.getObjectByIDReturns
	fake.recordInvocation("GetObjectByID", []interface{}{arg1, arg2})
	fake.getObjectByIDMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeStoreReader) GetObjectByIDCallCount() int {
	fake.getObjectByIDMutex.RLock()
	defer fake.getObjectByIDMutex.RUnlock()
	return len(fake.getObjectByIDArgsForCall)
}

func (fake *FakeStoreReader) GetObjectByIDCalls(stub func(context.Context, string) (models.Object, error)) {
	fake.getObjectByIDMutex.Lock()
	defer fake.getObjectByIDMutex.Unlock()
	fake.GetObjectByIDStub = stub
}

func (fake *FakeStoreReader) GetObjectByIDArgsForCall(i int) (context.Context, string) {
	fake.getObjectByIDMutex.RLock()
	defer fake.getObjectByIDMutex.RUnlock()
	argsForCall := fake.getObjectByIDArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeStoreReader) GetObjectByIDReturns(result1 models.Object, result2 error) {
	fake.getObjectByIDMutex.Lock()
	defer fake.getObjectByIDMutex.Unlock()
	fake.GetObjectByIDStub = nil
	fake.getObjectByIDReturns = struct {
		result1 models.Object
		result2 error
	}{result1, result2}
}

func (fake *FakeStoreReader) GetObjectByIDReturnsOnCall(i int, result1 models.Object, result2 error) {
	fake.getObjectByIDMutex.Lock()
	defer fake.getObjectByIDMutex.Unlock()
	fake.GetObjectByIDStub = nil
	if fake.getObjectByIDReturnsOnCall == nil {
		fake.getObjectByIDReturnsOnCall = make(map[int]struct {
			result1 models.Object
			result2 error
		})
	}
	fake.getObjectByIDReturnsOnCall[i] = struct {
		result1 models.Object
		result2 error
	}{result1, result2}
}

func (fake *FakeStoreReader) GetObjects(arg1 context.Context, arg2 []string, arg3 store.QueryOption) (store.Iterator, error) {
	var arg2Copy []string
	if arg2 != nil {
		arg2Copy = make([]string, len(arg2))
		copy(arg2Copy, arg2)
	}
	fake.getObjectsMutex.Lock()
	ret, specificReturn := fake.getObjectsReturnsOnCall[len(fake.getObjectsArgsForCall)]
	fake.getObjectsArgsForCall = append(fake.getObjectsArgsForCall, struct {
		arg1 context.Context
		arg2 []string
		arg3 store.QueryOption
	}{arg1, arg2Copy, arg3})
	stub := fake.GetObjectsStub
	fakeReturns := fake.getObjectsReturns
	fake.recordInvocation("GetObjects", []interface{}{arg1, arg2Copy, arg3})
	fake.getObjectsMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeStoreReader) GetObjectsCallCount() int {
	fake.getObjectsMutex.RLock()
	defer fake.getObjectsMutex.RUnlock()
	return len(fake.getObjectsArgsForCall)
}

func (fake *FakeStoreReader) GetObjectsCalls(stub func(context.Context, []string, store.QueryOption) (store.Iterator, error)) {
	fake.getObjectsMutex.Lock()
	defer fake.getObjectsMutex.Unlock()
	fake.GetObjectsStub = stub
}

func (fake *FakeStoreReader) GetObjectsArgsForCall(i int) (context.Context, []string, store.QueryOption) {
	fake.getObjectsMutex.RLock()
	defer fake.getObjectsMutex.RUnlock()
	argsForCall := fake.getObjectsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeStoreReader) GetObjectsReturns(result1 store.Iterator, result2 error) {
	fake.getObjectsMutex.Lock()
	defer fake.getObjectsMutex.Unlock()
	fake.GetObjectsStub = nil
	fake.getObjectsReturns = struct {
		result1 store.Iterator
		result2 error
	}{result1, result2}
}

func (fake *FakeStoreReader) GetObjectsReturnsOnCall(i int, result1 store.Iterator, result2 error) {
	fake.getObjectsMutex.Lock()
	defer fake.getObjectsMutex.Unlock()
	fake.GetObjectsStub = nil
	if fake.getObjectsReturnsOnCall == nil {
		fake.getObjectsReturnsOnCall = make(map[int]struct {
			result1 store.Iterator
			result2 error
		})
	}
	fake.getObjectsReturnsOnCall[i] = struct {
		result1 store.Iterator
		result2 error
	}{result1, result2}
}

func (fake *FakeStoreReader) GetRoleBindings(arg1 context.Context) ([]models.RoleBinding, error) {
	fake.getRoleBindingsMutex.Lock()
	ret, specificReturn := fake.getRoleBindingsReturnsOnCall[len(fake.getRoleBindingsArgsForCall)]
	fake.getRoleBindingsArgsForCall = append(fake.getRoleBindingsArgsForCall, struct {
		arg1 context.Context
	}{arg1})
	stub := fake.GetRoleBindingsStub
	fakeReturns := fake.getRoleBindingsReturns
	fake.recordInvocation("GetRoleBindings", []interface{}{arg1})
	fake.getRoleBindingsMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeStoreReader) GetRoleBindingsCallCount() int {
	fake.getRoleBindingsMutex.RLock()
	defer fake.getRoleBindingsMutex.RUnlock()
	return len(fake.getRoleBindingsArgsForCall)
}

func (fake *FakeStoreReader) GetRoleBindingsCalls(stub func(context.Context) ([]models.RoleBinding, error)) {
	fake.getRoleBindingsMutex.Lock()
	defer fake.getRoleBindingsMutex.Unlock()
	fake.GetRoleBindingsStub = stub
}

func (fake *FakeStoreReader) GetRoleBindingsArgsForCall(i int) context.Context {
	fake.getRoleBindingsMutex.RLock()
	defer fake.getRoleBindingsMutex.RUnlock()
	argsForCall := fake.getRoleBindingsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeStoreReader) GetRoleBindingsReturns(result1 []models.RoleBinding, result2 error) {
	fake.getRoleBindingsMutex.Lock()
	defer fake.getRoleBindingsMutex.Unlock()
	fake.GetRoleBindingsStub = nil
	fake.getRoleBindingsReturns = struct {
		result1 []models.RoleBinding
		result2 error
	}{result1, result2}
}

func (fake *FakeStoreReader) GetRoleBindingsReturnsOnCall(i int, result1 []models.RoleBinding, result2 error) {
	fake.getRoleBindingsMutex.Lock()
	defer fake.getRoleBindingsMutex.Unlock()
	fake.GetRoleBindingsStub = nil
	if fake.getRoleBindingsReturnsOnCall == nil {
		fake.getRoleBindingsReturnsOnCall = make(map[int]struct {
			result1 []models.RoleBinding
			result2 error
		})
	}
	fake.getRoleBindingsReturnsOnCall[i] = struct {
		result1 []models.RoleBinding
		result2 error
	}{result1, result2}
}

func (fake *FakeStoreReader) GetRoles(arg1 context.Context) ([]models.Role, error) {
	fake.getRolesMutex.Lock()
	ret, specificReturn := fake.getRolesReturnsOnCall[len(fake.getRolesArgsForCall)]
	fake.getRolesArgsForCall = append(fake.getRolesArgsForCall, struct {
		arg1 context.Context
	}{arg1})
	stub := fake.GetRolesStub
	fakeReturns := fake.getRolesReturns
	fake.recordInvocation("GetRoles", []interface{}{arg1})
	fake.getRolesMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeStoreReader) GetRolesCallCount() int {
	fake.getRolesMutex.RLock()
	defer fake.getRolesMutex.RUnlock()
	return len(fake.getRolesArgsForCall)
}

func (fake *FakeStoreReader) GetRolesCalls(stub func(context.Context) ([]models.Role, error)) {
	fake.getRolesMutex.Lock()
	defer fake.getRolesMutex.Unlock()
	fake.GetRolesStub = stub
}

func (fake *FakeStoreReader) GetRolesArgsForCall(i int) context.Context {
	fake.getRolesMutex.RLock()
	defer fake.getRolesMutex.RUnlock()
	argsForCall := fake.getRolesArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeStoreReader) GetRolesReturns(result1 []models.Role, result2 error) {
	fake.getRolesMutex.Lock()
	defer fake.getRolesMutex.Unlock()
	fake.GetRolesStub = nil
	fake.getRolesReturns = struct {
		result1 []models.Role
		result2 error
	}{result1, result2}
}

func (fake *FakeStoreReader) GetRolesReturnsOnCall(i int, result1 []models.Role, result2 error) {
	fake.getRolesMutex.Lock()
	defer fake.getRolesMutex.Unlock()
	fake.GetRolesStub = nil
	if fake.getRolesReturnsOnCall == nil {
		fake.getRolesReturnsOnCall = make(map[int]struct {
			result1 []models.Role
			result2 error
		})
	}
	fake.getRolesReturnsOnCall[i] = struct {
		result1 []models.Role
		result2 error
	}{result1, result2}
}

func (fake *FakeStoreReader) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getAccessRulesMutex.RLock()
	defer fake.getAccessRulesMutex.RUnlock()
	fake.getAllObjectsMutex.RLock()
	defer fake.getAllObjectsMutex.RUnlock()
	fake.getObjectByIDMutex.RLock()
	defer fake.getObjectByIDMutex.RUnlock()
	fake.getObjectsMutex.RLock()
	defer fake.getObjectsMutex.RUnlock()
	fake.getRoleBindingsMutex.RLock()
	defer fake.getRoleBindingsMutex.RUnlock()
	fake.getRolesMutex.RLock()
	defer fake.getRolesMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeStoreReader) recordInvocation(key string, args []interface{}) {
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

var _ store.StoreReader = new(FakeStoreReader)
