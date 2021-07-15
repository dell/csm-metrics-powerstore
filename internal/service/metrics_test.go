// Copyright (c) 2021 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dell/csm-metrics-powerstore/internal/service"
	"github.com/dell/csm-metrics-powerstore/internal/service/mocks"

	"github.com/golang/mock/gomock"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/metric"
)

func Test_Metrics_Record(t *testing.T) {
	type checkFn func(*testing.T, error)
	checkFns := func(checkFns ...checkFn) []checkFn { return checkFns }

	verifyError := func(t *testing.T, err error) {
		if err == nil {
			t.Errorf("expected an error, got nil")
		}
	}

	verifyNoError := func(t *testing.T, err error) {
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	}

	metas := []interface{}{
		&service.VolumeMeta{
			ID: "123",
		},
	}

	tests := map[string]func(t *testing.T) ([]*service.MetricsWrapper, []checkFn){
		"success": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)

			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
				otMeter := global.Meter(prefix + "_test")
				readBW, err := otMeter.NewFloat64UpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.NewFloat64UpDownCounter(prefix + "write_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readIOPS, err := otMeter.NewFloat64UpDownCounter(prefix + "read_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeIOPS, err := otMeter.NewFloat64UpDownCounter(prefix + "write_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readLatency, err := otMeter.NewFloat64UpDownCounter(prefix + "read_latency_milliseconds")
				if err != nil {
					t.Fatal(err)
				}

				writeLatency, err := otMeter.NewFloat64UpDownCounter(prefix + "write_latency_milliseconds")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readBW, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(writeBW, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readIOPS, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(writeIOPS, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readLatency, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(writeLatency, nil)

				return &service.MetricsWrapper{
					Meter: meter,
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_volume_"),
			}

			return mws, checkFns(verifyNoError)
		},
		"error creating read_bw_megabytes_per_second": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)

			meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(metric.Float64UpDownCounter{}, errors.New("error")).Times(2)

			mws := []*service.MetricsWrapper{{Meter: meter}}

			return mws, checkFns(verifyError)
		},
		"error creating write_bw_megabytes_per_second": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
				otMeter := global.Meter(prefix + "_test")
				readBW, err := otMeter.NewFloat64UpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readBW, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(metric.Float64UpDownCounter{}, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter,
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_volume_"),
			}

			return mws, checkFns(verifyError)
		},
		"error creating read_iops_per_second": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
				otMeter := global.Meter(prefix + "_test")

				readBW, err := otMeter.NewFloat64UpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.NewFloat64UpDownCounter(prefix + "write_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readBW, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(writeBW, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(metric.Float64UpDownCounter{}, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter,
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_volume_"),
			}

			return mws, checkFns(verifyError)
		},
		"error creating write_iops_per_second": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
				otMeter := global.Meter(prefix + "_test")

				readBW, err := otMeter.NewFloat64UpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.NewFloat64UpDownCounter(prefix + "write_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readIOPS, err := otMeter.NewFloat64UpDownCounter(prefix + "read_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readBW, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(writeBW, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readIOPS, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(metric.Float64UpDownCounter{}, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter,
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_volume_"),
			}

			return mws, checkFns(verifyError)
		},
		"error creating read_latency_milliseconds": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
				otMeter := global.Meter(prefix + "_test")

				readBW, err := otMeter.NewFloat64UpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.NewFloat64UpDownCounter(prefix + "write_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readIOPS, err := otMeter.NewFloat64UpDownCounter(prefix + "read_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeIOPS, err := otMeter.NewFloat64UpDownCounter(prefix + "write_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readBW, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(writeBW, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readIOPS, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(writeIOPS, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(metric.Float64UpDownCounter{}, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter,
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_volume_"),
			}

			return mws, checkFns(verifyError)
		},
		"error creating write_latency_milliseconds": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
				otMeter := global.Meter(prefix + "_test")

				readBW, err := otMeter.NewFloat64UpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.NewFloat64UpDownCounter(prefix + "write_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readIOPS, err := otMeter.NewFloat64UpDownCounter(prefix + "read_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeIOPS, err := otMeter.NewFloat64UpDownCounter(prefix + "write_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readLatency, err := otMeter.NewFloat64UpDownCounter(prefix + "read_latency_milliseconds")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readBW, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(writeBW, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readIOPS, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(writeIOPS, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readLatency, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(metric.Float64UpDownCounter{}, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter,
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_volume_"),
			}

			return mws, checkFns(verifyError)
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mws, checks := tc(t)
			for i := range mws {
				err := mws[i].Record(context.Background(), metas[i], 1, 2, 3, 4, 5, 6)
				for _, check := range checks {
					check(t, err)
				}
			}
		})
	}
}

func Test__SpaceMetrics_Record(t *testing.T) {
	type checkFn func(*testing.T, error)
	checkFns := func(checkFns ...checkFn) []checkFn { return checkFns }

	verifyError := func(t *testing.T, err error) {
		if err == nil {
			t.Errorf("expected an error, got nil")
		}
	}

	verifyNoError := func(t *testing.T, err error) {
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	}

	metas := []interface{}{
		&service.VolumeMeta{
			ID: "123",
		},
	}

	tests := map[string]func(t *testing.T) ([]*service.MetricsWrapper, []checkFn){
		"success": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)

			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
				otMeter := global.Meter(prefix + "_test")
				readBW, err := otMeter.NewFloat64UpDownCounter(prefix + "logical_provisioned_megabytes")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.NewFloat64UpDownCounter(prefix + "logical_used_megabytes")
				if err != nil {
					t.Fatal(err)
				}

				readIOPS, err := otMeter.NewFloat64UpDownCounter(prefix + "thin_savings_megabytes")
				if err != nil {
					t.Fatal(err)
				}

				writeIOPS, err := otMeter.NewFloat64UpDownCounter(prefix + "max_thin_savings_megabytes")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readBW, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(writeBW, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readIOPS, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(writeIOPS, nil)

				return &service.MetricsWrapper{
					Meter: meter,
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_volume_"),
			}

			return mws, checkFns(verifyNoError)
		},
		"error creating logical_provisioned_megabytes": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)

			meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(metric.Float64UpDownCounter{}, errors.New("error")).Times(2)

			mws := []*service.MetricsWrapper{{Meter: meter}}

			return mws, checkFns(verifyError)
		},
		"error creating logical_used_megabytes": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
				otMeter := global.Meter(prefix + "_test")
				readBW, err := otMeter.NewFloat64UpDownCounter(prefix + "logical_provisioned_megabytes")
				if err != nil {
					t.Fatal(err)
				}
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readBW, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(metric.Float64UpDownCounter{}, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter,
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_volume_"),
			}

			return mws, checkFns(verifyError)
		},
		"error creating thin_savings_megabytes": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
				otMeter := global.Meter(prefix + "_test")

				readBW, err := otMeter.NewFloat64UpDownCounter(prefix + "logical_provisioned_megabytes")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.NewFloat64UpDownCounter(prefix + "logical_used_megabytes")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readBW, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(writeBW, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(metric.Float64UpDownCounter{}, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter,
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_volume_"),
			}

			return mws, checkFns(verifyError)
		},
		"error creating max_thin_savings_megabytes": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
				otMeter := global.Meter(prefix + "_test")

				readBW, err := otMeter.NewFloat64UpDownCounter(prefix + "logical_provisioned_megabytes")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.NewFloat64UpDownCounter(prefix + "logical_used_megabytes")
				if err != nil {
					t.Fatal(err)
				}

				readIOPS, err := otMeter.NewFloat64UpDownCounter(prefix + "thin_savings_megabytes")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readBW, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(writeBW, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readIOPS, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(metric.Float64UpDownCounter{}, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter,
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_volume_"),
			}

			return mws, checkFns(verifyError)
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mws, checks := tc(t)
			for i := range mws {
				err := mws[i].RecordSpaceMetrics(context.Background(), metas[i], 1, 2, 3, 4)
				for _, check := range checks {
					check(t, err)
				}
			}
		})
	}
}

func Test_ArraySpace_Metrics_Record(t *testing.T) {
	type checkFn func(*testing.T, error)
	checkFns := func(checkFns ...checkFn) []checkFn { return checkFns }

	verifyError := func(t *testing.T, err error) {
		if err == nil {
			t.Errorf("expected an error, got nil")
		}
	}

	verifyNoError := func(t *testing.T, err error) {
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	}

	tests := map[string]func(t *testing.T) ([]*service.MetricsWrapper, []checkFn){
		"success": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)

			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
				otMeter := global.Meter(prefix + "_test")
				provisioned, err := otMeter.NewFloat64UpDownCounter(prefix + "logical_provisioned_megabytes")
				if err != nil {
					t.Fatal(err)
				}

				used, err := otMeter.NewFloat64UpDownCounter(prefix + "logical_used_megabytes")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(provisioned, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(used, nil)

				return &service.MetricsWrapper{
					Meter: meter,
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_array_"),
			}

			return mws, checkFns(verifyNoError)
		},
		"error creating logical_provisioned_megabytes": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)

			meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(metric.Float64UpDownCounter{}, errors.New("error")).Times(2)

			mws := []*service.MetricsWrapper{{Meter: meter}}

			return mws, checkFns(verifyError)
		},
		"error creating logical_used_megabytes": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
				otMeter := global.Meter(prefix + "_test")
				readBW, err := otMeter.NewFloat64UpDownCounter(prefix + "logical_provisioned_megabytes")
				if err != nil {
					t.Fatal(err)
				}
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readBW, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(metric.Float64UpDownCounter{}, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter,
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_array_"),
			}

			return mws, checkFns(verifyError)
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mws, checks := tc(t)
			for i := range mws {
				err := mws[i].RecordArraySpaceMetrics(context.Background(), "123", 1, 2)
				for _, check := range checks {
					check(t, err)
				}
			}
		})
	}
}

func Test_StorageClassSpace_Metrics_Record(t *testing.T) {
	type checkFn func(*testing.T, error)
	checkFns := func(checkFns ...checkFn) []checkFn { return checkFns }

	verifyError := func(t *testing.T, err error) {
		if err == nil {
			t.Errorf("expected an error, got nil")
		}
	}

	verifyNoError := func(t *testing.T, err error) {
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	}

	tests := map[string]func(t *testing.T) ([]*service.MetricsWrapper, []checkFn){
		"success": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)

			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
				otMeter := global.Meter(prefix + "_test")
				provisioned, err := otMeter.NewFloat64UpDownCounter(prefix + "logical_provisioned_megabytes")
				if err != nil {
					t.Fatal(err)
				}

				used, err := otMeter.NewFloat64UpDownCounter(prefix + "logical_used_megabytes")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(provisioned, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(used, nil)

				return &service.MetricsWrapper{
					Meter: meter,
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_storage_class_"),
			}

			return mws, checkFns(verifyNoError)
		},
		"error creating logical_provisioned_megabytes": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)

			meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(metric.Float64UpDownCounter{}, errors.New("error")).Times(2)

			mws := []*service.MetricsWrapper{{Meter: meter}}

			return mws, checkFns(verifyError)
		},
		"error creating logical_used_megabytes": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
				otMeter := global.Meter(prefix + "_test")
				readBW, err := otMeter.NewFloat64UpDownCounter(prefix + "logical_provisioned_megabytes")
				if err != nil {
					t.Fatal(err)
				}
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readBW, nil)
				meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(metric.Float64UpDownCounter{}, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter,
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_storage_class_"),
			}

			return mws, checkFns(verifyError)
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mws, checks := tc(t)
			for i := range mws {
				err := mws[i].RecordStorageClassSpaceMetrics(context.Background(), "123", 1, 2)
				for _, check := range checks {
					check(t, err)
				}
			}
		})
	}
}
func Test_Volume_Metrics_Label_Update(t *testing.T) {
	metaFirst := &service.VolumeMeta{
		ID: "123",
	}

	metaSecond := &service.VolumeMeta{
		ID: "123",
	}

	expectedLables := []kv.KeyValue{
		kv.String("VolumeID", metaSecond.ID),
		kv.String("PlotWithMean", "No"),
	}

	ctrl := gomock.NewController(t)

	meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
	otMeter := global.Meter("powerstore_volume_test")
	readBW, err := otMeter.NewFloat64UpDownCounter("powerstore_volume_read_bw_megabytes_per_second")
	if err != nil {
		t.Fatal(err)
	}

	writeBW, err := otMeter.NewFloat64UpDownCounter("powerstore_volume_write_bw_megabytes_per_second")
	if err != nil {
		t.Fatal(err)
	}

	readIOPS, err := otMeter.NewFloat64UpDownCounter("powerstore_volume_read_iops_per_second")
	if err != nil {
		t.Fatal(err)
	}

	writeIOPS, err := otMeter.NewFloat64UpDownCounter("powerstore_volume_write_iops_per_second")
	if err != nil {
		t.Fatal(err)
	}

	readLatency, err := otMeter.NewFloat64UpDownCounter("powerstore_volume_read_latency_milliseconds")
	if err != nil {
		t.Fatal(err)
	}

	writeLatency, err := otMeter.NewFloat64UpDownCounter("powerstore_volume_write_latency_milliseconds")
	if err != nil {
		t.Fatal(err)
	}

	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readBW, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(writeBW, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readIOPS, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(writeIOPS, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readLatency, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(writeLatency, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readBW, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(writeBW, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readIOPS, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(writeIOPS, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(readLatency, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(writeLatency, nil)

	mw := &service.MetricsWrapper{
		Meter: meter,
	}

	t.Run("success: volume metric labels updated", func(t *testing.T) {
		err := mw.Record(context.Background(), metaFirst, 1, 2, 3, 4, 5, 6)
		if err != nil {
			t.Errorf("expected nil error (record #1), got %v", err)
		}
		err = mw.Record(context.Background(), metaSecond, 1, 2, 3, 4, 5, 6)
		if err != nil {
			t.Errorf("expected nil error (record #2), got %v", err)
		}

		newLabels, ok := mw.Labels.Load(metaFirst.ID)
		if !ok {
			t.Errorf("expected labels to exist for %v, but did not find them", metaFirst.ID)
		}
		labels := newLabels.([]kv.KeyValue)
		for _, l := range labels {
			for _, e := range expectedLables {
				if l.Key == e.Key {
					if l.Value.AsString() != e.Value.AsString() {
						t.Errorf("expected label %v to be updated to %v, but the value was %v", e.Key, e.Value.AsString(), l.Value.AsString())
					}
				}
			}
		}
	})
}

func Test_Space_Metrics_Label_Update(t *testing.T) {
	metaFirst := &service.VolumeMeta{
		ID:           "123",
		ArrayID:      "arr123",
		StorageClass: "powerstore",
	}

	metaSecond := &service.VolumeMeta{
		ID:           "123",
		ArrayID:      "arr123",
		StorageClass: "powerstore",
	}

	expectedLables := []kv.KeyValue{
		kv.String("VolumeID", metaSecond.ID),
		kv.String("PlotWithMean", "No"),
	}

	ctrl := gomock.NewController(t)

	meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
	otMeter := global.Meter("powerstore_volume_test")
	provisioned, err := otMeter.NewFloat64UpDownCounter("logical_provisioned_megabytes")
	if err != nil {
		t.Fatal(err)
	}

	used, err := otMeter.NewFloat64UpDownCounter("logical_used_megabytes")
	if err != nil {
		t.Fatal(err)
	}

	thin, err := otMeter.NewFloat64UpDownCounter("thin_savings_megabytes")
	if err != nil {
		t.Fatal(err)
	}

	maxthin, err := otMeter.NewFloat64UpDownCounter("max_thin_savings_megabytes")
	if err != nil {
		t.Fatal(err)
	}

	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(provisioned, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(used, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(thin, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(maxthin, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(provisioned, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(used, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(thin, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(maxthin, nil)

	mw := &service.MetricsWrapper{
		Meter: meter,
	}

	t.Run("success: volume metric labels updated", func(t *testing.T) {
		err := mw.RecordSpaceMetrics(context.Background(), metaFirst, 1, 2, 3, 4)
		if err != nil {
			t.Errorf("expected nil error (record #1), got %v", err)
		}
		err = mw.RecordSpaceMetrics(context.Background(), metaSecond, 1, 2, 3, 4)
		if err != nil {
			t.Errorf("expected nil error (record #2), got %v", err)
		}

		newLabels, ok := mw.Labels.Load(metaFirst.ID)
		if !ok {
			t.Errorf("expected labels to exist for %v, but did not find them", metaFirst.ID)
		}
		labels := newLabels.([]kv.KeyValue)
		for _, l := range labels {
			for _, e := range expectedLables {
				if l.Key == e.Key {
					if l.Value.AsString() != e.Value.AsString() {
						t.Errorf("expected label %v to be updated to %v, but the value was %v", e.Key, e.Value.AsString(), l.Value.AsString())
					}
				}
			}
		}
	})
}

func Test_ArraySpace_Metrics_Label_Update(t *testing.T) {

	array1 := "123"
	array2 := "123"

	expectedLables := []kv.KeyValue{
		kv.String("ArrayID", array1),
		kv.String("PlotWithMean", "No"),
	}

	ctrl := gomock.NewController(t)

	meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
	otMeter := global.Meter("powerstore_array_")
	provisioned, err := otMeter.NewFloat64UpDownCounter("logical_provisioned_megabytes")
	if err != nil {
		t.Fatal(err)
	}

	used, err := otMeter.NewFloat64UpDownCounter("logical_used_megabytes")
	if err != nil {
		t.Fatal(err)
	}

	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(provisioned, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(used, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(provisioned, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(used, nil)

	mw := &service.MetricsWrapper{
		Meter: meter,
	}

	t.Run("success: volume metric labels updated", func(t *testing.T) {
		err := mw.RecordArraySpaceMetrics(context.Background(), array1, 1, 2)
		if err != nil {
			t.Errorf("expected nil error (record #1), got %v", err)
		}
		err = mw.RecordArraySpaceMetrics(context.Background(), array2, 1, 2)
		if err != nil {
			t.Errorf("expected nil error (record #2), got %v", err)
		}

		newLabels, ok := mw.Labels.Load(array1)
		if !ok {
			t.Errorf("expected labels to exist for %v, but did not find them", array1)
		}
		labels := newLabels.([]kv.KeyValue)
		for _, l := range labels {
			for _, e := range expectedLables {
				if l.Key == e.Key {
					if l.Value.AsString() != e.Value.AsString() {
						t.Errorf("expected label %v to be updated to %v, but the value was %v", e.Key, e.Value.AsString(), l.Value.AsString())
					}
				}
			}
		}
	})
}

func Test_StrorageClass_Space_Metrics_Label_Update(t *testing.T) {

	array1 := "123"
	array2 := "123"

	expectedLables := []kv.KeyValue{
		kv.String("StorageClass", array1),
		kv.String("PlotWithMean", "No"),
	}

	ctrl := gomock.NewController(t)

	meter := mocks.NewMockFloat64UpDownCounterCreater(ctrl)
	otMeter := global.Meter("powerstore_storage_class_")
	provisioned, err := otMeter.NewFloat64UpDownCounter("logical_provisioned_megabytes")
	if err != nil {
		t.Fatal(err)
	}

	used, err := otMeter.NewFloat64UpDownCounter("logical_used_megabytes")
	if err != nil {
		t.Fatal(err)
	}

	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(provisioned, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(used, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(provisioned, nil)
	meter.EXPECT().NewFloat64UpDownCounter(gomock.Any()).Return(used, nil)

	mw := &service.MetricsWrapper{
		Meter: meter,
	}

	t.Run("success: volume metric labels updated", func(t *testing.T) {
		err := mw.RecordStorageClassSpaceMetrics(context.Background(), array1, 1, 2)
		if err != nil {
			t.Errorf("expected nil error (record #1), got %v", err)
		}
		err = mw.RecordStorageClassSpaceMetrics(context.Background(), array2, 1, 2)
		if err != nil {
			t.Errorf("expected nil error (record #2), got %v", err)
		}

		newLabels, ok := mw.Labels.Load(array1)
		if !ok {
			t.Errorf("expected labels to exist for %v, but did not find them", array1)
		}
		labels := newLabels.([]kv.KeyValue)
		for _, l := range labels {
			for _, e := range expectedLables {
				if l.Key == e.Key {
					if l.Value.AsString() != e.Value.AsString() {
						t.Errorf("expected label %v to be updated to %v, but the value was %v", e.Key, e.Value.AsString(), l.Value.AsString())
					}
				}
			}
		}
	})
}
