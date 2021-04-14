<!--
Copyright (c) 2020 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
-->

# I/O Performance

Storage system I/O performance metrics (IOPS, bandwidth, latency) are available by default and broken down by export node and volume.

To disable these metrics, set the ```powerstore_metrics_enabled``` field to false in helm/values.yaml.

The [Grafana reference dashboards](https://github.com/dell/karavi-observability/blob/main/grafana/dashboards) for I/O metrics can but uploaded to your Grafana instance.

## Available Metrics from the OpenTelemetry Collector

The following metrics are available from the OpenTelemetry collector endpoint.  Please see the [Getting Started Guide](https://github.com/dell/karavi-observability/blob/main/docs/GETTING_STARTED_GUIDE.md) for more information on deploying and configuring the OpenTelemetry collector.

### PowerStore Metrics

| Metric | Description | Example |
| ------ | ----------- | ------- |
