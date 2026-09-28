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
	"time"

	csmmodule "github.com/dell/csm-metrics-common/pkg/module"
	"github.com/prometheus/client_golang/prometheus"
)

const obsModuleLabel = "metrics-powerstore"

// PSTObsInstrumenter records observability self-metrics for csm-metrics-powerstore.
type PSTObsInstrumenter struct {
	instrumenter *csmmodule.ObsInstrumenter
}

// NewPSTObsInstrumenter creates and registers a PSTObsInstrumenter.
func NewPSTObsInstrumenter(reg prometheus.Registerer) *PSTObsInstrumenter {
	return &PSTObsInstrumenter{instrumenter: csmmodule.NewObsInstrumenter(reg, "", "global_id")}
}

// RecordCollectionRate sets the current collection rate.
func (i *PSTObsInstrumenter) RecordCollectionRate(globalID string, rate float64) {
	if i == nil || i.instrumenter == nil {
		return
	}
	i.instrumenter.RecordCollectionRate(obsModuleLabel, globalID, rate)
}

// RecordExportSuccess increments the export success counter.
func (i *PSTObsInstrumenter) RecordExportSuccess(globalID, status string) {
	if i == nil || i.instrumenter == nil {
		return
	}
	i.instrumenter.RecordExportSuccess(obsModuleLabel, globalID, status)
}

// SetArrayConnectivity sets the array connectivity gauge.
func (i *PSTObsInstrumenter) SetArrayConnectivity(globalID string, connected bool) {
	if i == nil || i.instrumenter == nil {
		return
	}
	i.instrumenter.RecordArrayConnectivity(obsModuleLabel, globalID, connected)
}

// RecordProcessingLatency observes a processing latency sample.
func (i *PSTObsInstrumenter) RecordProcessingLatency(globalID string, seconds float64) {
	if i == nil || i.instrumenter == nil {
		return
	}
	i.instrumenter.RecordProcessingLatency(obsModuleLabel, globalID, time.Duration(seconds*float64(time.Second)))
}
