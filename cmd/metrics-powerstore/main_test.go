package main

import (
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
TLS_ENABLED: false
`
	err := viper.ReadConfig(strings.NewReader(configContent))
	if err != nil {
		// Handle the error or log it
		log.Printf("Error reading config: %v", err)
	}

	// Test case: configuration file is readable
	// viper.SetConfigFile("testdata/config.yaml")
	// err = viper.ReadInConfig()
	// if err != nil {
	// 	t.Fatalf("unable to read Config file: %v", err)
	// }

	// // Create a directory at the given path
	// dirPath := "/etc/config"

	// // Remove the directory if available
	// err = os.RemoveAll(dirPath)
	// if err != nil {
	// 	fmt.Println("Error removing directory:", err)
	// 	return
	// }
	// err = os.Mkdir(dirPath, os.ModePerm)
	// if err != nil {
	// 	panic(err)
	// }

	// viper.Set("COLLECTOR_ADDR", "localhost:4317")
	// defer viper.Reset()

	// viper.Set("TLS_ENABLED", false)
	// // viper.Set("COLLECTOR_CERT_PATH", "testdata/cert.crt")

	// // Create a file inside the directory
	// filePath := filepath.Join(dirPath, "karavi-metrics-powerstore.yaml")
	// file, err := os.Create(filePath)
	// if err != nil {
	// 	panic(err)
	// }
	// defer file.Close()

	// logger := logrus.New()
	//

	tests := []struct {
		name                  string
		envVars               map[string]string
		expectedLogLevel      logrus.Level
		expectedCollectorAddr string
		expectedProvisioners  []string
		expectedCertPath      string
	}{
		{
			name: "SuccessfulInitializationWithDefaults",
			envVars: map[string]string{
				"LOG_LEVEL":                         "debug",
				"COLLECTOR_ADDR":                    "localhost:4317",
				"PROVISIONER_NAMES":                 "csi-powerstore",
				"POWERSTORE_VOLUME_METRICS_ENABLED": "true",
				"TLS_ENABLED":                       "false",
			},
			expectedLogLevel:      logrus.DebugLevel,
			expectedCollectorAddr: "localhost:4317",
			expectedProvisioners:  []string{"csi-isilon"},
			expectedCertPath:      otlexporters.DefaultCollectorCertPath,
		},
		{
			name: "TLSEnabledWithCustomCertPath",
			envVars: map[string]string{
				"LOG_LEVEL":           "info",
				"COLLECTOR_ADDR":      "collector:4317",
				"PROVISIONER_NAMES":   "csi-powerstore",
				"TLS_ENABLED":         "true",
				"COLLECTOR_CERT_PATH": "/custom/cert/path",
			},
			expectedLogLevel:      logrus.InfoLevel,
			expectedCollectorAddr: "collector:4317",
			expectedProvisioners:  []string{"csi-powerstore"},
			expectedCertPath:      "/custom/cert/path",
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
	// Test case: Update provisioner names with valid input
	t.Run("Update provisioner names with valid input", func(t *testing.T) {
		// Create a new logger for the test
		logger := logrus.New()
		// Create a new VolumeFinder for the test
		volumeFinder := &k8s.VolumeFinder{
			Logger: logger,
		}
		// Set the provisioner names in the test
		viper.Set("provisioner_names", "test-provisioner,test-provisioner-2")
		// Call the function
		updateProvisionerNames(volumeFinder, logger)
		// Assert the expected result
		assert.Equal(t, []string{"test-provisioner", "test-provisioner-2"}, volumeFinder.DriverNames)
	})
}

func Test_updateService(t *testing.T) {
	// Test case: Update service with valid maxPowerStoreConcurrentRequests
	t.Run("Update service with valid maxPowerStoreConcurrentRequests", func(t *testing.T) {
		// Create a new logger for the test
		logger := logrus.New()
		// Create a new PowerStoreService for the test
		powerStoreSvc := &service.PowerStoreService{
			Logger: logger,
		}
		// Set the maxPowerStoreConcurrentRequests in the test
		viper.Set("POWERSTORE_MAX_CONCURRENT_QUERIES", "10")
		// Call the function
		updateService(powerStoreSvc, logger)
		// Assert the expected result
		assert.Equal(t, 10, powerStoreSvc.MaxPowerStoreConnections)
	})

	// Test case: Update service with no maxPowerStoreConcurrentRequests
	t.Run("Update service with no maxPowerStoreConcurrentRequests", func(t *testing.T) {
		// Create a new logger for the test
		logger := logrus.New()
		// Create a new PowerStoreService for the test
		powerStoreSvc := &service.PowerStoreService{
			Logger: logger,
		}
		// Set the maxPowerStoreConcurrentRequests in the test
		viper.Set("POWERSTORE_MAX_CONCURRENT_QUERIES", "")
		// Call the function
		updateService(powerStoreSvc, logger)
		// Assert the expected result
		assert.Equal(t, service.DefaultMaxPowerStoreConnections, powerStoreSvc.MaxPowerStoreConnections)
	})
}

// func TestInitializeConfig(t *testing.T) {
// 	// Test case: configuration file is readable
// 	viper.SetConfigFile("testdata/config.yaml")
// 	err := viper.ReadInConfig()
// 	if err != nil {
// 		t.Fatalf("unable to read Config file: %v", err)
// 	}

// 	// Create a directory at the given path
// 	dirPath := "/etc/config"

// 	// Remove the directory if available
// 	err = os.RemoveAll(dirPath)
// 	if err != nil {
// 		fmt.Println("Error removing directory:", err)
// 		return
// 	}
// 	err = os.Mkdir(dirPath, os.ModePerm)
// 	if err != nil {
// 		panic(err)
// 	}

// 	viper.Set("COLLECTOR_ADDR", "localhost:4317")
// 	defer viper.Reset()

// 	viper.Set("TLS_ENABLED", false)
// 	// viper.Set("COLLECTOR_CERT_PATH", "testdata/cert.crt")

// 	// Create a file inside the directory
// 	filePath := filepath.Join(dirPath, "karavi-metrics-powerstore.yaml")
// 	file, err := os.Create(filePath)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer file.Close()

// 	logger := logrus.New()
// 	logger, config, powerStoreSvc, exporter = initializeConfig()

// 	if config.LeaderElector == nil {
// 		t.Errorf("expected LeaderElector to be initialized, got nil")
// 	}
// 	fmt.Println("Certpath ", config.CollectorCertPath)

// 	// if config.CollectorCertPath != "testdata/cert.crt" {
// 	// 	t.Errorf("expected CollectorCertPath to be 'testdata/cert.crt', got %s", config.CollectorCertPath)
// 	// }

// 	if config.Logger != logger {
// 		t.Errorf("expected Logger to be the same as the input logger, got different logger")
// 	}

// 	// Test case: configuration file is not readable
// 	viper.SetConfigFile("testdata/nonexistent.yaml")
// 	err = viper.ReadInConfig()
// 	if err == nil {
// 		t.Fatalf("expected unable to read Config file, got nil error")
// 	}

// 	logger = logrus.New()
// 	config = initializeConfig(logger)

// 	if config.LeaderElector == nil {
// 		t.Errorf("expected LeaderElector to be initialized, got nil")
// 	}

// 	if config.CollectorCertPath != "" {
// 		t.Errorf("expected CollectorCertPath to be empty, got %s", config.CollectorCertPath)
// 	}

// 	if config.Logger != logger {
// 		t.Errorf("expected Logger to be the same as the input logger, got different logger")
// 	}

// 	// Delete the file
// 	err = os.Remove(filePath)
// 	if err != nil {
// 		panic(err)
// 	}
// }

func TestGetCollectorCertPath(t *testing.T) {
	// Test case: TLS_ENABLED is not set
	os.Setenv("TLS_ENABLED", "")
	if getCollectorCertPath() != "" {
		t.Errorf("expected empty string, got %s", getCollectorCertPath())

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
}

// func TestInitializePowerStoreService(t *testing.T) {
// 	logger := logrus.New()
// 	config := &entrypoint.Config{}
// 	exporter := &otlexporters.OtlCollectorExporter{}
// 	powerStoreSvc := &service.PowerStoreService{}
// 	viper.Set("provisioner_names", "test-provisioner,test-provisioner-2")

// 	// Create a directory at the given path
// 	dirPath := "/powerstore-config"

// 	// Remove the directory if available
// 	err := os.RemoveAll(dirPath)
// 	if err != nil {
// 		fmt.Println("Error removing directory:", err)
// 		return
// 	}
// 	err = os.MkdirAll(dirPath, os.ModePerm)
// 	if err != nil {
// 		panic(err)
// 	}
// 	filePath := filepath.Join(dirPath, "config")
// 	file, err := os.Create(filePath)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer file.Close()

// 	// Test case: Initialize PowerStore service with valid inputs
// 	t.Run("Initialize PowerStore service with valid inputs", func(t *testing.T) {
// 		config, powerStoreSvc = initializePowerStoreService(logger)

// 		assert.NotNil(t, powerStoreSvc)
// 		assert.NotNil(t, powerStoreSvc.Logger)
// 		assert.NotNil(t, powerStoreSvc.VolumeFinder)
// 		assert.NotNil(t, powerStoreSvc.MetricsWrapper)

// 	})

// 	viper.Reset()

// }

func TestStartConfigWatchers(t *testing.T) {
	logger := logrus.New()
	config := &entrypoint.Config{}
	exporter := &otlexporters.OtlCollectorExporter{}
	powerStoreSvc := &service.PowerStoreService{}

	// Test case: Viper configuration is updated
	viper.Set("test_key", "test_value")
	viper.Set("POWERSTORE_VOLUME_IO_POLL_FREQUENCY", "30")
	viper.Set("POWERSTORE_SPACE_POLL_FREQUENCY", "20")
	viper.Set("POWERSTORE_ARRAY_POLL_FREQUENCY", "10")
	viper.Set("POWERSTORE_FILE_SYSTEM_POLL_FREQUENCY", "10")
	viper.Set("COLLECTOR_ADDR", "test_address")
	viper.Set("POWERSTORE_VOLUME_METRICS_ENABLED", "true")
	viper.Set("LOG_LEVEL", "debug")
	logger, config, powerStoreSvc, exporter = initializeConfig()

	startConfigWatchers(logger, config, exporter, powerStoreSvc)

	// Assert that the logging settings are updated
	assert.Equal(t, logrus.DebugLevel, logger.Level)
	assert.Equal(t, "test_value", viper.GetString("test_key"))
	// Assert that the metrics enabled flag is updated
	assert.Equal(t, true, config.VolumeMetricsEnabled)

	// Assert that the tick intervals are updated
	assert.Equal(t, time.Second*30, config.VolumeTickInterval)
	assert.Equal(t, time.Second*20, config.SpaceTickInterval)
	assert.Equal(t, time.Second*10, config.ArrayTickInterval)
	assert.Equal(t, time.Second*10, config.FileSystemTickInterval)

	viper.Reset()
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

		// getBindPort(logger)
		// expectedOutput := "port value is invalid"
		assert.Panics(t, func() { panic(getBindPort(logger)) })
	})
}

func TestUpdateTickIntervals(t *testing.T) {
	tests := []struct {
		name      string
		volFreq   string
		spaceFreq string
		arrayFreq string
		fsFreq    string
		// volMetricsFreq string
		expectedVolFreq   time.Duration
		expectedSpaceFreq time.Duration
		expectedArrayFreq time.Duration
		expectedFsFreq    time.Duration
		// expectedVolMetricsFreq  time.Duration
		expectPanic bool
	}{
		{
			name:      "Valid Values",
			volFreq:   "10",
			spaceFreq: "20",
			arrayFreq: "10",
			fsFreq:    "20",
			// volMetricsFreq: "10",
			expectedVolFreq:   10 * time.Second,
			expectedSpaceFreq: 20 * time.Second,
			expectedArrayFreq: 10 * time.Second,
			expectedFsFreq:    20 * time.Second,
			// expectedVolMetricsFreq:  10 * time.Second,
			expectPanic: false,
		},
		{
			name:      "InValid Values",
			volFreq:   "invalid",
			spaceFreq: "invalid",
			arrayFreq: "",
			fsFreq:    "invalidinvalid",
			// volMetricsFreq: "10",
			expectedVolFreq:   defaultTickInterval,
			expectedSpaceFreq: defaultTickInterval,
			expectedArrayFreq: defaultTickInterval,
			expectedFsFreq:    defaultTickInterval,
			// expectedVolMetricsFreq: defaultTickInterval,
			expectPanic: true,
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

			config := &entrypoint.Config{}
			logger := logrus.New()
			logger.ExitFunc = func(int) { panic("fatal") }

			if tt.expectPanic {
				assert.Panics(t, func() { updateTickIntervals(config, logger) })
			} else {
				assert.NotPanics(t, func() { updateTickIntervals(config, logger) })
				assert.Equal(t, time.Second*10, config.VolumeTickInterval)
				assert.Equal(t, time.Second*20, config.SpaceTickInterval)
				assert.Equal(t, time.Second*10, config.ArrayTickInterval)
				assert.Equal(t, time.Second*20, config.FileSystemTickInterval)
			}
		})
	}
}
