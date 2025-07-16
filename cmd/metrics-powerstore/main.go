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

package main

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/dell/csm-metrics-powerstore/internal/entrypoint"
	"github.com/dell/csm-metrics-powerstore/internal/k8s"
	"github.com/dell/csm-metrics-powerstore/internal/pstoreresource"
	"github.com/dell/csm-metrics-powerstore/internal/service"
	otlexporters "github.com/dell/csm-metrics-powerstore/opentelemetry/exporters"
	tracer "github.com/dell/csm-metrics-powerstore/opentelemetry/tracers"
	"github.com/sirupsen/logrus"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
)

const (
	defaultTickInterval            = 20 * time.Second
	defaultConfigFile              = "/etc/config/karavi-metrics-powerstore.yaml"
	defaultStorageSystemConfigFile = "/powerstore-config/config"
	defaultDebugPort               = "9090"
	defaultCertFile                = "/certs/localhost.crt"
	defaultKeyFile                 = "/certs/localhost.key"
)

// getPowerStoreArrays is a wrapper for pstoreresource.GetPowerStoreArrays
var getPowerStoreArrays = pstoreresource.GetPowerStoreArrays

// initTracing is a wrapper for tracer.InitTracing
var initTracing = tracer.InitTracing

func main() {
	logger, config, powerStoreSvc, exporter := initializeConfig()

	startConfigWatchers(logger, config, exporter, powerStoreSvc)
	startHTTPServer(logger)

	if err := entrypoint.Run(context.Background(), config, exporter, powerStoreSvc); err != nil {
		logger.WithError(err).Fatal("running service")
	}
}

func initializeConfig() (*logrus.Logger, *entrypoint.Config, *service.PowerStoreService, *otlexporters.OtlCollectorExporter) {
	logger := logrus.New()
	exporter := &otlexporters.OtlCollectorExporter{}
	viper.SetConfigFile(defaultConfigFile)
	err := viper.ReadInConfig()
	// if unable to read configuration file, proceed in case we use environment variables
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to read Config file: %v", err)
	}

	leaderElectorGetter := &k8s.LeaderElector{API: &k8s.LeaderElector{}}
	collectorCertPath := getCollectorCertPath()

	config := &entrypoint.Config{
		LeaderElector:     leaderElectorGetter,
		CollectorCertPath: collectorCertPath,
		Logger:            logger,
	}

	volumeFinder := &k8s.VolumeFinder{API: &k8s.API{}, Logger: logger}

	powerStoreSvc := &service.PowerStoreService{
		MetricsWrapper: &service.MetricsWrapper{Meter: otel.Meter("powerstore")},
		Logger:         logger,
		VolumeFinder:   volumeFinder,
	}

	updatePowerStoreConnection(powerStoreSvc, logger)
	applyInitialConfig(logger, config, exporter, powerStoreSvc, volumeFinder)

	return logger, config, powerStoreSvc, exporter
}

func applyInitialConfig(logger *logrus.Logger, config *entrypoint.Config, exporter *otlexporters.OtlCollectorExporter, powerStoreSvc *service.PowerStoreService, volumeFinder *k8s.VolumeFinder) {
	updateLoggingSettings(logger)
	updateCollectorAddress(config, exporter, logger)
	updateProvisionerNames(volumeFinder, logger)
	updateMetricsEnabled(config, logger)
	updateTickIntervals(config, logger)
	updateService(powerStoreSvc, logger)
	updateTracing(logger)
}

var updateLoggingSettings = func(logger *logrus.Logger) {
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

func getCollectorCertPath() string {
	if tls := os.Getenv("TLS_ENABLED"); tls == "true" {
		collectorCertPath := os.Getenv("COLLECTOR_CERT_PATH")
		if len(strings.TrimSpace(collectorCertPath)) < 1 {
			return otlexporters.DefaultCollectorCertPath
		}
		return collectorCertPath
	}
	return ""
}

func startConfigWatchers(logger *logrus.Logger, config *entrypoint.Config, exporter *otlexporters.OtlCollectorExporter, powerStoreSvc *service.PowerStoreService) {
	viper.WatchConfig()
	volumeFinder := &k8s.VolumeFinder{
		API:    &k8s.API{},
		Logger: logger,
	}
	viper.OnConfigChange(func(_ fsnotify.Event) {
		applyInitialConfig(logger, config, exporter, powerStoreSvc, volumeFinder)
	})

	configFileListener := viper.New()
	configFileListener.SetConfigFile(defaultStorageSystemConfigFile)
	configFileListener.WatchConfig()
	configFileListener.OnConfigChange(func(_ fsnotify.Event) {
		updatePowerStoreConnection(powerStoreSvc, logger)
	})
}

func startHTTPServer(logger *logrus.Logger) {
	viper.SetDefault("TLS_CERT_PATH", defaultCertFile)
	viper.SetDefault("TLS_KEY_PATH", defaultKeyFile)
	viper.SetDefault("PORT", defaultDebugPort)

	// TLS_CERT_PATH is only read as an environment variable
	certFile := viper.GetString("TLS_CERT_PATH")

	// TLS_KEY_PATH is only read as an environment variable
	keyFile := viper.GetString("TLS_KEY_PATH")

	bindPort := getBindPort(logger)

	go func() {
		expvar.NewString("service").Set("metrics-powerstore")
		expvar.Publish("goroutines", expvar.Func(func() interface{} {
			return fmt.Sprintf("%d", runtime.NumGoroutine())
		}))
		s := http.Server{
			Addr:              fmt.Sprintf(":%d", bindPort),
			Handler:           http.DefaultServeMux,
			ReadHeaderTimeout: 5 * time.Second,
		}
		if err := s.ListenAndServeTLS(certFile, keyFile); err != nil {
			logger.WithError(err).Error("debug listener closed")
		}
	}()
}

func getBindPort(logger *logrus.Logger) int {
	portEnv := viper.GetString("PORT")
	if portEnv != "" {
		bindPort, err := strconv.Atoi(portEnv)
		if err != nil {
			logger.WithError(err).WithField("port", portEnv).Fatal("port value is invalid")
		}
		return bindPort
	}
	return 0
}

func updateTracing(logger *logrus.Logger) {
	zipkinURI := viper.GetString("ZIPKIN_URI")
	zipkinServiceName := viper.GetString("ZIPKIN_SERVICE_NAME")
	zipkinProbability := viper.GetFloat64("ZIPKIN_PROBABILITY")

	tp, err := initTracing(zipkinURI, zipkinProbability)
	if err != nil {
		logger.WithError(err).Error("initializing tracer")
	}
	if tp != nil {
		logger.WithFields(logrus.Fields{
			"uri":          zipkinURI,
			"service_name": zipkinServiceName,
			"probablity":   zipkinProbability,
		}).Infof("setting zipkin tracing")
		otel.SetTracerProvider(tp)
	}
}

func updatePowerStoreConnection(powerStoreSvc *service.PowerStoreService, logger *logrus.Logger) {
	arrays, _, _, err := getPowerStoreArrays(defaultStorageSystemConfigFile, logger)
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

	powerstoreTopologyMetricsEnabled := true
	powerstoreTopologyMetricsEnabledValue := viper.GetString("POWERSTORE_TOPOLOGY_METRICS_ENABLED")
	if powerstoreTopologyMetricsEnabledValue == "false" {
		powerstoreTopologyMetricsEnabled = false
	}
	config.TopologyMetricsEnabled = powerstoreTopologyMetricsEnabled
	logger.WithField("topology_metrics_enabled", powerstoreTopologyMetricsEnabled).Debug("setting topology metrics enabled")
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
		numSeconds, err := strconv.Atoi(arrayPollFrequencySeconds)
		if err != nil {
			logger.WithError(err).Fatal("POWERSTORE_ARRAY_POLL_FREQUENCY was not set to a valid number")
		}
		arrayTickInterval = time.Duration(numSeconds) * time.Second
	}
	config.ArrayTickInterval = arrayTickInterval
	logger.WithField("array_tick_interval", fmt.Sprintf("%v", arrayTickInterval)).Debug("setting array tick interval")

	fileSystemTickInterval := defaultTickInterval
	fileSystemPollFrequencySeconds := viper.GetString("POWERSTORE_FILE_SYSTEM_POLL_FREQUENCY")
	if fileSystemPollFrequencySeconds != "" {
		numSeconds, err := strconv.Atoi(fileSystemPollFrequencySeconds)
		if err != nil {
			logger.WithError(err).Fatal("POWERSTORE_FILE_SYSTEM_POLL_FREQUENCY was not set to a valid number")
		}
		fileSystemTickInterval = time.Duration(numSeconds) * time.Second
	}
	config.FileSystemTickInterval = fileSystemTickInterval
	logger.WithField("file_tick_interval", fmt.Sprintf("%v", fileSystemTickInterval)).Debug("setting filesystem tick interval")

	topologyTickInterval := defaultTickInterval
	topologyPollFrequencySeconds := viper.GetString("POWERSTORE_TOPOLOGY_METRICS_POLL_FREQUENCY")
	if topologyPollFrequencySeconds != "" {
		numSeconds, err := strconv.Atoi(topologyPollFrequencySeconds)
		if err != nil {
			logger.WithError(err).Fatal("POWERSTORE_TOPOLOGY_METRICS_POLL_FREQUENCY was not set to a valid number")
		}
		topologyTickInterval = time.Duration(numSeconds) * time.Second
	}
	config.TopologyTickInterval = topologyTickInterval
	logger.WithField("topology_tick_interval", fmt.Sprintf("%v", topologyTickInterval)).Debug("setting topology tick interval")
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
