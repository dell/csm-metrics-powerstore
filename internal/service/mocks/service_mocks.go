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
// Source: github.com/dell/csm-metrics-powerstore/internal/service (interfaces: Service)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockService is a mock of Service interface.
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
}

// MockServiceMockRecorder is the mock recorder for MockService.
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance.
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// ExportArraySpaceMetrics mocks base method.
func (m *MockService) ExportArraySpaceMetrics(arg0 context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ExportArraySpaceMetrics", arg0)
}

// ExportArraySpaceMetrics indicates an expected call of ExportArraySpaceMetrics.
func (mr *MockServiceMockRecorder) ExportArraySpaceMetrics(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExportArraySpaceMetrics", reflect.TypeOf((*MockService)(nil).ExportArraySpaceMetrics), arg0)
}

// ExportFileSystemStatistics mocks base method.
func (m *MockService) ExportFileSystemStatistics(arg0 context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ExportFileSystemStatistics", arg0)
}

// ExportFileSystemStatistics indicates an expected call of ExportFileSystemStatistics.
func (mr *MockServiceMockRecorder) ExportFileSystemStatistics(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExportFileSystemStatistics", reflect.TypeOf((*MockService)(nil).ExportFileSystemStatistics), arg0)
}

// ExportSpaceVolumeMetrics mocks base method.
func (m *MockService) ExportSpaceVolumeMetrics(arg0 context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ExportSpaceVolumeMetrics", arg0)
}

// ExportSpaceVolumeMetrics indicates an expected call of ExportSpaceVolumeMetrics.
func (mr *MockServiceMockRecorder) ExportSpaceVolumeMetrics(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExportSpaceVolumeMetrics", reflect.TypeOf((*MockService)(nil).ExportSpaceVolumeMetrics), arg0)
}

// ExportVolumeStatistics mocks base method.
func (m *MockService) ExportVolumeStatistics(arg0 context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ExportVolumeStatistics", arg0)
}

// ExportVolumeStatistics indicates an expected call of ExportVolumeStatistics.
func (mr *MockServiceMockRecorder) ExportVolumeStatistics(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExportVolumeStatistics", reflect.TypeOf((*MockService)(nil).ExportVolumeStatistics), arg0)
}
