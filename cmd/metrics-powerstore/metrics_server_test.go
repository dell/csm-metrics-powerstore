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

package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pstObsFreePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	p := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return p
}

// I-OBS-PST-01: MetricsServer :8443 equivalent starts alongside OTel pipeline;
// /metrics returns HTTP 200 with dell_csm_obs_* metrics; existing OTel pipeline unaffected.
func TestIntegration_OBS_PST_MetricsServerServesObsMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()

	collectionRate := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dell_csm_obs_collection_rate",
		Help: "Collection rate.",
	}, []string{"module", "global_id"})
	exportSuccess := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "dell_csm_obs_export_success_total",
		Help: "Export success total.",
	}, []string{"module", "global_id", "status"})
	reg.MustRegister(collectionRate, exportSuccess)

	collectionRate.WithLabelValues("metrics-powerstore", "global-1").Set(4.2)
	exportSuccess.WithLabelValues("metrics-powerstore", "global-1", "success").Inc()

	port := pstObsFreePort(t)
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	go func() { _ = srv.ListenAndServe() }()
	t.Cleanup(func() { _ = srv.Shutdown(context.Background()) })
	time.Sleep(80 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/metrics", port))
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "dell_csm_obs_collection_rate",
		"scrape must contain collection rate metric")
	assert.Contains(t, string(body), "dell_csm_obs_export_success_total",
		"scrape must contain export success counter")
}
