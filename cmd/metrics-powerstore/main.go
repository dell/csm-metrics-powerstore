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
	"expvar"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	_ "net/http/pprof"

	"github.com/dell/csi-powerstore/pkg/array"
	"github.com/dell/csi-powerstore/pkg/common/fs"
	"github.com/dell/csm-metrics-powerstore/internal/entrypoint"
	"github.com/dell/csm-metrics-powerstore/internal/k8s"
	"github.com/dell/csm-metrics-powerstore/internal/service"
	otlexporters "github.com/dell/csm-metrics-powerstore/opentelemetry/exporters"
	tracer "github.com/dell/csm-metrics-powerstore/opentelemetry/tracers"
	"github.com/dell/gofsutil"
	"github.com/sirupsen/logrus"

	"os"

	"go.opentelemetry.io/otel/api/global"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

const (
	defaultTickInterval            = 20 * time.Second
	defaultConfigFile              = "/etc/config/karavi-metrics-powerstore.yaml"
	defaultStorageSystemConfigFile = "/powerstore-config/config"
	defaultDebugPort               = "9090"
	defaultCertFile                = "/certs/localhost.crt"
	defaultKeyFile                 = "/certs/localhost.key"
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
	updateMetricsEnabled(config, logger)
	updateTickIntervals(config, logger)
	updateService(powerStoreSvc, logger)
	updateTracing(logger)

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		updateLoggingSettings(logger)
		updateCollectorAddress(config, exporter, logger)
		updateProvisionerNames(volumeFinder, logger)
		updateMetricsEnabled(config, logger)
		updateTickIntervals(config, logger)
		updateService(powerStoreSvc, logger)
		updateTracing(logger)
	})

	configFileListener.WatchConfig()
	configFileListener.OnConfigChange(func(e fsnotify.Event) {
		updatePowerStoreConnection(powerStoreSvc, logger)
	})

	viper.SetDefault("TLS_CERT_PATH", defaultCertFile)
	viper.SetDefault("TLS_KEY_PATH", defaultKeyFile)
	viper.SetDefault("PORT", defaultDebugPort)

	// TLS_CERT_PATH is only read as an environment variable
	certFile := viper.GetString("TLS_CERT_PATH")

	// TLS_KEY_PATH is only read as an environment variable
	keyFile := viper.GetString("TLS_KEY_PATH")

	var bindPort int
	// PORT is only read as an environment variable
	portEnv := viper.GetString("PORT")
	if portEnv != "" {
		var err error
		if bindPort, err = strconv.Atoi(portEnv); err != nil {
			logger.WithError(err).WithField("port", portEnv).Fatal("port value is invalid")
		}
	}

	go func() {
		expvar.NewString("service").Set("metrics-powerstore")
		expvar.Publish("goroutines", expvar.Func(func() interface{} {
			return fmt.Sprintf("%d", runtime.NumGoroutine())
		}))
		s := http.Server{
			Addr:    fmt.Sprintf(":%d", bindPort),
			Handler: http.DefaultServeMux,
		}
		if err := s.ListenAndServeTLS(certFile, keyFile); err != nil {
			logger.WithError(err).Error("debug listener closed")
		}
	}()

	if err := entrypoint.Run(context.Background(), config, exporter, powerStoreSvc); err != nil {
		logger.WithError(err).Fatal("running service")
	}
}

func updateTracing(logger *logrus.Logger) {
	zipkinURI := viper.GetString("ZIPKIN_URI")
	zipkinServiceName := viper.GetString("ZIPKIN_SERVICE_NAME")
	zipkinProbability := viper.GetFloat64("ZIPKIN_PROBABILITY")

	tp, err := tracer.InitTracing(zipkinURI, zipkinServiceName, zipkinProbability)
	if err != nil {
		logger.WithError(err).Error("initializing tracer")
	}
	if tp != nil {
		logger.WithFields(logrus.Fields{"uri": zipkinURI,
			"service_name": zipkinServiceName,
			"probablity":   zipkinProbability,
		}).Infof("setting zipkin tracing")
		global.SetTraceProvider(tp)
	}
}

func updatePowerStoreConnection(powerStoreSvc *service.PowerStoreService, logger *logrus.Logger) {
	f := &fs.Fs{Util: &gofsutil.FS{}}
	arrays, _, _, err := array.GetPowerStoreArrays(f, defaultStorageSystemConfigFile)
	if err != nil {
		logger.WithError(err).Fatal("initialize arrays in controller service")
	}
	powerStoreClients := make(map[string]service.PowerStoreClient)
	for arrayIP, client := range arrays {
		powerStoreClients[arrayIP] = client.Client
		logger.WithField("array_ip", arrayIP).Debug("setting powerstore client from configuration")
	}
	powerStoreSvc.PowerStoreClients = powerStoreClients
}

func updateCollectorAddress(config *entrypoint.Config, exporter *otlexporters.OtlCollectorExporter, logger *logrus.Logger) {
	collectorAddress := viper.GetString("COLLECTOR_ADDR")
	if collectorAddress == "" {
		logger.Fatal("COLLECTOR_ADDR is required")
	}
	config.CollectorAddress = collectorAddress
	exporter.CollectorAddr = collectorAddress
	logger.WithField("collector_address", collectorAddress).Debug("setting collector address")
}

func updateProvisionerNames(volumeFinder *k8s.VolumeFinder, logger *logrus.Logger) {
	provisionerNamesValue := viper.GetString("provisioner_names")
	if provisionerNamesValue == "" {
		logger.Fatal("PROVISIONER_NAMES is required")
	}
	provisionerNames := strings.Split(provisionerNamesValue, ",")
	volumeFinder.DriverNames = provisionerNames
	logger.WithField("provisioner_names", provisionerNamesValue).Debug("setting provisioner names")
}

func updateMetricsEnabled(config *entrypoint.Config, logger *logrus.Logger) {
	powerstoreVolumeMetricsEnabled := true
	powerstoreVolumeMetricsEnabledValue := viper.GetString("POWERSTORE_VOLUME_METRICS_ENABLED")
	if powerstoreVolumeMetricsEnabledValue == "false" {
		powerstoreVolumeMetricsEnabled = false
	}
	config.VolumeMetricsEnabled = powerstoreVolumeMetricsEnabled
	logger.WithField("volume_metrics_enabled", powerstoreVolumeMetricsEnabled).Debug("setting volume metrics enabled")
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
	logger.WithField("volume_tick_interval", fmt.Sprintf("%v", volumeTickInterval)).Debug("setting volume tick interval")

	spaceTickInterval := defaultTickInterval
	spacePollFrequencySeconds := viper.GetString("POWERSTORE_SPACE_POLL_FREQUENCY")
	if spacePollFrequencySeconds != "" {
		numSeconds, err := strconv.Atoi(spacePollFrequencySeconds)
		if err != nil {
			logger.WithError(err).Fatal("POWERSTORE_SPACE_POLL_FREQUENCY was not set to a valid number")
		}
		spaceTickInterval = time.Duration(numSeconds) * time.Second
	}
	config.SpaceTickInterval = spaceTickInterval
	logger.WithField("space_tick_interval", fmt.Sprintf("%v", spaceTickInterval)).Debug("setting space tick interval")

	arrayTickInterval := defaultTickInterval
	arrayPollFrequencySeconds := viper.GetString("POWERSTORE_ARRAY_POLL_FREQUENCY")
	if arrayPollFrequencySeconds != "" {
		numSeconds, err := strconv.Atoi(spacePollFrequencySeconds)
		if err != nil {
			logger.WithError(err).Fatal("POWERSTORE_ARRAY_POLL_FREQUENCY was not set to a valid number")
		}
		arrayTickInterval = time.Duration(numSeconds) * time.Second
	}
	config.ArrayTickInterval = arrayTickInterval
	logger.WithField("array_tick_interval", fmt.Sprintf("%v", arrayTickInterval)).Debug("setting array tick interval")
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
	logger.WithField("max_connections", maxPowerStoreConcurrentRequests).Debug("setting max powerstore connections")
}
