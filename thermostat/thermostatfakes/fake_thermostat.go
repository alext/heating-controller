// Code generated by counterfeiter. DO NOT EDIT.
package thermostatfakes

import (
	"sync"

	"github.com/alext/heating-controller/thermostat"
	"github.com/alext/heating-controller/units"
)

type FakeThermostat struct {
	CurrentStub        func() units.Temperature
	currentMutex       sync.RWMutex
	currentArgsForCall []struct{}
	currentReturns     struct {
		result1 units.Temperature
	}
	currentReturnsOnCall map[int]struct {
		result1 units.Temperature
	}
	TargetStub        func() units.Temperature
	targetMutex       sync.RWMutex
	targetArgsForCall []struct{}
	targetReturns     struct {
		result1 units.Temperature
	}
	targetReturnsOnCall map[int]struct {
		result1 units.Temperature
	}
	SetStub        func(units.Temperature)
	setMutex       sync.RWMutex
	setArgsForCall []struct {
		arg1 units.Temperature
	}
	CloseStub        func()
	closeMutex       sync.RWMutex
	closeArgsForCall []struct{}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeThermostat) Current() units.Temperature {
	fake.currentMutex.Lock()
	ret, specificReturn := fake.currentReturnsOnCall[len(fake.currentArgsForCall)]
	fake.currentArgsForCall = append(fake.currentArgsForCall, struct{}{})
	fake.recordInvocation("Current", []interface{}{})
	fake.currentMutex.Unlock()
	if fake.CurrentStub != nil {
		return fake.CurrentStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.currentReturns.result1
}

func (fake *FakeThermostat) CurrentCallCount() int {
	fake.currentMutex.RLock()
	defer fake.currentMutex.RUnlock()
	return len(fake.currentArgsForCall)
}

func (fake *FakeThermostat) CurrentReturns(result1 units.Temperature) {
	fake.CurrentStub = nil
	fake.currentReturns = struct {
		result1 units.Temperature
	}{result1}
}

func (fake *FakeThermostat) CurrentReturnsOnCall(i int, result1 units.Temperature) {
	fake.CurrentStub = nil
	if fake.currentReturnsOnCall == nil {
		fake.currentReturnsOnCall = make(map[int]struct {
			result1 units.Temperature
		})
	}
	fake.currentReturnsOnCall[i] = struct {
		result1 units.Temperature
	}{result1}
}

func (fake *FakeThermostat) Target() units.Temperature {
	fake.targetMutex.Lock()
	ret, specificReturn := fake.targetReturnsOnCall[len(fake.targetArgsForCall)]
	fake.targetArgsForCall = append(fake.targetArgsForCall, struct{}{})
	fake.recordInvocation("Target", []interface{}{})
	fake.targetMutex.Unlock()
	if fake.TargetStub != nil {
		return fake.TargetStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.targetReturns.result1
}

func (fake *FakeThermostat) TargetCallCount() int {
	fake.targetMutex.RLock()
	defer fake.targetMutex.RUnlock()
	return len(fake.targetArgsForCall)
}

func (fake *FakeThermostat) TargetReturns(result1 units.Temperature) {
	fake.TargetStub = nil
	fake.targetReturns = struct {
		result1 units.Temperature
	}{result1}
}

func (fake *FakeThermostat) TargetReturnsOnCall(i int, result1 units.Temperature) {
	fake.TargetStub = nil
	if fake.targetReturnsOnCall == nil {
		fake.targetReturnsOnCall = make(map[int]struct {
			result1 units.Temperature
		})
	}
	fake.targetReturnsOnCall[i] = struct {
		result1 units.Temperature
	}{result1}
}

func (fake *FakeThermostat) Set(arg1 units.Temperature) {
	fake.setMutex.Lock()
	fake.setArgsForCall = append(fake.setArgsForCall, struct {
		arg1 units.Temperature
	}{arg1})
	fake.recordInvocation("Set", []interface{}{arg1})
	fake.setMutex.Unlock()
	if fake.SetStub != nil {
		fake.SetStub(arg1)
	}
}

func (fake *FakeThermostat) SetCallCount() int {
	fake.setMutex.RLock()
	defer fake.setMutex.RUnlock()
	return len(fake.setArgsForCall)
}

func (fake *FakeThermostat) SetArgsForCall(i int) units.Temperature {
	fake.setMutex.RLock()
	defer fake.setMutex.RUnlock()
	return fake.setArgsForCall[i].arg1
}

func (fake *FakeThermostat) Close() {
	fake.closeMutex.Lock()
	fake.closeArgsForCall = append(fake.closeArgsForCall, struct{}{})
	fake.recordInvocation("Close", []interface{}{})
	fake.closeMutex.Unlock()
	if fake.CloseStub != nil {
		fake.CloseStub()
	}
}

func (fake *FakeThermostat) CloseCallCount() int {
	fake.closeMutex.RLock()
	defer fake.closeMutex.RUnlock()
	return len(fake.closeArgsForCall)
}

func (fake *FakeThermostat) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.currentMutex.RLock()
	defer fake.currentMutex.RUnlock()
	fake.targetMutex.RLock()
	defer fake.targetMutex.RUnlock()
	fake.setMutex.RLock()
	defer fake.setMutex.RUnlock()
	fake.closeMutex.RLock()
	defer fake.closeMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeThermostat) recordInvocation(key string, args []interface{}) {
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

var _ thermostat.Thermostat = new(FakeThermostat)
