/*
 Copyright (c) 2021-2022 Dell Inc. or its subsidiaries. All Rights Reserved.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/dell/csm-metrics-powerstore/internal/service (interfaces: MetricsRecorder,Float64UpDownCounterCreater)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	asyncfloat64 "go.opentelemetry.io/otel/metric/instrument/asyncfloat64"
)

// MockMetricsRecorder is a mock of MetricsRecorder interface.
type MockMetricsRecorder struct {
	ctrl     *gomock.Controller
	recorder *MockMetricsRecorderMockRecorder
}

// MockMetricsRecorderMockRecorder is the mock recorder for MockMetricsRecorder.
type MockMetricsRecorderMockRecorder struct {
	mock *MockMetricsRecorder
}

// NewMockMetricsRecorder creates a new mock instance.
func NewMockMetricsRecorder(ctrl *gomock.Controller) *MockMetricsRecorder {
	mock := &MockMetricsRecorder{ctrl: ctrl}
	mock.recorder = &MockMetricsRecorderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMetricsRecorder) EXPECT() *MockMetricsRecorderMockRecorder {
	return m.recorder
}

// Record mocks base method.
func (m *MockMetricsRecorder) Record(arg0 context.Context, arg1 interface{}, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10 float32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Record", arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10)
	ret0, _ := ret[0].(error)
	return ret0
}

// Record indicates an expected call of Record.
func (mr *MockMetricsRecorderMockRecorder) Record(arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Record", reflect.TypeOf((*MockMetricsRecorder)(nil).Record), arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10)
}

// RecordArraySpaceMetrics mocks base method.
func (m *MockMetricsRecorder) RecordArraySpaceMetrics(arg0 context.Context, arg1, arg2 string, arg3, arg4 int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecordArraySpaceMetrics", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecordArraySpaceMetrics indicates an expected call of RecordArraySpaceMetrics.
func (mr *MockMetricsRecorderMockRecorder) RecordArraySpaceMetrics(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordArraySpaceMetrics", reflect.TypeOf((*MockMetricsRecorder)(nil).RecordArraySpaceMetrics), arg0, arg1, arg2, arg3, arg4)
}

// RecordFileSystemMetrics mocks base method.
func (m *MockMetricsRecorder) RecordFileSystemMetrics(arg0 context.Context, arg1 interface{}, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10 float32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecordFileSystemMetrics", arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecordFileSystemMetrics indicates an expected call of RecordFileSystemMetrics.
func (mr *MockMetricsRecorderMockRecorder) RecordFileSystemMetrics(arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordFileSystemMetrics", reflect.TypeOf((*MockMetricsRecorder)(nil).RecordFileSystemMetrics), arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10)
}

// RecordSpaceMetrics mocks base method.
func (m *MockMetricsRecorder) RecordSpaceMetrics(arg0 context.Context, arg1 interface{}, arg2, arg3 int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecordSpaceMetrics", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecordSpaceMetrics indicates an expected call of RecordSpaceMetrics.
func (mr *MockMetricsRecorderMockRecorder) RecordSpaceMetrics(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordSpaceMetrics", reflect.TypeOf((*MockMetricsRecorder)(nil).RecordSpaceMetrics), arg0, arg1, arg2, arg3)
}

// RecordStorageClassSpaceMetrics mocks base method.
func (m *MockMetricsRecorder) RecordStorageClassSpaceMetrics(arg0 context.Context, arg1, arg2 string, arg3, arg4 int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecordStorageClassSpaceMetrics", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecordStorageClassSpaceMetrics indicates an expected call of RecordStorageClassSpaceMetrics.
func (mr *MockMetricsRecorderMockRecorder) RecordStorageClassSpaceMetrics(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordStorageClassSpaceMetrics", reflect.TypeOf((*MockMetricsRecorder)(nil).RecordStorageClassSpaceMetrics), arg0, arg1, arg2, arg3, arg4)
}

// MockFloat64UpDownCounterCreater is a mock of Float64UpDownCounterCreater interface.
type MockFloat64UpDownCounterCreater struct {
	ctrl     *gomock.Controller
	recorder *MockFloat64UpDownCounterCreaterMockRecorder
}

// MockFloat64UpDownCounterCreaterMockRecorder is the mock recorder for MockFloat64UpDownCounterCreater.
type MockFloat64UpDownCounterCreaterMockRecorder struct {
	mock *MockFloat64UpDownCounterCreater
}

// NewMockFloat64UpDownCounterCreater creates a new mock instance.
func NewMockFloat64UpDownCounterCreater(ctrl *gomock.Controller) *MockFloat64UpDownCounterCreater {
	mock := &MockFloat64UpDownCounterCreater{ctrl: ctrl}
	mock.recorder = &MockFloat64UpDownCounterCreaterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFloat64UpDownCounterCreater) EXPECT() *MockFloat64UpDownCounterCreaterMockRecorder {
	return m.recorder
}

// AsyncFloat64 mocks base method.
func (m *MockFloat64UpDownCounterCreater) AsyncFloat64() asyncfloat64.InstrumentProvider {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AsyncFloat64")
	ret0, _ := ret[0].(asyncfloat64.InstrumentProvider)
	return ret0
}

// AsyncFloat64 indicates an expected call of AsyncFloat64.
func (mr *MockFloat64UpDownCounterCreaterMockRecorder) AsyncFloat64() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AsyncFloat64", reflect.TypeOf((*MockFloat64UpDownCounterCreater)(nil).AsyncFloat64))
}
