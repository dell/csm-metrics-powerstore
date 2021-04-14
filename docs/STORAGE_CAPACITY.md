<!--
Copyright (c) 2020 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
-->

# Storage Capacity

Provides visibility into the total, used, and available capacity for a storage class and associated underlying storage construct.

To disable these metrics, set the ```enable_powerstore_metrics``` field to false in helm/values.yaml.

The [Grafana reference dashboards](https://github.com/dell/karavi-observability/blob/main/grafana/dashboards/powerstore) for storage capacity/consumption can be uploaded to your Grafana instance.

## Available Metrics from the OpenTelemetry Collector

The following metrics are available from the OpenTelemetry collector endpoint.  Please see the [GETTING STARTED GUIDE](https://github.com/dell/karavi-observability/blob/main/docs/GETTING_STARTED_GUIDE.md) for more information on deploying and configuring the OpenTelemetry collector.

### PowerFlex Metrics

| Metric | Description | Example |
| ------ | ----------- | ------- |
