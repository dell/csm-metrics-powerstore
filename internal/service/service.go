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

	tracer "github.com/dell/csm-metrics-powerstore/opentelemetry/tracers"

	"github.com/dell/csi-powerstore/pkg/array"
	"github.com/dell/gopowerstore"

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
	ExportSpaceVolumeMetrics(context.Context)
	ExportArraySpaceMetrics(context.Context)
}

// PowerStoreClient contains operations for accessing the PowerStore API
//go:generate mockgen -destination=mocks/powerstore_client.go -package=mocks github.com/dell/csm-metrics-powerstore/internal/service PowerStoreClient
type PowerStoreClient interface {
	PerformanceMetricsByVolume(context.Context, string, gopowerstore.MetricsIntervalEnum) ([]gopowerstore.PerformanceMetricsByVolumeResponse, error)
	SpaceMetricsByVolume(context.Context, string, gopowerstore.MetricsIntervalEnum) ([]gopowerstore.SpaceMetricsByVolumeResponse, error)
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

// VolumeSpaceMetricsRecord used for holding output of the Volume space metrics query results
type VolumeSpaceMetricsRecord struct {
	volumeMeta                      *VolumeMeta
	logicalProvisioned, logicalUsed int64
	maxThinSavings, thinSavings     float32
}

// ArraySpaceMetricsRecord used for holding output of the Volume space metrics query results
type ArraySpaceMetricsRecord struct {
	arrayID, storageclass           string
	logicalProvisioned, logicalUsed int64
}

// StorageClassSpaceMetricsRecord used for holding output of the Volume space metrics query results
type StorageClassSpaceMetricsRecord struct {
	storageclass                    string
	logicalProvisioned, logicalUsed int64
}

// ExportVolumeStatistics records I/O statistics for the given list of Volumes
func (s *PowerStoreService) ExportVolumeStatistics(ctx context.Context) {
	ctx, span := tracer.GetTracer(ctx, "ExportVolumeStatistics")
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
		_, span := tracer.GetTracer(ctx, "volumeServer")
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
		ctx, span := tracer.GetTracer(ctx, "gatherVolumeMetrics")
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
				arrayID := volumeProperties[1]
				protocol := volumeProperties[2]

				// skip Persistent Volumes that don't have a protocol of 'scsi', such as nfs file systems
				if !strings.EqualFold(protocol, scsiProtocol) {
					s.Logger.WithFields(logrus.Fields{"protocol": protocol, "persistent_volume": volume.PersistentVolume}).Debugf("persistent volume is not %s", scsiProtocol)
					return
				}

				volumeMeta := &VolumeMeta{
					ID:                   volumeID,
					PersistentVolumeName: volume.PersistentVolume,
					ArrayID:              arrayID,
				}

				goPowerStoreClient, err := s.getPowerStoreClient(ctx, arrayID)
				if err != nil {
					s.Logger.WithError(err).WithField("ip", arrayID).Warn("no client found for PowerStore with IP")
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
					"array_ip":                   volumeMeta.ArrayID,
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
		ctx, span := tracer.GetTracer(ctx, "pushVolumeMetrics")
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

func toMegabytesInt64(bytes int64) int64 {
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

// gatherSpaceVolumeMetrics will return a channel of volume metrics based on the input of volumes
func (s *PowerStoreService) gatherSpaceVolumeMetrics(ctx context.Context, volumes <-chan k8s.VolumeInfo) <-chan *VolumeSpaceMetricsRecord {
	start := time.Now()
	defer s.timeSince(start, "gatherSpaceVolumeMetrics")

	ch := make(chan *VolumeSpaceMetricsRecord)
	var wg sync.WaitGroup
	sem := make(chan struct{}, s.MaxPowerStoreConnections)

	go func() {
		ctx, span := tracer.GetTracer(ctx, "gatherSpaceVolumeMetrics")
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
				arrayID := volumeProperties[1]
				protocol := volumeProperties[2]

				// skip Persistent Volumes that don't have a protocol of 'scsi', such as nfs file systems
				if !strings.EqualFold(protocol, scsiProtocol) {
					s.Logger.WithFields(logrus.Fields{"protocol": protocol, "persistent_volume": volume.PersistentVolume}).Debugf("persistent volume is not %s", scsiProtocol)
					return
				}

				volumeMeta := &VolumeMeta{
					ID:                   volumeID,
					PersistentVolumeName: volume.PersistentVolume,
					ArrayID:              arrayID,
					StorageClass:         volume.StorageClass,
				}

				goPowerStoreClient, err := s.getPowerStoreClient(ctx, arrayID)
				if err != nil {
					s.Logger.WithError(err).WithField("ip", arrayID).Warn("no client found for PowerStore with IP")
					return
				}

				metrics, err := goPowerStoreClient.SpaceMetricsByVolume(ctx, volumeID, gopowerstore.FiveMins)
				if err != nil {
					s.Logger.WithError(err).WithField("volume_id", volumeMeta.ID).Error("getting space metrics for volume")
					return
				}

				var logicalProvisioned, logicalUsed int64
				var maxThinSavings, thinSavings float32

				s.Logger.WithFields(logrus.Fields{
					"volume_space_metrics": len(metrics),
					"volume_id":            volumeMeta.ID,
					"array_id":             volumeMeta.ArrayID,
				}).Debug("volume space metrics returned for volume")

				if len(metrics) > 0 {
					latestMetric := metrics[len(metrics)-1]
					logicalProvisioned = toMegabytesInt64(*latestMetric.LogicalProvisioned)
					logicalUsed = toMegabytesInt64(*latestMetric.LogicalUsed)
					thinSavings = toMegabytes(latestMetric.ThinSavings)
					maxThinSavings = toMegabytes(latestMetric.MaxThinSavings)
				}

				s.Logger.WithFields(logrus.Fields{
					"volume_meta":         volumeMeta,
					"logical_provisioned": logicalProvisioned,
					"logical_used":        logicalUsed,
					"max_thin_savings":    maxThinSavings,
					"thin_savings":        thinSavings,
				}).Debug("volume space metrics")

				ch <- &VolumeSpaceMetricsRecord{
					volumeMeta:         volumeMeta,
					logicalProvisioned: logicalProvisioned,
					logicalUsed:        logicalUsed,
					maxThinSavings:     maxThinSavings,
					thinSavings:        thinSavings,
				}

			}(volume)
		}

		if !exported {
			// If no volumes metrics were exported, we need to export an "empty" metric to update the OT Collector
			// so that stale entries are removed
			ch <- &VolumeSpaceMetricsRecord{
				volumeMeta:         &VolumeMeta{},
				logicalProvisioned: 0,
				logicalUsed:        0,
				maxThinSavings:     0,
				thinSavings:        0,
			}
		}
		wg.Wait()
		close(ch)
		close(sem)
	}()
	return ch
}

// pushSpaceVolumeMetrics will push the provided channel of volume metrics to a data collector
func (s *PowerStoreService) pushSpaceVolumeMetrics(ctx context.Context, volumeSpaceMetrics <-chan *VolumeSpaceMetricsRecord) <-chan string {
	start := time.Now()
	defer s.timeSince(start, "pushSpaceVolumeMetrics")
	var wg sync.WaitGroup

	ch := make(chan string)
	go func() {
		ctx, span := tracer.GetTracer(ctx, "pushSpaceVolumeMetrics")
		defer span.End()

		for metrics := range volumeSpaceMetrics {
			wg.Add(1)
			go func(metrics *VolumeSpaceMetricsRecord) {
				defer wg.Done()
				err := s.MetricsWrapper.RecordSpaceMetrics(ctx,
					metrics.volumeMeta,
					metrics.logicalProvisioned,
					metrics.logicalUsed,
					metrics.maxThinSavings,
					metrics.thinSavings,
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

// ExportSpaceVolumeMetrics records I/O statistics for the given list of Volumes
func (s *PowerStoreService) ExportSpaceVolumeMetrics(ctx context.Context) {
	ctx, span := tracer.GetTracer(ctx, "ExportSpaceVolumeMetrics")
	defer span.End()

	start := time.Now()
	defer s.timeSince(start, "ExportSpaceVolumeMetrics")

	if s.MetricsWrapper == nil {
		s.Logger.Warn("no MetricsWrapper provided for getting ExportSpaceVolumeMetrics")
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

	for range s.pushSpaceVolumeMetrics(ctx, s.gatherSpaceVolumeMetrics(ctx, s.volumeServer(ctx, pvs))) {
		// consume the channel until it is empty and closed
	}
}

// gatherArraySpaceMetrics will return a channel of volume metrics based on the input of volumes
func (s *PowerStoreService) gatherArraySpaceMetrics(ctx context.Context, volumes <-chan k8s.VolumeInfo) <-chan *ArraySpaceMetricsRecord {
	start := time.Now()
	defer s.timeSince(start, "gatherArraySpaceMetrics")

	ch := make(chan *ArraySpaceMetricsRecord)
	var wg sync.WaitGroup
	sem := make(chan struct{}, s.MaxPowerStoreConnections)

	go func() {
		ctx, span := tracer.GetTracer(ctx, "gatherArraySpaceMetrics")
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
				arrayID := volumeProperties[1]
				protocol := volumeProperties[2]

				// skip Persistent Volumes that don't have a protocol of 'scsi', such as nfs file systems
				if !strings.EqualFold(protocol, scsiProtocol) {
					s.Logger.WithFields(logrus.Fields{"protocol": protocol, "persistent_volume": volume.PersistentVolume}).Debugf("persistent volume is not %s", scsiProtocol)
					return
				}

				goPowerStoreClient, err := s.getPowerStoreClient(ctx, arrayID)
				if err != nil {
					s.Logger.WithError(err).WithField("ip", arrayID).Warn("no client found for PowerStore with IP")
					return
				}

				metrics, err := goPowerStoreClient.SpaceMetricsByVolume(ctx, volumeID, gopowerstore.FiveMins)
				if err != nil {
					s.Logger.WithError(err).WithField("volume_id", volumeID).Error("getting space metrics for volume")
				}

				var logicalProvisioned, logicalUsed int64

				if len(metrics) > 0 {
					latestMetric := metrics[len(metrics)-1]
					logicalProvisioned = toMegabytesInt64(*latestMetric.LogicalProvisioned)
					logicalUsed = toMegabytesInt64(*latestMetric.LogicalUsed)
				}
				ch <- &ArraySpaceMetricsRecord{
					arrayID:            arrayID,
					storageclass:       volume.StorageClass,
					logicalProvisioned: logicalProvisioned,
					logicalUsed:        logicalUsed,
				}
				s.Logger.WithFields(logrus.Fields{
					"array_id":                  arrayID,
					"storageClass":              volume.StorageClass,
					"array_logical_provisioned": logicalProvisioned,
					"array_logical_used":        logicalUsed,
				}).Debug("array space metrics for array")
			}(volume)
		}

		if !exported {
			// If no volumes metrics were exported, we need to export an "empty" metric to update the OT Collector
			// so that stale entries are removed
			ch <- &ArraySpaceMetricsRecord{
				arrayID:            "",
				storageclass:       "",
				logicalProvisioned: 0,
				logicalUsed:        0,
			}
		}
		wg.Wait()
		close(ch)
		close(sem)
	}()
	return ch
}

// pushArraySpaceMetrics will push the provided channel of volume metrics to a data collector
func (s *PowerStoreService) pushArraySpaceMetrics(ctx context.Context, volumeSpaceMetrics <-chan *ArraySpaceMetricsRecord) <-chan string {
	start := time.Now()
	defer s.timeSince(start, "pushArraySpaceMetrics")
	var wg sync.WaitGroup

	ch := make(chan string)
	go func() {
		ctx, span := tracer.GetTracer(ctx, "pushArraySpaceMetrics")
		defer span.End()

		// Sum based on array id for total space metrics for array and storage class
		arrayIdMap := make(map[string]ArraySpaceMetricsRecord)
		storageClassMap := make(map[string]ArraySpaceMetricsRecord)

		for metrics := range volumeSpaceMetrics {

			// for array id cumulative
			if volMetrics, ok := arrayIdMap[metrics.arrayID]; !ok {
				arrayIdMap[metrics.arrayID] = ArraySpaceMetricsRecord{
					arrayID:            metrics.arrayID,
					logicalProvisioned: metrics.logicalProvisioned,
					logicalUsed:        metrics.logicalUsed,
				}
			} else {
				volMetrics.logicalProvisioned = volMetrics.logicalProvisioned + metrics.logicalProvisioned
				volMetrics.logicalUsed = volMetrics.logicalUsed + metrics.logicalUsed
				arrayIdMap[metrics.arrayID] = volMetrics
				s.Logger.WithFields(logrus.Fields{
					"array_id":                             metrics.arrayID,
					"cumulative_array_logical_provisioned": volMetrics.logicalProvisioned,
					"cumulative_array_logical_used":        volMetrics.logicalUsed,
				}).Debug("array cumulative space metrics")
			}

			// for Storage Class cumulative space
			if volMetrics, ok := storageClassMap[metrics.storageclass]; !ok {
				storageClassMap[metrics.storageclass] = ArraySpaceMetricsRecord{
					storageclass:       metrics.storageclass,
					logicalProvisioned: metrics.logicalProvisioned,
					logicalUsed:        metrics.logicalUsed,
				}
			} else {
				volMetrics.logicalProvisioned = volMetrics.logicalProvisioned + metrics.logicalProvisioned
				volMetrics.logicalUsed = volMetrics.logicalUsed + metrics.logicalUsed
				storageClassMap[metrics.storageclass] = volMetrics
				s.Logger.WithFields(logrus.Fields{
					"storage_class": metrics.storageclass,
					"cumulative_storage_class_logical_provisioned": volMetrics.logicalProvisioned,
					"cumulative_storage_class_logical_used":        volMetrics.logicalUsed,
				}).Debug("storage class cumulative space metrics")
			}
		}

		// for array id
		for _, metrics := range arrayIdMap {
			wg.Add(1)
			go func(metrics ArraySpaceMetricsRecord) {
				defer wg.Done()
				err := s.MetricsWrapper.RecordArraySpaceMetrics(ctx,
					metrics.arrayID,
					metrics.logicalProvisioned,
					metrics.logicalUsed,
				)
				if err != nil {
					s.Logger.WithError(err).WithField("array_id", metrics.arrayID).Error("recording statistics for array")
				} else {
					ch <- fmt.Sprintf(metrics.arrayID)
				}
			}(metrics)
		}

		// for storage class
		for _, metrics := range storageClassMap {
			wg.Add(1)
			go func(metrics ArraySpaceMetricsRecord) {
				defer wg.Done()
				err := s.MetricsWrapper.RecordStorageClassSpaceMetrics(ctx,
					metrics.storageclass,
					metrics.logicalProvisioned,
					metrics.logicalUsed,
				)
				if err != nil {
					s.Logger.WithError(err).WithField("storage_class", metrics.storageclass).Error("recording statistics for storage class")
				} else {
					ch <- fmt.Sprintf(metrics.storageclass)
				}
			}(metrics)
		}
		wg.Wait()
		close(ch)
	}()

	return ch
}

// ExportArraySpaceMetrics records I/O statistics for the given list of Volumes
func (s *PowerStoreService) ExportArraySpaceMetrics(ctx context.Context) {
	ctx, span := tracer.GetTracer(ctx, "ExportArraySpaceMetrics")
	defer span.End()

	start := time.Now()
	defer s.timeSince(start, "ExportArraySpaceMetrics")

	if s.MetricsWrapper == nil {
		s.Logger.Warn("no MetricsWrapper provided for getting ExportArraySpaceMetrics")
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

	for range s.pushArraySpaceMetrics(ctx, s.gatherArraySpaceMetrics(ctx, s.volumeServer(ctx, pvs))) {
		// consume the channel until it is empty and closed
	}
}
