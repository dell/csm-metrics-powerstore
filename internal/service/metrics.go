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
		lastLogicalProvisioned, lastLogicalUsed, logicalProvisioned, logicalUsed, uniquePhysicalUsed int64,
		maxThinSavings, thinSavings float32) error
}

// Float64UpDownCounterCreater creates a Float64UpDownCounter metric
type Float64UpDownCounterCreater interface {
	NewFloat64UpDownCounter(name string, options ...metric.InstrumentOption) (metric.Float64UpDownCounter, error)
}

// MetricsWrapper contains data used for pushing metrics data
type MetricsWrapper struct {
	Meter        Float64UpDownCounterCreater
	Metrics      sync.Map
	Labels       sync.Map
	SpaceMetrics sync.Map
}

// CapacityMetrics contains the metrics related to a capacity
type SpaceMetrics struct {
	LogicalProvisioned     metric.BoundFloat64UpDownCounter
	LogicalUsed            metric.BoundFloat64UpDownCounter
	LastLogicalProvisioned metric.BoundFloat64UpDownCounter
	LastLogicalUsed        metric.BoundFloat64UpDownCounter
	UniquePhysicalUsed     metric.BoundFloat64UpDownCounter
	MaxThinSavings         metric.BoundFloat64UpDownCounter
	ThinSavings            metric.BoundFloat64UpDownCounter
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
			kv.String("ArrayIP", v.ArrayID),
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
	/*
			LogicalProvisioned     metric.BoundFloat64UpDownCounter
		            metric.BoundFloat64UpDownCounter
		 metric.BoundFloat64UpDownCounter
		        metric.BoundFloat64UpDownCounter
		     metric.BoundFloat64UpDownCounter
		MaxThinSavings         metric.BoundFloat64UpDownCounter
		ThinSavings            metric.BoundFloat64UpDownCounter
	*/
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

	unboundLastLogicalProvisioned, err := mw.Meter.NewFloat64UpDownCounter(prefix + "last_logical_provisioned_megabytes")
	if err != nil {
		return nil, err
	}
	lastLogicalProvisioned := unboundLastLogicalProvisioned.Bind(labels...)

	unboundLastLogicalUsed, err := mw.Meter.NewFloat64UpDownCounter(prefix + "last_logical_provision_megabytes")
	if err != nil {
		return nil, err
	}
	lastLogicalUsed := unboundLastLogicalUsed.Bind(labels...)

	unboundUniquePhysicalUsed, err := mw.Meter.NewFloat64UpDownCounter(prefix + "unique_physical_used_megabytes")
	if err != nil {
		return nil, err
	}
	uniquePhysicalUsed := unboundUniquePhysicalUsed.Bind(labels...)

	unboundThinSavings, err := mw.Meter.NewFloat64UpDownCounter(prefix + "thin_savings_megabytes")
	if err != nil {
		return nil, err
	}
	thinSavings := unboundThinSavings.Bind(labels...)

	unboundMaxThinSavings, err := mw.Meter.NewFloat64UpDownCounter(prefix + "max_thin_savings_megabytes")
	if err != nil {
		return nil, err
	}
	maxThinSavings := unboundMaxThinSavings.Bind(labels...)

	metrics := &SpaceMetrics{
		LogicalProvisioned:     logicalProvisioned,
		LogicalUsed:            logicalUsed,
		LastLogicalProvisioned: lastLogicalProvisioned,
		LastLogicalUsed:        lastLogicalUsed,
		UniquePhysicalUsed:     uniquePhysicalUsed,
		MaxThinSavings:         maxThinSavings,
		ThinSavings:            thinSavings,
	}

	mw.SpaceMetrics.Store(metaID, metrics)
	mw.Labels.Store(metaID, labels)

	return metrics, nil
}

// RecordSpaceMetrics will publish space metrics data for a given instance
func (mw *MetricsWrapper) RecordSpaceMetrics(ctx context.Context, meta interface{},
	lastLogicalProvisioned, lastLogicalUsed, logicalProvisioned, logicalUsed, uniquePhysicalUsed int64,
	maxThinSavings, thinSavings float32) error {
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
			kv.String("StorageClass", v.StorageClass),
			kv.String("PlotWithMean", "No"),
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
	metrics.LastLogicalProvisioned.Add(ctx, float64(lastLogicalProvisioned))
	metrics.LastLogicalUsed.Add(ctx, float64(lastLogicalUsed))
	metrics.UniquePhysicalUsed.Add(ctx, float64(uniquePhysicalUsed))
	metrics.ThinSavings.Add(ctx, float64(thinSavings))
	metrics.MaxThinSavings.Add(ctx, float64(maxThinSavings))

	return nil
}
