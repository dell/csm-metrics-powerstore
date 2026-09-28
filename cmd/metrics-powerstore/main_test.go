/*
 *
 * Copyright © 2021-2026 Dell Inc. or its subsidiaries. All Rights Reserved.
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

package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dell/csm-metrics-powerstore/internal/entrypoint"
	"github.com/dell/csm-metrics-powerstore/internal/k8s"
	"github.com/dell/csm-metrics-powerstore/internal/service"
	otlexporters "github.com/dell/csm-metrics-powerstore/opentelemetry/exporters"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func TestInitializeConfig(t *testing.T) {
	// Mock getPowerScaleClusters to avoid file I/O
	originalGetPowerStoreArrays := getPowerStoreArrays
	defer func() { getPowerStoreArrays = originalGetPowerStoreArrays }()
	getPowerStoreArrays = func(_ string) (map[string]*service.PowerStoreArray, map[string]string, *service.PowerStoreArray, error) {
		//nolint:goconst // Test mock data with repeated strings
		return map[string]*service.PowerStoreArray{
			"PowerStore123": {
				Endpoint:  "10.10.10.10",
				GlobalID:  "PowerStore123",
				IsDefault: true,
				IP:        "2001:db8::1",
			},
		}, map[string]string{"cluster1": "10.10.10.10"}, &service.PowerStoreArray{
			Endpoint:  "10.10.10.10",
			GlobalID:  "PowerStore123",
			IsDefault: true,
			IP:        "2001:db8::1",
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
			expectedCollectorAddr:          "localhost:4317",
			expectedProvisioners:           []string{"csi-powerstore"},
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
			config, svc, exporter := initializeConfig()

			// Assert components are initialized
			assert.NotNil(t, config)
			assert.NotNil(t, exporter)
			assert.NotNil(t, svc)
			assert.Equal(t, "2001:db8::1", svc.PowerStoreArrays["PowerStore123"].IP)
			assert.Equal(t, "PowerStore123", svc.DefaultPowerStoreArray.GlobalID)
		})
	}
}

type stubFailureRecorder struct {
	callback func()
}

func (s *stubFailureRecorder) SetExportFailureRecorder(callback func()) {
	s.callback = callback
}

func TestWireExportFailureRecorder_SetsCallbackWithoutMetricsServer(t *testing.T) {
	exporter := &stubFailureRecorder{}
	svc := &service.PowerStoreService{}

	wirePowerStoreExportFailureRecorder(svc, exporter)

	require.NotNil(t, exporter.callback)
	assert.NotPanics(t, func() { exporter.callback() })
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("PROVISIONER_NAMES", tt.provisioners)

			vf := &k8s.VolumeFinder{}

			if tt.expectPanic {
				assert.Panics(t, func() { updateProvisionerNames(vf) })
			} else {
				assert.NotPanics(t, func() { updateProvisionerNames(vf) })
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
				startConfigWatchers(config, exporter, powerStoreSvc)
			}, "Expected setupConfigWatchers to not panic")
		})
	}
}

func TestGetBindPort(t *testing.T) {
	// Test case: Default port
	t.Run("Default port", func(t *testing.T) {
		startHTTPServer()

		result := getBindPort()

		assert.Equal(t, defaultDebugPort, strconv.Itoa(result))
	})

	// Test case: Custom port
	t.Run("Custom port", func(t *testing.T) {
		viper.Set("PORT", "8080")

		result := getBindPort()

		assert.Equal(t, 8080, result)
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
		{
			name:                 "Default Values",
			volFreq:              "",
			spaceFreq:            "",
			arrayFreq:            "",
			fsFreq:               "",
			topologyFreq:         "",
			expectedVolFreq:      300 * time.Second,
			expectedSpaceFreq:    300 * time.Second,
			expectedArrayFreq:    300 * time.Second,
			expectedFsFreq:       300 * time.Second,
			expectedTopologyFreq: 300 * time.Second,
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
			if tt.expectPanic {
				assert.Panics(t, func() { updateTickIntervals(config) })
			} else {
				assert.NotPanics(t, func() { updateTickIntervals(config) })
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
		t.Run(tt.name, func(_ *testing.T) {
			initTracing = tt.initTracing
			viper.Reset()
			for k, v := range tt.envVars {
				viper.Set(k, v)
				defer viper.Set(k, "")
			}
			updateTracing()

			// Note: updateTracing logs errors but doesn't return them, so we can't assert on error state
		})
	}
}

func TestUpdateService(t *testing.T) {
	tests := []struct {
		name                   string
		maxConcurrent          string
		expectedMaxConnections int
		expectPanic            bool
	}{
		{
			name:                   "Valid Value",
			maxConcurrent:          "10",
			expectedMaxConnections: 10,
			expectPanic:            false,
		},
		{
			name:                   "Non-default max concurrent queries value is applied",
			maxConcurrent:          "6",
			expectedMaxConnections: 6,
			expectPanic:            false,
		},
		{
			name:                   "Default max concurrent queries value is applied",
			maxConcurrent:          "",
			expectedMaxConnections: 5,
			expectPanic:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("POWERSTORE_MAX_CONCURRENT_QUERIES", tt.maxConcurrent)

			svc := &service.PowerStoreService{}

			if tt.expectPanic {
				assert.Panics(t, func() { updateService(svc) })
			} else {
				assert.NotPanics(t, func() { updateService(svc) })
				assert.Equal(t, tt.expectedMaxConnections, svc.MaxPowerStoreConnections)
			}
		})
	}
}

// I-OBS-PST-02: startMetricsServer wires ObsInstrumenter onto the service and
// exposes all four dell_csm_obs_* metrics at /metrics (HTTP 200).
func TestStartMetricsServer_WiresInstrumenterAndServesMetrics(t *testing.T) {
	port := pstObsFreePort(t)
	viper.Set(csiObsMetricsPortKey, fmt.Sprintf("%d", port))
	t.Cleanup(func() { viper.Set(csiObsMetricsPortKey, "") })

	svc := &service.PowerStoreService{}

	startMetricsServer(svc, &otlexporters.OtlCollectorExporter{})

	assert.NotNil(t, svc.ObsInstrumenter, "startMetricsServer must set ObsInstrumenter on the service")

	// Record one observation per metric so they appear in the /metrics output
	// (prometheus Vec metrics only appear after at least one observation).
	svc.ObsInstrumenter.RecordCollectionRate("test", 1)
	svc.ObsInstrumenter.RecordExportSuccess("test", "success")
	svc.ObsInstrumenter.SetArrayConnectivity("test", true)
	svc.ObsInstrumenter.RecordProcessingLatency("test", 0.001)

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/metrics", port))
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	for _, metric := range []string{
		"dell_csm_obs_collection_rate",
		"dell_csm_obs_export_success_total",
		"dell_csm_obs_array_connectivity",
		"dell_csm_obs_processing_latency_seconds",
	} {
		assert.Contains(t, string(body), metric, "metric %q must appear in /metrics output", metric)
	}
}

// I-OBS-PST-03: startMetricsServer wires ObsInstrumenter regardless of port source.
func TestStartMetricsServer_WiresInstrumenterEvenWithDefaultPort(t *testing.T) {
	port := pstObsFreePort(t)
	viper.Set(csiObsMetricsPortKey, fmt.Sprintf("%d", port))
	t.Cleanup(func() { viper.Set(csiObsMetricsPortKey, "") })

	svc := &service.PowerStoreService{}

	startMetricsServer(svc, &otlexporters.OtlCollectorExporter{})

	assert.NotNil(t, svc.ObsInstrumenter)
}

// I-OBS-PST-04: /metrics returns HTTP 404 for unknown paths.
func TestStartMetricsServer_UnknownPathReturns404(t *testing.T) {
	port := pstObsFreePort(t)
	viper.Set(csiObsMetricsPortKey, fmt.Sprintf("%d", port))
	t.Cleanup(func() { viper.Set(csiObsMetricsPortKey, "") })

	svc := &service.PowerStoreService{}

	startMetricsServer(svc, &otlexporters.OtlCollectorExporter{})
	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("http://localhost:%d/unknown", port), nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// I-OBS-PST-05: startMetricsServer serves metrics over HTTPS when TLS is enabled with valid certificates.
func TestStartMetricsServer_TLSServesMetricsWithValidCerts(t *testing.T) {
	viper.Reset()
	viper.AutomaticEnv()
	port := pstObsFreePort(t)
	t.Setenv(csiObsMetricsPortKey, fmt.Sprintf("%d", port))

	// Create temporary TLS certificate and key files
	certFile := t.TempDir() + "/test-cert.pem"
	keyFile := t.TempDir() + "/test-key.pem"

	// Generate a self-signed certificate for testing
	cmd := exec.Command("openssl", "req", "-new", "-x509", "-sha256", "-keyout", keyFile,
		"-out", certFile, "-days", "1", "-nodes", "-subj", "/CN=localhost")
	if err := cmd.Run(); err != nil {
		t.Skipf("Skipping TLS test: openssl not available: %v", err)
	}
	t.Cleanup(func() {
		os.Remove(certFile)
		os.Remove(keyFile)
	})

	t.Setenv(csiObsMetricsCertKey, certFile)
	t.Setenv(csiObsMetricsKeyKey, keyFile)

	svc := &service.PowerStoreService{}

	startMetricsServer(svc, &otlexporters.OtlCollectorExporter{})

	assert.NotNil(t, svc.ObsInstrumenter, "startMetricsServer must set ObsInstrumenter on the service")

	// Record observations so metrics appear in output
	svc.ObsInstrumenter.RecordCollectionRate("test", 1)
	svc.ObsInstrumenter.RecordExportSuccess("test", "success")
	svc.ObsInstrumenter.SetArrayConnectivity("test", true)
	svc.ObsInstrumenter.RecordProcessingLatency("test", 0.001)

	time.Sleep(100 * time.Millisecond)

	// Create HTTP client that skips TLS verification for self-signed cert
	client := &http.Client{
		Transport: &http.Transport{
			// #nosec G402 -- InsecureSkipVerify is intentional for self-signed test cert
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Get(fmt.Sprintf("https://localhost:%d/metrics", port))
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	for _, metric := range []string{
		"dell_csm_obs_collection_rate",
		"dell_csm_obs_export_success_total",
		"dell_csm_obs_array_connectivity",
		"dell_csm_obs_processing_latency_seconds",
	} {
		assert.Contains(t, string(body), metric, "metric %q must appear in /metrics output", metric)
	}
}

// I-OBS-PST-06: validateTLSFiles returns an error when certificate files are missing.
// Note: startMetricsServer calls csmlog.Fatal (os.Exit) on TLS validation failure, so we
// test the underlying validateTLSFiles function directly to avoid process termination in tests.
func TestStartMetricsServer_TLSFailsWithMissingCerts(t *testing.T) {
	err := validateTLSFiles("/nonexistent/cert.pem", "/nonexistent/key.pem")
	assert.Error(t, err, "validateTLSFiles should return error for missing files")
	assert.Contains(t, err.Error(), "does not exist", "error should indicate file does not exist")
}

// I-OBS-PST-07: validateTLSFiles returns error for empty cert/key paths.
func TestValidateTLSFiles_EmptyPaths(t *testing.T) {
	err := validateTLSFiles("", "")
	assert.Error(t, err, "validateTLSFiles should return error for empty paths")
	assert.Contains(t, err.Error(), "empty", "error should indicate empty path")

	err = validateTLSFiles("/valid/cert.pem", "")
	assert.Error(t, err, "validateTLSFiles should return error for empty key path")
	assert.Contains(t, err.Error(), "key", "error should mention key")

	err = validateTLSFiles("", "/valid/key.pem")
	assert.Error(t, err, "validateTLSFiles should return error for empty cert path")
	assert.Contains(t, err.Error(), "certificate", "error should mention certificate")
}

func TestUpdateLoggingSettings_JSONFormat(t *testing.T) {
	viper.Reset()
	viper.Set("LOG_FORMAT", "json")
	viper.Set("LOG_LEVEL", "invalid_level")

	assert.NotPanics(t, func() { updateLoggingSettings() })
}

func TestUpdateMetricsEnabled_Disabled(t *testing.T) {
	viper.Reset()
	viper.Set("POWERSTORE_VOLUME_METRICS_ENABLED", "false")
	viper.Set("POWERSTORE_TOPOLOGY_METRICS_ENABLED", "false")

	config := &entrypoint.Config{}
	updateMetricsEnabled(config)

	assert.False(t, config.VolumeMetricsEnabled)
	assert.False(t, config.TopologyMetricsEnabled)
}

func TestGetBindPort_EmptyPort(t *testing.T) {
	viper.Reset()
	// PORT not set → should return 0
	result := getBindPort()
	assert.Equal(t, 0, result)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("COLLECTOR_ADDR", tt.addr)

			config := &entrypoint.Config{}
			exporter := &otlexporters.OtlCollectorExporter{}

			if tt.expectPanic {
				assert.Panics(t, func() { updateCollectorAddress(config, exporter) })
			} else {
				assert.NotPanics(t, func() { updateCollectorAddress(config, exporter) })
				assert.Equal(t, tt.addr, config.CollectorAddress)
				assert.Equal(t, tt.addr, exporter.CollectorAddr)
			}
		})
	}
}

func TestValidateTLSFiles_CertAccessError(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "notadir")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	keyFile := t.TempDir() + "/test-key.pem"
	err = os.WriteFile(keyFile, []byte("key"), 0o600)
	require.NoError(t, err)

	certPath := tmpFile.Name() + "/cert.pem"
	err = validateTLSFiles(certPath, keyFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to access TLS certificate file")
}

func TestValidateTLSFiles_KeyAccessError(t *testing.T) {
	tmpDir := t.TempDir()
	certFile := tmpDir + "/test-cert.pem"
	err := os.WriteFile(certFile, []byte("cert"), 0o600)
	require.NoError(t, err)

	notADir, err := os.CreateTemp("", "notadir")
	require.NoError(t, err)
	defer os.Remove(notADir.Name())
	notADir.Close()

	keyPath := notADir.Name() + "/key.pem"
	err = validateTLSFiles(certFile, keyPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to access TLS key file")
}

func TestObsGlobalID_WithValidService(t *testing.T) {
	svc := &service.PowerStoreService{
		DefaultPowerStoreArray: &service.PowerStoreArray{
			GlobalID: "PS123456",
		},
	}
	result := obsGlobalID(svc)
	assert.Equal(t, "PS123456", result)
}

func TestObsGlobalID_WithNilService(t *testing.T) {
	result := obsGlobalID(nil)
	assert.Equal(t, "powerstore", result)
}

func TestObsGlobalID_WithNilDefaultArray(t *testing.T) {
	svc := &service.PowerStoreService{
		DefaultPowerStoreArray: nil,
	}
	result := obsGlobalID(svc)
	assert.Equal(t, "powerstore", result)
}

func TestObsGlobalID_WithEmptyGlobalID(t *testing.T) {
	svc := &service.PowerStoreService{
		DefaultPowerStoreArray: &service.PowerStoreArray{
			GlobalID: "",
		},
	}
	result := obsGlobalID(svc)
	assert.Equal(t, "powerstore", result)
}
