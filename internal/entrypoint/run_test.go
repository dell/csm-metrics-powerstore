/*
 *
 * Copyright © 2021-2024 Dell Inc. or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

/*
 Copyright (c) 2025 Dell Inc. or its subsidiaries. All Rights Reserved.

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

package entrypoint_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/dell/csm-metrics-powerstore/internal/entrypoint"
	"github.com/dell/csm-metrics-powerstore/internal/service"
	pStoreServices "github.com/dell/csm-metrics-powerstore/internal/service"
	"github.com/dell/csm-metrics-powerstore/internal/service/mocks"
	metrics "github.com/dell/csm-metrics-powerstore/internal/service/mocks"
	otlexporters "github.com/dell/csm-metrics-powerstore/opentelemetry/exporters"
	exportermocks "github.com/dell/csm-metrics-powerstore/opentelemetry/exporters/mocks"

	"github.com/golang/mock/gomock"
)

func Test_Run(t *testing.T) {
	tests := map[string]func(t *testing.T) (expectError bool, config *entrypoint.Config, exporter otlexporters.Otlexporter, pStoreSvc pStoreServices.Service, prevConfigValidationFunc func(*entrypoint.Config) error, ctrl *gomock.Controller, validatingConfig bool){
		"success": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pStoreServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)

			leaderElector := mocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-metrics-powerstore", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				VolumeMetricsEnabled:   true,
				TopologyMetricsEnabled: true,
				LeaderElector:          leaderElector,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter(gomock.Any(), gomock.Any()).Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metrics.NewMockService(ctrl)
			svc.EXPECT().ExportVolumeStatistics(gomock.Any()).AnyTimes()
			svc.EXPECT().ExportSpaceVolumeMetrics(gomock.Any()).AnyTimes()
			svc.EXPECT().ExportArraySpaceMetrics(gomock.Any()).AnyTimes()
			svc.EXPECT().ExportFileSystemStatistics(gomock.Any()).AnyTimes()
			svc.EXPECT().ExportTopologyMetrics(gomock.Any()).AnyTimes()

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"error with invalid volume ticker interval": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pStoreServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			leaderElector := mocks.NewMockLeaderElector(ctrl)
			clients := make(map[string]service.PowerStoreClient)
			clients["test"] = mocks.NewMockPowerStoreClient(ctrl)
			config := &entrypoint.Config{
				VolumeMetricsEnabled: true,
				LeaderElector:        leaderElector,
				VolumeTickInterval:   1 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			e := exportermocks.NewMockOtlexporter(ctrl)
			svc := metrics.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, true
		},
		"error with invalid space ticker interval": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pStoreServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			leaderElector := mocks.NewMockLeaderElector(ctrl)
			clients := make(map[string]service.PowerStoreClient)
			clients["test"] = mocks.NewMockPowerStoreClient(ctrl)
			config := &entrypoint.Config{
				VolumeMetricsEnabled: true,
				LeaderElector:        leaderElector,
				VolumeTickInterval:   200 * time.Second,
				SpaceTickInterval:    1 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			e := exportermocks.NewMockOtlexporter(ctrl)
			svc := metrics.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, true
		},
		"error with invalid array ticker interval": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pStoreServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			leaderElector := mocks.NewMockLeaderElector(ctrl)
			clients := make(map[string]service.PowerStoreClient)
			clients["test"] = mocks.NewMockPowerStoreClient(ctrl)
			config := &entrypoint.Config{
				VolumeMetricsEnabled: true,
				LeaderElector:        leaderElector,
				VolumeTickInterval:   200 * time.Second,
				SpaceTickInterval:    200 * time.Second,
				ArrayTickInterval:    1 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			e := exportermocks.NewMockOtlexporter(ctrl)
			svc := metrics.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, true
		},
		"error with invalid filesystem ticker interval": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pStoreServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			leaderElector := mocks.NewMockLeaderElector(ctrl)
			clients := make(map[string]service.PowerStoreClient)
			clients["test"] = mocks.NewMockPowerStoreClient(ctrl)
			config := &entrypoint.Config{
				VolumeMetricsEnabled:   true,
				LeaderElector:          leaderElector,
				VolumeTickInterval:     200 * time.Second,
				SpaceTickInterval:      200 * time.Second,
				ArrayTickInterval:      200 * time.Second,
				FileSystemTickInterval: 1 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			e := exportermocks.NewMockOtlexporter(ctrl)
			svc := metrics.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, true
		},
		"error with invalid topology ticker interval": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pStoreServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			leaderElector := mocks.NewMockLeaderElector(ctrl)
			config := &entrypoint.Config{
				VolumeMetricsEnabled:   true,
				LeaderElector:          leaderElector,
				VolumeTickInterval:     200 * time.Second,
				SpaceTickInterval:      200 * time.Second,
				ArrayTickInterval:      200 * time.Second,
				FileSystemTickInterval: 200 * time.Second,
				TopologyTickInterval:   1 * time.Second,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			e := exportermocks.NewMockOtlexporter(ctrl)
			svc := metrics.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, true
		},
		"error nil config": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pStoreServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)
			e := exportermocks.NewMockOtlexporter(ctrl)

			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			svc := metrics.NewMockService(ctrl)

			return true, nil, e, svc, prevConfigValidationFunc, ctrl, true
		},
		"error initializing exporter": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pStoreServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)

			leaderElector := mocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				LeaderElector: leaderElector,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter(gomock.Any(), gomock.Any()).Return(fmt.Errorf("An error occurred while initializing the exporter"))
			e.EXPECT().StopExporter().Return(nil)

			svc := metrics.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"error no LeaderElector": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pStoreServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)

			config := &entrypoint.Config{
				LeaderElector: nil,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = entrypoint.ValidateConfig

			e := exportermocks.NewMockOtlexporter(ctrl)

			svc := metrics.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"success even if leader is false": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pStoreServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)

			leaderElector := mocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-metrics-powerstore", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(false)

			config := &entrypoint.Config{
				LeaderElector: leaderElector,
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter(gomock.Any(), gomock.Any()).Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metrics.NewMockService(ctrl)

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"success using TLS": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pStoreServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)

			leaderElector := mocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-metrics-powerstore", "karavi").Times(1).Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				LeaderElector:     leaderElector,
				CollectorCertPath: "testdata/test-cert.crt",
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter(gomock.Any(), gomock.Any()).Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metrics.NewMockService(ctrl)

			return false, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
		"error reading certificate": func(*testing.T) (bool, *entrypoint.Config, otlexporters.Otlexporter, pStoreServices.Service, func(*entrypoint.Config) error, *gomock.Controller, bool) {
			ctrl := gomock.NewController(t)

			leaderElector := mocks.NewMockLeaderElector(ctrl)
			leaderElector.EXPECT().InitLeaderElection("karavi-metrics-powerstore", "karavi").AnyTimes().Return(nil)
			leaderElector.EXPECT().IsLeader().AnyTimes().Return(true)

			config := &entrypoint.Config{
				LeaderElector:     leaderElector,
				CollectorCertPath: "testdata/bad-cert.crt",
			}
			prevConfigValidationFunc := entrypoint.ConfigValidatorFunc
			entrypoint.ConfigValidatorFunc = noCheckConfig

			e := exportermocks.NewMockOtlexporter(ctrl)
			e.EXPECT().InitExporter(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			e.EXPECT().StopExporter().Return(nil)

			svc := metrics.NewMockService(ctrl)

			return true, config, e, svc, prevConfigValidationFunc, ctrl, false
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			expectError, config, exporter, svc, prevConfValidation, ctrl, validateConfig := test(t)
			ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
			defer cancel()
			if config != nil {
				config.Logger = logrus.New()
				if !validateConfig {
					// The configuration is not nil and the test is not attempting to validate the configuration.
					// In this case, we can use smaller intervals for testing purposes.
					config.VolumeTickInterval = 100 * time.Millisecond
					config.SpaceTickInterval = 100 * time.Millisecond
					config.ArrayTickInterval = 100 * time.Millisecond
					config.FileSystemTickInterval = 100 * time.Millisecond
					config.TopologyTickInterval = 100 * time.Millisecond
				}
			}
			err := entrypoint.Run(ctx, config, exporter, svc)
			errorOccurred := err != nil
			if expectError != errorOccurred {
				t.Errorf("Unexpected result from test \"%v\": wanted error (%v), but got (%v)", name, expectError, errorOccurred)
			}
			entrypoint.ConfigValidatorFunc = prevConfValidation
			ctrl.Finish()
		})
	}
}

func noCheckConfig(_ *entrypoint.Config) error {
	return nil
}

func Test_ValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *entrypoint.Config
		wantErr bool
	}{
		{"valid config", &entrypoint.Config{VolumeTickInterval: 10 * time.Second, SpaceTickInterval: 10 * time.Second, ArrayTickInterval: 10 * time.Second, FileSystemTickInterval: 10 * time.Second, TopologyTickInterval: 10 * time.Second}, false},
		{"nil config", nil, true},
		{"invalid volume tick interval", &entrypoint.Config{VolumeTickInterval: 1 * time.Second}, true},
		{"invalid space tick interval", &entrypoint.Config{SpaceTickInterval: 1 * time.Second}, true},
		{"invalid array tick interval", &entrypoint.Config{ArrayTickInterval: 1 * time.Second}, true},
		{"invalid filesystem tick interval", &entrypoint.Config{FileSystemTickInterval: 1 * time.Second}, true},
		{"invalid topology tick interval", &entrypoint.Config{TopologyTickInterval: 1 * time.Second}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := entrypoint.ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
