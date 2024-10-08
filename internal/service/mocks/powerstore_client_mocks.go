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
// Source: github.com/dell/csm-metrics-powerstore/internal/service (interfaces: PowerStoreClient)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gopowerstore "github.com/dell/gopowerstore"
	gomock "github.com/golang/mock/gomock"
)

// MockPowerStoreClient is a mock of PowerStoreClient interface.
type MockPowerStoreClient struct {
	ctrl     *gomock.Controller
	recorder *MockPowerStoreClientMockRecorder
}

// MockPowerStoreClientMockRecorder is the mock recorder for MockPowerStoreClient.
type MockPowerStoreClientMockRecorder struct {
	mock *MockPowerStoreClient
}

// NewMockPowerStoreClient creates a new mock instance.
func NewMockPowerStoreClient(ctrl *gomock.Controller) *MockPowerStoreClient {
	mock := &MockPowerStoreClient{ctrl: ctrl}
	mock.recorder = &MockPowerStoreClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPowerStoreClient) EXPECT() *MockPowerStoreClientMockRecorder {
	return m.recorder
}

// GetFS mocks base method.
func (m *MockPowerStoreClient) GetFS(arg0 context.Context, arg1 string) (gopowerstore.FileSystem, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFS", arg0, arg1)
	ret0, _ := ret[0].(gopowerstore.FileSystem)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFS indicates an expected call of GetFS.
func (mr *MockPowerStoreClientMockRecorder) GetFS(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFS", reflect.TypeOf((*MockPowerStoreClient)(nil).GetFS), arg0, arg1)
}

// PerformanceMetricsByFileSystem mocks base method.
func (m *MockPowerStoreClient) PerformanceMetricsByFileSystem(arg0 context.Context, arg1 string, arg2 gopowerstore.MetricsIntervalEnum) ([]gopowerstore.PerformanceMetricsByFileSystemResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PerformanceMetricsByFileSystem", arg0, arg1, arg2)
	ret0, _ := ret[0].([]gopowerstore.PerformanceMetricsByFileSystemResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

//VolumeMirrorTransferRate mock base method.
func (m *MockPowerStoreClient) VolumeMirrorTransferRate(arg0 context.Context, arg1 string, arg2 gopowerstore.MetricsIntervalEnum)([]gopowerstore.VolumeMirrorTransferRateResponse, error)  {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VolumeMirrorTransferRate", arg0, arg1, arg2)
	ret0, _ := ret[0].([]gopowerstore.VolumeMirrorTransferRateResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

//VolumeMirrorTransferRate indicates and expected call of VolumeMirrorTransferRate
func (mr *MockPowerStoreClientMockRecorder) VolumeMirrorTransferRate(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VolumeMirrorTransferRate", reflect.TypeOf((*MockPowerStoreClient)(nil).VolumeMirrorTransferRate), arg0, arg1, arg2)
}

// PerformanceMetricsByFileSystem indicates an expected call of PerformanceMetricsByFileSystem.
func (mr *MockPowerStoreClientMockRecorder) PerformanceMetricsByFileSystem(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PerformanceMetricsByFileSystem", reflect.TypeOf((*MockPowerStoreClient)(nil).PerformanceMetricsByFileSystem), arg0, arg1, arg2)
}

// PerformanceMetricsByVolume mocks base method.
func (m *MockPowerStoreClient) PerformanceMetricsByVolume(arg0 context.Context, arg1 string, arg2 gopowerstore.MetricsIntervalEnum) ([]gopowerstore.PerformanceMetricsByVolumeResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PerformanceMetricsByVolume", arg0, arg1, arg2)
	ret0, _ := ret[0].([]gopowerstore.PerformanceMetricsByVolumeResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PerformanceMetricsByVolume indicates an expected call of PerformanceMetricsByVolume.
func (mr *MockPowerStoreClientMockRecorder) PerformanceMetricsByVolume(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PerformanceMetricsByVolume", reflect.TypeOf((*MockPowerStoreClient)(nil).PerformanceMetricsByVolume), arg0, arg1, arg2)
}

// SpaceMetricsByVolume mocks base method.
func (m *MockPowerStoreClient) SpaceMetricsByVolume(arg0 context.Context, arg1 string, arg2 gopowerstore.MetricsIntervalEnum) ([]gopowerstore.SpaceMetricsByVolumeResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SpaceMetricsByVolume", arg0, arg1, arg2)
	ret0, _ := ret[0].([]gopowerstore.SpaceMetricsByVolumeResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SpaceMetricsByVolume indicates an expected call of SpaceMetricsByVolume.
func (mr *MockPowerStoreClientMockRecorder) SpaceMetricsByVolume(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SpaceMetricsByVolume", reflect.TypeOf((*MockPowerStoreClient)(nil).SpaceMetricsByVolume), arg0, arg1, arg2)
}
