/*
 Copyright (c) 2025-2026 Dell Inc. or its subsidiaries. All Rights Reserved.

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

	csmserver "github.com/dell/csm-metrics-common/pkg/server"
	"github.com/dell/csm-metrics-powerstore/internal/entrypoint"
	"github.com/dell/csm-metrics-powerstore/internal/k8s"
	"github.com/dell/csm-metrics-powerstore/internal/pstoreresource"
	"github.com/dell/csm-metrics-powerstore/internal/service"
	otlexporters "github.com/dell/csm-metrics-powerstore/opentelemetry/exporters"
	tracer "github.com/dell/csm-metrics-powerstore/opentelemetry/tracers"
	"github.com/dell/csmlog"

	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
)

const (
	defaultTickInterval            = 300 * time.Second
	defaultConfigFile              = "/etc/config/karavi-metrics-powerstore.yaml"
	defaultStorageSystemConfigFile = "/powerstore-config/config"
	defaultDebugPort               = "9090"
	defaultCertFile                = "/certs/localhost.crt"
	defaultKeyFile                 = "/certs/localhost.key"
	defaultObsMetricsPort          = "8443"
	defaultObsMetricsScheme        = "http"
	defaultObsMetricsCert          = "/etc/metrics-tls/tls.crt"
	defaultObsMetricsKey           = "/etc/metrics-tls/tls.key"
)

const (
	csiObsMetricsEnabledKey = "X_CSI_METRICS_ENABLED"
	csiObsMetricsPortKey    = "X_CSI_METRICS_PORT"
	csiObsMetricsCertKey    = "X_CSI_METRICS_TLS_CERT_FILE"
	csiObsMetricsKeyKey     = "X_CSI_METRICS_TLS_KEY_FILE"
)

// getPowerStoreArrays is a wrapper for pstoreresource.GetPowerStoreArrays
var getPowerStoreArrays = pstoreresource.GetPowerStoreArrays

// initTracing is a wrapper for tracer.InitTracing
var initTracing = tracer.InitTracing

type exportFailureRecorder interface {
	SetExportFailureRecorder(func())
}

func wirePowerStoreExportFailureRecorder(powerStoreSvc *service.PowerStoreService, exporter exportFailureRecorder) {
	if exporter == nil {
		return
	}
	exporter.SetExportFailureRecorder(func() {
		if powerStoreSvc != nil && powerStoreSvc.ObsInstrumenter != nil {
			powerStoreSvc.ObsInstrumenter.RecordExportSuccess(obsGlobalID(powerStoreSvc), "failure")
		}
	})
}

func main() {
	config, powerStoreSvc, exporter := initializeConfig()

	wirePowerStoreExportFailureRecorder(powerStoreSvc, exporter)
	startConfigWatchers(config, exporter, powerStoreSvc)
	startHTTPServer()
	if strings.EqualFold(viper.GetString(csiObsMetricsEnabledKey), "true") {
		startMetricsServer(powerStoreSvc, exporter)
	}

	if err := entrypoint.Run(context.Background(), config, exporter, powerStoreSvc); err != nil {
		csmlog.WithFields(csmlog.Fields{
			"error": err,
		}).Fatal("running service")
	}
}

func initializeConfig() (*entrypoint.Config, *service.PowerStoreService, *otlexporters.OtlCollectorExporter) {
	exporter := &otlexporters.OtlCollectorExporter{}
	viper.SetConfigFile(defaultConfigFile)
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	// if unable to read configuration file, proceed in case we use environment variables
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to read Config file: %v", err)
	}

	updateLoggingSettings()

	leaderElectorGetter := &k8s.LeaderElector{API: &k8s.LeaderElector{}}
	collectorCertPath := getCollectorCertPath()

	config := &entrypoint.Config{
		LeaderElector:     leaderElectorGetter,
		CollectorCertPath: collectorCertPath,
	}

	volumeFinder := &k8s.VolumeFinder{API: &k8s.API{}}

	powerStoreSvc := &service.PowerStoreService{
		MetricsWrapper: &service.MetricsWrapper{Meter: otel.Meter("powerstore")},
		VolumeFinder:   volumeFinder,
	}

	updatePowerStoreConnection(powerStoreSvc)
	applyInitialConfig(config, exporter, powerStoreSvc, volumeFinder)

	return config, powerStoreSvc, exporter
}

func applyInitialConfig(config *entrypoint.Config, exporter *otlexporters.OtlCollectorExporter, powerStoreSvc *service.PowerStoreService, volumeFinder *k8s.VolumeFinder) {
	updateCollectorAddress(config, exporter)
	updateProvisionerNames(volumeFinder)
	updateMetricsEnabled(config)
	updateTickIntervals(config)
	updateService(powerStoreSvc)
	updateTracing()
	updateLoggingSettings()
}

var updateLoggingSettings = func() {
	logFormat := viper.GetString("LOG_FORMAT")
	if strings.EqualFold(logFormat, "json") {
		csmlog.SetFormat("json")
	} else {
		// use text formatter by default
		csmlog.SetFormat("text")
	}
	logLevel := viper.GetString("LOG_LEVEL")
	level, err := csmlog.ParseLevel(logLevel)
	if err != nil {
		// use INFO level by default
		level = csmlog.InfoLevel
	}
	csmlog.SetLevel(level)
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

func startConfigWatchers(config *entrypoint.Config, exporter *otlexporters.OtlCollectorExporter, powerStoreSvc *service.PowerStoreService) {
	viper.WatchConfig()
	volumeFinder := &k8s.VolumeFinder{
		API: &k8s.API{},
	}
	viper.OnConfigChange(func(_ fsnotify.Event) {
		applyInitialConfig(config, exporter, powerStoreSvc, volumeFinder)
	})

	configFileListener := viper.New()
	configFileListener.SetConfigFile(defaultStorageSystemConfigFile)
	configFileListener.WatchConfig()
	configFileListener.OnConfigChange(func(_ fsnotify.Event) {
		updatePowerStoreConnection(powerStoreSvc)
	})
}

func startMetricsServer(powerStoreSvc *service.PowerStoreService, _ *otlexporters.OtlCollectorExporter) {
	reg := prometheus.NewRegistry()
	powerStoreSvc.ObsInstrumenter = service.NewPSTObsInstrumenter(reg)

	// Read X_CSI_OBS_METRICS_PORT from env, default to 8080
	viper.SetDefault(csiObsMetricsPortKey, defaultObsMetricsPort)
	metricsPort := viper.GetString(csiObsMetricsPortKey)

	// Infer the metrics scheme from cert/key env vars: if either is provided,
	// use HTTPS; otherwise fall back to HTTP.
	certFile := strings.TrimSpace(viper.GetString(csiObsMetricsCertKey))
	keyFile := strings.TrimSpace(viper.GetString(csiObsMetricsKeyKey))
	scheme := defaultObsMetricsScheme
	if certFile != "" || keyFile != "" {
		scheme = "https"
	}

	// Resolve TLS cert/key paths and validate them BEFORE launching the goroutine.
	// This ensures configuration errors are detected synchronously and can safely
	// call Fatal from the main goroutine, preventing a running service with a
	// non-functional metrics endpoint.
	if scheme == "https" {
		if certFile == "" {
			certFile = defaultObsMetricsCert
		}
		if keyFile == "" {
			keyFile = defaultObsMetricsKey
		}

		if err := validateTLSFiles(certFile, keyFile); err != nil {
			csmlog.WithFields(csmlog.Fields{
				"error": err,
				"cert":  certFile,
				"key":   keyFile,
			}).Fatal("observability metrics server failed to start: invalid TLS configuration")
			return
		}
	}

	srv := csmserver.NewMetricsServer(csmserver.Config{
		Port:     fmt.Sprintf(":%s", metricsPort),
		CertFile: certFile,
		KeyFile:  keyFile,
		Registry: reg,
	})

	go func() {
		csmlog.WithFields(csmlog.Fields{
			"port":   metricsPort,
			"scheme": scheme,
		}).Info("starting observability metrics server")
		if err := srv.Start(); err != nil {
			csmlog.WithFields(csmlog.Fields{"error": err}).Error("observability metrics server closed")
		}
	}()
}

func obsGlobalID(powerStoreSvc *service.PowerStoreService) string {
	if powerStoreSvc != nil && powerStoreSvc.DefaultPowerStoreArray != nil && powerStoreSvc.DefaultPowerStoreArray.GlobalID != "" {
		return powerStoreSvc.DefaultPowerStoreArray.GlobalID
	}
	return "powerstore"
}

func validateTLSFiles(certFile, keyFile string) error {
	if certFile == "" {
		return fmt.Errorf("TLS certificate file path is empty")
	}
	if keyFile == "" {
		return fmt.Errorf("TLS key file path is empty")
	}

	// Check if cert file exists and is accessible
	if _, err := os.Stat(certFile); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("TLS certificate file does not exist: %s", certFile)
		}
		return fmt.Errorf("failed to access TLS certificate file %s: %v", certFile, err)
	}

	// Check if key file exists and is accessible
	if _, err := os.Stat(keyFile); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("TLS key file does not exist: %s", keyFile)
		}
		return fmt.Errorf("failed to access TLS key file %s: %v", keyFile, err)
	}

	return nil
}

func startHTTPServer() {
	viper.SetDefault("TLS_CERT_PATH", defaultCertFile)
	viper.SetDefault("TLS_KEY_PATH", defaultKeyFile)
	viper.SetDefault("PORT", defaultDebugPort)

	// TLS_CERT_PATH is only read as an environment variable
	certFile := viper.GetString("TLS_CERT_PATH")

	// TLS_KEY_PATH is only read as an environment variable
	keyFile := viper.GetString("TLS_KEY_PATH")

	bindPort := getBindPort()

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

		// Validate TLS certificate files before attempting to start the server.
		// Use Error (not Fatal) here — the debug listener is non-critical and its
		// failure should not crash the main service.
		if err := validateTLSFiles(certFile, keyFile); err != nil {
			csmlog.WithFields(csmlog.Fields{
				"error": err,
				"cert":  certFile,
				"key":   keyFile,
			}).Error("debug listener failed to start: invalid TLS configuration")
			return
		}

		if err := s.ListenAndServeTLS(certFile, keyFile); err != nil {
			csmlog.WithFields(csmlog.Fields{
				"error": err,
			}).Error("debug listener closed")
		}
	}()
}

func getBindPort() int {
	portEnv := viper.GetString("PORT")
	if portEnv != "" {
		bindPort, err := strconv.Atoi(portEnv)
		if err != nil {
			csmlog.WithFields(csmlog.Fields{
				"error": err,
				"port":  portEnv,
			}).Fatal("port value is invalid")
		}
		return bindPort
	}
	return 0
}

func updateTracing() {
	zipkinURI := viper.GetString("ZIPKIN_URI")
	zipkinServiceName := viper.GetString("ZIPKIN_SERVICE_NAME")
	zipkinProbability := viper.GetFloat64("ZIPKIN_PROBABILITY")

	tp, err := initTracing(zipkinURI, zipkinProbability)
	if err != nil {
		csmlog.WithFields(csmlog.Fields{
			"error": err,
		}).Error("initializing tracer")
	}
	if tp != nil {
		csmlog.WithFields(csmlog.Fields{
			"uri":          zipkinURI,
			"service_name": zipkinServiceName,
			"probability":  zipkinProbability,
		}).Info("setting zipkin tracing")
		otel.SetTracerProvider(tp)
	}
}

func updatePowerStoreConnection(powerStoreSvc *service.PowerStoreService) {
	arrays, _, defaultArray, err := getPowerStoreArrays(defaultStorageSystemConfigFile)
	if err != nil {
		csmlog.WithFields(csmlog.Fields{
			"error": err,
		}).Fatal("initialize arrays in controller service")
	}
	powerStoreClients := make(map[string]service.PowerStoreClient)
	for arrayIP, client := range arrays {
		powerStoreClients[arrayIP] = client.Client
		csmlog.WithFields(csmlog.Fields{
			"array_ip": arrayIP,
		}).Debug("setting powerstore client from configuration")
	}
	powerStoreSvc.PowerStoreClients = powerStoreClients
	powerStoreSvc.PowerStoreArrays = arrays
	powerStoreSvc.DefaultPowerStoreArray = defaultArray
}

func updateCollectorAddress(config *entrypoint.Config, exporter *otlexporters.OtlCollectorExporter) {
	collectorAddress := viper.GetString("COLLECTOR_ADDR")
	if collectorAddress == "" {
		csmlog.Fatal("COLLECTOR_ADDR is required")
	}
	config.CollectorAddress = collectorAddress
	exporter.CollectorAddr = collectorAddress
	csmlog.WithFields(csmlog.Fields{
		"collector_address": collectorAddress,
	}).Debug("setting collector address")
}

func updateProvisionerNames(volumeFinder *k8s.VolumeFinder) {
	provisionerNamesValue := viper.GetString("provisioner_names")
	if provisionerNamesValue == "" {
		csmlog.Fatal("PROVISIONER_NAMES is required")
	}
	provisionerNames := strings.Split(provisionerNamesValue, ",")
	volumeFinder.DriverNames = provisionerNames
	csmlog.WithFields(csmlog.Fields{
		"provisioner_names": provisionerNamesValue,
	}).Debug("setting provisioner names")
}

func updateMetricsEnabled(config *entrypoint.Config) {
	powerstoreVolumeMetricsEnabled := true
	powerstoreVolumeMetricsEnabledValue := viper.GetString("POWERSTORE_VOLUME_METRICS_ENABLED")
	if powerstoreVolumeMetricsEnabledValue == "false" {
		powerstoreVolumeMetricsEnabled = false
	}
	config.VolumeMetricsEnabled = powerstoreVolumeMetricsEnabled
	csmlog.WithFields(csmlog.Fields{
		"volume_metrics_enabled": powerstoreVolumeMetricsEnabled,
	}).Debug("setting volume metrics enabled")

	powerstoreTopologyMetricsEnabled := true
	powerstoreTopologyMetricsEnabledValue := viper.GetString("POWERSTORE_TOPOLOGY_METRICS_ENABLED")
	if powerstoreTopologyMetricsEnabledValue == "false" {
		powerstoreTopologyMetricsEnabled = false
	}
	config.TopologyMetricsEnabled = powerstoreTopologyMetricsEnabled
	csmlog.WithFields(csmlog.Fields{
		"topology_metrics_enabled": powerstoreTopologyMetricsEnabled,
	}).Debug("setting topology metrics enabled")
}

func updateTickIntervals(config *entrypoint.Config) {
	volumeTickInterval := defaultTickInterval
	volIoPollFrequencySeconds := viper.GetString("POWERSTORE_VOLUME_IO_POLL_FREQUENCY")
	if volIoPollFrequencySeconds != "" {
		numSeconds, err := strconv.Atoi(volIoPollFrequencySeconds)
		if err != nil {
			csmlog.WithFields(csmlog.Fields{
				"error": err,
			}).Fatal("POWERSTORE_VOLUME_IO_POLL_FREQUENCY was not set to a valid number")
		}
		volumeTickInterval = time.Duration(numSeconds) * time.Second
	}
	config.VolumeTickInterval = volumeTickInterval
	csmlog.WithFields(csmlog.Fields{
		"volume_tick_interval": fmt.Sprintf("%v", volumeTickInterval),
	}).Debug("setting volume tick interval")

	spaceTickInterval := defaultTickInterval
	spacePollFrequencySeconds := viper.GetString("POWERSTORE_SPACE_POLL_FREQUENCY")
	if spacePollFrequencySeconds != "" {
		numSeconds, err := strconv.Atoi(spacePollFrequencySeconds)
		if err != nil {
			csmlog.WithFields(csmlog.Fields{
				"error": err,
			}).Fatal("POWERSTORE_SPACE_POLL_FREQUENCY was not set to a valid number")
		}
		spaceTickInterval = time.Duration(numSeconds) * time.Second
	}
	config.SpaceTickInterval = spaceTickInterval
	csmlog.WithFields(csmlog.Fields{
		"space_tick_interval": fmt.Sprintf("%v", spaceTickInterval),
	}).Debug("setting space tick interval")

	arrayTickInterval := defaultTickInterval
	arrayPollFrequencySeconds := viper.GetString("POWERSTORE_ARRAY_POLL_FREQUENCY")
	if arrayPollFrequencySeconds != "" {
		numSeconds, err := strconv.Atoi(arrayPollFrequencySeconds)
		if err != nil {
			csmlog.WithFields(csmlog.Fields{
				"error": err,
			}).Fatal("POWERSTORE_ARRAY_POLL_FREQUENCY was not set to a valid number")
		}
		arrayTickInterval = time.Duration(numSeconds) * time.Second
	}
	config.ArrayTickInterval = arrayTickInterval
	csmlog.WithFields(csmlog.Fields{
		"array_tick_interval": fmt.Sprintf("%v", arrayTickInterval),
	}).Debug("setting array tick interval")

	fileSystemTickInterval := defaultTickInterval
	fileSystemPollFrequencySeconds := viper.GetString("POWERSTORE_FILE_SYSTEM_POLL_FREQUENCY")
	if fileSystemPollFrequencySeconds != "" {
		numSeconds, err := strconv.Atoi(fileSystemPollFrequencySeconds)
		if err != nil {
			csmlog.WithFields(csmlog.Fields{
				"error": err,
			}).Fatal("POWERSTORE_FILE_SYSTEM_POLL_FREQUENCY was not set to a valid number")
		}
		fileSystemTickInterval = time.Duration(numSeconds) * time.Second
	}
	config.FileSystemTickInterval = fileSystemTickInterval
	csmlog.WithFields(csmlog.Fields{
		"file_tick_interval": fmt.Sprintf("%v", fileSystemTickInterval),
	}).Debug("setting filesystem tick interval")

	topologyTickInterval := defaultTickInterval
	topologyPollFrequencySeconds := viper.GetString("POWERSTORE_TOPOLOGY_METRICS_POLL_FREQUENCY")
	if topologyPollFrequencySeconds != "" {
		numSeconds, err := strconv.Atoi(topologyPollFrequencySeconds)
		if err != nil {
			csmlog.WithFields(csmlog.Fields{
				"error": err,
			}).Fatal("POWERSTORE_TOPOLOGY_METRICS_POLL_FREQUENCY was not set to a valid number")
		}
		topologyTickInterval = time.Duration(numSeconds) * time.Second
	}
	config.TopologyTickInterval = topologyTickInterval
	csmlog.WithFields(csmlog.Fields{
		"topology_tick_interval": fmt.Sprintf("%v", topologyTickInterval),
	}).Debug("setting topology tick interval")
}

func updateService(pstoreSvc *service.PowerStoreService) {
	maxPowerStoreConcurrentRequests := service.DefaultMaxPowerStoreConnections
	maxPowerStoreConcurrentRequestsVar := viper.GetString("POWERSTORE_MAX_CONCURRENT_QUERIES")
	if maxPowerStoreConcurrentRequestsVar != "" {
		var err error
		maxPowerStoreConcurrentRequests, err = strconv.Atoi(maxPowerStoreConcurrentRequestsVar)
		if err != nil {
			csmlog.WithFields(csmlog.Fields{
				"error": err,
			}).Fatal("POWERSTORE_MAX_CONCURRENT_QUERIES was not set to a valid number")
		}
		if maxPowerStoreConcurrentRequests <= 0 {
			csmlog.WithFields(csmlog.Fields{
				"value": maxPowerStoreConcurrentRequests,
			}).Fatal("POWERSTORE_MAX_CONCURRENT_QUERIES value was invalid (<= 0)")
		}
	}
	pstoreSvc.MaxPowerStoreConnections = maxPowerStoreConcurrentRequests
	csmlog.WithFields(csmlog.Fields{
		"max_connections": maxPowerStoreConcurrentRequests,
	}).Debug("setting max powerstore connections")
}
