// Copyright (c) 2021 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

package otlexporters

import "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"

// Otlexporter is an interface for all OpenTelemetry exporters
//go:generate mockgen -destination=mocks/otlexporters_mocks.go -package=exportermocks github.com/dell/csm-metrics-powerstore/opentelemetry/exporters Otlexporter
type Otlexporter interface {
	InitExporter(...otlpmetricgrpc.Option) error
	StopExporter() error
}
