package main

import (
	"strconv"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/dell/csm-metrics-powerstore/internal/service"
)

func TestUpdateService(t *testing.T) {
	t.Run("Default value when env var is not set", func(t *testing.T) {
		pstoreSvc := &service.PowerStoreService{}
		logger := logrus.New()

		viper.Set("POWERSTORE_MAX_CONCURRENT_QUERIES", "")
		updateService(pstoreSvc, logger)

		assert.Equal(t, service.DefaultMaxPowerStoreConnections, pstoreSvc.MaxPowerStoreConnections)
	})

	t.Run("Valid env var value", func(t *testing.T) {
		pstoreSvc := &service.PowerStoreService{}
		logger := logrus.New()
		validValue := 10
		viper.Set("POWERSTORE_MAX_CONCURRENT_QUERIES", strconv.Itoa(validValue))

		updateService(pstoreSvc, logger)
		assert.Equal(t, validValue, pstoreSvc.MaxPowerStoreConnections)
	})
}
