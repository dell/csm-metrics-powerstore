/*
 Copyright (c) 2026 Dell Inc. or its subsidiaries. All Rights Reserved.

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

package service

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func gatherPSTObsMetric(t *testing.T, reg prometheus.Gatherer, name string) *dto.MetricFamily {
	t.Helper()
	mfs, err := reg.Gather()
	require.NoError(t, err)
	for _, mf := range mfs {
		if mf.GetName() == name {
			return mf
		}
	}
	return nil
}

func metricLabelsMatchPSTObs(m *dto.Metric, labels map[string]string) bool {
	got := make(map[string]string)
	for _, lp := range m.GetLabel() {
		got[lp.GetName()] = lp.GetValue()
	}
	for k, v := range labels {
		if got[k] != v {
			return false
		}
	}
	return true
}

func gaugePSTObs(mf *dto.MetricFamily, labels map[string]string) (float64, bool) {
	for _, m := range mf.GetMetric() {
		if metricLabelsMatchPSTObs(m, labels) {
			return m.GetGauge().GetValue(), true
		}
	}
	return 0, false
}

func counterPSTObs(mf *dto.MetricFamily, labels map[string]string) (float64, bool) {
	for _, m := range mf.GetMetric() {
		if metricLabelsMatchPSTObs(m, labels) {
			return m.GetCounter().GetValue(), true
		}
	}
	return 0, false
}

func histogramPSTObsCount(mf *dto.MetricFamily, labels map[string]string) (uint64, bool) {
	for _, m := range mf.GetMetric() {
		if metricLabelsMatchPSTObs(m, labels) {
			return m.GetHistogram().GetSampleCount(), true
		}
	}
	return 0, false
}

// U-OBS-PST-01: ObsInstrumenter with global_id label sets collection rate.
func TestPSTObsInstrumenter_RecordCollectionRate_GlobalIDLabel(t *testing.T) {
	reg := prometheus.NewRegistry()
	inst := NewPSTObsInstrumenter(reg)

	inst.RecordCollectionRate("global-1", 8.0)

	mf := gatherPSTObsMetric(t, reg, "dell_csm_obs_collection_rate")
	require.NotNil(t, mf, "dell_csm_obs_collection_rate must be registered")

	v, ok := gaugePSTObs(mf, map[string]string{
		"module": "metrics-powerstore", "global_id": "global-1",
	})
	require.True(t, ok, "global_id label must be present")
	assert.Equal(t, 8.0, v, "collection rate must match")
}

// U-OBS-PST-02: All four self-metrics are recorded with correct labels and values.
func TestPSTObsInstrumenter_RecordsAllSelfMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	inst := NewPSTObsInstrumenter(reg)

	inst.RecordCollectionRate("global-1", 2.5)
	inst.RecordExportSuccess("global-1", "success")
	inst.SetArrayConnectivity("global-1", true)
	inst.RecordProcessingLatency("global-1", 0.25)

	labels := map[string]string{"module": "metrics-powerstore", "global_id": "global-1"}

	rateMetric := gatherPSTObsMetric(t, reg, "dell_csm_obs_collection_rate")
	require.NotNil(t, rateMetric)
	rate, ok := gaugePSTObs(rateMetric, labels)
	require.True(t, ok)
	assert.Equal(t, 2.5, rate)

	exportMetric := gatherPSTObsMetric(t, reg, "dell_csm_obs_export_success_total")
	require.NotNil(t, exportMetric)
	exportCount, ok := counterPSTObs(exportMetric, map[string]string{"module": "metrics-powerstore", "global_id": "global-1", "status": "success"})
	require.True(t, ok)
	assert.Equal(t, 1.0, exportCount)

	connectivityMetric := gatherPSTObsMetric(t, reg, "dell_csm_obs_array_connectivity")
	require.NotNil(t, connectivityMetric)
	connectivity, ok := gaugePSTObs(connectivityMetric, labels)
	require.True(t, ok)
	assert.Equal(t, 1.0, connectivity)

	latencyMetric := gatherPSTObsMetric(t, reg, "dell_csm_obs_processing_latency_seconds")
	require.NotNil(t, latencyMetric)
	latencyCount, ok := histogramPSTObsCount(latencyMetric, labels)
	require.True(t, ok)
	assert.Equal(t, uint64(1), latencyCount)
}

// U-OBS-PST-03: collectionRatePerSecond returns a rate in samples/second.
func TestCollectionRatePerSecond(t *testing.T) {
	assert.Equal(t, 0.0, collectionRatePerSecond(0, time.Second))
	assert.Equal(t, 0.0, collectionRatePerSecond(10, 0))
	assert.InDelta(t, 2.0, collectionRatePerSecond(4, 2*time.Second), 1e-9)
}

// U-OBS-PST-04: A nil PSTObsInstrumenter does not panic on any method call.
func TestPSTObsInstrumenter_NilReceiverDoesNotPanic(t *testing.T) {
	var inst *PSTObsInstrumenter

	assert.NotPanics(t, func() { inst.RecordCollectionRate("global-1", 1.0) })
	assert.NotPanics(t, func() { inst.RecordExportSuccess("global-1", "success") })
	assert.NotPanics(t, func() { inst.SetArrayConnectivity("global-1", true) })
	assert.NotPanics(t, func() { inst.RecordProcessingLatency("global-1", 0.1) })
}
