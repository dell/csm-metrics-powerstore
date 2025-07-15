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

package entrypoint

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	pstoreServices "github.com/dell/csm-metrics-powerstore/internal/service"
	otlexporters "github.com/dell/csm-metrics-powerstore/opentelemetry/exporters"
	tracer "github.com/dell/csm-metrics-powerstore/opentelemetry/tracers"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"google.golang.org/grpc/credentials"
)

const (
	// MaximumVolTickInterval is the maximum allowed interval when querying volume metrics
	MaximumVolTickInterval = 10 * time.Minute
	// MinimumVolTickInterval is the minimum allowed interval when querying volume metrics
	MinimumVolTickInterval = 5 * time.Second
	// DefaultEndPoint for leader election path
	DefaultEndPoint = "karavi-metrics-powerstore"
	// DefaultNameSpace for PowerStore pod running metrics collection
	DefaultNameSpace = "karavi"
)

// ConfigValidatorFunc is used to override config validation in testing
var ConfigValidatorFunc = ValidateConfig

// Config holds data that will be used by the service
type Config struct {
	VolumeTickInterval     time.Duration
	SpaceTickInterval      time.Duration
	ArrayTickInterval      time.Duration
	FileSystemTickInterval time.Duration
	TopologyTickInterval   time.Duration
	LeaderElector          pstoreServices.LeaderElector
	VolumeMetricsEnabled   bool
	CollectorAddress       string
	CollectorCertPath      string
	Logger                 *logrus.Logger
	TopologyMetricsEnabled bool
}

// Run is the entry point for starting the service
func Run(ctx context.Context, config *Config, exporter otlexporters.Otlexporter, powerStoreSvc pstoreServices.Service) error {
	err := ConfigValidatorFunc(config)
	if err != nil {
		return err
	}
	logger := config.Logger

	errCh := make(chan error, 1)
	go func() {
		powerstoreEndpoint := os.Getenv("POWERSTORE_METRICS_ENDPOINT")
		if powerstoreEndpoint == "" {
			powerstoreEndpoint = DefaultEndPoint
		}
		powerstoreNamespace := os.Getenv("POWERSTORE_METRICS_NAMESPACE")
		if powerstoreNamespace == "" {
			powerstoreNamespace = DefaultNameSpace
		}
		errCh <- config.LeaderElector.InitLeaderElection(powerstoreEndpoint, powerstoreNamespace)
	}()

	go func() {
		options := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithEndpoint(config.CollectorAddress),
		}

		if config.CollectorCertPath != "" {
			transportCreds, err := credentials.NewClientTLSFromFile(config.CollectorCertPath, "")
			if err != nil {
				errCh <- err
			}
			options = append(options, otlpmetricgrpc.WithTLSCredentials(transportCreds))
		} else {
			options = append(options, otlpmetricgrpc.WithInsecure())
		}

		errCh <- exporter.InitExporter(options...)
	}()

	defer exporter.StopExporter()

	runtime.GOMAXPROCS(runtime.NumCPU())

	// set initial tick intervals
	VolumeTickInterval := config.VolumeTickInterval
	volumeTicker := time.NewTicker(VolumeTickInterval)
	SpaceTickInterval := config.SpaceTickInterval
	spaceTicker := time.NewTicker(SpaceTickInterval)
	ArrayTickInterval := config.ArrayTickInterval
	arrayTicker := time.NewTicker(ArrayTickInterval)
	FileSystemTickInterval := config.FileSystemTickInterval
	filesystemTicker := time.NewTicker(FileSystemTickInterval)
	topologyTickInterval := config.TopologyTickInterval
	topologyTicker := time.NewTicker(topologyTickInterval)

	for {
		select {
		case <-volumeTicker.C:
			ctx, span := tracer.GetTracer(ctx, "volume-metrics")
			if !config.LeaderElector.IsLeader() {
				logger.Info("not leader pod to collect metrics")
				span.End()
				continue
			}
			if !config.VolumeMetricsEnabled {
				logger.Info("powerstore volume metrics collection is disabled")
				span.End()
				continue
			}
			powerStoreSvc.ExportVolumeStatistics(ctx)
			span.End()

		case <-spaceTicker.C:
			ctx, span := tracer.GetTracer(ctx, "volume-space-metrics")
			if !config.LeaderElector.IsLeader() {
				logger.Info("not leader pod to collect metrics")
				span.End()
				continue
			}
			if !config.VolumeMetricsEnabled {
				logger.Info("powerstore volume metrics collection is disabled")
				span.End()
				continue
			}
			powerStoreSvc.ExportSpaceVolumeMetrics(ctx)
			span.End()

		case <-arrayTicker.C:
			ctx, span := tracer.GetTracer(ctx, "array-space-metrics")
			if !config.LeaderElector.IsLeader() {
				logger.Info("not leader pod to collect metrics")
				span.End()
				continue
			}
			if !config.VolumeMetricsEnabled {
				logger.Info("powerstore volume metrics collection is disabled")
				span.End()
				continue
			}
			powerStoreSvc.ExportArraySpaceMetrics(ctx)
			span.End()

		case <-filesystemTicker.C:
			ctx, span := tracer.GetTracer(ctx, "filesystem-metrics")
			if !config.LeaderElector.IsLeader() {
				logger.Info("not leader pod to collect metrics")
				span.End()
				continue
			}
			if !config.VolumeMetricsEnabled {
				logger.Info("powerstore filesystem metrics collection is disabled")
				span.End()
				continue
			}
			powerStoreSvc.ExportFileSystemStatistics(ctx)
			span.End()

		case <-topologyTicker.C:
			ctx, span := tracer.GetTracer(ctx, "topology-metrics")
			if !config.LeaderElector.IsLeader() {
				logger.Info("not leader pod to collect metrics")
				span.End()
				continue
			}
			if !config.TopologyMetricsEnabled {
				logger.Info("powerstore topology metrics collection is disabled")
				span.End()
				continue
			}
			powerStoreSvc.ExportTopologyMetrics(ctx)
			span.End()

		case err := <-errCh:
			if err == nil {
				continue
			}
			return err

		case <-ctx.Done():
			return nil
		}

		// check if tick interval config settings have changed
		if VolumeTickInterval != config.VolumeTickInterval {
			VolumeTickInterval = config.VolumeTickInterval
			volumeTicker = time.NewTicker(VolumeTickInterval)
		}
		if SpaceTickInterval != config.SpaceTickInterval {
			SpaceTickInterval = config.SpaceTickInterval
			spaceTicker = time.NewTicker(SpaceTickInterval)
		}
		if ArrayTickInterval != config.ArrayTickInterval {
			ArrayTickInterval = config.ArrayTickInterval
			arrayTicker = time.NewTicker(ArrayTickInterval)
		}
		if FileSystemTickInterval != config.FileSystemTickInterval {
			FileSystemTickInterval = config.FileSystemTickInterval
			filesystemTicker = time.NewTicker(FileSystemTickInterval)
		}
		if topologyTickInterval != config.TopologyTickInterval {
			topologyTickInterval = config.TopologyTickInterval
			topologyTicker = time.NewTicker(topologyTickInterval)
		}
	}
}

// ValidateConfig will validate the configuration and return any errors
func ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("no config provided")
	}

	if config.VolumeTickInterval > MaximumVolTickInterval || config.VolumeTickInterval < MinimumVolTickInterval {
		return fmt.Errorf("volume polling frequency not within allowed range of %v and %v", MinimumVolTickInterval.String(), MaximumVolTickInterval.String())
	}

	if config.SpaceTickInterval > MaximumVolTickInterval || config.SpaceTickInterval < MinimumVolTickInterval {
		return fmt.Errorf("volume space polling frequency not within allowed range of %v and %v", MinimumVolTickInterval.String(), MaximumVolTickInterval.String())
	}

	if config.ArrayTickInterval > MaximumVolTickInterval || config.ArrayTickInterval < MinimumVolTickInterval {
		return fmt.Errorf("array space polling frequency not within allowed range of %v and %v", MinimumVolTickInterval.String(), MaximumVolTickInterval.String())
	}

	if config.FileSystemTickInterval > MaximumVolTickInterval || config.FileSystemTickInterval < MinimumVolTickInterval {
		return fmt.Errorf("filesystem polling frequency not within allowed range of %v and %v", MinimumVolTickInterval.String(), MaximumVolTickInterval.String())
	}

	if config.TopologyTickInterval > MaximumVolTickInterval || config.TopologyTickInterval < MinimumVolTickInterval {
		return fmt.Errorf("topology polling frequency not within allowed range of %v and %v", MinimumVolTickInterval.String(), MaximumVolTickInterval.String())
	}

	return nil
}
