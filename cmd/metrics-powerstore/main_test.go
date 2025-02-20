package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/dell/csm-metrics-powerstore/internal/k8s"
	"github.com/dell/csm-metrics-powerstore/internal/service"
	otlexporters "github.com/dell/csm-metrics-powerstore/opentelemetry/exporters"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

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

func TestInitializeConfig(t *testing.T) {
	// Test case: configuration file is readable
	viper.SetConfigFile("testdata/config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		t.Fatalf("unable to read Config file: %v", err)
	}

	// Create a directory at the given path
	dirPath := "/etc/config"

	// Remove the directory if available
	err = os.RemoveAll(dirPath)
	if err != nil {
		fmt.Println("Error removing directory:", err)
		return
	}
	err = os.Mkdir(dirPath, os.ModePerm)
	if err != nil {
		panic(err)
	}

	viper.Set("COLLECTOR_ADDR", "localhost:4317")
	defer viper.Reset()

	viper.Set("TLS_ENABLED", false)
	// viper.Set("COLLECTOR_CERT_PATH", "testdata/cert.crt")

	// Create a file inside the directory
	filePath := filepath.Join(dirPath, "karavi-metrics-powerstore.yaml")
	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	logger := logrus.New()
	config := initializeConfig(logger)

	if config.LeaderElector == nil {
		t.Errorf("expected LeaderElector to be initialized, got nil")
	}
	fmt.Println("Certpath ", config.CollectorCertPath)

	// if config.CollectorCertPath != "testdata/cert.crt" {
	// 	t.Errorf("expected CollectorCertPath to be 'testdata/cert.crt', got %s", config.CollectorCertPath)
	// }

	if config.Logger != logger {
		t.Errorf("expected Logger to be the same as the input logger, got different logger")
	}

	// Test case: configuration file is not readable
	viper.SetConfigFile("testdata/nonexistent.yaml")
	err = viper.ReadInConfig()
	if err == nil {
		t.Fatalf("expected unable to read Config file, got nil error")
	}

	logger = logrus.New()
	config = initializeConfig(logger)

	if config.LeaderElector == nil {
		t.Errorf("expected LeaderElector to be initialized, got nil")
	}

	if config.CollectorCertPath != "" {
		t.Errorf("expected CollectorCertPath to be empty, got %s", config.CollectorCertPath)
	}

	if config.Logger != logger {
		t.Errorf("expected Logger to be the same as the input logger, got different logger")
	}

	// Delete the file
	err = os.Remove(filePath)
	if err != nil {
		panic(err)
	}
}

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

// func TestStartConfigWatchers(t *testing.T) {
// 	logger := logrus.New()
// 	config := &entrypoint.Config{}
// 	exporter := &otlexporters.OtlCollectorExporter{}
// 	powerStoreSvc := &service.PowerStoreService{}
// 	volumeFinder := &k8s.VolumeFinder{
// 		API:    &k8s.API{},
// 		Logger: logger,
// 	}

// 	// Test case: Viper configuration is updated
// 	viper.Set("test_key", "test_value")
// 	viper.Set("POWERSTORE_VOLUME_IO_POLL_FREQUENCY", "10")
// 	viper.Set("COLLECTOR_ADDR", "test_address")
// 	viper.Set("POWERSTORE_VOLUME_METRICS_ENABLED", "true")
// 	viper.OnConfigChange(func(_ fsnotify.Event) {
// 		updateLoggingSettings(logger)
// 		updateCollectorAddress(config, exporter, logger)
// 		updateProvisionerNames(volumeFinder, logger)
// 		updateMetricsEnabled(config, logger)
// 		updateTickIntervals(config, logger)
// 		updateService(powerStoreSvc, logger)
// 		updateTracing(logger)
// 	})

// 	startConfigWatchers(logger, config, exporter, powerStoreSvc)

// 	// Assert that the logging settings are updated
// 	assert.Equal(t, logrus.InfoLevel, logger.Level)
// 	assert.Equal(t, "test_value", viper.GetString("test_key"))

// 	// Assert that the collector address is updated
// 	// assert.Equal(t, "test_address", config.CollectorAddress)
// 	assert.Equal(t, "test_address", exporter.CollectorAddr)

// 	// Assert that the provisioner names are updated
// 	// assert.Equal(t, []string{"test_provisioner"}, powerStoreSvc.VolumeFinder.DriverNames)

// 	// Assert that the metrics enabled flag is updated
// 	assert.Equal(t, true, config.VolumeMetricsEnabled)

// 	// Assert that the tick intervals are updated
// 	assert.Equal(t, time.Second*10, config.VolumeTickInterval)
// 	assert.Equal(t, time.Second*10, config.SpaceTickInterval)
// 	assert.Equal(t, time.Second*10, config.ArrayTickInterval)
// 	assert.Equal(t, time.Second*10, config.FileSystemTickInterval)

// 	// Assert that the service is updated
// 	assert.Equal(t, time.Second*10, config.VolumeTickInterval)

// Assert that the tracing is updated
// assert.Equal(t, "test_uri", tracer.ZipkinURI)
// assert.Equal(t, float32(0.5), tracer.ZipkinProbability)
// assert.Equal(t, "test_service_name", tracer.ZipkinServiceName)
// }

func TestInitializePowerStoreService(t *testing.T) {
	logger := logrus.New()
	viper.Set("provisioner_names", "test-provisioner,test-provisioner-2")

	// Create a directory at the given path
	dirPath := "/powerstore-config"

	// Remove the directory if available
	err := os.RemoveAll(dirPath)
	if err != nil {
		fmt.Println("Error removing directory:", err)
		return
	}
	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		panic(err)
	}
	filePath := filepath.Join(dirPath, "config")
	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Test case: Initialize PowerStore service with valid inputs
	t.Run("Initialize PowerStore service with valid inputs", func(t *testing.T) {
		powerStoreSvc := initializePowerStoreService(logger)

		assert.NotNil(t, powerStoreSvc)
		assert.NotNil(t, powerStoreSvc.Logger)
		assert.NotNil(t, powerStoreSvc.VolumeFinder)
		assert.NotNil(t, powerStoreSvc.MetricsWrapper)

	})

}
