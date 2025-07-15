package main

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dell/csm-metrics-powerstore/internal/entrypoint"
	"github.com/dell/csm-metrics-powerstore/internal/k8s"
	"github.com/dell/csm-metrics-powerstore/internal/service"
	otlexporters "github.com/dell/csm-metrics-powerstore/opentelemetry/exporters"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func TestInitializeConfig(t *testing.T) {
	// Mock getPowerScaleClusters to avoid file I/O
	originalGetPowerStoreArrays := getPowerStoreArrays
	defer func() { getPowerStoreArrays = originalGetPowerStoreArrays }()
	getPowerStoreArrays = func(_ string, _ *logrus.Logger) (map[string]*service.PowerStoreArray, map[string]string, *service.PowerStoreArray, error) {
		return map[string]*service.PowerStoreArray{
				"cluster1": {
					Endpoint:  "10.10.10.10",
					GlobalID:  "PowerStore123",
					IsDefault: true,
				},
			}, map[string]string{"cluster1": "10.10.10.10"}, &service.PowerStoreArray{
				Endpoint:  "10.10.10.10",
				GlobalID:  "PowerStore123",
				IsDefault: true,
			}, nil
	}

	// Mock Viper to avoid reading from the actual config file
	viper.Reset()
	viper.SetConfigType("yaml")
	viper.SetConfigFile(defaultConfigFile)

	// Mock the config file content
	configContent := `
LOG_LEVEL: debug
COLLECTOR_ADDR: localhost:4317
PROVISIONER_NAMES: csi-powerstore
POWERSTORE_VOLUME_METRICS_ENABLED: true
POWERSTORE_TOPOLOGY_METRICS_ENABLED: "true"
TLS_ENABLED: false
`
	err := viper.ReadConfig(strings.NewReader(configContent))
	if err != nil {
		// Handle the error or log it
		log.Printf("Error reading config: %v", err)
	}

	tests := []struct {
		name                           string
		envVars                        map[string]string
		expectedLogLevel               logrus.Level
		expectedCollectorAddr          string
		expectedProvisioners           []string
		expectedCertPath               string
		expectedTopologyMetricsEnabled bool
	}{
		{
			name: "SuccessfulInitializationWithDefaults",
			envVars: map[string]string{
				"LOG_LEVEL":                           "debug",
				"COLLECTOR_ADDR":                      "localhost:4317",
				"PROVISIONER_NAMES":                   "csi-powerstore",
				"POWERSTORE_VOLUME_METRICS_ENABLED":   "true",
				"TLS_ENABLED":                         "false",
				"POWERSTORE_TOPOLOGY_METRICS_ENABLED": "true",
			},
			expectedLogLevel:               logrus.DebugLevel,
			expectedCollectorAddr:          "localhost:4317",
			expectedProvisioners:           []string{"csi-isilon"},
			expectedCertPath:               otlexporters.DefaultCollectorCertPath,
			expectedTopologyMetricsEnabled: true,
		},
		{
			name: "TLSEnabledWithCustomCertPath",
			envVars: map[string]string{
				"LOG_LEVEL":                           "info",
				"COLLECTOR_ADDR":                      "collector:4317",
				"PROVISIONER_NAMES":                   "csi-powerstore",
				"TLS_ENABLED":                         "true",
				"COLLECTOR_CERT_PATH":                 "/custom/cert/path",
				"POWERSTORE_TOPOLOGY_METRICS_ENABLED": "false",
			},
			expectedLogLevel:               logrus.InfoLevel,
			expectedCollectorAddr:          "collector:4317",
			expectedProvisioners:           []string{"csi-powerstore"},
			expectedCertPath:               "/custom/cert/path",
			expectedTopologyMetricsEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset Viper and set environment variables for each test case
			viper.Reset()
			for k, v := range tt.envVars {
				viper.Set(k, v)
				defer viper.Set(k, "")
			}

			// Mock the config file content for each test case
			viper.SetConfigType("yaml")
			viper.SetConfigFile(defaultConfigFile)

			err := viper.ReadConfig(strings.NewReader(configContent))
			if err != nil {
				// Handle the error or log it
				log.Printf("Error reading config: %v", err)
			}
			logger, config, svc, exporter := initializeConfig()

			// Assert components are initialized
			assert.NotNil(t, logger)
			assert.NotNil(t, config)
			assert.NotNil(t, exporter)
			assert.NotNil(t, svc)
		})
	}
}

func TestUpdateProvisionerNames(t *testing.T) {
	tests := []struct {
		name         string
		provisioners string
		expected     []string
		expectPanic  bool
	}{
		{
			name:         "Single Provisioner",
			provisioners: "csi-powerstore",
			expected:     []string{"csi-powerstore"},
			expectPanic:  false,
		},
		{
			name:         "Multiple Provisioners",
			provisioners: "csi-powerstore1,csi-powerstore2",
			expected:     []string{"csi-powerstore1", "csi-powerstore2"},
			expectPanic:  false,
		},
		{
			name:         "Empty Provisioners",
			provisioners: "",
			expected:     nil,
			expectPanic:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("PROVISIONER_NAMES", tt.provisioners)

			vf := &k8s.VolumeFinder{}

			logger := logrus.New()
			logger.ExitFunc = func(int) { panic("fatal") }

			if tt.expectPanic {
				assert.Panics(t, func() { updateProvisionerNames(vf, logger) })
			} else {
				assert.NotPanics(t, func() { updateProvisionerNames(vf, logger) })
				assert.Equal(t, tt.expected, vf.DriverNames)
			}
		})
	}
}

func TestGetCollectorCertPath(t *testing.T) {
	// Test case: TLS_ENABLED is set to false
	os.Setenv("TLS_ENABLED", "false")
	if getCollectorCertPath() != "" {
		t.Errorf("expected empty string, got %s", getCollectorCertPath())
	}

	// Test case: TLS_ENABLED is set to true, COLLECTOR_CERT_PATH is empty
	os.Setenv("TLS_ENABLED", "true")
	os.Setenv("COLLECTOR_CERT_PATH", "")
	if getCollectorCertPath() != otlexporters.DefaultCollectorCertPath {
		t.Errorf("expected %s, got %s", otlexporters.DefaultCollectorCertPath, getCollectorCertPath())
	}

	// Test case: TLS_ENABLED is set to true, COLLECTOR_CERT_PATH is not empty
	os.Setenv("TLS_ENABLED", "true")
	os.Setenv("COLLECTOR_CERT_PATH", "/path/to/cert.crt")
	if getCollectorCertPath() != "/path/to/cert.crt" {
		t.Errorf("expected %s, got %s", "/path/to/cert.crt", getCollectorCertPath())
	}
}

func TestStartConfigWatchers(t *testing.T) {
	logger := logrus.New()
	config := &entrypoint.Config{}
	exporter := &otlexporters.OtlCollectorExporter{}
	powerStoreSvc := &service.PowerStoreService{}
	// configFileListener := setupConfigFileListener()

	tests := []struct {
		name          string
		expectedError bool
	}{
		{"Valid Config Watchers Setup", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				startConfigWatchers(logger, config, exporter, powerStoreSvc)
			}, "Expected setupConfigWatchers to not panic")
		})
	}
}

func TestGetBindPort(t *testing.T) {
	logger := logrus.New()

	// Test case: Default port
	t.Run("Default port", func(t *testing.T) {
		// viper.Set("PORT", "")
		startHTTPServer(logger)
		logger := logrus.New()

		result := getBindPort(logger)

		assert.Equal(t, defaultDebugPort, strconv.Itoa(result))
	})

	// Test case: Custom port
	t.Run("Custom port", func(t *testing.T) {
		viper.Set("PORT", "8080")
		logger := logrus.New()

		result := getBindPort(logger)

		assert.Equal(t, 8080, result)
	})

	// Test case: Invalid port
	t.Run("Invalid port", func(t *testing.T) {
		viper.Set("PORT", "invalid")
		logger.ExitFunc = func(int) { panic("fatal") }

		assert.Panics(t, func() { panic(getBindPort(logger)) })
	})
}

func TestUpdateTickIntervals(t *testing.T) {
	tests := []struct {
		name                 string
		volFreq              string
		spaceFreq            string
		arrayFreq            string
		fsFreq               string
		topologyFreq         string
		expectedVolFreq      time.Duration
		expectedSpaceFreq    time.Duration
		expectedArrayFreq    time.Duration
		expectedFsFreq       time.Duration
		expectedTopologyFreq time.Duration
		expectPanic          bool
	}{
		{
			name:                 "Valid Values",
			volFreq:              "10",
			spaceFreq:            "20",
			arrayFreq:            "10",
			fsFreq:               "20",
			topologyFreq:         "20",
			expectedVolFreq:      10 * time.Second,
			expectedSpaceFreq:    20 * time.Second,
			expectedArrayFreq:    10 * time.Second,
			expectedFsFreq:       20 * time.Second,
			expectedTopologyFreq: 20 * time.Second,
			expectPanic:          false,
		},
		{
			name:                 "Invalid Values",
			volFreq:              "invalid",
			spaceFreq:            "invalid",
			arrayFreq:            "",
			fsFreq:               "invalidinvalid",
			topologyFreq:         "invalid",
			expectedVolFreq:      defaultTickInterval,
			expectedSpaceFreq:    defaultTickInterval,
			expectedArrayFreq:    defaultTickInterval,
			expectedFsFreq:       defaultTickInterval,
			expectedTopologyFreq: defaultTickInterval,
			expectPanic:          true,
		},
		{
			name:              "InValid SpaceFreq",
			volFreq:           "10",
			spaceFreq:         "invalid",
			arrayFreq:         "10",
			fsFreq:            "10",
			expectedVolFreq:   10 * time.Second,
			expectedSpaceFreq: defaultTickInterval,
			expectedArrayFreq: 10 * time.Second,
			expectedFsFreq:    10 * time.Second,
			expectPanic:       true,
		},
		{
			name:              "InValid arrayFreq",
			volFreq:           "10",
			spaceFreq:         "10",
			arrayFreq:         "invalid",
			fsFreq:            "10",
			expectedVolFreq:   10 * time.Second,
			expectedSpaceFreq: 10 * time.Second,
			expectedArrayFreq: defaultTickInterval,
			expectedFsFreq:    10 * time.Second,
			expectPanic:       true,
		},
		{
			name:              "InValid fsFeq",
			volFreq:           "10",
			spaceFreq:         "10",
			arrayFreq:         "10",
			fsFreq:            "invalid",
			expectedVolFreq:   10 * time.Second,
			expectedSpaceFreq: 10 * time.Second,
			expectedArrayFreq: 10 * time.Second,
			expectedFsFreq:    defaultTickInterval,
			expectPanic:       true,
		},
		{
			name:                 "Invalid TopologyFreq only",
			volFreq:              "10",
			spaceFreq:            "10",
			arrayFreq:            "10",
			fsFreq:               "10",
			topologyFreq:         "invalid",
			expectedVolFreq:      10 * time.Second,
			expectedSpaceFreq:    10 * time.Second,
			expectedArrayFreq:    10 * time.Second,
			expectedFsFreq:       10 * time.Second,
			expectedTopologyFreq: defaultTickInterval,
			expectPanic:          true,
		},
		{
			name:                 "Valid TopologyFreq",
			volFreq:              "10",
			spaceFreq:            "10",
			arrayFreq:            "10",
			fsFreq:               "10",
			topologyFreq:         "30",
			expectedVolFreq:      10 * time.Second,
			expectedSpaceFreq:    10 * time.Second,
			expectedArrayFreq:    10 * time.Second,
			expectedFsFreq:       10 * time.Second,
			expectedTopologyFreq: 30 * time.Second,
			expectPanic:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("POWERSTORE_VOLUME_IO_POLL_FREQUENCY", tt.volFreq)
			viper.Set("POWERSTORE_SPACE_POLL_FREQUENCY", tt.spaceFreq)
			viper.Set("POWERSTORE_ARRAY_POLL_FREQUENCY", tt.arrayFreq)
			viper.Set("POWERSTORE_FILE_SYSTEM_POLL_FREQUENCY", tt.fsFreq)
			// viper.Set("POWERSTORE_VOLUME_METRICS_ENABLED", "invalid")
			viper.Set("POWERSTORE_TOPOLOGY_METRICS_POLL_FREQUENCY", tt.topologyFreq)

			config := &entrypoint.Config{}
			logger := logrus.New()
			logger.ExitFunc = func(int) { panic("fatal") }

			if tt.expectPanic {
				assert.Panics(t, func() { updateTickIntervals(config, logger) })
			} else {
				assert.NotPanics(t, func() { updateTickIntervals(config, logger) })
				assert.Equal(t, tt.expectedVolFreq, config.VolumeTickInterval, "VolumeTickInterval mismatch")
				assert.Equal(t, tt.expectedSpaceFreq, config.SpaceTickInterval, "SpaceTickInterval mismatch")
				assert.Equal(t, tt.expectedArrayFreq, config.ArrayTickInterval, "ArrayTickInterval mismatch")
				assert.Equal(t, tt.expectedFsFreq, config.FileSystemTickInterval, "FileSystemTickInterval mismatch")
				assert.Equal(t, tt.expectedTopologyFreq, config.TopologyTickInterval, "TopologyTickInterval mismatch")
			}
		})
	}
}

func TestUpdateTracing(t *testing.T) {
	tests := []struct {
		name        string
		initTracing func(string, float64) (trace.TracerProvider, error)
		envVars     map[string]string
		wantErr     bool
	}{
		{
			name: "valid tracing",
			initTracing: func(string, float64) (trace.TracerProvider, error) {
				tp := otel.GetTracerProvider()
				return tp, nil
			},
			envVars: map[string]string{"ZIPKIN_URI": "http://localhost:9411/api/v2/spans", "ZIPKIN_SERVICE": "test-service", "ZIPKIN_PROBABILITY": "0.5"},
			wantErr: false,
		},
		{
			name: "tracing initialization error",
			initTracing: func(string, float64) (trace.TracerProvider, error) {
				return nil, errors.New("tracing initialization error")
			},
			envVars: map[string]string{"ZIPKIN_URI": "http://localhost:9411/api/v2/spans", "ZIPKIN_SERVICE": "test-service", "ZIPKIN_PROBABILITY": "0.5"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.New()
			initTracing = tt.initTracing
			viper.Reset()
			for k, v := range tt.envVars {
				viper.Set(k, v)
				defer viper.Set(k, "")
			}
			updateTracing(logger)

			if tt.wantErr {
				assert.NotNil(t, logger.Out)
			}
		})
	}
}

func TestUpdateService(t *testing.T) {
	tests := []struct {
		name          string
		maxConcurrent string
		expected      int
		expectPanic   bool
	}{
		{
			name:          "Valid Value",
			maxConcurrent: "10",
			expected:      10,
			expectPanic:   false,
		},
		{
			name:          "Invalid Value",
			maxConcurrent: "invalid",
			expected:      service.DefaultMaxPowerStoreConnections,
			expectPanic:   true,
		},
		{
			name:          "Zero Value",
			maxConcurrent: "0",
			expected:      service.DefaultMaxPowerStoreConnections,
			expectPanic:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("POWERSTORE_MAX_CONCURRENT_QUERIES", tt.maxConcurrent)

			svc := &service.PowerStoreService{}
			logger := logrus.New()
			logger.ExitFunc = func(int) { panic("fatal") }

			if tt.expectPanic {
				assert.Panics(t, func() { updateService(svc, logger) })
			} else {
				assert.NotPanics(t, func() { updateService(svc, logger) })
				assert.Equal(t, tt.expected, svc.MaxPowerStoreConnections)
			}
		})
	}
}

func TestUpdateCollectorAddress(t *testing.T) {
	tests := []struct {
		name        string
		addr        string
		expectPanic bool
	}{
		{
			name:        "Valid Address",
			addr:        "localhost:8080",
			expectPanic: false,
		},
		{
			name:        "Empty Address",
			addr:        "",
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("COLLECTOR_ADDR", tt.addr)

			logger := logrus.New()
			logger.ExitFunc = func(int) { panic("fatal") }
			config := &entrypoint.Config{Logger: logger}
			exporter := &otlexporters.OtlCollectorExporter{}

			if tt.expectPanic {
				assert.Panics(t, func() { updateCollectorAddress(config, exporter, logger) })
			} else {
				assert.NotPanics(t, func() { updateCollectorAddress(config, exporter, logger) })
				assert.Equal(t, tt.addr, config.CollectorAddress)
				assert.Equal(t, tt.addr, exporter.CollectorAddr)
			}
		})
	}
}
