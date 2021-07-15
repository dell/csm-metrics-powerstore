// Copyright (c) 2021 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

package service

// VolumeMeta is the details of a volume in an SDC
type VolumeMeta struct {
	ID                   string
	PersistentVolumeName string
	ArrayID              string
	StorageClass         string
}

// StorageClassInfo is meta data about a storage class
type StorageClassInfo struct {
	ID              string
	Name            string
	Driver          string
	StorageSystemID string
}

// TotalSpaceMeta is meta data about array and storage class space
type TotalSpaceMeta struct {
	ArrayID      string
	StorageClass string
}
