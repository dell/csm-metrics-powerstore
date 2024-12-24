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

	"go.opentelemetry.io/otel/metric"

	"go.opentelemetry.io/otel/attribute"
)

// MetricsRecorder supports recording I/O metrics
//
//go:generate mockgen -destination=mocks/metrics_mocks.go -package=mocks github.com/dell/csm-metrics-powerstore/internal/service MetricsRecorder,MeterCreater
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

// MeterCreater interface is used to create and provide Meter instances, which are used to report measurements.
//
//go:generate mockgen -destination=mocks/meter_mocks.go -package=mocks go.opentelemetry.io/otel/metric Meter
type MeterCreater interface {
	// AsyncFloat64() asyncfloat64.InstrumentProvider
	MeterProvider() metric.Meter
	// metric.Float64ObservableUpDownCounter
}

// MetricsWrapper contains data used for pushing metrics data
type MetricsWrapper struct {
	Meter             metric.Meter
	Metrics           sync.Map
	Labels            sync.Map
	SpaceMetrics      sync.Map
	ArraySpaceMetrics sync.Map
}

// SpaceMetrics contains the metrics related to a capacity
type SpaceMetrics struct {
	LogicalProvisioned metric.Float64ObservableUpDownCounter
	LogicalUsed        metric.Float64ObservableUpDownCounter
}

// ArraySpaceMetrics contains the metrics related to a capacity
type ArraySpaceMetrics struct {
	LogicalProvisioned metric.Float64ObservableUpDownCounter
	LogicalUsed        metric.Float64ObservableUpDownCounter
}

// Metrics contains the list of metrics data that is collected
type Metrics struct {
	ReadBW           metric.Float64ObservableUpDownCounter
	WriteBW          metric.Float64ObservableUpDownCounter
	ReadIOPS         metric.Float64ObservableUpDownCounter
	WriteIOPS        metric.Float64ObservableUpDownCounter
	ReadLatency      metric.Float64ObservableUpDownCounter
	WriteLatency     metric.Float64ObservableUpDownCounter
	SyncronizationBW metric.Float64ObservableUpDownCounter
	MirrorBW         metric.Float64ObservableUpDownCounter
	DataRemaining    metric.Float64ObservableUpDownCounter
}

func (mw *MetricsWrapper) initMetrics(prefix, metaID string, labels []attribute.KeyValue) (*Metrics, error) {
	readBW, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "read_bw_megabytes_per_second")

	writeBW, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "write_bw_megabytes_per_second")

	readIOPS, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "read_iops_per_second")

	writeIOPS, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "write_iops_per_second")

	readLatency, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "read_latency_milliseconds")

	writeLatency, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "write_latency_milliseconds")

	syncBW, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "syncronization_bw_megabytes_per_second")

	mirrorBW, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "mirror_bw_megabytes_per_second")

	dataRemaining, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "data_remaining_bytes")

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
func (mw *MetricsWrapper) Record(_ context.Context, meta interface{},
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

	_, _ = mw.Meter.RegisterCallback(func(_ context.Context, obs metric.Observer) error {
		obs.ObserveFloat64(metrics.ReadBW, float64(readBW), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.WriteBW, float64(writeBW), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.ReadIOPS, float64(readIOPS), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.WriteIOPS, float64(writeIOPS), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.ReadLatency, float64(readLatency), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.WriteLatency, float64(writeLatency), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.SyncronizationBW, float64(syncronizationBW), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.MirrorBW, float64(mirrorBW), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.DataRemaining, float64(dataRemaining), metric.ObserveOption(metric.WithAttributes(labels...)))

		return nil
	},
		metrics.ReadBW,
		metrics.WriteBW,
		metrics.ReadIOPS,
		metrics.WriteIOPS,
		metrics.ReadLatency,
		metrics.WriteLatency,
		metrics.SyncronizationBW,
		metrics.MirrorBW,
		metrics.DataRemaining,
	)

	return nil
}

func (mw *MetricsWrapper) initSpaceMetrics(prefix, metaID string, labels []attribute.KeyValue) (*SpaceMetrics, error) {
	logicalProvisioned, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "logical_provisioned_megabytes")

	logicalUsed, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "logical_used_megabytes")

	metrics := &SpaceMetrics{
		LogicalProvisioned: logicalProvisioned,
		LogicalUsed:        logicalUsed,
	}

	mw.SpaceMetrics.Store(metaID, metrics)
	mw.Labels.Store(metaID, labels)

	return metrics, nil
}

// RecordSpaceMetrics will publish space metrics data for a given instance
func (mw *MetricsWrapper) RecordSpaceMetrics(_ context.Context, meta interface{},
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

	_, _ = mw.Meter.RegisterCallback(func(_ context.Context, obs metric.Observer) error {
		obs.ObserveFloat64(metrics.LogicalProvisioned, float64(logicalProvisioned), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.LogicalUsed, float64(logicalUsed), metric.ObserveOption(metric.WithAttributes(labels...)))
		return nil
	},
		metrics.LogicalProvisioned,
		metrics.LogicalUsed,
	)

	return nil
}

func (mw *MetricsWrapper) initArraySpaceMetrics(prefix, metaID string, labels []attribute.KeyValue) (*ArraySpaceMetrics, error) {
	logicalProvisioned, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "logical_provisioned_megabytes")

	logicalUsed, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "logical_used_megabytes")

	metrics := &ArraySpaceMetrics{
		LogicalProvisioned: logicalProvisioned,
		LogicalUsed:        logicalUsed,
	}

	mw.ArraySpaceMetrics.Store(metaID, metrics)
	mw.Labels.Store(metaID, labels)

	return metrics, nil
}

// RecordArraySpaceMetrics will publish space metrics data for a given instance
func (mw *MetricsWrapper) RecordArraySpaceMetrics(_ context.Context, arrayID, driver string,
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
	_, _ = mw.Meter.RegisterCallback(func(_ context.Context, obs metric.Observer) error {
		obs.ObserveFloat64(metrics.LogicalProvisioned, float64(logicalProvisioned), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.LogicalUsed, float64(logicalUsed), metric.ObserveOption(metric.WithAttributes(labels...)))
		return nil
	},
		metrics.LogicalProvisioned,
		metrics.LogicalUsed,
	)

	return nil
}

// RecordStorageClassSpaceMetrics will publish space metrics for storage class
func (mw *MetricsWrapper) RecordStorageClassSpaceMetrics(_ context.Context, storageclass, driver string,
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
	_, _ = mw.Meter.RegisterCallback(func(_ context.Context, obs metric.Observer) error {
		obs.ObserveFloat64(metrics.LogicalProvisioned, float64(logicalProvisioned), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.LogicalUsed, float64(logicalUsed), metric.ObserveOption(metric.WithAttributes(labels...)))
		return nil
	},
		metrics.LogicalProvisioned,
		metrics.LogicalUsed,
	)

	return nil
}

func (mw *MetricsWrapper) initFileSystemMetrics(prefix, metaID string, labels []attribute.KeyValue) (*Metrics, error) {
	readBW, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "read_bw_megabytes_per_second")

	writeBW, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "write_bw_megabytes_per_second")

	readIOPS, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "read_iops_per_second")

	writeIOPS, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "write_iops_per_second")

	readLatency, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "read_latency_milliseconds")

	writeLatency, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "write_latency_milliseconds")

	syncBW, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "syncronization_bw_megabytes_per_second")

	mirrorBW, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "mirror_bw_megabytes_per_second")

	dataRemaining, _ := mw.Meter.Float64ObservableUpDownCounter(prefix + "data_remaining_bytes")

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
func (mw *MetricsWrapper) RecordFileSystemMetrics(_ context.Context, meta interface{},
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

	_, _ = mw.Meter.RegisterCallback(func(_ context.Context, obs metric.Observer) error {
		obs.ObserveFloat64(metrics.ReadBW, float64(readBW), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.WriteBW, float64(writeBW), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.ReadIOPS, float64(readIOPS), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.WriteIOPS, float64(writeIOPS), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.ReadLatency, float64(readLatency), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.WriteLatency, float64(writeLatency), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.SyncronizationBW, float64(syncBW), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.MirrorBW, float64(mirrorBW), metric.ObserveOption(metric.WithAttributes(labels...)))
		obs.ObserveFloat64(metrics.DataRemaining, float64(dataRemaining), metric.ObserveOption(metric.WithAttributes(labels...)))
		return nil
	},
		metrics.ReadBW,
		metrics.WriteBW,
		metrics.ReadIOPS,
		metrics.WriteIOPS,
		metrics.ReadLatency,
		metrics.WriteLatency,
		metrics.SyncronizationBW,
		metrics.MirrorBW,
		metrics.DataRemaining,
	)

	return nil
}
