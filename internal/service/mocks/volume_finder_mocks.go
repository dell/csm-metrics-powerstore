// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/dell/csm-metrics-powerstore/internal/service (interfaces: VolumeFinder)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	k8s "github.com/dell/csm-metrics-powerstore/internal/k8s"
	gomock "github.com/golang/mock/gomock"
)

// MockVolumeFinder is a mock of VolumeFinder interface.
type MockVolumeFinder struct {
	ctrl     *gomock.Controller
	recorder *MockVolumeFinderMockRecorder
}

// MockVolumeFinderMockRecorder is the mock recorder for MockVolumeFinder.
type MockVolumeFinderMockRecorder struct {
	mock *MockVolumeFinder
}

// NewMockVolumeFinder creates a new mock instance.
func NewMockVolumeFinder(ctrl *gomock.Controller) *MockVolumeFinder {
	mock := &MockVolumeFinder{ctrl: ctrl}
	mock.recorder = &MockVolumeFinderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockVolumeFinder) EXPECT() *MockVolumeFinderMockRecorder {
	return m.recorder
}

// GetPersistentVolumes mocks base method.
func (m *MockVolumeFinder) GetPersistentVolumes(arg0 context.Context) ([]k8s.VolumeInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPersistentVolumes", arg0)
	ret0, _ := ret[0].([]k8s.VolumeInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPersistentVolumes indicates an expected call of GetPersistentVolumes.
func (mr *MockVolumeFinderMockRecorder) GetPersistentVolumes(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPersistentVolumes", reflect.TypeOf((*MockVolumeFinder)(nil).GetPersistentVolumes), arg0)
}
