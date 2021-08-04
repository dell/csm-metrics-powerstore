// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/dell/csm-metrics-powerstore/internal/service (interfaces: MetricsRecorder,Float64UpDownCounterCreater)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	metric "go.opentelemetry.io/otel/api/metric"
	reflect "reflect"
)

// MockMetricsRecorder is a mock of MetricsRecorder interface
type MockMetricsRecorder struct {
	ctrl     *gomock.Controller
	recorder *MockMetricsRecorderMockRecorder
}

// MockMetricsRecorderMockRecorder is the mock recorder for MockMetricsRecorder
type MockMetricsRecorderMockRecorder struct {
	mock *MockMetricsRecorder
}

// NewMockMetricsRecorder creates a new mock instance
func NewMockMetricsRecorder(ctrl *gomock.Controller) *MockMetricsRecorder {
	mock := &MockMetricsRecorder{ctrl: ctrl}
	mock.recorder = &MockMetricsRecorderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMetricsRecorder) EXPECT() *MockMetricsRecorderMockRecorder {
	return m.recorder
}

// Record mocks base method
func (m *MockMetricsRecorder) Record(arg0 context.Context, arg1 interface{}, arg2, arg3, arg4, arg5, arg6, arg7 float32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Record", arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	ret0, _ := ret[0].(error)
	return ret0
}

// Record indicates an expected call of Record
func (mr *MockMetricsRecorderMockRecorder) Record(arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Record", reflect.TypeOf((*MockMetricsRecorder)(nil).Record), arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
}

// RecordArraySpaceMetrics mocks base method
func (m *MockMetricsRecorder) RecordArraySpaceMetrics(arg0 context.Context, arg1, arg2 string, arg3, arg4 int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecordArraySpaceMetrics", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecordArraySpaceMetrics indicates an expected call of RecordArraySpaceMetrics
func (mr *MockMetricsRecorderMockRecorder) RecordArraySpaceMetrics(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordArraySpaceMetrics", reflect.TypeOf((*MockMetricsRecorder)(nil).RecordArraySpaceMetrics), arg0, arg1, arg2, arg3, arg4)
}

// RecordFileSystemMetrics mocks base method
func (m *MockMetricsRecorder) RecordFileSystemMetrics(arg0 context.Context, arg1 interface{}, arg2, arg3, arg4, arg5, arg6, arg7 float32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecordFileSystemMetrics", arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecordFileSystemMetrics indicates an expected call of RecordFileSystemMetrics
func (mr *MockMetricsRecorderMockRecorder) RecordFileSystemMetrics(arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordFileSystemMetrics", reflect.TypeOf((*MockMetricsRecorder)(nil).RecordFileSystemMetrics), arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
}

// RecordSpaceMetrics mocks base method
func (m *MockMetricsRecorder) RecordSpaceMetrics(arg0 context.Context, arg1 interface{}, arg2, arg3 int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecordSpaceMetrics", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecordSpaceMetrics indicates an expected call of RecordSpaceMetrics
func (mr *MockMetricsRecorderMockRecorder) RecordSpaceMetrics(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordSpaceMetrics", reflect.TypeOf((*MockMetricsRecorder)(nil).RecordSpaceMetrics), arg0, arg1, arg2, arg3)
}

// RecordStorageClassSpaceMetrics mocks base method
func (m *MockMetricsRecorder) RecordStorageClassSpaceMetrics(arg0 context.Context, arg1, arg2 string, arg3, arg4 int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecordStorageClassSpaceMetrics", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecordStorageClassSpaceMetrics indicates an expected call of RecordStorageClassSpaceMetrics
func (mr *MockMetricsRecorderMockRecorder) RecordStorageClassSpaceMetrics(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordStorageClassSpaceMetrics", reflect.TypeOf((*MockMetricsRecorder)(nil).RecordStorageClassSpaceMetrics), arg0, arg1, arg2, arg3, arg4)
}

// MockFloat64UpDownCounterCreater is a mock of Float64UpDownCounterCreater interface
type MockFloat64UpDownCounterCreater struct {
	ctrl     *gomock.Controller
	recorder *MockFloat64UpDownCounterCreaterMockRecorder
}

// MockFloat64UpDownCounterCreaterMockRecorder is the mock recorder for MockFloat64UpDownCounterCreater
type MockFloat64UpDownCounterCreaterMockRecorder struct {
	mock *MockFloat64UpDownCounterCreater
}

// NewMockFloat64UpDownCounterCreater creates a new mock instance
func NewMockFloat64UpDownCounterCreater(ctrl *gomock.Controller) *MockFloat64UpDownCounterCreater {
	mock := &MockFloat64UpDownCounterCreater{ctrl: ctrl}
	mock.recorder = &MockFloat64UpDownCounterCreaterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockFloat64UpDownCounterCreater) EXPECT() *MockFloat64UpDownCounterCreaterMockRecorder {
	return m.recorder
}

// NewFloat64UpDownCounter mocks base method
func (m *MockFloat64UpDownCounterCreater) NewFloat64UpDownCounter(arg0 string, arg1 ...metric.InstrumentOption) (metric.Float64UpDownCounter, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "NewFloat64UpDownCounter", varargs...)
	ret0, _ := ret[0].(metric.Float64UpDownCounter)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewFloat64UpDownCounter indicates an expected call of NewFloat64UpDownCounter
func (mr *MockFloat64UpDownCounterCreaterMockRecorder) NewFloat64UpDownCounter(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewFloat64UpDownCounter", reflect.TypeOf((*MockFloat64UpDownCounterCreater)(nil).NewFloat64UpDownCounter), varargs...)
}
