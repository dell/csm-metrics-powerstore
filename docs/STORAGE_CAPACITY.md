<!--
Copyright (c) 2021 Dell Inc., or its subsidiaries. All Rights Reserved.

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

### PowerStore Metrics

| Metric | Description | Example |
| ------ | ----------- | ------- |
| powerstore_array_logical_provisioned_megabytes | Total provisioned logical storage on a given array managed by CSI driver | powerstore_array_logical_provisioned_megabytes{ArrayID="PSf16fa81b20d5",Driver="csi-powerstore.dellemc.com",PlotWithMean="No"} 15360 |
| powerstore_array_logical_used_megabytes | Total used logical storage on a given array | powerstore_array_logical_used_megabytes{ArrayID="PSf16fa81b20d5",Driver="csi-powerstore.dellemc.com",PlotWithMean="No"} 1845 |
| powerstore_storage_class_logical_provisioned_megabytes | Total provisioned logical storage for a given storage class |powerstore_storage_class_logical_provisioned_megabytes{Driver="csi-powerstore.dellemc.com",PlotWithMean="No",StorageClass="powerstore"} 15360 |
| powerstore_storage_class_logical_used_megabytes | Total used logical storage for a given storage class | powerstore_storage_class_logical_used_megabytes{Driver="csi-powerstore.dellemc.com",PlotWithMean="No",StorageClass="powerstore"} 1845 |
| powerstore_volume_logical_provisioned_megabytes | Logical provisioned storage for a volume | powerstore_volume_logical_provisioned_megabytes{ArrayID="PSf16fa81b20d5",Driver="csi-powerstore.dellemc.com",PersistentVolumeName="csi-0e8bd5eda6",PlotWithMean="No",StorageClass="powerstore",VolumeID="859e07f3-ef07-4a1f-bd8b-bd409715837e"} 3072 |
| powerstore_volume_logical_used_megabytes | Logical used storage for a volume | powerstore_volume_logical_used_megabytes{ArrayID="PSf16fa81b20d5",Driver="csi-powerstore.dellemc.com",PersistentVolumeName="csi-0e8bd5eda6",PlotWithMean="No",StorageClass="powerstore",VolumeID="859e07f3-ef07-4a1f-bd8b-bd409715837e"} 369 |
| powerstore_filesystem_logical_provisioned_megabytes | Logical provisioned storage for a filesystem | powerstore_filesystem_logical_provisioned_megabytes{ArrayID="PS4ef018459192",Driver="csi-powerstore.dellemc.com",FileSystemID="61098008-a344-f94e-6d84-5233df2e3f67",PersistentVolumeName="csi-42f21317ef",PlotWithMean="No",Protocol="nfs",StorageClass="powerstore-nfs"} 4608 |
| powerstore_filesystem_logical_used_megabytes | Logical used storage for a filesystem | powerstore_filesystem_logical_used_megabytes{ArrayID="PS4ef018459192",Driver="csi-powerstore.dellemc.com",FileSystemID="61098008-a344-f94e-6d84-5233df2e3f67",PersistentVolumeName="csi-42f21317ef",PlotWithMean="No",Protocol="nfs",StorageClass="powerstore-nfs"} 1546 |
