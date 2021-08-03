// Copyright (c) 2021 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

package service

import (
	"context"
	"errors"
	"strings"
	"sync"

	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/metric"
)

// MetricsRecorder supports recording I/O metrics
//go:generate mockgen -destination=mocks/metrics_mocks.go -package=mocks github.com/dell/csm-metrics-powerstore/internal/service MetricsRecorder,Float64UpDownCounterCreater
type MetricsRecorder interface {
	Record(ctx context.Context, meta interface{},
		readBW, writeBW,
		readIOPS, writeIOPS,
		readLatency, writeLatency float32) error
	RecordSpaceMetrics(ctx context.Context, meta interface{},
		logicalProvisioned, logicalUsed int64) error
	RecordArraySpaceMetrics(ctx context.Context, arrayID, driver string,
		logicalProvisioned, logicalUsed int64) error
	RecordStorageClassSpaceMetrics(ctx context.Context, storageclass, driver string,
		logicalProvisioned, logicalUsed int64) error
	RecordFileSystemMetrics(ctx context.Context, meta interface{},
		readBW, writeBW,
		readIOPS, writeIOPS,
		readLatency, writeLatency float32) error
}

// Float64UpDownCounterCreater creates a Float64UpDownCounter metric
type Float64UpDownCounterCreater interface {
	NewFloat64UpDownCounter(name string, options ...metric.InstrumentOption) (metric.Float64UpDownCounter, error)
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
	LogicalProvisioned metric.BoundFloat64UpDownCounter
	LogicalUsed        metric.BoundFloat64UpDownCounter
}

// ArraySpaceMetrics contains the metrics related to a capacity
type ArraySpaceMetrics struct {
	LogicalProvisioned metric.BoundFloat64UpDownCounter
	LogicalUsed        metric.BoundFloat64UpDownCounter
}

// Metrics contains the list of metrics data that is collected
type Metrics struct {
	ReadBW       metric.BoundFloat64UpDownCounter
	WriteBW      metric.BoundFloat64UpDownCounter
	ReadIOPS     metric.BoundFloat64UpDownCounter
	WriteIOPS    metric.BoundFloat64UpDownCounter
	ReadLatency  metric.BoundFloat64UpDownCounter
	WriteLatency metric.BoundFloat64UpDownCounter
}

func (mw *MetricsWrapper) initMetrics(prefix, metaID string, labels []kv.KeyValue) (*Metrics, error) {
	unboundReadBW, err := mw.Meter.NewFloat64UpDownCounter(prefix + "read_bw_megabytes_per_second")
	if err != nil {
		return nil, err
	}
	readBW := unboundReadBW.Bind(labels...)

	unboundWriteBW, err := mw.Meter.NewFloat64UpDownCounter(prefix + "write_bw_megabytes_per_second")
	if err != nil {
		return nil, err
	}
	writeBW := unboundWriteBW.Bind(labels...)

	unboundReadIOPS, err := mw.Meter.NewFloat64UpDownCounter(prefix + "read_iops_per_second")
	if err != nil {
		return nil, err
	}
	readIOPS := unboundReadIOPS.Bind(labels...)

	unboundWriteIOPS, err := mw.Meter.NewFloat64UpDownCounter(prefix + "write_iops_per_second")
	if err != nil {
		return nil, err
	}
	writeIOPS := unboundWriteIOPS.Bind(labels...)

	unboundReadLatency, err := mw.Meter.NewFloat64UpDownCounter(prefix + "read_latency_milliseconds")
	if err != nil {
		return nil, err
	}
	readLatency := unboundReadLatency.Bind(labels...)

	unboundWriteLatency, err := mw.Meter.NewFloat64UpDownCounter(prefix + "write_latency_milliseconds")
	if err != nil {
		return nil, err
	}
	writeLatency := unboundWriteLatency.Bind(labels...)

	metrics := &Metrics{
		ReadBW:       readBW,
		WriteBW:      writeBW,
		ReadIOPS:     readIOPS,
		WriteIOPS:    writeIOPS,
		ReadLatency:  readLatency,
		WriteLatency: writeLatency,
	}

	mw.Metrics.Store(metaID, metrics)
	mw.Labels.Store(metaID, labels)

	return metrics, nil
}

// Record will publish metrics data for a given instance
func (mw *MetricsWrapper) Record(ctx context.Context, meta interface{},
	readBW, writeBW,
	readIOPS, writeIOPS,
	readLatency, writeLatency float32,
) error {

	var prefix string
	var metaID string
	var labels []kv.KeyValue
	switch v := meta.(type) {
	case *VolumeMeta:
		prefix, metaID = "powerstore_volume_", v.ID
		labels = []kv.KeyValue{
			kv.String("VolumeID", v.ID),
			kv.String("ArrayID", v.ArrayID),
			kv.String("PersistentVolumeName", v.PersistentVolumeName),
			kv.String("PlotWithMean", "No"),
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
			currentLabels := currentLabels.([]kv.KeyValue)
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

	metrics.ReadBW.Add(ctx, float64(readBW))
	metrics.WriteBW.Add(ctx, float64(writeBW))
	metrics.ReadIOPS.Add(ctx, float64(readIOPS))
	metrics.WriteIOPS.Add(ctx, float64(writeIOPS))
	metrics.ReadLatency.Add(ctx, float64(readLatency))
	metrics.WriteLatency.Add(ctx, float64(writeLatency))

	return nil
}

func (mw *MetricsWrapper) initSpaceMetrics(prefix, metaID string, labels []kv.KeyValue) (*SpaceMetrics, error) {

	unboundLogicalProvisioned, err := mw.Meter.NewFloat64UpDownCounter(prefix + "logical_provisioned_megabytes")
	if err != nil {
		return nil, err
	}
	logicalProvisioned := unboundLogicalProvisioned.Bind(labels...)

	unboundLogicalUsed, err := mw.Meter.NewFloat64UpDownCounter(prefix + "logical_used_megabytes")
	if err != nil {
		return nil, err
	}
	logicalUsed := unboundLogicalUsed.Bind(labels...)

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
	logicalProvisioned, logicalUsed int64) error {
	var prefix string
	var metaID string
	var labels []kv.KeyValue
	switch v := meta.(type) {
	case *SpaceVolumeMeta:
		if strings.EqualFold(v.Protocol, scsiProtocol) {
			prefix, metaID = "powerstore_volume_", v.ID
			labels = []kv.KeyValue{
				kv.String("VolumeID", v.ID),
				kv.String("ArrayID", v.ArrayID),
				kv.String("PersistentVolumeName", v.PersistentVolumeName),
				kv.String("StorageClass", v.StorageClass),
				kv.String("Driver", v.Driver),
				kv.String("Protocol", v.Protocol),
				kv.String("PlotWithMean", "No"),
			}
		}
		if strings.EqualFold(v.Protocol, nfsProtocol) {
			prefix, metaID = "powerstore_filesystem_", v.ID
			labels = []kv.KeyValue{
				kv.String("FileSystemID", v.ID),
				kv.String("ArrayID", v.ArrayID),
				kv.String("PersistentVolumeName", v.PersistentVolumeName),
				kv.String("StorageClass", v.StorageClass),
				kv.String("Driver", v.Driver),
				kv.String("Protocol", v.Protocol),
				kv.String("PlotWithMean", "No"),
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
			currentLabels := currentLabels.([]kv.KeyValue)
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
	metrics.LogicalProvisioned.Add(ctx, float64(logicalProvisioned))
	metrics.LogicalUsed.Add(ctx, float64(logicalUsed))
	return nil
}

func (mw *MetricsWrapper) initArraySpaceMetrics(prefix, metaID string, labels []kv.KeyValue) (*ArraySpaceMetrics, error) {

	unboundLogicalProvisioned, err := mw.Meter.NewFloat64UpDownCounter(prefix + "logical_provisioned_megabytes")
	if err != nil {
		return nil, err
	}
	logicalProvisioned := unboundLogicalProvisioned.Bind(labels...)

	unboundLogicalUsed, err := mw.Meter.NewFloat64UpDownCounter(prefix + "logical_used_megabytes")
	if err != nil {
		return nil, err
	}
	logicalUsed := unboundLogicalUsed.Bind(labels...)

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
	logicalProvisioned, logicalUsed int64) error {
	var prefix string
	var metaID string
	var labels []kv.KeyValue

	prefix, metaID = "powerstore_array_", arrayID
	labels = []kv.KeyValue{
		kv.String("ArrayID", arrayID),
		kv.String("Driver", driver),
		kv.String("PlotWithMean", "No"),
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
			currentLabels := currentLabels.([]kv.KeyValue)
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
	metrics.LogicalProvisioned.Add(ctx, float64(logicalProvisioned))
	metrics.LogicalUsed.Add(ctx, float64(logicalUsed))

	return nil
}

// RecordStorageClassSpaceMetrics will publish space metrics for storage class
func (mw *MetricsWrapper) RecordStorageClassSpaceMetrics(ctx context.Context, storageclass, driver string,
	logicalProvisioned, logicalUsed int64) error {
	var prefix string
	var metaID string
	var labels []kv.KeyValue

	prefix, metaID = "powerstore_storage_class_", storageclass
	labels = []kv.KeyValue{
		kv.String("StorageClass", storageclass),
		kv.String("Driver", driver),
		kv.String("PlotWithMean", "No"),
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
			currentLabels := currentLabels.([]kv.KeyValue)
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
	metrics.LogicalProvisioned.Add(ctx, float64(logicalProvisioned))
	metrics.LogicalUsed.Add(ctx, float64(logicalUsed))

	return nil
}

func (mw *MetricsWrapper) initFileSystemMetrics(prefix, metaID string, labels []kv.KeyValue) (*Metrics, error) {
	unboundReadBW, err := mw.Meter.NewFloat64UpDownCounter(prefix + "read_bw_megabytes_per_second")
	if err != nil {
		return nil, err
	}
	readBW := unboundReadBW.Bind(labels...)

	unboundWriteBW, err := mw.Meter.NewFloat64UpDownCounter(prefix + "write_bw_megabytes_per_second")
	if err != nil {
		return nil, err
	}
	writeBW := unboundWriteBW.Bind(labels...)

	unboundReadIOPS, err := mw.Meter.NewFloat64UpDownCounter(prefix + "read_iops_per_second")
	if err != nil {
		return nil, err
	}
	readIOPS := unboundReadIOPS.Bind(labels...)

	unboundWriteIOPS, err := mw.Meter.NewFloat64UpDownCounter(prefix + "write_iops_per_second")
	if err != nil {
		return nil, err
	}
	writeIOPS := unboundWriteIOPS.Bind(labels...)

	unboundReadLatency, err := mw.Meter.NewFloat64UpDownCounter(prefix + "read_latency_milliseconds")
	if err != nil {
		return nil, err
	}
	readLatency := unboundReadLatency.Bind(labels...)

	unboundWriteLatency, err := mw.Meter.NewFloat64UpDownCounter(prefix + "write_latency_milliseconds")
	if err != nil {
		return nil, err
	}
	writeLatency := unboundWriteLatency.Bind(labels...)

	metrics := &Metrics{
		ReadBW:       readBW,
		WriteBW:      writeBW,
		ReadIOPS:     readIOPS,
		WriteIOPS:    writeIOPS,
		ReadLatency:  readLatency,
		WriteLatency: writeLatency,
	}

	mw.Metrics.Store(metaID, metrics)
	mw.Labels.Store(metaID, labels)

	return metrics, nil
}

// RecordFileSystemMetrics will publish filesystem metrics data for a given instance
func (mw *MetricsWrapper) RecordFileSystemMetrics(ctx context.Context, meta interface{},
	readBW, writeBW,
	readIOPS, writeIOPS,
	readLatency, writeLatency float32,
) error {

	var prefix string
	var metaID string
	var labels []kv.KeyValue
	switch v := meta.(type) {
	case *VolumeMeta:
		prefix, metaID = "powerstore_filesystem_", v.ID
		labels = []kv.KeyValue{
			kv.String("FileSystemID", v.ID),
			kv.String("ArrayID", v.ArrayID),
			kv.String("PersistentVolumeName", v.PersistentVolumeName),
			kv.String("StorageClass", v.StorageClass),
			kv.String("PlotWithMean", "No"),
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
			currentLabels := currentLabels.([]kv.KeyValue)
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

	metrics.ReadBW.Add(ctx, float64(readBW))
	metrics.WriteBW.Add(ctx, float64(writeBW))
	metrics.ReadIOPS.Add(ctx, float64(readIOPS))
	metrics.WriteIOPS.Add(ctx, float64(writeIOPS))
	metrics.ReadLatency.Add(ctx, float64(readLatency))
	metrics.WriteLatency.Add(ctx, float64(writeLatency))

	return nil
}
