/*
 Copyright (c) 2025 Dell Inc. or its subsidiaries. All Rights Reserved.

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
	"testing"

	otlexporters "github.com/dell/csm-metrics-powerstore/opentelemetry/exporters"

	"github.com/dell/csm-metrics-powerstore/internal/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func TestMetricsWrapper_Record(t *testing.T) {
	mw := &service.MetricsWrapper{
		Meter: otel.Meter("powerstore-test"),
	}
	volumeMetas := []interface{}{
		&service.VolumeMeta{
			ID: "123",
		},
	}
	spaceMetas := []interface{}{
		&service.SpaceVolumeMeta{
			ID: "123",
		},
	}

	exporter := &otlexporters.OtlCollectorExporter{}
	err := exporter.InitExporter()
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx              context.Context
		meta             interface{}
		readBW           float32
		writeBW          float32
		readIOPS         float32
		writeIOPS        float32
		readLatency      float32
		writeLatency     float32
		syncronizationBW float32
		mirrorBW         float32
		dataRemaining    float32
	}
	tests := []struct {
		name    string
		mw      *service.MetricsWrapper
		args    args
		wantErr bool
	}{
		{
			name: "success",
			mw:   mw,
			args: args{
				ctx:              context.Background(),
				meta:             volumeMetas[0],
				readBW:           1,
				writeBW:          2,
				readIOPS:         3,
				writeIOPS:        4,
				readLatency:      5,
				writeLatency:     6,
				syncronizationBW: 7,
				mirrorBW:         8,
				dataRemaining:    9,
			},
			wantErr: false,
		},
		{
			name: "fail",
			mw:   mw,
			args: args{
				ctx:              context.Background(),
				meta:             spaceMetas[0],
				readBW:           1,
				writeBW:          2,
				readIOPS:         3,
				writeIOPS:        4,
				readLatency:      5,
				writeLatency:     6,
				syncronizationBW: 7,
				mirrorBW:         8,
				dataRemaining:    9,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.mw.Record(tt.args.ctx, tt.args.meta, tt.args.readBW, tt.args.writeBW, tt.args.readIOPS, tt.args.writeIOPS, tt.args.readLatency, tt.args.writeLatency, tt.args.syncronizationBW, tt.args.mirrorBW, tt.args.dataRemaining); (err != nil) != tt.wantErr {
				t.Errorf("MetricsWrapper.Record() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetricsWrapper_Record_Label_Update(t *testing.T) {
	mw := &service.MetricsWrapper{
		Meter: otel.Meter("powerstore-test"),
	}

	exporter := &otlexporters.OtlCollectorExporter{}
	err := exporter.InitExporter()
	if err != nil {
		t.Fatal(err)
	}

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
	metaThird := &service.VolumeMeta{
		ID:                        "123",
		PersistentVolumeName:      "pvol1",
		PersistentVolumeClaimName: "pvc1",
		Namespace:                 "namespace0",
	}

	expectedLables := []attribute.KeyValue{
		attribute.String("VolumeID", metaSecond.ID),
		attribute.String("PlotWithMean", "No"),
		attribute.String("PersistentVolumeName", metaSecond.PersistentVolumeName),
		attribute.String("PersistentVolumeClaimName", metaSecond.PersistentVolumeClaimName),
		attribute.String("Namespace", metaSecond.Namespace),
	}
	expectedLablesUpdate := []attribute.KeyValue{
		attribute.String("VolumeID", metaThird.ID),
		attribute.String("PlotWithMean", "No"),
		attribute.String("PersistentVolumeName", metaThird.PersistentVolumeName),
		attribute.String("PersistentVolumeClaimName", metaThird.PersistentVolumeClaimName),
		attribute.String("Namespace", metaThird.Namespace),
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

	t.Run("success: volume metric labels updated with PV Name and PVC Update", func(t *testing.T) {
		err := mw.Record(context.Background(), metaFirst, 1, 2, 3, 4, 5, 6, 7, 8, 9)
		if err != nil {
			t.Errorf("expected nil error (record #1), got %v", err)
		}
		err = mw.Record(context.Background(), metaThird, 1, 2, 3, 4, 5, 6, 7, 8, 9)
		if err != nil {
			t.Errorf("expected nil error (record #2), got %v", err)
		}

		newLabels, ok := mw.Labels.Load(metaThird.ID)
		if !ok {
			t.Errorf("expected labels to exist for %v, but did not find them", metaThird.ID)
		}
		labels := newLabels.([]attribute.KeyValue)
		for _, l := range labels {
			for _, e := range expectedLablesUpdate {
				if l.Key == e.Key {
					if l.Value.AsString() != e.Value.AsString() {
						t.Errorf("expected label %v to be updated to %v, but the value was %v", e.Key, e.Value.AsString(), l.Value.AsString())
					}
				}
			}
		}
	})
}

func TestMetricsWrapper_RecordSpaceMetrics(t *testing.T) {
	mw := &service.MetricsWrapper{
		Meter: otel.Meter("powerstore-test"),
	}
	spaceMetas := []interface{}{
		&service.SpaceVolumeMeta{
			ID: "123",
		},
	}
	spaceMetasNFS := []interface{}{
		&service.SpaceVolumeMeta{
			ID:       "123",
			Protocol: "nfs",
		},
	}
	volumeMetas := []interface{}{
		&service.VolumeMeta{
			ID: "123",
		},
	}

	exporter := &otlexporters.OtlCollectorExporter{}
	err := exporter.InitExporter()
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx                context.Context
		meta               interface{}
		logicalProvisioned int64
		logicalUsed        int64
	}
	tests := []struct {
		name    string
		mw      *service.MetricsWrapper
		args    args
		wantErr bool
	}{
		{
			name: "success",
			mw:   mw,
			args: args{
				ctx:                context.Background(),
				meta:               spaceMetas[0],
				logicalProvisioned: 1,
				logicalUsed:        2,
			},
			wantErr: false,
		},
		{
			name: "success nfs",
			mw:   mw,
			args: args{
				ctx:                context.Background(),
				meta:               spaceMetasNFS[0],
				logicalProvisioned: 1,
				logicalUsed:        2,
			},
			wantErr: false,
		},
		{
			name: "fail",
			mw:   mw,
			args: args{
				ctx:                context.Background(),
				meta:               volumeMetas[0],
				logicalProvisioned: 1,
				logicalUsed:        2,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.mw.RecordSpaceMetrics(tt.args.ctx, tt.args.meta, tt.args.logicalProvisioned, tt.args.logicalUsed); (err != nil) != tt.wantErr {
				t.Errorf("MetricsWrapper.RecordSpaceMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetricsWrapper_RecordSpaceMetrics_Label_Update(t *testing.T) {
	mw := &service.MetricsWrapper{
		Meter: otel.Meter("powerstore-test"),
	}
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

	metaThird := &service.SpaceVolumeMeta{
		ID:           "123",
		ArrayID:      "arr125",
		StorageClass: "powerstore",
		Protocol:     "scsi",
	}

	exporter := &otlexporters.OtlCollectorExporter{}
	err := exporter.InitExporter()
	if err != nil {
		t.Fatal(err)
	}

	expectedLables := []attribute.KeyValue{
		attribute.String("VolumeID", metaSecond.ID),
		attribute.String("PersistentVolumeName", metaSecond.PersistentVolumeName),
		attribute.String("PersistentVolumeClaimName", metaSecond.PersistentVolumeClaimName),
		attribute.String("Namespace", metaSecond.Namespace),
		attribute.String("PlotWithMean", "No"),
	}

	expectedLablesUpdate := []attribute.KeyValue{
		attribute.String("VolumeID", metaThird.ID),
		attribute.String("PersistentVolumeName", metaThird.PersistentVolumeName),
		attribute.String("PersistentVolumeClaimName", metaThird.PersistentVolumeClaimName),
		attribute.String("Namespace", metaThird.Namespace),
		attribute.String("PlotWithMean", "No"),
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

	t.Run("success: volume metric labels with ArrayID updated", func(t *testing.T) {
		err := mw.RecordSpaceMetrics(context.Background(), metaFirst, 1, 2)
		if err != nil {
			t.Errorf("expected nil error (record #1), got %v", err)
		}
		err = mw.RecordSpaceMetrics(context.Background(), metaThird, 1, 2)
		if err != nil {
			t.Errorf("expected nil error (record #2), got %v", err)
		}

		newLabels, ok := mw.Labels.Load(metaThird.ID)
		if !ok {
			t.Errorf("expected labels to exist for %v, but did not find them", metaThird.ID)
		}
		labels := newLabels.([]attribute.KeyValue)
		for _, l := range labels {
			for _, e := range expectedLablesUpdate {
				if l.Key == e.Key {
					if l.Value.AsString() != e.Value.AsString() {
						t.Errorf("expected label %v to be updated to %v, but the value was %v", e.Key, e.Value.AsString(), l.Value.AsString())
					}
				}
			}
		}
	})
}

func TestMetricsWrapper_RecordArraySpaceMetrics(t *testing.T) {
	mw := &service.MetricsWrapper{
		Meter: otel.Meter("powerstore-test"),
	}

	exporter := &otlexporters.OtlCollectorExporter{}
	err := exporter.InitExporter()
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx                context.Context
		arrayID            string
		driver             string
		logicalProvisioned int64
		logicalUsed        int64
	}
	tests := []struct {
		name    string
		mw      *service.MetricsWrapper
		args    args
		wantErr bool
	}{
		{
			name: "success",
			mw:   mw,
			args: args{
				ctx:                context.Background(),
				arrayID:            "123",
				driver:             "driver",
				logicalProvisioned: 1,
				logicalUsed:        2,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.mw.RecordArraySpaceMetrics(tt.args.ctx, tt.args.arrayID, tt.args.driver, tt.args.logicalProvisioned, tt.args.logicalUsed); (err != nil) != tt.wantErr {
				t.Errorf("MetricsWrapper.RecordArraySpaceMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetricsWrapper_RecordArraySpaceMetrics_Label_Update(t *testing.T) {
	mw := &service.MetricsWrapper{
		Meter: otel.Meter("powerstore-test"),
	}

	exporter := &otlexporters.OtlCollectorExporter{}
	err := exporter.InitExporter()
	if err != nil {
		t.Fatal(err)
	}

	array1 := "123"
	array2 := "123"
	array3 := "125"

	expectedLables := []attribute.KeyValue{
		attribute.String("ArrayID", array2),
		attribute.String("PlotWithMean", "No"),
	}
	expectedLablesUpdate := []attribute.KeyValue{
		attribute.String("ArrayID", array3),
		attribute.String("PlotWithMean", "No"),
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

		newLabels, ok := mw.Labels.Load(array2)
		if !ok {
			t.Errorf("expected labels to exist for %v, but did not find them", array2)
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

	t.Run("success: volume metric labels updated", func(t *testing.T) {
		err := mw.RecordArraySpaceMetrics(context.Background(), array1, "driver", 1, 2)
		if err != nil {
			t.Errorf("expected nil error (record #1), got %v", err)
		}
		err = mw.RecordArraySpaceMetrics(context.Background(), array3, "driver", 1, 2)
		if err != nil {
			t.Errorf("expected nil error (record #2), got %v", err)
		}

		newLabels, ok := mw.Labels.Load(array3)
		if !ok {
			t.Errorf("expected labels to exist for %v, but did not find them", array3)
		}
		labels := newLabels.([]attribute.KeyValue)
		for _, l := range labels {
			for _, e := range expectedLablesUpdate {
				if l.Key == e.Key {
					if l.Value.AsString() != e.Value.AsString() {
						t.Errorf("expected label %v to be updated to %v, but the value was %v", e.Key, e.Value.AsString(), l.Value.AsString())
					}
				}
			}
		}
	})
	t.Run("success: label change triggers metric reinitialization", func(t *testing.T) {
		arrayID := "array-456"

		// Initial call with one driver
		err := mw.RecordArraySpaceMetrics(context.Background(), arrayID, "driverA", 100, 50)
		if err != nil {
			t.Errorf("expected nil error (initial record), got %v", err)
		}

		// Simulate label change by manually storing different labels
		mw.Labels.Store(arrayID, []attribute.KeyValue{
			attribute.String("ArrayID", arrayID),
			attribute.String("Driver", "driverB"), // Different driver to trigger label update
			attribute.String("PlotWithMean", "No"),
		})

		// Second call with updated driver
		err = mw.RecordArraySpaceMetrics(context.Background(), arrayID, "driverA", 200, 100)
		if err != nil {
			t.Errorf("expected nil error (updated record), got %v", err)
		}

		// Validate that the label was updated
		newLabels, ok := mw.Labels.Load(arrayID)
		if !ok {
			t.Errorf("expected labels to exist for %v, but did not find them", arrayID)
		}
		labels := newLabels.([]attribute.KeyValue)
		for _, l := range labels {
			if l.Key == "Driver" && l.Value.AsString() != "driverA" {
				t.Errorf("expected Driver to be updated to 'driverA', got %v", l.Value.AsString())
			}
		}
	})
}

func TestMetricsWrapper_RecordStorageClassSpaceMetrics(t *testing.T) {
	mw := &service.MetricsWrapper{
		Meter: otel.Meter("powerstore-test"),
	}

	exporter := &otlexporters.OtlCollectorExporter{}
	err := exporter.InitExporter()
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx                context.Context
		storageclass       string
		driver             string
		logicalProvisioned int64
		logicalUsed        int64
	}
	tests := []struct {
		name    string
		mw      *service.MetricsWrapper
		args    args
		wantErr bool
	}{
		{
			name: "success",
			mw:   mw,
			args: args{
				ctx:                context.Background(),
				storageclass:       "storageclass",
				driver:             "driver",
				logicalProvisioned: 1,
				logicalUsed:        2,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.mw.RecordStorageClassSpaceMetrics(tt.args.ctx, tt.args.storageclass, tt.args.driver, tt.args.logicalProvisioned, tt.args.logicalUsed); (err != nil) != tt.wantErr {
				t.Errorf("MetricsWrapper.RecordStorageClassSpaceMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetricsWrapper_RecordStorageClassSpaceMetrics_Label_Update(t *testing.T) {
	mw := &service.MetricsWrapper{
		Meter: otel.Meter("powerstore-test"),
	}

	exporter := &otlexporters.OtlCollectorExporter{}
	err := exporter.InitExporter()
	if err != nil {
		t.Fatal(err)
	}

	array1 := "storageclass"
	array2 := "storageclass"
	array3 := "storageclass2"

	expectedLables := []attribute.KeyValue{
		attribute.String("StorageClass", array2),
		attribute.String("PlotWithMean", "No"),
	}

	expectedLablesUpdate := []attribute.KeyValue{
		attribute.String("StorageClass", array3),
		attribute.String("PlotWithMean", "No"),
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

		newLabels, ok := mw.Labels.Load(array2)
		if !ok {
			t.Errorf("expected labels to exist for %v, but did not find them", array2)
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

	t.Run("success: volume metric labels updated with array updated", func(t *testing.T) {
		err := mw.RecordStorageClassSpaceMetrics(context.Background(), array1, "driver", 1, 2)
		if err != nil {
			t.Errorf("expected nil error (record #1), got %v", err)
		}
		err = mw.RecordStorageClassSpaceMetrics(context.Background(), array3, "driver", 1, 2)
		if err != nil {
			t.Errorf("expected nil error (record #2), got %v", err)
		}

		newLabels, ok := mw.Labels.Load(array3)
		if !ok {
			t.Errorf("expected labels to exist for %v, but did not find them", array3)
		}
		labels := newLabels.([]attribute.KeyValue)
		for _, l := range labels {
			for _, e := range expectedLablesUpdate {
				if l.Key == e.Key {
					if l.Value.AsString() != e.Value.AsString() {
						t.Errorf("expected label %v to be updated to %v, but the value was %v", e.Key, e.Value.AsString(), l.Value.AsString())
					}
				}
			}
		}
	})
	t.Run("success: existing labels are updated when changed", func(t *testing.T) {
		// Step 1: Record initial metrics with one set of labels
		err := mw.RecordStorageClassSpaceMetrics(context.Background(), "sc-original", "driverA", 100, 50)
		if err != nil {
			t.Errorf("expected nil error (initial record), got %v", err)
		}

		// Step 2: Manually override the stored labels to simulate a change
		mw.Labels.Store("sc-original", []attribute.KeyValue{
			attribute.String("StorageClass", "sc-original"),
			attribute.String("Driver", "driverA"),
			attribute.String("PlotWithMean", "Yes"), // This differs from the new label
		})

		// Step 3: Call again with updated label value to trigger label update logic
		err = mw.RecordStorageClassSpaceMetrics(context.Background(), "sc-original", "driverA", 200, 100)
		if err != nil {
			t.Errorf("expected nil error (updated record), got %v", err)
		}

		// Step 4: Verify that the label was updated
		newLabels, ok := mw.Labels.Load("sc-original")
		if !ok {
			t.Errorf("expected labels to exist for %v, but did not find them", "sc-original")
		}
		labels := newLabels.([]attribute.KeyValue)
		for _, l := range labels {
			if l.Key == "PlotWithMean" && l.Value.AsString() != "No" {
				t.Errorf("expected PlotWithMean to be updated to 'No', got %v", l.Value.AsString())
			}
		}
	})
}

func TestMetricsWrapper_RecordFileSystemMetrics(t *testing.T) {
	mw := &service.MetricsWrapper{
		Meter: otel.Meter("powerstore-test"),
	}
	volumeMetas := []interface{}{
		&service.VolumeMeta{
			ID: "123",
		},
	}
	spaceMetas := []interface{}{
		&service.SpaceVolumeMeta{
			ID: "123",
		},
	}

	exporter := &otlexporters.OtlCollectorExporter{}
	err := exporter.InitExporter()
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx           context.Context
		meta          interface{}
		readBW        float32
		writeBW       float32
		readIOPS      float32
		writeIOPS     float32
		readLatency   float32
		writeLatency  float32
		syncBW        float32
		mirrorBW      float32
		dataRemaining float32
	}
	tests := []struct {
		name    string
		mw      *service.MetricsWrapper
		args    args
		wantErr bool
	}{
		{
			name: "success",
			mw:   mw,
			args: args{
				ctx:           context.Background(),
				meta:          volumeMetas[0],
				readBW:        1,
				writeBW:       2,
				readIOPS:      3,
				writeIOPS:     4,
				readLatency:   5,
				writeLatency:  6,
				syncBW:        7,
				mirrorBW:      8,
				dataRemaining: 9,
			},
			wantErr: false,
		},
		{
			name: "fail",
			mw:   mw,
			args: args{
				ctx:           context.Background(),
				meta:          spaceMetas[0],
				readBW:        1,
				writeBW:       2,
				readIOPS:      3,
				writeIOPS:     4,
				readLatency:   5,
				writeLatency:  6,
				syncBW:        7,
				mirrorBW:      8,
				dataRemaining: 9,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.mw.RecordFileSystemMetrics(tt.args.ctx, tt.args.meta, tt.args.readBW, tt.args.writeBW, tt.args.readIOPS, tt.args.writeIOPS, tt.args.readLatency, tt.args.writeLatency, tt.args.syncBW, tt.args.mirrorBW, tt.args.dataRemaining); (err != nil) != tt.wantErr {
				t.Errorf("MetricsWrapper.RecordFileSystemMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetricsWrapper_RecordFileSystemMetrics_Label_Update(t *testing.T) {
	mw := &service.MetricsWrapper{
		Meter: otel.Meter("powerstore-test"),
	}

	exporter := &otlexporters.OtlCollectorExporter{}
	err := exporter.InitExporter()
	if err != nil {
		t.Fatal(err)
	}

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

	metaThird := &service.VolumeMeta{
		ID:                        "123",
		PersistentVolumeName:      "pvol1",
		PersistentVolumeClaimName: "pvc1",
		Namespace:                 "namespace0",
	}

	expectedLables := []attribute.KeyValue{
		attribute.String("FileSystemID", metaSecond.ID),
		attribute.String("PlotWithMean", "No"),
		attribute.String("PersistentVolumeName", metaSecond.PersistentVolumeName),
		attribute.String("PersistentVolumeClaimName", metaSecond.PersistentVolumeClaimName),
		attribute.String("Namespace", metaSecond.Namespace),
	}

	expectedLablesUpdate := []attribute.KeyValue{
		attribute.String("FileSystemID", metaThird.ID),
		attribute.String("PlotWithMean", "No"),
		attribute.String("PersistentVolumeName", metaThird.PersistentVolumeName),
		attribute.String("PersistentVolumeClaimName", metaThird.PersistentVolumeClaimName),
		attribute.String("Namespace", metaThird.Namespace),
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

		newLabels, ok := mw.Labels.Load(metaSecond.ID)
		if !ok {
			t.Errorf("expected labels to exist for %v, but did not find them", metaSecond.ID)
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

	t.Run("success: filesystem metric labels with PV and PVC namesupdated", func(t *testing.T) {
		err := mw.RecordFileSystemMetrics(context.Background(), metaFirst, 1, 2, 3, 4, 5, 6, 7, 8, 9)
		if err != nil {
			t.Errorf("expected nil error (record #1), got %v", err)
		}
		err = mw.RecordFileSystemMetrics(context.Background(), metaThird, 1, 2, 3, 4, 5, 6, 7, 8, 9)
		if err != nil {
			t.Errorf("expected nil error (record #2), got %v", err)
		}

		newLabels, ok := mw.Labels.Load(metaThird.ID)
		if !ok {
			t.Errorf("expected labels to exist for %v, but did not find them", metaThird.ID)
		}
		labels := newLabels.([]attribute.KeyValue)
		for _, l := range labels {
			for _, e := range expectedLablesUpdate {
				if l.Key == e.Key {
					if l.Value.AsString() != e.Value.AsString() {
						t.Errorf("expected label %v to be updated to %v, but the value was %v", e.Key, e.Value.AsString(), l.Value.AsString())
					}
				}
			}
		}
	})
}

func TestMetricsWrapper_RecordTopologyMetrics(t *testing.T) {
	mw := &service.MetricsWrapper{
		Meter: otel.Meter("powerstore-topology-test"),
	}

	exporter := &otlexporters.OtlCollectorExporter{}
	err := exporter.InitExporter()
	if err != nil {
		t.Fatal(err)
	}

	// Pre-populate sync.Maps to cover the else branch in the method
	metaID := "pv-123"
	initialLabels := []attribute.KeyValue{
		attribute.String("Namespace", "default"),
		attribute.String("PersistentVolumeClaim", "pvc-123"),
		attribute.String("PersistentVolumeStatus", "Pending"), // Different to force label update
		attribute.String("VolumeClaimName", "claim-123"),
		attribute.String("PersistentVolume", metaID),
		attribute.String("StorageClass", "standard"),
		attribute.String("Driver", "csi-powerstore"),
		attribute.String("ProvisionedSize", "100Gi"),
		attribute.String("StorageSystemVolumeName", "vol-123"),
		attribute.String("StorageSystem", "system-1"),
		attribute.String("Protocol", "iSCSI"),
		attribute.String("CreatedTime", "2023-01-01T00:00:00Z"),
		attribute.String("PlotWithMean", "No"),
	}

	// Dummy metrics to satisfy the method usage
	dummyMetrics := &service.TopologyMetrics{
		PvAvailabilityMetric: nil,
	}

	mw.TopologyMetrics.Store(metaID, dummyMetrics)
	mw.Labels.Store(metaID, initialLabels)

	type args struct {
		ctx             context.Context
		meta            interface{}
		topologyMetrics *service.TopologyMetricsRecord
	}

	tests := []struct {
		name    string
		mw      *service.MetricsWrapper
		args    args
		wantErr bool
	}{
		{
			name: "success",
			mw:   mw,
			args: args{
				ctx: context.Background(),
				meta: &service.TopologyMeta{
					Namespace:               "default",
					PersistentVolumeClaim:   "pvc-123",
					PersistentVolumeStatus:  "Bound",
					VolumeClaimName:         "claim-123",
					PersistentVolume:        metaID,
					StorageClass:            "standard",
					Driver:                  "csi-powerstore",
					ProvisionedSize:         "100Gi",
					StorageSystemVolumeName: "vol-123",
					StorageSystem:           "system-1",
					Protocol:                "iSCSI",
					CreatedTime:             "2023-01-01T00:00:00Z",
				},
				topologyMetrics: &service.TopologyMetricsRecord{
					PvAvailable: 1024,
				},
			},
			wantErr: false,
		},
		{
			name: "label update triggers reinit",
			mw:   mw,
			args: args{
				ctx: context.Background(),
				meta: &service.TopologyMeta{
					Namespace:               "default",
					PersistentVolumeClaim:   "pvc-123",
					PersistentVolumeStatus:  "Bound",
					VolumeClaimName:         "claim-123",
					PersistentVolume:        metaID,
					StorageClass:            "standard",
					Driver:                  "csi-powerstore",
					ProvisionedSize:         "100Gi",
					StorageSystemVolumeName: "vol-123",
					StorageSystem:           "system-1",
					Protocol:                "iSCSI",
					CreatedTime:             "2023-01-01T00:00:00Z",
				},
				topologyMetrics: &service.TopologyMetricsRecord{
					PvAvailable: 2048,
				},
			},
			wantErr: false,
		},
		{
			name: "existing metrics no label change",
			mw:   mw,
			args: args{
				ctx: context.Background(),
				meta: &service.TopologyMeta{
					Namespace:               "default",
					PersistentVolumeClaim:   "pvc-123",
					PersistentVolumeStatus:  "Pending",
					VolumeClaimName:         "claim-123",
					PersistentVolume:        metaID,
					StorageClass:            "standard",
					Driver:                  "csi-powerstore",
					ProvisionedSize:         "100Gi",
					StorageSystemVolumeName: "vol-123",
					StorageSystem:           "system-1",
					Protocol:                "iSCSI",
					CreatedTime:             "2023-01-01T00:00:00Z",
				},
				topologyMetrics: &service.TopologyMetricsRecord{
					PvAvailable: 512,
				},
			},
			wantErr: false,
		},
		{
			name: "new metrics initialized when not found in TopologyMetrics map",
			mw: &service.MetricsWrapper{
				Meter: otel.Meter("powerstore-topology-test"),
			},
			args: args{
				ctx: context.Background(),
				meta: &service.TopologyMeta{
					Namespace:               "default",
					PersistentVolumeClaim:   "pvc-new",
					PersistentVolumeStatus:  "Bound",
					VolumeClaimName:         "claim-new",
					PersistentVolume:        "pv-new", // This metaID is not preloaded
					StorageClass:            "gold",
					Driver:                  "csi-powerstore",
					ProvisionedSize:         "200Gi",
					StorageSystemVolumeName: "vol-new",
					StorageSystem:           "system-2",
					Protocol:                "NVMe",
					CreatedTime:             "2024-01-01T00:00:00Z",
				},
				topologyMetrics: &service.TopologyMetricsRecord{
					PvAvailable: 1024,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.mw.RecordTopologyMetrics(tt.args.ctx, tt.args.meta, tt.args.topologyMetrics); (err != nil) != tt.wantErr {
				t.Errorf("MetricsWrapper.RecordTopologyMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
