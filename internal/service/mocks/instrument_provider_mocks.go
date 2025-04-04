// Code generated by MockGen. DO NOT EDIT.
// Source: go.opentelemetry.io/otel/metric (interfaces: Float64ObservableUpDownCounter)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockFloat64ObservableUpDownCounter is a mock of Float64ObservableUpDownCounter interface.
type MockFloat64ObservableUpDownCounter struct {
	ctrl     *gomock.Controller
	recorder *MockFloat64ObservableUpDownCounterMockRecorder
}

// MockFloat64ObservableUpDownCounterMockRecorder is the mock recorder for MockFloat64ObservableUpDownCounter.
type MockFloat64ObservableUpDownCounterMockRecorder struct {
	mock *MockFloat64ObservableUpDownCounter
}

// NewMockFloat64ObservableUpDownCounter creates a new mock instance.
func NewMockFloat64ObservableUpDownCounter(ctrl *gomock.Controller) *MockFloat64ObservableUpDownCounter {
	mock := &MockFloat64ObservableUpDownCounter{ctrl: ctrl}
	mock.recorder = &MockFloat64ObservableUpDownCounterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFloat64ObservableUpDownCounter) EXPECT() *MockFloat64ObservableUpDownCounterMockRecorder {
	return m.recorder
}

// float64Observable mocks base method.
func (m *MockFloat64ObservableUpDownCounter) float64Observable() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "float64Observable")
}

// float64Observable indicates an expected call of float64Observable.
func (mr *MockFloat64ObservableUpDownCounterMockRecorder) float64Observable() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "float64Observable", reflect.TypeOf((*MockFloat64ObservableUpDownCounter)(nil).float64Observable))
}

// float64ObservableUpDownCounter mocks base method.
func (m *MockFloat64ObservableUpDownCounter) float64ObservableUpDownCounter() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "float64ObservableUpDownCounter")
}

// float64ObservableUpDownCounter indicates an expected call of float64ObservableUpDownCounter.
func (mr *MockFloat64ObservableUpDownCounterMockRecorder) float64ObservableUpDownCounter() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "float64ObservableUpDownCounter", reflect.TypeOf((*MockFloat64ObservableUpDownCounter)(nil).float64ObservableUpDownCounter))
}

// observable mocks base method.
func (m *MockFloat64ObservableUpDownCounter) observable() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "observable")
}

// observable indicates an expected call of observable.
func (mr *MockFloat64ObservableUpDownCounterMockRecorder) observable() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "observable", reflect.TypeOf((*MockFloat64ObservableUpDownCounter)(nil).observable))
}
