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

package service_test

import (
	"context"
	"testing"

	"github.com/dell/csm-metrics-powerstore/internal/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func hasMetricLabel(labels []attribute.KeyValue, key attribute.Key) bool {
	for _, label := range labels {
		if label.Key == key {
			return true
		}
	}
	return false
}

func assertMainProtocolLabel(t *testing.T, labelsValue any) {
	t.Helper()
	labels, ok := labelsValue.([]attribute.KeyValue)
	if !ok {
		t.Fatalf("expected metric labels, got %T", labelsValue)
	}
	if !hasMetricLabel(labels, "Protocol") {
		t.Fatal("expected canonical Protocol label used by main")
	}
	if hasMetricLabel(labels, "protocol") {
		t.Fatal("unexpected non-canonical lowercase protocol label")
	}
}

func TestMetricsWrapperUsesCanonicalProtocolLabel(t *testing.T) {
	tests := []struct {
		name   string
		record func(*service.MetricsWrapper) error
		key    string
	}{
		{
			name: "filesystem",
			record: func(mw *service.MetricsWrapper) error {
				return mw.RecordSpaceMetrics(context.Background(), &service.SpaceVolumeMeta{
					ID:       "filesystem-1",
					Protocol: "nfs",
				}, 1, 2)
			},
			key: "filesystem-1",
		},
		{
			name: "volume",
			record: func(mw *service.MetricsWrapper) error {
				return mw.RecordSpaceMetrics(context.Background(), &service.SpaceVolumeMeta{
					ID:       "volume-1",
					Protocol: "scsi",
				}, 1, 2)
			},
			key: "volume-1",
		},
		{
			name: "array space",
			record: func(mw *service.MetricsWrapper) error {
				return mw.RecordArraySpaceMetrics(context.Background(), "2001:db8::1", "driver", 1, 2)
			},
			key: "2001:db8::1",
		},
		{
			name: "storage class space",
			record: func(mw *service.MetricsWrapper) error {
				return mw.RecordStorageClassSpaceMetrics(context.Background(), "gold", "driver", 1, 2)
			},
			key: "gold",
		},
		{
			name: "topology",
			record: func(mw *service.MetricsWrapper) error {
				return mw.RecordTopologyMetrics(context.Background(), &service.TopologyMeta{
					PersistentVolume: "pv-1",
					Protocol:         "nfs",
				}, &service.TopologyMetricsRecord{PvAvailable: 1})
			},
			key: "pv-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := &service.MetricsWrapper{Meter: otel.Meter("canonical-protocol-" + tt.name)}
			if err := tt.record(mw); err != nil {
				t.Fatalf("record metrics: %v", err)
			}
			labelsValue, ok := mw.Labels.Load(tt.key)
			if !ok {
				t.Fatalf("expected labels for %q", tt.key)
			}
			assertMainProtocolLabel(t, labelsValue)
		})
	}
}
