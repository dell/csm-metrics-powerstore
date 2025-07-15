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
	"fmt"
	"strings"
	"sync"
	"time"

	tracer "github.com/dell/csm-metrics-powerstore/opentelemetry/tracers"

	"github.com/dell/gopowerstore"

	"github.com/dell/csm-metrics-powerstore/internal/k8s"
	"github.com/dell/csm-sharednfs/nfs"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultMaxPowerStoreConnections is the number of workers that can query powerstore at a time
	DefaultMaxPowerStoreConnections = 10
	// ExpectedVolumeHandleProperties is the number of properties that the VolumeHandle contains
	ExpectedVolumeHandleProperties = 3
	// ExpectedVolumeHandleMetroProperties is the number of properties that the VolumeHandle of metro volumes contains
	ExpectedVolumeHandleMetroProperties = 4
	scsiProtocol                        = "scsi"
	nfsProtocol                         = "nfs"
)

var _ Service = (*PowerStoreService)(nil)

func (s *PowerStoreService) getPowerStoreClient(_ context.Context, arrayIP string) (PowerStoreClient, error) {
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

// Service contains operations that would be used to interact with a PowerStore system
//
//go:generate mockgen -destination=mocks/service_mocks.go -package=mocks github.com/dell/csm-metrics-powerstore/internal/service Service
type Service interface {
	ExportVolumeStatistics(context.Context)
	ExportSpaceVolumeMetrics(context.Context)
	ExportArraySpaceMetrics(context.Context)
	ExportFileSystemStatistics(context.Context)
	ExportTopologyMetrics(context.Context)
}

// PowerStoreClient contains operations for accessing the PowerStore API
//
//go:generate mockgen -destination=mocks/powerstore_client_mocks.go -package=mocks github.com/dell/csm-metrics-powerstore/internal/service PowerStoreClient
type PowerStoreClient interface {
	PerformanceMetricsByVolume(context.Context, string, gopowerstore.MetricsIntervalEnum) ([]gopowerstore.PerformanceMetricsByVolumeResponse, error)
	SpaceMetricsByVolume(context.Context, string, gopowerstore.MetricsIntervalEnum) ([]gopowerstore.SpaceMetricsByVolumeResponse, error)
	PerformanceMetricsByFileSystem(context.Context, string, gopowerstore.MetricsIntervalEnum) ([]gopowerstore.PerformanceMetricsByFileSystemResponse, error)
	GetFS(context.Context, string) (gopowerstore.FileSystem, error)
	VolumeMirrorTransferRate(ctx context.Context, id string) ([]gopowerstore.VolumeMirrorTransferRateResponse, error)
}

// PowerStoreService represents the service for getting metrics data for a PowerStore system
type PowerStoreService struct {
	MetricsWrapper           MetricsRecorder
	MaxPowerStoreConnections int
	Logger                   *logrus.Logger
	PowerStoreClients        map[string]PowerStoreClient
	DefaultPowerStoreArray   *PowerStoreArray
	VolumeFinder             VolumeFinder
}

// VolumeFinder is used to find volume information in kubernetes
//
//go:generate mockgen -destination=mocks/volume_finder_mocks.go -package=mocks github.com/dell/csm-metrics-powerstore/internal/service VolumeFinder
type VolumeFinder interface {
	GetPersistentVolumes(context.Context) ([]k8s.VolumeInfo, error)
}

// LeaderElector will elect a leader
//
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
	readLatency, writeLatency,
	synchronizationBW, mirrorBW, remainingData float32
}

// VolumeSpaceMetricsRecord used for holding output of the Volume space metrics query results
type VolumeSpaceMetricsRecord struct {
	spaceMeta                       *SpaceVolumeMeta
	logicalProvisioned, logicalUsed int64
}

// ArraySpaceMetricsRecord used for holding output of the Volume space metrics query results
type ArraySpaceMetricsRecord struct {
	arrayID, storageclass, driver   string
	logicalProvisioned, logicalUsed int64
}

// TopologyMetricsRecord used for holding output of the Topology metrics query results
type TopologyMetricsRecord struct {
	TopologyMeta *TopologyMeta
	PVCSize      int64
}

var PrevPVList = make(map[string]bool)
var CurrentPVList map[string]bool

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
	} // revive:disable-line:empty-block
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

	var volumeID, arrayID, protocol string
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
				if len(volumeProperties) == ExpectedVolumeHandleProperties {
					volumeID = volumeProperties[0]
					arrayID = volumeProperties[1]
					protocol = volumeProperties[2]
				} else if len(volumeProperties) == ExpectedVolumeHandleMetroProperties {
					volumeID = volumeProperties[0]
					arrayID = volumeProperties[1]
					protocol = strings.Split(volumeProperties[2], ":")[0]
				} else {
					s.Logger.WithField("volume_handle", volume.VolumeHandle).Warn("unable to get Volume ID and Array IP from volume handle")
					return
				}

				// skip Persistent Volumes that don't have a protocol of 'scsi', such as nfs file systems
				if !strings.EqualFold(protocol, scsiProtocol) {
					s.Logger.WithFields(logrus.Fields{"protocol": protocol, "persistent_volume": volume.PersistentVolume}).Debugf("persistent volume is not %s", scsiProtocol)
					return
				}

				volumeMeta := &VolumeMeta{
					ID:                        volumeID,
					PersistentVolumeName:      volume.PersistentVolume,
					PersistentVolumeClaimName: volume.VolumeClaimName,
					Namespace:                 volume.Namespace,
					ArrayID:                   arrayID,
				}

				goPowerStoreClient, err := s.getPowerStoreClient(ctx, arrayID)
				if err != nil {
					s.Logger.WithError(err).WithField("ip", arrayID).Warn("no client found for PowerStore with IP")
					return
				}

				if nfs.IsNFSVolumeID(volumeID) {
					volumeID = nfs.ToArrayVolumeID(volumeID)
				}

				s.Logger.WithFields(logrus.Fields{
					"volume_id": volumeID,
				}).Infof("Getting the Performance Metrics")

				metrics, err := goPowerStoreClient.PerformanceMetricsByVolume(ctx, volumeID, gopowerstore.TwentySec)
				if err != nil {
					s.Logger.WithError(err).WithField("volume_id", volumeID).Error("getting performance metrics for volume")
					return
				}

				var readBW, writeBW, readIOPS, writeIOPS, readLatency, writeLatency, syncBW, mirrorBW, remainingData float32

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

				// Read the replication parameter
				replicationMetrics, err := goPowerStoreClient.VolumeMirrorTransferRate(ctx, volumeID)

				s.Logger.WithFields(logrus.Fields{
					"volume_replication_metrics": len(replicationMetrics),
					"volume_id":                  volumeMeta.ID,
					"array_ip":                   volumeMeta.ArrayID,
				}).Debug("volume replication metrics returned for volume")

				if len(replicationMetrics) > 0 {
					latestRepMetrics := replicationMetrics[len(replicationMetrics)-1]
					syncBW = toMegabytes(latestRepMetrics.SynchronizationBandwidth)
					mirrorBW = toMegabytes(latestRepMetrics.MirrorBandwidth)
					remainingData = toMegabytes(latestRepMetrics.DataRemaining)
				}

				s.Logger.WithFields(logrus.Fields{
					"volume_meta":              volumeMeta,
					"read_bandwidth":           readBW,
					"write_bandwidth":          writeBW,
					"read_iops":                readIOPS,
					"write_iops":               writeIOPS,
					"read_latency":             readLatency,
					"write_latency":            writeLatency,
					"syncronization_bandwidth": syncBW,
					"mirror_bandwidth":         mirrorBW,
					"remaining_data":           remainingData,
				}).Debug("volume metrics")

				ch <- &VolumeMetricsRecord{
					volumeMeta: volumeMeta,
					readBW:     readBW, writeBW: writeBW,
					readIOPS: readIOPS, writeIOPS: writeIOPS,
					readLatency: readLatency, writeLatency: writeLatency,
					synchronizationBW: syncBW, mirrorBW: mirrorBW, remainingData: remainingData,
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
				synchronizationBW: 0, mirrorBW: 0, remainingData: 0,
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
					metrics.synchronizationBW, metrics.mirrorBW, metrics.remainingData,
				)
				if err != nil {
					s.Logger.WithError(err).WithField("volume_id", metrics.volumeMeta.ID).Error("recording statistics for volume")
				} else {
					ch <- metrics.volumeMeta.ID
				}
			}(metrics)
		}
		wg.Wait()
		close(ch)
	}()

	return ch
}

// gatherSpaceVolumeMetrics will return a channel of space volume metrics based on the input of volumes
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

				var logicalProvisioned, logicalUsed int64

				// VolumeHandle is of the format "volume-id/array-ip/protocol"
				volumeProperties := strings.Split(volume.VolumeHandle, "/")
				if len(volumeProperties) != ExpectedVolumeHandleProperties {
					s.Logger.WithField("volume_handle", volume.VolumeHandle).Warn("unable to get Volume ID and Array IP from volume handle")
					return
				}

				volumeID := volumeProperties[0]
				arrayID := volumeProperties[1]
				protocol := volumeProperties[2]

				spaceMeta := &SpaceVolumeMeta{
					ID:                        volumeID,
					PersistentVolumeName:      volume.PersistentVolume,
					PersistentVolumeClaimName: volume.VolumeClaimName,
					Namespace:                 volume.Namespace,
					ArrayID:                   arrayID,
					StorageClass:              volume.StorageClass,
					Driver:                    volume.Driver,
					Protocol:                  protocol,
				}

				goPowerStoreClient, err := s.getPowerStoreClient(ctx, arrayID)
				if err != nil {
					s.Logger.WithError(err).WithField("ip", arrayID).Warn("no client found for PowerStore with IP")
					return
				}

				switch protocol {
				case nfsProtocol: // nfs space metrics
					fs, err := goPowerStoreClient.GetFS(ctx, volumeID)
					if err != nil {
						s.Logger.WithError(err).WithField("filesystem_id", volumeID).Error("getting space metrics for filesystem")
						return
					}

					logicalProvisioned = toMegabytesInt64(fs.SizeTotal)
					logicalUsed = toMegabytesInt64(fs.SizeUsed)
					s.Logger.WithFields(logrus.Fields{"filesystem": volumeID, "persistent_volume": volume.PersistentVolume}).Debugf("got data %d %d", logicalProvisioned, logicalUsed)

				default: // space metrics for Persistent Volumes
					metrics, err := goPowerStoreClient.SpaceMetricsByVolume(ctx, volumeID, gopowerstore.FiveMins)
					if err != nil {
						s.Logger.WithError(err).WithField("volume_id", spaceMeta.ID).Error("getting space metrics for volume")
						return
					}
					if len(metrics) > 0 {
						latestMetric := metrics[len(metrics)-1]
						logicalProvisioned = toMegabytesInt64(*latestMetric.LogicalProvisioned)
						logicalUsed = toMegabytesInt64(*latestMetric.LogicalUsed)
					}
					s.Logger.WithFields(logrus.Fields{
						"space_metrics": len(metrics),
						"id":            spaceMeta.ID,
						"array_id":      spaceMeta.ArrayID,
						"storage_class": spaceMeta.StorageClass,
					}).Debug("volume space metrics returned for volume")
				}

				s.Logger.WithFields(logrus.Fields{
					"space_meta":          spaceMeta,
					"logical_provisioned": logicalProvisioned,
					"logical_used":        logicalUsed,
				}).Debug("volume space metrics")

				ch <- &VolumeSpaceMetricsRecord{
					spaceMeta:          spaceMeta,
					logicalProvisioned: logicalProvisioned,
					logicalUsed:        logicalUsed,
				}
			}(volume)
		}

		if !exported {
			// If no volumes metrics were exported, we need to export an "empty" metric to update the OT Collector
			// so that stale entries are removed
			ch <- &VolumeSpaceMetricsRecord{
				spaceMeta:          &SpaceVolumeMeta{},
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

// pushSpaceVolumeMetrics will push the provided channel of space metrics to a data collector
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
					metrics.spaceMeta,
					metrics.logicalProvisioned,
					metrics.logicalUsed,
				)
				if err != nil {
					s.Logger.WithError(err).WithField("id", metrics.spaceMeta.ID).Error("recording statistics for volume")
				} else {
					ch <- metrics.spaceMeta.ID
				}
			}(metrics)
		}
		wg.Wait()
		close(ch)
	}()

	return ch
}

// ExportSpaceVolumeMetrics records space capacity for the given list of Volumes
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
	} // revive:disable-line:empty-block
}

// gatherArraySpaceMetrics will return a channel of array space metrics based on the input of volumes
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

				goPowerStoreClient, err := s.getPowerStoreClient(ctx, arrayID)
				if err != nil {
					s.Logger.WithError(err).WithField("ip", arrayID).Warn("no client found for PowerStore with IP")
					return
				}

				var logicalProvisioned, logicalUsed int64
				switch protocol {
				// filesystem: nfs protocol space metrics
				case nfsProtocol:
					fs, err := goPowerStoreClient.GetFS(ctx, volumeID)
					if err != nil {
						s.Logger.WithError(err).WithField("filesystem_id", volumeID).Error("getting space metrics for filesystem")
						return
					}

					logicalProvisioned = toMegabytesInt64(fs.SizeTotal)
					logicalUsed = toMegabytesInt64(fs.SizeUsed)
					s.Logger.WithFields(logrus.Fields{"filesystem": volumeID, "persistent_volume": volume.PersistentVolume}).Debugf("got data %d %d", logicalProvisioned, logicalUsed)

					// volume space metrics: scsi as default protocol
				default:
					metrics, err := goPowerStoreClient.SpaceMetricsByVolume(ctx, volumeID, gopowerstore.FiveMins)
					if err != nil {
						s.Logger.WithError(err).WithField("volume_id", volumeID).Error("getting space metrics for volume")
						return
					}
					if len(metrics) > 0 {
						latestMetric := metrics[len(metrics)-1]
						logicalProvisioned = toMegabytesInt64(*latestMetric.LogicalProvisioned)
						logicalUsed = toMegabytesInt64(*latestMetric.LogicalUsed)
					}
				}

				ch <- &ArraySpaceMetricsRecord{
					arrayID:            arrayID,
					storageclass:       volume.StorageClass,
					driver:             volume.Driver,
					logicalProvisioned: logicalProvisioned,
					logicalUsed:        logicalUsed,
				}
				s.Logger.WithFields(logrus.Fields{
					"array_id":                  arrayID,
					"storageClass":              volume.StorageClass,
					"driver":                    volume.Driver,
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

// pushArraySpaceMetrics will push the provided channel of array and storageclass space metrics to a data collector
func (s *PowerStoreService) pushArraySpaceMetrics(ctx context.Context, volumeSpaceMetrics <-chan *ArraySpaceMetricsRecord) <-chan string {
	start := time.Now()
	defer s.timeSince(start, "pushArraySpaceMetrics")
	var wg sync.WaitGroup

	ch := make(chan string)
	go func() {
		ctx, span := tracer.GetTracer(ctx, "pushArraySpaceMetrics")
		defer span.End()

		// Sum based on array id for total space metrics for array and storage class
		arrayIDMap := make(map[string]ArraySpaceMetricsRecord)
		storageClassMap := make(map[string]ArraySpaceMetricsRecord)

		for metrics := range volumeSpaceMetrics {

			// for array id cumulative
			if volMetrics, ok := arrayIDMap[metrics.arrayID]; !ok {
				arrayIDMap[metrics.arrayID] = ArraySpaceMetricsRecord{
					arrayID:            metrics.arrayID,
					driver:             metrics.driver,
					logicalProvisioned: metrics.logicalProvisioned,
					logicalUsed:        metrics.logicalUsed,
				}
			} else {
				volMetrics.logicalProvisioned = volMetrics.logicalProvisioned + metrics.logicalProvisioned
				volMetrics.logicalUsed = volMetrics.logicalUsed + metrics.logicalUsed
				arrayIDMap[metrics.arrayID] = volMetrics
				s.Logger.WithFields(logrus.Fields{
					"array_id":                             metrics.arrayID,
					"driver":                               metrics.driver,
					"cumulative_array_logical_provisioned": volMetrics.logicalProvisioned,
					"cumulative_array_logical_used":        volMetrics.logicalUsed,
				}).Debug("array cumulative space metrics")
			}

			// for Storage Class cumulative space
			if volMetrics, ok := storageClassMap[metrics.storageclass]; !ok {
				storageClassMap[metrics.storageclass] = ArraySpaceMetricsRecord{
					storageclass:       metrics.storageclass,
					driver:             metrics.driver,
					logicalProvisioned: metrics.logicalProvisioned,
					logicalUsed:        metrics.logicalUsed,
				}
			} else {
				volMetrics.logicalProvisioned = volMetrics.logicalProvisioned + metrics.logicalProvisioned
				volMetrics.logicalUsed = volMetrics.logicalUsed + metrics.logicalUsed
				storageClassMap[metrics.storageclass] = volMetrics
				s.Logger.WithFields(logrus.Fields{
					"storage_class": metrics.storageclass,
					"driver":        metrics.driver,
					"cumulative_storage_class_logical_provisioned": volMetrics.logicalProvisioned,
					"cumulative_storage_class_logical_used":        volMetrics.logicalUsed,
				}).Debug("storage class cumulative space metrics")
			}
		}

		// for array id
		for _, metrics := range arrayIDMap {
			wg.Add(1)
			go func(metrics ArraySpaceMetricsRecord) {
				defer wg.Done()
				err := s.MetricsWrapper.RecordArraySpaceMetrics(ctx,
					metrics.arrayID, metrics.driver,
					metrics.logicalProvisioned,
					metrics.logicalUsed,
				)
				if err != nil {
					s.Logger.WithError(err).WithField("array_id", metrics.arrayID).Error("recording statistics for array")
				} else {
					ch <- metrics.arrayID
				}
			}(metrics)
		}

		// for storage class
		for _, metrics := range storageClassMap {
			wg.Add(1)
			go func(metrics ArraySpaceMetricsRecord) {
				defer wg.Done()
				err := s.MetricsWrapper.RecordStorageClassSpaceMetrics(ctx,
					metrics.storageclass, metrics.driver,
					metrics.logicalProvisioned,
					metrics.logicalUsed,
				)
				if err != nil {
					s.Logger.WithError(err).WithField("storage_class", metrics.storageclass).Error("recording statistics for storage class")
				} else {
					ch <- metrics.storageclass
				}
			}(metrics)
		}
		wg.Wait()
		close(ch)
	}()

	return ch
}

// ExportArraySpaceMetrics records space capacity statistics for the given list of arrays
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
	} // revive:disable-line:empty-block
}

// ExportFileSystemStatistics records I/O statistics for the given list of Volumes
func (s *PowerStoreService) ExportFileSystemStatistics(ctx context.Context) {
	ctx, span := tracer.GetTracer(ctx, "ExportFileSystemStatistics")
	defer span.End()

	start := time.Now()
	defer s.timeSince(start, "ExportFileSystemStatistics")

	if s.MetricsWrapper == nil {
		s.Logger.Warn("no MetricsWrapper provided for getting ExportFileSystemStatistics")
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

	for range s.pushFileSystemMetrics(ctx, s.gatherFileSystemMetrics(ctx, s.volumeServer(ctx, pvs))) {
		// consume the channel until it is empty and closed
	} // revive:disable-line:empty-block
}

// gatherFileSystemMetrics will return a channel of filesystem metrics based on the input of volumes
func (s *PowerStoreService) gatherFileSystemMetrics(ctx context.Context, volumes <-chan k8s.VolumeInfo) <-chan *VolumeMetricsRecord {
	start := time.Now()
	defer s.timeSince(start, "gatherFileSystemMetrics")

	ch := make(chan *VolumeMetricsRecord)
	var wg sync.WaitGroup
	sem := make(chan struct{}, s.MaxPowerStoreConnections)

	var volumeID, arrayID, protocol string

	go func() {
		ctx, span := tracer.GetTracer(ctx, "gatherFileSystemMetrics")
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
				if len(volumeProperties) == ExpectedVolumeHandleProperties {
					volumeID = volumeProperties[0]
					arrayID = volumeProperties[1]
					protocol = volumeProperties[2]
					// VolumeHandle is of the format "src-volume-id/array-ip/protocol:dest-volume-id/dest-array-ip"
				} else if len(volumeProperties) == ExpectedVolumeHandleMetroProperties {
					volumeID = volumeProperties[0]
					arrayID = volumeProperties[1]
					protocol = strings.Split(volumeProperties[2], ":")[0]
				} else {
					s.Logger.WithField("volume_handle", volume.VolumeHandle).Warn("unable to get Volume ID and Array IP from volume handle")
					return
				}

				// skip Persistent Volumes that don't have a protocol of 'nfs'
				if !strings.EqualFold(protocol, nfsProtocol) {
					s.Logger.WithFields(logrus.Fields{"protocol": protocol, "persistent_volume": volume.PersistentVolume}).Debugf("persistent volume is not %s", nfsProtocol)
					return
				}

				volumeMeta := &VolumeMeta{
					ID:                        volumeID,
					PersistentVolumeName:      volume.PersistentVolume,
					PersistentVolumeClaimName: volume.VolumeClaimName,
					Namespace:                 volume.Namespace,
					ArrayID:                   arrayID,
					StorageClass:              volume.StorageClass,
				}

				goPowerStoreClient, err := s.getPowerStoreClient(ctx, arrayID)
				if err != nil {
					s.Logger.WithError(err).WithField("ip", arrayID).Warn("no client found for PowerStore with IP")
					return
				}

				metrics, err := goPowerStoreClient.PerformanceMetricsByFileSystem(ctx, volumeID, gopowerstore.TwentySec)
				if err != nil {
					s.Logger.WithError(err).WithField("volume_id", volumeMeta.ID).Error("getting performance metrics for volume")
					return
				}

				var readBW, writeBW, readIOPS, writeIOPS, readLatency, writeLatency, syncBW, mirrorBW, remainingData float32

				s.Logger.WithFields(logrus.Fields{
					"filesystem_performance_metrics": len(metrics),
					"filesystem_id":                  volumeMeta.ID,
					"array_ip":                       volumeMeta.ArrayID,
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

				// Read the replication parameter
				replicationMetrics, err := goPowerStoreClient.VolumeMirrorTransferRate(ctx, volumeID)

				s.Logger.WithFields(logrus.Fields{
					"volume_replication_metrics": len(replicationMetrics),
					"volume_id":                  volumeMeta.ID,
					"array_ip":                   volumeMeta.ArrayID,
				}).Debug("volume replication metrics returned for volume")

				if len(replicationMetrics) > 0 {
					latestRepMetrics := replicationMetrics[len(replicationMetrics)-1]
					syncBW = toMegabytes(latestRepMetrics.SynchronizationBandwidth)
					mirrorBW = toMegabytes(latestRepMetrics.MirrorBandwidth)
					remainingData = toMegabytes(latestRepMetrics.DataRemaining)
				}

				s.Logger.WithFields(logrus.Fields{
					"volume_meta":              volumeMeta,
					"read_bandwidth":           readBW,
					"write_bandwidth":          writeBW,
					"read_iops":                readIOPS,
					"write_iops":               writeIOPS,
					"read_latency":             readLatency,
					"write_latency":            writeLatency,
					"syncronization_bandwidth": syncBW,
					"mirror_bandwidth":         mirrorBW,
					"remaining_data":           remainingData,
				}).Debug("volume metrics")

				ch <- &VolumeMetricsRecord{
					volumeMeta: volumeMeta,
					readBW:     readBW, writeBW: writeBW,
					readIOPS: readIOPS, writeIOPS: writeIOPS,
					readLatency: readLatency, writeLatency: writeLatency,
					synchronizationBW: syncBW, mirrorBW: mirrorBW, remainingData: remainingData,
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
				synchronizationBW: 0, mirrorBW: 0, remainingData: 0,
			}
		}
		wg.Wait()
		close(ch)
		close(sem)
	}()
	return ch
}

// pushFileSystemMetrics will push the provided channel of filesystem metrics to a data collector
func (s *PowerStoreService) pushFileSystemMetrics(ctx context.Context, volumeMetrics <-chan *VolumeMetricsRecord) <-chan string {
	start := time.Now()
	defer s.timeSince(start, "pushFileSystemMetrics")
	var wg sync.WaitGroup

	ch := make(chan string)
	go func() {
		ctx, span := tracer.GetTracer(ctx, "pushFileSystemMetrics")
		defer span.End()

		for metrics := range volumeMetrics {
			wg.Add(1)
			go func(metrics *VolumeMetricsRecord) {
				defer wg.Done()
				err := s.MetricsWrapper.RecordFileSystemMetrics(ctx,
					metrics.volumeMeta,
					metrics.readBW, metrics.writeBW,
					metrics.readIOPS, metrics.writeIOPS,
					metrics.readLatency, metrics.writeLatency,
					metrics.synchronizationBW, metrics.mirrorBW, metrics.remainingData,
				)
				if err != nil {
					s.Logger.WithError(err).WithField("volume_id", metrics.volumeMeta.ID).Error("recording statistics for volume")
				} else {
					ch <- metrics.volumeMeta.ID
				}
			}(metrics)
		}
		wg.Wait()
		close(ch)
	}()

	return ch
}

func (s *PowerStoreService) gatherTopologyMetrics(ctx context.Context, volumes <-chan k8s.VolumeInfo) <-chan *TopologyMetricsRecord {
	start := time.Now()
	defer s.timeSince(start, "gatherTopologyMetrics")

	ch := make(chan *TopologyMetricsRecord)
	var wg sync.WaitGroup

	go func() {
		exported := false
		for volume := range volumes {
			exported = true
			wg.Add(1)

			go func(volume k8s.VolumeInfo) {
				defer wg.Done()

				// VolumeHandle format: "volume-id/array-ip/protocol"
				volumeProperties := strings.Split(volume.VolumeHandle, "/")
				if len(volumeProperties) != ExpectedVolumeHandleProperties {
					s.Logger.WithField("volume_handle", volume.VolumeHandle).Warn("unable to parse volume handle")
					return
				}

				topologyMeta := &TopologyMeta{
					PersistentVolumeClaim:   volume.VolumeClaimName,
					VolumeClaimName:         volume.VolumeClaimName,
					PersistentVolumeStatus:  volume.PersistentVolumeStatus,
					PersistentVolume:        volume.PersistentVolume,
					Driver:                  volume.Driver,
					ProvisionedSize:         volume.ProvisionedSize,
					StorageSystemVolumeName: volume.StorageSystemVolumeName,
					StoragePoolName:         volume.StoragePoolName,
					StorageSystem:           volume.StorageSystem,
					Protocol:                volume.Protocol,
					CreatedTime:             volume.CreatedTime,
					StorageClass:            volume.StorageClass,
				}

				pvcSize := int64(1) // Placeholder value

				metric := &TopologyMetricsRecord{
					TopologyMeta: topologyMeta,
					PVCSize:      pvcSize,
				}

				s.Logger.Debugf("topology metrics - PV: %s, PVC Size: %d, Provisioned: %s",
					metric.TopologyMeta.PersistentVolume, metric.PVCSize, topologyMeta.ProvisionedSize)

				ch <- metric
			}(volume)
		}

		if !exported {
			ch <- &TopologyMetricsRecord{
				TopologyMeta: &TopologyMeta{},
				PVCSize:      0,
			}
		}

		wg.Wait()
		close(ch)
	}()

	return ch
}

func (s *PowerStoreService) pushTopologyMetrics(ctx context.Context, topologyMetrics <-chan *TopologyMetricsRecord, listOfPVs []string) <-chan *TopologyMetricsRecord {
	start := time.Now()
	defer s.timeSince(start, "pushTopologyMetrics")

	var wg sync.WaitGroup
	ch := make(chan *TopologyMetricsRecord)

	go func() {
		ctx, span := tracer.GetTracer(ctx, "pushTopologyMetrics")
		defer span.End()

		for metrics := range topologyMetrics {
			wg.Add(1)
			go func(metrics *TopologyMetricsRecord) {
				defer wg.Done()

				err := s.MetricsWrapper.RecordTopologyMetrics(ctx, metrics.TopologyMeta, metrics, listOfPVs)
				if err != nil {
					s.Logger.WithError(err).
						WithField("volume_id", metrics.TopologyMeta.PersistentVolume).
						Error("recording topology metrics for volume")
				} else {
					s.Logger.Debugf("recorded topology metrics for volume %s and size %s",
						metrics.TopologyMeta.PersistentVolume, metrics.TopologyMeta.ProvisionedSize)
					ch <- metrics
				}
			}(metrics)
		}

		wg.Wait()
		close(ch)
	}()

	return ch
}

// ExportTopologyMetrics records topology metrics for PowerStore volumes.
func (s *PowerStoreService) ExportTopologyMetrics(ctx context.Context) {
	ctx, span := tracer.GetTracer(ctx, "ExportTopologyMetrics")
	defer span.End()

	start := time.Now()
	defer s.timeSince(start, "ExportTopologyMetrics")

	if s.MetricsWrapper == nil {
		s.Logger.Warn("no MetricsWrapper provided for getting ExportTopologyMetrics")
		return
	}

	pvs, err := s.VolumeFinder.GetPersistentVolumes(ctx)
	if err != nil {
		s.Logger.WithError(err).Error("getting persistent volumes")
		return
	}

	// Track deleted PVs
	CurrentPVList = make(map[string]bool)
	for _, pv := range pvs {
		CurrentPVList[pv.PersistentVolume] = true
	}

	var deletedPVs []string
	for pv := range PrevPVList {
		if _, exists := CurrentPVList[pv]; !exists {
			deletedPVs = append(deletedPVs, pv)
		}
	}

	if len(deletedPVs) > 0 {
		s.Logger.Infof("Deleted PVs: %v", deletedPVs)
	}

	// Update PrevPVList
	PrevPVList = make(map[string]bool, len(CurrentPVList))
	for key, val := range CurrentPVList {
		PrevPVList[key] = val
	}

	// Trigger metric collection and push
	for range s.pushTopologyMetrics(ctx, s.gatherTopologyMetrics(ctx, s.volumeServer(ctx, pvs)), deletedPVs) {
		// consume the channel until it is empty and closed
	} // revive:disable-line:empty-block
}
