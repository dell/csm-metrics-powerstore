<!--
Copyright (c) 2021 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
-->

# I/O Performance

Storage system I/O performance metrics (IOPS, bandwidth, latency) are available by default and broken down by volume.

To disable these volume metrics, set the ```karaviMetricsPowerstore.volumeMetricsEnabled``` field to false in helm/values.yaml.

The [Grafana reference dashboards](https://github.com/dell/karavi-observability/blob/main/grafana/dashboards/powerstore) for I/O metrics can but uploaded to your Grafana instance.

## Available Metrics from the OpenTelemetry Collector

The following metrics are available from the OpenTelemetry collector endpoint.  Please see the [Getting Started Guide](https://github.com/dell/karavi-observability/blob/main/docs/GETTING_STARTED_GUIDE.md) for more information on deploying and configuring the OpenTelemetry collector.

### PowerStore Metrics

| Metric | Description | Example |
| ------ | ----------- | ------- |
| powerstore_volume_read_bw_megabytes_per_second	| The volume read bandwidth (MB/s) | powerstore_volume_read_bw_megabytes_per_second{ArrayID="10.0.0.1",PersistentVolumeName="pvname-f7b7382a76",PlotWithMean="No",VolumeID="daaec0fa-475b-4308-a867-c15412611a97"} 0.15727519989013672 |
| powerstore_volume_write_bw_megabytes_per_second | The volume write bandwidth (MB/s) | powerstore_volume_write_bw_megabytes_per_second{ArrayID="10.0.0.1",PersistentVolumeName="pvname-f7b7382a76",PlotWithMean="No",VolumeID="daaec0fa-475b-4308-a867-c15412611a97"} 6.03456974029541
| powerstore_volume_read_latency_milliseconds | The time (in ms) to complete read operations to a volume | powerstore_volume_read_latency_milliseconds{ArrayID="10.0.0.1",PersistentVolumeName="pvname-f7b7382a76",PlotWithMean="No",VolumeID="daaec0fa-475b-4308-a867-c15412611a97"} 0.18000000715255737 |
| powerstore_volume_write_latency_milliseconds | The time (in ms) to complete write operations to a volume | powerstore_volume_write_latency_milliseconds{ArrayID="10.0.0.1",PersistentVolumeName="pvname-f7b7382a76",PlotWithMean="No",VolumeID="daaec0fa-475b-4308-a867-c15412611a97"} 0.2669999599456787 |
| powerstore_volume_read_iops_per_second | The number of read operations performed against a volume (per second) | powerstore_volume_read_iops_per_second{ArrayID="10.0.0.1",PersistentVolumeName="pvname-f7b7382a76",PlotWithMean="No",VolumeID="daaec0fa-475b-4308-a867-c15412611a97"} 8
| powerstore_volume_write_iops_per_second | The number of write operations performed against a volume (per second) | powerstore_volume_write_iops_per_second{ArrayID="10.0.0.1",PersistentVolumeName="pvname-f7b7382a76",PlotWithMean="No",VolumeID="daaec0fa-475b-4308-a867-c15412611a97"} 12 |
