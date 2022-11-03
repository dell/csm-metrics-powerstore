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
	"github.com/dell/gopowerstore"
	"github.com/sirupsen/logrus"

	"github.com/dell/csm-metrics-powerstore/internal/k8s"

	"github.com/golang/mock/gomock"
)

func Test_ExportVolumeStatistics(t *testing.T) {
	type setup struct {
		Service *service.PowerStoreService
	}

	tests := map[string]func(t *testing.T) (service.PowerStoreService, *gomock.Controller){
		"success": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{
				k8s.VolumeInfo{
					PersistentVolume: "pv-1",
					VolumeHandle:     "volume-1/127.0.0.1/scsi",
				},
				k8s.VolumeInfo{
					PersistentVolume: "pv-2",
					VolumeHandle:     "volume-2/127.0.0.1/scsi",
				},
				k8s.VolumeInfo{
					PersistentVolume: "pv-3",
					VolumeHandle:     "volume-2/127.0.0.1/scsi",
				},
			}, nil).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().PerformanceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Return([]gopowerstore.PerformanceMetricsByVolumeResponse{
				gopowerstore.PerformanceMetricsByVolumeResponse{
					CommonMaxAvgIopsBandwidthFields: gopowerstore.CommonMaxAvgIopsBandwidthFields{
						ReadBandwidth:  1,
						WriteBandwidth: 1,
						ReadIops:       1,
						WriteIops:      1,
					},
					CommonAvgFields: gopowerstore.CommonAvgFields{
						AvgReadLatency:  1,
						AvgWriteLatency: 1,
					},
				},
			}, nil).Times(3)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed if volume does not have scsi protocol": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{
				k8s.VolumeInfo{
					PersistentVolume: "pv-1",
					VolumeHandle:     "volume-1/127.0.0.1/nfs",
				},
			}, nil).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().PerformanceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed if error getting volume metrics": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{
				k8s.VolumeInfo{
					PersistentVolume: "pv-1",
					VolumeHandle:     "volume-1/127.0.0.1/scsi",
				},
			}, nil).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().PerformanceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed if client not found for array ip in volume handle": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{
				k8s.VolumeInfo{
					PersistentVolume: "pv-1",
					VolumeHandle:     "volume-1/127.0.0.1/scsi",
				},
			}, nil).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().PerformanceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.2"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed if volume handle is invalid": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{
				k8s.VolumeInfo{
					PersistentVolume: "pv-1",
					VolumeHandle:     "invalid-volume-handle",
				},
			}, nil).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().PerformanceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed if volume finder returns error": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return(nil, errors.New("error")).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().PerformanceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed if metrics wrapper is nil": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Times(0)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().PerformanceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    nil,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed with 0 volumes": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			volFinder := mocks.NewMockVolumeFinder(ctrl)

			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{}, nil)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().PerformanceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{MetricsWrapper: metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			metrics.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			return service, ctrl
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			service, ctrl := tc(t)
			service.Logger = logrus.New()
			service.ExportVolumeStatistics(context.Background())
			ctrl.Finish()
		})
	}
}

func Test_ExportSpaceVolumeMetrics(t *testing.T) {
	type setup struct {
		Service *service.PowerStoreService
	}

	tests := map[string]func(t *testing.T) (service.PowerStoreService, *gomock.Controller){
		"success": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().RecordSpaceMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{
				k8s.VolumeInfo{
					PersistentVolume: "pv-1",
					VolumeHandle:     "volume-1/127.0.0.1/scsi",
				},
				k8s.VolumeInfo{
					PersistentVolume: "pv-2",
					VolumeHandle:     "volume-2/127.0.0.1/scsi",
				},
				k8s.VolumeInfo{
					PersistentVolume: "pv-3",
					VolumeHandle:     "volume-2/127.0.0.1/scsi",
				},
			}, nil).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().SpaceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Return([]gopowerstore.SpaceMetricsByVolumeResponse{
				gopowerstore.SpaceMetricsByVolumeResponse{
					LogicalProvisioned:     new(int64),
					LogicalUsed:            new(int64),
					LastLogicalProvisioned: new(int64),
					LastLogicalUsed:        new(int64),
					ThinSavings:            1,
					MaxThinSavings:         1,
				},
			}, nil).Times(3)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"success for filesystem": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().RecordSpaceMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{
				k8s.VolumeInfo{
					PersistentVolume: "pv-1",
					VolumeHandle:     "volume-1/127.0.0.1/nfs",
				},
				k8s.VolumeInfo{
					PersistentVolume: "pv-2",
					VolumeHandle:     "volume-2/127.0.0.1/nfs",
				},
				k8s.VolumeInfo{
					PersistentVolume: "pv-3",
					VolumeHandle:     "volume-2/127.0.0.1/nfs",
				},
			}, nil).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().GetFS(gomock.Any(), gomock.Any()).Return(gopowerstore.FileSystem{
				SizeTotal: 10,
				SizeUsed:  2,
			}, nil).Times(3)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed if error getting space metrics": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().RecordSpaceMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{
				k8s.VolumeInfo{
					PersistentVolume: "pv-1",
					VolumeHandle:     "volume-1/127.0.0.1/scsi",
				},
			}, nil).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().SpaceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed if client not found for array ip in volume handle": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().RecordSpaceMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{
				k8s.VolumeInfo{
					PersistentVolume: "pv-1",
					VolumeHandle:     "volume-1/127.0.0.1/scsi",
				},
			}, nil).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().SpaceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.2"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed if volume handle is invalid": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().RecordSpaceMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{
				k8s.VolumeInfo{
					PersistentVolume: "pv-1",
					VolumeHandle:     "invalid-volume-handle",
				},
			}, nil).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().SpaceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed if volume finder returns error": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().RecordSpaceMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return(nil, errors.New("error")).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().SpaceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed if metrics wrapper is nil": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Times(0)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().SpaceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    nil,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed with 0 volumes": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			volFinder := mocks.NewMockVolumeFinder(ctrl)

			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{}, nil)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().SpaceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{MetricsWrapper: metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			metrics.EXPECT().RecordSpaceMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			return service, ctrl
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			service, ctrl := tc(t)
			service.Logger = logrus.New()
			service.ExportSpaceVolumeMetrics(context.Background())
			ctrl.Finish()
		})
	}
}

func Test_ExportArraySpaceMetrics(t *testing.T) {
	type setup struct {
		Service *service.PowerStoreService
	}

	tests := map[string]func(t *testing.T) (service.PowerStoreService, *gomock.Controller){
		"success": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().RecordArraySpaceMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
			metrics.EXPECT().RecordStorageClassSpaceMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{
				k8s.VolumeInfo{
					PersistentVolume: "pv-1",
					VolumeHandle:     "volume-1/127.0.0.1/scsi",
				},
				k8s.VolumeInfo{
					PersistentVolume: "pv-2",
					VolumeHandle:     "volume-2/127.0.0.1/scsi",
				},
				k8s.VolumeInfo{
					PersistentVolume: "pv-3",
					VolumeHandle:     "volume-2/127.0.0.1/scsi",
				},
			}, nil).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().SpaceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Return([]gopowerstore.SpaceMetricsByVolumeResponse{
				gopowerstore.SpaceMetricsByVolumeResponse{
					LogicalProvisioned:     new(int64),
					LogicalUsed:            new(int64),
					LastLogicalProvisioned: new(int64),
					LastLogicalUsed:        new(int64),
					ThinSavings:            1,
					MaxThinSavings:         1,
				},
			}, nil).Times(3)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"success for filesystem": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().RecordArraySpaceMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
			metrics.EXPECT().RecordStorageClassSpaceMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{
				k8s.VolumeInfo{
					PersistentVolume: "pv-1",
					VolumeHandle:     "volume-1/127.0.0.1/nfs",
				},
				k8s.VolumeInfo{
					PersistentVolume: "pv-2",
					VolumeHandle:     "volume-2/127.0.0.1/nfs",
				},
				k8s.VolumeInfo{
					PersistentVolume: "pv-3",
					VolumeHandle:     "volume-2/127.0.0.1/nfs",
				},
			}, nil).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().GetFS(gomock.Any(), gomock.Any()).Return(gopowerstore.FileSystem{
				SizeTotal: 10,
				SizeUsed:  2,
			}, nil).Times(3)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed if error getting space metrics": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().RecordArraySpaceMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
			metrics.EXPECT().RecordStorageClassSpaceMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{
				k8s.VolumeInfo{
					PersistentVolume: "pv-1",
					VolumeHandle:     "volume-1/127.0.0.1/scsi",
				},
			}, nil).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().SpaceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed if client not found for array ip in volume handle": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().RecordArraySpaceMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			metrics.EXPECT().RecordStorageClassSpaceMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{
				k8s.VolumeInfo{
					PersistentVolume: "pv-1",
					VolumeHandle:     "volume-1/127.0.0.1/scsi",
				},
			}, nil).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().SpaceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.2"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed if volume handle is invalid": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().RecordArraySpaceMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			metrics.EXPECT().RecordStorageClassSpaceMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{
				k8s.VolumeInfo{
					PersistentVolume: "pv-1",
					VolumeHandle:     "invalid-volume-handle",
				},
			}, nil).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().SpaceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed if volume finder returns error": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().RecordArraySpaceMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			metrics.EXPECT().RecordStorageClassSpaceMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return(nil, errors.New("error")).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().SpaceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed if metrics wrapper is nil": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Times(0)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().SpaceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    nil,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed with 0 volumes": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			volFinder := mocks.NewMockVolumeFinder(ctrl)

			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{}, nil)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().SpaceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{MetricsWrapper: metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			metrics.EXPECT().RecordArraySpaceMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			metrics.EXPECT().RecordStorageClassSpaceMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			return service, ctrl
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			service, ctrl := tc(t)
			service.Logger = logrus.New()
			service.ExportArraySpaceMetrics(context.Background())
			ctrl.Finish()
		})
	}
}

func Test_ExportFileSystemStatistics(t *testing.T) {
	type setup struct {
		Service *service.PowerStoreService
	}

	tests := map[string]func(t *testing.T) (service.PowerStoreService, *gomock.Controller){
		"success": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().RecordFileSystemMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{
				k8s.VolumeInfo{
					PersistentVolume: "pv-1",
					VolumeHandle:     "volume-1/127.0.0.1/nfs",
				},
				k8s.VolumeInfo{
					PersistentVolume: "pv-2",
					VolumeHandle:     "volume-2/127.0.0.1/nfs",
				},
				k8s.VolumeInfo{
					PersistentVolume: "pv-3",
					VolumeHandle:     "volume-2/127.0.0.1/nfs",
				},
			}, nil).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().PerformanceMetricsByFileSystem(gomock.Any(), gomock.Any(), gomock.Any()).Return([]gopowerstore.PerformanceMetricsByFileSystemResponse{
				gopowerstore.PerformanceMetricsByFileSystemResponse{
					ReadBandwidth:   1,
					WriteBandwidth:  1,
					ReadIops:        1,
					WriteIops:       1,
					AvgReadLatency:  1,
					AvgWriteLatency: 1,
				},
			}, nil).Times(3)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed if volume does not have scsi protocol": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().RecordFileSystemMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{
				k8s.VolumeInfo{
					PersistentVolume: "pv-1",
					VolumeHandle:     "volume-1/127.0.0.1/scsi",
				},
			}, nil).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().PerformanceMetricsByFileSystem(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed if error getting volume metrics": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().RecordFileSystemMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{
				k8s.VolumeInfo{
					PersistentVolume: "pv-1",
					VolumeHandle:     "volume-1/127.0.0.1/nfs",
				},
			}, nil).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().PerformanceMetricsByFileSystem(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed if client not found for array ip in volume handle": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().RecordFileSystemMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{
				k8s.VolumeInfo{
					PersistentVolume: "pv-1",
					VolumeHandle:     "volume-1/127.0.0.1/nfs",
				},
			}, nil).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().PerformanceMetricsByFileSystem(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.2"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed if volume handle is invalid": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().RecordFileSystemMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{
				k8s.VolumeInfo{
					PersistentVolume: "pv-1",
					VolumeHandle:     "invalid-volume-handle",
				},
			}, nil).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().PerformanceMetricsByFileSystem(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed if volume finder returns error": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			metrics.EXPECT().RecordFileSystemMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return(nil, errors.New("error")).Times(1)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().PerformanceMetricsByFileSystem(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed if metrics wrapper is nil": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			volFinder := mocks.NewMockVolumeFinder(ctrl)
			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Times(0)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().PerformanceMetricsByFileSystem(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{
				MetricsWrapper:    nil,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			return service, ctrl
		},
		"metrics not pushed with 0 volumes": func(*testing.T) (service.PowerStoreService, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			volFinder := mocks.NewMockVolumeFinder(ctrl)

			volFinder.EXPECT().GetPersistentVolumes(gomock.Any()).Return([]k8s.VolumeInfo{}, nil)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().PerformanceMetricsByFileSystem(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{MetricsWrapper: metrics,
				VolumeFinder:      volFinder,
				PowerStoreClients: clients,
			}
			metrics.EXPECT().RecordFileSystemMetrics(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			return service, ctrl
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			service, ctrl := tc(t)
			service.Logger = logrus.New()
			service.ExportFileSystemStatistics(context.Background())
			ctrl.Finish()
		})
	}
}
