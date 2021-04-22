// Copyright (c) 2021 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dell/csi-powerstore/pkg/array"
	"github.com/dell/csi-powerstore/pkg/common/fs"
	"github.com/dell/csm-metrics-powerstore/internal/entrypoint"
	"github.com/dell/csm-metrics-powerstore/internal/k8s"
	"github.com/dell/csm-metrics-powerstore/internal/service"
	otlexporters "github.com/dell/csm-metrics-powerstore/opentelemetry/exporters"
	"github.com/dell/gofsutil"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/api/global"

	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

const (
	defaultTickInterval            = 5 * time.Second
	defaultConfigFile              = "/etc/config/csm-metrics-powerstore.yaml"
	defaultStorageSystemConfigFile = "/powerstore-config/config"
)

func main() {

	logger := logrus.New()

	viper.SetConfigFile(defaultConfigFile)

	err := viper.ReadInConfig()
	// if unable to read configuration file, proceed in case we use environment variables
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to read Config file: %v", err)
	}

	configFileListener := viper.New()
	configFileListener.SetConfigFile(defaultStorageSystemConfigFile)

	leaderElectorGetter := &k8s.LeaderElector{
		API: &k8s.LeaderElector{},
	}

	volumeFinder := &k8s.VolumeFinder{
		API: &k8s.API{},
	}

	updateLoggingSettings := func(logger *logrus.Logger) {
		logFormat := viper.GetString("LOG_FORMAT")
		if strings.EqualFold(logFormat, "json") {
			logger.SetFormatter(&logrus.JSONFormatter{})
		} else {
			// use text formatter by default
			logger.SetFormatter(&logrus.TextFormatter{})
		}
		logLevel := viper.GetString("LOG_LEVEL")
		level, err := logrus.ParseLevel(logLevel)
		if err != nil {
			// use INFO level by default
			level = logrus.InfoLevel
		}
		logger.SetLevel(level)
	}

	updateLoggingSettings(logger)
	updateProvisionerNames(volumeFinder, logger)

	var collectorCertPath string
	if tls := os.Getenv("TLS_ENABLED"); tls == "true" {
		collectorCertPath = os.Getenv("COLLECTOR_CERT_PATH")
		if len(strings.TrimSpace(collectorCertPath)) < 1 {
			collectorCertPath = otlexporters.DefaultCollectorCertPath
		}
	}

	config := &entrypoint.Config{
		LeaderElector:     leaderElectorGetter,
		CollectorCertPath: collectorCertPath,
		Logger:            logger,
	}

	exporter := &otlexporters.OtlCollectorExporter{}

	powerStoreSvc := &service.PowerStoreService{
		MetricsWrapper: &service.MetricsWrapper{
			Meter: global.Meter("powerstore"),
		},
		Logger:       logger,
		VolumeFinder: volumeFinder,
	}

	updatePowerStoreConnection(powerStoreSvc, logger)
	updateCollectorAddress(config, exporter, logger)
	updateMetricsEnabled(config)
	updateTickIntervals(config, logger)
	updateService(powerStoreSvc, logger)

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		updateLoggingSettings(logger)
		updateCollectorAddress(config, exporter, logger)
		updateProvisionerNames(volumeFinder, logger)
		updateMetricsEnabled(config)
		updateTickIntervals(config, logger)
		updateService(powerStoreSvc, logger)
	})

	configFileListener.WatchConfig()
	configFileListener.OnConfigChange(func(e fsnotify.Event) {
		updatePowerStoreConnection(powerStoreSvc, logger)
	})

	if err := entrypoint.Run(context.Background(), config, exporter, powerStoreSvc); err != nil {
		logger.WithError(err).Fatal("running service")
	}
}

func updatePowerStoreConnection(powerStoreSvc *service.PowerStoreService, logger *logrus.Logger) {
	f := &fs.Fs{Util: &gofsutil.FS{}}
	arrays, _, err := array.GetPowerStoreArrays(f, defaultStorageSystemConfigFile)
	if err != nil {
		logger.WithError(err).Fatal("initialize arrays in controller service")
	}
	newArrays := make(map[string]service.PowerStoreClient)
	for k, v := range arrays {
		newArrays[k] = v.Client
	}
	powerStoreSvc.PowerStoreClients = newArrays
}

func updateCollectorAddress(config *entrypoint.Config, exporter *otlexporters.OtlCollectorExporter, logger *logrus.Logger) {
	collectorAddress := viper.GetString("COLLECTOR_ADDR")
	if collectorAddress == "" {
		logger.Fatal("COLLECTOR_ADDR is required")
	}
	config.CollectorAddress = collectorAddress
	exporter.CollectorAddr = collectorAddress
}

func updateProvisionerNames(volumeFinder *k8s.VolumeFinder, logger *logrus.Logger) {
	provisionerNamesValue := viper.GetString("provisioner_names")
	if provisionerNamesValue == "" {
		logger.Fatal("PROVISIONER_NAMES is required")
	}
	provisionerNames := strings.Split(provisionerNamesValue, ",")
	volumeFinder.DriverNames = provisionerNames
}

func updateMetricsEnabled(config *entrypoint.Config) {
	powerstoreVolumeMetricsEnabled := true
	powerstoreVolumeMetricsEnabledValue := viper.GetString("POWERSTORE_VOLUME_METRICS_ENABLED")
	if powerstoreVolumeMetricsEnabledValue == "false" {
		powerstoreVolumeMetricsEnabled = false
	}
	config.VolumeMetricsEnabled = powerstoreVolumeMetricsEnabled
}

func updateTickIntervals(config *entrypoint.Config, logger *logrus.Logger) {
	volumeTickInterval := defaultTickInterval
	volIoPollFrequencySeconds := viper.GetString("POWERSTORE_VOLUME_IO_POLL_FREQUENCY")
	if volIoPollFrequencySeconds != "" {
		numSeconds, err := strconv.Atoi(volIoPollFrequencySeconds)
		if err != nil {
			logger.WithError(err).Fatal("POWERSTORE_VOLUME_IO_POLL_FREQUENCY was not set to a valid number")
		}
		volumeTickInterval = time.Duration(numSeconds) * time.Second
	}
	config.VolumeTickInterval = volumeTickInterval
}

func updateService(pstoreSvc *service.PowerStoreService, logger *logrus.Logger) {
	maxPowerStoreConcurrentRequests := service.DefaultMaxPowerStoreConnections
	maxPowerStoreConcurrentRequestsVar := viper.GetString("POWERSTORE_MAX_CONCURRENT_QUERIES")
	if maxPowerStoreConcurrentRequestsVar != "" {
		maxPowerStoreConcurrentRequests, err := strconv.Atoi(maxPowerStoreConcurrentRequestsVar)
		if err != nil {
			logger.WithError(err).Fatal("POWERSTORE_MAX_CONCURRENT_QUERIES was not set to a valid number")
		}
		if maxPowerStoreConcurrentRequests <= 0 {
			logger.WithError(err).Fatal("POWERSTORE_MAX_CONCURRENT_QUERIES value was invalid (<= 0)")
		}
	}
	pstoreSvc.MaxPowerStoreConnections = maxPowerStoreConcurrentRequests
}
