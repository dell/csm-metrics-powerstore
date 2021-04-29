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
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dell/csi-powerstore/pkg/array"
	"github.com/dell/gopowerstore"
	"go.opentelemetry.io/otel/api/global"

	"github.com/dell/csm-metrics-powerstore/internal/k8s"
	"github.com/sirupsen/logrus"
)

var _ Service = (*PowerStoreService)(nil)

const (
	// DefaultMaxPowerStoreConnections is the number of workers that can query powerstore at a time
	DefaultMaxPowerStoreConnections = 10
	// ExpectedVolumeHandleProperties is the number of properties that the VolumeHandle contains
	ExpectedVolumeHandleProperties = 3
	scsiProtocol                   = "scsi"
)

// Service contains operations that would be used to interact with a PowerStore system
//go:generate mockgen -destination=mocks/service_mocks.go -package=mocks github.com/dell/csm-metrics-powerstore/internal/service Service
type Service interface {
	ExportVolumeStatistics(context.Context)
}

// PowerStoreClient contains operations for accessing the PowerStore API
//go:generate mockgen -destination=mocks/powerstore_client.go -package=mocks github.com/dell/csm-metrics-powerstore/internal/service PowerStoreClient
type PowerStoreClient interface {
	PerformanceMetricsByVolume(context.Context, string, gopowerstore.MetricsIntervalEnum) ([]gopowerstore.PerformanceMetricsByVolumeResponse, error)
}

// PowerStoreService represents the service for getting metrics data for a PowerStore system
type PowerStoreService struct {
	MetricsWrapper           MetricsRecorder
	MaxPowerStoreConnections int
	Logger                   *logrus.Logger
	PowerStoreClients        map[string]PowerStoreClient
	DefaultPowerStoreArray   *array.PowerStoreArray
	VolumeFinder             VolumeFinder
}

// VolumeFinder is used to find volume information in kubernetes
//go:generate mockgen -destination=mocks/volume_finder_mocks.go -package=mocks github.com/dell/csm-metrics-powerstore/internal/service VolumeFinder
type VolumeFinder interface {
	GetPersistentVolumes(context.Context) ([]k8s.VolumeInfo, error)
}

// LeaderElector will elect a leader
//go:generate mockgen -destination=mocks/leader_elector_mocks.go -package=mocks github.com/dell/csm-metrics-powerstore/internal/service LeaderElector
type LeaderElector interface {
	InitLeaderElection(string, string) error
	IsLeader() bool
}

// VolumeMetricsRecord used for holding output of the Volume stat query results
type VolumeMetricsRecord struct {
	volumeMeta *VolumeMeta
	readBW, writeBW,
	readIOPS, writeIOPS,
	readLatency, writeLatency float32
}

// ExportVolumeStatistics records I/O statistics for the given list of Volumes
func (s *PowerStoreService) ExportVolumeStatistics(ctx context.Context) {
	tr := global.TraceProvider().Tracer("metrics-powerstore")
	ctx, span := tr.Start(ctx, "ExportVolumeStatistics")
	defer span.End()

	start := time.Now()
	defer s.timeSince(start, "ExportVolumeStatistics")

	if s.MetricsWrapper == nil {
		s.Logger.Warn("no MetricsWrapper provided for getting ExportVolumeStatistics")
		return
	}

	if s.MaxPowerStoreConnections == 0 {
		s.Logger.Debug("Using DefaultMaxPowerStoreConnections")
		s.MaxPowerStoreConnections = DefaultMaxPowerStoreConnections
	}

	pvs, err := s.VolumeFinder.GetPersistentVolumes(ctx)
	if err != nil {
		s.Logger.WithError(err).Error("getting persistent volumes")
		return
	}

	for range s.pushVolumeMetrics(ctx, s.gatherVolumeMetrics(ctx, s.volumeServer(ctx, pvs))) {
		// consume the channel until it is empty and closed
	}
}

// volumeServer will return a channel of volumes that can provide statistics about each volume
func (s *PowerStoreService) volumeServer(ctx context.Context, volumes []k8s.VolumeInfo) <-chan k8s.VolumeInfo {
	volumeChannel := make(chan k8s.VolumeInfo, len(volumes))
	go func() {
		tr := global.TraceProvider().Tracer("metrics-powerstore")
		_, span := tr.Start(ctx, "volumeServer")
		defer span.End()

		for _, volume := range volumes {
			volumeChannel <- volume
		}
		close(volumeChannel)
	}()
	return volumeChannel
}

// gatherVolumeMetrics will return a channel of volume metrics based on the input of volumes
func (s *PowerStoreService) gatherVolumeMetrics(ctx context.Context, volumes <-chan k8s.VolumeInfo) <-chan *VolumeMetricsRecord {
	start := time.Now()
	defer s.timeSince(start, "gatherVolumeMetrics")

	ch := make(chan *VolumeMetricsRecord)
	var wg sync.WaitGroup
	sem := make(chan struct{}, s.MaxPowerStoreConnections)

	go func() {
		tr := global.TraceProvider().Tracer("metrics-powerstore")
		ctx, span := tr.Start(ctx, "gatherVolumeMetrics")
		defer span.End()

		exported := false
		for volume := range volumes {
			exported = true
			wg.Add(1)
			sem <- struct{}{}
			go func(volume k8s.VolumeInfo) {
				defer func() {
					wg.Done()
					<-sem
				}()

				// VolumeHandle is of the format "volume-id/array-ip/protocol"
				volumeProperties := strings.Split(volume.VolumeHandle, "/")
				if len(volumeProperties) != ExpectedVolumeHandleProperties {
					s.Logger.WithField("volume_handle", volume.VolumeHandle).Warn("unable to get Volume ID and Array IP from volume handle")
					return
				}

				volumeID := volumeProperties[0]
				arrayIP := volumeProperties[1]
				protocol := volumeProperties[2]

				// skip Persistent Volumes that don't have a protocol of 'scsi', such as nfs file systems
				if !strings.EqualFold(protocol, scsiProtocol) {
					s.Logger.WithFields(logrus.Fields{"protocol": protocol, "persistent_volume": volume.PersistentVolume}).Debugf("persistent volume is not %s", scsiProtocol)
					return
				}

				volumeMeta := &VolumeMeta{
					ID:                   volumeID,
					PersistentVolumeName: volume.PersistentVolume,
					ArrayIP:              arrayIP,
				}

				goPowerStoreClient, err := s.getPowerStoreClient(ctx, arrayIP)
				if err != nil {
					s.Logger.WithError(err).WithField("ip", arrayIP).Warn("no client found for PowerStore with IP")
					return
				}

				metrics, err := goPowerStoreClient.PerformanceMetricsByVolume(ctx, volumeID, gopowerstore.TwentySec)
				if err != nil {
					s.Logger.WithError(err).WithField("volume_id", volumeMeta.ID).Error("getting performance metrics for volume")
					return
				}

				var readBW, writeBW, readIOPS, writeIOPS, readLatency, writeLatency float32

				s.Logger.WithFields(logrus.Fields{
					"volume_performance_metrics": len(metrics),
					"volume_id":                  volumeMeta.ID,
					"array_ip":                   volumeMeta.ArrayIP,
				}).Debug("volume performance metrics returned for volume")

				if len(metrics) > 0 {
					latestMetric := metrics[len(metrics)-1]
					readBW = toMegabytes(latestMetric.ReadBandwidth)
					readIOPS = latestMetric.ReadIops
					readLatency = toMilliseconds(latestMetric.AvgReadLatency)
					writeBW = toMegabytes(latestMetric.WriteBandwidth)
					writeIOPS = latestMetric.WriteIops
					writeLatency = toMilliseconds(latestMetric.AvgWriteLatency)
				}

				s.Logger.WithFields(logrus.Fields{
					"volume_meta":     volumeMeta,
					"read_bandwidth":  readBW,
					"write_bandwidth": writeBW,
					"read_iops":       readIOPS,
					"write_iops":      writeIOPS,
					"read_latency":    readLatency,
					"write_latency":   writeLatency,
				}).Debug("volume metrics")

				ch <- &VolumeMetricsRecord{
					volumeMeta: volumeMeta,
					readBW:     readBW, writeBW: writeBW,
					readIOPS: readIOPS, writeIOPS: writeIOPS,
					readLatency: readLatency, writeLatency: writeLatency,
				}

			}(volume)
		}

		if !exported {
			// If no volumes metrics were exported, we need to export an "empty" metric to update the OT Collector
			// so that stale entries are removed
			ch <- &VolumeMetricsRecord{
				volumeMeta: &VolumeMeta{},
				readBW:     0, writeBW: 0,
				readIOPS: 0, writeIOPS: 0,
				readLatency: 0, writeLatency: 0,
			}
		}
		wg.Wait()
		close(ch)
		close(sem)
	}()
	return ch
}

// pushVolumeMetrics will push the provided channel of volume metrics to a data collector
func (s *PowerStoreService) pushVolumeMetrics(ctx context.Context, volumeMetrics <-chan *VolumeMetricsRecord) <-chan string {
	start := time.Now()
	defer s.timeSince(start, "pushVolumeMetrics")
	var wg sync.WaitGroup

	ch := make(chan string)
	go func() {
		tr := global.TraceProvider().Tracer("metrics-powerstore")
		ctx, span := tr.Start(ctx, "pushVolumeMetrics")
		defer span.End()

		for metrics := range volumeMetrics {
			wg.Add(1)
			go func(metrics *VolumeMetricsRecord) {
				defer wg.Done()
				err := s.MetricsWrapper.Record(ctx,
					metrics.volumeMeta,
					metrics.readBW, metrics.writeBW,
					metrics.readIOPS, metrics.writeIOPS,
					metrics.readLatency, metrics.writeLatency,
				)
				if err != nil {
					s.Logger.WithError(err).WithField("volume_id", metrics.volumeMeta.ID).Error("recording statistics for volume")
				} else {
					ch <- fmt.Sprintf(metrics.volumeMeta.ID)
				}
			}(metrics)
		}
		wg.Wait()
		close(ch)
	}()

	return ch
}

func (s *PowerStoreService) getPowerStoreClient(ctx context.Context, arrayIP string) (PowerStoreClient, error) {
	if goPowerStoreClient, ok := s.PowerStoreClients[arrayIP]; ok {
		return goPowerStoreClient, nil
	}
	return nil, fmt.Errorf("unable to find client")
}

func toMegabytes(bytes float32) float32 {
	return bytes / 1024 / 1024
}

func toMilliseconds(microseconds float32) float32 {
	return microseconds / 1000
}

// timeSince will log the amount of time spent in a given function
func (s *PowerStoreService) timeSince(start time.Time, fName string) {
	s.Logger.WithFields(logrus.Fields{
		"duration": fmt.Sprintf("%v", time.Since(start)),
		"function": fName,
	}).Info("function duration")
}
