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

package service

import (
	"context"
	"errors"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument/asyncfloat64"
)

// MetricsRecorder supports recording I/O metrics
//
//go:generate mockgen -destination=mocks/metrics_mocks.go -package=mocks github.com/dell/csm-metrics-powerstore/internal/service MetricsRecorder,Float64UpDownCounterCreater
type MetricsRecorder interface {
	Record(ctx context.Context, meta interface{},
		readBW, writeBW,
		readIOPS, writeIOPS,
		readLatency, writeLatency, syncronizationBW, mirrorBW, dataRemaining float32) error
	RecordSpaceMetrics(ctx context.Context, meta interface{},
		logicalProvisioned, logicalUsed int64) error
	RecordArraySpaceMetrics(ctx context.Context, arrayID, driver string,
		logicalProvisioned, logicalUsed int64) error
	RecordStorageClassSpaceMetrics(ctx context.Context, storageclass, driver string,
		logicalProvisioned, logicalUsed int64) error
	RecordFileSystemMetrics(ctx context.Context, meta interface{},
		readBW, writeBW,
		readIOPS, writeIOPS,
		readLatency, writeLatency, syncronizationBW, mirrorBW, dataRemaining float32) error
}

// Float64UpDownCounterCreater creates a Float64UpDownCounter InstrumentProvider
//
//go:generate mockgen -destination=mocks/instrument_provider_mocks.go -package=mocks go.opentelemetry.io/otel/metric/instrument/asyncfloat64 InstrumentProvider
type Float64UpDownCounterCreater interface {
	AsyncFloat64() asyncfloat64.InstrumentProvider
}

// MetricsWrapper contains data used for pushing metrics data
type MetricsWrapper struct {
	Meter             Float64UpDownCounterCreater
	Metrics           sync.Map
	Labels            sync.Map
	SpaceMetrics      sync.Map
	ArraySpaceMetrics sync.Map
}

// SpaceMetrics contains the metrics related to a capacity
type SpaceMetrics struct {
	LogicalProvisioned asyncfloat64.UpDownCounter
	LogicalUsed        asyncfloat64.UpDownCounter
}

// ArraySpaceMetrics contains the metrics related to a capacity
type ArraySpaceMetrics struct {
	LogicalProvisioned asyncfloat64.UpDownCounter
	LogicalUsed        asyncfloat64.UpDownCounter
}

// Metrics contains the list of metrics data that is collected
type Metrics struct {
	ReadBW           asyncfloat64.UpDownCounter
	WriteBW          asyncfloat64.UpDownCounter
	ReadIOPS         asyncfloat64.UpDownCounter
	WriteIOPS        asyncfloat64.UpDownCounter
	ReadLatency      asyncfloat64.UpDownCounter
	WriteLatency     asyncfloat64.UpDownCounter
	SyncronizationBW asyncfloat64.UpDownCounter
	MirrorBW         asyncfloat64.UpDownCounter
	DataRemaining    asyncfloat64.UpDownCounter
}

func (mw *MetricsWrapper) initMetrics(prefix, metaID string, labels []attribute.KeyValue) (*Metrics, error) {
	readBW, err := mw.Meter.AsyncFloat64().UpDownCounter(prefix + "read_bw_megabytes_per_second")
	if err != nil {
		return nil, err
	}

	writeBW, err := mw.Meter.AsyncFloat64().UpDownCounter(prefix + "write_bw_megabytes_per_second")
	if err != nil {
		return nil, err
	}

	readIOPS, err := mw.Meter.AsyncFloat64().UpDownCounter(prefix + "read_iops_per_second")
	if err != nil {
		return nil, err
	}

	writeIOPS, err := mw.Meter.AsyncFloat64().UpDownCounter(prefix + "write_iops_per_second")
	if err != nil {
		return nil, err
	}

	readLatency, err := mw.Meter.AsyncFloat64().UpDownCounter(prefix + "read_latency_milliseconds")
	if err != nil {
		return nil, err
	}

	writeLatency, err := mw.Meter.AsyncFloat64().UpDownCounter(prefix + "write_latency_milliseconds")
	if err != nil {
		return nil, err
	}

	syncBW, err := mw.Meter.AsyncFloat64().UpDownCounter(prefix + "syncronization_bw_megabytes_per_second")
	if err != nil {
		return nil, err
	}

	mirrorBW, err := mw.Meter.AsyncFloat64().UpDownCounter(prefix + "mirror_bw_megabytes_per_second")
	if err != nil {
		return nil, err
	}

	dataRemaining, err := mw.Meter.AsyncFloat64().UpDownCounter(prefix + "data_remaining_bytes")
	if err != nil {
		return nil, err
	}

	metrics := &Metrics{
		ReadBW:           readBW,
		WriteBW:          writeBW,
		ReadIOPS:         readIOPS,
		WriteIOPS:        writeIOPS,
		ReadLatency:      readLatency,
		WriteLatency:     writeLatency,
		SyncronizationBW: syncBW,
		MirrorBW:         mirrorBW,
		DataRemaining:    dataRemaining,
	}

	mw.Metrics.Store(metaID, metrics)
	mw.Labels.Store(metaID, labels)

	return metrics, nil
}

// Record will publish metrics data for a given instance
func (mw *MetricsWrapper) Record(ctx context.Context, meta interface{},
	readBW, writeBW,
	readIOPS, writeIOPS,
	readLatency, writeLatency, syncronizationBW, mirrorBW, dataRemaining float32,
) error {
	var prefix string
	var metaID string
	var labels []attribute.KeyValue
	switch v := meta.(type) {
	case *VolumeMeta:
		prefix, metaID = "powerstore_volume_", v.ID
		labels = []attribute.KeyValue{
			attribute.String("VolumeID", v.ID),
			attribute.String("ArrayID", v.ArrayID),
			attribute.String("PersistentVolumeName", v.PersistentVolumeName),
			attribute.String("PersistentVolumeClaimName", v.PersistentVolumeClaimName),
			attribute.String("Namespace", v.Namespace),
			attribute.String("PlotWithMean", "No"),
		}
	default:
		return errors.New("unknown MetaData type")
	}

	metricsMapValue, ok := mw.Metrics.Load(metaID)
	if !ok {
		newMetrics, err := mw.initMetrics(prefix, metaID, labels)
		if err != nil {
			return err
		}
		metricsMapValue = newMetrics
	} else {
		// If Metrics for this MetricsWrapper exist, then check if any labels have changed and update them
		currentLabels, ok := mw.Labels.Load(metaID)
		if !ok {
			newMetrics, err := mw.initMetrics(prefix, metaID, labels)
			if err != nil {
				return err
			}
			metricsMapValue = newMetrics
		} else {
			currentLabels := currentLabels.([]attribute.KeyValue)
			updatedLabels := currentLabels
			haveLabelsChanged := false
			for i, current := range currentLabels {
				for _, new := range labels {
					if current.Key == new.Key {
						if current.Value != new.Value {
							updatedLabels[i].Value = new.Value
							haveLabelsChanged = true
						}
					}
				}
			}
			if haveLabelsChanged {
				newMetrics, err := mw.initMetrics(prefix, metaID, updatedLabels)
				if err != nil {
					return err
				}
				metricsMapValue = newMetrics
			}
		}
	}

	metrics := metricsMapValue.(*Metrics)

	metrics.ReadBW.Observe(ctx, float64(readBW), labels...)
	metrics.WriteBW.Observe(ctx, float64(writeBW), labels...)
	metrics.ReadIOPS.Observe(ctx, float64(readIOPS), labels...)
	metrics.WriteIOPS.Observe(ctx, float64(writeIOPS), labels...)
	metrics.ReadLatency.Observe(ctx, float64(readLatency), labels...)
	metrics.WriteLatency.Observe(ctx, float64(writeLatency), labels...)
	metrics.SyncronizationBW.Observe(ctx, float64(syncronizationBW), labels...)
	metrics.MirrorBW.Observe(ctx, float64(mirrorBW), labels...)
	metrics.DataRemaining.Observe(ctx, float64(dataRemaining), labels...)

	return nil
}

func (mw *MetricsWrapper) initSpaceMetrics(prefix, metaID string, labels []attribute.KeyValue) (*SpaceMetrics, error) {
	logicalProvisioned, err := mw.Meter.AsyncFloat64().UpDownCounter(prefix + "logical_provisioned_megabytes")
	if err != nil {
		return nil, err
	}

	logicalUsed, err := mw.Meter.AsyncFloat64().UpDownCounter(prefix + "logical_used_megabytes")
	if err != nil {
		return nil, err
	}

	metrics := &SpaceMetrics{
		LogicalProvisioned: logicalProvisioned,
		LogicalUsed:        logicalUsed,
	}

	mw.SpaceMetrics.Store(metaID, metrics)
	mw.Labels.Store(metaID, labels)

	return metrics, nil
}

// RecordSpaceMetrics will publish space metrics data for a given instance
func (mw *MetricsWrapper) RecordSpaceMetrics(ctx context.Context, meta interface{},
	logicalProvisioned, logicalUsed int64,
) error {
	var prefix string
	var metaID string
	var labels []attribute.KeyValue
	switch v := meta.(type) {
	case *SpaceVolumeMeta:
		switch v.Protocol {
		case nfsProtocol:
			prefix, metaID = "powerstore_filesystem_", v.ID
			labels = []attribute.KeyValue{
				attribute.String("FileSystemID", v.ID),
				attribute.String("ArrayID", v.ArrayID),
				attribute.String("PersistentVolumeName", v.PersistentVolumeName),
				attribute.String("PersistentVolumeClaimName", v.PersistentVolumeClaimName),
				attribute.String("Namespace", v.Namespace),
				attribute.String("StorageClass", v.StorageClass),
				attribute.String("Driver", v.Driver),
				attribute.String("Protocol", v.Protocol),
				attribute.String("PlotWithMean", "No"),
			}
		// volume metrics as default
		default:
			prefix, metaID = "powerstore_volume_", v.ID
			labels = []attribute.KeyValue{
				attribute.String("VolumeID", v.ID),
				attribute.String("ArrayID", v.ArrayID),
				attribute.String("PersistentVolumeName", v.PersistentVolumeName),
				attribute.String("PersistentVolumeClaimName", v.PersistentVolumeClaimName),
				attribute.String("Namespace", v.Namespace),
				attribute.String("StorageClass", v.StorageClass),
				attribute.String("Driver", v.Driver),
				attribute.String("Protocol", v.Protocol),
				attribute.String("PlotWithMean", "No"),
			}
		}
	default:
		return errors.New("unknown MetaData type")
	}

	metricsMapValue, ok := mw.SpaceMetrics.Load(metaID)
	if !ok {
		newMetrics, err := mw.initSpaceMetrics(prefix, metaID, labels)
		if err != nil {
			return err
		}
		metricsMapValue = newMetrics
	} else {
		// If Metrics for this MetricsWrapper exist, then check if any labels have changed and update them
		currentLabels, ok := mw.Labels.Load(metaID)
		if !ok {
			newMetrics, err := mw.initSpaceMetrics(prefix, metaID, labels)
			if err != nil {
				return err
			}
			metricsMapValue = newMetrics
		} else {
			currentLabels := currentLabels.([]attribute.KeyValue)
			updatedLabels := currentLabels
			haveLabelsChanged := false
			for i, current := range currentLabels {
				for _, new := range labels {
					if current.Key == new.Key {
						if current.Value != new.Value {
							updatedLabels[i].Value = new.Value
							haveLabelsChanged = true
						}
					}
				}
			}
			if haveLabelsChanged {
				newMetrics, err := mw.initSpaceMetrics(prefix, metaID, updatedLabels)
				if err != nil {
					return err
				}
				metricsMapValue = newMetrics
			}
		}
	}

	metrics := metricsMapValue.(*SpaceMetrics)
	metrics.LogicalProvisioned.Observe(ctx, float64(logicalProvisioned), labels...)
	metrics.LogicalUsed.Observe(ctx, float64(logicalUsed), labels...)
	return nil
}

func (mw *MetricsWrapper) initArraySpaceMetrics(prefix, metaID string, labels []attribute.KeyValue) (*ArraySpaceMetrics, error) {
	logicalProvisioned, err := mw.Meter.AsyncFloat64().UpDownCounter(prefix + "logical_provisioned_megabytes")
	if err != nil {
		return nil, err
	}

	logicalUsed, err := mw.Meter.AsyncFloat64().UpDownCounter(prefix + "logical_used_megabytes")
	if err != nil {
		return nil, err
	}

	metrics := &ArraySpaceMetrics{
		LogicalProvisioned: logicalProvisioned,
		LogicalUsed:        logicalUsed,
	}

	mw.ArraySpaceMetrics.Store(metaID, metrics)
	mw.Labels.Store(metaID, labels)

	return metrics, nil
}

// RecordArraySpaceMetrics will publish space metrics data for a given instance
func (mw *MetricsWrapper) RecordArraySpaceMetrics(ctx context.Context, arrayID, driver string,
	logicalProvisioned, logicalUsed int64,
) error {
	var prefix string
	var metaID string
	var labels []attribute.KeyValue

	prefix, metaID = "powerstore_array_", arrayID
	labels = []attribute.KeyValue{
		attribute.String("ArrayID", arrayID),
		attribute.String("Driver", driver),
		attribute.String("PlotWithMean", "No"),
	}

	metricsMapValue, ok := mw.ArraySpaceMetrics.Load(metaID)
	if !ok {
		newMetrics, err := mw.initArraySpaceMetrics(prefix, metaID, labels)
		if err != nil {
			return err
		}
		metricsMapValue = newMetrics
	} else {
		// If Metrics for this MetricsWrapper exist, then check if any labels have changed and update them
		currentLabels, ok := mw.Labels.Load(metaID)
		if !ok {
			newMetrics, err := mw.initArraySpaceMetrics(prefix, metaID, labels)
			if err != nil {
				return err
			}
			metricsMapValue = newMetrics
		} else {
			currentLabels := currentLabels.([]attribute.KeyValue)
			updatedLabels := currentLabels
			haveLabelsChanged := false
			for i, current := range currentLabels {
				for _, new := range labels {
					if current.Key == new.Key {
						if current.Value != new.Value {
							updatedLabels[i].Value = new.Value
							haveLabelsChanged = true
						}
					}
				}
			}
			if haveLabelsChanged {
				newMetrics, err := mw.initArraySpaceMetrics(prefix, metaID, updatedLabels)
				if err != nil {
					return err
				}
				metricsMapValue = newMetrics
			}
		}
	}

	metrics := metricsMapValue.(*ArraySpaceMetrics)
	metrics.LogicalProvisioned.Observe(ctx, float64(logicalProvisioned), labels...)
	metrics.LogicalUsed.Observe(ctx, float64(logicalUsed), labels...)

	return nil
}

// RecordStorageClassSpaceMetrics will publish space metrics for storage class
func (mw *MetricsWrapper) RecordStorageClassSpaceMetrics(ctx context.Context, storageclass, driver string,
	logicalProvisioned, logicalUsed int64,
) error {
	var prefix string
	var metaID string
	var labels []attribute.KeyValue

	prefix, metaID = "powerstore_storage_class_", storageclass
	labels = []attribute.KeyValue{
		attribute.String("StorageClass", storageclass),
		attribute.String("Driver", driver),
		attribute.String("PlotWithMean", "No"),
	}

	metricsMapValue, ok := mw.ArraySpaceMetrics.Load(metaID)
	if !ok {
		newMetrics, err := mw.initArraySpaceMetrics(prefix, metaID, labels)
		if err != nil {
			return err
		}
		metricsMapValue = newMetrics
	} else {
		// If Metrics for this MetricsWrapper exist, then check if any labels have changed and update them
		currentLabels, ok := mw.Labels.Load(metaID)
		if !ok {
			newMetrics, err := mw.initArraySpaceMetrics(prefix, metaID, labels)
			if err != nil {
				return err
			}
			metricsMapValue = newMetrics
		} else {
			currentLabels := currentLabels.([]attribute.KeyValue)
			updatedLabels := currentLabels
			haveLabelsChanged := false
			for i, current := range currentLabels {
				for _, new := range labels {
					if current.Key == new.Key {
						if current.Value != new.Value {
							updatedLabels[i].Value = new.Value
							haveLabelsChanged = true
						}
					}
				}
			}
			if haveLabelsChanged {
				newMetrics, err := mw.initArraySpaceMetrics(prefix, metaID, updatedLabels)
				if err != nil {
					return err
				}
				metricsMapValue = newMetrics
			}
		}
	}

	metrics := metricsMapValue.(*ArraySpaceMetrics)
	metrics.LogicalProvisioned.Observe(ctx, float64(logicalProvisioned), labels...)
	metrics.LogicalUsed.Observe(ctx, float64(logicalUsed), labels...)

	return nil
}

func (mw *MetricsWrapper) initFileSystemMetrics(prefix, metaID string, labels []attribute.KeyValue) (*Metrics, error) {
	readBW, err := mw.Meter.AsyncFloat64().UpDownCounter(prefix + "read_bw_megabytes_per_second")
	if err != nil {
		return nil, err
	}

	writeBW, err := mw.Meter.AsyncFloat64().UpDownCounter(prefix + "write_bw_megabytes_per_second")
	if err != nil {
		return nil, err
	}

	readIOPS, err := mw.Meter.AsyncFloat64().UpDownCounter(prefix + "read_iops_per_second")
	if err != nil {
		return nil, err
	}

	writeIOPS, err := mw.Meter.AsyncFloat64().UpDownCounter(prefix + "write_iops_per_second")
	if err != nil {
		return nil, err
	}

	readLatency, err := mw.Meter.AsyncFloat64().UpDownCounter(prefix + "read_latency_milliseconds")
	if err != nil {
		return nil, err
	}

	writeLatency, err := mw.Meter.AsyncFloat64().UpDownCounter(prefix + "write_latency_milliseconds")
	if err != nil {
		return nil, err
	}

	syncBW, err := mw.Meter.AsyncFloat64().UpDownCounter(prefix + "syncronization_bw_megabytes_per_second")
	if err != nil {
		return nil, err
	}

	mirrorBW, err := mw.Meter.AsyncFloat64().UpDownCounter(prefix + "mirror_bw_megabytes_per_second")
	if err != nil {
		return nil, err
	}

	dataRemaining, err := mw.Meter.AsyncFloat64().UpDownCounter(prefix + "data_remaining_bytes")
	if err != nil {
		return nil, err
	}

	metrics := &Metrics{
		ReadBW:           readBW,
		WriteBW:          writeBW,
		ReadIOPS:         readIOPS,
		WriteIOPS:        writeIOPS,
		ReadLatency:      readLatency,
		WriteLatency:     writeLatency,
		SyncronizationBW: syncBW,
		MirrorBW:         mirrorBW,
		DataRemaining:    dataRemaining,
	}

	mw.Metrics.Store(metaID, metrics)
	mw.Labels.Store(metaID, labels)

	return metrics, nil
}

// RecordFileSystemMetrics will publish filesystem metrics data for a given instance
func (mw *MetricsWrapper) RecordFileSystemMetrics(ctx context.Context, meta interface{},
	readBW, writeBW,
	readIOPS, writeIOPS,
	readLatency, writeLatency, syncBW, mirrorBW, dataRemaining float32,
) error {
	var prefix string
	var metaID string
	var labels []attribute.KeyValue
	switch v := meta.(type) {
	case *VolumeMeta:
		prefix, metaID = "powerstore_filesystem_", v.ID
		labels = []attribute.KeyValue{
			attribute.String("FileSystemID", v.ID),
			attribute.String("ArrayID", v.ArrayID),
			attribute.String("PersistentVolumeName", v.PersistentVolumeName),
			attribute.String("PersistentVolumeClaimName", v.PersistentVolumeClaimName),
			attribute.String("Namespace", v.Namespace),
			attribute.String("StorageClass", v.StorageClass),
			attribute.String("PlotWithMean", "No"),
		}
	default:
		return errors.New("unknown MetaData type")
	}

	metricsMapValue, ok := mw.Metrics.Load(metaID)
	if !ok {
		newMetrics, err := mw.initFileSystemMetrics(prefix, metaID, labels)
		if err != nil {
			return err
		}
		metricsMapValue = newMetrics
	} else {
		// If Metrics for this MetricsWrapper exist, then check if any labels have changed and update them
		currentLabels, ok := mw.Labels.Load(metaID)
		if !ok {
			newMetrics, err := mw.initFileSystemMetrics(prefix, metaID, labels)
			if err != nil {
				return err
			}
			metricsMapValue = newMetrics
		} else {
			currentLabels := currentLabels.([]attribute.KeyValue)
			updatedLabels := currentLabels
			haveLabelsChanged := false
			for i, current := range currentLabels {
				for _, new := range labels {
					if current.Key == new.Key {
						if current.Value != new.Value {
							updatedLabels[i].Value = new.Value
							haveLabelsChanged = true
						}
					}
				}
			}
			if haveLabelsChanged {
				newMetrics, err := mw.initFileSystemMetrics(prefix, metaID, updatedLabels)
				if err != nil {
					return err
				}
				metricsMapValue = newMetrics
			}
		}
	}

	metrics := metricsMapValue.(*Metrics)

	metrics.ReadBW.Observe(ctx, float64(readBW), labels...)
	metrics.WriteBW.Observe(ctx, float64(writeBW), labels...)
	metrics.ReadIOPS.Observe(ctx, float64(readIOPS), labels...)
	metrics.WriteIOPS.Observe(ctx, float64(writeIOPS), labels...)
	metrics.ReadLatency.Observe(ctx, float64(readLatency), labels...)
	metrics.WriteLatency.Observe(ctx, float64(writeLatency), labels...)
	metrics.DataRemaining.Observe(ctx, float64(dataRemaining), labels...)
	metrics.SyncronizationBW.Observe(ctx, float64(syncBW), labels...)
	metrics.MirrorBW.Observe(ctx, float64(mirrorBW), labels...)

	return nil
}
