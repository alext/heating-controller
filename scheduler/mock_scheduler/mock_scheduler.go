// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/alext/heating-controller/scheduler (interfaces: Scheduler)

package mock_scheduler

import (
	scheduler "github.com/alext/heating-controller/scheduler"
	time "time"
	gomock "code.google.com/p/gomock/gomock"
)

// Mock of Scheduler interface
type MockScheduler struct {
	ctrl     *gomock.Controller
	recorder *_MockSchedulerRecorder
}

// Recorder for MockScheduler (not exported)
type _MockSchedulerRecorder struct {
	mock *MockScheduler
}

func NewMockScheduler(ctrl *gomock.Controller) *MockScheduler {
	mock := &MockScheduler{ctrl: ctrl}
	mock.recorder = &_MockSchedulerRecorder{mock}
	return mock
}

func (_m *MockScheduler) EXPECT() *_MockSchedulerRecorder {
	return _m.recorder
}

func (_m *MockScheduler) AddEvent(_param0 scheduler.Event) {
	_m.ctrl.Call(_m, "AddEvent", _param0)
}

func (_mr *_MockSchedulerRecorder) AddEvent(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "AddEvent", arg0)
}

func (_m *MockScheduler) Boost(_param0 time.Duration) {
	_m.ctrl.Call(_m, "Boost", _param0)
}

func (_mr *_MockSchedulerRecorder) Boost(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Boost", arg0)
}

func (_m *MockScheduler) Boosted() bool {
	ret := _m.ctrl.Call(_m, "Boosted")
	ret0, _ := ret[0].(bool)
	return ret0
}

func (_mr *_MockSchedulerRecorder) Boosted() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Boosted")
}

func (_m *MockScheduler) CancelBoost() {
	_m.ctrl.Call(_m, "CancelBoost")
}

func (_mr *_MockSchedulerRecorder) CancelBoost() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "CancelBoost")
}

func (_m *MockScheduler) NextEvent() *scheduler.Event {
	ret := _m.ctrl.Call(_m, "NextEvent")
	ret0, _ := ret[0].(*scheduler.Event)
	return ret0
}

func (_mr *_MockSchedulerRecorder) NextEvent() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "NextEvent")
}

func (_m *MockScheduler) ReadEvents() []scheduler.Event {
	ret := _m.ctrl.Call(_m, "ReadEvents")
	ret0, _ := ret[0].([]scheduler.Event)
	return ret0
}

func (_mr *_MockSchedulerRecorder) ReadEvents() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ReadEvents")
}

func (_m *MockScheduler) Running() bool {
	ret := _m.ctrl.Call(_m, "Running")
	ret0, _ := ret[0].(bool)
	return ret0
}

func (_mr *_MockSchedulerRecorder) Running() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Running")
}

func (_m *MockScheduler) Start() {
	_m.ctrl.Call(_m, "Start")
}

func (_mr *_MockSchedulerRecorder) Start() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Start")
}

func (_m *MockScheduler) Stop() {
	_m.ctrl.Call(_m, "Stop")
}

func (_mr *_MockSchedulerRecorder) Stop() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Stop")
}
