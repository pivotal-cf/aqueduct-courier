// Code generated by counterfeiter. DO NOT EDIT.
package consumptionfakes

import (
	"io"
	"sync"
)

type FakeConsumptionService struct {
	AppUsagesStub        func() (io.Reader, error)
	appUsagesMutex       sync.RWMutex
	appUsagesArgsForCall []struct{}
	appUsagesReturns     struct {
		result1 io.Reader
		result2 error
	}
	appUsagesReturnsOnCall map[int]struct {
		result1 io.Reader
		result2 error
	}
	ServiceUsagesStub        func() (io.Reader, error)
	serviceUsagesMutex       sync.RWMutex
	serviceUsagesArgsForCall []struct{}
	serviceUsagesReturns     struct {
		result1 io.Reader
		result2 error
	}
	serviceUsagesReturnsOnCall map[int]struct {
		result1 io.Reader
		result2 error
	}
	TaskUsagesStub        func() (io.Reader, error)
	taskUsagesMutex       sync.RWMutex
	taskUsagesArgsForCall []struct{}
	taskUsagesReturns     struct {
		result1 io.Reader
		result2 error
	}
	taskUsagesReturnsOnCall map[int]struct {
		result1 io.Reader
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeConsumptionService) AppUsages() (io.Reader, error) {
	fake.appUsagesMutex.Lock()
	ret, specificReturn := fake.appUsagesReturnsOnCall[len(fake.appUsagesArgsForCall)]
	fake.appUsagesArgsForCall = append(fake.appUsagesArgsForCall, struct{}{})
	fake.recordInvocation("AppUsages", []interface{}{})
	fake.appUsagesMutex.Unlock()
	if fake.AppUsagesStub != nil {
		return fake.AppUsagesStub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.appUsagesReturns.result1, fake.appUsagesReturns.result2
}

func (fake *FakeConsumptionService) AppUsagesCallCount() int {
	fake.appUsagesMutex.RLock()
	defer fake.appUsagesMutex.RUnlock()
	return len(fake.appUsagesArgsForCall)
}

func (fake *FakeConsumptionService) AppUsagesReturns(result1 io.Reader, result2 error) {
	fake.AppUsagesStub = nil
	fake.appUsagesReturns = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeConsumptionService) AppUsagesReturnsOnCall(i int, result1 io.Reader, result2 error) {
	fake.AppUsagesStub = nil
	if fake.appUsagesReturnsOnCall == nil {
		fake.appUsagesReturnsOnCall = make(map[int]struct {
			result1 io.Reader
			result2 error
		})
	}
	fake.appUsagesReturnsOnCall[i] = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeConsumptionService) ServiceUsages() (io.Reader, error) {
	fake.serviceUsagesMutex.Lock()
	ret, specificReturn := fake.serviceUsagesReturnsOnCall[len(fake.serviceUsagesArgsForCall)]
	fake.serviceUsagesArgsForCall = append(fake.serviceUsagesArgsForCall, struct{}{})
	fake.recordInvocation("ServiceUsages", []interface{}{})
	fake.serviceUsagesMutex.Unlock()
	if fake.ServiceUsagesStub != nil {
		return fake.ServiceUsagesStub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.serviceUsagesReturns.result1, fake.serviceUsagesReturns.result2
}

func (fake *FakeConsumptionService) ServiceUsagesCallCount() int {
	fake.serviceUsagesMutex.RLock()
	defer fake.serviceUsagesMutex.RUnlock()
	return len(fake.serviceUsagesArgsForCall)
}

func (fake *FakeConsumptionService) ServiceUsagesReturns(result1 io.Reader, result2 error) {
	fake.ServiceUsagesStub = nil
	fake.serviceUsagesReturns = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeConsumptionService) ServiceUsagesReturnsOnCall(i int, result1 io.Reader, result2 error) {
	fake.ServiceUsagesStub = nil
	if fake.serviceUsagesReturnsOnCall == nil {
		fake.serviceUsagesReturnsOnCall = make(map[int]struct {
			result1 io.Reader
			result2 error
		})
	}
	fake.serviceUsagesReturnsOnCall[i] = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeConsumptionService) TaskUsages() (io.Reader, error) {
	fake.taskUsagesMutex.Lock()
	ret, specificReturn := fake.taskUsagesReturnsOnCall[len(fake.taskUsagesArgsForCall)]
	fake.taskUsagesArgsForCall = append(fake.taskUsagesArgsForCall, struct{}{})
	fake.recordInvocation("TaskUsages", []interface{}{})
	fake.taskUsagesMutex.Unlock()
	if fake.TaskUsagesStub != nil {
		return fake.TaskUsagesStub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.taskUsagesReturns.result1, fake.taskUsagesReturns.result2
}

func (fake *FakeConsumptionService) TaskUsagesCallCount() int {
	fake.taskUsagesMutex.RLock()
	defer fake.taskUsagesMutex.RUnlock()
	return len(fake.taskUsagesArgsForCall)
}

func (fake *FakeConsumptionService) TaskUsagesReturns(result1 io.Reader, result2 error) {
	fake.TaskUsagesStub = nil
	fake.taskUsagesReturns = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeConsumptionService) TaskUsagesReturnsOnCall(i int, result1 io.Reader, result2 error) {
	fake.TaskUsagesStub = nil
	if fake.taskUsagesReturnsOnCall == nil {
		fake.taskUsagesReturnsOnCall = make(map[int]struct {
			result1 io.Reader
			result2 error
		})
	}
	fake.taskUsagesReturnsOnCall[i] = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeConsumptionService) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.appUsagesMutex.RLock()
	defer fake.appUsagesMutex.RUnlock()
	fake.serviceUsagesMutex.RLock()
	defer fake.serviceUsagesMutex.RUnlock()
	fake.taskUsagesMutex.RLock()
	defer fake.taskUsagesMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeConsumptionService) recordInvocation(key string, args []interface{}) {
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
