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
}

// Float64UpDownCounterCreater creates a Float64UpDownCounter metric
type Float64UpDownCounterCreater interface {
	NewFloat64UpDownCounter(name string, options ...metric.InstrumentOption) (metric.Float64UpDownCounter, error)
}

// MetricsWrapper contains data used for pushing metrics data
type MetricsWrapper struct {
	Meter           Float64UpDownCounterCreater
	Metrics         sync.Map
	Labels          sync.Map
	CapacityMetrics sync.Map
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
			kv.String("ArrayIP", v.ArrayIP),
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
