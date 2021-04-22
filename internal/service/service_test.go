// Copyright (c) 2021 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

package service_test

import (
	"context"
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

	tests := map[string]func(t *testing.T) (service.PowerStoreService, map[string]service.PowerStoreClient, service.VolumeFinder, *gomock.Controller){
		"success": func(*testing.T) (service.PowerStoreService, map[string]service.PowerStoreClient, service.VolumeFinder, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			volFinder := mocks.NewMockVolumeFinder(ctrl)

			volFinder.EXPECT().GetPersistentVolumes().Return([]k8s.VolumeInfo{
				k8s.VolumeInfo{
					PersistentVolume: "pv-1",
					VolumeHandle:     "volume-1/127.0.0.1/protocol",
				},
				k8s.VolumeInfo{
					PersistentVolume: "pv-2",
					VolumeHandle:     "volume-2/127.0.0.1/protocol",
				},
				k8s.VolumeInfo{
					PersistentVolume: "pv-3",
					VolumeHandle:     "volume-2/127.0.0.1/protocol",
				},
			}, nil)

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

			service := service.PowerStoreService{MetricsWrapper: metrics}
			metrics.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)
			return service, clients, volFinder, ctrl
		},
		"success with 0 volumes": func(*testing.T) (service.PowerStoreService, map[string]service.PowerStoreClient, service.VolumeFinder, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			metrics := mocks.NewMockMetricsRecorder(ctrl)
			volFinder := mocks.NewMockVolumeFinder(ctrl)

			volFinder.EXPECT().GetPersistentVolumes().Return([]k8s.VolumeInfo{}, nil)

			clients := make(map[string]service.PowerStoreClient)
			c := mocks.NewMockPowerStoreClient(ctrl)
			c.EXPECT().PerformanceMetricsByVolume(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			clients["127.0.0.1"] = c

			service := service.PowerStoreService{MetricsWrapper: metrics}
			metrics.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			return service, clients, volFinder, ctrl
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			service, clients, volFinder, ctrl := tc(t)
			service.Logger = logrus.New()
			service.ExportVolumeStatistics(context.Background(), clients, volFinder)
			ctrl.Finish()
		})
	}
}
