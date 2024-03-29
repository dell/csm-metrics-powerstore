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
// Source: github.com/dell/csm-metrics-powerstore/internal/k8s (interfaces: LeaderElectorGetter)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockLeaderElectorGetter is a mock of LeaderElectorGetter interface.
type MockLeaderElectorGetter struct {
	ctrl     *gomock.Controller
	recorder *MockLeaderElectorGetterMockRecorder
}

// MockLeaderElectorGetterMockRecorder is the mock recorder for MockLeaderElectorGetter.
type MockLeaderElectorGetterMockRecorder struct {
	mock *MockLeaderElectorGetter
}

// NewMockLeaderElectorGetter creates a new mock instance.
func NewMockLeaderElectorGetter(ctrl *gomock.Controller) *MockLeaderElectorGetter {
	mock := &MockLeaderElectorGetter{ctrl: ctrl}
	mock.recorder = &MockLeaderElectorGetterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLeaderElectorGetter) EXPECT() *MockLeaderElectorGetterMockRecorder {
	return m.recorder
}

// InitLeaderElection mocks base method.
func (m *MockLeaderElectorGetter) InitLeaderElection(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InitLeaderElection", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// InitLeaderElection indicates an expected call of InitLeaderElection.
func (mr *MockLeaderElectorGetterMockRecorder) InitLeaderElection(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitLeaderElection", reflect.TypeOf((*MockLeaderElectorGetter)(nil).InitLeaderElection), arg0, arg1)
}

// IsLeader mocks base method.
func (m *MockLeaderElectorGetter) IsLeader() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsLeader")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsLeader indicates an expected call of IsLeader.
func (mr *MockLeaderElectorGetterMockRecorder) IsLeader() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsLeader", reflect.TypeOf((*MockLeaderElectorGetter)(nil).IsLeader))
}
