// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/dell/csm-metrics-powerstore/internal/service (interfaces: MetricsRecorder,MeterCreater)
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	service "github.com/dell/csm-metrics-powerstore/internal/service"
	gomock "github.com/golang/mock/gomock"
	metric "go.opentelemetry.io/otel/metric"
)

// MockMetricsRecorder is a mock of MetricsRecorder interface.
type MockMetricsRecorder struct {
	ctrl     *gomock.Controller
	recorder *MockMetricsRecorderMockRecorder
	isgomock struct{}
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
func (m *MockMetricsRecorder) Record(ctx context.Context, meta any, readBW, writeBW, readIOPS, writeIOPS, readLatency, writeLatency, syncronizationBW, mirrorBW, dataRemaining float32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Record", ctx, meta, readBW, writeBW, readIOPS, writeIOPS, readLatency, writeLatency, syncronizationBW, mirrorBW, dataRemaining)
	ret0, _ := ret[0].(error)
	return ret0
}

// Record indicates an expected call of Record.
func (mr *MockMetricsRecorderMockRecorder) Record(ctx, meta, readBW, writeBW, readIOPS, writeIOPS, readLatency, writeLatency, syncronizationBW, mirrorBW, dataRemaining any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Record", reflect.TypeOf((*MockMetricsRecorder)(nil).Record), ctx, meta, readBW, writeBW, readIOPS, writeIOPS, readLatency, writeLatency, syncronizationBW, mirrorBW, dataRemaining)
}

// RecordArraySpaceMetrics mocks base method.
func (m *MockMetricsRecorder) RecordArraySpaceMetrics(ctx context.Context, arrayID, driver string, logicalProvisioned, logicalUsed int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecordArraySpaceMetrics", ctx, arrayID, driver, logicalProvisioned, logicalUsed)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecordArraySpaceMetrics indicates an expected call of RecordArraySpaceMetrics.
func (mr *MockMetricsRecorderMockRecorder) RecordArraySpaceMetrics(ctx, arrayID, driver, logicalProvisioned, logicalUsed any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordArraySpaceMetrics", reflect.TypeOf((*MockMetricsRecorder)(nil).RecordArraySpaceMetrics), ctx, arrayID, driver, logicalProvisioned, logicalUsed)
}

// RecordFileSystemMetrics mocks base method.
func (m *MockMetricsRecorder) RecordFileSystemMetrics(ctx context.Context, meta any, readBW, writeBW, readIOPS, writeIOPS, readLatency, writeLatency, syncronizationBW, mirrorBW, dataRemaining float32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecordFileSystemMetrics", ctx, meta, readBW, writeBW, readIOPS, writeIOPS, readLatency, writeLatency, syncronizationBW, mirrorBW, dataRemaining)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecordFileSystemMetrics indicates an expected call of RecordFileSystemMetrics.
func (mr *MockMetricsRecorderMockRecorder) RecordFileSystemMetrics(ctx, meta, readBW, writeBW, readIOPS, writeIOPS, readLatency, writeLatency, syncronizationBW, mirrorBW, dataRemaining any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordFileSystemMetrics", reflect.TypeOf((*MockMetricsRecorder)(nil).RecordFileSystemMetrics), ctx, meta, readBW, writeBW, readIOPS, writeIOPS, readLatency, writeLatency, syncronizationBW, mirrorBW, dataRemaining)
}

// RecordSpaceMetrics mocks base method.
func (m *MockMetricsRecorder) RecordSpaceMetrics(ctx context.Context, meta any, logicalProvisioned, logicalUsed int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecordSpaceMetrics", ctx, meta, logicalProvisioned, logicalUsed)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecordSpaceMetrics indicates an expected call of RecordSpaceMetrics.
func (mr *MockMetricsRecorderMockRecorder) RecordSpaceMetrics(ctx, meta, logicalProvisioned, logicalUsed any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordSpaceMetrics", reflect.TypeOf((*MockMetricsRecorder)(nil).RecordSpaceMetrics), ctx, meta, logicalProvisioned, logicalUsed)
}

// RecordStorageClassSpaceMetrics mocks base method.
func (m *MockMetricsRecorder) RecordStorageClassSpaceMetrics(ctx context.Context, storageclass, driver string, logicalProvisioned, logicalUsed int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecordStorageClassSpaceMetrics", ctx, storageclass, driver, logicalProvisioned, logicalUsed)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecordStorageClassSpaceMetrics indicates an expected call of RecordStorageClassSpaceMetrics.
func (mr *MockMetricsRecorderMockRecorder) RecordStorageClassSpaceMetrics(ctx, storageclass, driver, logicalProvisioned, logicalUsed any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordStorageClassSpaceMetrics", reflect.TypeOf((*MockMetricsRecorder)(nil).RecordStorageClassSpaceMetrics), ctx, storageclass, driver, logicalProvisioned, logicalUsed)
}

// RecordTopologyMetrics mocks base method.
func (m *MockMetricsRecorder) RecordTopologyMetrics(ctx context.Context, meta any, topologyMetrics *service.TopologyMetricsRecord) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecordTopologyMetrics", ctx, meta, topologyMetrics)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecordTopologyMetrics indicates an expected call of RecordTopologyMetrics.
func (mr *MockMetricsRecorderMockRecorder) RecordTopologyMetrics(ctx, meta, topologyMetrics any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordTopologyMetrics", reflect.TypeOf((*MockMetricsRecorder)(nil).RecordTopologyMetrics), ctx, meta, topologyMetrics)
}

// MockMeterCreater is a mock of MeterCreater interface.
type MockMeterCreater struct {
	ctrl     *gomock.Controller
	recorder *MockMeterCreaterMockRecorder
	isgomock struct{}
}

// MockMeterCreaterMockRecorder is the mock recorder for MockMeterCreater.
type MockMeterCreaterMockRecorder struct {
	mock *MockMeterCreater
}

// NewMockMeterCreater creates a new mock instance.
func NewMockMeterCreater(ctrl *gomock.Controller) *MockMeterCreater {
	mock := &MockMeterCreater{ctrl: ctrl}
	mock.recorder = &MockMeterCreaterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMeterCreater) EXPECT() *MockMeterCreaterMockRecorder {
	return m.recorder
}

// MeterProvider mocks base method.
func (m *MockMeterCreater) MeterProvider() metric.Meter {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MeterProvider")
	ret0, _ := ret[0].(metric.Meter)
	return ret0
}

// MeterProvider indicates an expected call of MeterProvider.
func (mr *MockMeterCreaterMockRecorder) MeterProvider() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MeterProvider", reflect.TypeOf((*MockMeterCreater)(nil).MeterProvider))
}
