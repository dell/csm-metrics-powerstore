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

package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dell/csm-metrics-powerstore/internal/service"
	"github.com/dell/csm-metrics-powerstore/internal/service/mocks"

	"github.com/golang/mock/gomock"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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
				meter := mocks.NewMockMeterCreater(ctrl)
				provider := mocks.NewMockMeter(ctrl)
				otMeter := otel.Meter(prefix + "_test")
				readBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "write_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readIOPS, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeIOPS, err := otMeter.Float64ObservableUpDownCounter(prefix + "write_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readLatency, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_latency_milliseconds")
				if err != nil {
					t.Fatal(err)
				}

				writeLatency, err := otMeter.Float64ObservableUpDownCounter(prefix + "write_latency_milliseconds")
				if err != nil {
					t.Fatal(err)
				}

				syncBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "syncronization_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				mirrorBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "mirror_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				dataRemaining, err := otMeter.Float64ObservableUpDownCounter(prefix + "data_remaining_bytes")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().MeterProvider().Times(9)

				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readLatency, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeLatency, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(syncBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(mirrorBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(dataRemaining, nil)

				return &service.MetricsWrapper{
					Meter: meter.MeterProvider(),
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_volume_"),
			}

			return mws, checkFns(verifyNoError)
		},
		"error creating read_bw_megabytes_per_second": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			meter := mocks.NewMockMeterCreater(ctrl)
			provider := mocks.NewMockMeter(ctrl)
			otMeter := otel.Meter("")
			empty, err := otMeter.Float64ObservableUpDownCounter("")
			if err != nil {
				t.Fatal(err)
			}
			meter.EXPECT().MeterProvider().Return(provider)
			// meter.EXPECT().MeterProvider()
			provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error")).Times(1)
			// provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error")).Times(1)

			mws := []*service.MetricsWrapper{{Meter: meter.MeterProvider()}}

			return mws, checkFns(verifyError)
		},
		"error creating write_bw_megabytes_per_second": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockMeterCreater(ctrl)
				provider := mocks.NewMockMeter(ctrl)
				otMeter := otel.Meter(prefix + "_test")
				readBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}
				empty, err := otMeter.Float64ObservableUpDownCounter("")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().MeterProvider().Times(2)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter.MeterProvider(),
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
				meter := mocks.NewMockMeterCreater(ctrl)
				otMeter := otel.Meter(prefix + "_test")
				provider := mocks.NewMockMeter(ctrl)

				readBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "write_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				empty, err := otMeter.Float64ObservableUpDownCounter("")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().MeterProvider().Times(3)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter.MeterProvider(),
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
				meter := mocks.NewMockMeterCreater(ctrl)
				otMeter := otel.Meter(prefix + "_test")
				provider := mocks.NewMockMeter(ctrl)

				readBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "write_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readIOPS, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				empty, err := otMeter.Float64ObservableUpDownCounter("")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().MeterProvider().Times(4)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter.MeterProvider(),
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
				meter := mocks.NewMockMeterCreater(ctrl)
				otMeter := otel.Meter(prefix + "_test")
				provider := mocks.NewMockMeter(ctrl)

				readBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "write_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readIOPS, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeIOPS, err := otMeter.Float64ObservableUpDownCounter(prefix + "write_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				empty, err := otMeter.Float64ObservableUpDownCounter("")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().MeterProvider().Times(5)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter.MeterProvider(),
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
				meter := mocks.NewMockMeterCreater(ctrl)
				otMeter := otel.Meter(prefix + "_test")
				provider := mocks.NewMockMeter(ctrl)

				readBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "write_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readIOPS, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeIOPS, err := otMeter.Float64ObservableUpDownCounter(prefix + "write_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readLatency, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_latency_milliseconds")
				if err != nil {
					t.Fatal(err)
				}

				empty, err := otMeter.Float64ObservableUpDownCounter("")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().MeterProvider().Times(6)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readLatency, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter.MeterProvider(),
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
				err := mws[i].Record(context.Background(), metas[i], 1, 2, 3, 4, 5, 6, 7, 8, 9)
				for _, check := range checks {
					check(t, err)
				}
			}
		})
	}
}

func Test_Metrics_Record_Meta(t *testing.T) {
	type checkFn func(*testing.T, error)
	checkFns := func(checkFns ...checkFn) []checkFn { return checkFns }

	verifyNoError := func(t *testing.T, err error) {
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	}

	metas := []interface{}{
		&service.VolumeMeta{},
	}

	tests := map[string]func(t *testing.T) ([]*service.MetricsWrapper, []checkFn){
		"success": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)

			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockMeterCreater(ctrl)
				provider := mocks.NewMockMeter(ctrl)
				otMeter := otel.Meter(prefix + "_test")
				readBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "write_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readIOPS, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeIOPS, err := otMeter.Float64ObservableUpDownCounter(prefix + "write_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readLatency, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_latency_milliseconds")
				if err != nil {
					t.Fatal(err)
				}

				writeLatency, err := otMeter.Float64ObservableUpDownCounter(prefix + "write_latency_milliseconds")
				if err != nil {
					t.Fatal(err)
				}

				syncBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "syncronization_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				mirrorBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "mirror_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				dataRemaining, err := otMeter.Float64ObservableUpDownCounter(prefix + "data_remaining_bytes")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().MeterProvider().Times(9)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readLatency, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeLatency, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(syncBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(mirrorBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(dataRemaining, nil)

				return &service.MetricsWrapper{
					Meter: meter.MeterProvider(),
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_volume_"),
			}

			return mws, checkFns(verifyNoError)
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mws, checks := tc(t)
			for i := range mws {
				err := mws[i].Record(context.Background(), metas[i], 1, 2, 3, 4, 5, 6, 7, 8, 9)
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
		&service.SpaceVolumeMeta{
			ID: "123",
		},
	}

	tests := map[string]func(t *testing.T) ([]*service.MetricsWrapper, []checkFn){
		"success": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)

			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockMeterCreater(ctrl)
				provider := mocks.NewMockMeter(ctrl)
				otMeter := otel.Meter(prefix + "_test")
				provisioned, err := otMeter.Float64ObservableUpDownCounter(prefix + "logical_provisioned_megabytes")
				if err != nil {
					t.Fatal(err)
				}

				used, err := otMeter.Float64ObservableUpDownCounter(prefix + "logical_used_megabytes")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().MeterProvider().Times(2)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(provisioned, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(used, nil)

				return &service.MetricsWrapper{
					Meter: meter.MeterProvider(),
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_volume_"),
			}

			return mws, checkFns(verifyNoError)
		},
		"error creating logical_provisioned_megabytes": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			meter := mocks.NewMockMeterCreater(ctrl)
			provider := mocks.NewMockMeter(ctrl)
			otMeter := otel.Meter("")
			empty, err := otMeter.Float64ObservableUpDownCounter("")
			if err != nil {
				t.Fatal(err)
			}
			meter.EXPECT().MeterProvider()
			provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error")).Times(1)

			mws := []*service.MetricsWrapper{{Meter: meter.MeterProvider()}}

			return mws, checkFns(verifyError)
		},
		"error creating logical_used_megabytes": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockMeterCreater(ctrl)
				provider := mocks.NewMockMeter(ctrl)
				otMeter := otel.Meter(prefix + "_test")
				provisioned, err := otMeter.Float64ObservableUpDownCounter(prefix + "logical_provisioned_megabytes")
				if err != nil {
					t.Fatal(err)
				}
				empty, err := otMeter.Float64ObservableUpDownCounter("")
				if err != nil {
					t.Fatal(err)
				}
				meter.EXPECT().MeterProvider().Times(2)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(provisioned, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter.MeterProvider(),
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
				err := mws[i].RecordSpaceMetrics(context.Background(), metas[i], 1, 2)
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
				meter := mocks.NewMockMeterCreater(ctrl)
				provider := mocks.NewMockMeter(ctrl)
				otMeter := otel.Meter(prefix + "_test")
				provisioned, err := otMeter.Float64ObservableUpDownCounter(prefix + "logical_provisioned_megabytes")
				if err != nil {
					t.Fatal(err)
				}

				used, err := otMeter.Float64ObservableUpDownCounter(prefix + "logical_used_megabytes")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().MeterProvider().Times(2)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(provisioned, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(used, nil)

				return &service.MetricsWrapper{
					Meter: meter.MeterProvider(),
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_array_"),
			}

			return mws, checkFns(verifyNoError)
		},
		"error creating logical_provisioned_megabytes": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			meter := mocks.NewMockMeterCreater(ctrl)
			provider := mocks.NewMockMeter(ctrl)
			otMeter := otel.Meter("")
			empty, err := otMeter.Float64ObservableUpDownCounter("")
			if err != nil {
				t.Fatal(err)
			}

			meter.EXPECT().MeterProvider()
			provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error")).Times(1)

			mws := []*service.MetricsWrapper{{Meter: meter.MeterProvider()}}

			return mws, checkFns(verifyError)
		},
		"error creating logical_used_megabytes": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockMeterCreater(ctrl)
				provider := mocks.NewMockMeter(ctrl)
				otMeter := otel.Meter(prefix + "_test")
				readBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "logical_provisioned_megabytes")
				if err != nil {
					t.Fatal(err)
				}
				empty, err := otMeter.Float64ObservableUpDownCounter("")
				if err != nil {
					t.Fatal(err)
				}
				meter.EXPECT().MeterProvider().Times(2)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter.MeterProvider(),
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
				err := mws[i].RecordArraySpaceMetrics(context.Background(), "123", "driver", 1, 2)
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
				meter := mocks.NewMockMeterCreater(ctrl)
				provider := mocks.NewMockMeter(ctrl)
				otMeter := otel.Meter(prefix + "_test")
				provisioned, err := otMeter.Float64ObservableUpDownCounter(prefix + "logical_provisioned_megabytes")
				if err != nil {
					t.Fatal(err)
				}

				used, err := otMeter.Float64ObservableUpDownCounter(prefix + "logical_used_megabytes")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().MeterProvider().Times(2)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(provisioned, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(used, nil)

				return &service.MetricsWrapper{
					Meter: meter.MeterProvider(),
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_storage_class_"),
			}

			return mws, checkFns(verifyNoError)
		},
		"error creating logical_provisioned_megabytes": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			meter := mocks.NewMockMeterCreater(ctrl)
			provider := mocks.NewMockMeter(ctrl)
			otMeter := otel.Meter("")
			empty, err := otMeter.Float64ObservableUpDownCounter("")
			if err != nil {
				t.Fatal(err)
			}

			meter.EXPECT().MeterProvider()
			provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error")).Times(1)

			mws := []*service.MetricsWrapper{{Meter: meter.MeterProvider()}}

			return mws, checkFns(verifyError)
		},
		"error creating logical_used_megabytes": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockMeterCreater(ctrl)
				provider := mocks.NewMockMeter(ctrl)
				otMeter := otel.Meter(prefix + "_test")
				readBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "logical_provisioned_megabytes")
				if err != nil {
					t.Fatal(err)
				}
				empty, err := otMeter.Float64ObservableUpDownCounter("")
				if err != nil {
					t.Fatal(err)
				}
				meter.EXPECT().MeterProvider().Times(2)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter.MeterProvider(),
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
				err := mws[i].RecordStorageClassSpaceMetrics(context.Background(), "123", "driver", 1, 2)
				for _, check := range checks {
					check(t, err)
				}
			}
		})
	}
}

func Test_Volume_Metrics_Label_Update(t *testing.T) {
	metaFirst := &service.VolumeMeta{
		ID:                        "123",
		PersistentVolumeName:      "pvol0",
		PersistentVolumeClaimName: "pvc0",
		Namespace:                 "namespace0",
	}

	metaSecond := &service.VolumeMeta{
		ID:                        "123",
		PersistentVolumeName:      "pvol0",
		PersistentVolumeClaimName: "pvc0",
		Namespace:                 "namespace0",
	}

	expectedLables := []attribute.KeyValue{
		attribute.String("VolumeID", metaSecond.ID),
		attribute.String("PlotWithMean", "No"),
		attribute.String("PersistentVolumeName", metaSecond.PersistentVolumeName),
		attribute.String("PersistentVolumeClaimName", metaSecond.PersistentVolumeClaimName),
		attribute.String("Namespace", metaSecond.Namespace),
	}

	ctrl := gomock.NewController(t)

	meter := mocks.NewMockMeterCreater(ctrl)
	provider := mocks.NewMockMeter(ctrl)
	otMeter := otel.Meter("powerstore_volume_test")
	readBW, err := otMeter.Float64ObservableUpDownCounter("powerstore_volume_read_bw_megabytes_per_second")
	if err != nil {
		t.Fatal(err)
	}

	writeBW, err := otMeter.Float64ObservableUpDownCounter("powerstore_volume_write_bw_megabytes_per_second")
	if err != nil {
		t.Fatal(err)
	}

	readIOPS, err := otMeter.Float64ObservableUpDownCounter("powerstore_volume_read_iops_per_second")
	if err != nil {
		t.Fatal(err)
	}

	writeIOPS, err := otMeter.Float64ObservableUpDownCounter("powerstore_volume_write_iops_per_second")
	if err != nil {
		t.Fatal(err)
	}

	readLatency, err := otMeter.Float64ObservableUpDownCounter("powerstore_volume_read_latency_milliseconds")
	if err != nil {
		t.Fatal(err)
	}

	writeLatency, err := otMeter.Float64ObservableUpDownCounter("powerstore_volume_write_latency_milliseconds")
	if err != nil {
		t.Fatal(err)
	}

	syncBW, err := otMeter.Float64ObservableUpDownCounter("syncronization_bw_megabytes_per_second")
	if err != nil {
		t.Fatal(err)
	}

	mirrorBW, err := otMeter.Float64ObservableUpDownCounter("mirror_bw_megabytes_per_second")
	if err != nil {
		t.Fatal(err)
	}

	dataRemaining, err := otMeter.Float64ObservableUpDownCounter("data_remaining_bytes")
	if err != nil {
		t.Fatal(err)
	}

	meter.EXPECT().MeterProvider().Times(9)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeBW, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(readIOPS, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeIOPS, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(readLatency, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeLatency, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(syncBW, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(mirrorBW, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(dataRemaining, nil)

	mw := &service.MetricsWrapper{
		Meter: meter.MeterProvider(),
	}

	t.Run("success: volume metric labels updated", func(t *testing.T) {
		err := mw.Record(context.Background(), metaFirst, 1, 2, 3, 4, 5, 6, 7, 8, 9)
		if err != nil {
			t.Errorf("expected nil error (record #1), got %v", err)
		}
		err = mw.Record(context.Background(), metaSecond, 1, 2, 3, 4, 5, 6, 7, 8, 9)
		if err != nil {
			t.Errorf("expected nil error (record #2), got %v", err)
		}

		newLabels, ok := mw.Labels.Load(metaFirst.ID)
		if !ok {
			t.Errorf("expected labels to exist for %v, but did not find them", metaFirst.ID)
		}
		labels := newLabels.([]attribute.KeyValue)
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
	metaFirst := &service.SpaceVolumeMeta{
		ID:           "123",
		ArrayID:      "arr123",
		StorageClass: "powerstore",
		Protocol:     "scsi",
	}

	metaSecond := &service.SpaceVolumeMeta{
		ID:           "123",
		ArrayID:      "arr123",
		StorageClass: "powerstore",
		Protocol:     "scsi",
	}

	expectedLables := []attribute.KeyValue{
		attribute.String("VolumeID", metaSecond.ID),
		attribute.String("PersistentVolumeName", metaSecond.PersistentVolumeName),
		attribute.String("PersistentVolumeClaimName", metaSecond.PersistentVolumeClaimName),
		attribute.String("Namespace", metaSecond.Namespace),
		attribute.String("PlotWithMean", "No"),
	}

	ctrl := gomock.NewController(t)

	meter := mocks.NewMockMeterCreater(ctrl)
	provider := mocks.NewMockMeter(ctrl)
	otMeter := otel.Meter("powerstore_volume_test")
	provisioned, err := otMeter.Float64ObservableUpDownCounter("logical_provisioned_megabytes")
	if err != nil {
		t.Fatal(err)
	}

	used, err := otMeter.Float64ObservableUpDownCounter("logical_used_megabytes")
	if err != nil {
		t.Fatal(err)
	}

	meter.EXPECT().MeterProvider().Times(2)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(provisioned, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(used, nil)

	mw := &service.MetricsWrapper{
		Meter: meter.MeterProvider(),
	}

	t.Run("success: volume metric labels updated", func(t *testing.T) {
		err := mw.RecordSpaceMetrics(context.Background(), metaFirst, 1, 2)
		if err != nil {
			t.Errorf("expected nil error (record #1), got %v", err)
		}
		err = mw.RecordSpaceMetrics(context.Background(), metaSecond, 1, 2)
		if err != nil {
			t.Errorf("expected nil error (record #2), got %v", err)
		}

		newLabels, ok := mw.Labels.Load(metaFirst.ID)
		if !ok {
			t.Errorf("expected labels to exist for %v, but did not find them", metaFirst.ID)
		}
		labels := newLabels.([]attribute.KeyValue)
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

func Test_Filesystem_Metrics_Label_Update(t *testing.T) {
	metaFirst := &service.SpaceVolumeMeta{
		ID:           "123",
		ArrayID:      "arr123",
		StorageClass: "powerstore",
		Protocol:     "nfs",
	}

	metaSecond := &service.SpaceVolumeMeta{
		ID:           "123",
		ArrayID:      "arr123",
		StorageClass: "powerstore",
		Protocol:     "nfs",
	}

	expectedLables := []attribute.KeyValue{
		attribute.String("VolumeID", metaSecond.ID),
		attribute.String("PlotWithMean", "No"),
	}

	ctrl := gomock.NewController(t)

	meter := mocks.NewMockMeterCreater(ctrl)
	provider := mocks.NewMockMeter(ctrl)
	otMeter := otel.Meter("powerstore_volume_test")
	provisioned, err := otMeter.Float64ObservableUpDownCounter("logical_provisioned_megabytes")
	if err != nil {
		t.Fatal(err)
	}

	used, err := otMeter.Float64ObservableUpDownCounter("logical_used_megabytes")
	if err != nil {
		t.Fatal(err)
	}

	meter.EXPECT().MeterProvider().Times(2)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(provisioned, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(used, nil)

	mw := &service.MetricsWrapper{
		Meter: meter.MeterProvider(),
	}

	t.Run("success: volume metric labels updated", func(t *testing.T) {
		err := mw.RecordSpaceMetrics(context.Background(), metaFirst, 1, 2)
		if err != nil {
			t.Errorf("expected nil error (record #1), got %v", err)
		}
		err = mw.RecordSpaceMetrics(context.Background(), metaSecond, 1, 2)
		if err != nil {
			t.Errorf("expected nil error (record #2), got %v", err)
		}

		newLabels, ok := mw.Labels.Load(metaFirst.ID)
		if !ok {
			t.Errorf("expected labels to exist for %v, but did not find them", metaFirst.ID)
		}
		labels := newLabels.([]attribute.KeyValue)
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

	expectedLables := []attribute.KeyValue{
		attribute.String("ArrayID", array1),
		attribute.String("PlotWithMean", "No"),
	}

	ctrl := gomock.NewController(t)

	meter := mocks.NewMockMeterCreater(ctrl)
	provider := mocks.NewMockMeter(ctrl)
	otMeter := otel.Meter("powerstore_array_")
	provisioned, err := otMeter.Float64ObservableUpDownCounter("logical_provisioned_megabytes")
	if err != nil {
		t.Fatal(err)
	}

	used, err := otMeter.Float64ObservableUpDownCounter("logical_used_megabytes")
	if err != nil {
		t.Fatal(err)
	}

	meter.EXPECT().MeterProvider().Times(2)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(provisioned, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(used, nil)

	mw := &service.MetricsWrapper{
		Meter: meter.MeterProvider(),
	}

	t.Run("success: volume metric labels updated", func(t *testing.T) {
		err := mw.RecordArraySpaceMetrics(context.Background(), array1, "driver", 1, 2)
		if err != nil {
			t.Errorf("expected nil error (record #1), got %v", err)
		}
		err = mw.RecordArraySpaceMetrics(context.Background(), array2, "driver", 1, 2)
		if err != nil {
			t.Errorf("expected nil error (record #2), got %v", err)
		}

		newLabels, ok := mw.Labels.Load(array1)
		if !ok {
			t.Errorf("expected labels to exist for %v, but did not find them", array1)
		}
		labels := newLabels.([]attribute.KeyValue)
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

	expectedLables := []attribute.KeyValue{
		attribute.String("StorageClass", array1),
		attribute.String("PlotWithMean", "No"),
	}

	ctrl := gomock.NewController(t)

	meter := mocks.NewMockMeterCreater(ctrl)
	provider := mocks.NewMockMeter(ctrl)
	otMeter := otel.Meter("powerstore_storage_class_")
	provisioned, err := otMeter.Float64ObservableUpDownCounter("logical_provisioned_megabytes")
	if err != nil {
		t.Fatal(err)
	}

	used, err := otMeter.Float64ObservableUpDownCounter("logical_used_megabytes")
	if err != nil {
		t.Fatal(err)
	}

	meter.EXPECT().MeterProvider().Times(2)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(provisioned, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(used, nil)

	mw := &service.MetricsWrapper{
		Meter: meter.MeterProvider(),
	}

	t.Run("success: volume metric labels updated", func(t *testing.T) {
		err := mw.RecordStorageClassSpaceMetrics(context.Background(), array1, "driver", 1, 2)
		if err != nil {
			t.Errorf("expected nil error (record #1), got %v", err)
		}
		err = mw.RecordStorageClassSpaceMetrics(context.Background(), array2, "driver", 1, 2)
		if err != nil {
			t.Errorf("expected nil error (record #2), got %v", err)
		}

		newLabels, ok := mw.Labels.Load(array1)
		if !ok {
			t.Errorf("expected labels to exist for %v, but did not find them", array1)
		}
		labels := newLabels.([]attribute.KeyValue)
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

func Test_FileSystem_Metrics_Label_Update(t *testing.T) {
	metaFirst := &service.VolumeMeta{
		ID:                        "123",
		PersistentVolumeName:      "pvol0",
		PersistentVolumeClaimName: "pvc0",
		Namespace:                 "namespace0",
	}

	metaSecond := &service.VolumeMeta{
		ID:                        "123",
		PersistentVolumeName:      "pvol0",
		PersistentVolumeClaimName: "pvc0",
		Namespace:                 "namespace0",
	}

	expectedLables := []attribute.KeyValue{
		attribute.String("FileSystemID", metaSecond.ID),
		attribute.String("PlotWithMean", "No"),
		attribute.String("PersistentVolumeName", metaSecond.PersistentVolumeName),
		attribute.String("PersistentVolumeClaimName", metaSecond.PersistentVolumeClaimName),
		attribute.String("Namespace", metaSecond.Namespace),
	}

	ctrl := gomock.NewController(t)

	meter := mocks.NewMockMeterCreater(ctrl)
	provider := mocks.NewMockMeter(ctrl)
	otMeter := otel.Meter("powerstore_filesystem_test")
	readBW, err := otMeter.Float64ObservableUpDownCounter("powerstore_filesystem_read_bw_megabytes_per_second")
	if err != nil {
		t.Fatal(err)
	}

	writeBW, err := otMeter.Float64ObservableUpDownCounter("powerstore_filesystem_write_bw_megabytes_per_second")
	if err != nil {
		t.Fatal(err)
	}

	readIOPS, err := otMeter.Float64ObservableUpDownCounter("powerstore_filesystem_read_iops_per_second")
	if err != nil {
		t.Fatal(err)
	}

	writeIOPS, err := otMeter.Float64ObservableUpDownCounter("powerstore_filesystem_write_iops_per_second")
	if err != nil {
		t.Fatal(err)
	}

	readLatency, err := otMeter.Float64ObservableUpDownCounter("powerstore_filesystem_read_latency_milliseconds")
	if err != nil {
		t.Fatal(err)
	}

	writeLatency, err := otMeter.Float64ObservableUpDownCounter("powerstore_filesystem_write_latency_milliseconds")
	if err != nil {
		t.Fatal(err)
	}

	syncBW, err := otMeter.Float64ObservableUpDownCounter("syncronization_bw_megabytes_per_second")
	if err != nil {
		t.Fatal(err)
	}

	mirrorBW, err := otMeter.Float64ObservableUpDownCounter("mirror_bw_megabytes_per_second")
	if err != nil {
		t.Fatal(err)
	}

	dataRemaining, err := otMeter.Float64ObservableUpDownCounter("data_remaining_bytes")
	if err != nil {
		t.Fatal(err)
	}

	meter.EXPECT().MeterProvider().Times(9)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeBW, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(readIOPS, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeIOPS, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(readLatency, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeLatency, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(syncBW, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(mirrorBW, nil)
	provider.EXPECT().UpDownCounter(gomock.Any()).Return(dataRemaining, nil)

	mw := &service.MetricsWrapper{
		Meter: meter.MeterProvider(),
	}

	t.Run("success: filesystem metric labels updated", func(t *testing.T) {
		err := mw.RecordFileSystemMetrics(context.Background(), metaFirst, 1, 2, 3, 4, 5, 6, 7, 8, 9)
		if err != nil {
			t.Errorf("expected nil error (record #1), got %v", err)
		}
		err = mw.RecordFileSystemMetrics(context.Background(), metaSecond, 1, 2, 3, 4, 5, 6, 7, 8, 9)
		if err != nil {
			t.Errorf("expected nil error (record #2), got %v", err)
		}

		newLabels, ok := mw.Labels.Load(metaFirst.ID)
		if !ok {
			t.Errorf("expected labels to exist for %v, but did not find them", metaFirst.ID)
		}
		labels := newLabels.([]attribute.KeyValue)
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

func Test_Record_FileSystem_Metrics(t *testing.T) {
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
				meter := mocks.NewMockMeterCreater(ctrl)
				provider := mocks.NewMockMeter(ctrl)
				otMeter := otel.Meter(prefix + "_test")
				readBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "write_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readIOPS, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeIOPS, err := otMeter.Float64ObservableUpDownCounter(prefix + "write_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readLatency, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_latency_milliseconds")
				if err != nil {
					t.Fatal(err)
				}

				writeLatency, err := otMeter.Float64ObservableUpDownCounter(prefix + "write_latency_milliseconds")
				if err != nil {
					t.Fatal(err)
				}

				syncBW, err := otMeter.Float64ObservableUpDownCounter("syncronization_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				mirrorBW, err := otMeter.Float64ObservableUpDownCounter("mirror_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				dataRemaining, err := otMeter.Float64ObservableUpDownCounter("data_remaining_bytes")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().MeterProvider().Times(9)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readLatency, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeLatency, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(syncBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(mirrorBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(dataRemaining, nil)

				return &service.MetricsWrapper{
					Meter: meter.MeterProvider(),
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_filesystem_"),
			}

			return mws, checkFns(verifyNoError)
		},
		"error creating read_bw_megabytes_per_second": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			meter := mocks.NewMockMeterCreater(ctrl)
			provider := mocks.NewMockMeter(ctrl)
			otMeter := otel.Meter("")
			empty, err := otMeter.Float64ObservableUpDownCounter("")
			if err != nil {
				t.Fatal(err)
			}
			meter.EXPECT().MeterProvider()
			provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error")).Times(1)

			mws := []*service.MetricsWrapper{{Meter: meter.MeterProvider()}}

			return mws, checkFns(verifyError)
		},
		"error creating write_bw_megabytes_per_second": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockMeterCreater(ctrl)
				provider := mocks.NewMockMeter(ctrl)
				otMeter := otel.Meter(prefix + "_test")
				readBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}
				empty, err := otMeter.Float64ObservableUpDownCounter("")
				if err != nil {
					t.Fatal(err)
				}
				meter.EXPECT().MeterProvider().Times(2)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter.MeterProvider(),
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_filesystem_"),
			}

			return mws, checkFns(verifyError)
		},
		"error creating read_iops_per_second": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockMeterCreater(ctrl)
				provider := mocks.NewMockMeter(ctrl)
				otMeter := otel.Meter(prefix + "_test")

				readBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "write_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}
				empty, err := otMeter.Float64ObservableUpDownCounter("")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().MeterProvider().Times(3)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter.MeterProvider(),
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_filesystem_"),
			}

			return mws, checkFns(verifyError)
		},
		"error creating write_iops_per_second": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockMeterCreater(ctrl)
				provider := mocks.NewMockMeter(ctrl)
				otMeter := otel.Meter(prefix + "_test")

				readBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "write_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readIOPS, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				empty, err := otMeter.Float64ObservableUpDownCounter("")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().MeterProvider().Times(4)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter.MeterProvider(),
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_filesystem_"),
			}

			return mws, checkFns(verifyError)
		},
		"error creating read_latency_milliseconds": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockMeterCreater(ctrl)
				provider := mocks.NewMockMeter(ctrl)
				otMeter := otel.Meter(prefix + "_test")

				readBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "write_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readIOPS, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeIOPS, err := otMeter.Float64ObservableUpDownCounter(prefix + "write_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				empty, err := otMeter.Float64ObservableUpDownCounter("")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().MeterProvider().Times(5)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter.MeterProvider(),
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_filesystem_"),
			}

			return mws, checkFns(verifyError)
		},
		"error creating write_latency_milliseconds": func(t *testing.T) ([]*service.MetricsWrapper, []checkFn) {
			ctrl := gomock.NewController(t)
			getMeter := func(prefix string) *service.MetricsWrapper {
				meter := mocks.NewMockMeterCreater(ctrl)
				provider := mocks.NewMockMeter(ctrl)
				otMeter := otel.Meter(prefix + "_test")

				readBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeBW, err := otMeter.Float64ObservableUpDownCounter(prefix + "write_bw_megabytes_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readIOPS, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				writeIOPS, err := otMeter.Float64ObservableUpDownCounter(prefix + "write_iops_per_second")
				if err != nil {
					t.Fatal(err)
				}

				readLatency, err := otMeter.Float64ObservableUpDownCounter(prefix + "read_latency_milliseconds")
				if err != nil {
					t.Fatal(err)
				}

				empty, err := otMeter.Float64ObservableUpDownCounter("")
				if err != nil {
					t.Fatal(err)
				}

				meter.EXPECT().MeterProvider().Times(6)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeBW, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(writeIOPS, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(readLatency, nil)
				provider.EXPECT().UpDownCounter(gomock.Any()).Return(empty, errors.New("error"))

				return &service.MetricsWrapper{
					Meter: meter.MeterProvider(),
				}
			}

			mws := []*service.MetricsWrapper{
				getMeter("powerstore_filesystem_"),
			}

			return mws, checkFns(verifyError)
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mws, checks := tc(t)
			for i := range mws {
				err := mws[i].RecordFileSystemMetrics(context.Background(), metas[i], 1, 2, 3, 4, 5, 6, 7, 8, 9)
				for _, check := range checks {
					check(t, err)
				}
			}
		})
	}
}
