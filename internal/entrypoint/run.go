// Copyright (c) 2021 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

package entrypoint

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/dell/csm-metrics-powerstore/internal/service"
	pstoreServices "github.com/dell/csm-metrics-powerstore/internal/service"
	otlexporters "github.com/dell/csm-metrics-powerstore/opentelemetry/exporters"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/exporters/otlp"
	"google.golang.org/grpc/credentials"
)

const (
	// MaximumVolTickInterval is the maximum allowed interval when querying volume metrics
	MaximumVolTickInterval = 10 * time.Minute
	// MinimumVolTickInterval is the minimum allowed interval when querying volume metrics
	MinimumVolTickInterval = 5 * time.Second
	// DefaultEndPoint for leader election path
	DefaultEndPoint = "csm-metrics-powerstore"
	// DefaultNameSpace for PowerStore pod running metrics collection
	DefaultNameSpace = "karavi"
)

var (
	// ConfigValidatorFunc is used to override config validation in testing
	ConfigValidatorFunc func(*Config) error = ValidateConfig
)

// Config holds data that will be used by the service
type Config struct {
	VolumeTickInterval   time.Duration
	LeaderElector        service.LeaderElector
	VolumeMetricsEnabled bool
	CollectorAddress     string
	CollectorCertPath    string
	Logger               *logrus.Logger
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
		options := []otlp.ExporterOption{
			otlp.WithAddress(config.CollectorAddress),
		}

		if config.CollectorCertPath != "" {
			transportCreds, err := credentials.NewClientTLSFromFile(config.CollectorCertPath, "")
			if err != nil {
				errCh <- err
			}
			options = append(options, otlp.WithTLSCredentials(transportCreds))
		} else {
			options = append(options, otlp.WithInsecure())
		}

		errCh <- exporter.InitExporter(options...)
	}()

	defer exporter.StopExporter()

	runtime.GOMAXPROCS(runtime.NumCPU())

	//set initial tick intervals
	VolumeTickInterval := config.VolumeTickInterval
	volumeTicker := time.NewTicker(VolumeTickInterval)
	for {
		select {
		case <-volumeTicker.C:
			if !config.LeaderElector.IsLeader() {
				logger.Info("not leader pod to collect metrics")
				continue
			}
			if !config.VolumeMetricsEnabled {
				logger.Info("powerstore volume metrics collection is disabled")
				continue
			}
			powerStoreSvc.ExportVolumeStatistics(ctx)
		case err := <-errCh:
			if err == nil {
				continue
			}
			return err
		case <-ctx.Done():
			return nil
		}

		//check if tick interval config settings have changed
		if VolumeTickInterval != config.VolumeTickInterval {
			VolumeTickInterval = config.VolumeTickInterval
			volumeTicker = time.NewTicker(VolumeTickInterval)
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

	return nil
}
