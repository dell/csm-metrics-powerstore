/*
 Copyright (c) 2021-2022 Dell Inc. or its subsidiaries. All Rights Reserved.

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
	"github.com/dell/gopowerstore"
)

// VolumeMeta is the details of a volume in an SDC
type VolumeMeta struct {
	ID                        string
	PersistentVolumeName      string
	PersistentVolumeClaimName string
	Namespace                 string
	ArrayID                   string
	StorageClass              string
}

// SpaceVolumeMeta is the details of a volume in an SDC
type SpaceVolumeMeta struct {
	ID                        string
	PersistentVolumeName      string
	PersistentVolumeClaimName string
	Namespace                 string
	ArrayID                   string
	StorageClass              string
	Driver                    string
	Protocol                  string
}

type TopologyMeta struct {
	Namespace               string
	PersistentVolumeClaim   string
	PersistentVolumeStatus  string
	VolumeClaimName         string
	PersistentVolume        string
	StorageClass            string
	Driver                  string
	ProvisionedSize         string
	StorageSystemVolumeName string
	StorageSystem           string
	Protocol                string
	CreatedTime             string
}

// TransportType differentiates different SCSI transport protocols (FC, iSCSI, Auto, None)
type TransportType string

// PowerStoreArray is a struct that stores all PowerStore connection information.
// It stores gopowerstore client that can be directly used to invoke PowerStore API calls.
// This structure is supposed to be parsed from config and mainly is created by GetPowerStoreArrays function.
type PowerStoreArray struct {
	Endpoint      string        `yaml:"endpoint"`
	GlobalID      string        `yaml:"globalID"`
	Username      string        `yaml:"username"`
	Password      string        `yaml:"password"`
	NasName       string        `yaml:"nasName"`
	BlockProtocol TransportType `yaml:"blockProtocol"`
	Insecure      bool          `yaml:"skipCertificateValidation"`
	IsDefault     bool          `yaml:"isDefault"`

	Client gopowerstore.Client
	IP     string
}
