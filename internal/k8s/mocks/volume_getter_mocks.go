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
// Source: github.com/dell/csm-metrics-powerstore/internal/k8s (interfaces: VolumeGetter)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1 "k8s.io/api/core/v1"
)

// MockVolumeGetter is a mock of VolumeGetter interface.
type MockVolumeGetter struct {
	ctrl     *gomock.Controller
	recorder *MockVolumeGetterMockRecorder
}

// MockVolumeGetterMockRecorder is the mock recorder for MockVolumeGetter.
type MockVolumeGetterMockRecorder struct {
	mock *MockVolumeGetter
}

// NewMockVolumeGetter creates a new mock instance.
func NewMockVolumeGetter(ctrl *gomock.Controller) *MockVolumeGetter {
	mock := &MockVolumeGetter{ctrl: ctrl}
	mock.recorder = &MockVolumeGetterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockVolumeGetter) EXPECT() *MockVolumeGetterMockRecorder {
	return m.recorder
}

// GetPersistentVolumes mocks base method.
func (m *MockVolumeGetter) GetPersistentVolumes() (*v1.PersistentVolumeList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPersistentVolumes")
	ret0, _ := ret[0].(*v1.PersistentVolumeList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPersistentVolumes indicates an expected call of GetPersistentVolumes.
func (mr *MockVolumeGetterMockRecorder) GetPersistentVolumes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPersistentVolumes", reflect.TypeOf((*MockVolumeGetter)(nil).GetPersistentVolumes))
}
